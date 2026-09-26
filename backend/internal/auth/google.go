package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/boopsrc/boop-flag/backend/internal/domain/user"
)

// GoogleProvider encapsula a interação OAuth2 com o Google: montar a URL de
// consentimento, trocar o código por token e buscar o perfil do usuário.
type GoogleProvider struct {
	cfg        *oauth2.Config
	httpClient *http.Client
}

const googleUserInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"

func NewGoogleProvider(clientID, clientSecret, redirectURL string) *GoogleProvider {
	return &GoogleProvider{
		cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     google.Endpoint,
			Scopes:       []string{"openid", "email", "profile"},
		},
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// AuthCodeURL devolve a URL de consentimento do Google para o state informado.
func (p *GoogleProvider) AuthCodeURL(state string) string {
	return p.cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "select_account"))
}

// Exchange troca o código de autorização pelo perfil do usuário no Google.
func (p *GoogleProvider) Exchange(ctx context.Context, code string) (user.GoogleProfile, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	token, err := p.cfg.Exchange(ctx, code)
	if err != nil {
		return user.GoogleProfile{}, fmt.Errorf("exchange code: %w", err)
	}

	client := p.cfg.Client(ctx, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return user.GoogleProfile{}, fmt.Errorf("build userinfo request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return user.GoogleProfile{}, fmt.Errorf("fetch userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return user.GoogleProfile{}, fmt.Errorf("userinfo returned status %d", resp.StatusCode)
	}

	var info struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return user.GoogleProfile{}, fmt.Errorf("decode userinfo: %w", err)
	}

	return user.GoogleProfile{
		Sub:           info.Sub,
		Email:         info.Email,
		EmailVerified: info.EmailVerified,
		Name:          info.Name,
		Picture:       info.Picture,
	}, nil
}
