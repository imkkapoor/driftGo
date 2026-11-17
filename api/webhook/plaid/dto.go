package plaid

type PlaidWebhookEvent struct {
	WebhookType string `json:"webhook_type"`
	WebhookCode string `json:"webhook_code"`
	Environment string `json:"environment"`
}
