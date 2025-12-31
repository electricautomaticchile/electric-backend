package data

import (
"context"
"electric-backend/config"
"electric-backend/infrastructure/entities"
"electric-backend/types"
"time"

"go.mongodb.org/mongo-driver/bson"
"go.mongodb.org/mongo-driver/mongo"
)

type ConfiguracionRepository struct {
	collection *mongo.Collection
}

func NewConfiguracionRepository() *ConfiguracionRepository {
	return &ConfiguracionRepository{
		collection: config.MongoDB.Collection(entities.ConfiguracionEntity{}.CollectionName()),
	}
}

func (r *ConfiguracionRepository) FindAll(ctx context.Context) ([]*entities.ConfiguracionEntity, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return []*entities.ConfiguracionEntity{}, nil
	}
	defer cursor.Close(ctx)

	var configuraciones []*entities.ConfiguracionEntity
	if err := cursor.All(ctx, &configuraciones); err != nil {
		return []*entities.ConfiguracionEntity{}, nil
	}
	if configuraciones == nil {
		return []*entities.ConfiguracionEntity{}, nil
	}
	return configuraciones, nil
}

func (r *ConfiguracionRepository) FindByClave(ctx context.Context, clave string) (*entities.ConfiguracionEntity, error) {
	var configuracion entities.ConfiguracionEntity
	err := r.collection.FindOne(ctx, bson.M{"clave": clave}).Decode(&configuracion)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Configuración no encontrada")
		}
		return nil, err
	}
	return &configuracion, nil
}

func (r *ConfiguracionRepository) Actualizar(ctx context.Context, clave string, valor interface{}) error {
	now := time.Now()

	_, err := r.collection.UpdateOne(
ctx,
bson.M{"clave": clave},
bson.M{
"$set": bson.M{
"valor":              valor,
"fechaActualizacion": now,
},
},
nil,
)
	return err
}
