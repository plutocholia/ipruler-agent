package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(app *gin.Engine) {
	app.GET("/", ip)
	app.POST("/update", update)
}
