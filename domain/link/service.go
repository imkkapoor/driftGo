package link

import (
	"context"
	"driftGo/api/common/utils"

	"github.com/plaid/plaid-go/v35/plaid"
)

/*
Service handles all Plaid link-related operations
*/
type Service struct {
	client *plaid.APIClient
}

/*
NewService creates a new link service with the provided Plaid credentials
*/
func NewService(clientID, secret, env string) (*Service, error) {
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	configuration.AddDefaultHeader("PLAID-SECRET", secret)
	configuration.UseEnvironment(plaid.Environment(env))

	client := plaid.NewAPIClient(configuration)
	return &Service{client: client}, nil
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
