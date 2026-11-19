package config

import (
	"driftGo/pkg/logger"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	ProjectID          string
	Secret             string
	SignupMagicLinkURL string
	Port               string
	PlaidClientID      string
	PlaidSecret        string
	PlaidEnv           string
	PlaidWebhookURL    string
	AllowedOrigins     []string
	Env                string
	DatabaseURL        string
	WebhookSecret      string
	EncryptionKey      string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		logger.Log.Println("No .env file found, reading from environment variables instead.")
	}

	ProjectID = os.Getenv("STYTCH_PROJECT_ID")
	Secret = os.Getenv("STYTCH_SECRET")
	SignupMagicLinkURL = os.Getenv("STYTCH_SIGNUP_REDIRECT_URL")
	Port = os.Getenv("PORT")
	PlaidClientID = os.Getenv("PLAID_CLIENT_ID")
	PlaidSecret = os.Getenv("PLAID_SECRET")
	PlaidEnv = os.Getenv("PLAID_ENV")
	PlaidWebhookURL = os.Getenv("PLAID_WEBHOOK_URL")
	AllowedOrigins = parseOrigins(os.Getenv("ALLOWED_ORIGINS"))
	Env = os.Getenv("ENV")
	DatabaseURL = os.Getenv("DATABASE_URL")
	WebhookSecret = os.Getenv("STYTCH_WEBHOOK_SECRET")
	encryptionKeyStr := os.Getenv("ENCRYPTION_KEY")

	if ProjectID == "" || Secret == "" {
		logger.Log.Fatal("Missing required environment variables: STYTCH_PROJECT_ID and/or STYTCH_SECRET")
	}

	if PlaidClientID == "" || PlaidSecret == "" {
		logger.Log.Fatal("Missing required environment variables: PLAID_CLIENT_ID and/or PLAID_SECRET")
	}

	if PlaidEnv == "" {
		PlaidEnv = "sandbox" // Default to sandbox environment
	}

	if WebhookSecret == "" {
		logger.Log.Fatal("Missing required environment variable: STYTCH_WEBHOOK_SECRET")
	}

	// Validate encryption key
	if encryptionKeyStr == "" {
		logger.Log.Fatal("Missing required environment variable: ENCRYPTION_KEY")
	}

	EncryptionKey = encryptionKeyStr
	if len(EncryptionKey) < 32 {
		logger.Log.Fatal("ENCRYPTION_KEY must be at least 32 characters long")
	}
}

/*
SecureCookies reports whether cookies should be marked Secure (HTTPS only).
They are in production; local development runs over plain HTTP.
*/
func SecureCookies() bool {
	return Env == "production"
}

/*
parseOrigins splits a comma-separated ALLOWED_ORIGINS value, defaulting to the
local web app.
*/
func parseOrigins(raw string) []string {
	var origins []string
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	if len(origins) == 0 {
		return []string{"http://localhost:3000"}
	}
	return origins
}
