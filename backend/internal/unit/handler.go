package unit

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/logger"
	"gorm.io/gorm"
)

// Handler handles unit HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new unit handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// List handles GET /api/v1/units
// Returns all units, optionally filtered by ?q= (searches name and code) and ?departmentId=
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	// Get user context (auth middleware already applied)
	_ = auth.MustGetUserContext(r.Context())

	// Parse query parameters
	q := r.URL.Query().Get("q")
	departmentID := r.URL.Query().Get("departmentId")

	units, err := h.service.List(r.Context(), q, departmentID)
	if err != nil {
		logger.Error("failed to list units: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list units")
		return
	}

	// Return array directly (no pagination for lookup tables)
	httputil.RespondJSON(w, http.StatusOK, units)
}

// Create handles POST /api/v1/units
// Admin only - creates a new unit
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	// Get user context
	userCtx := auth.MustGetUserContext(r.Context())

	// Check admin role
	if userCtx.Role != models.RoleAdmin {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "Admin access required")
		return
	}

	// Decode request body
	var req CreateUnitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	unit, err := h.service.Create(r.Context(), req)
	if err != nil {
		logger.Error("failed to create unit: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Return created resource with 201 status
	httputil.RespondJSON(w, http.StatusCreated, *unit)
}

// Get handles GET /api/v1/units/{unitId}
// Returns a single unit by ID
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	// Get user context
	_ = auth.MustGetUserContext(r.Context())

	// Extract ID from path parameter
	id := r.PathValue("unitId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "unit ID is required")
		return
	}

	unit, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "unit not found")
			return
		}
		logger.Error("failed to get unit: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get unit")
		return
	}

	// Return unit
	httputil.RespondJSON(w, http.StatusOK, *unit)
}
