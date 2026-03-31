package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wardflow/backend/internal/httputil"
)

func TestCorrelationID(t *testing.T) {
	t.Run("sets X-Correlation-ID header", func(t *testing.T) {
		handler := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		correlationID := rr.Header().Get("X-Correlation-ID")
		assert.NotEmpty(t, correlationID, "X-Correlation-ID header should be set")
	})

	t.Run("generates a valid UUID v4", func(t *testing.T) {
		handler := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		correlationID := rr.Header().Get("X-Correlation-ID")
		
		// UUID v4 format: 8-4-4-4-12 characters
		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, correlationID)
		
		// Check version bit (4xxx for UUID v4)
		assert.Equal(t, "4", string(correlationID[14]), "UUID should be version 4")
	})

	t.Run("stores correlation ID in request context", func(t *testing.T) {
		var contextCorrelationID string
		handler := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract correlation ID from context
			if id, ok := r.Context().Value(httputil.CorrelationIDKey).(string); ok {
				contextCorrelationID = id
			}
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		headerCorrelationID := rr.Header().Get("X-Correlation-ID")
		assert.Equal(t, headerCorrelationID, contextCorrelationID, "context correlation ID should match header")
	})

	t.Run("generates unique IDs for different requests", func(t *testing.T) {
		handler := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// First request
		r1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, r1)
		id1 := rr1.Header().Get("X-Correlation-ID")

		// Second request
		r2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, r2)
		id2 := rr2.Header().Get("X-Correlation-ID")

		assert.NotEqual(t, id1, id2, "correlation IDs should be unique")
	})

	t.Run("passes request to next handler", func(t *testing.T) {
		handlerCalled := false
		handler := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.True(t, handlerCalled, "next handler should be called")
	})
}

func TestGenerateUUID(t *testing.T) {
	t.Run("generates valid UUID format", func(t *testing.T) {
		uuid := generateUUID()
		
		// UUID v4 format: 8-4-4-4-12 hex characters
		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, uuid)
	})

	t.Run("generates UUID v4 with correct version", func(t *testing.T) {
		uuid := generateUUID()
		
		// Version 4 UUID has '4' in the version position (character 14)
		assert.Equal(t, "4", string(uuid[14]))
	})

	t.Run("generates UUID v4 with correct variant", func(t *testing.T) {
		uuid := generateUUID()
		
		// Variant bits should be 10xx (8, 9, a, or b at position 19)
		variantChar := string(uuid[19])
		assert.Contains(t, []string{"8", "9", "a", "b"}, variantChar)
	})

	t.Run("generates unique UUIDs", func(t *testing.T) {
		uuid1 := generateUUID()
		uuid2 := generateUUID()
		uuid3 := generateUUID()
		
		assert.NotEqual(t, uuid1, uuid2)
		assert.NotEqual(t, uuid2, uuid3)
		assert.NotEqual(t, uuid1, uuid3)
	})
}
