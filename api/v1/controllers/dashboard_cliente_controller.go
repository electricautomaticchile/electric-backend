package controllers

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/config"
	"electric-backend/domain/facades"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/arduino"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DashboardClienteController struct {
	clienteFacade     *facades.ClienteFacade
	dispositivoFacade *facades.DispositivoFacade
	boletaService     *services.BoletaService
	dashboardService  *services.DashboardService
	arduinoBridge     *arduino.SerialBridge
}

func NewDashboardClienteController(
	clienteFacade *facades.ClienteFacade,
	dispositivoFacade *facades.DispositivoFacade,
	boletaService *services.BoletaService,
	dashboardService *services.DashboardService,
	arduinoBridge *arduino.SerialBridge,
) *DashboardClienteController {
	return &DashboardClienteController{
		clienteFacade:     clienteFacade,
		dispositivoFacade: dispositivoFacade,
		boletaService:     boletaService,
		dashboardService:  dashboardService,
		arduinoBridge:     arduinoBridge,
	}
}

func (ctrl *DashboardClienteController) SetupRoutes(router *gin.RouterGroup) {
	dashboard := router.Group("/dashboard/cliente")
	dashboard.Use(middleware.AuthMiddleware())
	{
		dashboard.GET("", ctrl.ObtenerTodo)
		dashboard.GET("/resumen", ctrl.ObtenerResumen)
		dashboard.GET("/dispositivos", ctrl.ObtenerDispositivos)
		dashboard.GET("/consumo", ctrl.ObtenerConsumo)
		dashboard.GET("/perfil", ctrl.ObtenerPerfil)
		dashboard.GET("/boletas", ctrl.ObtenerBoletas)
		dashboard.PUT("/perfil", ctrl.ActualizarPerfil)
	}

	servicio := router.Group("/servicio-electrico")
	servicio.Use(middleware.AuthMiddleware())
	{
		servicio.GET("/:clienteId", ctrl.ObtenerEstadoServicio)
		servicio.POST("/:clienteId/cortar", ctrl.CortarServicio)
		servicio.POST("/:clienteId/restablecer", ctrl.RestablecerServicio)
		servicio.POST("/:clienteId/restablecer-empresa", ctrl.RestablecerServicio)
	}

	clientes := router.Group("/clientes")
	clientes.Use(middleware.AuthMiddleware())
	{
		clientes.GET("/mi-dispositivo", ctrl.ObtenerMiDispositivo)
	}
}

func (ctrl *DashboardClienteController) ObtenerTodo(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	type result struct {
		key  string
		data any
		err  error
	}

	ch := make(chan result, 4)

	go func() {
		cliente, err := ctrl.clienteFacade.ObtenerPorID(ctx, userID)
		if err != nil {
			ch <- result{key: "resumen", err: err}
			return
		}
		dispositivos, _ := ctrl.dispositivoFacade.ObtenerPorCliente(ctx, userID)
		dispositivosActivos := 0
		consumoActual := 0.0
		costoActual := 0.0
		var ultimaLectura any
		for _, disp := range dispositivos {
			if disp.Estado == "activo" {
				dispositivosActivos++
			}
			if disp.UltimaLectura != nil {
				consumoActual += disp.UltimaLectura.Energy
				costoActual += disp.UltimaLectura.Cost
				ultimaLectura = disp.UltimaLectura
			}
		}
		ch <- result{key: "resumen", data: gin.H{
			"cliente": gin.H{
				"nombre":           cliente.Nombre,
				"numeroCliente":    cliente.NumeroCliente,
				"correo":           cliente.Correo,
				"telefono":         cliente.Telefono,
				"direccion":        cliente.Direccion,
				"imagenPerfil":     cliente.ImagenPerfil,
				"passwordTemporal": cliente.PasswordTemporal != "",
			},
			"estadisticas": gin.H{
				"dispositivosActivos": dispositivosActivos,
				"dispositivosTotal":   len(dispositivos),
				"consumoMensual":      consumoActual,
				"costoMensual":        costoActual,
				"ultimaLectura":       ultimaLectura,
				"boletasPendientes":   0,
			},
		}}
	}()

	go func() {
		dispositivos, err := ctrl.dispositivoFacade.ObtenerPorCliente(ctx, userID)
		ch <- result{key: "dispositivos", data: dispositivos, err: err}
	}()

	go func() {
		consumo, err := ctrl.dashboardService.ObtenerConsumoCliente(ctx, userID)
		ch <- result{key: "consumo", data: consumo, err: err}
	}()

	go func() {
		boletas, err := ctrl.boletaService.ObtenerPorCliente(ctx, userID)
		ch <- result{key: "boletas", data: boletas, err: err}
	}()

	combined := make(map[string]any)
	for i := 0; i < 4; i++ {
		r := <-ch
		if r.err != nil {
			combined[r.key] = nil
		} else {
			combined[r.key] = r.data
		}
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    combined,
	})
}

func (ctrl *DashboardClienteController) ObtenerResumen(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	// Contexto propio para no depender del timeout del request HTTP
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cliente, err := ctrl.clienteFacade.ObtenerPorID(ctx, userID)
	if err != nil {
		gctx.Error(err)
		return
	}

	dispositivos, _ := ctrl.dispositivoFacade.ObtenerPorCliente(ctx, userID)

	dispositivosActivos := 0
	consumoActual := 0.0
	costoActual := 0.0
	var ultimaLectura interface{}

	for _, disp := range dispositivos {
		if disp.Estado == "activo" {
			dispositivosActivos++
		}
		// Datos en tiempo real desde ultimaLectura del dispositivo (actualizado por Arduino)
		if disp.UltimaLectura != nil {
			consumoActual += disp.UltimaLectura.Energy
			costoActual += disp.UltimaLectura.Cost
			ultimaLectura = disp.UltimaLectura
		}
	}

	resumen := gin.H{
		"cliente": gin.H{
			"nombre":           cliente.Nombre,
			"numeroCliente":    cliente.NumeroCliente,
			"correo":           cliente.Correo,
			"telefono":         cliente.Telefono,
			"direccion":        cliente.Direccion,
			"imagenPerfil":     cliente.ImagenPerfil,
			"passwordTemporal": cliente.PasswordTemporal != "",
		},
		"estadisticas": gin.H{
			"dispositivosActivos": dispositivosActivos,
			"dispositivosTotal":   len(dispositivos),
			"consumoMensual":      consumoActual,
			"costoMensual":        costoActual,
			"ultimaLectura":       ultimaLectura,
			"boletasPendientes":   0,
		},
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    resumen,
	})
}

func (ctrl *DashboardClienteController) ObtenerDispositivos(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	dispositivos, err := ctrl.dispositivoFacade.ObtenerPorCliente(gctx.Request.Context(), userID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    dispositivos,
	})
}

func (ctrl *DashboardClienteController) ObtenerConsumo(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	consumo, err := ctrl.dashboardService.ObtenerConsumoCliente(gctx.Request.Context(), userID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    consumo,
	})
}

func (ctrl *DashboardClienteController) ObtenerPerfil(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	cliente, err := ctrl.clienteFacade.ObtenerPorID(gctx.Request.Context(), userID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    cliente,
	})
}

func (ctrl *DashboardClienteController) ObtenerBoletas(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	boletas, err := ctrl.boletaService.ObtenerPorCliente(gctx.Request.Context(), userID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    boletas,
	})
}

func (ctrl *DashboardClienteController) ObtenerMiDispositivo(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	dispositivos, err := ctrl.dispositivoFacade.ObtenerPorCliente(gctx.Request.Context(), userID)
	if err != nil || len(dispositivos) == 0 {
		gctx.JSON(http.StatusOK, types.ApiResponse{
			Success: true,
			Data:    nil,
			Message: "Sin dispositivo asignado",
		})
		return
	}

	d := dispositivos[0]
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"dispositivoId":     d.ID,
			"numeroDispositivo": d.NumeroDispositivo,
			"nombre":            d.Nombre,
			"estado":            d.Estado,
			"ultimaLectura":     d.UltimaLectura,
		},
	})
}

func (ctrl *DashboardClienteController) ActualizarPerfil(gctx *gin.Context) {
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	var r recipe.ActualizarClienteRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	cliente, err := ctrl.clienteFacade.Actualizar(gctx.Request.Context(), userID, &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    cliente,
		Message: "Perfil actualizado correctamente",
	})
}

func (ctrl *DashboardClienteController) ObtenerEstadoServicio(gctx *gin.Context) {
	clienteId := gctx.Param("clienteId")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	boletas, err := ctrl.boletaService.ObtenerPorCliente(ctx, clienteId)
	if err != nil {
		boletas = nil
	}

	boletasPagadas := 0
	boletasPendientes := 0
	montoDeuda := 0.0

	for _, b := range boletas {
		if b.Estado == "pagada" || b.FechaPago != nil {
			boletasPagadas++
		} else {
			boletasPendientes++
			montoDeuda += b.Monto
		}
	}

	// Leer estado persistido desde MongoDB
	estadoServicio := leerEstadoServicioCliente(clienteId)
	motivoCorte := ""

	// Si tiene 3+ boletas pendientes, forzar cortado
	if boletasPendientes >= 3 && estadoServicio == "activo" {
		estadoServicio = "cortado"
		motivoCorte = "Más de 3 boletas pendientes de pago"
		persistirEstadoServicioCliente(clienteId, "cortado")
	}
	if estadoServicio == "cortado" && motivoCorte == "" {
		motivoCorte = "Corte manual por empresa"
	}

	puedeRestablecer := estadoServicio == "cortado" && boletasPendientes < 3

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"clienteId":           clienteId,
			"estadoServicio":      estadoServicio,
			"boletasPagadas":      boletasPagadas,
			"boletasPendientes":   boletasPendientes,
			"montoDeuda":          montoDeuda,
			"ultimaActualizacion": time.Now().Format(time.RFC3339),
			"puedeRestablecer":    puedeRestablecer,
			"motivoCorte":         motivoCorte,
			"historialCambios":    []gin.H{},
		},
	})
}

func (ctrl *DashboardClienteController) enviarComandoDispositivo(clienteId string, comando string) {
	if ctrl.arduinoBridge == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dispositivos, err := ctrl.dispositivoFacade.ObtenerPorCliente(ctx, clienteId)
	if err != nil || len(dispositivos) == 0 {
		return
	}
	for _, d := range dispositivos {
		if d.NumeroDispositivo != "" {
			ctrl.arduinoBridge.SendCommandToDevice(d.NumeroDispositivo, comando)
			return
		}
	}
}

func persistirEstadoServicioCliente(clienteId string, estado string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(clienteId)
	filter := bson.M{"clienteId": clienteId}
	if err == nil {
		filter = bson.M{"clienteId": oid}
	}
	config.MongoDB.Collection("dispositivos").UpdateMany(ctx,
		filter,
		bson.M{"$set": bson.M{"estadoServicio": estado}},
	)
}

func leerEstadoServicioCliente(clienteId string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(clienteId)
	filter := bson.M{"clienteId": clienteId}
	if err == nil {
		filter = bson.M{"clienteId": oid}
	}

	var doc bson.M
	if err := config.MongoDB.Collection("dispositivos").FindOne(ctx, filter).Decode(&doc); err != nil {
		return "activo"
	}
	if estado, ok := doc["estadoServicio"].(string); ok && estado != "" {
		return estado
	}
	return "activo"
}

func (ctrl *DashboardClienteController) CortarServicio(gctx *gin.Context) {
	clienteId := gctx.Param("clienteId")

	persistirEstadoServicioCliente(clienteId, "cortado")
	ctrl.enviarComandoDispositivo(clienteId, "DESACTIVAR_SERVICIO")

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"clienteId":           clienteId,
			"estadoServicio":      "cortado",
			"boletasPendientes":   0,
			"boletasPagadas":      0,
			"montoDeuda":          0,
			"ultimaActualizacion": time.Now().Format(time.RFC3339),
			"puedeRestablecer":    true,
			"motivoCorte":         "Corte manual por empresa",
			"historialCambios":    []gin.H{},
		},
		Message: "Servicio cortado correctamente",
	})
}

func (ctrl *DashboardClienteController) RestablecerServicio(gctx *gin.Context) {
	clienteId := gctx.Param("clienteId")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	boletas, _ := ctrl.boletaService.ObtenerPorCliente(ctx, clienteId)

	boletasPendientes := 0
	montoDeuda := 0.0
	for _, b := range boletas {
		if b.Estado != "pagada" && b.FechaPago == nil {
			boletasPendientes++
			montoDeuda += b.Monto
		}
	}

	if boletasPendientes >= 3 {
		gctx.JSON(http.StatusOK, types.ApiResponse{
			Success: false,
			Error:   "No puedes restablecer el servicio con 3 o más boletas pendientes",
		})
		return
	}

	persistirEstadoServicioCliente(clienteId, "activo")
	ctrl.enviarComandoDispositivo(clienteId, "ACTIVAR_SERVICIO")

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"clienteId":           clienteId,
			"estadoServicio":      "activo",
			"boletasPendientes":   boletasPendientes,
			"boletasPagadas":      0,
			"montoDeuda":          montoDeuda,
			"ultimaActualizacion": time.Now().Format(time.RFC3339),
			"puedeRestablecer":    false,
			"motivoCorte":         "",
			"historialCambios":    []gin.H{},
		},
		Message: "Servicio restablecido correctamente",
	})
}
