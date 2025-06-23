package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"os"
	_ "person-service/docs"
	"person-service/internal/config"
	"person-service/internal/handler"
	"person-service/internal/repository"
	"person-service/internal/service"
)

// @title Person Service API
// @version 1.0
// @description REST API for managing persons with data enrichment from external APIs
// @host localhost:8080
// @BasePath /api
func initLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)
	return log
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		logrus.Info("Health check requested")
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return router
}

func main() {
	log := initLogger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}
	gin.SetMode(cfg.GinMode)

	// Test db connection
	ctx := context.Background()
	if err := repository.RunMigrations(
		ctx,
		cfg.DatabaseDSN,
		log,
	); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}
	db, err := repository.NewDB(ctx, cfg)
	if err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}
	defer db.Close()

	personRepo := repository.NewPersonRepository(db, log)
	personService := service.NewPersonService(personRepo, log)
	personHandler := handler.NewPersonHandler(personService, log)

	// Init router
	r := setupRouter()

	api := r.Group("/api")
	{
		api.POST("/person", personHandler.CreatePerson)
		api.GET("/person/:id", personHandler.GetPerson)
		api.GET("/people", personHandler.GetAll)
		api.PUT("/person/:id", personHandler.Update)
		api.DELETE("/person/:id", personHandler.Delete)
	}

	// Load server
	log.Infof("Server started on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
