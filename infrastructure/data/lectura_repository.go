package data

import (
	"context"
	"electric-backend/config"
	"electric-backend/infrastructure/entities"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LecturaRepository struct {
	collection    *mongo.Collection
	collectionAgg *mongo.Collection
}

func NewLecturaRepository() *LecturaRepository {
	return &LecturaRepository{
		collection:    config.MongoDB.Collection(entities.LecturaEntity{}.CollectionName()),
		collectionAgg: config.MongoDB.Collection(entities.LecturaAgregadaEntity{}.CollectionName()),
	}
}

func (r *LecturaRepository) GetLecturasAgregadas(ctx context.Context, dispositivoID string, desde, hasta time.Time, limit int64) ([]entities.LecturaAgregadaEntity, error) {
	filter := bson.M{
		"dispositivoId": dispositivoID,
		"timestamp": bson.M{
			"$gte": desde,
			"$lte": hasta,
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collectionAgg.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var lecturas []entities.LecturaAgregadaEntity
	if err := cursor.All(ctx, &lecturas); err != nil {
		return nil, err
	}

	return lecturas, nil
}

func (r *LecturaRepository) GetLecturasPorCliente(ctx context.Context, clienteID string, desde, hasta time.Time, limit int64) ([]entities.LecturaAgregadaEntity, error) {
	filter := bson.M{
		"clienteId": clienteID,
		"timestamp": bson.M{
			"$gte": desde,
			"$lte": hasta,
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collectionAgg.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var lecturas []entities.LecturaAgregadaEntity
	if err := cursor.All(ctx, &lecturas); err != nil {
		return nil, err
	}

	return lecturas, nil
}

func (r *LecturaRepository) GetLecturasPorEmpresa(ctx context.Context, empresaID string, desde, hasta time.Time, limit int64) ([]entities.LecturaAgregadaEntity, error) {
	filter := bson.M{
		"empresaId": empresaID,
		"timestamp": bson.M{
			"$gte": desde,
			"$lte": hasta,
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.collectionAgg.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var lecturas []entities.LecturaAgregadaEntity
	if err := cursor.All(ctx, &lecturas); err != nil {
		return nil, err
	}

	return lecturas, nil
}

func (r *LecturaRepository) GetResumenDiario(ctx context.Context, dispositivoID string, fecha time.Time) (*ResumenConsumo, error) {
	inicio := time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 0, 0, 0, 0, fecha.Location())
	fin := inicio.Add(24 * time.Hour)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"dispositivoId": dispositivoID,
				"timestamp": bson.M{
					"$gte": inicio,
					"$lt":  fin,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"potenciaAvg": bson.M{"$avg": "$potenciaAvg"},
				"potenciaMax": bson.M{"$max": "$potenciaMax"},
				"potenciaMin": bson.M{"$min": "$potenciaMin"},
				"energiaInicio": bson.M{"$first": "$energiaInicio"},
				"energiaFin": bson.M{"$last": "$energiaFin"},
				"costoInicio": bson.M{"$first": "$costoInicio"},
				"costoFin": bson.M{"$last": "$costoFin"},
				"numLecturas": bson.M{"$sum": "$numLecturas"},
			},
		},
	}

	cursor, err := r.collectionAgg.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &ResumenConsumo{}, nil
	}

	result := results[0]
	return &ResumenConsumo{
		PotenciaAvg:    getFloat64(result, "potenciaAvg"),
		PotenciaMax:    getFloat64(result, "potenciaMax"),
		PotenciaMin:    getFloat64(result, "potenciaMin"),
		EnergiaInicio:  getFloat64(result, "energiaInicio"),
		EnergiaFin:     getFloat64(result, "energiaFin"),
		EnergiaTotal:   getFloat64(result, "energiaFin") - getFloat64(result, "energiaInicio"),
		CostoInicio:    getFloat64(result, "costoInicio"),
		CostoFin:       getFloat64(result, "costoFin"),
		CostoTotal:     getFloat64(result, "costoFin") - getFloat64(result, "costoInicio"),
		NumLecturas:    getInt(result, "numLecturas"),
	}, nil
}

type ResumenConsumo struct {
	PotenciaAvg   float64
	PotenciaMax   float64
	PotenciaMin   float64
	EnergiaInicio float64
	EnergiaFin    float64
	EnergiaTotal  float64
	CostoInicio   float64
	CostoFin      float64
	CostoTotal    float64
	NumLecturas   int
}

func getFloat64(m bson.M, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

func getInt(m bson.M, key string) int {
	if v, ok := m[key].(int32); ok {
		return int(v)
	}
	if v, ok := m[key].(int64); ok {
		return int(v)
	}
	return 0
}
