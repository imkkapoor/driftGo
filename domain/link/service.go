package link

import (
	"context"
	"driftGo/api/common/utils"
	"driftGo/domain/user"
	"errors"

	"github.com/plaid/plaid-go/v35/plaid"
)

/*
Service handles all Plaid link-related operations
*/
type Service struct {
	client      *plaid.APIClient
	userService user.UserInterface
}

/*
NewService creates a new link service with the provided Plaid credentials
*/
func NewService(clientID, secret, env string, userService user.UserInterface) (*Service, error) {
	var plaidEnv plaid.Environment
	switch env {
	case "sandbox":
		plaidEnv = plaid.Sandbox
	case "production":
		plaidEnv = plaid.Production
	default:
		return nil, errors.New(" Invalid Plaid environment: " + env)
	}

	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	configuration.AddDefaultHeader("PLAID-SECRET", secret)
	configuration.UseEnvironment(plaidEnv)

	client := plaid.NewAPIClient(configuration)
	return &Service{client: client, userService: userService}, nil
}

/*
CreateLinkToken creates a new link token for the Plaid Link flow.
This is used to initialize the Plaid Link interface for a user.
The link token is required to start the Plaid Link flow.
*/
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

/*
CreateLinkTokenForUser creates a link token for a specific user by their Stytch ID.
This demonstrates how to use the user service within the link service.
*/
func (s *Service) CreateLinkTokenForUser(ctx context.Context, stytchUserID string) (*LinkTokenCallResponse, error) {
	// Get user from user service
	user, err := s.userService.GetUserByStytchID(ctx, stytchUserID)
	if err != nil {
		return nil, err
	}

	// Use the user's UUID for Plaid
	plaidUser := plaid.LinkTokenCreateRequestUser{
		ClientUserId: user.Uuid.String(),
	}

	request := plaid.NewLinkTokenCreateRequest(
		"drift",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_CA},
		plaidUser,
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

/*
ExchangePublicToken exchanges a public token for an access token.
This is used after a user successfully links their bank account through Plaid Link.
The public token is exchanged for an access token that can be used to access the user's bank account data.
*/
func (s *Service) ExchangePublicToken(ctx context.Context, publicToken string) (*AccessTokenCallResponse, error) {
	request := plaid.NewItemPublicTokenExchangeRequest(publicToken)

	response, _, err := s.client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*request).Execute()
	if err != nil {
		return nil, err
	}

	return &AccessTokenCallResponse{
		AccessToken: response.GetAccessToken(),
		ItemID:      response.GetItemId(),
		RequestID:   response.GetRequestId(),
	}, nil
}

/*
GetInstitutionMetadata gets the institution ID and name for a given access token.
*/
func (s *Service) GetInstitutionMetadata(ctx context.Context, accessToken string) (institutionId string, institutionName string, err error) {
	institutionId, err = s.GetInstitutionId(ctx, accessToken)
	if err != nil {
		return "", "", err
	}

	institutionName, err = s.GetInstitutionName(ctx, accessToken)
	if err != nil {
		return "", "", err
	}

	return institutionId, institutionName, nil
}

/*
GetInstitutionId gets the institution ID for a given access token.
*/
func (s *Service) GetInstitutionId(ctx context.Context, accessToken string) (institutionId string, err error) {
	request := plaid.NewItemGetRequest(accessToken)

	response, _, err := s.client.PlaidApi.ItemGet(ctx).ItemGetRequest(*request).Execute()
	if err != nil {
		return "", err
	}

	return *response.GetItem().InstitutionId.Get(), nil
}

/*
GetInstitutionName gets the institution name for a given access token.
*/
func (s *Service) GetInstitutionName(ctx context.Context, accessToken string) (institutionName string, err error) {
	request := plaid.NewItemGetRequest(accessToken)

	response, _, err := s.client.PlaidApi.ItemGet(ctx).ItemGetRequest(*request).Execute()
	if err != nil {
		return "", err
	}

	return *response.GetItem().InstitutionName.Get(), nil
}

/*
GetAccounts gets the accounts for a given access token.
*/
func (s *Service) GetAccounts(ctx context.Context, accessToken string) ([]plaid.AccountBase, error) {
	request := plaid.NewAccountsGetRequest(accessToken)

	response, _, err := s.client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(*request).Execute()
	if err != nil {
		return nil, err
	}

	return response.GetAccounts(), nil
}
