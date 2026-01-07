package data

import (
	"context"
	"electric-backend/config"
	"electric-backend/domain/models"
	"electric-backend/infrastructure/entities"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TarifaRepository struct{}

func NewTarifaRepository() *TarifaRepository {
	return &TarifaRepository{}
}

func (r *TarifaRepository) Create(ctx context.Context, tarifa *models.TarifaModel) error {
	collection := config.MongoDB.Collection(entities.TarifaEntity{}.CollectionName())

	vigenciaDesde, _ := time.Parse(time.RFC3339, tarifa.VigenciaDesde.Format(time.RFC3339))
	vigenciaHasta, _ := time.Parse(time.RFC3339, tarifa.VigenciaHasta.Format(time.RFC3339))

	entity := &entities.TarifaEntity{
		Distribuidora:        tarifa.Distribuidora,
		TipoTarifa:           tarifa.TipoTarifa,
		Comuna:               tarifa.Comuna,
		RedTipo:              tarifa.RedTipo,
		VigenciaDesde:        vigenciaDesde,
		VigenciaHasta:        vigenciaHasta,
		CargoEnergia:         tarifa.CargoEnergia,
		CargoTransmision:     tarifa.CargoTransmision,
		CargoServicioPublico: tarifa.CargoServicioPublico,
		PrecioKwhBase:        tarifa.PrecioKwhBase,
		PeajeDistribucion:    tarifa.PeajeDistribucion,
		TramosEstabilizacion: entities.TramosEstabilizacion{
			Hasta350Kwh:  tarifa.TramosEstabilizacion.Hasta350Kwh,
			Entre350Y500: tarifa.TramosEstabilizacion.Entre350Y500,
			Mayor500Kwh:  tarifa.TramosEstabilizacion.Mayor500Kwh,
		},
		Activa:             true,
		FechaCreacion:      time.Now(),
		FechaActualizacion: time.Now(),
	}

	result, err := collection.InsertOne(ctx, entity)
	if err != nil {
		return err
	}

	tarifa.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *TarifaRepository) FindByID(ctx context.Context, id string) (*models.TarifaModel, error) {
	collection := config.MongoDB.Collection(entities.TarifaEntity{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var entity entities.TarifaEntity
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&entity)
	if err != nil {
		return nil, err
	}

	return r.entityToModel(&entity), nil
}

func (r *TarifaRepository) FindActiva(ctx context.Context, comuna, tipoTarifa string) (*models.TarifaModel, error) {
	collection := config.MongoDB.Collection(entities.TarifaEntity{}.CollectionName())

	now := time.Now()
	filter := bson.M{
		"comuna":         comuna,
		"tipoTarifa":     tipoTarifa,
		"activa":         true,
		"vigenciaDesde":  bson.M{"$lte": now},
		"vigenciaHasta":  bson.M{"$gte": now},
	}

	var entity entities.TarifaEntity
	err := collection.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		return nil, err
	}

	return r.entityToModel(&entity), nil
}

func (r *TarifaRepository) FindAll(ctx context.Context) ([]*models.TarifaModel, error) {
	collection := config.MongoDB.Collection(entities.TarifaEntity{}.CollectionName())

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []entities.TarifaEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}

	tarifas := make([]*models.TarifaModel, len(entities))
	for i, entity := range entities {
		tarifas[i] = r.entityToModel(&entity)
	}

	return tarifas, nil
}

func (r *TarifaRepository) Update(ctx context.Context, id string, tarifa *models.TarifaModel) error {
	collection := config.MongoDB.Collection(entities.TarifaEntity{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"distribuidora":        tarifa.Distribuidora,
			"tipoTarifa":           tarifa.TipoTarifa,
			"comuna":               tarifa.Comuna,
			"redTipo":              tarifa.RedTipo,
			"vigenciaDesde":        tarifa.VigenciaDesde,
			"vigenciaHasta":        tarifa.VigenciaHasta,
			"cargoEnergia":         tarifa.CargoEnergia,
			"cargoTransmision":     tarifa.CargoTransmision,
			"cargoServicioPublico": tarifa.CargoServicioPublico,
			"precioKwhBase":        tarifa.PrecioKwhBase,
			"peajeDistribucion":    tarifa.PeajeDistribucion,
			"tramosEstabilizacion": bson.M{
				"hasta350Kwh":  tarifa.TramosEstabilizacion.Hasta350Kwh,
				"entre350Y500": tarifa.TramosEstabilizacion.Entre350Y500,
				"mayor500Kwh":  tarifa.TramosEstabilizacion.Mayor500Kwh,
			},
			"activa":             tarifa.Activa,
			"fechaActualizacion": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *TarifaRepository) Delete(ctx context.Context, id string) error {
	collection := config.MongoDB.Collection(entities.TarifaEntity{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

func (r *TarifaRepository) entityToModel(entity *entities.TarifaEntity) *models.TarifaModel {
	return &models.TarifaModel{
		ID:                   entity.ID.Hex(),
		Distribuidora:        entity.Distribuidora,
		TipoTarifa:           entity.TipoTarifa,
		Comuna:               entity.Comuna,
		RedTipo:              entity.RedTipo,
		VigenciaDesde:        entity.VigenciaDesde,
		VigenciaHasta:        entity.VigenciaHasta,
		CargoEnergia:         entity.CargoEnergia,
		CargoTransmision:     entity.CargoTransmision,
		CargoServicioPublico: entity.CargoServicioPublico,
		PrecioKwhBase:        entity.PrecioKwhBase,
		PeajeDistribucion:    entity.PeajeDistribucion,
		TramosEstabilizacion: models.TramosEstabilizacion{
			Hasta350Kwh:  entity.TramosEstabilizacion.Hasta350Kwh,
			Entre350Y500: entity.TramosEstabilizacion.Entre350Y500,
			Mayor500Kwh:  entity.TramosEstabilizacion.Mayor500Kwh,
		},
		Activa:             entity.Activa,
		FechaCreacion:      entity.FechaCreacion,
		FechaActualizacion: entity.FechaActualizacion,
	}
}
