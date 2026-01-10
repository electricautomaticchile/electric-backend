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
)

type EmpresaRepository struct {
	collection *mongo.Collection
}

func NewEmpresaRepository() *EmpresaRepository {
	return &EmpresaRepository{
		collection: config.MongoDB.Collection(entities.EmpresaEntity{}.CollectionName()),
	}
}

func (r *EmpresaRepository) entityToModel(entity *entities.EmpresaEntity) *models.EmpresaModel {
	var fechaCreacion time.Time
	if fc, ok := entity.FechaCreacion.(int64); ok {
		fechaCreacion = time.UnixMilli(fc)
	}

	return &models.EmpresaModel{
		ID:            entity.ID.Hex(),
		NombreEmpresa: entity.NombreEmpresa,
		RazonSocial:   entity.RazonSocial,
		Rut:           entity.RUT,
		Correo:        entity.Correo,
		Telefono:      entity.Telefono,
		Direccion:     entity.Direccion,
		Ciudad:        entity.Ciudad,
		Region:        entity.Region,
		ContactoPrincipal: models.ContactoPrincipal{
			Nombre:       entity.ContactoPrincipal.Nombre,
			Cargo:        entity.ContactoPrincipal.Cargo,
			Telefono:     entity.ContactoPrincipal.Telefono,
			Correo:       entity.ContactoPrincipal.Correo,
			ImagenPerfil: entity.ContactoPrincipal.ImagenPerfil,
		},
		NumeroCliente:    entity.NumeroCliente,
		Password:         entity.Password,
		PasswordTemporal: entity.PasswordTemporal,
		Role:             entity.Role,
		TipoUsuario:      entity.TipoUsuario,
		Estado:           entity.Estado,
		FechaCreacion:    fechaCreacion,
	}
}

func (r *EmpresaRepository) modelToEntity(model *models.EmpresaModel) *entities.EmpresaEntity {
	entity := &entities.EmpresaEntity{
		NombreEmpresa: model.NombreEmpresa,
		RazonSocial:   model.RazonSocial,
		Correo:        model.Correo,
		Telefono:      model.Telefono,
		Direccion:     model.Direccion,
		Ciudad:        model.Ciudad,
		Region:        model.Region,
		RUT:           model.Rut,
		ContactoPrincipal: entities.ContactoPrincipalEntity{
			Nombre:       model.ContactoPrincipal.Nombre,
			Cargo:        model.ContactoPrincipal.Cargo,
			Telefono:     model.ContactoPrincipal.Telefono,
			Correo:       model.ContactoPrincipal.Correo,
			ImagenPerfil: model.ContactoPrincipal.ImagenPerfil,
		},
		NumeroCliente:    model.NumeroCliente,
		Password:         model.Password,
		PasswordTemporal: model.PasswordTemporal,
		Role:             model.Role,
		TipoUsuario:      model.TipoUsuario,
		Estado:           model.Estado,
	}

	if model.ID != "" {
		if oid, err := primitive.ObjectIDFromHex(model.ID); err == nil {
			entity.ID = oid
		}
	}

	return entity
}

func (r *EmpresaRepository) FindAll(ctx context.Context) ([]*models.EmpresaModel, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return []*models.EmpresaModel{}, nil
	}
	defer cursor.Close(ctx)

	var entities []*entities.EmpresaEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return []*models.EmpresaModel{}, nil
	}

	if entities == nil {
		return []*models.EmpresaModel{}, nil
	}

	empresas := make([]*models.EmpresaModel, len(entities))
	for i, entity := range entities {
		empresas[i] = r.entityToModel(entity)
	}

	return empresas, nil
}

func (r *EmpresaRepository) FindByID(ctx context.Context, id string) (*models.EmpresaModel, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, types.ThrowData("ID inválido")
	}

	var entity entities.EmpresaEntity
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Empresa no encontrada")
		}
		return nil, err
	}
	return r.entityToModel(&entity), nil
}

func (r *EmpresaRepository) FindByNumeroCliente(ctx context.Context, numeroCliente string) (*models.EmpresaModel, error) {
	var entity entities.EmpresaEntity
	err := r.collection.FindOne(ctx, bson.M{"numeroCliente": numeroCliente}).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Empresa no encontrada")
		}
		return nil, err
	}
	return r.entityToModel(&entity), nil
}

func (r *EmpresaRepository) FindByCorreo(ctx context.Context, correo string) (*models.EmpresaModel, error) {
	var entity entities.EmpresaEntity
	err := r.collection.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"correo": correo},
			{"email": correo},
		},
	}).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Empresa no encontrada")
		}
		return nil, err
	}
	return r.entityToModel(&entity), nil
}

func (r *EmpresaRepository) Create(ctx context.Context, model *models.EmpresaModel) error {
	entity := r.modelToEntity(model)
	entity.FechaCreacion = time.Now().UnixMilli()
	entity.Activo = true

	result, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return err
	}

	model.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *EmpresaRepository) Update(ctx context.Context, id string, model *models.EmpresaModel) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	entity := r.modelToEntity(model)
	now := time.Now().UnixMilli()
	entity.FechaActualizacion = &now

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": entity})
	return err
}

func (r *EmpresaRepository) UpdateUltimoAcceso(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	now := time.Now().UnixMilli()
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": bson.M{"ultimoAcceso": now}})
	return err
}

func (r *EmpresaRepository) UpdatePassword(ctx context.Context, id string, hashedPassword string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	now := time.Now().UnixMilli()
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": bson.M{"password": hashedPassword, "passwordTemporal": false, "fechaActualizacion": now}})
	return err
}

func (r *EmpresaRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

func (r *EmpresaRepository) CambiarEstado(ctx context.Context, id string, activo bool) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	estado := "inactivo"
	if activo {
		estado = "activo"
	}

	now := time.Now().UnixMilli()
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": bson.M{"estado": estado, "activo": activo, "fechaActualizacion": now}})
	return err
}
