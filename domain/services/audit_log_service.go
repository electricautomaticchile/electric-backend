package services

import (
	"context"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/types"
	"time"
)

type AuditLogService struct {
	auditLogRepo ports.PortAuditLog
}

func NewAuditLogService(auditLogRepo ports.PortAuditLog) *AuditLogService {
	return &AuditLogService{
		auditLogRepo: auditLogRepo,
	}
}

func (s *AuditLogService) Log(ctx context.Context, log *models.AuditLogModel) error {
	return s.auditLogRepo.Create(ctx, log)
}

func (s *AuditLogService) GetLogs(ctx context.Context, empresaID string, params types.PaginationParams, filters types.FilterParams) (*types.PaginatedResponse, error) {
	return s.auditLogRepo.FindAll(ctx, empresaID, params, filters)
}

func (s *AuditLogService) GetUserLogs(ctx context.Context, userID string, params types.PaginationParams) (*types.PaginatedResponse, error) {
	return s.auditLogRepo.FindByUser(ctx, userID, params)
}

func (s *AuditLogService) GetResourceHistory(ctx context.Context, resource string, resourceID string) ([]*models.AuditLogModel, error) {
	return s.auditLogRepo.FindByResource(ctx, resource, resourceID)
}

func (s *AuditLogService) CleanOldLogs(ctx context.Context, days int) (int64, error) {
	return s.auditLogRepo.DeleteOlderThan(ctx, days)
}

func (s *AuditLogService) GetStatistics(ctx context.Context, empresaID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	return s.auditLogRepo.GetStatistics(ctx, empresaID, startDate, endDate)
}
