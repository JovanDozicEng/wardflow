package models

// FieldError represents a single field validation error
type FieldError struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

// APIError represents the error object in the spec-compliant error response
type APIError struct {
	Code          string       `json:"code"`
	Message       string       `json:"message"`
	Details       []FieldError `json:"details,omitempty"`
	CorrelationID string       `json:"correlationId,omitempty"`
}

// ErrorEnvelope wraps APIError for spec-compliant error responses
type ErrorEnvelope struct {
	Error APIError `json:"error"`
}

// Standard paginated response with both offset and cursor support
type PaginatedResponse struct {
	Data   interface{} `json:"data"`
	Total  int64       `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
	Cursor string      `json:"cursor,omitempty"` // for cursor-based pagination
}

// Health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version,omitempty"`
	Timestamp string `json:"timestamp"`
}
