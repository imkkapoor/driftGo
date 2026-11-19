package transfer

import (
	"context"
	"driftGo/api/common/utils"
	"driftGo/domain/user"
	"driftGo/pkg/encryption"
	"driftGo/pkg/logger"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plaid/plaid-go/v35/plaid"
)

/*
Service handles all transfer-related operations
*/
type Service struct {
	client      *plaid.APIClient
	userService user.UserInterface
	database    Querier
	encryptor   *encryption.Encryptor
	pool        *pgxpool.Pool

	// lookupAccount returns the owner and access token for a stored (encrypted) account ID.
	// Injected to avoid a circular dependency with the link domain.
	lookupAccount AccountLookup
}

/*
NewService creates a new transfer service with the provided Plaid credentials and database
*/
func NewService(
	clientID, secret, env string,
	userService user.UserInterface,
	lookupAccount AccountLookup,
	db *pgxpool.Pool,
	encryptionKey string,
) (*Service, error) {
	var plaidEnv plaid.Environment
	switch env {
	case "sandbox":
		plaidEnv = plaid.Sandbox
	case "production":
		plaidEnv = plaid.Production
	default:
		return nil, errors.New("Invalid Plaid environment: " + env)
	}

	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	configuration.AddDefaultHeader("PLAID-SECRET", secret)
	configuration.UseEnvironment(plaidEnv)

	client := plaid.NewAPIClient(configuration)

	encryptor, err := encryption.NewEncryptor(encryptionKey)
	if err != nil {
		return nil, err
	}

	return &Service{
		client:        client,
		userService:   userService,
		database:      New(db),
		encryptor:     encryptor,
		pool:          db,
		lookupAccount: lookupAccount,
	}, nil
}

/*
InitiateTransfer orchestrates a two-leg transfer:
1. Validates the sender and receiver
2. Creates a DB record
3. Authorizes and creates the debit transfer (sender → Plaid Ledger)
The credit leg (Plaid Ledger → receiver) is triggered asynchronously when the debit settles via webhook
*/
func (s *Service) InitiateTransfer(ctx context.Context, senderAccountID string, receiverUserID int64, receiverAccountID string, amount string, description string) (*Transfer, error) {
	senderUserID := utils.GetUserID(ctx)
	if senderUserID == 0 {
		return nil, errors.New("user ID not found in context")
	}

	// Validate no self-transfer
	if senderUserID == receiverUserID {
		return nil, ErrSelfTransfer
	}

	// Validate amount
	amountVal, ok := new(big.Float).SetString(amount)
	if !ok {
		return nil, ErrInvalidAmount
	}
	if amountVal.Cmp(big.NewFloat(0)) <= 0 {
		return nil, ErrInvalidAmount
	}

	// Plaid requires the account holder's legal name on each leg.
	sender, err := s.userService.GetUserByID(ctx, senderUserID)
	if err != nil {
		return nil, fmt.Errorf("sender user not found: %w", err)
	}
	senderName, err := legalName(sender)
	if err != nil {
		return nil, err
	}

	receiver, err := s.userService.GetUserByID(ctx, receiverUserID)
	if err != nil {
		return nil, fmt.Errorf("receiver user not found: %w", err)
	}
	if _, err := legalName(receiver); err != nil {
		return nil, err
	}

	// Account IDs arrive in their stored (encrypted) form. Both must exist and
	// belong to the right people before any money moves.
	senderOwnerID, senderAccessToken, err := s.lookupAccount(ctx, senderAccountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to look up sender account: %w", err)
	}
	if senderOwnerID != senderUserID {
		return nil, ErrAccountNotOwned
	}

	receiverOwnerID, _, err := s.lookupAccount(ctx, receiverAccountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReceiverNoAccount
		}
		return nil, fmt.Errorf("failed to look up receiver account: %w", err)
	}
	if receiverOwnerID != receiverUserID {
		return nil, ErrReceiverNoAccount
	}

	// Decrypt account IDs for Plaid API calls
	decryptedSenderAccountID, err := s.encryptor.Decrypt(senderAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt sender account ID: %w", err)
	}

	// Generate idempotency key
	idempotencyKey := uuid.New().String()

	// Parse amount for DB storage
	pgAmount := pgtype.Numeric{}
	if err := pgAmount.Scan(amount); err != nil {
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}

	// Create transfer record (account IDs are already encrypted from the client)
	transfer, err := s.database.CreateTransfer(ctx, CreateTransferParams{
		SenderUserID:      senderUserID,
		ReceiverUserID:    receiverUserID,
		SenderAccountID:   senderAccountID,
		ReceiverAccountID: receiverAccountID,
		Amount:            pgAmount,
		Currency:          "USD",
		Description:       pgtype.Text{String: description, Valid: description != ""},
		IdempotencyKey:    idempotencyKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer record: %w", err)
	}

	// Authorize the debit transfer (use decrypted account ID for Plaid)
	authorizationID, err := s.authorizeTransfer(ctx, senderAccessToken, decryptedSenderAccountID, senderName, amount, plaid.TRANSFERTYPE_DEBIT, idempotencyKey+"-debit")
	if err != nil {
		_ = s.database.UpdateTransferStatus(ctx, UpdateTransferStatusParams{
			ID:            transfer.ID,
			Status:        TransferStatusFailed,
			FailureReason: pgtype.Text{String: fmt.Sprintf("debit authorization failed: %v", err), Valid: true},
		})
		return nil, fmt.Errorf("debit authorization failed: %w", err)
	}

	// Create the debit transfer (use decrypted account ID for Plaid)
	debitTransferID, err := s.createTransfer(ctx, authorizationID, senderAccessToken, decryptedSenderAccountID, amount, description, plaid.TRANSFERTYPE_DEBIT, idempotencyKey+"-debit")
	if err != nil {
		_ = s.database.UpdateTransferStatus(ctx, UpdateTransferStatusParams{
			ID:            transfer.ID,
			Status:        TransferStatusFailed,
			FailureReason: pgtype.Text{String: fmt.Sprintf("debit transfer creation failed: %v", err), Valid: true},
		})
		return nil, fmt.Errorf("debit transfer creation failed: %w", err)
	}

	// Update transfer record with debit info
	err = s.database.UpdateTransferDebit(ctx, UpdateTransferDebitParams{
		ID:                   transfer.ID,
		DebitAuthorizationID: pgtype.Text{String: authorizationID, Valid: true},
		DebitTransferID:      pgtype.Text{String: debitTransferID, Valid: true},
		Status:               TransferStatusDebitPending,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update transfer with debit info: %w", err)
	}

	// Fetch the updated transfer
	updatedTransfer, err := s.database.GetTransferByID(ctx, transfer.ID)
	if err != nil {
		return nil, err
	}

	// Decrypt account IDs before returning
	return s.decryptTransfer(&updatedTransfer)
}

/*
InitiateCreditLeg authorizes and creates the credit transfer (Plaid Ledger → receiver)
This is called when the debit settles, triggered via webhook
*/
func (s *Service) InitiateCreditLeg(ctx context.Context, transferID int64) error {
	transfer, err := s.database.GetTransferByID(ctx, transferID)
	if err != nil {
		return fmt.Errorf("failed to get transfer: %w", err)
	}

	// Only start the credit leg once, and only from debit_settled.
	if transfer.CreditTransferID.Valid || transfer.Status != TransferStatusDebitSettled {
		logger.Log.WithField("transfer_id", transfer.ID).WithField("status", transfer.Status).Info("Credit leg already started or not ready, skipping")
		return nil
	}

	// Look up by the stored (encrypted) ID; decrypt only for the Plaid call.
	_, receiverAccessToken, err := s.lookupAccount(ctx, transfer.ReceiverAccountID)
	if err != nil {
		return fmt.Errorf("failed to get receiver access token: %w", err)
	}

	receiverAccountID, err := s.encryptor.Decrypt(transfer.ReceiverAccountID)
	if err != nil {
		return fmt.Errorf("failed to decrypt receiver account ID: %w", err)
	}

	receiver, err := s.userService.GetUserByID(ctx, transfer.ReceiverUserID)
	if err != nil {
		return fmt.Errorf("receiver user not found: %w", err)
	}
	receiverName, err := legalName(receiver)
	if err != nil {
		return err
	}

	// Get the transfer amount as string
	amountStr, err := NumericToString(transfer.Amount)
	if err != nil {
		return fmt.Errorf("failed to convert amount: %w", err)
	}

	description := ""
	if transfer.Description.Valid {
		description = transfer.Description.String
	}

	// Authorize the credit transfer
	authorizationID, err := s.authorizeTransfer(ctx, receiverAccessToken, receiverAccountID, receiverName, amountStr, plaid.TRANSFERTYPE_CREDIT, transfer.IdempotencyKey+"-credit")
	if err != nil {
		_ = s.database.UpdateTransferStatus(ctx, UpdateTransferStatusParams{
			ID:            transfer.ID,
			Status:        TransferStatusFailed,
			FailureReason: pgtype.Text{String: fmt.Sprintf("credit authorization failed: %v", err), Valid: true},
		})
		return fmt.Errorf("credit authorization failed: %w", err)
	}

	// Create the credit transfer
	creditTransferID, err := s.createTransfer(ctx, authorizationID, receiverAccessToken, receiverAccountID, amountStr, description, plaid.TRANSFERTYPE_CREDIT, transfer.IdempotencyKey+"-credit")
	if err != nil {
		_ = s.database.UpdateTransferStatus(ctx, UpdateTransferStatusParams{
			ID:            transfer.ID,
			Status:        TransferStatusFailed,
			FailureReason: pgtype.Text{String: fmt.Sprintf("credit transfer creation failed: %v", err), Valid: true},
		})
		return fmt.Errorf("credit transfer creation failed: %w", err)
	}

	// Update transfer record with credit info
	err = s.database.UpdateTransferCredit(ctx, UpdateTransferCreditParams{
		ID:                    transfer.ID,
		CreditAuthorizationID: pgtype.Text{String: authorizationID, Valid: true},
		CreditTransferID:      pgtype.Text{String: creditTransferID, Valid: true},
		Status:                TransferStatusCreditPending,
	})
	if err != nil {
		return fmt.Errorf("failed to update transfer with credit info: %w", err)
	}

	logger.Log.WithField("transfer_id", transfer.ID).Info("Credit leg initiated successfully")
	return nil
}

/*
authorizeTransfer calls the Plaid Transfer Authorization Create API
*/
func (s *Service) authorizeTransfer(ctx context.Context, accessToken, accountID, legalName, amount string, transferType plaid.TransferType, idempotencyKey string) (string, error) {
	request := plaid.NewTransferAuthorizationCreateRequest(
		accessToken,
		accountID,
		transferType,
		plaid.TRANSFERNETWORK_SAME_DAY_ACH,
		amount,
		*plaid.NewTransferAuthorizationUserInRequest(legalName),
	)
	request.SetAchClass(*plaid.ACHCLASS_PPD.Ptr())
	request.SetIdempotencyKey(idempotencyKey)

	response, _, err := s.client.PlaidApi.TransferAuthorizationCreate(ctx).TransferAuthorizationCreateRequest(*request).Execute()
	if err != nil {
		if plaidErr, ok := err.(plaid.GenericOpenAPIError); ok {
			logger.Log.Error("Plaid authorization error: ", string(plaidErr.Body()))
		}
		return "", err
	}

	authorization := response.GetAuthorization()

	// Check if authorization was declined
	decision := authorization.GetDecision()
	if decision == plaid.TRANSFERAUTHORIZATIONDECISION_DECLINED {
		rationale := authorization.GetDecisionRationale()
		return "", fmt.Errorf("transfer authorization declined: %s - %s", rationale.GetCode(), rationale.GetDescription())
	}

	return authorization.GetId(), nil
}

/*
createTransfer calls the Plaid Transfer Create API
*/
func (s *Service) createTransfer(ctx context.Context, authorizationID, accessToken, accountID, amount, description string, transferType plaid.TransferType, idempotencyKey string) (string, error) {
	request := plaid.NewTransferCreateRequest(
		accessToken,
		accountID,
		authorizationID,
		description,
	)
	request.SetAmount(amount)

	response, _, err := s.client.PlaidApi.TransferCreate(ctx).TransferCreateRequest(*request).Execute()
	if err != nil {
		if plaidErr, ok := err.(plaid.GenericOpenAPIError); ok {
			logger.Log.Error("Plaid transfer creation error: ", string(plaidErr.Body()))
		}
		return "", err
	}

	plaidTransfer := response.GetTransfer()
	return plaidTransfer.GetId(), nil
}

/*
legalName returns the account holder's full name as Plaid expects it.
*/
func legalName(u *user.User) (string, error) {
	name := strings.TrimSpace(strings.TrimSpace(u.FirstName.String) + " " + strings.TrimSpace(u.LastName.String))
	if name == "" {
		return "", ErrMissingLegalName
	}
	return name, nil
}

/*
NumericToString converts a pgtype.Numeric to a string representation
*/
func NumericToString(n pgtype.Numeric) (string, error) {
	if !n.Valid {
		return "0", nil
	}

	// Use big.Float for proper decimal representation
	f := new(big.Float)
	f.SetInt(n.Int)
	if n.Exp < 0 {
		divisor := new(big.Float).SetFloat64(1)
		for i := 0; i < int(-n.Exp); i++ {
			divisor.Mul(divisor, new(big.Float).SetFloat64(10))
		}
		f.Quo(f, divisor)
	} else if n.Exp > 0 {
		multiplier := new(big.Float).SetFloat64(1)
		for i := 0; i < int(n.Exp); i++ {
			multiplier.Mul(multiplier, new(big.Float).SetFloat64(10))
		}
		f.Mul(f, multiplier)
	}

	return f.Text('f', 2), nil
}
