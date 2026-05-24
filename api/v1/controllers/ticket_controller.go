package controllers

import (
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TicketController struct {
	ticketService *services.TicketService
}

func NewTicketController(ticketService *services.TicketService) *TicketController {
	return &TicketController{
		ticketService: ticketService,
	}
}

// SetupRoutes configura las rutas del controlador
func (ctrl *TicketController) SetupRoutes(router *gin.RouterGroup) {
	g := router.Group("/tickets")
	g.Use(middleware.AuthMiddleware(), middleware.CSRFMiddleware())
	g.GET("", ctrl.ObtenerTodos)
	g.GET("/:id", ctrl.ObtenerPorID)
	g.GET("/cliente/:clienteId", ctrl.ObtenerPorCliente)
	g.GET("/empresa/:empresaId", ctrl.ObtenerPorEmpresa)
	g.POST("", ctrl.Crear)
	g.PUT("/:id/responder", ctrl.AgregarRespuesta)
	g.PUT("/:id/estado", ctrl.ActualizarEstado)
	g.DELETE("/:id", ctrl.Eliminar)
}

func (ctrl *TicketController) ObtenerTodos(gctx *gin.Context) {
	page, _ := strconv.Atoi(gctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(gctx.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	skip := (page - 1) * limit

	tickets, err := ctrl.ticketService.ObtenerTodos(gctx.Request.Context())
	if err != nil {
		gctx.Error(err)
		return
	}

	total := len(tickets)
	end := skip + limit
	if end > total {
		end = total
	}

	paginatedTickets := tickets
	if skip < total {
		paginatedTickets = tickets[skip:end]
	} else {
		paginatedTickets = []*models.TicketModel{}
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    paginatedTickets,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	})
}

func (ctrl *TicketController) ObtenerPorID(gctx *gin.Context) {
	id := gctx.Param("id")

	ticket, err := ctrl.ticketService.ObtenerPorID(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ticket,
	})
}

func (ctrl *TicketController) Crear(gctx *gin.Context) {
	var r recipe.CrearTicketRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Datos inválidos: " + err.Error(),
		})
		return
	}

	ticket, err := ctrl.ticketService.Crear(gctx.Request.Context(), &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    ticket,
		"message": "Ticket creado correctamente",
	})
}

func (ctrl *TicketController) AgregarRespuesta(gctx *gin.Context) {
	id := gctx.Param("id")
	userID := gctx.Request.Context().Value(types.ContextKeyUserID).(string)

	var r recipe.AgregarRespuestaRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	err := ctrl.ticketService.AgregarRespuesta(gctx.Request.Context(), id, &r, userID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Respuesta agregada correctamente",
	})
}

func (ctrl *TicketController) ActualizarEstado(gctx *gin.Context) {
	id := gctx.Param("id")

	var r recipe.ActualizarEstadoTicketRecipe
	if err := gctx.ShouldBindJSON(&r); err != nil {
		gctx.Error(types.ThrowRecipe("Datos inválidos", ""))
		return
	}

	err := ctrl.ticketService.ActualizarEstado(gctx.Request.Context(), id, &r)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Estado actualizado correctamente",
	})
}

func (ctrl *TicketController) Eliminar(gctx *gin.Context) {
	id := gctx.Param("id")

	err := ctrl.ticketService.Eliminar(gctx.Request.Context(), id)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Ticket eliminado correctamente",
	})
}

func (ctrl *TicketController) ObtenerPorCliente(gctx *gin.Context) {
	clienteID := gctx.Param("clienteId")

	tickets, err := ctrl.ticketService.ObtenerPorCliente(gctx.Request.Context(), clienteID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tickets,
	})
}

func (ctrl *TicketController) ObtenerPorEmpresa(gctx *gin.Context) {
	empresaID := gctx.Param("empresaId")

	tickets, err := ctrl.ticketService.ObtenerPorEmpresa(gctx.Request.Context(), empresaID)
	if err != nil {
		gctx.Error(err)
		return
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tickets,
	})
}
