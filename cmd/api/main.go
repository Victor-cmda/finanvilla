package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"finanvilla/internal/domain/services"
	"finanvilla/internal/infrastructure/repositories"
	"finanvilla/internal/interfaces/http/handlers"
	"finanvilla/pkg/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type HealthStatus struct {
	Status      string    `json:"status"`
	Time        time.Time `json:"timestamp"`
	Environment string    `json:"environment"`

	// Informações do Sistema
	System struct {
		CPUUsage    float64 `json:"cpu_usage"`
		MemoryUsage struct {
			Total     uint64  `json:"total"`
			Used      uint64  `json:"used"`
			Free      uint64  `json:"free"`
			UsagePerc float64 `json:"usage_percentage"`
		} `json:"memory_usage"`
		DiskUsage struct {
			Total     uint64  `json:"total"`
			Used      uint64  `json:"used"`
			Free      uint64  `json:"free"`
			UsagePerc float64 `json:"usage_percentage"`
		} `json:"disk_usage"`
	} `json:"system"`

	// Informações da Aplicação
	Application struct {
		Version    string `json:"version"`
		GoVersion  string `json:"go_version"`
		Goroutines int    `json:"goroutines"`
		StartTime  string `json:"start_time"`
		UpTime     string `json:"uptime"`
	} `json:"application"`
}

var startTime = time.Now()

func main() {
	// Carrega configurações
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Define o modo do Gin baseado no ambiente
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Conecta ao banco de dados
	db, err := setupDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Configura o servidor
	server := setupServer(db)

	// Inicia o servidor
	log.Printf("Server starting on port %s in %s mode", cfg.Server.Port, cfg.Environment)
	if err := server.Run(":" + cfg.Server.Port); err != nil {
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

	// Configurações adicionais do banco de dados
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Define limites de conexão
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db, nil
}

func setupServer(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// Configuração do CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Middleware de recuperação de pânico
	router.Use(gin.Recovery())

	// Middleware de logging
	router.Use(gin.Logger())

	// Inicializa repositórios
	userRepo := repositories.NewPostgresUserRepository(db)

	// Inicializa serviços
	userService := services.NewUserService(userRepo)

	// Inicializa handlers
	userHandler := handlers.NewUserHandler(userService)

	// Configura rotas
	api := router.Group("/api/v1")
	{
		// Todas as rotas sem proteção
		api.GET("/health", healthCheck)

		// Rotas de usuários (agora sem proteção)
		users := api.Group("/users")
		{
			users.POST("/", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUserByID)
			users.GET("/", userHandler.ListUsers)
			users.PUT("/:id/settings", userHandler.UpdateSettings)
		}
	}

	return router
}

func healthCheck(c *gin.Context) {
	health := HealthStatus{
		Status: "healthy",
		Time:   time.Now(),
	}

	// Coleta métricas do sistema
	if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
		health.System.CPUUsage = cpuPercent[0]
	}

	if memInfo, err := mem.VirtualMemory(); err == nil {
		health.System.MemoryUsage.Total = memInfo.Total
		health.System.MemoryUsage.Used = memInfo.Used
		health.System.MemoryUsage.Free = memInfo.Free
		health.System.MemoryUsage.UsagePerc = memInfo.UsedPercent
	}

	if diskInfo, err := disk.Usage("/"); err == nil {
		health.System.DiskUsage.Total = diskInfo.Total
		health.System.DiskUsage.Used = diskInfo.Used
		health.System.DiskUsage.Free = diskInfo.Free
		health.System.DiskUsage.UsagePerc = diskInfo.UsedPercent
	}

	// Informações da aplicação
	health.Application.Version = "1.0.0" // Substitua pela versão real da sua aplicação
	health.Application.GoVersion = runtime.Version()
	health.Application.Goroutines = runtime.NumGoroutine()
	health.Application.StartTime = startTime.Format(time.RFC3339)
	health.Application.UpTime = time.Since(startTime).String()

	// Define o código de status HTTP com base no estado geral
	statusCode := 200
	if health.Status != "healthy" {
		statusCode = 503 // Service Unavailable
	}

	c.JSON(statusCode, health)
}
