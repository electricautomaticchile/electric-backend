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

type UsuarioEmpresaRepository struct{}

func NewUsuarioEmpresaRepository() *UsuarioEmpresaRepository {
	return &UsuarioEmpresaRepository{}
}

func (r *UsuarioEmpresaRepository) Create(ctx context.Context, usuario *models.UsuarioEmpresaModel) error {
	collection := config.MongoDB.Collection("usuarios_empresa")

	empresaID, _ := primitive.ObjectIDFromHex(usuario.EmpresaID)

	entity := &entities.UsuarioEmpresaEntity{
		EmpresaID:          empresaID,
		Nombre:             usuario.Nombre,
		Email:              usuario.Email,
		Password:           usuario.Password,
		Role:               usuario.Role,
		Telefono:           usuario.Telefono,
		Cargo:              usuario.Cargo,
		Activo:             true,
		PasswordTemporal:   usuario.PasswordTemporal,
		FechaCreacion:      time.Now(),
		FechaActualizacion: time.Now(),
	}

	result, err := collection.InsertOne(ctx, entity)
	if err != nil {
		return err
	}

	usuario.ID = result.InsertedID.(primitive.ObjectID).Hex()
	usuario.FechaCreacion = entity.FechaCreacion
	usuario.FechaActualizacion = entity.FechaActualizacion
	usuario.Activo = true

	return nil
}

func (r *UsuarioEmpresaRepository) FindAll(ctx context.Context, empresaID string) ([]*models.UsuarioEmpresaModel, error) {
	collection := config.MongoDB.Collection("usuarios_empresa")

	empresaOID, _ := primitive.ObjectIDFromHex(empresaID)

	cursor, err := collection.Find(ctx, bson.M{"empresaId": empresaOID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []*entities.UsuarioEmpresaEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}

	usuarios := make([]*models.UsuarioEmpresaModel, len(entities))
	for i, entity := range entities {
		usuarios[i] = r.entityToModel(entity)
	}

	return usuarios, nil
}

func (r *UsuarioEmpresaRepository) FindByID(ctx context.Context, id string) (*models.UsuarioEmpresaModel, error) {
	collection := config.MongoDB.Collection("usuarios_empresa")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var entity entities.UsuarioEmpresaEntity
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&entity)
	if err != nil {
		return nil, err
	}

	return r.entityToModel(&entity), nil
}

func (r *UsuarioEmpresaRepository) FindByEmail(ctx context.Context, email string) (*models.UsuarioEmpresaModel, error) {
	collection := config.MongoDB.Collection("usuarios_empresa")

	var entity entities.UsuarioEmpresaEntity
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&entity)
	if err != nil {
		return nil, err
	}

	return r.entityToModel(&entity), nil
}

func (r *UsuarioEmpresaRepository) Update(ctx context.Context, id string, usuario *models.UsuarioEmpresaModel) error {
	collection := config.MongoDB.Collection("usuarios_empresa")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	empresaID, _ := primitive.ObjectIDFromHex(usuario.EmpresaID)

	update := bson.M{
		"$set": bson.M{
			"empresaId":          empresaID,
			"nombre":             usuario.Nombre,
			"email":              usuario.Email,
			"role":               usuario.Role,
			"telefono":           usuario.Telefono,
			"cargo":              usuario.Cargo,
			"activo":             usuario.Activo,
			"fechaActualizacion": time.Now(),
		},
	}

	if usuario.Password != "" {
		update["$set"].(bson.M)["password"] = usuario.Password
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *UsuarioEmpresaRepository) Delete(ctx context.Context, id string) error {
	collection := config.MongoDB.Collection("usuarios_empresa")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

func (r *UsuarioEmpresaRepository) UpdateUltimoAcceso(ctx context.Context, id string) error {
	collection := config.MongoDB.Collection("usuarios_empresa")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"ultimoAcceso": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *UsuarioEmpresaRepository) entityToModel(entity *entities.UsuarioEmpresaEntity) *models.UsuarioEmpresaModel {
	return &models.UsuarioEmpresaModel{
		ID:                 entity.ID.Hex(),
		EmpresaID:          entity.EmpresaID.Hex(),
		Nombre:             entity.Nombre,
		Email:              entity.Email,
		Password:           entity.Password,
		Role:               entity.Role,
		Telefono:           entity.Telefono,
		Cargo:              entity.Cargo,
		Activo:             entity.Activo,
		PasswordTemporal:   entity.PasswordTemporal,
		UltimoAcceso:       entity.UltimoAcceso,
		FechaCreacion:      entity.FechaCreacion,
		FechaActualizacion: entity.FechaActualizacion,
	}
}
