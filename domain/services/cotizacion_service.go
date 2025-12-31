package services

import (
	"context"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
)

type CotizacionService struct {
	cotizacionRepo ports.PortCotizacion
}

func NewCotizacionService(cotizacionRepo ports.PortCotizacion) *CotizacionService {
	return &CotizacionService{
		cotizacionRepo: cotizacionRepo,
	}
}

func (s *CotizacionService) ObtenerTodas(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.CotizacionModel, int64, error) {
	return s.cotizacionRepo.FindAll(ctx, page, limit, filters)
}

func (s *CotizacionService) ObtenerPorID(ctx context.Context, id string) (*models.CotizacionModel, error) {
	return s.cotizacionRepo.FindByID(ctx, id)
}

func (s *CotizacionService) ObtenerPorNumero(ctx context.Context, numero string) (*models.CotizacionModel, error) {
	return s.cotizacionRepo.FindByNumero(ctx, numero)
}

func (s *CotizacionService) Crear(ctx context.Context, nombre, email, empresa, telefono, servicio, plazo, mensaje string) (*models.CotizacionModel, error) {
	// Asignar prioridad basada en plazo
	prioridad := "baja"
	switch plazo {
	case "urgente":
		prioridad = "critica"
	case "pronto":
		prioridad = "alta"
	case "normal":
		prioridad = "media"
	}

	model := &models.CotizacionModel{
		Nombre:    nombre,
		Email:     email,
		Empresa:   empresa,
		Telefono:  telefono,
		Servicio:  servicio,
		Plazo:     plazo,
		Mensaje:   mensaje,
		Prioridad: prioridad,
	}

	if err := s.cotizacionRepo.Create(ctx, model); err != nil {
		return nil, err
	}

	return model, nil
}

func (s *CotizacionService) Actualizar(ctx context.Context, id string, updates map[string]interface{}) (*models.CotizacionModel, error) {
	cotizacion, err := s.cotizacionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if nombre, ok := updates["nombre"].(string); ok && nombre != "" {
		cotizacion.Nombre = nombre
	}
	if email, ok := updates["email"].(string); ok && email != "" {
		cotizacion.Email = email
	}
	if empresa, ok := updates["empresa"].(string); ok {
		cotizacion.Empresa = empresa
	}
	if telefono, ok := updates["telefono"].(string); ok {
		cotizacion.Telefono = telefono
	}
	if estado, ok := updates["estado"].(string); ok && estado != "" {
		cotizacion.Estado = estado
	}
	if prioridad, ok := updates["prioridad"].(string); ok && prioridad != "" {
		cotizacion.Prioridad = prioridad
	}

	if err := s.cotizacionRepo.Update(ctx, id, cotizacion); err != nil {
		return nil, err
	}

	return cotizacion, nil
}

func (s *CotizacionService) ActualizarEstado(ctx context.Context, id string, estado string) error {
	return s.cotizacionRepo.UpdateEstado(ctx, id, estado)
}

func (s *CotizacionService) Eliminar(ctx context.Context, id string) error {
	return s.cotizacionRepo.Delete(ctx, id)
}
