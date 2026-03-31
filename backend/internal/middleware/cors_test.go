package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	t.Run("sets CORS headers for allowed origin", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000", "https://example.com"}
		middleware := CORSMiddleware(allowedOrigins)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		r.Header.Set("Origin", "http://localhost:3000")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "http://localhost:3000", rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Accept, Authorization, Content-Type, X-CSRF-Token", rr.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "86400", rr.Header().Get("Access-Control-Max-Age"))
	})

	t.Run("sets CORS headers for wildcard origin", func(t *testing.T) {
		allowedOrigins := []string{"*"}
		middleware := CORSMiddleware(allowedOrigins)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		r.Header.Set("Origin", "http://any-origin.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "http://any-origin.com", rr.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("does not set CORS headers for disallowed origin", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000"}
		middleware := CORSMiddleware(allowedOrigins)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		r.Header.Set("Origin", "http://evil-site.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("does not set CORS headers when Origin is missing", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000"}
		middleware := CORSMiddleware(allowedOrigins)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		// No Origin header
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("handles OPTIONS preflight request", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000"}
		middleware := CORSMiddleware(allowedOrigins)

		handlerCalled := false
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodOptions, "/api/test", nil)
		r.Header.Set("Origin", "http://localhost:3000")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.False(t, handlerCalled, "handler should not be called for OPTIONS preflight")
		assert.Equal(t, "http://localhost:3000", rr.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("allows normal request to proceed after OPTIONS", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000"}
		middleware := CORSMiddleware(allowedOrigins)

		handlerCalled := false
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodPost, "/api/test", nil)
		r.Header.Set("Origin", "http://localhost:3000")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.True(t, handlerCalled, "handler should be called for non-OPTIONS request")
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("matches exact origin from multiple allowed", func(t *testing.T) {
		allowedOrigins := []string{"http://localhost:3000", "https://app.example.com", "https://staging.example.com"}
		middleware := CORSMiddleware(allowedOrigins)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		r.Header.Set("Origin", "https://staging.example.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "https://staging.example.com", rr.Header().Get("Access-Control-Allow-Origin"))
	})
}
