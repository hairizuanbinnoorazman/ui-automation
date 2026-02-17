package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

// parseUUID parses a UUID from the request path parameters.
func parseUUID(r *http.Request, paramName string) (uuid.UUID, error) {
	vars := mux.Vars(r)
	uuidStr := vars[paramName]
	return uuid.Parse(uuidStr)
}

// parseUUIDOrRespond parses a UUID from path parameters and responds with an error if invalid.
// Returns the UUID and true if successful, or uuid.Nil and false if parsing failed (error response already sent).
func parseUUIDOrRespond(w http.ResponseWriter, r *http.Request, paramName, entityName string) (uuid.UUID, bool) {
	id, err := parseUUID(r, paramName)
	if err != nil {
		respondError(w, http.StatusBadRequest,
			fmt.Sprintf("invalid %s ID: must be a valid UUID", entityName))
		return uuid.Nil, false
	}
	return id, true
}
