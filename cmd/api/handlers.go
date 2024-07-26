package api

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/plutocholia/ipruler/cmd/ipruler"
)

var (
	configLifeCycle *ipruler.ConfigLifeCycle
)

func ip(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ip":        c.ClientIP(),
		"remote-ip": c.RemoteIP(),
	})
}

func update(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}

	if string(body) != "" {
		configLifeCycle.WeaveSync(body)
		// configLifeCycle.PersistState()
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"status": "BadConfigFile"})
}

func init() {
	configLifeCycle = ipruler.CreateConfigLifeCycle()
}
