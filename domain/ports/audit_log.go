package ports

import (
	"context"
	"electric-backend/domain/models"
	"electric-backend/types"
	"time"
)

type PortAuditLog interface {
	Create(ctx context.Context, log *models.AuditLogModel) error
	FindAll(ctx context.Context, empresaID string, params types.PaginationParams, filters types.FilterParams) (*types.PaginatedResponse, error)
	FindByUser(ctx context.Context, userID string, params types.PaginationParams) (*types.PaginatedResponse, error)
	FindByResource(ctx context.Context, resource string, resourceID string) ([]*models.AuditLogModel, error)
	DeleteOlderThan(ctx context.Context, days int) (int64, error)
	GetStatistics(ctx context.Context, empresaID string, startDate, endDate time.Time) (map[string]interface{}, error)
}
