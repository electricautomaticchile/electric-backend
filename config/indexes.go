package config

import (
	"context"
	"electric-backend/infrastructure/logger"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []struct {
		collection string
		indexes    []mongo.IndexModel
	}{
		{
			collection: "clientes",
			indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "empresaId", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "numeroCliente", Value: 1},
					},
					Options: options.Index().SetUnique(true),
				},
				{
					Keys: bson.D{
						{Key: "rut", Value: 1},
					},
					Options: options.Index().SetUnique(true),
				},
				{
					Keys: bson.D{
						{Key: "correo", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "empresaId", Value: 1},
						{Key: "estado", Value: 1},
					},
				},
			},
		},
		{
			collection: "dispositivos",
			indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "empresaId", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "clienteId", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "numeroDispositivo", Value: 1},
					},
					Options: options.Index().SetUnique(true),
				},
				{
					Keys: bson.D{
						{Key: "empresaId", Value: 1},
						{Key: "estado", Value: 1},
					},
				},
			},
		},
		{
			collection: "boletas",
			indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "clienteId", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "clienteId", Value: 1},
						{Key: "estado", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "fechaVencimiento", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "numeroBoleta", Value: 1},
					},
					Options: options.Index().SetUnique(true),
				},
			},
		},
		{
			collection: "alertas",
			indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "empresaId", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "dispositivoId", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "empresaId", Value: 1},
						{Key: "estado", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "fechaCreacion", Value: -1},
					},
				},
			},
		},
		{
			collection: "tickets",
			indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "clienteId", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "empresaId", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "estado", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "fechaCreacion", Value: -1},
					},
				},
			},
		},
		{
			collection: "empresas",
			indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "rut", Value: 1},
					},
					Options: options.Index().SetUnique(true),
				},
				{
					Keys: bson.D{
						{Key: "email", Value: 1},
					},
					Options: options.Index().SetUnique(true),
				},
			},
		},
		{
			collection: "usuarios_empresa",
			indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "empresaId", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "email", Value: 1},
					},
					Options: options.Index().SetUnique(true),
				},
				{
					Keys: bson.D{
						{Key: "empresaId", Value: 1},
						{Key: "rol", Value: 1},
					},
				},
			},
		},
		{
			collection: "lecturas",
			indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "dispositivoId", Value: 1},
						{Key: "timestamp", Value: -1},
					},
				},
				{
					Keys: bson.D{
						{Key: "clienteId", Value: 1},
						{Key: "timestamp", Value: -1},
					},
				},
			},
		},
		{
			collection: "notificaciones",
			indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "destinatarioId", Value: 1},
						{Key: "leida", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "fechaCreacion", Value: -1},
					},
				},
			},
		},
		{
			collection: "refresh_tokens",
			indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "userId", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "expiresAt", Value: 1},
					},
					Options: options.Index().SetExpireAfterSeconds(0),
				},
			},
		},
		{
			collection: "recovery_tokens",
			indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "email", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "expiresAt", Value: 1},
					},
					Options: options.Index().SetExpireAfterSeconds(0),
				},
			},
		},
		{
			collection: "audit_logs",
			indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "timestamp", Value: -1},
					},
				},
				{
					Keys: bson.D{
						{Key: "userId", Value: 1},
						{Key: "timestamp", Value: -1},
					},
				},
				{
					Keys: bson.D{
						{Key: "ipAddress", Value: 1},
						{Key: "success", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "timestamp", Value: 1},
					},
					Options: options.Index().SetExpireAfterSeconds(7776000), // TTL 90 días
				},
			},
		},
		{
			collection: "tarifas",
			indexes: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "distribuidora", Value: 1},
						{Key: "comuna", Value: 1},
						{Key: "activa", Value: 1},
					},
				},
				{
					Keys: bson.D{
						{Key: "vigenciaDesde", Value: 1},
						{Key: "vigenciaHasta", Value: 1},
					},
				},
			},
		},
	}

	for _, idx := range indexes {
		collection := db.Collection(idx.collection)
		
		_, err := collection.Indexes().CreateMany(ctx, idx.indexes)
		if err != nil {
			logger.Error().Str("collection", idx.collection).Err(err).Msg("Error creando índices")
			return err
		}
		
		logger.Info().Str("collection", idx.collection).Msg("Índices creados para colección")
	}

	logger.Info().Msg("Todos los índices de MongoDB creados exitosamente")
	return nil
}
