package domain

import "context"

// MemberRepository defines the interface for member persistence.
type MemberRepository interface {
	GetByOIDCSubject(ctx context.Context, oidcSubject string) (*Member, error)
	Create(ctx context.Context, oidcSubject, displayName string) (*Member, error)
}
