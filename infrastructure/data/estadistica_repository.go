package data

import (
"context"
"electric-backend/config"

"go.mongodb.org/mongo-driver/bson"
"go.mongodb.org/mongo-driver/bson/primitive"
)

type EstadisticaRepository struct{}

func NewEstadisticaRepository() *EstadisticaRepository {
	return &EstadisticaRepository{}
}

func (r *EstadisticaRepository) ObtenerConsumoCliente(ctx context.Context, clienteID string) (map[string]interface{}, error) {
	clienteObjectID, _ := primitive.ObjectIDFromHex(clienteID)

	collection := config.MongoDB.Collection("dispositivos")
	cursor, err := collection.Find(ctx, bson.M{"clienteId": clienteObjectID})
	if err != nil {
		return map[string]interface{}{
			"consumoTotal":    0,
			"costoTotal":      0,
			"promedioVoltage": 0,
			"promedioCurrent": 0,
		}, nil
	}
	defer cursor.Close(ctx)

	var totalConsumo, totalCosto, totalVoltage, totalCurrent float64
	var count int

	for cursor.Next(ctx) {
		var dispositivo bson.M
		if err := cursor.Decode(&dispositivo); err != nil {
			continue
		}

		if ultimaLectura, ok := dispositivo["ultimaLectura"].(bson.M); ok {
			if energy, ok := ultimaLectura["energy"].(float64); ok {
				totalConsumo += energy
			}
			if cost, ok := ultimaLectura["cost"].(float64); ok {
				totalCosto += cost
			}
			if voltage, ok := ultimaLectura["voltage"].(float64); ok {
				totalVoltage += voltage
			}
			if current, ok := ultimaLectura["current"].(float64); ok {
				totalCurrent += current
			}
			count++
		}
	}

	promedioVoltage := 0.0
	promedioCurrent := 0.0
	if count > 0 {
		promedioVoltage = totalVoltage / float64(count)
		promedioCurrent = totalCurrent / float64(count)
	}

	return map[string]interface{}{
		"consumoTotal":    totalConsumo,
		"costoTotal":      totalCosto,
		"promedioVoltage": promedioVoltage,
		"promedioCurrent": promedioCurrent,
	}, nil
}

func (r *EstadisticaRepository) ObtenerEstadisticasGlobales(ctx context.Context) (map[string]interface{}, error) {
	clientesCount, _ := config.MongoDB.Collection("clientes").CountDocuments(ctx, bson.M{})
	dispositivosCount, _ := config.MongoDB.Collection("dispositivos").CountDocuments(ctx, bson.M{})
	dispositivosActivos, _ := config.MongoDB.Collection("dispositivos").CountDocuments(ctx, bson.M{"activo": true})

	return map[string]interface{}{
		"totalClientes":       clientesCount,
		"totalDispositivos":   dispositivosCount,
		"dispositivosActivos": dispositivosActivos,
		"consumoTotalKwh":     0,
	}, nil
}
