package provider

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/samber/do/v2"
)

// OIDCVerifier wraps an OIDC ID token verifier.
type OIDCVerifier struct {
	verifier *oidc.IDTokenVerifier
}

// ProvideOIDCVerifier creates an OIDC verifier using the Keycloak issuer.
func ProvideOIDCVerifier(i do.Injector) (*OIDCVerifier, error) {
	cfg := do.MustInvoke[Config](i)

	provider, err := oidc.NewProvider(context.Background(), cfg.OIDCIssuerURL)
	if err != nil {
		log.Printf("OIDC provider not available: %v", err)
		return &OIDCVerifier{}, nil
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.OIDCClientID,
	})

	return &OIDCVerifier{verifier: verifier}, nil
}

// Middleware returns an HTTP middleware that validates OIDC tokens.
// Health check endpoints are excluded from authentication.
func (v *OIDCVerifier) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/health") {
			next.ServeHTTP(w, r)
			return
		}

		if v.verifier == nil {
			http.Error(w, "authentication service unavailable", http.StatusServiceUnavailable)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token == auth {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		_, err := v.verifier.Verify(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
