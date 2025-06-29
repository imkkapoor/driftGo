package stytch

import (
	"driftGo/api/common/errors"
	"driftGo/domain/user"
	"encoding/json"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

/*
Handler handles Stytch webhook events
*/
type Handler struct {
	userService *user.Service
	secret      string
}

/*
NewHandler creates a new Stytch webhook handler
*/
func NewHandler(userService *user.Service, secret string) *Handler {
	return &Handler{
		userService: userService,
		secret:      secret,
	}
}

/*
This function is used to handle the incoming Stytch webhook events.
It reads the request body, verifies the webhook signature, parses the webhook event, and processes the event based on the action.

Events supported:
- CREATE: Create a new user
- UPDATE: Update an existing user
- DELETE: Delete a user
*/
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Error("Failed to read webhook request body")
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	if err := VerifySignature(h.secret, r.Header, body); err != nil {
		log.WithError(err).Error("Invalid webhook signature")
		errors.RequestErrorHandler(w, errors.NewErrorWithCode(http.StatusUnauthorized, "Invalid webhook signature", errors.ErrCodeAuthentication))
		return
	}

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.WithError(err).Error("Failed to parse webhook event")
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	log.Info("Incoming stytch webhook event with action: ", event.Action, " and source: ", event.Source)

	switch event.Action {
	case "CREATE":
		if event.User == nil {
			log.Error("User data is missing in CREATE event")
			errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
			return
		}

		existingUser, err := h.userService.GetUserByStytchID(r.Context(), event.StytchUserID)
		if err != nil && err != user.ErrUserNotFound {
			log.WithError(err).Error("Failed to check if user exists")
			errors.InternalErrorHandler(w)
			return
		}
		if existingUser != nil {
			log.WithField("stytch_user_id", event.StytchUserID).Info("User already exists, skipping CREATE event")
			w.WriteHeader(http.StatusOK)
			return
		}

		var email string
		if len(event.User.Emails) > 0 {
			email = event.User.Emails[0].Email
		}

		_, err = h.userService.CreateUser(
			r.Context(),
			event.StytchUserID,
			event.User.Name.FirstName,
			event.User.Name.LastName,
			email,
			event.User.Status,
		)
		if err != nil {
			log.WithError(err).Error("Failed to create user")
			errors.InternalErrorHandler(w)
			return
		}

	case "UPDATE":
		if event.User == nil {
			log.Error("User data is missing in UPDATE event")
			errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
			return
		}

		var email string
		if len(event.User.Emails) > 0 {
			email = event.User.Emails[0].Email
		}

		_, err := h.userService.UpdateUser(
			r.Context(),
			event.StytchUserID,
			event.User.Name.FirstName,
			event.User.Name.LastName,
			email,
			event.User.Status,
		)
		if err != nil {
			log.WithError(err).Error("Failed to update user")
			errors.InternalErrorHandler(w)
			return
		}

	case "DELETE":
		if err := h.userService.DeleteUser(r.Context(), event.StytchUserID); err != nil {
			log.WithError(err).Error("Failed to process user deletion")
			errors.InternalErrorHandler(w)
			return
		}

	default:
		log.WithField("action", event.Action).Info("Received unknown event type")
	}

	w.WriteHeader(http.StatusOK)
}
