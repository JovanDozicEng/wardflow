package patient

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/wardflow/backend/internal/audit"
	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
	"github.com/wardflow/backend/pkg/logger"
)

// Handler handles patient HTTP requests
type Handler struct {
	service *Service
	db      *database.DB
}

// NewHandler creates a new patient handler
func NewHandler(service *Service, db *database.DB) *Handler {
	return &Handler{
		service: service,
		db:      db,
	}
}

// List handles GET /api/v1/patients
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	// Get user context (auth middleware already applied)
	_ = auth.MustGetUserContext(r.Context())

	// Parse query parameters
	query := r.URL.Query()
	filter := ListPatientsFilter{
		Q:      query.Get("q"),
		Limit:  20, // default
		Offset: 0,
	}

	// Parse numeric parameters with validation
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Call service
	patients, total, err := h.service.List(r.Context(), filter)
	if err != nil {
		logger.Error("failed to list patients: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list patients")
		return
	}

	// Respond with paginated response
	response := models.PaginatedResponse{
		Data:   patients,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	httputil.RespondJSON(w, http.StatusOK, response)
}

// Create handles POST /api/v1/patients
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	// Get user context
	userCtx := auth.MustGetUserContext(r.Context())

	// Decode request body
	var req CreatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	// Call service
	patient, err := h.service.Create(r.Context(), &req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to create patient: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Write audit log
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "patient",
		EntityID:   patient.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      patient,
	})

	// Respond with created resource (201)
	httputil.RespondJSON(w, http.StatusCreated, patient)
}

// Get handles GET /api/v1/patients/{patientId}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	// Get user context
	_ = auth.MustGetUserContext(r.Context())

	// Extract ID from path (Go 1.22+ pattern with PathValue)
	id := r.PathValue("patientId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "patient ID is required")
		return
	}

	// Call service
	patient, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "patient not found")
			return
		}
		logger.Error("failed to get patient: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get patient")
		return
	}

	// Respond
	httputil.RespondJSON(w, http.StatusOK, patient)
}
