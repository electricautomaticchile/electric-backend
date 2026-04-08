package facades

import "electric-backend/domain/services"

// FacadeContainer agrupa todas las facades de la aplicación.
type FacadeContainer struct {
	AuthFacade        *AuthFacade
	ClienteFacade     *ClienteFacade
	EmpresaFacade     *EmpresaFacade
	DispositivoFacade *DispositivoFacade
	CotizacionFacade  *CotizacionFacade
}

func Build(svc *services.ServiceContainer) *FacadeContainer {
	return &FacadeContainer{
		AuthFacade:        NewAuthFacade(svc.AuthService),
		ClienteFacade:     NewClienteFacade(svc.ClienteService),
		EmpresaFacade:     NewEmpresaFacade(svc.EmpresaService),
		DispositivoFacade: NewDispositivoFacade(svc.DispositivoService),
		CotizacionFacade:  NewCotizacionFacade(svc.CotizacionService),
	}
}
