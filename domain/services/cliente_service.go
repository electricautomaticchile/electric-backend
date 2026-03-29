package services

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/email"
	"electric-backend/infrastructure/validation"
	"electric-backend/types"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type ClienteService struct {
	clienteRepo  ports.PortCliente
	emailService email.EmailService
}

func NewClienteService(clienteRepo ports.PortCliente, emailService email.EmailService) *ClienteService {
	return &ClienteService{
		clienteRepo:  clienteRepo,
		emailService: emailService,
	}
}

func (s *ClienteService) ObtenerTodos(ctx context.Context, empresaID string) ([]*models.ClienteModel, error) {
	return s.clienteRepo.FindAll(ctx, empresaID)
}

func (s *ClienteService) ObtenerTodosPaginado(ctx context.Context, empresaID string, params types.PaginationParams, filters types.FilterParams) ([]*models.ClienteModel, int64, error) {
	return s.clienteRepo.FindAllPaginated(ctx, empresaID, params, filters)
}

func (s *ClienteService) ObtenerPorID(ctx context.Context, id string) (*models.ClienteModel, error) {
	return s.clienteRepo.FindByID(ctx, id)
}

func (s *ClienteService) ObtenerPorNumero(ctx context.Context, numeroCliente string) (*models.ClienteModel, error) {
	return s.clienteRepo.FindByNumeroCliente(ctx, numeroCliente)
}

func (s *ClienteService) Crear(ctx context.Context, r *recipe.CrearClienteRecipe) (*models.ClienteModel, error) {
	if r.EmpresaID == "" {
		return nil, types.ThrowRecipe("EmpresaID es requerido", "empresaId")
	}

	numeroCliente := r.NumeroCliente
	if numeroCliente == "" {
		numeroCliente = s.generarNumeroCliente()
	}

	passwordTemporal := s.generarPasswordTemporal()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordTemporal), 12)
	if err != nil {
		return nil, err
	}

	rutNormalizado := ""
	if r.Rut != "" {
		rutNormalizado = validation.NormalizarRUT(r.Rut)
	}

	telefonoNormalizado := ""
	if r.Telefono != "" {
		telefonoNormalizado = validation.NormalizarTelefono(r.Telefono)
	}

	model := &models.ClienteModel{
		Nombre:           validation.SanitizeString(r.Nombre),
		Correo:           validation.SanitizeEmail(r.Correo),
		NumeroCliente:    validation.SanitizeNumeroCliente(numeroCliente),
		Telefono:         telefonoNormalizado,
		Direccion:        validation.SanitizeString(r.Direccion),
		Ciudad:           validation.SanitizeString(r.Ciudad),
		Rut:              rutNormalizado,
		TipoCliente:      r.TipoCliente,
		Empresa:          validation.SanitizeString(r.Empresa),
		EmpresaID:        r.EmpresaID,
		Password:         string(hashedPassword),
		PasswordTemporal: passwordTemporal,
		Role:             "cliente",
		TipoUsuario:      "cliente",
	}

	if err := s.clienteRepo.Create(ctx, model); err != nil {
		return nil, err
	}

	go s.enviarCredencialesPorEmail(model.Correo, model.Nombre, numeroCliente, passwordTemporal)

	return model, nil
}

func (s *ClienteService) enviarCredencialesPorEmail(correo, nombre, numeroCliente, passwordTemporal string) {
	err := s.emailService.EnviarCredenciales(correo, nombre, numeroCliente, passwordTemporal)
	if err != nil {
		log.Printf("Error enviando credenciales por email: %v", err)
	}
}

func (s *ClienteService) generarPasswordTemporal() string {
	return "Temp" + time.Now().Format("0601021504")
}

func (s *ClienteService) generarNumeroCliente() string {
	numero := 1000000 + (int(time.Now().UnixNano()) % 9000000)
	dv := s.calcularDigitoVerificador(numero)
	return fmt.Sprintf("%d-%s", numero, dv)
}

func (s *ClienteService) calcularDigitoVerificador(numero int) string {
	suma := 0
	multiplicador := 2
	numeroStr := fmt.Sprintf("%d", numero)

	for i := len(numeroStr) - 1; i >= 0; i-- {
		digito := int(numeroStr[i] - '0')
		suma += digito * multiplicador
		multiplicador++
		if multiplicador > 7 {
			multiplicador = 2
		}
	}

	resto := suma % 11
	dv := 11 - resto

	if dv == 11 {
		return "0"
	} else if dv == 10 {
		return "K"
	}
	return fmt.Sprintf("%d", dv)
}

func (s *ClienteService) Actualizar(ctx context.Context, id string, r *recipe.ActualizarClienteRecipe) (*models.ClienteModel, error) {
	cliente, err := s.clienteRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if r.Nombre != "" {
		cliente.Nombre = validation.SanitizeString(r.Nombre)
	}
	if r.Correo != "" {
		cliente.Correo = validation.SanitizeEmail(r.Correo)
	}
	if r.Telefono != "" {
		cliente.Telefono = validation.NormalizarTelefono(r.Telefono)
	}
	if r.Direccion != "" {
		cliente.Direccion = validation.SanitizeString(r.Direccion)
	}
	if r.Ciudad != "" {
		cliente.Ciudad = validation.SanitizeString(r.Ciudad)
	}
	if r.Rut != "" {
		cliente.Rut = validation.NormalizarRUT(r.Rut)
	}
	if r.Activo != nil {
		cliente.Activo = *r.Activo
	}

	if err := s.clienteRepo.Update(ctx, id, cliente); err != nil {
		return nil, err
	}

	return cliente, nil
}

func (s *ClienteService) Eliminar(ctx context.Context, id string) error {
	return s.clienteRepo.Delete(ctx, id)
}

func (s *ClienteService) ObtenerClientesConUbicacion(ctx context.Context, empresaID string) ([]map[string]interface{}, error) {
	clientes, err := s.clienteRepo.FindAll(ctx, empresaID)
	if err != nil {
		return []map[string]interface{}{}, nil
	}

	resultado := make([]map[string]interface{}, 0)
	for _, c := range clientes {
		if c.Latitud != 0 && c.Longitud != 0 {
			resultado = append(resultado, map[string]interface{}{
				"id":            c.ID,
				"nombre":        c.Nombre,
				"numeroCliente": c.NumeroCliente,
				"direccion":     c.Direccion,
				"ciudad":        c.Ciudad,
				"latitud":       c.Latitud,
				"longitud":      c.Longitud,
				"activo":        c.Activo,
			})
		}
	}

	return resultado, nil
}
