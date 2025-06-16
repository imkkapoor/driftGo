package stytch

import (
	"fmt"
	"net/http"

	svix "github.com/svix/svix-webhooks/go"
)

func VerifySignature(webhookSecret string, headers http.Header, body []byte) error {
	wh, err := svix.NewWebhook(webhookSecret)
	if err != nil {
		return fmt.Errorf("failed to create webhook instance: %w", err)
	}

	err = wh.Verify(body, headers)
	if err != nil {
		return fmt.Errorf("webhook verification failed: %w", err)
	}

	return nil
}
