package data

import (
	"context"
	"crypto/rand"
	"electric-backend/config"
	"electric-backend/infrastructure/entities"
	"electric-backend/types"
	"encoding/hex"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RefreshTokenRepository struct {
	collection *mongo.Collection
}

func NewRefreshTokenRepository() *RefreshTokenRepository {
	return &RefreshTokenRepository{
		collection: config.MongoDB.Collection(entities.RefreshTokenEntity{}.CollectionName()),
	}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, userID, userType string) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", types.ThrowData("ID inválido")
	}

	entity := &entities.RefreshTokenEntity{
		Token:     token,
		UserID:    userObjectID,
		UserType:  userType,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	_, err = r.collection.InsertOne(ctx, entity)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (r *RefreshTokenRepository) Validate(ctx context.Context, token string) (*entities.RefreshTokenEntity, error) {
	var entity entities.RefreshTokenEntity
	err := r.collection.FindOne(ctx, bson.M{
		"token":   token,
		"revoked": false,
		"expiresAt": bson.M{"$gt": time.Now()},
	}).Decode(&entity)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowAuth("Token inválido o expirado")
		}
		return nil, err
	}

	return &entity, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"token": token},
		bson.M{
			"$set": bson.M{
				"revoked":   true,
				"revokedAt": now,
			},
		},
	)
	return err
}

func (r *RefreshTokenRepository) RevokeAllByUser(ctx context.Context, userID string) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	now := time.Now()
	_, err = r.collection.UpdateMany(
		ctx,
		bson.M{"userId": userObjectID, "revoked": false},
		bson.M{
			"$set": bson.M{
				"revoked":   true,
				"revokedAt": now,
			},
		},
	)
	return err
}

func (r *RefreshTokenRepository) CleanExpired(ctx context.Context) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{
		"expiresAt": bson.M{"$lt": time.Now()},
	})
	return err
}
