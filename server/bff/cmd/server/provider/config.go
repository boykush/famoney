package provider

// Config holds the BFF service configuration.
type Config struct {
	HTTPPort              string `env:"BFF_HTTP_PORT" envDefault:"8080"`
	ExpenseServiceAddr    string `env:"EXPENSE_SERVICE_ADDR" envDefault:"expense-service:50052"`
	FamilyServiceAddr     string `env:"FAMILY_SERVICE_ADDR" envDefault:"family-service:50053"`
	OIDCIssuerURL         string `env:"OIDC_ISSUER_URL" envDefault:"http://keycloak/realms/famoney"`
	OIDCIssuerExternalURL string `env:"OIDC_ISSUER_EXTERNAL_URL" envDefault:"http://localhost:8180/realms/famoney"`
	OIDCClientID          string `env:"OIDC_CLIENT_ID" envDefault:"famoney-bff"`
	AuthCallbackURL       string `env:"AUTH_CALLBACK_URL" envDefault:"http://localhost:8080/api/v1/auth/callback"`
	FrontendURL           string `env:"FRONTEND_URL" envDefault:"http://localhost:3000"`
}
