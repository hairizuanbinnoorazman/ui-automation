package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hairizuan-noorazman/ui-automation/logger"
)

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response with a message.
type SuccessResponse struct {
	Message string `json:"message"`
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response with the given status code.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}

// respondSuccess writes a success response with the given message.
func respondSuccess(w http.ResponseWriter, message string) {
	respondJSON(w, http.StatusOK, SuccessResponse{Message: message})
}

// parseJSON parses JSON from the request body into the given destination.
func parseJSON(r *http.Request, dest interface{}, log logger.Logger) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		log.Error(r.Context(), "failed to parse JSON", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}
	return nil
}
