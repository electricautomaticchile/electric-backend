package services

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/entities"
	"electric-backend/types"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DispositivoService struct {
	dispositivoRepo ports.PortDispositivo
	wsNotifier      *WebSocketNotifierService
}

func NewDispositivoService(dispositivoRepo ports.PortDispositivo, wsNotifier *WebSocketNotifierService) *DispositivoService {
	return &DispositivoService{
		dispositivoRepo: dispositivoRepo,
		wsNotifier:      wsNotifier,
	}
}

func (s *DispositivoService) ObtenerTodos(ctx context.Context, empresaID string) ([]*models.DispositivoModel, error) {
	dispositivos, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return []*models.DispositivoModel{}, nil
	}

	models := make([]*models.DispositivoModel, len(dispositivos))
	for i, dispositivo := range dispositivos {
		models[i] = s.entityToModel(dispositivo)
	}

	return models, nil
}

func (s *DispositivoService) ObtenerPorID(ctx context.Context, id string) (*models.DispositivoModel, error) {
	dispositivo, err := s.dispositivoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.entityToModel(dispositivo), nil
}

func (s *DispositivoService) ObtenerPorNumero(ctx context.Context, numeroDispositivo string) (*models.DispositivoModel, error) {
	dispositivo, err := s.dispositivoRepo.FindByNumero(ctx, numeroDispositivo)
	if err != nil {
		return nil, err
	}

	return s.entityToModel(dispositivo), nil
}

func (s *DispositivoService) ObtenerPorCliente(ctx context.Context, clienteID string) ([]*models.DispositivoModel, error) {
	dispositivos, err := s.dispositivoRepo.FindByCliente(ctx, clienteID)
	if err != nil {
		return []*models.DispositivoModel{}, nil
	}

	models := make([]*models.DispositivoModel, len(dispositivos))
	for i, dispositivo := range dispositivos {
		models[i] = s.entityToModel(dispositivo)
	}

	return models, nil
}

func (s *DispositivoService) Crear(ctx context.Context, r *recipe.CrearDispositivoRecipe) (*models.DispositivoModel, error) {
	clienteID, err := primitive.ObjectIDFromHex(r.ClienteID)
	if err != nil {
		return nil, types.ThrowRecipe("ClienteID inválido", "clienteId")
	}

	empresaID, err := primitive.ObjectIDFromHex(r.EmpresaID)
	if err != nil {
		return nil, types.ThrowRecipe("EmpresaID inválido", "empresaId")
	}

	entity := &entities.DispositivoEntity{
		NumeroDispositivo: r.NumeroDispositivo,
		Nombre:            r.Nombre,
		Tipo:              r.Tipo,
		ClienteID:         clienteID,
		EmpresaID:         empresaID,
	}

	if err := s.dispositivoRepo.Create(ctx, entity); err != nil {
		return nil, err
	}

	return s.entityToModel(entity), nil
}

func (s *DispositivoService) Actualizar(ctx context.Context, id string, r *recipe.ActualizarDispositivoRecipe) (*models.DispositivoModel, error) {
	dispositivo, err := s.dispositivoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if r.Nombre != "" {
		dispositivo.Nombre = r.Nombre
	}
	if r.Tipo != "" {
		dispositivo.Tipo = r.Tipo
	}

	if err := s.dispositivoRepo.Update(ctx, id, dispositivo); err != nil {
		return nil, err
	}

	return s.entityToModel(dispositivo), nil
}

func (s *DispositivoService) AsignarCliente(ctx context.Context, id string, clienteID string) (*models.DispositivoModel, error) {
	dispositivo, err := s.dispositivoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if clienteID == "" {
		dispositivo.ClienteID = primitive.NilObjectID
	} else {
		clienteOID, err := primitive.ObjectIDFromHex(clienteID)
		if err != nil {
			return nil, types.ThrowRecipe("ClienteID inválido", "clienteId")
		}
		dispositivo.ClienteID = clienteOID
	}

	if err := s.dispositivoRepo.Update(ctx, id, dispositivo); err != nil {
		return nil, err
	}

	return s.entityToModel(dispositivo), nil
}

func (s *DispositivoService) ActualizarUltimaLectura(ctx context.Context, numeroDispositivo string, r *recipe.ActualizarLecturaRecipe) error {
	lectura := &entities.LecturaDispositivo{
		Voltage:     r.Voltage,
		Current:     r.Current,
		ActivePower: r.ActivePower,
		Energy:      r.Energy,
		Cost:        r.Cost,
		Timestamp:   time.Now(),
	}

	if err := s.dispositivoRepo.UpdateUltimaLectura(ctx, numeroDispositivo, lectura); err != nil {
		return err
	}

	if s.wsNotifier != nil {
		dispositivo, err := s.dispositivoRepo.FindByNumero(ctx, numeroDispositivo)
		if err == nil {
			s.wsNotifier.NotificarActualizacionDispositivo(dispositivo)
		}
	}

	return nil
}

func (s *DispositivoService) CambiarEstado(ctx context.Context, id string, r *recipe.CambiarEstadoDispositivoRecipe) error {
	return s.dispositivoRepo.CambiarEstado(ctx, id, r.Estado)
}

func (s *DispositivoService) Eliminar(ctx context.Context, id string) error {
	return s.dispositivoRepo.Delete(ctx, id)
}

func (s *DispositivoService) entityToModel(entity *entities.DispositivoEntity) *models.DispositivoModel {
	model := &models.DispositivoModel{
		ID:                  entity.ID.Hex(),
		NumeroDispositivo:   entity.NumeroDispositivo,
		Nombre:              entity.Nombre,
		Tipo:                entity.Tipo,
		ClienteID:           entity.ClienteID.Hex(),
		EmpresaID:           entity.EmpresaID.Hex(),
		Estado:              entity.Estado,
		Latitud:             entity.Latitud,
		Longitud:            entity.Longitud,
		Direccion:           entity.Direccion,
		Configuracion:       entity.Configuracion,
		Activo:              entity.Activo,
		FechaCreacion:       entity.FechaCreacion,
		FechaActualizacion:  entity.FechaActualizacion,
	}

	if entity.UltimaLectura != nil {
		model.UltimaLectura = &models.LecturaDispositivoModel{
			Voltage:     entity.UltimaLectura.Voltage,
			Current:     entity.UltimaLectura.Current,
			ActivePower: entity.UltimaLectura.ActivePower,
			Energy:      entity.UltimaLectura.Energy,
			Cost:        entity.UltimaLectura.Cost,
			Timestamp:   entity.UltimaLectura.Timestamp,
		}
	}

	return model
}

func (s *DispositivoService) ObtenerDispositivosConUbicacion(ctx context.Context, empresaID string) ([]map[string]interface{}, error) {
	dispositivos, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return []map[string]interface{}{}, nil
	}

	resultado := make([]map[string]interface{}, 0)
	for _, d := range dispositivos {
		if d.Latitud != 0 && d.Longitud != 0 {
			item := map[string]interface{}{
				"id":                d.ID.Hex(),
				"numeroDispositivo": d.NumeroDispositivo,
				"nombre":            d.Nombre,
				"tipo":              d.Tipo,
				"estado":            d.Estado,
				"latitud":           d.Latitud,
				"longitud":          d.Longitud,
				"direccion":         d.Direccion,
				"clienteId":         d.ClienteID.Hex(),
				"activo":            d.Activo,
			}

			if d.UltimaLectura != nil {
				item["consumo"] = d.UltimaLectura.Energy
				item["ultimaLectura"] = d.UltimaLectura.Timestamp
			}

			resultado = append(resultado, item)
		}
	}

	return resultado, nil
}
