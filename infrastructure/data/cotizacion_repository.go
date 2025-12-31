package data

import (
	"context"
	"electric-backend/config"
	"electric-backend/domain/models"
	"electric-backend/infrastructure/entities"
	"electric-backend/types"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CotizacionRepository struct {
	collection *mongo.Collection
}

func NewCotizacionRepository() *CotizacionRepository {
	return &CotizacionRepository{
		collection: config.MongoDB.Collection(entities.CotizacionEntity{}.CollectionName()),
	}
}

// Conversión de entity a model
func (r *CotizacionRepository) entityToModel(entity *entities.CotizacionEntity) *models.CotizacionModel {
	var fechaCreacion time.Time
	if fc, ok := entity.FechaCreacion.(int64); ok {
		fechaCreacion = time.UnixMilli(fc)
	} else if fc, ok := entity.FechaCreacion.(time.Time); ok {
		fechaCreacion = fc
	}

	var fechaActualizacion *time.Time
	if entity.FechaActualizacion != nil {
		if fa, ok := entity.FechaActualizacion.(int64); ok {
			t := time.UnixMilli(fa)
			fechaActualizacion = &t
		} else if fa, ok := entity.FechaActualizacion.(time.Time); ok {
			fechaActualizacion = &fa
		}
	}

	var id string
	if !entity.ID.IsZero() {
		id = entity.ID.Hex()
	} else {
		id = primitive.NewObjectID().Hex()
	}

	return &models.CotizacionModel{
		ID:                 id,
		Numero:             entity.Numero,
		Nombre:             entity.Nombre,
		Email:              entity.Email,
		Empresa:            entity.Empresa,
		Telefono:           entity.Telefono,
		Servicio:           entity.Servicio,
		Plazo:              entity.Plazo,
		Mensaje:            entity.Mensaje,
		Estado:             entity.Estado,
		Prioridad:          entity.Prioridad,
		FechaCreacion:      fechaCreacion,
		FechaActualizacion: fechaActualizacion,
	}
}

// Conversión de model a entity
func (r *CotizacionRepository) modelToEntity(model *models.CotizacionModel) *entities.CotizacionEntity {
	entity := &entities.CotizacionEntity{
		Numero:    model.Numero,
		Nombre:    model.Nombre,
		Email:     model.Email,
		Empresa:   model.Empresa,
		Telefono:  model.Telefono,
		Servicio:  model.Servicio,
		Plazo:     model.Plazo,
		Mensaje:   model.Mensaje,
		Estado:    model.Estado,
		Prioridad: model.Prioridad,
	}

	if model.ID != "" {
		if oid, err := primitive.ObjectIDFromHex(model.ID); err == nil {
			entity.ID = oid
		}
	}

	return entity
}

func (r *CotizacionRepository) FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.CotizacionModel, int64, error) {
	// Construir filtros
	filter := bson.M{}
	if estado, ok := filters["estado"].(string); ok && estado != "" {
		filter["estado"] = estado
	}
	if servicio, ok := filters["servicio"].(string); ok && servicio != "" {
		filter["servicio"] = servicio
	}
	if prioridad, ok := filters["prioridad"].(string); ok && prioridad != "" {
		filter["prioridad"] = prioridad
	}

	// Contar total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Opciones de paginación
	skip := int64((page - 1) * limit)
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "fechaCreacion", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return []*models.CotizacionModel{}, 0, nil
	}
	defer cursor.Close(ctx)

	var entities []*entities.CotizacionEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return []*models.CotizacionModel{}, 0, nil
	}

	if entities == nil {
		return []*models.CotizacionModel{}, 0, nil
	}

	// Convertir entities a models
	cotizaciones := make([]*models.CotizacionModel, len(entities))
	for i, entity := range entities {
		cotizaciones[i] = r.entityToModel(entity)
	}

	return cotizaciones, total, nil
}

func (r *CotizacionRepository) FindByID(ctx context.Context, id string) (*models.CotizacionModel, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, types.ThrowData("ID inválido")
	}

	var entity entities.CotizacionEntity
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Cotización no encontrada")
		}
		return nil, err
	}
	return r.entityToModel(&entity), nil
}

func (r *CotizacionRepository) FindByNumero(ctx context.Context, numero string) (*models.CotizacionModel, error) {
	var entity entities.CotizacionEntity
	err := r.collection.FindOne(ctx, bson.M{"numero": numero}).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Cotización no encontrada")
		}
		return nil, err
	}
	return r.entityToModel(&entity), nil
}

func (r *CotizacionRepository) Create(ctx context.Context, model *models.CotizacionModel) error {
	entity := r.modelToEntity(model)
	entity.FechaCreacion = time.Now().UnixMilli()
	
	if entity.Estado == "" {
		entity.Estado = "pendiente"
	}
	
	// Generar número de cotización
	if entity.Numero == "" {
		year := time.Now().Year()
		count, _ := r.collection.CountDocuments(ctx, bson.M{})
		entity.Numero = fmt.Sprintf("COT-%d-%03d", year, count+1)
	}

	result, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return err
	}

	model.ID = result.InsertedID.(primitive.ObjectID).Hex()
	model.Numero = entity.Numero
	model.Estado = entity.Estado
	model.FechaCreacion = time.UnixMilli(entity.FechaCreacion.(int64))
	return nil
}

func (r *CotizacionRepository) Update(ctx context.Context, id string, model *models.CotizacionModel) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	entity := r.modelToEntity(model)
	now := time.Now().UnixMilli()
	entity.FechaActualizacion = &now

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": entity},
	)
	return err
}

func (r *CotizacionRepository) UpdateEstado(ctx context.Context, id string, estado string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	now := time.Now().UnixMilli()
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"estado":             estado,
				"fechaActualizacion": now,
			},
		},
	)
	return err
}

func (r *CotizacionRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
