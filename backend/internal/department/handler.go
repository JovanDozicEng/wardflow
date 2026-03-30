package department

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
	"github.com/wardflow/backend/pkg/logger"
	"gorm.io/gorm"
)

// Handler handles department HTTP requests
type Handler struct {
	db *database.DB
}

// NewHandler creates a new department handler
func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// List handles GET /api/v1/departments
// Returns all departments, optionally filtered by ?q= (searches name and code)
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	// Get user context (auth middleware already applied)
	_ = auth.MustGetUserContext(r.Context())

	// Parse query parameter
	q := r.URL.Query().Get("q")

	var departments []Department
	tx := h.db.DB.Order("name ASC")

	// Apply search filter if provided
	if q != "" {
		searchPattern := "%" + q + "%"
		tx = tx.Where("name ILIKE ? OR code ILIKE ?", searchPattern, searchPattern)
	}

	if err := tx.Find(&departments).Error; err != nil {
		logger.Error("failed to list departments: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list departments")
		return
	}

	// Return array directly (no pagination for lookup tables)
	httputil.RespondJSON(w, http.StatusOK, departments)
}

// Create handles POST /api/v1/departments
// Admin only - creates a new department
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	// Get user context
	userCtx := auth.MustGetUserContext(r.Context())

	// Check admin role
	if userCtx.Role != models.RoleAdmin {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "Admin access required")
		return
	}

	// Decode request body
	var req CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "name is required")
		return
	}
	if req.Code == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "code is required")
		return
	}

	// Create department
	department := &Department{
		Name: req.Name,
		Code: req.Code,
	}

	if err := h.db.WithContext(r.Context()).Create(department).Error; err != nil {
		logger.Error("failed to create department: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "failed to create department")
		return
	}

	// Return created resource with 201 status
	httputil.RespondJSON(w, http.StatusCreated, department)
}

// Get handles GET /api/v1/departments/{departmentId}
// Returns a single department by ID
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	// Get user context
	_ = auth.MustGetUserContext(r.Context())

	// Extract ID from path parameter
	id := r.PathValue("departmentId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "department ID is required")
		return
	}

	// Fetch department by ID
	var department Department
	err := h.db.WithContext(r.Context()).Where("id = ?", id).First(&department).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "department not found")
			return
		}
		logger.Error("failed to get department: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get department")
		return
	}

	// Return department
	httputil.RespondJSON(w, http.StatusOK, department)
}
