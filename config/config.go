package config

import (
	"log"
	"os"

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
	Env                string
	DatabaseURL        string
	WebhookSecret      string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from environment variables instead.")
	}

	ProjectID = os.Getenv("STYTCH_PROJECT_ID")
	Secret = os.Getenv("STYTCH_SECRET")
	SignupMagicLinkURL = os.Getenv("STYTCH_SIGNUP_REDIRECT_URL")
	Port = os.Getenv("PORT")
	PlaidClientID = os.Getenv("PLAID_CLIENT_ID")
	PlaidSecret = os.Getenv("PLAID_SECRET")
	PlaidEnv = os.Getenv("PLAID_ENV")
	Env = os.Getenv("ENV")
	DatabaseURL = os.Getenv("DATABASE_URL")
	WebhookSecret = os.Getenv("STYTCH_WEBHOOK_SECRET")

	if ProjectID == "" || Secret == "" {
		log.Fatal("Missing required environment variables: STYTCH_PROJECT_ID and/or STYTCH_SECRET")
	}

	if PlaidClientID == "" || PlaidSecret == "" {
		log.Fatal("Missing required environment variables: PLAID_CLIENT_ID and/or PLAID_SECRET")
	}

	if PlaidEnv == "" {
		PlaidEnv = "sandbox" // Default to sandbox environment
	}

	if WebhookSecret == "" {
		log.Fatal("Missing required environment variable: STYTCH_WEBHOOK_SECRET")
	}
}
