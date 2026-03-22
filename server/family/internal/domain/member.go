package domain

import "time"

// Member represents a family member linked to an OIDC identity.
type Member struct {
	ID          int
	OIDCSubject string
	DisplayName string
	CreatedAt   time.Time
}
