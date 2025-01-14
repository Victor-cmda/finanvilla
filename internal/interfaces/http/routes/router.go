package routes

import (
	"finanvilla/internal/interfaces/http/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(userHandler *handlers.UserHandler) *gin.Engine {
	router := gin.Default()

	// Grupo de rotas para API v1
	v1 := router.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUserByID)
			users.GET("/", userHandler.ListUsers)
			users.PUT("/:id/settings", userHandler.UpdateSettings)
		}
	}

	return router
}
