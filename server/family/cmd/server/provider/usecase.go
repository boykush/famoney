package provider

import (
	"github.com/boykush/famoney/server/family/internal/domain"
	"github.com/boykush/famoney/server/family/internal/usecase"
	"github.com/samber/do/v2"
)

// ProvideMemberUsecase creates a MemberUsecase.
func ProvideMemberUsecase(i do.Injector) (*usecase.MemberUsecase, error) {
	repo := do.MustInvoke[domain.MemberRepository](i)
	return usecase.NewMemberUsecase(repo), nil
}
