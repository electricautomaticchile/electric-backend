package services

import (
	"context"
	"sync"
	"time"

	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/entities"
)

// FeatureFlagService gestiona flags de funcionalidad para rollout gradual.
// Mantiene un caché en memoria con TTL corto para no consultar Mongo en cada
// verificación (los flags cambian poco y se toleran unos segundos de rezago).
type FeatureFlagService struct {
	repo ports.PortFeatureFlag

	mu       sync.RWMutex
	cache    map[string]*entities.FeatureFlagEntity
	cachedAt time.Time
	cacheTTL time.Duration
}

func NewFeatureFlagService(repo ports.PortFeatureFlag) *FeatureFlagService {
	return &FeatureFlagService{
		repo:     repo,
		cache:    map[string]*entities.FeatureFlagEntity{},
		cacheTTL: 30 * time.Second,
	}
}

// IsEnabled indica si un flag está activo para una empresa dada.
//   - flag inexistente o apagado → false.
//   - encendido sin lista de empresas → true para todas.
//   - encendido con lista → true solo si empresaID está en la lista.
func (s *FeatureFlagService) IsEnabled(ctx context.Context, key, empresaID string) bool {
	flag := s.getCached(ctx, key)
	if flag == nil || !flag.Enabled {
		return false
	}
	if len(flag.EmpresaIDs) == 0 {
		return true
	}
	for _, id := range flag.EmpresaIDs {
		if id == empresaID {
			return true
		}
	}
	return false
}

func (s *FeatureFlagService) getCached(ctx context.Context, key string) *entities.FeatureFlagEntity {
	s.mu.RLock()
	fresh := time.Since(s.cachedAt) < s.cacheTTL
	if fresh {
		flag := s.cache[key]
		s.mu.RUnlock()
		return flag
	}
	s.mu.RUnlock()
	return s.refreshAndGet(ctx, key)
}

func (s *FeatureFlagService) refreshAndGet(ctx context.Context, key string) *entities.FeatureFlagEntity {
	flags, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil
	}
	next := make(map[string]*entities.FeatureFlagEntity, len(flags))
	for _, f := range flags {
		next[f.Key] = f
	}
	s.mu.Lock()
	s.cache = next
	s.cachedAt = time.Now()
	s.mu.Unlock()
	return next[key]
}

func (s *FeatureFlagService) invalidate() {
	s.mu.Lock()
	s.cachedAt = time.Time{}
	s.mu.Unlock()
}

// Listar devuelve todos los flags.
func (s *FeatureFlagService) Listar(ctx context.Context) ([]*models.FeatureFlagModel, error) {
	flags, err := s.repo.FindAll(ctx)
	if err != nil {
		return []*models.FeatureFlagModel{}, nil
	}
	result := make([]*models.FeatureFlagModel, len(flags))
	for i, f := range flags {
		result[i] = s.entityToModel(f)
	}
	return result, nil
}

// Set crea o actualiza un flag (uso administrativo).
func (s *FeatureFlagService) Set(ctx context.Context, key, descripcion string, enabled bool, empresaIDs []string) error {
	if empresaIDs == nil {
		empresaIDs = []string{}
	}
	err := s.repo.Upsert(ctx, &entities.FeatureFlagEntity{
		Key:         key,
		Descripcion: descripcion,
		Enabled:     enabled,
		EmpresaIDs:  empresaIDs,
	})
	if err == nil {
		s.invalidate()
	}
	return err
}

// Eliminar borra un flag.
func (s *FeatureFlagService) Eliminar(ctx context.Context, key string) error {
	err := s.repo.Delete(ctx, key)
	if err == nil {
		s.invalidate()
	}
	return err
}

func (s *FeatureFlagService) entityToModel(f *entities.FeatureFlagEntity) *models.FeatureFlagModel {
	ids := f.EmpresaIDs
	if ids == nil {
		ids = []string{}
	}
	return &models.FeatureFlagModel{
		ID:          f.ID.Hex(),
		Key:         f.Key,
		Descripcion: f.Descripcion,
		Enabled:     f.Enabled,
		EmpresaIDs:  ids,
		UpdatedAt:   f.UpdatedAt,
	}
}
