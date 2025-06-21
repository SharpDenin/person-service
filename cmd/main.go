package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"person-service/internal/config"
)

func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		logrus.Info("Health check requested")
		c.JSON(200, gin.H{"status": "ok"})
	})
	return router
}

func main() {
	initLogger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatal("Failed to load config: ", err)
	}
	gin.SetMode(cfg.GinMode)

	// Init router
	r := setupRouter()

	// Load server
	logrus.Infof("Server started on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		logrus.Fatal("Failed to start server: ", err)
	}
}
