package plaid

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// maxWebhookAge is how old a Plaid webhook signature may be before it's rejected.
const maxWebhookAge = 5 * time.Minute

var errInvalidSignature = errors.New("invalid Plaid webhook signature")

/*
KeyFetcher returns the public key Plaid used to sign a webhook, by key ID.
*/
type KeyFetcher func(ctx context.Context, keyID string) (*ecdsa.PublicKey, error)

/*
Verifier checks the Plaid-Verification header on incoming webhooks: an ES256 JWT
signed by Plaid whose request_body_sha256 claim must match the raw body.
See https://plaid.com/docs/api/webhooks/webhook-verification/
*/
type Verifier struct {
	fetch KeyFetcher
	now   func() time.Time

	mu   sync.Mutex
	keys map[string]*ecdsa.PublicKey
}

/*
NewVerifier creates a webhook verifier that fetches signing keys with fetch and caches them.
*/
func NewVerifier(fetch KeyFetcher) *Verifier {
	return &Verifier{fetch: fetch, now: time.Now, keys: map[string]*ecdsa.PublicKey{}}
}

/*
Verify returns nil only if header is a valid, fresh Plaid signature over body.
*/
func (v *Verifier) Verify(ctx context.Context, header string, body []byte) error {
	if header == "" {
		return fmt.Errorf("%w: missing Plaid-Verification header", errInvalidSignature)
	}

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(header, claims, func(token *jwt.Token) (any, error) {
		keyID, _ := token.Header["kid"].(string)
		if keyID == "" {
			return nil, errors.New("missing kid")
		}
		return v.key(ctx, keyID)
	}, jwt.WithValidMethods([]string{"ES256"}), jwt.WithIssuedAt())
	if err != nil {
		return fmt.Errorf("%w: %v", errInvalidSignature, err)
	}

	issuedAt, err := claims.GetIssuedAt()
	if err != nil || issuedAt == nil {
		return fmt.Errorf("%w: missing iat", errInvalidSignature)
	}
	if v.now().Sub(issuedAt.Time) > maxWebhookAge {
		return fmt.Errorf("%w: signature is older than %s", errInvalidSignature, maxWebhookAge)
	}

	wantHash, _ := claims["request_body_sha256"].(string)
	gotHash := sha256.Sum256(body)
	if subtle.ConstantTimeCompare([]byte(wantHash), []byte(hex.EncodeToString(gotHash[:]))) != 1 {
		return fmt.Errorf("%w: body does not match signature", errInvalidSignature)
	}

	return nil
}

func (v *Verifier) key(ctx context.Context, keyID string) (*ecdsa.PublicKey, error) {
	v.mu.Lock()
	key, ok := v.keys[keyID]
	v.mu.Unlock()
	if ok {
		return key, nil
	}

	key, err := v.fetch(ctx, keyID)
	if err != nil {
		return nil, err
	}

	v.mu.Lock()
	v.keys[keyID] = key
	v.mu.Unlock()
	return key, nil
}
