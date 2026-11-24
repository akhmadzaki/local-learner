package helpers

import (
	"encoding/json"
	"log"
	"net/http"
)

// ── Standard JSON response helpers ────────────────────────────────────────────

// ErrorResponse is the standard error envelope returned on failures.
// It is also used as the swagger schema for error responses.
type ErrorResponse struct {
	Error   string `json:"error" example:"Bad Request"`
	Message string `json:"message" example:"Invalid request body"`
}

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Failed to encode JSON response: %v", err)
		}
	}
}

func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, map[string]interface{}{
		"error":   http.StatusText(status),
		"message": message,
	})
}

func RespondSuccess(w http.ResponseWriter, data interface{}) {
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

func RespondCreated(w http.ResponseWriter, data interface{}) {
	RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

func RespondPaginated(w http.ResponseWriter, data interface{}, page, perPage, total int) {
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "success",
		"data":        data,
		"page":        page,
		"per_page":    perPage,
		"total":       total,
		"total_pages": (total + perPage - 1) / perPage,
	})
}