package repository

import (
	"context"

	"github.com/boykush/famoney/server/family/internal/domain"
	"github.com/boykush/famoney/server/family/internal/infra/ent"
	"github.com/boykush/famoney/server/family/internal/infra/ent/member"
)

// MemberRepository implements domain.MemberRepository using Ent.
type MemberRepository struct {
	client *ent.Client
}

// NewMemberRepository creates a new MemberRepository.
func NewMemberRepository(client *ent.Client) *MemberRepository {
	return &MemberRepository{client: client}
}

func (r *MemberRepository) GetByOIDCSubject(ctx context.Context, oidcSubject string) (*domain.Member, error) {
	m, err := r.client.Member.Query().
		Where(member.OidcSubject(oidcSubject)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainMember(m), nil
}

func (r *MemberRepository) Create(ctx context.Context, oidcSubject, displayName string) (*domain.Member, error) {
	m, err := r.client.Member.Create().
		SetOidcSubject(oidcSubject).
		SetDisplayName(displayName).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainMember(m), nil
}

func toDomainMember(m *ent.Member) *domain.Member {
	return &domain.Member{
		ID:          m.ID,
		OIDCSubject: m.OidcSubject,
		DisplayName: m.DisplayName,
		CreatedAt:   m.CreatedAt,
	}
}
