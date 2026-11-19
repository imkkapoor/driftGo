package transfer

import (
	"driftGo/api/common/errors"
	"driftGo/api/common/validation"
	"driftGo/domain/transfer"
	"driftGo/pkg/logger"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

/*
Handler holds the service instance for handling transfer-related operations
*/
type Handler struct {
	service *transfer.Service
}

/*
SetupRoutes sets up the routes for the transfer package.
It registers the handlers for the various transfer-related endpoints.
*/
func SetupRoutes(r chi.Router, service *transfer.Service) {
	handler := &Handler{service: service}
	r.Post("/create", handler.initiateTransfer)
	r.Get("/list", handler.listTransfers)
	r.Get("/{id}", handler.getTransfer)
}

/*
initiateTransfer handles the request to initiate a new transfer between two users.
The sender must be the authenticated user. Funds are debited from the sender's linked
bank account and credited to the receiver's linked bank account via Plaid Ledger.
*/
func (h *Handler) initiateTransfer(w http.ResponseWriter, r *http.Request) {
	var req InitiateTransferRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.WithError(err).Error("Failed to decode initiate transfer request")
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	if !validation.ValidateRequest(w, req) {
		return
	}

	result, err := h.service.InitiateTransfer(
		r.Context(),
		req.SenderAccountID,
		req.ReceiverUserID,
		req.ReceiverAccountID,
		req.Amount,
		req.Description,
	)
	if err != nil {
		switch err {
		case transfer.ErrSelfTransfer:
			errors.RequestErrorHandler(w, errors.NewErrorWithCode(http.StatusBadRequest, "Cannot transfer to yourself", errors.ErrCodeInvalidRequest))
		case transfer.ErrInvalidAmount:
			errors.RequestErrorHandler(w, errors.NewErrorWithCode(http.StatusBadRequest, "Invalid transfer amount", errors.ErrCodeInvalidRequest))
		case transfer.ErrAccountNotFound, transfer.ErrAccountNotOwned:
			errors.RequestErrorHandler(w, errors.NewErrorWithCode(http.StatusBadRequest, "Sender account not found", errors.ErrCodeInvalidRequest))
		case transfer.ErrMissingLegalName:
			errors.RequestErrorHandler(w, errors.NewErrorWithCode(http.StatusBadRequest, "Sender and receiver must have a first and last name on file", errors.ErrCodeValidationError))
		case transfer.ErrReceiverNoAccount:
			errors.RequestErrorHandler(w, errors.NewErrorWithCode(http.StatusBadRequest, "Receiver has no linked account", errors.ErrCodeInvalidRequest))
		default:
			logger.Log.WithError(err).Error("Failed to initiate transfer")
			errors.InternalErrorHandler(w)
		}
		return
	}

	response := toTransferResponse(result)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.WithError(err).Error("Failed to encode transfer response")
		errors.InternalErrorHandler(w)
		return
	}
}

/*
getTransfer handles the request to get a specific transfer by ID.
The requesting user must be the sender or receiver of the transfer.
*/
func (h *Handler) getTransfer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errors.RequestErrorHandler(w, errors.NewErrorWithCode(http.StatusBadRequest, "Invalid transfer ID", errors.ErrCodeInvalidRequest))
		return
	}

	result, err := h.service.GetTransferByID(r.Context(), id)
	if err != nil {
		if err == transfer.ErrTransferNotFound {
			errors.NotFoundErrorHandler(w, "Transfer not found")
			return
		}
		logger.Log.WithError(err).Error("Failed to get transfer")
		errors.InternalErrorHandler(w)
		return
	}

	response := toTransferResponse(result)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.WithError(err).Error("Failed to encode transfer response")
		errors.InternalErrorHandler(w)
		return
	}
}

/*
listTransfers handles the request to list all transfers for the authenticated user.
Returns transfers where the user is either the sender or receiver.
*/
func (h *Handler) listTransfers(w http.ResponseWriter, r *http.Request) {
	transfers, err := h.service.GetTransfersByUserID(r.Context())
	if err != nil {
		logger.Log.WithError(err).Error("Failed to list transfers")
		errors.InternalErrorHandler(w)
		return
	}

	transferResponses := make([]TransferResponse, len(transfers))
	for i, t := range transfers {
		transferResponses[i] = *toTransferResponse(&t)
	}

	response := TransferListResponse{
		Transfers: transferResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.WithError(err).Error("Failed to encode transfers response")
		errors.InternalErrorHandler(w)
		return
	}
}

/*
toTransferResponse converts a domain Transfer to an API TransferResponse
*/
func toTransferResponse(t *transfer.Transfer) *TransferResponse {
	amountStr, _ := transfer.NumericToString(t.Amount)

	resp := &TransferResponse{
		ID:                t.ID,
		SenderUserID:      t.SenderUserID,
		ReceiverUserID:    t.ReceiverUserID,
		SenderAccountID:   t.SenderAccountID,
		ReceiverAccountID: t.ReceiverAccountID,
		Amount:            amountStr,
		Currency:          t.Currency,
		Status:            string(t.Status),
	}

	if t.Description.Valid {
		resp.Description = t.Description.String
	}

	if t.CreatedAt.Valid {
		createdAt := t.CreatedAt.Time.UTC()
		resp.CreatedAt = &createdAt
	}

	if t.UpdatedAt.Valid {
		updatedAt := t.UpdatedAt.Time.UTC()
		resp.UpdatedAt = &updatedAt
	}

	if t.CompletedAt.Valid {
		completedAt := t.CompletedAt.Time.UTC()
		resp.CompletedAt = &completedAt
	}

	return resp
}
