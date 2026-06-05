package controllers

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/config"
	"electric-backend/infrastructure/entities"
	"electric-backend/infrastructure/leads"
	"electric-backend/infrastructure/middleware"
	"electric-backend/infrastructure/validation"
	"electric-backend/types"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	defaultLeadPageSize = 50
	maxLeadPageSize     = 1000
)

type LeadController struct {
	collection *mongo.Collection
	httpClient *http.Client
}

func NewLeadController() *LeadController {
	return &LeadController{
		collection: config.MongoDB.Collection(entities.LeadEntity{}.CollectionName()),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (ctrl *LeadController) SetupRoutes(router *gin.RouterGroup) {
	leads := router.Group("/leads")
	{
		leads.POST("", ctrl.Crear)

		leads.Use(
			middleware.AuthMiddleware(),
			middleware.CSRFMiddleware(),
			middleware.RequireRole(
				"empresa",
				"admin",
				"superadmin",
				"super_admin",
				"EMPRESA_ADMIN",
				"EMPRESA_OPERADOR",
				"EMPRESA_SOPORTE",
				"EMPRESA_FINANCIERO",
			),
		)
		{
			leads.GET("", ctrl.ObtenerTodas)
			leads.PUT("/:id/status", ctrl.ActualizarEstado)
		}
	}
}

func (ctrl *LeadController) Crear(gctx *gin.Context) {
	var r recipe.CrearLeadRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	if err := ctrl.verifyTurnstile(gctx.Request.Context(), r.TurnstileToken, gctx.ClientIP()); err != nil {
		gctx.Error(types.ThrowRecipe(err.Error(), "turnstileToken"))
		return
	}

	now := time.Now().UTC()
	lead := entities.LeadEntity{
		ID:           primitive.NewObjectID(),
		Type:         r.Type,
		Status:       "new",
		Name:         validation.SanitizeString(r.Name),
		Email:        validation.SanitizeEmail(r.Email),
		Organization: validation.SanitizeString(r.Organization),
		Message:      validation.SanitizeString(r.Message),
		Extra:        sanitizeLeadExtra(r.Extra),
		IPAddress:    gctx.ClientIP(),
		UserAgent:    validation.SanitizeString(gctx.Request.UserAgent()),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if lead.Name == "" || lead.Email == "" {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	if ingestor := leads.DefaultLeadIngestor(); ingestor != nil {
		if !ingestor.Enqueue(lead) {
			gctx.JSON(http.StatusServiceUnavailable, types.ApiResponse{
				Success: false,
				Error:   "Servicio temporalmente saturado",
			})
			return
		}
		gctx.JSON(http.StatusCreated, types.ApiResponse{
			Success: true,
			Data:    lead,
			Message: "Lead recibido correctamente",
		})
		return
	}

	if _, err := ctrl.collection.InsertOne(gctx.Request.Context(), lead); err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusCreated, types.ApiResponse{
		Success: true,
		Data:    lead,
		Message: "Lead recibido correctamente",
	})
}

func (ctrl *LeadController) ObtenerTodas(gctx *gin.Context) {
	offset, limit, page := leadPagination(gctx)
	filter, err := buildLeadFilter(gctx)
	if err != nil {
		gctx.Error(err)
		return
	}

	total, err := ctrl.collection.CountDocuments(gctx.Request.Context(), filter)
	if err != nil {
		gctx.Error(err)
		return
	}

	cursor, err := ctrl.collection.Find(
		gctx.Request.Context(),
		filter,
		options.Find().
			SetSkip(int64(offset)).
			SetLimit(int64(limit)).
			SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		gctx.Error(err)
		return
	}
	defer cursor.Close(gctx.Request.Context())

	items := make([]entities.LeadEntity, 0, limit)
	if err := cursor.All(gctx.Request.Context(), &items); err != nil {
		gctx.Error(err)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data: gin.H{
			"items":    items,
			"total":    total,
			"offset":   offset,
			"limit":    limit,
			"hasMore":  int64(offset+len(items)) < total,
			"page":     page,
			"pageSize": limit,
		},
		Pagination: &types.PaginationResponse{
			CurrentPage:  page,
			TotalPages:   totalPages,
			TotalItems:   int(total),
			ItemsPerPage: limit,
		},
	})
}

func (ctrl *LeadController) ActualizarEstado(gctx *gin.Context) {
	id, err := primitive.ObjectIDFromHex(strings.TrimSpace(gctx.Param("id")))
	if err != nil {
		gctx.Error(types.ThrowRecipe("ID inválido", "id"))
		return
	}

	var r recipe.ActualizarEstadoLeadRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	update := bson.M{
		"$set": bson.M{
			"status":    r.Status,
			"updatedAt": time.Now().UTC(),
		},
	}

	result := ctrl.collection.FindOneAndUpdate(
		gctx.Request.Context(),
		bson.M{"_id": id},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var lead entities.LeadEntity
	if err := result.Decode(&lead); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			gctx.Error(types.ThrowData("Lead no encontrado"))
			return
		}
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, types.ApiResponse{
		Success: true,
		Data:    lead,
		Message: "Estado actualizado correctamente",
	})
}

func leadPagination(gctx *gin.Context) (offset int, limit int, page int) {
	limit, _ = strconv.Atoi(gctx.DefaultQuery("limit", strconv.Itoa(defaultLeadPageSize)))
	if limit < 1 {
		limit = defaultLeadPageSize
	}
	if limit > maxLeadPageSize {
		limit = maxLeadPageSize
	}

	if rawOffset := strings.TrimSpace(gctx.Query("offset")); rawOffset != "" {
		offset, _ = strconv.Atoi(rawOffset)
		if offset < 0 {
			offset = 0
		}
		page = (offset / limit) + 1
		return offset, limit, page
	}

	page, _ = strconv.Atoi(gctx.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset = (page - 1) * limit
	return offset, limit, page
}

func buildLeadFilter(gctx *gin.Context) (bson.M, error) {
	filter := bson.M{}

	if leadType := strings.TrimSpace(gctx.Query("type")); leadType != "" && leadType != "all" {
		if leadType != "investor" && leadType != "distributor" {
			return nil, types.ThrowRecipe("Tipo de lead inválido", "type")
		}
		filter["type"] = leadType
	}

	if status := strings.TrimSpace(gctx.Query("status")); status != "" && status != "all" {
		switch status {
		case "new", "contacted", "qualified", "discarded":
			filter["status"] = status
		default:
			return nil, types.ThrowRecipe("Estado de lead inválido", "status")
		}
	}

	if from := strings.TrimSpace(gctx.Query("from")); from != "" {
		fromTime, err := parseLeadDate(from, false)
		if err != nil {
			return nil, types.ThrowRecipe("Fecha desde inválida", "from")
		}
		filter["createdAt"] = bson.M{"$gte": fromTime}
	}

	if to := strings.TrimSpace(gctx.Query("to")); to != "" {
		toTime, err := parseLeadDate(to, true)
		if err != nil {
			return nil, types.ThrowRecipe("Fecha hasta inválida", "to")
		}
		if createdAt, ok := filter["createdAt"].(bson.M); ok {
			createdAt["$lte"] = toTime
		} else {
			filter["createdAt"] = bson.M{"$lte": toTime}
		}
	}

	if search := validation.SanitizeString(gctx.Query("search")); search != "" {
		pattern := regexp.QuoteMeta(search)
		rx := primitive.Regex{Pattern: pattern, Options: "i"}
		filter["$or"] = bson.A{
			bson.M{"name": rx},
			bson.M{"email": rx},
			bson.M{"organization": rx},
			bson.M{"message": rx},
		}
	}

	return filter, nil
}

func parseLeadDate(value string, endOfDay bool) (time.Time, error) {
	if ts, err := time.Parse(time.RFC3339, value); err == nil {
		return ts.UTC(), nil
	}
	ts, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, err
	}
	if endOfDay {
		ts = ts.Add(24*time.Hour - time.Nanosecond)
	}
	return ts.UTC(), nil
}

func sanitizeLeadExtra(extra map[string]interface{}) map[string]interface{} {
	if len(extra) == 0 {
		return nil
	}

	out := make(map[string]interface{}, len(extra))
	count := 0
	for key, value := range extra {
		if count >= 20 {
			break
		}

		cleanKey := validation.SanitizeString(key)
		cleanKey = strings.TrimLeft(cleanKey, "$.")
		if cleanKey == "" || strings.Contains(cleanKey, ".") {
			continue
		}

		if cleanValue, ok := sanitizeLeadExtraValue(value, 0); ok {
			out[cleanKey] = cleanValue
			count++
		}
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

func sanitizeLeadExtraValue(value interface{}, depth int) (interface{}, bool) {
	if depth > 2 {
		return nil, false
	}

	switch v := value.(type) {
	case string:
		clean := validation.SanitizeString(v)
		if len(clean) > 500 {
			clean = clean[:500]
		}
		return clean, true
	case bool:
		return v, true
	case float64:
		return v, true
	case int:
		return v, true
	case int64:
		return v, true
	case nil:
		return nil, true
	case []interface{}:
		if len(v) > 20 {
			v = v[:20]
		}
		out := make([]interface{}, 0, len(v))
		for _, item := range v {
			if clean, ok := sanitizeLeadExtraValue(item, depth+1); ok {
				out = append(out, clean)
			}
		}
		return out, true
	case map[string]interface{}:
		return sanitizeLeadExtra(v), true
	default:
		return nil, false
	}
}

func (ctrl *LeadController) verifyTurnstile(ctx context.Context, token string, remoteIP string) error {
	secret := strings.TrimSpace(os.Getenv("TURNSTILE_SECRET_KEY"))
	if secret == "" {
		return nil
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("Completa la verificación")
	}

	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return errors.New("No se pudo verificar el desafío")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := ctrl.httpClient.Do(req)
	if err != nil {
		return errors.New("No se pudo verificar el desafío")
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return errors.New("No se pudo verificar el desafío")
	}
	if resp.StatusCode != http.StatusOK || !result.Success {
		return errors.New("Verificación inválida")
	}
	return nil
}
