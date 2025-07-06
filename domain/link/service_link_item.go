package link

import (
	"context"
	"driftGo/api/common/utils"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

/*
returns one
*/
func (s *Service) CreateLinkItem(ctx context.Context, userID int64, accessToken, itemID, institutionID, institutionName string) (*LinkItem, error) {
	encryptedAccessToken, err := s.encryptor.Encrypt(accessToken)
	if err != nil {
		return nil, err
	}

	params := CreateLinkItemParams{
		UserID:          userID,
		AccessToken:     encryptedAccessToken,
		ItemID:          itemID,
		InstitutionID:   pgtype.Text{String: institutionID, Valid: institutionID != ""},
		InstitutionName: pgtype.Text{String: institutionName, Valid: institutionName != ""},
	}

	linkItem, err := s.database.CreateLinkItem(ctx, params)
	if err != nil {
		return nil, err
	}

	decryptedAccessToken, err := s.encryptor.Decrypt(linkItem.AccessToken)
	if err != nil {
		return nil, err
	}
	linkItem.AccessToken = decryptedAccessToken

	return &linkItem, nil
}

/*
returns one
*/
func (s *Service) GetLinkItemByID(ctx context.Context, ID int64) (*LinkItem, error) {
	linkItem, err := s.database.GetLinkItemByID(ctx, ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	return &linkItem, nil
}

/*
returns one
*/
func (s *Service) GetLinkItemByItemID(ctx context.Context, itemID string) (*LinkItem, error) {
	linkItem, err := s.database.GetLinkItemByItemID(ctx, itemID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	return &linkItem, nil
}

/*
returns many
*/
func (s *Service) GetLinkItemsByUser(ctx context.Context) ([]LinkItem, error) {
	userID := utils.GetUserID(ctx)
	if userID == 0 {
		return nil, errors.New("user ID not found in context")
	}

	linkItems, err := s.database.GetLinkItemsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range linkItems {
		decryptedAccessToken, err := s.encryptor.Decrypt(linkItems[i].AccessToken)
		if err != nil {
			return nil, err
		}
		linkItems[i].AccessToken = decryptedAccessToken
	}

	return linkItems, nil
}

/*
returns one
*/
func (s *Service) GetAccessTokenByAccountID(ctx context.Context, accountID string) (string, error) {
	accessToken, err := s.database.GetAccessTokenByAccountID(ctx, accountID)
	if err != nil {
		return "", err
	}

	decryptedAccessToken, err := s.encryptor.Decrypt(accessToken)
	if err != nil {
		return "", err
	}

	return decryptedAccessToken, nil
}

/*
exec
*/
func (s *Service) DeleteLinkItemByID(ctx context.Context, ID int64) error {
	return s.database.DeleteLinkItem(ctx, ID)
}

/*
exec
*/
func (s *Service) DeleteLinkItemByItemID(ctx context.Context, itemID string) error {
	return s.database.DeleteLinkItemByItemID(ctx, itemID)
}
