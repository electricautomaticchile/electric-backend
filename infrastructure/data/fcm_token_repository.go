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
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FCMTokenRepository struct {
	collection *mongo.Collection
}

func NewFCMTokenRepository() *FCMTokenRepository {
	return &FCMTokenRepository{
		collection: config.MongoDB.Collection(entities.FCMTokenEntity{}.CollectionName()),
	}
}

// Upsert registra o actualiza un token de forma idempotente por token.
// Si el token ya existe, actualiza userId/plataforma/updatedAt. Permite
// múltiples tokens por usuario (multi-dispositivo).
func (r *FCMTokenRepository) Upsert(ctx context.Context, token *entities.FCMTokenEntity) error {
	now := time.Now()

	filter := bson.M{"token": token.Token}
	update := bson.M{
		"$set": bson.M{
			"userId":     token.UserID,
			"plataforma": token.Plataforma,
			"updatedAt":  now,
		},
		"$setOnInsert": bson.M{
			"token":     token.Token,
			"createdAt": now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

func (r *FCMTokenRepository) FindByUsuario(ctx context.Context, userID string) ([]*entities.FCMTokenEntity, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return []*entities.FCMTokenEntity{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{"userId": userObjectID})
	if err != nil {
		return []*entities.FCMTokenEntity{}, nil
	}
	defer cursor.Close(ctx)

	var tokens []*entities.FCMTokenEntity
	if err := cursor.All(ctx, &tokens); err != nil {
		return []*entities.FCMTokenEntity{}, nil
	}

	if tokens == nil {
		return []*entities.FCMTokenEntity{}, nil
	}

	return tokens, nil
}

func (r *FCMTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	if token == "" {
		return types.ThrowData("Token inválido")
	}

	_, err := r.collection.DeleteOne(ctx, bson.M{"token": token})
	return err
}
