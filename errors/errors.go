package errors

import (
	"encoding/json"
	"net/http"
)

// AppError represents a typed domain error with an HTTP status code.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Standard domain errors.
var (
	ErrDocumentNotFound = &AppError{Code: "DOCUMENT_NOT_FOUND", Message: "Document not found", HTTPStatus: http.StatusNotFound}
	ErrIngestionFailed  = &AppError{Code: "INGESTION_FAILED", Message: "Document ingestion failed", HTTPStatus: http.StatusInternalServerError}
	ErrInvalidRequest   = &AppError{Code: "INVALID_REQUEST", Message: "Invalid request", HTTPStatus: http.StatusBadRequest}
	ErrEmbeddingFailed  = &AppError{Code: "EMBEDDING_FAILED", Message: "Embedding generation failed", HTTPStatus: http.StatusInternalServerError}
	ErrSearchFailed     = &AppError{Code: "SEARCH_FAILED", Message: "Search failed", HTTPStatus: http.StatusInternalServerError}
	ErrNotFound         = &AppError{Code: "NOT_FOUND", Message: "Resource not found", HTTPStatus: http.StatusNotFound}
	ErrRateLimited      = &AppError{Code: "RATE_LIMITED", Message: "Too many requests", HTTPStatus: http.StatusTooManyRequests}
)

// WriteJSON writes the AppError as a JSON response to the http.ResponseWriter.
func (e *AppError) WriteJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.HTTPStatus)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   e.Code,
		"message": e.Message,
	})
}

// NewAppError creates a new AppError with a custom message.
func NewAppError(code string, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

// Wrap wraps an existing error into an AppError.
func Wrap(err error, code string, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}
