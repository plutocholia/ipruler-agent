package api

import (
	"log"
	"os"
	"reflect"
	"time"

	"github.com/plutocholia/ipruler/internal/ipruler"
)

func SetupConfigfileBasedMode(configPath string, enablePersistence bool, configReloadDuration uint) {
	var oldData []byte

	configLifeCycle := ipruler.CreateConfigLifeCycle()

	for {
		data, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if !reflect.DeepEqual(data, oldData) {
			log.Println("detected changes in config")
			oldData = data
		}

		configLifeCycle.WaveSync(data)
		if enablePersistence {
			configLifeCycle.PersistState()
		}

		time.Sleep(time.Duration(configReloadDuration) * time.Second)
	}
}
