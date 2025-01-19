package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"finanvilla/internal/domain/entities"
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
	refreshTokenRepo := repositories.NewPostgresRefreshTokenRepository(db)

	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(
		userService,
		refreshTokenRepo,
		cfg.JWT.Secret,
		cfg.JWT.RefreshSecret,
	)

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

	go startRefreshTokenCleanup(refreshTokenRepo)

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

	// Auto Migrate para a tabela de refresh tokens
	if err := db.AutoMigrate(&entities.RefreshToken{}); err != nil {
		return nil, fmt.Errorf("failed to migrate refresh_tokens table: %w", err)
	}

	return db, nil
}

func startRefreshTokenCleanup(repo *repositories.PostgresRefreshTokenRepository) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if err := repo.DeleteExpired(context.Background()); err != nil {
			log.Printf("Error cleaning up expired refresh tokens: %v", err)
		}
	}
}
