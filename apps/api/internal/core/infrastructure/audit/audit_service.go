package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/data/models"
	"gorm.io/gorm"
)

type AuditService interface {
	Log(ctx context.Context, action string, targetID string, metadata map[string]interface{})
	LogWithReason(ctx context.Context, action string, targetID string, reason string, metadata map[string]interface{})
	LogWithChanges(ctx context.Context, action string, targetID string, metadata map[string]interface{}, changes interface{})
	LogWithChangesFull(ctx context.Context, action string, targetID string, reason string, metadata map[string]interface{}, changes interface{})
}

type databaseAuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) AuditService {
	return &databaseAuditService{
		db: db,
	}
}

func (s *databaseAuditService) Log(ctx context.Context, action string, targetID string, metadata map[string]interface{}) {
	s.LogWithChangesFull(ctx, action, targetID, "", metadata, nil)
}

func (s *databaseAuditService) LogWithReason(ctx context.Context, action string, targetID string, reason string, metadata map[string]interface{}) {
	s.LogWithChangesFull(ctx, action, targetID, reason, metadata, nil)
}

func (s *databaseAuditService) LogWithChanges(ctx context.Context, action string, targetID string, metadata map[string]interface{}, changes interface{}) {
	s.LogWithChangesFull(ctx, action, targetID, "", metadata, changes)
}

func (s *databaseAuditService) LogWithChangesFull(ctx context.Context, action string, targetID string, reason string, metadata map[string]interface{}, changes interface{}) {
	// Extract actor from context (assumes AuthMiddleware sets "user_id" and "user_email")
	actorID := ""
	if v := ctx.Value("user_id"); v != nil {
		actorID = fmt.Sprintf("%v", v)
	}

	ip := ""
	userAgent := ""
	if v := ctx.Value("client_ip"); v != nil {
		ip = fmt.Sprintf("%v", v)
	}
	if v := ctx.Value("user_agent"); v != nil {
		userAgent = fmt.Sprintf("%v", v)
	}

	// Permission Code is often "resource.action".
	// We can infer it or pass it. Using action as PermissionCode for now.

	if metadata == nil {
		metadata = map[string]interface{}{}
	}

	metaJSON, _ := json.Marshal(metadata)
	changesStr := "null"
	if changes != nil {
		if cBytes, err := json.Marshal(changes); err == nil {
			changesStr = string(cBytes)
		}
	}

	auditLog := &models.AuditLog{
		ActorID:        actorID,
		PermissionCode: action, // Using action name as permission code ref
		TargetID:       targetID,
		Action:         action,
		IPAddress:      ip,
		UserAgent:      userAgent,
		Reason:         reason,
		Metadata:       string(metaJSON),
		Changes:        changesStr,
		ResultStatus:   "success", // Default to success if logged after action
		CreatedAt:      apptime.Now(),
	}

	// Synchronous write to ensure audit trail consistency.
	// Use a background context to avoid request cancellation dropping audit entries.
	dbCtx := context.Background()
	if err := s.db.WithContext(dbCtx).Create(auditLog).Error; err != nil {
		log.Printf("[audit] failed to write audit log: action=%s target_id=%s err=%v", action, targetID, err)
	}
}
