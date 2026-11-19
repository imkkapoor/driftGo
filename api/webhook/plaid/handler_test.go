package plaid

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type countingProcessor struct{ calls int }

func (c *countingProcessor) ProcessTransferWebhook(context.Context) error {
	c.calls++
	return nil
}

const body = `{"webhook_type":"TRANSFER","webhook_code":"TRANSFER_EVENTS_UPDATE"}`

func sign(t *testing.T, key *ecdsa.PrivateKey, kid string, payload string, iat time.Time) string {
	t.Helper()
	sum := sha256.Sum256([]byte(payload))
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iat":                 iat.Unix(),
		"request_body_sha256": hex.EncodeToString(sum[:]),
	})
	tok.Header["kid"] = kid
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestPlaidWebhookSignature(t *testing.T) {
	plaidKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	attackerKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	fetch := func(_ context.Context, kid string) (*ecdsa.PublicKey, error) {
		if kid != "plaid-kid" {
			return nil, errInvalidSignature
		}
		return &plaidKey.PublicKey, nil
	}
	now := time.Now()

	hmac := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"iat": now.Unix()})
	hmac.Header["kid"] = "plaid-kid"
	hmacToken, _ := hmac.SignedString([]byte("guess"))

	cases := []struct {
		name      string
		header    string
		body      string
		wantCode  int
		wantCalls int
	}{
		{"valid signature", sign(t, plaidKey, "plaid-kid", body, now), body, http.StatusOK, 1},
		{"no header", "", body, http.StatusUnauthorized, 0},
		{"signed by someone else", sign(t, attackerKey, "plaid-kid", body, now), body, http.StatusUnauthorized, 0},
		{"body tampered", sign(t, plaidKey, "plaid-kid", body, now), strings.Replace(body, "TRANSFER", "ITEM", 1), http.StatusUnauthorized, 0},
		{"replayed old signature", sign(t, plaidKey, "plaid-kid", body, now.Add(-10*time.Minute)), body, http.StatusUnauthorized, 0},
		{"unknown key id", sign(t, plaidKey, "other-kid", body, now), body, http.StatusUnauthorized, 0},
		{"wrong algorithm", hmacToken, body, http.StatusUnauthorized, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			proc := &countingProcessor{}
			h := NewHandler(proc, NewVerifier(fetch))

			req := httptest.NewRequest(http.MethodPost, "/webhook/plaid", strings.NewReader(tc.body))
			if tc.header != "" {
				req.Header.Set("Plaid-Verification", tc.header)
			}
			rec := httptest.NewRecorder()
			h.HandleWebhook(rec, req)

			if rec.Code != tc.wantCode || proc.calls != tc.wantCalls {
				t.Fatalf("got status %d and %d syncs, want %d and %d", rec.Code, proc.calls, tc.wantCode, tc.wantCalls)
			}
		})
	}
}
