package middlewares

import (
	"finanvilla/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				errStr := ""
				switch v := err.(type) {
				case error:
					errStr = v.Error()
				case string:
					errStr = v
				default:
					errStr = "unknown error"
				}

				logger.Error("Panic recovered",
					zap.String("error_type", "panic"),
					zap.String("error", errStr),
					zap.String("path", c.Request.URL.Path),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})
			}
		}()

		c.Next()
	}
}
