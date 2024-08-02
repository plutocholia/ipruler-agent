package ipruler

// Define a new type for your custom error
type EmptyConfig struct {
	Message string
}

func (e *EmptyConfig) Error() string {
	return e.Message
}

func CreateEmptyConfigError() error {
	return &EmptyConfig{
		Message: "The given config is parsed as an empty config. skipped",
	}
}
