package provider

import "github.com/samber/do/v2"

// Register registers all family service providers with the injector.
func Register(injector do.Injector) {
	do.Provide(injector, ProvideEntClient)
	do.Provide(injector, ProvideMemberRepository)
	do.Provide(injector, ProvideMemberUsecase)
	do.Provide(injector, ProvideServer)
	do.Provide(injector, ProvideGRPCServer)
}
