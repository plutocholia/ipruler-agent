package api

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/plutocholia/ipruler/cmd/ipruler"
)

var (
	configLifeCycle *ipruler.ConfigLifeCycle
)

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "good, thank you for asking",
	})
}

func update(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}

	err = configLifeCycle.WeaveSync(body)
	if _err, ok := err.(*ipruler.EmptyConfig); ok {
		log.Println(_err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": _err.Error()})
		return
	}
	// configLifeCycle.PersistState()
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func init() {
	configLifeCycle = ipruler.CreateConfigLifeCycle()
}
