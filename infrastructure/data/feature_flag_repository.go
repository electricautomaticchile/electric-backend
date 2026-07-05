package data

import (
	"context"
	"electric-backend/config"
	"electric-backend/infrastructure/entities"
	"electric-backend/types"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FeatureFlagRepository struct {
	collection *mongo.Collection
}

func NewFeatureFlagRepository() *FeatureFlagRepository {
	return &FeatureFlagRepository{
		collection: config.MongoDB.Collection(entities.FeatureFlagEntity{}.CollectionName()),
	}
}

func (r *FeatureFlagRepository) FindAll(ctx context.Context) ([]*entities.FeatureFlagEntity, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return []*entities.FeatureFlagEntity{}, nil
	}
	defer cursor.Close(ctx)

	var flags []*entities.FeatureFlagEntity
	if err := cursor.All(ctx, &flags); err != nil {
		return []*entities.FeatureFlagEntity{}, nil
	}
	if flags == nil {
		return []*entities.FeatureFlagEntity{}, nil
	}
	return flags, nil
}

func (r *FeatureFlagRepository) FindByKey(ctx context.Context, key string) (*entities.FeatureFlagEntity, error) {
	var flag entities.FeatureFlagEntity
	err := r.collection.FindOne(ctx, bson.M{"key": key}).Decode(&flag)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &flag, nil
}

// Upsert crea o actualiza un flag de forma idempotente por key.
func (r *FeatureFlagRepository) Upsert(ctx context.Context, flag *entities.FeatureFlagEntity) error {
	if flag.Key == "" {
		return types.ThrowData("key requerida")
	}
	if flag.EmpresaIDs == nil {
		flag.EmpresaIDs = []string{}
	}

	filter := bson.M{"key": flag.Key}
	update := bson.M{
		"$set": bson.M{
			"descripcion": flag.Descripcion,
			"enabled":     flag.Enabled,
			"empresaIds":  flag.EmpresaIDs,
			"updatedAt":   time.Now(),
		},
		"$setOnInsert": bson.M{"key": flag.Key},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

func (r *FeatureFlagRepository) Delete(ctx context.Context, key string) error {
	if key == "" {
		return types.ThrowData("key requerida")
	}
	_, err := r.collection.DeleteOne(ctx, bson.M{"key": key})
	return err
}
