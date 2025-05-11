package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	ProjectID string
	Secret    string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from environment variables instead.")
	}

	ProjectID = os.Getenv("STYTCH_PROJECT_ID")
	Secret = os.Getenv("STYTCH_SECRET")

	if ProjectID == "" || Secret == "" {
		log.Fatal("Missing required environment variables: STYTCH_PROJECT_ID and/or STYTCH_SECRET")
	}
}
