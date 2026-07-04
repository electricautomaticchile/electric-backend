package services

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/email"
	"electric-backend/infrastructure/entities"
	"electric-backend/infrastructure/validation"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TicketService struct {
	ticketRepo       ports.PortTicket
	notificacionRepo ports.PortNotificacion
	emailService     email.EmailService
	clienteRepo      ports.PortCliente
	empresaRepo      ports.PortEmpresa
}

func NewTicketService(ticketRepo ports.PortTicket, notificacionRepo ports.PortNotificacion, emailService email.EmailService, clienteRepo ports.PortCliente, empresaRepo ports.PortEmpresa) *TicketService {
	return &TicketService{
		ticketRepo:       ticketRepo,
		notificacionRepo: notificacionRepo,
		emailService:     emailService,
		clienteRepo:      clienteRepo,
		empresaRepo:      empresaRepo,
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
	prioridad := r.Prioridad
	if prioridad == "" {
		prioridad = "media"
	}

	categoria := r.Categoria
	if categoria == "" {
		categoria = "general"
	}

	entity := &entities.TicketEntity{
		Titulo:      validation.SanitizeString(r.Asunto),
		Descripcion: validation.SanitizeString(r.Descripcion),
		Prioridad:   prioridad,
		Categoria:   categoria,
	}

	if r.ClienteID != "" {
		clienteID, _ := primitive.ObjectIDFromHex(r.ClienteID)
		entity.ClienteID = clienteID
	}

	if r.EmpresaID != "" {
		empresaID, _ := primitive.ObjectIDFromHex(r.EmpresaID)
		entity.EmpresaID = empresaID
	}

	if err := s.ticketRepo.Create(ctx, entity); err != nil {
		return nil, err
	}

	if !entity.EmpresaID.IsZero() {
		notificacion := &entities.NotificacionEntity{
			DestinatarioID: entity.EmpresaID,
			Tipo:           "ticket",
			Titulo:         "Nuevo Ticket Recibido",
			Mensaje:        "Se ha creado un nuevo ticket #" + entity.NumeroTicket + ": " + entity.Titulo,
			Leida:          false,
			FechaCreacion:  time.Now(),
		}
		s.notificacionRepo.Create(ctx, notificacion)
	}

	return s.entityToModel(entity), nil
}

func (s *TicketService) AgregarRespuesta(ctx context.Context, id string, r *recipe.AgregarRespuestaRecipe, usuarioID string) error {
	respuesta := &entities.RespuestaTicket{
		Mensaje:   validation.SanitizeString(r.Mensaje),
		UsuarioID: usuarioID,
	}

	if err := s.ticketRepo.AgregarRespuesta(ctx, id, respuesta); err != nil {
		return err
	}

	ticket, err := s.ticketRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	var destinatarioID primitive.ObjectID
	if usuarioID == ticket.ClienteID.Hex() {
		destinatarioID = ticket.EmpresaID
	} else {
		destinatarioID = ticket.ClienteID
	}

	if !destinatarioID.IsZero() {
		notificacion := &entities.NotificacionEntity{
			DestinatarioID: destinatarioID,
			Tipo:           "ticket",
			Titulo:         "Nueva Respuesta en Ticket",
			Mensaje:        "Hay una nueva respuesta en el ticket #" + ticket.NumeroTicket,
			Leida:          false,
			FechaCreacion:  time.Now(),
		}
		s.notificacionRepo.Create(ctx, notificacion)

		go s.enviarEmailRespuestaTicket(ctx, destinatarioID.Hex(), ticket.NumeroTicket, ticket.Titulo, r.Mensaje)
	}

	return nil
}

func (s *TicketService) enviarEmailRespuestaTicket(ctx context.Context, destinatarioID, numeroTicket, asunto, mensaje string) {
	cliente, err := s.clienteRepo.FindByID(ctx, destinatarioID)
	if err == nil && cliente.Correo != "" {
		err = s.emailService.EnviarNotificacionTicket(
			cliente.Correo,
			cliente.Nombre,
			numeroTicket,
			asunto,
			mensaje,
		)
		if err != nil {
			log.Printf("Error enviando email de ticket: %v", err)
		}
		return
	}

	empresa, err := s.empresaRepo.FindByID(ctx, destinatarioID)
	if err == nil && empresa.Correo != "" {
		err = s.emailService.EnviarNotificacionTicket(
			empresa.Correo,
			empresa.NombreEmpresa,
			numeroTicket,
			asunto,
			mensaje,
		)
		if err != nil {
			log.Printf("Error enviando email de ticket: %v", err)
		}
	}
}

func (s *TicketService) ActualizarEstado(ctx context.Context, id string, r *recipe.ActualizarEstadoTicketRecipe) error {
	if err := s.ticketRepo.ActualizarEstado(ctx, id, r.Estado); err != nil {
		return err
	}

	ticket, err := s.ticketRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	estadoTexto := ""
	switch r.Estado {
	case "abierto":
		estadoTexto = "abierto"
	case "en_proceso":
		estadoTexto = "en proceso"
	case "resuelto":
		estadoTexto = "resuelto"
	case "cerrado":
		estadoTexto = "cerrado"
	default:
		estadoTexto = r.Estado
	}

	notificacion := &entities.NotificacionEntity{
		DestinatarioID: ticket.ClienteID,
		Tipo:           "ticket",
		Titulo:         "Estado de Ticket Actualizado",
		Mensaje:        "Tu ticket #" + ticket.NumeroTicket + " ha cambiado a estado: " + estadoTexto,
		Leida:          false,
		FechaCreacion:  time.Now(),
	}
	s.notificacionRepo.Create(ctx, notificacion)

	return nil
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
		Categoria:     entity.Categoria,
		Respuestas:    respuestas,
		FechaCreacion: entity.FechaCreacion,
	}

	if !entity.ClienteID.IsZero() {
		model.ClienteID = entity.ClienteID.Hex()
	}

	if !entity.EmpresaID.IsZero() {
		model.EmpresaID = entity.EmpresaID.Hex()
	}

	return model
}

func (s *TicketService) ObtenerPorCliente(ctx context.Context, clienteID string) ([]*models.TicketModel, error) {
	tickets, err := s.ticketRepo.FindByCliente(ctx, clienteID)
	if err != nil {
		return []*models.TicketModel{}, nil
	}

	models := make([]*models.TicketModel, len(tickets))
	for i, ticket := range tickets {
		models[i] = s.entityToModel(ticket)
	}

	return models, nil
}

func (s *TicketService) ObtenerPorEmpresa(ctx context.Context, empresaID string) ([]*models.TicketModel, error) {
	tickets, err := s.ticketRepo.FindByEmpresa(ctx, empresaID)
	if err != nil {
		return []*models.TicketModel{}, nil
	}

	models := make([]*models.TicketModel, len(tickets))
	for i, ticket := range tickets {
		models[i] = s.entityToModel(ticket)
	}

	return models, nil
}
