package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/pkg/logger"
)

// CorrelationID middleware generates a unique correlation ID per request
func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate UUID v4
		correlationID := generateUUID()
		
		// Set response header
		w.Header().Set("X-Correlation-ID", correlationID)
		
		// Store correlation ID in context using the typed key from httputil
		ctx := context.WithValue(r.Context(), httputil.CorrelationIDKey, correlationID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateUUID generates a simple UUID v4
func generateUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		logger.Warn("failed to generate UUID: %v", err)
		return "00000000-0000-0000-0000-000000000000"
	}
	
	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
