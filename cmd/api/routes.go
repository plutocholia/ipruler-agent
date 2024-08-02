package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(app *gin.Engine) {
	app.GET("/health", health)
	app.POST("/update", update)
}
