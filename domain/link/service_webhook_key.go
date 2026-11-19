package link

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/plaid/plaid-go/v35/plaid"
)

/*
GetWebhookVerificationKey fetches the public key Plaid used to sign a webhook.
Expired keys are rejected.
*/
func (s *Service) GetWebhookVerificationKey(ctx context.Context, keyID string) (*ecdsa.PublicKey, error) {
	request := plaid.NewWebhookVerificationKeyGetRequest(keyID)
	response, _, err := s.client.PlaidApi.WebhookVerificationKeyGet(ctx).WebhookVerificationKeyGetRequest(*request).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Plaid webhook key: %w", err)
	}

	jwk := response.GetKey()
	if expiredAt, ok := jwk.GetExpiredAtOk(); ok && expiredAt != nil {
		return nil, fmt.Errorf("plaid webhook key %s has expired", keyID)
	}
	if jwk.GetKty() != "EC" || jwk.GetCrv() != "P-256" {
		return nil, fmt.Errorf("unexpected Plaid webhook key type %s/%s", jwk.GetKty(), jwk.GetCrv())
	}

	x, err := base64.RawURLEncoding.DecodeString(jwk.GetX())
	if err != nil {
		return nil, fmt.Errorf("invalid Plaid webhook key x: %w", err)
	}
	y, err := base64.RawURLEncoding.DecodeString(jwk.GetY())
	if err != nil {
		return nil, fmt.Errorf("invalid Plaid webhook key y: %w", err)
	}

	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: new(big.Int).SetBytes(x), Y: new(big.Int).SetBytes(y)}, nil
}
