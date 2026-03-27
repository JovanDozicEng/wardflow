package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/database"
	"github.com/wardflow/backend/pkg/logger"
)

// Entry holds data for a single audit log write
type Entry struct {
	EntityType string
	EntityID   string
	Action     string      // CREATE | UPDATE | DELETE | OVERRIDE
	ByUserID   string
	Reason     *string
	Source     string      // user_action | system_event (default: user_action)
	Before     interface{} // will be JSON-marshaled into BeforeJSON
	After      interface{} // will be JSON-marshaled into AfterJSON
}

// Log writes an audit entry; non-fatal on error (logs warning)
func Log(ctx context.Context, db *database.DB, r *http.Request, entry Entry) {
	// Set default source if empty
	if entry.Source == "" {
		entry.Source = "user_action"
	}

	// Extract IP from RemoteAddr (strip port if present)
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	// Extract User-Agent
	userAgent := r.UserAgent()

	// Extract correlation ID from context
	correlationID := httputil.CorrelationIDFromContext(ctx)

	// Marshal Before/After to JSON strings
	var beforeJSON, afterJSON *string
	if entry.Before != nil {
		data, err := json.Marshal(entry.Before)
		if err != nil {
			logger.Warn("failed to marshal Before data for audit log: %v", err)
		} else {
			str := string(data)
			beforeJSON = &str
		}
	}
	if entry.After != nil {
		data, err := json.Marshal(entry.After)
		if err != nil {
			logger.Warn("failed to marshal After data for audit log: %v", err)
		} else {
			str := string(data)
			afterJSON = &str
		}
	}

	// Create audit log entry
	auditLog := models.AuditLog{
		EntityType:    entry.EntityType,
		EntityID:      entry.EntityID,
		Action:        entry.Action,
		At:            time.Now().UTC(),
		ByUserID:      entry.ByUserID,
		IP:            &ip,
		UserAgent:     &userAgent,
		Reason:        entry.Reason,
		Source:        entry.Source,
		BeforeJSON:    beforeJSON,
		AfterJSON:     afterJSON,
		CorrelationID: &correlationID,
	}

	// Write to database
	if err := db.Create(&auditLog).Error; err != nil {
		logger.Warn("failed to write audit log: %v", err)
	}
}
