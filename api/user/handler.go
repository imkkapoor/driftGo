package user

import (
	"driftGo/api/common/errors"
	"driftGo/api/common/utils"
	"driftGo/domain/user"
	"driftGo/pkg/logger"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

/*
SetupRoutes sets up the routes for the user package.
It registers the handlers for the various user-related endpoints.
*/
func SetupRoutes(r chi.Router, service *user.Service) {
	handler := &Handler{service: service}
	r.Get("/", handler.getUser)
}

/*
Handler holds the service instance
*/
type Handler struct {
	service *user.Service
}

/*
getUser handles the request to get the authenticated user's information.
This endpoint retrieves the user details for the currently authenticated user.
*/
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserID(r.Context())
	if userID == 0 {
		logger.Log.Error("User ID not found in context")
		errors.UnauthorizedErrorHandler(w, "User not authenticated")
		return
	}

	dbUser, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		logger.Log.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		errors.RequestErrorHandler(w, errors.NewErrorWithCode(http.StatusNotFound, "User not found", errors.ErrCodeNotFound))
		return
	}

	response := GetUserResponse{
		ID:        dbUser.ID,
		FirstName: dbUser.FirstName.String,
		LastName:  dbUser.LastName.String,
		Email:     dbUser.Email,
		Status:    string(dbUser.Status),
		CreatedAt: dbUser.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: dbUser.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.WithError(err).Error("Failed to encode user response")
		errors.InternalErrorHandler(w)
		return
	}
}
