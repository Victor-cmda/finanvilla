// cmd/api/main.go
package main

import (
	"fmt"
	"log"

	"finanvilla/internal/domain/services"
	"finanvilla/internal/infrastructure/repositories"
	"finanvilla/internal/interfaces/http/handlers"
	"finanvilla/internal/interfaces/http/routes"
	"finanvilla/pkg/config"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	AppVersion = "1.0.0"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := setupDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	userRepo := repositories.NewPostgresUserRepository(db)

	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userService, cfg.JWT.Secret)

	userHandler := handlers.NewUserHandler(userService)
	healthHandler := handlers.NewHealthHandler(cfg.Environment, AppVersion)
	authHandler := handlers.NewAuthHandler(authService)

	routerConfig := routes.RouterConfig{
		UserHandler:   userHandler,
		HealthHandler: healthHandler,
		AuthHandler:   authHandler,
		JWTSecret:     cfg.JWT.Secret,
	}

	router := routes.SetupRouter(routerConfig)

	log.Printf("Server starting on port %s in %s mode", cfg.Server.Port, cfg.Environment)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db, nil
}
