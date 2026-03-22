package provider

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/samber/do/v2"
	"golang.org/x/oauth2"
)

// AuthHandler handles OIDC Authorization Code Flow endpoints.
type AuthHandler struct {
	oauth2Config *oauth2.Config
	frontendURL  string
}

// ProvideAuthHandler creates an AuthHandler with OAuth2 configuration.
func ProvideAuthHandler(i do.Injector) (*AuthHandler, error) {
	cfg := do.MustInvoke[Config](i)

	oauth2Config := &oauth2.Config{
		ClientID: cfg.OIDCClientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.OIDCIssuerExternalURL + "/protocol/openid-connect/auth",
			TokenURL: cfg.OIDCIssuerURL + "/protocol/openid-connect/token",
		},
		RedirectURL: cfg.AuthCallbackURL,
		Scopes:      []string{"openid", "profile"},
	}

	return &AuthHandler{oauth2Config: oauth2Config, frontendURL: cfg.FrontendURL}, nil
}

// HandleLogin redirects the user to the Keycloak authorization endpoint.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}
	authURL := h.oauth2Config.AuthCodeURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleCallback handles the authorization code callback from Keycloak.
// It exchanges the code for a token and redirects to the frontend with the access token.
func (h *AuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	token, err := h.oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("token exchange failed: %v", err)
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}

	redirectURL := fmt.Sprintf("%s/login/callback?access_token=%s",
		h.frontendURL,
		url.QueryEscape(token.AccessToken),
	)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
