package provider

// Config holds the family service configuration.
type Config struct {
	Port        string `env:"FAMILY_SERVICE_PORT" envDefault:"50053"`
	DatabaseURL string `env:"DATABASE_URL,required"`
}
