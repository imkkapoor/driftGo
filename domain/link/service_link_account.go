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
func (s *Service) CreateLinkAccount(ctx context.Context, accountID string, itemID, userID int64, name, officialName, mask, subtype, accountType string) (*LinkAccount, error) {
	encryptedAccountID, err := s.encryptor.Encrypt(accountID)
	if err != nil {
		return nil, err
	}

	params := CreateLinkAccountParams{
		AccountID:    encryptedAccountID,
		ItemID:       itemID,
		UserID:       userID,
		Name:         pgtype.Text{String: name, Valid: name != ""},
		OfficialName: pgtype.Text{String: officialName, Valid: officialName != ""},
		Mask:         pgtype.Text{String: mask, Valid: mask != ""},
		Subtype:      pgtype.Text{String: subtype, Valid: subtype != ""},
		Type:         pgtype.Text{String: accountType, Valid: accountType != ""},
	}

	linkAccount, err := s.database.CreateLinkAccount(ctx, params)
	if err != nil {
		return nil, err
	}

	decryptedAccountID, err := s.encryptor.Decrypt(linkAccount.AccountID)
	if err != nil {
		return nil, err
	}
	linkAccount.AccountID = decryptedAccountID

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

	decryptedAccountID, err := s.encryptor.Decrypt(linkAccount.AccountID)
	if err != nil {
		return nil, err
	}
	linkAccount.AccountID = decryptedAccountID

	return &linkAccount, nil
}

/*
returns one
*/
func (s *Service) GetLinkAccountByAccountID(ctx context.Context, accountID string) (*LinkAccount, error) {
	// Encrypt the search parameter since the database stores encrypted values
	encryptedAccountID, err := s.encryptor.Encrypt(accountID)
	if err != nil {
		return nil, err
	}

	linkAccount, err := s.database.GetLinkAccountByAccountID(ctx, encryptedAccountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	decryptedAccountID, err := s.encryptor.Decrypt(linkAccount.AccountID)
	if err != nil {
		return nil, err
	}
	linkAccount.AccountID = decryptedAccountID

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
		decryptedAccountID, err := s.encryptor.Decrypt(linkAccounts[i].AccountID)
		if err != nil {
			return nil, err
		}
		linkAccounts[i].AccountID = decryptedAccountID
	}

	return linkAccounts, nil
}

/*
returns many
*/
func (s *Service) GetLinkAccountsByUser(ctx context.Context) ([]LinkAccount, error) {
	userID := utils.GetUserID(ctx)
	if userID == 0 {
		return nil, errors.New("user ID not found in context")
	}

	linkAccounts, err := s.database.GetLinkAccountsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range linkAccounts {
		decryptedAccountID, err := s.encryptor.Decrypt(linkAccounts[i].AccountID)
		if err != nil {
			return nil, err
		}
		linkAccounts[i].AccountID = decryptedAccountID
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
func (s *Service) DeleteLinkAccountByAccountID(ctx context.Context, accountID string) error {
	encryptedAccountID, err := s.encryptor.Encrypt(accountID)
	if err != nil {
		return err
	}

	return s.database.DeleteLinkAccountByAccountID(ctx, encryptedAccountID)
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
