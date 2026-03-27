package models

// Standard API error response
type ErrorResponse struct {
	Error       string                 `json:"error"`
	Message     string                 `json:"message"`
	FieldErrors map[string]string      `json:"fieldErrors,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// Standard paginated response
type PaginatedResponse struct {
	Data   interface{} `json:"data"`
	Total  int64       `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

// Health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version,omitempty"`
	Timestamp string `json:"timestamp"`
}
