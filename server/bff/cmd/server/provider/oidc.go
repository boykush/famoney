package provider

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/samber/do/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	familypb "github.com/boykush/famoney/server/bff/gen/go/family"
)

// OIDCVerifier wraps an OIDC ID token verifier.
type OIDCVerifier struct {
	verifier     *oidc.IDTokenVerifier
	familyClient familypb.FamilyServiceClient
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

	conn, err := grpc.NewClient(cfg.FamilyServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	familyClient := familypb.NewFamilyServiceClient(conn)

	return &OIDCVerifier{verifier: verifier, familyClient: familyClient}, nil
}

// Middleware returns an HTTP middleware that validates OIDC tokens.
// Health check endpoints are excluded from authentication.
// On successful authentication, it ensures a family member exists for the OIDC subject.
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

		idToken, err := v.verifier.Verify(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Ensure family member exists for the authenticated user
		if v.familyClient != nil {
			v.ensureMember(r.Context(), idToken)
		}

		next.ServeHTTP(w, r)
	})
}

// ensureMember ensures a family member record exists for the OIDC subject.
// It attempts to get the member first, and creates one if not found.
func (v *OIDCVerifier) ensureMember(ctx context.Context, idToken *oidc.IDToken) {
	subject := idToken.Subject

	_, err := v.familyClient.GetMemberBySubject(ctx, &familypb.GetMemberBySubjectRequest{
		OidcSubject: subject,
	})
	if err == nil {
		return
	}

	if status.Code(err) != codes.NotFound {
		log.Printf("failed to get member by subject: %v", err)
		return
	}

	// Extract display name from token claims
	var claims struct {
		Name              string `json:"name"`
		PreferredUsername  string `json:"preferred_username"`
	}
	if err := idToken.Claims(&claims); err != nil {
		log.Printf("failed to parse token claims: %v", err)
		return
	}

	displayName := claims.Name
	if displayName == "" {
		displayName = claims.PreferredUsername
	}
	if displayName == "" {
		displayName = subject
	}

	_, err = v.familyClient.CreateMember(ctx, &familypb.CreateMemberRequest{
		OidcSubject: subject,
		DisplayName: displayName,
	})
	if err != nil {
		log.Printf("failed to create member: %v", err)
	}
}
