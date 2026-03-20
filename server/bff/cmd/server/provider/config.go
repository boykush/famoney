package provider

// Config holds the BFF service configuration.
type Config struct {
	HTTPPort           string `env:"BFF_HTTP_PORT" envDefault:"8080"`
	ExpenseServiceAddr string `env:"EXPENSE_SERVICE_ADDR" envDefault:"expense-service:50052"`
}
