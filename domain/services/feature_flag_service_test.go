package services

import (
	"context"
	"testing"

	"electric-backend/infrastructure/entities"
)

// fakeFeatureFlagRepo implementa ports.PortFeatureFlag en memoria para tests.
type fakeFeatureFlagRepo struct {
	flags []*entities.FeatureFlagEntity
}

func (r *fakeFeatureFlagRepo) FindAll(ctx context.Context) ([]*entities.FeatureFlagEntity, error) {
	return r.flags, nil
}
func (r *fakeFeatureFlagRepo) FindByKey(ctx context.Context, key string) (*entities.FeatureFlagEntity, error) {
	for _, f := range r.flags {
		if f.Key == key {
			return f, nil
		}
	}
	return nil, nil
}
func (r *fakeFeatureFlagRepo) Upsert(ctx context.Context, flag *entities.FeatureFlagEntity) error {
	r.flags = append(r.flags, flag)
	return nil
}
func (r *fakeFeatureFlagRepo) Delete(ctx context.Context, key string) error { return nil }

func TestFeatureFlag_IsEnabled(t *testing.T) {
	repo := &fakeFeatureFlagRepo{flags: []*entities.FeatureFlagEntity{
		{Key: "apagado", Enabled: false},
		{Key: "para-todos", Enabled: true},
		{Key: "solo-empresa-A", Enabled: true, EmpresaIDs: []string{"empA"}},
	}}
	svc := NewFeatureFlagService(repo)
	ctx := context.Background()

	casos := []struct {
		nombre    string
		key       string
		empresaID string
		esperado  bool
	}{
		{"flag inexistente", "no-existe", "empA", false},
		{"flag apagado", "apagado", "empA", false},
		{"encendido para todos", "para-todos", "cualquiera", true},
		{"encendido solo empresa A, es A", "solo-empresa-A", "empA", true},
		{"encendido solo empresa A, es B", "solo-empresa-A", "empB", false},
	}

	for _, c := range casos {
		if got := svc.IsEnabled(ctx, c.key, c.empresaID); got != c.esperado {
			t.Errorf("%s: IsEnabled(%q,%q)=%v, esperado %v", c.nombre, c.key, c.empresaID, got, c.esperado)
		}
	}
}
