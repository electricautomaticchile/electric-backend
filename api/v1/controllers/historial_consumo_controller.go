package controllers

import (
	"context"
	"electric-backend/config"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type HistorialConsumoController struct{}

func NewHistorialConsumoController() *HistorialConsumoController {
	return &HistorialConsumoController{}
}

func (ctrl *HistorialConsumoController) SetupRoutes(router *gin.RouterGroup) {
	h := router.Group("/historial-consumo")
	h.Use(middleware.AuthMiddleware(), middleware.CSRFMiddleware())
	{
		h.GET("/:clienteId", ctrl.ObtenerHistorial)
		h.GET("/:clienteId/estadisticas", ctrl.ObtenerEstadisticas)
		h.POST("/seed", ctrl.SeedDatosPrueba)
	}
}

func (ctrl *HistorialConsumoController) ObtenerHistorial(gctx *gin.Context) {
	clienteId := gctx.Param("clienteId")
	agregacion := gctx.Query("agregacion") // vacío = sin agregación (lecturas crudas)
	limiteStr := gctx.DefaultQuery("limite", "100")
	desdeStr := gctx.Query("desde")
	hastaStr := gctx.Query("hasta")

	limite, _ := strconv.ParseInt(limiteStr, 10, 64)
	if limite <= 0 || limite > 500 {
		limite = 100
	}

	hasta := time.Now()
	desde := hasta.Add(-24 * time.Hour)

	switch agregacion {
	case "mes":
		desde = hasta.AddDate(-1, 0, 0)
	case "dia":
		desde = hasta.AddDate(0, 0, -30)
	}

	if desdeStr != "" {
		if t, err := time.Parse(time.RFC3339, desdeStr); err == nil {
			desde = t
		}
	}
	if hastaStr != "" {
		if t, err := time.Parse(time.RFC3339, hastaStr); err == nil {
			hasta = t
		}
	}

	clienteOID, err := primitive.ObjectIDFromHex(clienteId)
	if err != nil {
		gctx.Error(types.ThrowData("ID de cliente inválido"))
		return
	}

	col := config.MongoDB.Collection("lecturas")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchFilter := bson.M{
		"clienteId": clienteOID,
		"timestamp": bson.M{"$gte": desde, "$lte": hasta},
	}

	// Sin agregación: lecturas crudas (para "último valor")
	if agregacion == "" {
		opts := options.Find().
			SetSort(bson.D{{Key: "timestamp", Value: -1}}).
			SetLimit(limite)
		cursor, err := col.Find(ctx, matchFilter, opts)
		if err != nil {
			gctx.Error(err)
			return
		}
		defer cursor.Close(ctx)
		var resultados []bson.M
		if err := cursor.All(ctx, &resultados); err != nil {
			gctx.Error(err)
			return
		}
		if resultados == nil {
			resultados = []bson.M{}
		}
		// Mapear campos a los que espera el frontend
		mapped := make([]bson.M, len(resultados))
		for i, r := range resultados {
			mapped[i] = bson.M{
				"_id":            r["_id"],
				"timestamp":      r["timestamp"],
				"dispositivoId":  r["dispositivoId"],
				"clienteId":      r["clienteId"],
				"potenciaActiva": r["potencia"],
				"energia":        r["energia"],
				"costo":          r["costo"],
				"voltaje":        r["voltaje"],
				"corriente":      r["corriente"],
			}
		}
		gctx.JSON(http.StatusOK, types.ApiResponse{Success: true, Data: mapped})
		return
	}

	// Con agregación por período
	var groupFormat string
	switch agregacion {
	case "mes":
		groupFormat = "%Y-%m"
	case "dia":
		groupFormat = "%Y-%m-%d"
	default: // hora
		groupFormat = "%Y-%m-%dT%H:00"
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"periodo":       bson.M{"$dateToString": bson.M{"format": groupFormat, "date": "$timestamp"}},
				"dispositivoId": "$dispositivoId",
			},
			"timestamp":        bson.M{"$first": "$timestamp"},
			"timestampFinal":   bson.M{"$last": "$timestamp"},
			"potenciaPromedio": bson.M{"$avg": "$potencia"},
			"potenciaMaxima":   bson.M{"$max": "$potencia"},
			"potenciaMinima":   bson.M{"$min": "$potencia"},
			"energiaInicial":   bson.M{"$first": "$energia"},
			"energiaFinal":     bson.M{"$last": "$energia"},
			"energiaTotal":     bson.M{"$sum": "$energia"},
			"costoInicial":     bson.M{"$first": "$costo"},
			"costoFinal":       bson.M{"$last": "$costo"},
			"costoTotal":       bson.M{"$sum": "$costo"},
			"voltaje":          bson.M{"$avg": "$voltaje"},
			"corriente":        bson.M{"$avg": "$corriente"},
			"numeroMuestras":   bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id.periodo", Value: 1}}}},
		{{Key: "$limit", Value: limite}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		gctx.Error(err)
		return
	}
	defer cursor.Close(ctx)

	var resultados []bson.M
	if err := cursor.All(ctx, &resultados); err != nil {
		gctx.Error(err)
		return
	}
	if resultados == nil {
		resultados = []bson.M{}
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{Success: true, Data: resultados})
}

func (ctrl *HistorialConsumoController) ObtenerEstadisticas(gctx *gin.Context) {
	clienteId := gctx.Param("clienteId")
	desdeStr := gctx.Query("desde")
	hastaStr := gctx.Query("hasta")

	hasta := time.Now()
	desde := hasta.AddDate(0, -1, 0)

	if desdeStr != "" {
		if t, err := time.Parse(time.RFC3339, desdeStr); err == nil {
			desde = t
		}
	}
	if hastaStr != "" {
		if t, err := time.Parse(time.RFC3339, hastaStr); err == nil {
			hasta = t
		}
	}

	clienteOID, err := primitive.ObjectIDFromHex(clienteId)
	if err != nil {
		gctx.Error(types.ThrowData("ID de cliente inválido"))
		return
	}

	col := config.MongoDB.Collection("lecturas")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchFilter := bson.M{
		"clienteId": clienteOID,
		"timestamp": bson.M{"$gte": desde, "$lte": hasta},
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{Key: "$group", Value: bson.M{
			"_id":               nil,
			"consumoTotal":      bson.M{"$sum": "$energia"},
			"costoTotal":        bson.M{"$sum": "$costo"},
			"potenciaPromedio":  bson.M{"$avg": "$potencia"},
			"potenciaMaxima":    bson.M{"$max": "$potencia"},
			"potenciaMinima":    bson.M{"$min": "$potencia"},
			"voltajePromedio":   bson.M{"$avg": "$voltaje"},
			"corrientePromedio": bson.M{"$avg": "$corriente"},
			"numeroMuestras":    bson.M{"$sum": 1},
		}}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		gctx.Error(err)
		return
	}
	defer cursor.Close(ctx)

	var resultados []bson.M
	if err := cursor.All(ctx, &resultados); err != nil {
		gctx.Error(err)
		return
	}

	var stats bson.M
	if len(resultados) > 0 {
		stats = resultados[0]
		delete(stats, "_id")
	} else {
		stats = bson.M{
			"consumoTotal": 0, "costoTotal": 0,
			"potenciaPromedio": 0, "potenciaMaxima": 0, "potenciaMinima": 0,
			"voltajePromedio": 0, "corrientePromedio": 0, "numeroMuestras": 0,
		}
	}

	// Último registro
	var ultimaLectura bson.M
	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	col.FindOne(ctx, bson.M{"clienteId": clienteOID}, opts).Decode(&ultimaLectura)
	if ultimaLectura != nil {
		stats["ultimaLectura"] = ultimaLectura
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{Success: true, Data: stats})
}

// SeedDatosPrueba inserta lecturas de prueba para los últimos 7 días
func (ctrl *HistorialConsumoController) SeedDatosPrueba(gctx *gin.Context) {
	var body struct {
		ClienteID     string `json:"clienteId"`
		DispositivoID string `json:"dispositivoId"`
	}
	if err := gctx.ShouldBindJSON(&body); err != nil || body.ClienteID == "" {
		gctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "clienteId requerido"})
		return
	}

	clienteOID, err := primitive.ObjectIDFromHex(body.ClienteID)
	if err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "clienteId inválido"})
		return
	}

	// Buscar dispositivoId si no se pasó
	var dispositivoOID primitive.ObjectID
	if body.DispositivoID != "" {
		dispositivoOID, _ = primitive.ObjectIDFromHex(body.DispositivoID)
	} else {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		var disp bson.M
		config.MongoDB.Collection("dispositivos").FindOne(ctx2, bson.M{"clienteId": clienteOID}).Decode(&disp)
		if disp != nil {
			if oid, ok := disp["_id"].(primitive.ObjectID); ok {
				dispositivoOID = oid
			}
		}
	}

	col := config.MongoDB.Collection("lecturas")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	now := time.Now()
	docs := make([]interface{}, 0, 7*24)

	// 7 días × 24 horas = 168 lecturas (una por hora)
	energiaAcum := 0.0855 // partir desde el valor actual del dispositivo
	for d := 6; d >= 0; d-- {
		for h := 0; h < 24; h++ {
			ts := now.AddDate(0, 0, -d).Truncate(24 * time.Hour).Add(time.Duration(h) * time.Hour)
			// Simular variación realista: más consumo en horario diurno
			factor := 0.3 + 0.7*float64(h%12)/12.0
			potencia := 80.0 + 60.0*factor
			corriente := potencia / 220.0
			energia := potencia / 1000.0 // kWh por hora
			energiaAcum += energia
			costo := energiaAcum * 150.0 // tarifa $150/kWh

			docs = append(docs, bson.M{
				"timestamp":     ts,
				"dispositivoId": dispositivoOID,
				"clienteId":     clienteOID,
				"voltaje":       220.0,
				"corriente":     corriente,
				"potencia":      potencia,
				"energia":       energiaAcum,
				"costo":         costo,
			})
		}
	}

	result, err := col.InsertMany(ctx, docs)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success":    true,
		"insertados": len(result.InsertedIDs),
		"mensaje":    "Datos de prueba insertados correctamente",
	})
}
