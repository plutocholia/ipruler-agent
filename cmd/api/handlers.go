package api

import (
	"io"
	"log"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/plutocholia/ipruler/cmd/ipruler"
)

var (
	configLifeCycle *ipruler.ConfigLifeCycle
	lock            sync.Mutex
	data            []byte
)

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "good, thank you for asking",
	})
}

func update(c *gin.Context) {
	lock.Lock()
	defer lock.Unlock()

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}

	err = configLifeCycle.WaveSync(body)
	if _err, ok := err.(*ipruler.EmptyConfig); ok {
		log.Println(_err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": _err.Error()})
		return
	}
	// configLifeCycle.PersistState()
	data = body
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func init() {
	configLifeCycle = ipruler.CreateConfigLifeCycle()
}

func BackgroundSync(configReloadDuration uint) {
	var oldData []byte

	clc := ipruler.CreateConfigLifeCycle()

	for {
		lock.Lock()
		if !reflect.DeepEqual(data, oldData) {
			log.Println("detected changes in config")
			oldData = data
		}
		clc.WaveSync(data)
		lock.Unlock()
		time.Sleep(time.Duration(configReloadDuration) * time.Second)
	}
}
