package provider

import "github.com/samber/do/v2"

// Register registers all BFF service providers with the injector.
func Register(injector do.Injector) {
	do.Provide(injector, ProvideOIDCVerifier)
	do.Provide(injector, ProvideHTTPServer)
}
