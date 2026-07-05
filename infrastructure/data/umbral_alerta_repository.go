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

type UmbralAlertaRepository struct {
	collection *mongo.Collection
}

func NewUmbralAlertaRepository() *UmbralAlertaRepository {
	return &UmbralAlertaRepository{
		collection: config.MongoDB.Collection(entities.UmbralAlertaEntity{}.CollectionName()),
	}
}

func (r *UmbralAlertaRepository) FindByEmpresa(ctx context.Context, empresaID string) (*entities.UmbralAlertaEntity, error) {
	var umbral entities.UmbralAlertaEntity
	err := r.collection.FindOne(ctx, bson.M{"empresaId": empresaID}).Decode(&umbral)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &umbral, nil
}

// Upsert crea o actualiza el umbral de una empresa de forma idempotente por empresaId.
func (r *UmbralAlertaRepository) Upsert(ctx context.Context, umbral *entities.UmbralAlertaEntity) error {
	if umbral.EmpresaID == "" {
		return types.ThrowData("empresaId requerido")
	}

	filter := bson.M{"empresaId": umbral.EmpresaID}
	update := bson.M{
		"$set": bson.M{
			"voltajeMin":   umbral.VoltajeMin,
			"voltajeMax":   umbral.VoltajeMax,
			"corrienteMax": umbral.CorrienteMax,
			"consumoMax":   umbral.ConsumoMax,
			"updatedAt":    time.Now(),
		},
		"$setOnInsert": bson.M{"empresaId": umbral.EmpresaID},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}
