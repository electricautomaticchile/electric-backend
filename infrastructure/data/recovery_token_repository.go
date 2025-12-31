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

type RecoveryTokenRepository struct {
	collection *mongo.Collection
}

func NewRecoveryTokenRepository() *RecoveryTokenRepository {
	return &RecoveryTokenRepository{
		collection: config.MongoDB.Collection(entities.RecoveryTokenEntity{}.CollectionName()),
	}
}

func (r *RecoveryTokenRepository) Create(ctx context.Context, token *entities.RecoveryTokenEntity) error {
	token.FechaCreacion = time.Now()
	token.Usado = false
	
	result, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		return err
	}
	
	token.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *RecoveryTokenRepository) FindByToken(ctx context.Context, token string) (*entities.RecoveryTokenEntity, error) {
	var recoveryToken entities.RecoveryTokenEntity
	err := r.collection.FindOne(ctx, bson.M{
		"token": token,
		"usado": false,
		"fechaExpiracion": bson.M{"$gt": time.Now()},
	}).Decode(&recoveryToken)
	
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Token inválido o expirado")
		}
		return nil, err
	}
	return &recoveryToken, nil
}

func (r *RecoveryTokenRepository) MarkAsUsed(ctx context.Context, tokenID string) error {
	objectID, err := primitive.ObjectIDFromHex(tokenID)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"usado": true}},
	)
	return err
}

func (r *RecoveryTokenRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{
		"fechaExpiracion": bson.M{"$lt": time.Now()},
	})
	return err
}
