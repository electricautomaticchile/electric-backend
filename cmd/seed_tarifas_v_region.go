package main

import (
	"context"
	"electric-backend/config"
	"electric-backend/domain/models"
	"electric-backend/infrastructure/data"
	"log"
	"time"
)

func main() {
	config.LoadConfig()

	if err := config.ConnectDatabase(config.AppConfig.MongoURI); err != nil {
		log.Fatalf("Error conectando a MongoDB: %v", err)
	}
	defer config.DisconnectDatabase()

	log.Println("✅ Conectado a MongoDB")

	tarifaRepo := data.NewTarifaRepository()
	ctx := context.Background()

	vigenciaDesde := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	vigenciaHasta := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

	comunesTransmision := 31.827
	comunesServicioBase := 0.855

	tarifas := []models.TarifaModel{
		{
			Distribuidora:         "Chilquinta",
			TipoTarifa:            "BT1",
			Comuna:                "Villa Alemana",
			RedTipo:               "C0 Aéreo",
			VigenciaDesde:         vigenciaDesde,
			VigenciaHasta:         vigenciaHasta,
			CargoEnergia:          187.204,
			CargoTransmision:      comunesTransmision,
			CargoServicioPublico:  comunesServicioBase,
			PrecioKwhBase:         187.204 + comunesTransmision + 21.962 + comunesServicioBase,
			PeajeDistribucion:     21.962,
			TramosEstabilizacion: models.TramosEstabilizacion{
				Hasta350Kwh:   0.0,
				Entre350Y500:  0.923,
				Mayor500Kwh:   2.076,
			},
			Activa: true,
		},
		{
			Distribuidora:         "Chilquinta",
			TipoTarifa:            "BT1",
			Comuna:                "Quilpué",
			RedTipo:               "C0 Aéreo",
			VigenciaDesde:         vigenciaDesde,
			VigenciaHasta:         vigenciaHasta,
			CargoEnergia:          109.783,
			CargoTransmision:      comunesTransmision,
			CargoServicioPublico:  comunesServicioBase,
			PrecioKwhBase:         109.783 + comunesTransmision + 10.985 + comunesServicioBase,
			PeajeDistribucion:     10.985,
			TramosEstabilizacion: models.TramosEstabilizacion{
				Hasta350Kwh:   0.0,
				Entre350Y500:  0.923,
				Mayor500Kwh:   2.076,
			},
			Activa: true,
		},
		{
			Distribuidora:         "Chilquinta",
			TipoTarifa:            "BT1",
			Comuna:                "La Calera",
			RedTipo:               "C0 Aéreo",
			VigenciaDesde:         vigenciaDesde,
			VigenciaHasta:         vigenciaHasta,
			CargoEnergia:          187.204,
			CargoTransmision:      comunesTransmision,
			CargoServicioPublico:  comunesServicioBase,
			PrecioKwhBase:         187.204 + comunesTransmision + 21.962 + comunesServicioBase,
			PeajeDistribucion:     21.962,
			TramosEstabilizacion: models.TramosEstabilizacion{
				Hasta350Kwh:   0.0,
				Entre350Y500:  0.923,
				Mayor500Kwh:   2.076,
			},
			Activa: true,
		},
		{
			Distribuidora:         "Chilquinta",
			TipoTarifa:            "BT1",
			Comuna:                "Limache",
			RedTipo:               "C0 Aéreo",
			VigenciaDesde:         vigenciaDesde,
			VigenciaHasta:         vigenciaHasta,
			CargoEnergia:          177.941,
			CargoTransmision:      comunesTransmision,
			CargoServicioPublico:  comunesServicioBase,
			PrecioKwhBase:         177.941 + comunesTransmision + 20.648 + comunesServicioBase,
			PeajeDistribucion:     20.648,
			TramosEstabilizacion: models.TramosEstabilizacion{
				Hasta350Kwh:   0.0,
				Entre350Y500:  0.923,
				Mayor500Kwh:   2.076,
			},
			Activa: true,
		},
		{
			Distribuidora:         "Chilquinta",
			TipoTarifa:            "BT1",
			Comuna:                "Quillota",
			RedTipo:               "C0 Aéreo",
			VigenciaDesde:         vigenciaDesde,
			VigenciaHasta:         vigenciaHasta,
			CargoEnergia:          187.204,
			CargoTransmision:      comunesTransmision,
			CargoServicioPublico:  comunesServicioBase,
			PrecioKwhBase:         187.204 + comunesTransmision + 21.962 + comunesServicioBase,
			PeajeDistribucion:     21.962,
			TramosEstabilizacion: models.TramosEstabilizacion{
				Hasta350Kwh:   0.0,
				Entre350Y500:  0.923,
				Mayor500Kwh:   2.076,
			},
			Activa: true,
		},
		{
			Distribuidora:         "Chilquinta",
			TipoTarifa:            "BT1",
			Comuna:                "Valparaíso",
			RedTipo:               "ZS1",
			VigenciaDesde:         vigenciaDesde,
			VigenciaHasta:         vigenciaHasta,
			CargoEnergia:          187.204,
			CargoTransmision:      comunesTransmision,
			CargoServicioPublico:  comunesServicioBase,
			PrecioKwhBase:         187.204 + comunesTransmision + 21.962 + comunesServicioBase,
			PeajeDistribucion:     21.962,
			TramosEstabilizacion: models.TramosEstabilizacion{
				Hasta350Kwh:   0.0,
				Entre350Y500:  0.923,
				Mayor500Kwh:   2.076,
			},
			Activa: true,
		},
		{
			Distribuidora:         "Chilquinta",
			TipoTarifa:            "BT1",
			Comuna:                "Viña del Mar",
			RedTipo:               "ZS2",
			VigenciaDesde:         vigenciaDesde,
			VigenciaHasta:         vigenciaHasta,
			CargoEnergia:          187.204,
			CargoTransmision:      comunesTransmision,
			CargoServicioPublico:  comunesServicioBase,
			PrecioKwhBase:         187.204 + comunesTransmision + 21.962 + comunesServicioBase,
			PeajeDistribucion:     21.962,
			TramosEstabilizacion: models.TramosEstabilizacion{
				Hasta350Kwh:   0.0,
				Entre350Y500:  0.923,
				Mayor500Kwh:   2.076,
			},
			Activa: true,
		},
		{
			Distribuidora:         "Chilquinta",
			TipoTarifa:            "BT41",
			Comuna:                "Villa Alemana",
			RedTipo:               "C0 Aéreo",
			VigenciaDesde:         vigenciaDesde,
			VigenciaHasta:         vigenciaHasta,
			CargoEnergia:          187.204,
			CargoTransmision:      comunesTransmision,
			CargoServicioPublico:  comunesServicioBase,
			PrecioKwhBase:         187.204 + comunesTransmision + 21.962 + comunesServicioBase,
			PeajeDistribucion:     21.962,
			TramosEstabilizacion: models.TramosEstabilizacion{
				Hasta350Kwh:   0.0,
				Entre350Y500:  0.0,
				Mayor500Kwh:   0.0,
			},
			Activa: true,
		},
		{
			Distribuidora:         "Chilquinta",
			TipoTarifa:            "BT41",
			Comuna:                "Quilpué",
			RedTipo:               "C0 Aéreo",
			VigenciaDesde:         vigenciaDesde,
			VigenciaHasta:         vigenciaHasta,
			CargoEnergia:          109.783,
			CargoTransmision:      comunesTransmision,
			CargoServicioPublico:  comunesServicioBase,
			PrecioKwhBase:         109.783 + comunesTransmision + 10.985 + comunesServicioBase,
			PeajeDistribucion:     10.985,
			TramosEstabilizacion: models.TramosEstabilizacion{
				Hasta350Kwh:   0.0,
				Entre350Y500:  0.0,
				Mayor500Kwh:   0.0,
			},
			Activa: true,
		},
		{
			Distribuidora:         "Chilquinta",
			TipoTarifa:            "BT41",
			Comuna:                "La Calera",
			RedTipo:               "C0 Aéreo",
			VigenciaDesde:         vigenciaDesde,
			VigenciaHasta:         vigenciaHasta,
			CargoEnergia:          187.204,
			CargoTransmision:      comunesTransmision,
			CargoServicioPublico:  comunesServicioBase,
			PrecioKwhBase:         187.204 + comunesTransmision + 21.962 + comunesServicioBase,
			PeajeDistribucion:     21.962,
			TramosEstabilizacion: models.TramosEstabilizacion{
				Hasta350Kwh:   0.0,
				Entre350Y500:  0.0,
				Mayor500Kwh:   0.0,
			},
			Activa: true,
		},
	}

	for _, tarifa := range tarifas {
		if err := tarifaRepo.Create(ctx, &tarifa); err != nil {
			log.Printf("❌ Error al crear tarifa %s %s: %v", tarifa.Comuna, tarifa.TipoTarifa, err)
			continue
		}
		log.Printf("✅ Tarifa creada: %s - %s (%s) - $%.3f/kWh", 
			tarifa.Comuna, tarifa.TipoTarifa, tarifa.RedTipo, tarifa.PrecioKwhBase)
	}

	log.Println("✅ Seed completado - Tarifas V Región cargadas")
	log.Println("✅ Desconectado de MongoDB")
}
