package services

import (
	"context"
	"testing"

	"electric-backend/infrastructure/entities"
)

// fakeUmbralRepo implementa ports.PortUmbralAlerta en memoria para tests.
type fakeUmbralRepo struct {
	umbrales map[string]*entities.UmbralAlertaEntity
}

func newFakeUmbralRepo() *fakeUmbralRepo {
	return &fakeUmbralRepo{umbrales: map[string]*entities.UmbralAlertaEntity{}}
}

func (r *fakeUmbralRepo) FindByEmpresa(ctx context.Context, empresaID string) (*entities.UmbralAlertaEntity, error) {
	return r.umbrales[empresaID], nil
}
func (r *fakeUmbralRepo) Upsert(ctx context.Context, umbral *entities.UmbralAlertaEntity) error {
	r.umbrales[umbral.EmpresaID] = umbral
	return nil
}

func TestUmbralService_DefaultsCuandoNoHayConfig(t *testing.T) {
	svc := NewUmbralService(newFakeUmbralRepo())
	u := svc.ObtenerUmbrales(context.Background(), "empresa-sin-config")

	if !u.EsDefault {
		t.Error("esperaba EsDefault=true cuando no hay configuración")
	}
	if u.VoltajeMin != DefaultVoltajeMin || u.VoltajeMax != DefaultVoltajeMax {
		t.Errorf("voltaje por defecto incorrecto: got (%.1f,%.1f), esperado (%.1f,%.1f)",
			u.VoltajeMin, u.VoltajeMax, DefaultVoltajeMin, DefaultVoltajeMax)
	}
	if u.CorrienteMax != DefaultCorrienteMax {
		t.Errorf("corrienteMax por defecto incorrecto: got %.1f, esperado %.1f", u.CorrienteMax, DefaultCorrienteMax)
	}
	if u.ConsumoMax != DefaultConsumoMax {
		t.Errorf("consumoMax por defecto incorrecto: got %.1f, esperado %.1f", u.ConsumoMax, DefaultConsumoMax)
	}
}

func TestUmbralService_ValoresPropiosCuandoHayConfig(t *testing.T) {
	repo := newFakeUmbralRepo()
	repo.umbrales["empA"] = &entities.UmbralAlertaEntity{
		EmpresaID:    "empA",
		VoltajeMin:   210,
		VoltajeMax:   230,
		CorrienteMax: 30,
		ConsumoMax:   80,
	}
	svc := NewUmbralService(repo)

	u := svc.ObtenerUmbrales(context.Background(), "empA")
	if u.EsDefault {
		t.Error("esperaba EsDefault=false cuando hay configuración propia")
	}
	if u.VoltajeMin != 210 || u.VoltajeMax != 230 || u.CorrienteMax != 30 || u.ConsumoMax != 80 {
		t.Errorf("umbrales propios incorrectos: %+v", u)
	}
}

func TestUmbralService_GuardarInvalidaCacheYPersiste(t *testing.T) {
	repo := newFakeUmbralRepo()
	svc := NewUmbralService(repo)
	ctx := context.Background()

	// Primero cachea los defaults.
	if u := svc.ObtenerUmbrales(ctx, "empB"); !u.EsDefault {
		t.Fatal("esperaba defaults iniciales")
	}

	// Guardar debe invalidar el caché y reflejar los nuevos valores.
	if err := svc.Guardar(ctx, "empB", 205, 235, 40, 90); err != nil {
		t.Fatalf("Guardar falló: %v", err)
	}

	u := svc.ObtenerUmbrales(ctx, "empB")
	if u.EsDefault {
		t.Error("tras Guardar, esperaba EsDefault=false (caché debió invalidarse)")
	}
	if u.VoltajeMin != 205 || u.VoltajeMax != 235 || u.CorrienteMax != 40 || u.ConsumoMax != 90 {
		t.Errorf("valores tras Guardar incorrectos: %+v", u)
	}
}
