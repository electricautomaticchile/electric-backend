package controllers

import (
	"electric-backend/infrastructure/data"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type IoTStatusController struct {
	dispositivoRepo *data.DispositivoRepository
}

func NewIoTStatusController(dispositivoRepo *data.DispositivoRepository) *IoTStatusController {
	return &IoTStatusController{dispositivoRepo: dispositivoRepo}
}

type DeviceStatus struct {
	ID                string    `json:"id"`
	NumeroDispositivo string    `json:"numeroDispositivo"`
	Nombre            string    `json:"nombre"`
	Estado            string    `json:"estado"`
	Online            bool      `json:"online"`
	UltimaLectura     *time.Time `json:"ultimaLectura,omitempty"`
	Voltaje           float64   `json:"voltaje,omitempty"`
	Corriente         float64   `json:"corriente,omitempty"`
	Potencia          float64   `json:"potencia,omitempty"`
	Energia           float64   `json:"energia,omitempty"`
	SenalGSM          string    `json:"senalGSM"`
}

// GetAllStatus GET /api/v1/iot/status — Estado de todos los dispositivos
func (ctrl *IoTStatusController) GetAllStatus(gctx *gin.Context) {
	empresaID := gctx.Query("empresaId")
	if empresaID == "" {
		gctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "empresaId requerido"})
		return
	}

	dispositivos, err := ctrl.dispositivoRepo.FindAll(gctx.Request.Context(), empresaID)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	ahora := time.Now()
	statuses := make([]DeviceStatus, 0, len(dispositivos))

	for _, d := range dispositivos {
		status := DeviceStatus{
			ID:                d.ID.Hex(),
			NumeroDispositivo: d.NumeroDispositivo,
			Nombre:            d.Nombre,
			Estado:            d.Estado,
			SenalGSM:          "desconocida",
		}

		if d.UltimaLectura != nil {
			ts := d.UltimaLectura.Timestamp
			status.UltimaLectura = &ts
			status.Voltaje = d.UltimaLectura.Voltage
			status.Corriente = d.UltimaLectura.Current
			status.Potencia = d.UltimaLectura.ActivePower
			status.Energia = d.UltimaLectura.Energy

			// Online si reportó en los últimos 10 minutos
			status.Online = ahora.Sub(ts) < 10*time.Minute

			// Estimar señal GSM por frecuencia de reportes
			if ahora.Sub(ts) < 2*time.Minute {
				status.SenalGSM = "excelente"
			} else if ahora.Sub(ts) < 5*time.Minute {
				status.SenalGSM = "buena"
			} else if ahora.Sub(ts) < 10*time.Minute {
				status.SenalGSM = "débil"
			} else {
				status.SenalGSM = "sin señal"
			}
		}

		statuses = append(statuses, status)
	}

	// Resumen
	var online, offline int
	for _, s := range statuses {
		if s.Online {
			online++
		} else {
			offline++
		}
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    statuses,
		"resumen": gin.H{
			"total":   len(statuses),
			"online":  online,
			"offline": offline,
		},
	})
}
