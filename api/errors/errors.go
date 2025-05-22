/*
Package errors provides utility functions for handling and formatting errors in the API.
It includes functions for writing error responses in JSON format, as well as specific error handlers for different types of errors.
*/
package errors

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Code    int
	Message string
}

func writeError(w http.ResponseWriter, message string, code int) {
	resp := Error{
		Code:    code,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(resp)
}

var (
	RequestErrorHandler = func(w http.ResponseWriter, err error) {
		writeError(w, err.Error(), http.StatusBadRequest)
	}
	InternalErrorHandler = func(w http.ResponseWriter) {
		writeError(w, "An Unexpected Error Occurred.", http.StatusInternalServerError)
	}
	UnauthorizedErrorHandler = func(w http.ResponseWriter, message string) {
		writeError(w, message, http.StatusUnauthorized)
	}
)
