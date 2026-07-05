package services

import (
	"context"
	"sync"
	"time"

	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/entities"
)

// Umbrales de alerta por defecto. Coinciden con las constantes que estaban
// hardcodeadas en MonitoreoService, para no cambiar el comportamiento cuando
// una empresa no tiene configuración propia.
const (
	DefaultVoltajeMin   = 200.0
	DefaultVoltajeMax   = 240.0
	DefaultCorrienteMax = 50.0
	DefaultConsumoMax   = 100.0
)

// Umbrales agrupa los valores resueltos para una empresa (propios o defaults).
type Umbrales struct {
	VoltajeMin   float64
	VoltajeMax   float64
	CorrienteMax float64
	ConsumoMax   float64
	EsDefault    bool
}

// UmbralesPorDefecto devuelve los umbrales por defecto del sistema.
func UmbralesPorDefecto() Umbrales {
	return Umbrales{
		VoltajeMin:   DefaultVoltajeMin,
		VoltajeMax:   DefaultVoltajeMax,
		CorrienteMax: DefaultCorrienteMax,
		ConsumoMax:   DefaultConsumoMax,
		EsDefault:    true,
	}
}

// UmbralService gestiona los umbrales de alerta configurables por empresa.
// Mantiene un caché en memoria con TTL corto por empresa para no consultar
// Mongo en cada verificación de monitoreo.
type UmbralService struct {
	repo ports.PortUmbralAlerta

	mu       sync.RWMutex
	cache    map[string]cachedUmbral
	cacheTTL time.Duration
}

type cachedUmbral struct {
	umbral Umbrales
	at     time.Time
}

func NewUmbralService(repo ports.PortUmbralAlerta) *UmbralService {
	return &UmbralService{
		repo:     repo,
		cache:    map[string]cachedUmbral{},
		cacheTTL: 60 * time.Second,
	}
}

// ObtenerUmbrales devuelve los umbrales de la empresa o los defaults si no hay
// configuración. Ante un error de repositorio también cae en los defaults para
// preservar el comportamiento actual del monitoreo.
func (s *UmbralService) ObtenerUmbrales(ctx context.Context, empresaID string) Umbrales {
	if empresaID == "" {
		return UmbralesPorDefecto()
	}

	s.mu.RLock()
	if c, ok := s.cache[empresaID]; ok && time.Since(c.at) < s.cacheTTL {
		s.mu.RUnlock()
		return c.umbral
	}
	s.mu.RUnlock()

	umbral := UmbralesPorDefecto()
	if ent, err := s.repo.FindByEmpresa(ctx, empresaID); err == nil && ent != nil {
		umbral = Umbrales{
			VoltajeMin:   ent.VoltajeMin,
			VoltajeMax:   ent.VoltajeMax,
			CorrienteMax: ent.CorrienteMax,
			ConsumoMax:   ent.ConsumoMax,
			EsDefault:    false,
		}
	}

	s.mu.Lock()
	s.cache[empresaID] = cachedUmbral{umbral: umbral, at: time.Now()}
	s.mu.Unlock()
	return umbral
}

// ObtenerModelo devuelve los umbrales de una empresa en formato de API.
func (s *UmbralService) ObtenerModelo(ctx context.Context, empresaID string) *models.UmbralAlertaModel {
	u := s.ObtenerUmbrales(ctx, empresaID)
	return &models.UmbralAlertaModel{
		EmpresaID:    empresaID,
		VoltajeMin:   u.VoltajeMin,
		VoltajeMax:   u.VoltajeMax,
		CorrienteMax: u.CorrienteMax,
		ConsumoMax:   u.ConsumoMax,
		EsDefault:    u.EsDefault,
	}
}

// Guardar crea o actualiza los umbrales de una empresa e invalida el caché.
func (s *UmbralService) Guardar(ctx context.Context, empresaID string, voltajeMin, voltajeMax, corrienteMax, consumoMax float64) error {
	err := s.repo.Upsert(ctx, &entities.UmbralAlertaEntity{
		EmpresaID:    empresaID,
		VoltajeMin:   voltajeMin,
		VoltajeMax:   voltajeMax,
		CorrienteMax: corrienteMax,
		ConsumoMax:   consumoMax,
	})
	if err == nil {
		s.mu.Lock()
		delete(s.cache, empresaID)
		s.mu.Unlock()
	}
	return err
}
