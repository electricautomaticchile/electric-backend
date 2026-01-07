package data

import (
	"context"
	"electric-backend/config"
	"electric-backend/domain/models"
	"electric-backend/infrastructure/entities"
	"electric-backend/types"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AlertaRepository struct {
	collection *mongo.Collection
}

func NewAlertaRepository() *AlertaRepository {
	return &AlertaRepository{
		collection: config.MongoDB.Collection(entities.AlertaEntity{}.CollectionName()),
	}
}

// Conversión de entity a model
func (r *AlertaRepository) entityToModel(entity *entities.AlertaEntity) *models.AlertaModel {
	var fechaResolucion *time.Time
	if entity.FechaResolucion != nil {
		fechaResolucion = entity.FechaResolucion
	}

	model := &models.AlertaModel{
		ID:               entity.ID.Hex(),
		EmpresaID:        entity.EmpresaID.Hex(),
		Tipo:             entity.Tipo,
		Titulo:           entity.Mensaje,
		Mensaje:          entity.Mensaje,
		Resuelta:         entity.Resuelta,
		AccionesTomadas:  entity.Resolucion,
		FechaCreacion:    entity.FechaCreacion,
		FechaResolucion:  fechaResolucion,
		Importante:       false,
		Leida:            false,
	}

	if !entity.DispositivoID.IsZero() {
		model.Dispositivo = entity.DispositivoID.Hex()
	}

	return model
}

// Conversión de model a entity
func (r *AlertaRepository) modelToEntity(model *models.AlertaModel) *entities.AlertaEntity {
	entity := &entities.AlertaEntity{
		Tipo:       model.Tipo,
		Severidad:  model.Tipo, // Mapear tipo a severidad
		Mensaje:    model.Mensaje,
		Resuelta:   model.Resuelta,
		Resolucion: model.AccionesTomadas,
	}

	if model.ID != "" {
		if oid, err := primitive.ObjectIDFromHex(model.ID); err == nil {
			entity.ID = oid
		}
	}

	if model.EmpresaID != "" {
		if oid, err := primitive.ObjectIDFromHex(model.EmpresaID); err == nil {
			entity.EmpresaID = oid
		}
	}

	if model.Dispositivo != "" {
		if oid, err := primitive.ObjectIDFromHex(model.Dispositivo); err == nil {
			entity.DispositivoID = oid
		}
	}

	return entity
}

func (r *AlertaRepository) FindAll(ctx context.Context) ([]*models.AlertaModel, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return []*models.AlertaModel{}, nil
	}
	defer cursor.Close(ctx)

	var entities []*entities.AlertaEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return []*models.AlertaModel{}, nil
	}
	if entities == nil {
		return []*models.AlertaModel{}, nil
	}

	alertas := make([]*models.AlertaModel, len(entities))
	for i, entity := range entities {
		alertas[i] = r.entityToModel(entity)
	}
	return alertas, nil
}

func (r *AlertaRepository) FindAllPaginated(ctx context.Context, empresaID string, params types.PaginationParams, filters types.FilterParams) ([]*models.AlertaModel, int64, error) {
	filter := filters.BuildMongoFilter()
	
	if empresaID != "" {
		empresaObjectID, err := primitive.ObjectIDFromHex(empresaID)
		if err == nil {
			filter["empresaId"] = empresaObjectID
		}
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return []*models.AlertaModel{}, 0, err
	}

	opts := options.Find()
	opts.SetSkip(int64(params.GetSkip()))
	opts.SetLimit(int64(params.GetLimit()))
	
	if params.SortBy != "" {
		sortOrder := 1
		if params.SortDir == "desc" {
			sortOrder = -1
		}
		opts.SetSort(bson.D{{Key: params.SortBy, Value: sortOrder}})
	} else {
		opts.SetSort(bson.D{{Key: "fechaCreacion", Value: -1}})
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return []*models.AlertaModel{}, 0, err
	}
	defer cursor.Close(ctx)

	var entities []*entities.AlertaEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return []*models.AlertaModel{}, 0, err
	}

	if entities == nil {
		return []*models.AlertaModel{}, total, nil
	}

	alertas := make([]*models.AlertaModel, len(entities))
	for i, entity := range entities {
		alertas[i] = r.entityToModel(entity)
	}

	return alertas, total, nil
}

func (r *AlertaRepository) FindActivas(ctx context.Context) ([]*models.AlertaModel, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"resuelta": false})
	if err != nil {
		return []*models.AlertaModel{}, nil
	}
	defer cursor.Close(ctx)

	var entities []*entities.AlertaEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return []*models.AlertaModel{}, nil
	}
	if entities == nil {
		return []*models.AlertaModel{}, nil
	}

	alertas := make([]*models.AlertaModel, len(entities))
	for i, entity := range entities {
		alertas[i] = r.entityToModel(entity)
	}
	return alertas, nil
}

func (r *AlertaRepository) FindByEmpresa(ctx context.Context, empresaID string) ([]*models.AlertaModel, error) {
	empresaObjectID, err := primitive.ObjectIDFromHex(empresaID)
	if err != nil {
		return []*models.AlertaModel{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{"empresaId": empresaObjectID})
	if err != nil {
		return []*models.AlertaModel{}, nil
	}
	defer cursor.Close(ctx)

	var entities []*entities.AlertaEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return []*models.AlertaModel{}, nil
	}
	if entities == nil {
		return []*models.AlertaModel{}, nil
	}

	alertas := make([]*models.AlertaModel, len(entities))
	for i, entity := range entities {
		alertas[i] = r.entityToModel(entity)
	}
	return alertas, nil
}

func (r *AlertaRepository) FindByID(ctx context.Context, id string) (*models.AlertaModel, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, types.ThrowData("ID inválido")
	}

	var entity entities.AlertaEntity
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Alerta no encontrada")
		}
		return nil, err
	}
	return r.entityToModel(&entity), nil
}

func (r *AlertaRepository) Create(ctx context.Context, model *models.AlertaModel) error {
	entity := r.modelToEntity(model)
	entity.FechaCreacion = time.Now()
	entity.Resuelta = false

	result, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return err
	}

	model.ID = result.InsertedID.(primitive.ObjectID).Hex()
	model.FechaCreacion = entity.FechaCreacion
	model.Resuelta = false
	return nil
}

func (r *AlertaRepository) Resolver(ctx context.Context, id string, resolucion string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	now := time.Now()
	_, err = r.collection.UpdateOne(
ctx,
bson.M{"_id": objectID},
bson.M{
"$set": bson.M{
"resuelta":        true,
"resolucion":      resolucion,
"fechaResolucion": now,
},
},
)
	return err
}

func (r *AlertaRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
