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

func (r *BoletaRepository) FindByClientePaginated(ctx context.Context, clienteID string, params types.PaginationParams, filters types.FilterParams) ([]*entities.BoletaEntity, int64, error) {
	filter := filters.BuildMongoFilter()
	
	if clienteID != "" {
		clienteObjectID, err := primitive.ObjectIDFromHex(clienteID)
		if err == nil {
			filter["clienteId"] = clienteObjectID
		}
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return []*entities.BoletaEntity{}, 0, err
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
		opts.SetSort(bson.D{{Key: "fechaEmision", Value: -1}})
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return []*entities.BoletaEntity{}, 0, err
	}
	defer cursor.Close(ctx)

	var boletas []*entities.BoletaEntity
	if err := cursor.All(ctx, &boletas); err != nil {
		return []*entities.BoletaEntity{}, 0, err
	}

	if boletas == nil {
		return []*entities.BoletaEntity{}, total, nil
	}

	return boletas, total, nil
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
	if boleta.Estado == "" {
		boleta.Estado = "pendiente"
	}

	result, err := r.collection.InsertOne(ctx, boleta)
	if err != nil {
		return err
	}

	boleta.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *BoletaRepository) Update(ctx context.Context, id string, boleta *entities.BoletaEntity) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.ReplaceOne(ctx, bson.M{"_id": objectID}, boleta)
	return err
}

func (r *BoletaRepository) UpdateEstado(ctx context.Context, id string, estado string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{
		"$set": bson.M{"estado": estado},
	})
	return err
}

func (r *BoletaRepository) UpdateNotificacionEnviada(ctx context.Context, id string, campo string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{
		"$set": bson.M{"notificacionesEnviadas." + campo: true},
	})
	return err
}

func (r *BoletaRepository) FindVencidasByCliente(ctx context.Context, clienteID string) ([]*entities.BoletaEntity, error) {
	clienteObjectID, err := primitive.ObjectIDFromHex(clienteID)
	if err != nil {
		return []*entities.BoletaEntity{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{
		"clienteId": clienteObjectID,
		"estado":    "vencido",
	})
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

func (r *BoletaRepository) FindPendientesByCliente(ctx context.Context, clienteID string) ([]*entities.BoletaEntity, error) {
	clienteObjectID, err := primitive.ObjectIDFromHex(clienteID)
	if err != nil {
		return []*entities.BoletaEntity{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{
		"clienteId": clienteObjectID,
		"estado":    bson.M{"$in": []string{"pendiente", "por_vencer", "vencido"}},
	})
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

func (r *BoletaRepository) FindPorVencer(ctx context.Context, diasAntes int) ([]*entities.BoletaEntity, error) {
	ahora := time.Now()
	limite := ahora.AddDate(0, 0, diasAntes)

	cursor, err := r.collection.Find(ctx, bson.M{
		"estado":           "pendiente",
		"fechaVencimiento": bson.M{"$lte": limite, "$gt": ahora},
	})
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

func (r *BoletaRepository) FindVencidas(ctx context.Context) ([]*entities.BoletaEntity, error) {
	ahora := time.Now()

	cursor, err := r.collection.Find(ctx, bson.M{
		"estado":           bson.M{"$in": []string{"pendiente", "por_vencer"}},
		"fechaVencimiento": bson.M{"$lte": ahora},
	})
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

// FindClienteIDsConBoletasVencidas retorna solo los clienteIDs que tienen boletas vencidas.
// Usa aggregate para evitar traer todos los clientes de la base de datos.
func (r *BoletaRepository) FindClienteIDsConBoletasVencidas(ctx context.Context) ([]string, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"estado": "vencido"}},
		bson.M{"$group": bson.M{
			"_id":   "$clienteId",
			"count": bson.M{"$sum": 1},
		}},
		bson.M{"$match": bson.M{"count": bson.M{"$gte": 2}}}, // Solo clientes con 2+ vencidas (advertencia o corte)
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return []string{}, nil
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return []string{}, nil
	}

	ids := make([]string, len(results))
	for i, r := range results {
		ids[i] = r.ID.Hex()
	}
	return ids, nil
}

func (r *BoletaRepository) ObtenerPorCliente(ctx context.Context, clienteID interface{}) ([]*entities.BoletaEntity, error) {
	var clienteObjectID primitive.ObjectID
	var err error

	switch v := clienteID.(type) {
	case string:
		clienteObjectID, err = primitive.ObjectIDFromHex(v)
		if err != nil {
			return []*entities.BoletaEntity{}, nil
		}
	case primitive.ObjectID:
		clienteObjectID = v
	default:
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
