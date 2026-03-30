package httputil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/logger"
)

// CorrelationIDKey is the context key for correlation IDs
type contextKey string

const CorrelationIDKey contextKey = "correlationID"

// RespondJSON writes a JSON response with the given status code
func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("failed to encode JSON response: %v", err)
	}
}

// RespondError writes a spec-compliant error response
func RespondError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	correlationID := CorrelationIDFromContext(r.Context())
	
	apiError := models.APIError{
		Code:          code,
		Message:       message,
		CorrelationID: correlationID,
	}
	
	envelope := models.ErrorEnvelope{
		Error: apiError,
	}
	
	RespondJSON(w, status, envelope)
}

// RespondValidationError writes a spec-compliant validation error response
func RespondValidationError(w http.ResponseWriter, r *http.Request, details []models.FieldError) {
	correlationID := CorrelationIDFromContext(r.Context())
	
	apiError := models.APIError{
		Code:          "VALIDATION_ERROR",
		Message:       "Request validation failed",
		Details:       details,
		CorrelationID: correlationID,
	}
	
	envelope := models.ErrorEnvelope{
		Error: apiError,
	}
	
	RespondJSON(w, http.StatusBadRequest, envelope)
}

// CorrelationIDFromContext extracts the correlation ID from context
func CorrelationIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

// DecodeJSON decodes JSON request body into the given value
func DecodeJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()
	
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Strict parsing
	
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	
	return nil
}
