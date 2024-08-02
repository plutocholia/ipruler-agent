package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"reflect"
	"time"

	env "github.com/Netflix/go-env"
	"github.com/gin-gonic/gin"
	"github.com/plutocholia/ipruler/internal/api"
	"github.com/plutocholia/ipruler/internal/ipruler"
)

var (
	envirnment  Environment
	LogLevelMap map[string]int = map[string]int{
		"INFO":  int(slog.LevelInfo),
		"DEBUG": int(slog.LevelDebug),
		"WARN":  int(slog.LevelWarn),
		"ERROR": int(slog.LevelError),
	}
)

type Environment struct {
	Mode                 string `env:"MODE,default=api"`
	EnablePersistence    bool   `env:"ENABLE_PERSISTENCE,default=false"`
	APIPort              string `env:"API_PORT,default=8080"`
	ConfigPath           string `env:"CONFIG_PATH,default=./config/config.yaml"`
	ConfigReloadDuration uint   `env:"CONFIG_RELOAD_DURATION_SECONDS,default=15"`
	LogLevel             string `env:"LOG_LEVEL,default=INFO"`
}

func (e *Environment) String() string {
	return fmt.Sprintf(`
Environments:
	Mode: %s
	EnablePersistence: %t
	APIPort: %s
	ConfigPath: %s
	ConfigReloadDuration: %d
	LogLevel: %s
`, e.Mode, e.EnablePersistence, e.APIPort, e.ConfigPath, e.ConfigReloadDuration, e.LogLevel)
}

func apiMode() {
	app := gin.Default()
	api.SetupRoutes(app)
	go api.BackgroundSync(envirnment.ConfigReloadDuration)
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
		}

		configLifeCycle.WaveSync(data)
		if envirnment.EnablePersistence {
			configLifeCycle.PersistState()
		}

		time.Sleep(time.Duration(envirnment.ConfigReloadDuration) * time.Second)
	}
}

func main() {
	log.Println(envirnment.String())

	// // setup slog
	// if value, exists := LogLevelMap[envirnment.LogLevel]; exists {
	// 	logger := slog.New(slog.NewTextHandler(os.Stderr,
	// 		&slog.HandlerOptions{Level: slog.Level(value)}))
	// 	slog.SetDefault(logger)
	// } else {
	// 	log.Fatalf("loglevel %s is not valid", envirnment.LogLevel)
	// }

	switch envirnment.Mode {
	case "api":
		apiMode()
	case "ConfigBased":
		configBasedMode()
	default:
		log.Fatalf("mode %s is not defined", envirnment.Mode)
	}
}

func init() {
	if _, err := env.UnmarshalFromEnviron(&envirnment); err != nil {
		log.Fatal(err)
	}
}
