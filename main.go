package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	env "github.com/Netflix/go-env"
	"github.com/gin-gonic/gin"
	"github.com/plutocholia/ipruler/cmd/api"
	"github.com/plutocholia/ipruler/cmd/ipruler"
)

var (
	envirnment Environment
)

type Environment struct {
	Mode                 string `env:"MODE,default=api"`
	EnablePersistence    bool   `env:"ENABLE_PERSISTENCE,default=false"`
	APIPort              string `env:"API_PORT,default=8080"`
	ConfigPath           string `env:"CONFIG_PATH,default=./config/config.yaml"`
	ConfigReloadDuration uint   `env:"CONFIG_RELOAD_DURATION_SECONDS,default=15"`
}

func (e *Environment) String() string {
	return fmt.Sprintf(`
Environments:
	Mode: %s
	EnablePersistence: %t
	APIPort: %s
	ConfigPath: %s
	ConfigReloadDuration: %d
`, e.Mode, e.EnablePersistence, e.APIPort, e.ConfigPath, e.ConfigReloadDuration)
}

func apiMode() {
	app := gin.Default()

	api.SetupRoutes(app)

	app.Run(fmt.Sprintf("0.0.0.0:%s", envirnment.APIPort))
}

func configBasedMode() {
	var oldData []byte

	configLifeCycle := ipruler.CreateConfigLifeCycle()

	for {
		data, err := os.ReadFile(envirnment.ConfigPath)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if !reflect.DeepEqual(data, oldData) {
			log.Println("detected changes in config")
			oldData = data
			configLifeCycle.Update(data)
			ipruler.SyncState(configLifeCycle)
			if envirnment.EnablePersistence {
				ipruler.PersistState(configLifeCycle)
			}
		} else {
			configLifeCycle.Update(data)
			ipruler.SyncState(configLifeCycle)
		}

		time.Sleep(time.Duration(envirnment.ConfigReloadDuration) * time.Second)
	}
}

func main() {
	log.Println(envirnment.String())
	if envirnment.Mode == "api" {
		apiMode()
	} else if envirnment.Mode == "ConfigBased" {
		configBasedMode()
	} else {
		log.Fatalf("mode %s is not defined", envirnment.Mode)
	}
}

func init() {
	_, err := env.UnmarshalFromEnviron(&envirnment)
	if err != nil {
		log.Fatal(err)
	}
}
