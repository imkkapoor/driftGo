package link

import (
	"context"
	"driftGo/api/common/utils"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/plaid/plaid-go/v35/plaid"
)

/*
returns one
*/
func (s *Service) CreateLinkAccount(ctx context.Context, plaidAccountID string, itemID, userID int64, name, officialName, mask, subtype, accountType string) (*LinkAccount, error) {
	encryptedPlaidAccountID, err := s.encryptor.Encrypt(plaidAccountID)
	if err != nil {
		return nil, err
	}

	params := CreateLinkAccountParams{
		PlaidAccountID: encryptedPlaidAccountID,
		ItemID:         itemID,
		UserID:         userID,
		Name:           pgtype.Text{String: name, Valid: name != ""},
		OfficialName:   pgtype.Text{String: officialName, Valid: officialName != ""},
		Mask:           pgtype.Text{String: mask, Valid: mask != ""},
		Subtype:        pgtype.Text{String: subtype, Valid: subtype != ""},
		Type:           pgtype.Text{String: accountType, Valid: accountType != ""},
	}

	linkAccount, err := s.database.CreateLinkAccount(ctx, params)
	if err != nil {
		return nil, err
	}

	decryptedPlaidAccountID, err := s.encryptor.Decrypt(linkAccount.PlaidAccountID)
	if err != nil {
		return nil, err
	}
	linkAccount.PlaidAccountID = decryptedPlaidAccountID

	return &linkAccount, nil
}

/*
returns one
*/
func (s *Service) GetLinkAccountByID(ctx context.Context, ID int64) (*LinkAccount, error) {
	linkAccount, err := s.database.GetLinkAccountByID(ctx, ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	decryptedPlaidAccountID, err := s.encryptor.Decrypt(linkAccount.PlaidAccountID)
	if err != nil {
		return nil, err
	}
	linkAccount.PlaidAccountID = decryptedPlaidAccountID

	return &linkAccount, nil
}

/*
returns one
*/
func (s *Service) GetLinkAccountByPlaidAccountID(ctx context.Context, plaidAccountID string) (*LinkAccount, error) {
	// Encrypt the search parameter since the database stores encrypted values
	encryptedPlaidAccountID, err := s.encryptor.Encrypt(plaidAccountID)
	if err != nil {
		return nil, err
	}

	linkAccount, err := s.database.GetLinkAccountByPlaidAccountID(ctx, encryptedPlaidAccountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	decryptedPlaidAccountID, err := s.encryptor.Decrypt(linkAccount.PlaidAccountID)
	if err != nil {
		return nil, err
	}
	linkAccount.PlaidAccountID = decryptedPlaidAccountID

	return &linkAccount, nil
}

/*
returns many
*/
func (s *Service) GetLinkAccountsByItemID(ctx context.Context, itemID int64) ([]LinkAccount, error) {
	linkAccounts, err := s.database.GetLinkAccountsByItemID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	for i := range linkAccounts {
		decryptedPlaidAccountID, err := s.encryptor.Decrypt(linkAccounts[i].PlaidAccountID)
		if err != nil {
			return nil, err
		}
		linkAccounts[i].PlaidAccountID = decryptedPlaidAccountID
	}

	return linkAccounts, nil
}

/*
returns many
*/
func (s *Service) GetLinkAccountsByUser(ctx context.Context) ([]LinkAccount, error) {
	userID := utils.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}

	user, err := s.userService.GetUserByStytchID(ctx, userID)
	if err != nil {
		return nil, err
	}

	linkAccounts, err := s.database.GetLinkAccountsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	for i := range linkAccounts {
		decryptedPlaidAccountID, err := s.encryptor.Decrypt(linkAccounts[i].PlaidAccountID)
		if err != nil {
			return nil, err
		}
		linkAccounts[i].PlaidAccountID = decryptedPlaidAccountID
	}

	return linkAccounts, nil
}

/*
exec
*/
func (s *Service) DeleteLinkAccount(ctx context.Context, id int64) error {
	return s.database.DeleteLinkAccount(ctx, id)
}

/*
exec
*/
func (s *Service) DeleteLinkAccountByPlaidAccountID(ctx context.Context, plaidAccountID string) error {
	encryptedPlaidAccountID, err := s.encryptor.Encrypt(plaidAccountID)
	if err != nil {
		return err
	}

	return s.database.DeleteLinkAccountByPlaidAccountID(ctx, encryptedPlaidAccountID)
}

/*
exec
*/
func (s *Service) DeleteLinkAccountsByItemID(ctx context.Context, itemID int64) error {
	return s.database.DeleteLinkAccountsByItemID(ctx, itemID)
}

/*
exec
*/
func (s *Service) SaveAccountsFromPlaid(ctx context.Context, accounts []plaid.AccountBase, itemID, userID int64) error {
	for _, account := range accounts {
		name := account.GetName()

		officialName := ""
		if account.OfficialName.IsSet() {
			officialName = *account.OfficialName.Get()
		}

		mask := ""
		if account.Mask.IsSet() {
			mask = *account.Mask.Get()
		}

		subtype := ""
		if account.Subtype.IsSet() {
			subtype = string(*account.Subtype.Get())
		}

		accountType := string(account.GetType())

		_, err := s.CreateLinkAccount(ctx, account.GetAccountId(), itemID, userID, name, officialName, mask, subtype, accountType)
		if err != nil {
			return err
		}
	}

	return nil
}
