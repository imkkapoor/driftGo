package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/stytchauth/stytch-go/v15/stytch/consumer/stytchapi"
	emailAPI "github.com/stytchauth/stytch-go/v15/stytch/consumer/magiclinks/email"
)

// StytchClient is the client for the Stytch API
var StytchClient *stytchapi.API

// Initialize the client
func init() {
	// Load .env file
	envPath := filepath.Join(".env")
	err := godotenv.Load(envPath)
	if err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	projectID := os.Getenv("STYTCH_PROJECT_ID")
	secret := os.Getenv("STYTCH_SECRET")
	
	// Check if environment variables are set
	if projectID == "" || secret == "" {
		fmt.Println("ERROR: STYTCH_PROJECT_ID and/or STYTCH_SECRET environment variables are not set")
		fmt.Println("Please set these environment variables before running the application")
		fmt.Println("Example: export STYTCH_PROJECT_ID=project-test-xxxx")
		fmt.Println("         export STYTCH_SECRET=secret-test-xxxx")
		os.Exit(1)
	}
	
	// Debug info to verify credentials
	fmt.Printf("Initializing Stytch with Project ID: %s (length: %d)\n", 
		maskString(projectID), len(projectID))
	
	// Create a new client with test environment
	client, err := stytchapi.NewClient(projectID, secret)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Stytch client: %v", err))
	}
	StytchClient = client
}

// Helper function to mask a string for logging
func maskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}

// SendMagicLink sends a magic link to the provided email address
func SendMagicLink(email string) error {
	params := &emailAPI.LoginOrCreateParams{
		Email: email,
	}

	_, err := StytchClient.MagicLinks.Email.LoginOrCreate(
		context.Background(),
		params,
	)
	return err
}

type CreateUserRequest struct {
	Email string `json:"email"`
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := SendMagicLink(req.Email)
	if err != nil {
		http.Error(w, "Failed to send magic link", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Magic link sent"))
}
