package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/facades"
	"electric-backend/domain/models"
	"electric-backend/infrastructure/entities"
	"electric-backend/infrastructure/iot"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type DispositivoController struct {
	dispositivoFacade *facades.DispositivoFacade
}

func NewDispositivoController(dispositivoFacade *facades.DispositivoFacade) *DispositivoController {
	return &DispositivoController{
		dispositivoFacade: dispositivoFacade,
	}
}

// SetupRoutes configura las rutas CRUD del controlador (con AuthMiddleware)
func (ctrl *DispositivoController) SetupRoutes(router *gin.RouterGroup) {
	g := router.Group("/dispositivos")
	g.Use(middleware.AuthMiddleware(), middleware.CSRFMiddleware())
	g.GET("", ctrl.ObtenerTodos)
	g.GET("/:id", ctrl.ObtenerPorID)
	g.POST("", ctrl.Crear)
	g.PUT("/:id", ctrl.Actualizar)
	g.PUT("/:id/asignar", ctrl.AsignarCliente)
	g.PUT("/:id/desasignar", ctrl.DesasignarCliente)
	g.DELETE("/:id", ctrl.Eliminar)
}

// SetupIoTRoutes configura las rutas IoT (con IoTAPIKeyMiddleware)
func (ctrl *DispositivoController) SetupIoTRoutes(router *gin.RouterGroup) {
	g := router.Group("/iot")
	g.Use(middleware.IoTAPIKeyMiddleware())
	g.POST("/lectura", ctrl.RecibirLecturaIoT)
	g.POST("/comando-ejecutado", ctrl.ComandoEjecutado)
}

func (ctrl *DispositivoController) ObtenerTodos(gctx *gin.Context) {
	empresaID := gctx.Request.Context().Value(types.ContextKeyEmpresaID)
	if empresaID == nil {
		gctx.Error(types.ThrowPower("No tienes acceso a esta empresa"))
		return
	}

	page, _ := strconv.Atoi(gctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(gctx.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	skip := (page - 1) * limit

	dispositivos, err := ctrl.dispositivoFacade.ObtenerTodos(gctx.Request.Context(), empresaID.(string))
	if err != nil {
		gctx.Error(err)
		return
	}

	total := len(dispositivos)
	end := skip + limit
	if end > total {
		end = total
	}

	paginatedDispositivos := dispositivos
	if skip < total {
		paginatedDispositivos = dispositivos[skip:end]
	} else {
		paginatedDispositivos = []*models.DispositivoModel{}
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    paginatedDispositivos,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	})
}

func (ctrl *DispositivoController) ObtenerPorID(gctx *gin.Context) {
	id := gctx.Param("id")

	dispositivo, err := ctrl.dispositivoFacade.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}
	if !middleware.CanAccessResource(gctx, dispositivo.ClienteID, dispositivo.EmpresaID, true) {
		gctx.Error(types.ThrowPower("No tienes acceso a este dispositivo"))
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dispositivo,
	})
}

func (ctrl *DispositivoController) ObtenerPorCliente(gctx *gin.Context) {
	clienteID := gctx.Param("clienteId")
	if middleware.IsClienteContext(gctx) && middleware.UserID(gctx) != clienteID {
		gctx.Error(types.ThrowPower("No tienes acceso a los dispositivos de este cliente"))
		return
	}

	dispositivos, err := ctrl.dispositivoFacade.ObtenerPorCliente(gctx.Request.Context(), clienteID)
	if err != nil {
		gctx.Error(err)
		return
	}
	for _, dispositivo := range dispositivos {
		if !middleware.CanAccessResource(gctx, dispositivo.ClienteID, dispositivo.EmpresaID, true) {
			gctx.Error(types.ThrowPower("No tienes acceso a los dispositivos de este cliente"))
			return
		}
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dispositivos,
	})
}

func (ctrl *DispositivoController) Crear(gctx *gin.Context) {
	if !middleware.IsEmpresaContext(gctx) || middleware.EmpresaID(gctx) == "" {
		gctx.Error(types.ThrowPower("Solo una empresa autenticada puede crear dispositivos"))
		return
	}

	var r recipe.CrearDispositivoRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}
	r.EmpresaID = middleware.EmpresaID(gctx)

	dispositivo, err := ctrl.dispositivoFacade.Crear(gctx.Request.Context(), &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    dispositivo,
		"message": "Dispositivo creado correctamente",
	})
}

func (ctrl *DispositivoController) Actualizar(gctx *gin.Context) {
	id := gctx.Param("id")

	dispositivoActual, err := ctrl.dispositivoFacade.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}
	if !middleware.CanAccessResource(gctx, dispositivoActual.ClienteID, dispositivoActual.EmpresaID, false) {
		gctx.Error(types.ThrowPower("No tienes permisos para actualizar este dispositivo"))
		return
	}

	var r recipe.ActualizarDispositivoRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	dispositivo, err := ctrl.dispositivoFacade.Actualizar(gctx.Request.Context(), id, &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dispositivo,
		"message": "Dispositivo actualizado correctamente",
	})
}

func (ctrl *DispositivoController) AsignarCliente(gctx *gin.Context) {
	id := gctx.Param("id")

	dispositivoActual, err := ctrl.dispositivoFacade.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}
	if !middleware.CanAccessResource(gctx, dispositivoActual.ClienteID, dispositivoActual.EmpresaID, false) {
		gctx.Error(types.ThrowPower("No tienes permisos para asignar este dispositivo"))
		return
	}

	var r recipe.AsignarDispositivoRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	dispositivo, err := ctrl.dispositivoFacade.AsignarCliente(gctx.Request.Context(), id, r.ClienteID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dispositivo,
		"message": "Dispositivo asignado correctamente",
	})
}

func (ctrl *DispositivoController) DesasignarCliente(gctx *gin.Context) {
	id := gctx.Param("id")

	dispositivoActual, err := ctrl.dispositivoFacade.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}
	if !middleware.CanAccessResource(gctx, dispositivoActual.ClienteID, dispositivoActual.EmpresaID, false) {
		gctx.Error(types.ThrowPower("No tienes permisos para desasignar este dispositivo"))
		return
	}

	dispositivo, err := ctrl.dispositivoFacade.AsignarCliente(gctx.Request.Context(), id, "")
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dispositivo,
		"message": "Dispositivo desasignado correctamente",
	})
}

func (ctrl *DispositivoController) ActualizarLectura(gctx *gin.Context) {
	numeroDispositivo := gctx.Param("numeroDispositivo")

	var r recipe.ActualizarLecturaRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	err := ctrl.dispositivoFacade.ActualizarUltimaLectura(gctx.Request.Context(), numeroDispositivo, &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Lectura actualizada correctamente",
	})
}

func (ctrl *DispositivoController) CambiarEstado(gctx *gin.Context) {
	id := gctx.Param("id")

	dispositivoActual, err := ctrl.dispositivoFacade.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}
	if !middleware.CanAccessResource(gctx, dispositivoActual.ClienteID, dispositivoActual.EmpresaID, false) {
		gctx.Error(types.ThrowPower("No tienes permisos para cambiar este dispositivo"))
		return
	}

	var r recipe.CambiarEstadoDispositivoRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	err = ctrl.dispositivoFacade.CambiarEstado(gctx.Request.Context(), id, &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Estado actualizado correctamente",
	})
}

func (ctrl *DispositivoController) Eliminar(gctx *gin.Context) {
	id := gctx.Param("id")

	dispositivoActual, err := ctrl.dispositivoFacade.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}
	if !middleware.CanAccessResource(gctx, dispositivoActual.ClienteID, dispositivoActual.EmpresaID, false) {
		gctx.Error(types.ThrowPower("No tienes permisos para eliminar este dispositivo"))
		return
	}

	err = ctrl.dispositivoFacade.Eliminar(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dispositivo eliminado correctamente",
	})
}

func (ctrl *DispositivoController) RecibirLecturaIoT(gctx *gin.Context) {
	var payload struct {
		DeviceID       string   `json:"deviceId"`
		Voltaje        float64  `json:"voltaje"`
		Corriente      float64  `json:"corriente"`
		Potencia       float64  `json:"potencia"`
		Energia        float64  `json:"energia"`
		Frecuencia     float64  `json:"frecuencia"`
		FactorPotencia float64  `json:"factorPotencia"`
		Latitud        *float64 `json:"latitud"`
		Longitud       *float64 `json:"longitud"`
		Timestamp      int64    `json:"timestamp"`
	}

	if err := gctx.ShouldBindJSON(&payload); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido"})
		return
	}

	payload.DeviceID = strings.TrimSpace(payload.DeviceID)
	if tokenDeviceID, ok := gctx.Get("iot_device_id"); ok {
		deviceID, _ := tokenDeviceID.(string)
		if deviceID != "" {
			if payload.DeviceID != "" && payload.DeviceID != deviceID {
				gctx.JSON(http.StatusForbidden, gin.H{"error": "token no corresponde al dispositivo"})
				return
			}
			payload.DeviceID = deviceID
		}
	}
	if payload.DeviceID == "" {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": "deviceId requerido"})
		return
	}

	// Idempotencia: si el ESP32 reenvía una lectura ya recibida (mismo deviceId
	// + timestamp del dispositivo tras un corte de red), la descartamos para no
	// duplicar consumo ni métricas. Solo aplica si el dispositivo mandó su epoch.
	if iot.IsDuplicateReading(payload.DeviceID, payload.Timestamp) {
		gctx.JSON(http.StatusOK, gin.H{"ok": true, "duplicate": true})
		return
	}

	lectura := &entities.LecturaDispositivo{
		Voltage:        payload.Voltaje,
		Current:        payload.Corriente,
		ActivePower:    payload.Potencia,
		Energy:         payload.Energia,
		Frecuencia:     payload.Frecuencia,
		FactorPotencia: payload.FactorPotencia,
		Timestamp:      parseIoTTimestamp(payload.Timestamp),
	}

	// Solo tratamos el GPS como ubicación válida cuando ambos valores están
	// presentes, dentro de rango y no son (0,0) (típico "sin fix" del ESP32).
	lat, lng := normalizeIoTUbicacion(payload.Latitud, payload.Longitud)

	if ingestor := iot.DefaultReadingIngestor(); ingestor != nil {
		if !ingestor.Enqueue(iot.Reading{
			DeviceID:   payload.DeviceID,
			Lectura:    lectura,
			ReceivedAt: time.Now().UTC(),
			Latitud:    lat,
			Longitud:   lng,
		}) {
			gctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "cola de lecturas saturada"})
			return
		}
		gctx.JSON(http.StatusOK, gin.H{"ok": true, "queued": true})
		return
	}

	err := ctrl.dispositivoFacade.ActualizarLectura(
		gctx.Request.Context(),
		payload.DeviceID,
		payload.Voltaje,
		payload.Corriente,
		payload.Potencia,
		payload.Energia,
		payload.Frecuencia,
		payload.FactorPotencia,
		lat,
		lng,
	)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	gctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// normalizeIoTUbicacion valida las coordenadas GPS reportadas por el ESP32.
// Devuelve (nil, nil) si faltan, están fuera de rango o son (0,0) — que suele
// indicar que el módulo GPS aún no tiene fix.
func normalizeIoTUbicacion(lat, lng *float64) (*float64, *float64) {
	if lat == nil || lng == nil {
		return nil, nil
	}
	if *lat < -90 || *lat > 90 || *lng < -180 || *lng > 180 {
		return nil, nil
	}
	if *lat == 0 && *lng == 0 {
		return nil, nil
	}
	return lat, lng
}

func parseIoTTimestamp(timestamp int64) time.Time {
	if timestamp <= 0 {
		return time.Now().UTC()
	}
	if timestamp > 1_000_000_000_000 {
		return time.UnixMilli(timestamp).UTC()
	}
	return time.Unix(timestamp, 0).UTC()
}

func (ctrl *DispositivoController) ComandoEjecutado(gctx *gin.Context) {
	var payload struct {
		DeviceID  string `json:"deviceId"`
		Comando   string `json:"comando"`
		Estado    string `json:"estado"`
		Timestamp int64  `json:"timestamp"`
	}

	if err := gctx.ShouldBindJSON(&payload); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido"})
		return
	}

	gctx.JSON(http.StatusOK, gin.H{"ok": true})
}
