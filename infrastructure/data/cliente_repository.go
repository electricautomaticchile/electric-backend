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

type ClienteRepository struct {
	collection *mongo.Collection
}

func NewClienteRepository() *ClienteRepository {
	return &ClienteRepository{
		collection: config.MongoDB.Collection(entities.ClienteEntity{}.CollectionName()),
	}
}

func (r *ClienteRepository) entityToModel(entity *entities.ClienteEntity) *models.ClienteModel {
	var fechaCreacion time.Time
	if fc, ok := entity.FechaCreacion.(int64); ok {
		fechaCreacion = time.UnixMilli(fc)
	} else if fc, ok := entity.FechaCreacion.(time.Time); ok {
		fechaCreacion = fc
	}

	var fechaActivacion *time.Time
	if entity.FechaActualizacion != nil {
		if fa, ok := entity.FechaActualizacion.(int64); ok {
			t := time.UnixMilli(fa)
			fechaActivacion = &t
		} else if fa, ok := entity.FechaActualizacion.(time.Time); ok {
			fechaActivacion = &fa
		}
	}

	var ultimoAcceso *time.Time
	if entity.UltimoAcceso != nil {
		if ua, ok := entity.UltimoAcceso.(int64); ok {
			t := time.UnixMilli(ua)
			ultimoAcceso = &t
		} else if ua, ok := entity.UltimoAcceso.(time.Time); ok {
			ultimoAcceso = &ua
		}
	}

	correo := entity.Correo

	passwordTemporal := ""
	if entity.PasswordTemporal {
		passwordTemporal = "true"
	}

	model := &models.ClienteModel{
		ID:               entity.ID.Hex(),
		Nombre:           entity.Nombre,
		Correo:           correo,
		NumeroCliente:    entity.NumeroCliente,
		Password:         entity.Password,
		PasswordTemporal: passwordTemporal,
		Telefono:         entity.Telefono,
		Direccion:        entity.Direccion,
		Ciudad:           entity.Ciudad,
		Rut:              entity.Rut,
		ImagenPerfil:     entity.ImagenPerfil,
		TipoCliente:      entity.TipoCliente,
		Empresa:          entity.Empresa,
		Role:             entity.Role,
		TipoUsuario:      entity.TipoUsuario,
		Activo:           entity.Activo,
		FechaRegistro:    fechaCreacion,
		FechaActivacion:  fechaActivacion,
		UltimoAcceso:     ultimoAcceso,
	}

	if !entity.EmpresaID.IsZero() {
		model.EmpresaID = entity.EmpresaID.Hex()
	}

	if !entity.UsuarioID.IsZero() {
		model.UsuarioID = entity.UsuarioID.Hex()
	}

	return model
}

func (r *ClienteRepository) modelToEntity(model *models.ClienteModel) *entities.ClienteEntity {
	entity := &entities.ClienteEntity{
		Nombre:           model.Nombre,
		Correo:           model.Correo,
		NumeroCliente:    model.NumeroCliente,
		Password:         model.Password,
		PasswordTemporal: model.PasswordTemporal != "",
		Telefono:         model.Telefono,
		Direccion:        model.Direccion,
		Ciudad:           model.Ciudad,
		Rut:              model.Rut,
		ImagenPerfil:     model.ImagenPerfil,
		TipoCliente:      model.TipoCliente,
		Empresa:          model.Empresa,
		Role:             model.Role,
		TipoUsuario:      model.TipoUsuario,
		Activo:           model.Activo,
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

	if model.UsuarioID != "" {
		if oid, err := primitive.ObjectIDFromHex(model.UsuarioID); err == nil {
			entity.UsuarioID = oid
		}
	}

	return entity
}

func (r *ClienteRepository) FindAll(ctx context.Context, empresaID string) ([]*models.ClienteModel, error) {
	empresaObjectID, err := primitive.ObjectIDFromHex(empresaID)
	if err != nil {
		return []*models.ClienteModel{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{"empresaId": empresaObjectID})
	if err != nil {
		return []*models.ClienteModel{}, nil
	}
	defer cursor.Close(ctx)

	var entities []*entities.ClienteEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return []*models.ClienteModel{}, nil
	}

	if entities == nil {
		return []*models.ClienteModel{}, nil
	}

	clientes := make([]*models.ClienteModel, len(entities))
	for i, entity := range entities {
		clientes[i] = r.entityToModel(entity)
	}

	return clientes, nil
}

func (r *ClienteRepository) FindAllPaginated(ctx context.Context, empresaID string, params types.PaginationParams, filters types.FilterParams) ([]*models.ClienteModel, int64, error) {
	filter := filters.BuildMongoFilter()

	if empresaID != "" {
		empresaObjectID, err := primitive.ObjectIDFromHex(empresaID)
		if err == nil {
			filter["empresaId"] = empresaObjectID
		}
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return []*models.ClienteModel{}, 0, err
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
		return []*models.ClienteModel{}, 0, err
	}
	defer cursor.Close(ctx)

	var entities []*entities.ClienteEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return []*models.ClienteModel{}, 0, err
	}

	if entities == nil {
		return []*models.ClienteModel{}, total, nil
	}

	clientes := make([]*models.ClienteModel, len(entities))
	for i, entity := range entities {
		clientes[i] = r.entityToModel(entity)
	}

	return clientes, total, nil
}

func (r *ClienteRepository) FindByID(ctx context.Context, id string) (*models.ClienteModel, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, types.ThrowData("ID inválido")
	}

	var entity entities.ClienteEntity
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Cliente no encontrado")
		}
		return nil, err
	}
	return r.entityToModel(&entity), nil
}

func (r *ClienteRepository) FindByNumeroCliente(ctx context.Context, numeroCliente string) (*models.ClienteModel, error) {
	var entity entities.ClienteEntity
	err := r.collection.FindOne(ctx, bson.M{"numeroCliente": numeroCliente}).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Cliente no encontrado")
		}
		return nil, err
	}
	return r.entityToModel(&entity), nil
}

func (r *ClienteRepository) FindByRut(ctx context.Context, rut string) (*models.ClienteModel, error) {
	var entity entities.ClienteEntity
	err := r.collection.FindOne(ctx, bson.M{"rut": rut}).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Cliente no encontrado")
		}
		return nil, err
	}
	return r.entityToModel(&entity), nil
}

func (r *ClienteRepository) FindByCorreo(ctx context.Context, correo string) (*models.ClienteModel, error) {
	var entity entities.ClienteEntity
	err := r.collection.FindOne(ctx, bson.M{"correo": correo}).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Cliente no encontrado")
		}
		return nil, err
	}
	return r.entityToModel(&entity), nil
}

func (r *ClienteRepository) Create(ctx context.Context, model *models.ClienteModel) error {
	entity := r.modelToEntity(model)
	entity.FechaCreacion = time.Now().UnixMilli()
	entity.Activo = true

	result, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return err
	}

	model.ID = result.InsertedID.(primitive.ObjectID).Hex()
	model.Activo = true
	model.FechaRegistro = time.UnixMilli(entity.FechaCreacion.(int64))
	return nil
}

func (r *ClienteRepository) Update(ctx context.Context, id string, model *models.ClienteModel) error {
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

func (r *ClienteRepository) UpdateUltimoAcceso(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	now := time.Now().UnixMilli()
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"ultimoAcceso": now}},
	)
	return err
}

func (r *ClienteRepository) UpdatePassword(ctx context.Context, id string, hashedPassword string) error {
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
				"password":           hashedPassword,
				"passwordTemporal":   false,
				"fechaActualizacion": now,
			},
		},
	)
	return err
}

func (r *ClienteRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
