package server

import (
	"electric-backend/config"
	"electric-backend/infrastructure/arduino"
	"electric-backend/infrastructure/middleware"
	"electric-backend/infrastructure/websocket"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

func registerHealthRoutes(router *gin.Engine, wsHub *websocket.Hub, arduinoBridge *arduino.SerialBridge) {
	router.GET("/health", func(c *gin.Context) {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
			"memory": gin.H{
				"alloc_mb":       memStats.Alloc / 1024 / 1024,
				"total_alloc_mb": memStats.TotalAlloc / 1024 / 1024,
				"sys_mb":         memStats.Sys / 1024 / 1024,
				"gc_cycles":      memStats.NumGC,
			},
			"goroutines": runtime.NumGoroutine(),
		})
	})

	router.GET("/api/admin/health", middleware.AuthMiddleware(), middleware.RequireRole("empresa", "admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "OK",
			"message":     "API Electricautomaticchile funcionando correctamente",
			"timestamp":   time.Now().Format(time.RFC3339),
			"version":     "2.0.0",
			"environment": config.AppConfig.Environment,
			"database":    gin.H{"connected": config.MongoDB != nil},
			"redis":       gin.H{"connected": config.RedisClient != nil},
			"websocket":   gin.H{"clients": wsHub.GetConnectedClients()},
			"arduino":     gin.H{"connected": arduinoBridge.IsConnected(), "devices": len(arduinoBridge.GetDevices())},
		})
	})
}
