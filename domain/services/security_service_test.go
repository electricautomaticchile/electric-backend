package services

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/infrastructure/entities"
	"electric-backend/types"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type fakeClienteRepo struct {
	clientes map[string]*models.ClienteModel
	updated  string
}

func (f *fakeClienteRepo) FindAll(context.Context, string) ([]*models.ClienteModel, error) {
	return nil, nil
}
func (f *fakeClienteRepo) FindAllPaginated(context.Context, string, types.PaginationParams, types.FilterParams) ([]*models.ClienteModel, int64, error) {
	return nil, 0, nil
}
func (f *fakeClienteRepo) FindByID(_ context.Context, id string) (*models.ClienteModel, error) {
	if c, ok := f.clientes[id]; ok {
		return c, nil
	}
	return nil, types.ThrowData("Cliente no encontrado")
}
func (f *fakeClienteRepo) FindByNumeroCliente(context.Context, string) (*models.ClienteModel, error) {
	return nil, nil
}
func (f *fakeClienteRepo) FindByRut(_ context.Context, rut string) (*models.ClienteModel, error) {
	for _, cliente := range f.clientes {
		if cliente.Rut == "" {
			continue
		}
		if cliente.Rut == rut {
			return cliente, nil
		}
	}
	return nil, types.ThrowData("Cliente no encontrado")
}
func (f *fakeClienteRepo) FindByCorreo(context.Context, string) (*models.ClienteModel, error) {
	return nil, nil
}
func (f *fakeClienteRepo) Create(context.Context, *models.ClienteModel) error { return nil }
func (f *fakeClienteRepo) Update(context.Context, string, *models.ClienteModel) error {
	return nil
}
func (f *fakeClienteRepo) UpdateUltimoAcceso(context.Context, string) error { return nil }
func (f *fakeClienteRepo) UpdatePassword(_ context.Context, _ string, hashedPassword string) error {
	f.updated = hashedPassword
	return nil
}
func (f *fakeClienteRepo) Delete(context.Context, string) error { return nil }

type fakeEmpresaRepo struct{}

func (f *fakeEmpresaRepo) FindAll(context.Context) ([]*models.EmpresaModel, error) { return nil, nil }
func (f *fakeEmpresaRepo) FindByID(context.Context, string) (*models.EmpresaModel, error) {
	return nil, types.ThrowData("Empresa no encontrada")
}
func (f *fakeEmpresaRepo) FindByNumeroCliente(context.Context, string) (*models.EmpresaModel, error) {
	return nil, nil
}
func (f *fakeEmpresaRepo) FindByCorreo(context.Context, string) (*models.EmpresaModel, error) {
	return nil, nil
}
func (f *fakeEmpresaRepo) Create(context.Context, *models.EmpresaModel) error { return nil }
func (f *fakeEmpresaRepo) Update(context.Context, string, *models.EmpresaModel) error {
	return nil
}
func (f *fakeEmpresaRepo) UpdateUltimoAcceso(context.Context, string) error { return nil }
func (f *fakeEmpresaRepo) UpdatePassword(context.Context, string, string) error {
	return nil
}
func (f *fakeEmpresaRepo) Delete(context.Context, string) error { return nil }
func (f *fakeEmpresaRepo) CambiarEstado(context.Context, string, bool) error {
	return nil
}

func TestCambiarPasswordRequiresCurrentPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("Actual123!"), 4)
	if err != nil {
		t.Fatal(err)
	}
	clienteRepo := &fakeClienteRepo{
		clientes: map[string]*models.ClienteModel{
			"cliente-a": {ID: "cliente-a", Password: string(hash)},
		},
	}
	service := NewAuthService(&fakeEmpresaRepo{}, clienteRepo, nil, nil, nil, nil)

	err = service.CambiarPassword(context.Background(), "cliente-a", &recipe.CambiarPasswordRecipe{
		PasswordNuevo: "Nueva123!",
	})
	if err == nil {
		t.Fatal("deberia exigir passwordActual")
	}

	err = service.CambiarPassword(context.Background(), "cliente-a", &recipe.CambiarPasswordRecipe{
		PasswordActual: "Actual123!",
		PasswordNuevo:  "Nueva123!",
	})
	if err != nil {
		t.Fatalf("password actual correcta deberia pasar: %v", err)
	}
	if clienteRepo.updated == "" {
		t.Fatal("deberia guardar nuevo hash")
	}
}

func TestLoginRejectsInvalidClientSession(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("Correcta123!"), 4)
	if err != nil {
		t.Fatal(err)
	}
	clienteRepo := &fakeClienteRepo{
		clientes: map[string]*models.ClienteModel{
			"cliente-a": {
				ID:       primitive.NewObjectID().Hex(),
				Rut:      "11111111-1",
				Password: string(hash),
				Activo:   true,
			},
			"cliente-b": {
				ID:       primitive.NewObjectID().Hex(),
				Rut:      "22222222-2",
				Password: string(hash),
				Activo:   false,
			},
		},
	}
	service := NewAuthService(&fakeEmpresaRepo{}, clienteRepo, nil, nil, nil, nil)

	_, err = service.Login(context.Background(), &recipe.LoginRecipe{
		Rut:      "11.111.111-1",
		Password: "Incorrecta123!",
	})
	if err == nil {
		t.Fatal("login no deberia aceptar password incorrecta")
	}

	_, err = service.Login(context.Background(), &recipe.LoginRecipe{
		Rut:      "22.222.222-2",
		Password: "Correcta123!",
	})
	if err == nil {
		t.Fatal("login no deberia aceptar cliente inactivo")
	}
}

type fakeBoletaRepo struct {
	created *entities.BoletaEntity
}

func (f *fakeBoletaRepo) FindByCliente(context.Context, string) ([]*entities.BoletaEntity, error) {
	return nil, nil
}
func (f *fakeBoletaRepo) FindByID(context.Context, string) (*entities.BoletaEntity, error) {
	return nil, nil
}
func (f *fakeBoletaRepo) Create(_ context.Context, boleta *entities.BoletaEntity) error {
	f.created = boleta
	return nil
}
func (f *fakeBoletaRepo) Update(context.Context, string, *entities.BoletaEntity) error {
	return nil
}
func (f *fakeBoletaRepo) FindVencidasByCliente(context.Context, string) ([]*entities.BoletaEntity, error) {
	return nil, nil
}
func (f *fakeBoletaRepo) FindPendientesByCliente(context.Context, string) ([]*entities.BoletaEntity, error) {
	return nil, nil
}
func (f *fakeBoletaRepo) FindPorVencer(context.Context, int) ([]*entities.BoletaEntity, error) {
	return nil, nil
}
func (f *fakeBoletaRepo) FindVencidas(context.Context) ([]*entities.BoletaEntity, error) {
	return nil, nil
}
func (f *fakeBoletaRepo) FindClienteIDsConBoletasVencidas(context.Context) ([]string, error) {
	return nil, nil
}
func (f *fakeBoletaRepo) UpdateEstado(context.Context, string, string) error { return nil }
func (f *fakeBoletaRepo) UpdateNotificacionEnviada(context.Context, string, string) error {
	return nil
}

func TestCrearBoletaUsesClienteEmpresa(t *testing.T) {
	empresaID := primitive.NewObjectID()
	clienteID := primitive.NewObjectID()
	clienteRepo := &fakeClienteRepo{
		clientes: map[string]*models.ClienteModel{
			clienteID.Hex(): {ID: clienteID.Hex(), EmpresaID: empresaID.Hex()},
		},
	}
	boletaRepo := &fakeBoletaRepo{}
	service := NewBoletaService(boletaRepo, clienteRepo, nil)

	_, err := service.Crear(context.Background(), &recipe.CrearBoletaRecipe{
		ClienteID: clienteID.Hex(),
		Monto:     1000,
		Periodo:   "2026-05",
	})
	if err != nil {
		t.Fatalf("crear boleta deberia pasar: %v", err)
	}
	if boletaRepo.created == nil || boletaRepo.created.EmpresaID != empresaID {
		t.Fatal("boleta debe heredar empresaId del cliente")
	}
}

type fakeDispositivoRepo struct {
	device   *entities.DispositivoEntity
	reading  *entities.LecturaDispositivo
	created  *entities.DispositivoEntity
	createID primitive.ObjectID
}

func (f *fakeDispositivoRepo) FindAll(context.Context, string) ([]*entities.DispositivoEntity, error) {
	return nil, nil
}
func (f *fakeDispositivoRepo) FindByID(context.Context, string) (*entities.DispositivoEntity, error) {
	if f.device == nil {
		return nil, errors.New("not found")
	}
	return f.device, nil
}
func (f *fakeDispositivoRepo) FindByNumero(context.Context, string) (*entities.DispositivoEntity, error) {
	return f.device, nil
}
func (f *fakeDispositivoRepo) FindByCliente(context.Context, string) ([]*entities.DispositivoEntity, error) {
	return nil, nil
}
func (f *fakeDispositivoRepo) Create(_ context.Context, dispositivo *entities.DispositivoEntity) error {
	if f.createID.IsZero() {
		f.createID = primitive.NewObjectID()
	}
	dispositivo.ID = f.createID
	f.created = dispositivo
	return nil
}
func (f *fakeDispositivoRepo) Update(context.Context, string, *entities.DispositivoEntity) error {
	return nil
}
func (f *fakeDispositivoRepo) UpdateUltimaLectura(_ context.Context, _ string, lectura *entities.LecturaDispositivo) (*entities.DispositivoEntity, error) {
	f.reading = lectura
	return f.device, nil
}
func (f *fakeDispositivoRepo) CambiarEstado(context.Context, string, string) error { return nil }
func (f *fakeDispositivoRepo) Delete(context.Context, string) error                { return nil }

func TestDispositivoCrearValidaEmpresaCliente(t *testing.T) {
	empresaID := primitive.NewObjectID()
	clienteID := primitive.NewObjectID()
	clienteRepo := &fakeClienteRepo{
		clientes: map[string]*models.ClienteModel{
			clienteID.Hex(): {ID: clienteID.Hex(), EmpresaID: empresaID.Hex()},
		},
	}
	dispositivoRepo := &fakeDispositivoRepo{}
	service := NewDispositivoService(dispositivoRepo, clienteRepo, nil)

	_, err := service.Crear(context.Background(), &recipe.CrearDispositivoRecipe{
		NumeroDispositivo: "MED-1",
		Nombre:            "Medidor",
		Tipo:              "medidor",
		ClienteID:         clienteID.Hex(),
		EmpresaID:         primitive.NewObjectID().Hex(),
	})
	if err == nil {
		t.Fatal("no deberia crear dispositivo con cliente de otra empresa")
	}

	_, err = service.Crear(context.Background(), &recipe.CrearDispositivoRecipe{
		NumeroDispositivo: "MED-1",
		Nombre:            "Medidor",
		Tipo:              "medidor",
		ClienteID:         clienteID.Hex(),
		EmpresaID:         empresaID.Hex(),
	})
	if err != nil {
		t.Fatalf("deberia crear dispositivo con empresa correcta: %v", err)
	}
	if dispositivoRepo.created == nil || dispositivoRepo.created.EmpresaID != empresaID {
		t.Fatal("dispositivo debe guardar empresaId autenticado")
	}
}

func TestActualizarLecturaIoTStoresReading(t *testing.T) {
	dispositivoRepo := &fakeDispositivoRepo{device: &entities.DispositivoEntity{ID: primitive.NewObjectID()}}
	service := NewDispositivoService(dispositivoRepo, &fakeClienteRepo{}, nil)

	err := service.ActualizarUltimaLectura(context.Background(), "MED-1", &recipe.ActualizarLecturaRecipe{
		Voltage:     220,
		Current:     10,
		ActivePower: 2000,
		Energy:      12.5,
		Cost:        1500,
	})
	if err != nil {
		t.Fatalf("lectura IoT deberia pasar: %v", err)
	}
	if dispositivoRepo.reading == nil || dispositivoRepo.reading.Energy != 12.5 || dispositivoRepo.reading.ActivePower != 2000 {
		t.Fatal("lectura IoT no fue mapeada correctamente")
	}
}
