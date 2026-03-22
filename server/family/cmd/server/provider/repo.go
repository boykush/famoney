package provider

import (
	"github.com/boykush/famoney/server/family/internal/domain"
	"github.com/boykush/famoney/server/family/internal/infra/ent"
	"github.com/boykush/famoney/server/family/internal/infra/repository"
	"github.com/samber/do/v2"
)

// ProvideMemberRepository creates a MemberRepository backed by Ent.
func ProvideMemberRepository(i do.Injector) (domain.MemberRepository, error) {
	client := do.MustInvoke[*ent.Client](i)
	return repository.NewMemberRepository(client), nil
}
