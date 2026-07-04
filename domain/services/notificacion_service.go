package services

import (
	"context"
	"log"
	"time"

	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/entities"
	"electric-backend/infrastructure/validation"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// notificacionEmailSender es la porción mínima del servicio de email que
// necesita NotificacionService para enviar cualquier notificación por correo.
type notificacionEmailSender interface {
	EnviarNotificacion(destinatario, nombreCliente, tipo, titulo, mensaje string) error
}

type NotificacionService struct {
	notificacionRepo ports.PortNotificacion
	wsNotifier       *WebSocketNotifierService
	emailService     notificacionEmailSender
	clienteRepo      ports.PortCliente
	empresaRepo      ports.PortEmpresa
}

func NewNotificacionService(
	notificacionRepo ports.PortNotificacion,
	wsNotifier *WebSocketNotifierService,
	emailService notificacionEmailSender,
	clienteRepo ports.PortCliente,
	empresaRepo ports.PortEmpresa,
) *NotificacionService {
	return &NotificacionService{
		notificacionRepo: notificacionRepo,
		wsNotifier:       wsNotifier,
		emailService:     emailService,
		clienteRepo:      clienteRepo,
		empresaRepo:      empresaRepo,
	}
}

func (s *NotificacionService) Listar(ctx context.Context, destinatarioID string) ([]*models.NotificacionModel, error) {
	notificaciones, err := s.notificacionRepo.FindByDestinatario(ctx, destinatarioID)
	if err != nil {
		return []*models.NotificacionModel{}, nil
	}

	models := make([]*models.NotificacionModel, len(notificaciones))
	for i, notificacion := range notificaciones {
		models[i] = s.entityToModel(notificacion)
	}

	return models, nil
}

func (s *NotificacionService) ListarPorEmpresa(ctx context.Context, empresaID string) ([]*models.NotificacionModel, error) {
	notificaciones, err := s.notificacionRepo.FindByEmpresa(ctx, empresaID)
	if err != nil {
		return []*models.NotificacionModel{}, nil
	}

	models := make([]*models.NotificacionModel, len(notificaciones))
	for i, notificacion := range notificaciones {
		models[i] = s.entityToModel(notificacion)
	}

	return models, nil
}

func (s *NotificacionService) ListarActivas(ctx context.Context, empresaID string) ([]*models.NotificacionModel, error) {
	notificaciones, err := s.notificacionRepo.FindActivas(ctx, empresaID)
	if err != nil {
		return []*models.NotificacionModel{}, nil
	}

	models := make([]*models.NotificacionModel, len(notificaciones))
	for i, notificacion := range notificaciones {
		models[i] = s.entityToModel(notificacion)
	}

	return models, nil
}

func (s *NotificacionService) MarcarLeida(ctx context.Context, id string) error {
	return s.notificacionRepo.MarcarLeida(ctx, id)
}

func (s *NotificacionService) MarcarTodasLeidas(ctx context.Context, destinatarioID string) error {
	return s.notificacionRepo.MarcarTodasLeidas(ctx, destinatarioID)
}

func (s *NotificacionService) Resolver(ctx context.Context, id string, resolucion string) error {
	return s.notificacionRepo.Resolver(ctx, id, resolucion)
}

func (s *NotificacionService) Eliminar(ctx context.Context, id string) error {
	return s.notificacionRepo.Delete(ctx, id)
}

func (s *NotificacionService) Crear(ctx context.Context, r *recipe.CrearNotificacionRecipe) (*models.NotificacionModel, error) {
	destinatarioID, _ := primitive.ObjectIDFromHex(r.DestinatarioID)

	entity := &entities.NotificacionEntity{
		DestinatarioID: destinatarioID,
		Titulo:         validation.SanitizeString(r.Titulo),
		Mensaje:        validation.SanitizeString(r.Mensaje),
		Tipo:           r.Tipo,
		Severidad:      r.Severidad,
		Importante:     r.Importante,
		Metadatos:      r.Metadatos,
	}

	if r.DispositivoID != "" {
		if dispositivoID, err := primitive.ObjectIDFromHex(r.DispositivoID); err == nil {
			entity.DispositivoID = dispositivoID
		}
	}

	if err := s.notificacionRepo.Create(ctx, entity); err != nil {
		return nil, err
	}

	if s.wsNotifier != nil {
		s.wsNotifier.NotificarNuevaNotificacion(entity)
	}

	// Envío por email en segundo plano: resolvemos el correo del destinatario
	// (cliente o empresa) y despachamos sin bloquear la respuesta HTTP.
	s.enviarPorEmail(r.DestinatarioID, r.Tipo, r.Titulo, r.Mensaje)

	return s.entityToModel(entity), nil
}

// enviarPorEmail resuelve el correo del destinatario (primero cliente, luego
// empresa) y envía la notificación por email en una goroutine. Es tolerante a
// dependencias nil para no romper flujos que construyan el servicio sin email.
func (s *NotificacionService) enviarPorEmail(destinatarioID, tipo, titulo, mensaje string) {
	if s.emailService == nil || destinatarioID == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		correo, nombre := s.resolverDestinatario(ctx, destinatarioID)
		if correo == "" {
			return
		}
		if err := s.emailService.EnviarNotificacion(correo, nombre, tipo, titulo, mensaje); err != nil {
			log.Printf("notificacion: error enviando email a %s: %v", correo, err)
		}
	}()
}

// resolverDestinatario devuelve (correo, nombre) buscando el ID como cliente y,
// si no existe, como empresa.
func (s *NotificacionService) resolverDestinatario(ctx context.Context, id string) (string, string) {
	if s.clienteRepo != nil {
		if cliente, err := s.clienteRepo.FindByID(ctx, id); err == nil && cliente != nil && cliente.Correo != "" {
			return cliente.Correo, cliente.Nombre
		}
	}
	if s.empresaRepo != nil {
		if empresa, err := s.empresaRepo.FindByID(ctx, id); err == nil && empresa != nil && empresa.Correo != "" {
			return empresa.Correo, empresa.NombreEmpresa
		}
	}
	return "", ""
}

func (s *NotificacionService) entityToModel(entity *entities.NotificacionEntity) *models.NotificacionModel {
	model := &models.NotificacionModel{
		ID:             entity.ID.Hex(),
		DestinatarioID: entity.DestinatarioID.Hex(),
		Titulo:         entity.Titulo,
		Mensaje:        entity.Mensaje,
		Tipo:           entity.Tipo,
		Severidad:      entity.Severidad,
		Leida:          entity.Leida,
		Resuelta:       entity.Resuelta,
		Importante:     entity.Importante,
		Resolucion:     entity.Resolucion,
		FechaCreacion:  entity.FechaCreacion,
		FechaResolucion: entity.FechaResolucion,
		Metadatos:      entity.Metadatos,
	}

	if !entity.DispositivoID.IsZero() {
		model.DispositivoID = entity.DispositivoID.Hex()
	}

	return model
}
