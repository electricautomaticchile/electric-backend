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

type NotificacionRepository struct {
	collection *mongo.Collection
}

func NewNotificacionRepository() *NotificacionRepository {
	return &NotificacionRepository{
		collection: config.MongoDB.Collection(entities.NotificacionEntity{}.CollectionName()),
	}
}

func (r *NotificacionRepository) FindByDestinatario(ctx context.Context, destinatarioID string) ([]*entities.NotificacionEntity, error) {
	destinatarioObjectID, err := primitive.ObjectIDFromHex(destinatarioID)
	if err != nil {
		return []*entities.NotificacionEntity{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{"destinatarioId": destinatarioObjectID})
	if err != nil {
		return []*entities.NotificacionEntity{}, nil
	}
	defer cursor.Close(ctx)

	var notificaciones []*entities.NotificacionEntity
	if err := cursor.All(ctx, &notificaciones); err != nil {
		return []*entities.NotificacionEntity{}, nil
	}

	if notificaciones == nil {
		return []*entities.NotificacionEntity{}, nil
	}

	return notificaciones, nil
}

func (r *NotificacionRepository) FindByID(ctx context.Context, id string) (*entities.NotificacionEntity, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, types.ThrowData("ID inválido")
	}

	var notificacion entities.NotificacionEntity
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&notificacion)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Notificación no encontrada")
		}
		return nil, err
	}
	return &notificacion, nil
}

func (r *NotificacionRepository) Create(ctx context.Context, notificacion *entities.NotificacionEntity) error {
	notificacion.FechaCreacion = time.Now()
	notificacion.Leida = false

	result, err := r.collection.InsertOne(ctx, notificacion)
	if err != nil {
		return err
	}

	notificacion.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *NotificacionRepository) MarcarLeida(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.UpdateOne(
ctx,
bson.M{"_id": objectID},
bson.M{"$set": bson.M{"leida": true}},
)
	return err
}

func (r *NotificacionRepository) MarcarTodasLeidas(ctx context.Context, destinatarioID string) error {
	destinatarioObjectID, err := primitive.ObjectIDFromHex(destinatarioID)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.UpdateMany(
ctx,
bson.M{"destinatarioId": destinatarioObjectID, "leida": false},
bson.M{"$set": bson.M{"leida": true}},
)
	return err
}

func (r *NotificacionRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
