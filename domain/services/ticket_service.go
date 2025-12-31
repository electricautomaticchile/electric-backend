package services

import (
"context"
"electric-backend/api/v1/recipe"
"electric-backend/domain/models"
"electric-backend/domain/ports"
"electric-backend/infrastructure/entities"

"go.mongodb.org/mongo-driver/bson/primitive"
)

type TicketService struct {
	ticketRepo ports.PortTicket
}

func NewTicketService(ticketRepo ports.PortTicket) *TicketService {
	return &TicketService{
		ticketRepo: ticketRepo,
	}
}

func (s *TicketService) ObtenerTodos(ctx context.Context) ([]*models.TicketModel, error) {
	tickets, err := s.ticketRepo.FindAll(ctx)
	if err != nil {
		return []*models.TicketModel{}, nil
	}

	models := make([]*models.TicketModel, len(tickets))
	for i, ticket := range tickets {
		models[i] = s.entityToModel(ticket)
	}

	return models, nil
}

func (s *TicketService) ObtenerPorID(ctx context.Context, id string) (*models.TicketModel, error) {
	ticket, err := s.ticketRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.entityToModel(ticket), nil
}

func (s *TicketService) Crear(ctx context.Context, r *recipe.CrearTicketRecipe) (*models.TicketModel, error) {
	entity := &entities.TicketEntity{
		Titulo:      r.Titulo,
		Descripcion: r.Descripcion,
		Prioridad:   r.Prioridad,
	}

	if r.ClienteID != "" {
		clienteID, _ := primitive.ObjectIDFromHex(r.ClienteID)
		entity.ClienteID = clienteID
	}

	if err := s.ticketRepo.Create(ctx, entity); err != nil {
		return nil, err
	}

	return s.entityToModel(entity), nil
}

func (s *TicketService) AgregarRespuesta(ctx context.Context, id string, r *recipe.AgregarRespuestaRecipe, usuarioID string) error {
	respuesta := &entities.RespuestaTicket{
		Mensaje:   r.Mensaje,
		UsuarioID: usuarioID,
	}

	return s.ticketRepo.AgregarRespuesta(ctx, id, respuesta)
}

func (s *TicketService) ActualizarEstado(ctx context.Context, id string, r *recipe.ActualizarEstadoTicketRecipe) error {
	return s.ticketRepo.ActualizarEstado(ctx, id, r.Estado)
}

func (s *TicketService) Eliminar(ctx context.Context, id string) error {
	return s.ticketRepo.Delete(ctx, id)
}

func (s *TicketService) entityToModel(entity *entities.TicketEntity) *models.TicketModel {
	respuestas := make([]models.RespuestaTicketModel, len(entity.Respuestas))
	for i, r := range entity.Respuestas {
		respuestas[i] = models.RespuestaTicketModel{
			Mensaje:       r.Mensaje,
			UsuarioID:     r.UsuarioID,
			FechaCreacion: r.FechaCreacion,
		}
	}

	model := &models.TicketModel{
		ID:            entity.ID.Hex(),
		NumeroTicket:  entity.NumeroTicket,
		Titulo:        entity.Titulo,
		Descripcion:   entity.Descripcion,
		Estado:        entity.Estado,
		Prioridad:     entity.Prioridad,
		Respuestas:    respuestas,
		FechaCreacion: entity.FechaCreacion,
	}

	if !entity.ClienteID.IsZero() {
		model.ClienteID = entity.ClienteID.Hex()
	}

	return model
}
