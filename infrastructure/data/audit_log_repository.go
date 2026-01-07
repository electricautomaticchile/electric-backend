package data

import (
	"context"
	"electric-backend/config"
	"electric-backend/domain/models"
	"electric-backend/infrastructure/entities"
	"electric-backend/types"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuditLogRepository struct{}

func NewAuditLogRepository() *AuditLogRepository {
	return &AuditLogRepository{}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *models.AuditLogModel) error {
	collection := config.MongoDB.Collection(entities.AuditLogEntity{}.CollectionName())

	entity := &entities.AuditLogEntity{
		UserID:       log.UserID,
		UserType:     log.UserType,
		UserEmail:    log.UserEmail,
		EmpresaID:    log.EmpresaID,
		Action:       log.Action,
		Resource:     log.Resource,
		ResourceID:   log.ResourceID,
		Method:       log.Method,
		Endpoint:     log.Endpoint,
		IPAddress:    log.IPAddress,
		UserAgent:    log.UserAgent,
		StatusCode:   log.StatusCode,
		Success:      log.Success,
		ErrorMessage: log.ErrorMessage,
		RequestBody:  log.RequestBody,
		ResponseBody: log.ResponseBody,
		Duration:     log.Duration,
		Timestamp:    time.Now(),
		Metadata:     log.Metadata,
	}

	_, err := collection.InsertOne(ctx, entity)
	return err
}

func (r *AuditLogRepository) FindAll(ctx context.Context, empresaID string, params types.PaginationParams, filters types.FilterParams) (*types.PaginatedResponse, error) {
	collection := config.MongoDB.Collection(entities.AuditLogEntity{}.CollectionName())

	filter := bson.M{}
	if empresaID != "" {
		filter["empresaId"] = empresaID
	}

	if filterBson := filters.BuildMongoFilter(); len(filterBson) > 0 {
		for k, v := range filterBson {
			filter[k] = v
		}
	}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	sortDir := 1
	if params.SortDir == "desc" {
		sortDir = -1
	}

	opts := options.Find()
	opts.SetSkip(int64(params.GetSkip()))
	opts.SetLimit(int64(params.GetLimit()))
	opts.SetSort(bson.D{{Key: params.SortBy, Value: sortDir}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []entities.AuditLogEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}

	logs := make([]*models.AuditLogModel, len(entities))
	for i, entity := range entities {
		logs[i] = &models.AuditLogModel{
			ID:           entity.ID.Hex(),
			UserID:       entity.UserID,
			UserType:     entity.UserType,
			UserEmail:    entity.UserEmail,
			EmpresaID:    entity.EmpresaID,
			Action:       entity.Action,
			Resource:     entity.Resource,
			ResourceID:   entity.ResourceID,
			Method:       entity.Method,
			Endpoint:     entity.Endpoint,
			IPAddress:    entity.IPAddress,
			UserAgent:    entity.UserAgent,
			StatusCode:   entity.StatusCode,
			Success:      entity.Success,
			ErrorMessage: entity.ErrorMessage,
			RequestBody:  entity.RequestBody,
			ResponseBody: entity.ResponseBody,
			Duration:     entity.Duration,
			Timestamp:    entity.Timestamp,
			Metadata:     entity.Metadata,
		}
	}

	return &types.PaginatedResponse{
		Data:       logs,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: (int(total) + params.PageSize - 1) / params.PageSize,
		HasNext:    params.GetSkip()+len(logs) < int(total),
		HasPrev:    params.Page > 1,
	}, nil
}

func (r *AuditLogRepository) FindByUser(ctx context.Context, userID string, params types.PaginationParams) (*types.PaginatedResponse, error) {
	collection := config.MongoDB.Collection(entities.AuditLogEntity{}.CollectionName())

	filter := bson.M{"userId": userID}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	opts := options.Find()
	opts.SetSkip(int64(params.GetSkip()))
	opts.SetLimit(int64(params.GetLimit()))
	opts.SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []entities.AuditLogEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}

	logs := make([]*models.AuditLogModel, len(entities))
	for i, entity := range entities {
		logs[i] = &models.AuditLogModel{
			ID:           entity.ID.Hex(),
			UserID:       entity.UserID,
			UserType:     entity.UserType,
			UserEmail:    entity.UserEmail,
			EmpresaID:    entity.EmpresaID,
			Action:       entity.Action,
			Resource:     entity.Resource,
			ResourceID:   entity.ResourceID,
			Method:       entity.Method,
			Endpoint:     entity.Endpoint,
			IPAddress:    entity.IPAddress,
			UserAgent:    entity.UserAgent,
			StatusCode:   entity.StatusCode,
			Success:      entity.Success,
			ErrorMessage: entity.ErrorMessage,
			Duration:     entity.Duration,
			Timestamp:    entity.Timestamp,
		}
	}

	return &types.PaginatedResponse{
		Data:       logs,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: (int(total) + params.PageSize - 1) / params.PageSize,
		HasNext:    params.GetSkip()+len(logs) < int(total),
		HasPrev:    params.Page > 1,
	}, nil
}

func (r *AuditLogRepository) FindByResource(ctx context.Context, resource string, resourceID string) ([]*models.AuditLogModel, error) {
	collection := config.MongoDB.Collection(entities.AuditLogEntity{}.CollectionName())

	filter := bson.M{
		"resource":   resource,
		"resourceId": resourceID,
	}

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []entities.AuditLogEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}

	logs := make([]*models.AuditLogModel, len(entities))
	for i, entity := range entities {
		logs[i] = &models.AuditLogModel{
			ID:         entity.ID.Hex(),
			UserID:     entity.UserID,
			UserType:   entity.UserType,
			UserEmail:  entity.UserEmail,
			EmpresaID:  entity.EmpresaID,
			Action:     entity.Action,
			Resource:   entity.Resource,
			ResourceID: entity.ResourceID,
			Method:     entity.Method,
			Endpoint:   entity.Endpoint,
			IPAddress:  entity.IPAddress,
			StatusCode: entity.StatusCode,
			Success:    entity.Success,
			Duration:   entity.Duration,
			Timestamp:  entity.Timestamp,
		}
	}

	return logs, nil
}

func (r *AuditLogRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	collection := config.MongoDB.Collection(entities.AuditLogEntity{}.CollectionName())

	cutoffDate := time.Now().AddDate(0, 0, -days)
	filter := bson.M{
		"timestamp": bson.M{"$lt": cutoffDate},
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

func (r *AuditLogRepository) GetStatistics(ctx context.Context, empresaID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	collection := config.MongoDB.Collection(entities.AuditLogEntity{}.CollectionName())

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	if empresaID != "" {
		filter["empresaId"] = empresaID
	}

	pipeline := []bson.M{
		{"$match": filter},
		{
			"$group": bson.M{
				"_id": nil,
				"totalActions": bson.M{"$sum": 1},
				"successfulActions": bson.M{
					"$sum": bson.M{"$cond": bson.A{"$success", 1, 0}},
				},
				"failedActions": bson.M{
					"$sum": bson.M{"$cond": bson.A{"$success", 0, 1}},
				},
				"avgDuration": bson.M{"$avg": "$duration"},
				"uniqueUsers": bson.M{"$addToSet": "$userId"},
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return map[string]interface{}{
			"totalActions":       0,
			"successfulActions":  0,
			"failedActions":      0,
			"avgDuration":        0,
			"uniqueUsers":        0,
		}, nil
	}

	result := results[0]
	uniqueUsers := result["uniqueUsers"].(primitive.A)

	return map[string]interface{}{
		"totalActions":       result["totalActions"],
		"successfulActions":  result["successfulActions"],
		"failedActions":      result["failedActions"],
		"avgDuration":        result["avgDuration"],
		"uniqueUsers":        len(uniqueUsers),
	}, nil
}
