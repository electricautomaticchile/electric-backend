package server

import (
	"electric-backend/domain/facades"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/arduino"
	"electric-backend/infrastructure/data"
	"electric-backend/infrastructure/middleware"
	"electric-backend/infrastructure/websocket"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter crea el router Gin con todos los middleware y rutas registradas.
func SetupRouter(
	facadeContainer *facades.FacadeContainer,
	svc *services.ServiceContainer,
	repos *data.DataContainer,
	wsHub *websocket.Hub,
	arduinoBridge *arduino.SerialBridge,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/api/ws/connect"},
	}))
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)
		c.Next()
	})
	router.Use(middleware.CompressionMiddleware())
	router.Use(middleware.AuditMiddleware(svc.AuditLogService))
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.EndpointRateLimitMiddleware(rateLimits()))

	registerHealthRoutes(router, wsHub, arduinoBridge)

	api := router.Group("/api")
	registerRoutes(api, facadeContainer, svc, repos, wsHub, arduinoBridge)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, types.ApiResponse{Success: false, Error: "Ruta no encontrada"})
	})

	return router
}
