package main

import (
	"log"
	"net/http"
	"os"
	"syscall"

	"github.com/boykush/famoney/server/bff/cmd/server/provider"
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

	do.MustInvoke[*http.Server](injector)

	_, report := injector.ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)
	if report != nil && !report.Succeed {
		log.Fatalf("shutdown error: %v", report)
	}
}
