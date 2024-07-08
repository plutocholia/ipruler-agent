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

	configLifeCycle.Update(body)
	ipruler.SyncState(configLifeCycle)
	ipruler.PersistState(configLifeCycle)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func init() {
	configLifeCycle = ipruler.CreateConfigLifeCycle()
}
