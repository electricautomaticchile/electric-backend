package types

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FilterParams struct {
	Search      string
	DateFrom    string
	DateTo      string
	Status      string
	Type        string
	Active      *bool
	ClienteID   string
	EmpresaID   string
	Resuelta    *bool
	Estado      string
	CustomFilters map[string]interface{}
}

func (f *FilterParams) BuildMongoFilter() bson.M {
	filter := bson.M{}

	if f.Search != "" {
		searchRegex := primitive.Regex{Pattern: f.Search, Options: "i"}
		filter["$or"] = []bson.M{
			{"nombre": searchRegex},
			{"correo": searchRegex},
			{"numeroCliente": searchRegex},
			{"titulo": searchRegex},
			{"mensaje": searchRegex},
			{"asunto": searchRegex},
			{"descripcion": searchRegex},
		}
	}

	if f.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", f.DateFrom)
		if err == nil {
			if filter["fechaCreacion"] == nil {
				filter["fechaCreacion"] = bson.M{}
			}
			filter["fechaCreacion"].(bson.M)["$gte"] = dateFrom
		}
	}

	if f.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", f.DateTo)
		if err == nil {
			dateTo = dateTo.Add(24 * time.Hour)
			if filter["fechaCreacion"] == nil {
				filter["fechaCreacion"] = bson.M{}
			}
			filter["fechaCreacion"].(bson.M)["$lt"] = dateTo
		}
	}

	if f.Status != "" {
		filter["status"] = f.Status
	}

	if f.Type != "" {
		filter["tipo"] = f.Type
	}

	if f.Active != nil {
		filter["activo"] = *f.Active
	}

	if f.ClienteID != "" {
		clienteObjectID, err := primitive.ObjectIDFromHex(f.ClienteID)
		if err == nil {
			filter["clienteId"] = clienteObjectID
		}
	}

	if f.EmpresaID != "" {
		empresaObjectID, err := primitive.ObjectIDFromHex(f.EmpresaID)
		if err == nil {
			filter["empresaId"] = empresaObjectID
		}
	}

	if f.Resuelta != nil {
		filter["resuelta"] = *f.Resuelta
	}

	if f.Estado != "" {
		filter["estado"] = f.Estado
	}

	for key, value := range f.CustomFilters {
		filter[key] = value
	}

	return filter
}

func ParseBoolPtr(s string) *bool {
	if s == "" {
		return nil
	}
	lower := strings.ToLower(s)
	if lower == "true" || lower == "1" {
		val := true
		return &val
	}
	if lower == "false" || lower == "0" {
		val := false
		return &val
	}
	return nil
}
