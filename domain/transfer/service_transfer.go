package transfer

import (
	"context"
	"driftGo/api/common/utils"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetTransferByID fetches a transfer by ID and verifies the requesting user is the sender or receiver
func (s *Service) GetTransferByID(ctx context.Context, id int64) (*Transfer, error) {
	transfer, err := s.database.GetTransferByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransferNotFound
		}
		return nil, err
	}

	userID := utils.GetUserID(ctx)
	if userID != transfer.SenderUserID && userID != transfer.ReceiverUserID {
		return nil, ErrTransferNotFound
	}

	return s.decryptTransfer(&transfer)
}

// GetTransfersByUserID fetches all transfers where the authenticated user is sender or receiver
func (s *Service) GetTransfersByUserID(ctx context.Context) ([]Transfer, error) {
	userID := utils.GetUserID(ctx)
	if userID == 0 {
		return nil, errors.New("user ID not found in context")
	}

	transfers, err := s.database.GetTransfersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.decryptTransfers(transfers)
}

// GetTransferByDebitTransferID looks up a transfer by its Plaid debit transfer ID (for webhook processing)
func (s *Service) GetTransferByDebitTransferID(ctx context.Context, debitTransferID string) (*Transfer, error) {
	transfer, err := s.database.GetTransferByDebitTransferID(ctx, pgtype.Text{String: debitTransferID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransferNotFound
		}
		return nil, err
	}
	return &transfer, nil
}

// GetTransferByCreditTransferID looks up a transfer by its Plaid credit transfer ID (for webhook processing)
func (s *Service) GetTransferByCreditTransferID(ctx context.Context, creditTransferID string) (*Transfer, error) {
	transfer, err := s.database.GetTransferByCreditTransferID(ctx, pgtype.Text{String: creditTransferID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransferNotFound
		}
		return nil, err
	}
	return &transfer, nil
}

// UpdateTransferStatus updates the transfer status and optional failure reason
func (s *Service) UpdateTransferStatus(ctx context.Context, id int64, status TransferStatus, failureReason string) error {
	return s.database.UpdateTransferStatus(ctx, UpdateTransferStatusParams{
		ID:            id,
		Status:        status,
		FailureReason: pgtype.Text{String: failureReason, Valid: failureReason != ""},
	})
}

// MarkTransferCompleted marks a transfer as completed with a timestamp
func (s *Service) MarkTransferCompleted(ctx context.Context, id int64) error {
	return s.database.UpdateTransferCompleted(ctx, id)
}

// decryptTransfer decrypts the encrypted fields of a transfer record
func (s *Service) decryptTransfer(transfer *Transfer) (*Transfer, error) {
	decryptedSenderAccountID, err := s.encryptor.Decrypt(transfer.SenderAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt sender account ID: %w", err)
	}
	transfer.SenderAccountID = decryptedSenderAccountID

	decryptedReceiverAccountID, err := s.encryptor.Decrypt(transfer.ReceiverAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt receiver account ID: %w", err)
	}
	transfer.ReceiverAccountID = decryptedReceiverAccountID

	return transfer, nil
}

// decryptTransfers decrypts the encrypted fields of multiple transfer records
func (s *Service) decryptTransfers(transfers []Transfer) ([]Transfer, error) {
	for i := range transfers {
		decryptedSenderAccountID, err := s.encryptor.Decrypt(transfers[i].SenderAccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt sender account ID: %w", err)
		}
		transfers[i].SenderAccountID = decryptedSenderAccountID

		decryptedReceiverAccountID, err := s.encryptor.Decrypt(transfers[i].ReceiverAccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt receiver account ID: %w", err)
		}
		transfers[i].ReceiverAccountID = decryptedReceiverAccountID
	}

	return transfers, nil
}
