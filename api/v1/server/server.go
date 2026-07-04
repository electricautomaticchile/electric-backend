package server

import (
	"electric-backend/domain/facades"
	"electric-backend/domain/services"
	"electric-backend/infrastructure/arduino"
	"electric-backend/infrastructure/data"
	"electric-backend/infrastructure/middleware"
	"electric-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter crea el router Gin con todos los middleware y rutas registradas.
func SetupRouter(
	facadeContainer *facades.FacadeContainer,
	svc *services.ServiceContainer,
	repos *data.DataContainer,
	arduinoBridge *arduino.SerialBridge,
) *gin.Engine {
	router := gin.New()
	// Logging estructurado con zerolog (JSON en producción).
	router.Use(middleware.RequestLogger(
		"/health",
		"/api/leads",
		"/api/iot/lectura",
		"/api/iot/comando-ejecutado",
	))
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

	registerHealthRoutes(router, arduinoBridge)

	api := router.Group("/api")
	registerRoutes(api, facadeContainer, svc, repos, arduinoBridge)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, types.ApiResponse{Success: false, Error: "Ruta no encontrada"})
	})

	return router
}
