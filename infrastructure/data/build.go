package data

// DataContainer agrupa todos los repositorios de la aplicación.
// Centraliza la creación para inyección de dependencias limpia.
type DataContainer struct {
	RecoveryTokenRepo  *RecoveryTokenRepository
	ClienteRepo        *ClienteRepository
	EmpresaRepo        *EmpresaRepository
	DispositivoRepo    *DispositivoRepository
	NotificacionRepo   *NotificacionRepository
	BoletaRepo         *BoletaRepository
	TicketRepo         *TicketRepository
	EstadisticaRepo    *EstadisticaRepository
	CotizacionRepo     *CotizacionRepository
	TarifaRepo         *TarifaRepository
	UsuarioEmpresaRepo *UsuarioEmpresaRepository
	RefreshTokenRepo   *RefreshTokenRepository
	AuditLogRepo       *AuditLogRepository
	FCMTokenRepo       *FCMTokenRepository
	FeatureFlagRepo    *FeatureFlagRepository
}

func Build() *DataContainer {
	return &DataContainer{
		RecoveryTokenRepo:  NewRecoveryTokenRepository(),
		ClienteRepo:        NewClienteRepository(),
		EmpresaRepo:        NewEmpresaRepository(),
		DispositivoRepo:    NewDispositivoRepository(),
		NotificacionRepo:   NewNotificacionRepository(),
		BoletaRepo:         NewBoletaRepository(),
		TicketRepo:         NewTicketRepository(),
		EstadisticaRepo:    NewEstadisticaRepository(),
		CotizacionRepo:     NewCotizacionRepository(),
		TarifaRepo:         NewTarifaRepository(),
		UsuarioEmpresaRepo: NewUsuarioEmpresaRepository(),
		RefreshTokenRepo:   NewRefreshTokenRepository(),
		AuditLogRepo:       NewAuditLogRepository(),
		FCMTokenRepo:       NewFCMTokenRepository(),
		FeatureFlagRepo:    NewFeatureFlagRepository(),
	}
}
