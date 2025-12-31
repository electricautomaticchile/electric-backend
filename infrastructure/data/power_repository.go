package data

import (
	"context"
	"electric-backend/config"
	"electric-backend/infrastructure/entities"
	"electric-backend/types"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PowerRepository struct {
	collection *mongo.Collection
}

func NewPowerRepository() *PowerRepository {
	return &PowerRepository{
		collection: config.MongoDB.Collection(entities.PowerEntity{}.CollectionName()),
	}
}

func (r *PowerRepository) FindByUsuarioID(ctx context.Context, usuarioID string) ([]string, error) {
	objectID, err := primitive.ObjectIDFromHex(usuarioID)
	if err != nil {
		return []string{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{"usuarioId": objectID})
	if err != nil {
		return []string{}, nil
	}
	defer cursor.Close(ctx)

	var powers []entities.PowerEntity
	if err := cursor.All(ctx, &powers); err != nil {
		return []string{}, nil
	}

	result := make([]string, 0, len(powers))
	for _, p := range powers {
		result = append(result, p.Power)
	}

	return result, nil
}

func (r *PowerRepository) Create(ctx context.Context, power *entities.PowerEntity) error {
	power.FechaCreacion = time.Now()

	result, err := r.collection.InsertOne(ctx, power)
	if err != nil {
		return err
	}

	power.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *PowerRepository) Delete(ctx context.Context, usuarioID string, power string) error {
	objectID, err := primitive.ObjectIDFromHex(usuarioID)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{
		"usuarioId": objectID,
		"power":     power,
	})
	return err
}

func (r *PowerRepository) DeleteAllByUsuarioID(ctx context.Context, usuarioID string) error {
	objectID, err := primitive.ObjectIDFromHex(usuarioID)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.DeleteMany(ctx, bson.M{"usuarioId": objectID})
	return err
}
