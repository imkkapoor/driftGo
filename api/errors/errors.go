/*
Package errors provides utility functions for handling and formatting errors in the API.
It includes functions for writing error responses in JSON format, as well as specific error handlers for different types of errors.
*/
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error represents a standard API error response
type Error struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
}

// NewError creates a new Error with the given status code and message
func NewError(statusCode int, message string) *Error {
	return &Error{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewErrorWithCode creates a new Error with status code, message and error code
func NewErrorWithCode(statusCode int, message, code string) *Error {
	return &Error{
		StatusCode: statusCode,
		Message:    message,
		Code:       code,
	}
}

// Error implements the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("%s (status: %d)", e.Message, e.StatusCode)
}

// Common error codes
const (
	ErrCodeInvalidRequest  = "INVALID_REQUEST"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeInternalError   = "INTERNAL_ERROR"
	ErrCodeValidationError = "VALIDATION_ERROR"
	ErrCodeAuthentication  = "AUTHENTICATION_ERROR"
)

// Common error messages
const (
	MsgInvalidRequest  = "The request was invalid or malformed"
	MsgUnauthorized    = "You are not authorized to perform this action"
	MsgForbidden       = "You don't have permission to access this resource"
	MsgNotFound        = "The requested resource was not found"
	MsgInternalError   = "An unexpected error occurred. Please try again later"
	MsgValidationError = "The request failed validation"
	MsgAuthentication  = "Authentication failed"
)

func writeError(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(err)
}

// Error handlers
var (
	RequestErrorHandler = func(w http.ResponseWriter, err error) {
		var apiErr *Error
		if e, ok := err.(*Error); ok {
			apiErr = e
		} else {
			apiErr = NewErrorWithCode(http.StatusBadRequest, err.Error(), ErrCodeInvalidRequest)
		}
		writeError(w, apiErr)
	}

	InternalErrorHandler = func(w http.ResponseWriter) {
		writeError(w, NewErrorWithCode(http.StatusInternalServerError, MsgInternalError, ErrCodeInternalError))
	}

	UnauthorizedErrorHandler = func(w http.ResponseWriter, message string) {
		writeError(w, NewErrorWithCode(http.StatusUnauthorized, message, ErrCodeUnauthorized))
	}

	NotFoundErrorHandler = func(w http.ResponseWriter, message string) {
		writeError(w, NewErrorWithCode(http.StatusNotFound, message, ErrCodeNotFound))
	}

	ForbiddenErrorHandler = func(w http.ResponseWriter, message string) {
		writeError(w, NewErrorWithCode(http.StatusForbidden, message, ErrCodeForbidden))
	}

	ValidationErrorHandler = func(w http.ResponseWriter, message string) {
		writeError(w, NewErrorWithCode(http.StatusBadRequest, message, ErrCodeValidationError))
	}
)
