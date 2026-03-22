package usecase

import (
	"context"

	"github.com/boykush/famoney/server/family/internal/domain"
	"github.com/boykush/famoney/server/family/internal/infra/ent"
)

// MemberUsecase handles member business logic.
type MemberUsecase struct {
	repo domain.MemberRepository
}

// NewMemberUsecase creates a new MemberUsecase.
func NewMemberUsecase(repo domain.MemberRepository) *MemberUsecase {
	return &MemberUsecase{repo: repo}
}

// GetByOIDCSubject retrieves a member by OIDC subject.
func (u *MemberUsecase) GetByOIDCSubject(ctx context.Context, oidcSubject string) (*domain.Member, error) {
	return u.repo.GetByOIDCSubject(ctx, oidcSubject)
}

// Create creates a new member.
func (u *MemberUsecase) Create(ctx context.Context, oidcSubject, displayName string) (*domain.Member, error) {
	return u.repo.Create(ctx, oidcSubject, displayName)
}

// GetOrCreate retrieves a member by OIDC subject, creating one if not found.
func (u *MemberUsecase) GetOrCreate(ctx context.Context, oidcSubject, displayName string) (*domain.Member, error) {
	m, err := u.repo.GetByOIDCSubject(ctx, oidcSubject)
	if err == nil {
		return m, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	return u.repo.Create(ctx, oidcSubject, displayName)
}
