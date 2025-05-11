package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"driftGo/config"
	
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/stytchapi"
)

var client *stytchapi.API

func init() {
	var err error
	client, err = stytchapi.NewClient(config.ProjectID, config.Secret)
	if err != nil {
		log.Fatalf("failed to initialize Stytch client: %v", err)
	}
}


func SendMagicLinkCall(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is POST.
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON body.
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	
	if req.Email == "" {
		http.Error(w, "missing email field", http.StatusBadRequest)
		return
	}

	// Call the service function to send the magic link.
	resp, err := SendMagicLink(r.Context(), req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}
