package link

import (
	"context"
	"driftGo/api/common/utils"
	"driftGo/domain/user"
	"driftGo/pkg/encryption"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plaid/plaid-go/v35/plaid"
)

var (
	ErrLinkNotFound = errors.New("link not found")
)

/*
Service handles all Plaid link-related operations
*/
type Service struct {
	client      *plaid.APIClient
	userService user.UserInterface
	database    Querier
	encryptor   *encryption.Encryptor
}

/*
NewService creates a new link service with the provided Plaid credentials and database
*/
func NewService(clientID, secret, env string, userService user.UserInterface, db *pgxpool.Pool, encryptionKey string) (*Service, error) {
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

	// Initialize encryptor
	encryptor, err := encryption.NewEncryptor(encryptionKey)
	if err != nil {
		return nil, err
	}

	return &Service{
		client:      client,
		userService: userService,
		database:    New(db),
		encryptor:   encryptor,
	}, nil
}

func (s *Service) CreateLinkToken(ctx context.Context) (*LinkTokenCallResponse, error) {
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: utils.GetUserID(ctx),
	}

	request := plaid.NewLinkTokenCreateRequest(
		"drift",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_CA},
		user,
	)

	request.SetProducts([]plaid.Products{plaid.PRODUCTS_AUTH, plaid.PRODUCTS_IDENTITY})

	linkToken, _, err := s.client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		return nil, err
	}

	return &LinkTokenCallResponse{
		LinkToken:  linkToken.GetLinkToken(),
		Expiration: linkToken.GetExpiration(),
		RequestID:  linkToken.GetRequestId(),
	}, nil
}

func (s *Service) ExchangePublicTokenAndSave(ctx context.Context, publicToken string) error {
	accessTokenResponse, err := s.exchangePublicToken(ctx, publicToken)
	if err != nil {
		return err
	}

	userID := utils.GetUserID(ctx)
	if userID == "" {
		return errors.New("user ID not found in context")
	}

	user, err := s.userService.GetUserByStytchID(ctx, userID)
	if err != nil {
		return err
	}

	institutionID, institutionName, err := s.GetInstitutionMetadata(ctx, accessTokenResponse.AccessToken)
	if err != nil {
		return err
	}

	linkItem, err := s.CreateLinkItem(ctx, user.ID, accessTokenResponse.AccessToken, accessTokenResponse.ItemID, institutionID, institutionName)
	if err != nil {
		return err
	}

	accounts, err := s.GetAccounts(ctx, accessTokenResponse.AccessToken)
	if err != nil {
		return err
	}

	err = s.SaveAccountsFromPlaid(ctx, accounts, linkItem.ID, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetAccounts(ctx context.Context, accessToken string) ([]plaid.AccountBase, error) {
	request := plaid.NewAccountsGetRequest(accessToken)

	response, _, err := s.client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(*request).Execute()
	if err != nil {
		return nil, err
	}

	return response.GetAccounts(), nil
}

func (s *Service) exchangePublicToken(ctx context.Context, publicToken string) (*AccessTokenCallResponse, error) {
	request := plaid.NewItemPublicTokenExchangeRequest(publicToken)

	response, _, err := s.client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*request).Execute()
	if err != nil {
		return nil, err
	}

	return &AccessTokenCallResponse{
		AccessToken: response.GetAccessToken(),
		ItemID:      response.GetItemId(),
	}, nil
}

func (s *Service) GetInstitutionMetadata(ctx context.Context, accessToken string) (institutionID string, institutionName string, err error) {
	institutionID, err = s.getInstitutionID(ctx, accessToken)
	if err != nil {
		return "", "", err
	}

	institutionName, err = s.getInstitutionName(ctx, accessToken)
	if err != nil {
		return "", "", err
	}

	return institutionID, institutionName, nil
}

func (s *Service) getInstitutionID(ctx context.Context, accessToken string) (institutionId string, err error) {
	request := plaid.NewItemGetRequest(accessToken)

	response, _, err := s.client.PlaidApi.ItemGet(ctx).ItemGetRequest(*request).Execute()
	if err != nil {
		return "", err
	}

	return *response.GetItem().InstitutionId.Get(), nil
}

func (s *Service) getInstitutionName(ctx context.Context, accessToken string) (institutionName string, err error) {
	request := plaid.NewItemGetRequest(accessToken)

	response, _, err := s.client.PlaidApi.ItemGet(ctx).ItemGetRequest(*request).Execute()
	if err != nil {
		return "", err
	}

	return *response.GetItem().InstitutionName.Get(), nil
}
