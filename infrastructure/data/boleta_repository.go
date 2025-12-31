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

type BoletaRepository struct {
	collection *mongo.Collection
}

func NewBoletaRepository() *BoletaRepository {
	return &BoletaRepository{
		collection: config.MongoDB.Collection(entities.BoletaEntity{}.CollectionName()),
	}
}

func (r *BoletaRepository) FindByCliente(ctx context.Context, clienteID string) ([]*entities.BoletaEntity, error) {
	clienteObjectID, err := primitive.ObjectIDFromHex(clienteID)
	if err != nil {
		return []*entities.BoletaEntity{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{"clienteId": clienteObjectID})
	if err != nil {
		return []*entities.BoletaEntity{}, nil
	}
	defer cursor.Close(ctx)

	var boletas []*entities.BoletaEntity
	if err := cursor.All(ctx, &boletas); err != nil {
		return []*entities.BoletaEntity{}, nil
	}
	if boletas == nil {
		return []*entities.BoletaEntity{}, nil
	}
	return boletas, nil
}

func (r *BoletaRepository) FindByID(ctx context.Context, id string) (*entities.BoletaEntity, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, types.ThrowData("ID inválido")
	}

	var boleta entities.BoletaEntity
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&boleta)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Boleta no encontrada")
		}
		return nil, err
	}
	return &boleta, nil
}

func (r *BoletaRepository) Create(ctx context.Context, boleta *entities.BoletaEntity) error {
	boleta.FechaCreacion = time.Now()
	boleta.Estado = "pendiente"

	result, err := r.collection.InsertOne(ctx, boleta)
	if err != nil {
		return err
	}

	boleta.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}
