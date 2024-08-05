package main

import (
	"fmt"
	"log"
	"log/slog"

	env "github.com/Netflix/go-env"
	"github.com/plutocholia/ipruler/internal/api"
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
	APIBindAddress       string `env:"API_BIND_ADDRESS,default=0.0.0.0"`
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

func main() {
	log.Println(envirnment.String())

	switch envirnment.Mode {
	case "api":
		api.SetupHttpApiMode(envirnment.ConfigReloadDuration, envirnment.APIPort, envirnment.APIBindAddress)
	case "ConfigBased":
		api.SetupConfigfileBasedMode(envirnment.ConfigPath, envirnment.EnablePersistence, envirnment.ConfigReloadDuration)
	default:
		log.Fatalf("mode %s is not defined", envirnment.Mode)
	}
}

func init() {
	if _, err := env.UnmarshalFromEnviron(&envirnment); err != nil {
		log.Fatal(err)
	}
}
