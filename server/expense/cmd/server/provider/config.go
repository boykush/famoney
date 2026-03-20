package provider

// Config holds the expense service configuration.
type Config struct {
	Port        string `env:"EXPENSE_SERVICE_PORT" envDefault:"50052"`
	DatabaseURL string `env:"DATABASE_URL,required"`
}
