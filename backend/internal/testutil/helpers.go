// Package testutil provides shared helpers for HTTP handler tests.
package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/auth"
)

// WithUser returns a copy of ctx with an authenticated UserContext injected.
// UnitIDs and DeptIDs default to empty slices unless provided via opts.
func WithUser(ctx context.Context, userID string, role models.Role, opts ...func(*auth.Claims)) context.Context {
	claims := &auth.Claims{
		UserID: userID,
		Email:  userID + "@test.local",
		Role:   role,
	}
	for _, opt := range opts {
		opt(claims)
	}
	return auth.SetUserContext(ctx, claims)
}

// NewRequest creates an httptest.Request with a JSON body and an authenticated
// UserContext pre-injected into the request context.
func NewRequest(method, target string, body any, userID string, role models.Role) *http.Request {
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewBuffer(b)
	}
	req := httptest.NewRequest(method, target, r)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	ctx := WithUser(req.Context(), userID, role)
	return req.WithContext(ctx)
}

// NewRequestNoAuth creates an httptest.Request without any user context.
func NewRequestNoAuth(method, target string, body any) *http.Request {
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewBuffer(b)
	}
	req := httptest.NewRequest(method, target, r)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

// DecodeJSON decodes the response body into v.
func DecodeJSON(t interface{ Fatal(...any) }, rr *httptest.ResponseRecorder, v any) {
	if err := json.NewDecoder(rr.Body).Decode(v); err != nil {
		t.Fatal("DecodeJSON:", err)
	}
}

// MustMarshal returns the JSON encoding of v or panics.
func MustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
