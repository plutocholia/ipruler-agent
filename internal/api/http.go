package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/plutocholia/ipruler/internal/ipruler"
)

type HttpApi struct {
	configLifeCycle *ipruler.ConfigLifeCycle
	lock            sync.Mutex
	data            []byte
}

func (a *HttpApi) setupRoutes(app *gin.Engine) {
	app.GET("/health", a.health)
	app.POST("/update", a.update)
}

func (a *HttpApi) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "good, thank you for asking",
	})
}

func (a *HttpApi) update(c *gin.Context) {
	a.lock.Lock()
	defer a.lock.Unlock()

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}

	err = a.configLifeCycle.WaveSync(body)
	if _err, ok := err.(*ipruler.EmptyConfig); ok {
		log.Println(_err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": _err.Error()})
		return
	}
	// configLifeCycle.PersistState()
	a.data = body
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (a *HttpApi) backgroundSync(configReloadDuration uint) {
	var oldData []byte

	clc := ipruler.CreateConfigLifeCycle()

	for {
		a.lock.Lock()
		if !reflect.DeepEqual(a.data, oldData) {
			log.Println("detected changes in config")
			oldData = a.data
		}
		clc.WaveSync(a.data)
		a.lock.Unlock()
		time.Sleep(time.Duration(configReloadDuration) * time.Second)
	}
}

func SetupHttpApiMode(configReloadDuration uint, port string, bind_address string) {
	api := HttpApi{
		configLifeCycle: ipruler.CreateConfigLifeCycle(),
	}
	app := gin.Default()
	api.setupRoutes(app)
	go api.backgroundSync(configReloadDuration)
	app.Run(fmt.Sprintf("%s:%s", bind_address, port))
}
