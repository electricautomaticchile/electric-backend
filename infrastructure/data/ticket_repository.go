package data

import (
	"context"
	"electric-backend/config"
	"electric-backend/infrastructure/entities"
	"electric-backend/types"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TicketRepository struct {
	collection *mongo.Collection
}

func NewTicketRepository() *TicketRepository {
	return &TicketRepository{
		collection: config.MongoDB.Collection(entities.TicketEntity{}.CollectionName()),
	}
}

func (r *TicketRepository) FindAll(ctx context.Context) ([]*entities.TicketEntity, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return []*entities.TicketEntity{}, nil
	}
	defer cursor.Close(ctx)

	var tickets []*entities.TicketEntity
	if err := cursor.All(ctx, &tickets); err != nil {
		return []*entities.TicketEntity{}, nil
	}
	if tickets == nil {
		return []*entities.TicketEntity{}, nil
	}
	return tickets, nil
}

func (r *TicketRepository) FindByID(ctx context.Context, id string) (*entities.TicketEntity, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, types.ThrowData("ID inválido")
	}

	var ticket entities.TicketEntity
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&ticket)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, types.ThrowData("Ticket no encontrado")
		}
		return nil, err
	}
	return &ticket, nil
}

func (r *TicketRepository) Create(ctx context.Context, ticket *entities.TicketEntity) error {
	ticket.FechaCreacion = time.Now()
	ticket.Estado = "abierto"
	ticket.NumeroTicket = fmt.Sprintf("TKT-%d", time.Now().Unix())
	ticket.Respuestas = []entities.RespuestaTicket{}

	result, err := r.collection.InsertOne(ctx, ticket)
	if err != nil {
		return err
	}

	ticket.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *TicketRepository) AgregarRespuesta(ctx context.Context, id string, respuesta *entities.RespuestaTicket) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	respuesta.FechaCreacion = time.Now()

	_, err = r.collection.UpdateOne(
ctx,
bson.M{"_id": objectID},
bson.M{"$push": bson.M{"respuestas": respuesta}},
)
	return err
}

func (r *TicketRepository) ActualizarEstado(ctx context.Context, id string, estado string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.UpdateOne(
ctx,
bson.M{"_id": objectID},
bson.M{"$set": bson.M{"estado": estado}},
)
	return err
}

func (r *TicketRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return types.ThrowData("ID inválido")
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

func (r *TicketRepository) FindByCliente(ctx context.Context, clienteID string) ([]*entities.TicketEntity, error) {
	objectID, err := primitive.ObjectIDFromHex(clienteID)
	if err != nil {
		return []*entities.TicketEntity{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{"clienteId": objectID})
	if err != nil {
		return []*entities.TicketEntity{}, nil
	}
	defer cursor.Close(ctx)

	var tickets []*entities.TicketEntity
	if err := cursor.All(ctx, &tickets); err != nil {
		return []*entities.TicketEntity{}, nil
	}
	if tickets == nil {
		return []*entities.TicketEntity{}, nil
	}
	return tickets, nil
}

func (r *TicketRepository) FindByEmpresa(ctx context.Context, empresaID string) ([]*entities.TicketEntity, error) {
	objectID, err := primitive.ObjectIDFromHex(empresaID)
	if err != nil {
		return []*entities.TicketEntity{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{"empresaId": objectID})
	if err != nil {
		return []*entities.TicketEntity{}, nil
	}
	defer cursor.Close(ctx)

	var tickets []*entities.TicketEntity
	if err := cursor.All(ctx, &tickets); err != nil {
		return []*entities.TicketEntity{}, nil
	}
	if tickets == nil {
		return []*entities.TicketEntity{}, nil
	}
	return tickets, nil
}
