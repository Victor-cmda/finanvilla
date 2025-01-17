// internal/interfaces/http/routes/routes.go
package routes

import (
	"finanvilla/internal/interfaces/http/handlers"
	"finanvilla/internal/interfaces/http/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	UserHandler   *handlers.UserHandler
	HealthHandler *handlers.HealthHandler
	AuthHandler   *handlers.AuthHandler
	JWTSecret     string
}

func SetupRouter(config RouterConfig) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	api := router.Group("/api/v1")
	{
		api.GET("/health", config.HealthHandler.Check)

		auth := api.Group("/auth")
		{
			auth.POST("/register", config.AuthHandler.Register)
			auth.POST("/login", config.AuthHandler.Login)
		}

		protected := api.Group("")
		protected.Use(middlewares.AuthMiddleware(config.JWTSecret))
		{
			users := protected.Group("/users")
			{
				users.POST("/", config.UserHandler.CreateUser)
				users.GET("/:id", config.UserHandler.GetUserByID)
				users.GET("/", config.UserHandler.ListUsers)
				users.PUT("/:id/settings", config.UserHandler.UpdateSettings)
			}

		}
	}

	return router
}
