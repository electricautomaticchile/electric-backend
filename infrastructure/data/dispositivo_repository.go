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

type DispositivoRepository struct {
	collection *mongo.Collection
}

func NewDispositivoRepository() *DispositivoRepository {
	return &DispositivoRepository{
		collection: config.MongoDB.Collection(entities.DispositivoEntity{}.CollectionName()),
	}
}

func (r *DispositivoRepository) FindAll(ctx context.Context, empresaID string) ([]*entities.DispositivoEntity, error) {
	empresaObjectID, err := primitive.ObjectIDFromHex(empresaID)
	if err != nil {
		return []*entities.DispositivoEntity{}, nil
	}

	filter := bson.M{"empresaId": empresaObjectID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return []*entities.DispositivoEntity{}, nil
	}
	defer cursor.Close(ctx)

	var dispositivos []*entities.DispositivoEntity
	if err := cursor.All(ctx, &dispositivos); err != nil {
		return []*entities.DispositivoEntity{}, nil
	}

	if dispositivos == nil {
		return []*entities.DispositivoEntity{}, nil
	}

	return dispositivos, nil
}

func (r *DispositivoRepository) FindAllPaginated(ctx context.Context, empresaID string, params types.PaginationParams, filters types.FilterParams) ([]*entities.DispositivoEntity, int64, error) {
	filter := filters.BuildMongoFilter()

	if empresaID != "" {
		empresaObjectID, err := primitive.ObjectIDFromHex(empresaID)
		if err == nil {
			filter["empresaId"] = empresaObjectID
		}
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return []*entities.DispositivoEntity{}, 0, err
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
		opts.SetSort(bson.D{{Key: "fechaInstalacion", Value: -1}})
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return []*entities.DispositivoEntity{}, 0, err
	}
	defer cursor.Close(ctx)

	var dispositivos []*entities.DispositivoEntity
	if err := cursor.All(ctx, &dispositivos); err != nil {
		return []*entities.DispositivoEntity{}, 0, err
	}

	if dispositivos == nil {
		return []*entities.DispositivoEntity{}, total, nil
	}

	return dispositivos, total, nil
}

func (r *DispositivoRepository) FindByID(ctx context.Context, id string) (*entities.DispositivoEntity, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, types.ThrowData("ID inválido")
	}

	var dispositivo entities.DispositivoEntity
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&dispositivo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Dispositivo no encontrado")
		}
		return nil, err
	}
	return &dispositivo, nil
}

func (r *DispositivoRepository) FindByNumero(ctx context.Context, numeroDispositivo string) (*entities.DispositivoEntity, error) {
	var dispositivo entities.DispositivoEntity
	err := r.collection.FindOne(ctx, bson.M{"numeroDispositivo": numeroDispositivo}).Decode(&dispositivo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Dispositivo no encontrado")
		}
		return nil, err
	}
	return &dispositivo, nil
}

func (r *DispositivoRepository) FindByCliente(ctx context.Context, clienteID string) ([]*entities.DispositivoEntity, error) {
	clienteObjectID, err := primitive.ObjectIDFromHex(clienteID)
	if err != nil {
		return []*entities.DispositivoEntity{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{"clienteId": clienteObjectID})
	if err != nil {
		return []*entities.DispositivoEntity{}, nil
	}
	defer cursor.Close(ctx)

	var dispositivos []*entities.DispositivoEntity
	if err := cursor.All(ctx, &dispositivos); err != nil {
		return []*entities.DispositivoEntity{}, nil
	}

	if dispositivos == nil {
		return []*entities.DispositivoEntity{}, nil
	}

	return dispositivos, nil
}

func (r *DispositivoRepository) Create(ctx context.Context, dispositivo *entities.DispositivoEntity) error {
	dispositivo.FechaCreacion = time.Now()
	dispositivo.Activo = true
	dispositivo.Estado = "activo"

	result, err := r.collection.InsertOne(ctx, dispositivo)
	if err != nil {
		return err
	}

	dispositivo.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *DispositivoRepository) Update(ctx context.Context, id string, dispositivo *entities.DispositivoEntity) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	now := time.Now()
	dispositivo.FechaActualizacion = &now

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": dispositivo},
	)
	return err
}

func (r *DispositivoRepository) UpdateUltimaLectura(ctx context.Context, numeroDispositivo string, lectura *entities.LecturaDispositivo) (*entities.DispositivoEntity, error) {
	now := time.Now()
	var updated entities.DispositivoEntity
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"numeroDispositivo": numeroDispositivo},
		bson.M{
			"$set": bson.M{
				"ultimaLectura":      lectura,
				"fechaActualizacion": now,
			},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &updated, nil
}

func (r *DispositivoRepository) CambiarEstado(ctx context.Context, id string, estado string) error {
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
				"estado":             estado,
				"fechaActualizacion": now,
			},
		},
	)
	return err
}

func (r *DispositivoRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

func (r *DispositivoRepository) ObtenerPorCliente(ctx context.Context, clienteID string) ([]*entities.DispositivoEntity, error) {
	return r.FindByCliente(ctx, clienteID)
}
