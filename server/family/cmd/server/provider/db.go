package provider

import (
	"github.com/boykush/famoney/server/family/internal/infra/ent"
	"github.com/samber/do/v2"

	_ "github.com/lib/pq"
)

// ProvideEntClient creates an Ent client connected to the database.
func ProvideEntClient(i do.Injector) (*ent.Client, error) {
	cfg := do.MustInvoke[Config](i)
	return ent.Open("postgres", cfg.DatabaseURL)
}
