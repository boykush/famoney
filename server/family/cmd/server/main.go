package main

import (
	"log"
	"os"
	"syscall"

	"github.com/boykush/famoney/server/family/cmd/server/provider"
	"github.com/caarlos0/env/v11"
	"github.com/samber/do/v2"
)

func main() {
	cfg, err := env.ParseAs[provider.Config]()
	if err != nil {
		log.Fatalf("failed to parse environment variables: %v", err)
	}

	injector := do.New()
	do.ProvideValue(injector, cfg)
	provider.Register(injector)

	do.MustInvoke[*provider.GRPCServer](injector)

	_, report := injector.ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)
	if report != nil && !report.Succeed {
		log.Fatalf("shutdown error: %v", report)
	}
}
