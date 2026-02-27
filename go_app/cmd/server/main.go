package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type oidcConfig struct {
	Provider *oidc.Provider
	Verifier *oidc.IDTokenVerifier
	OAuth2   *oauth2.Config
}

type appConfig struct {
	Port               string
	Issuer             string
	DiscoveryInternal  string // если задан — запросы к Issuer перенаправляются сюда (для Docker)
	ClientID           string
	ClientSecret       string
	RedirectURL        string
	SessionSecret      string
}

func loadConfig() appConfig {
	return appConfig{
		Port:              getenv("APP_PORT", "8212"),
		Issuer:            getenv("OIDC_ISSUER", ""),
		DiscoveryInternal: getenv("OIDC_DISCOVERY_INTERNAL", ""),
		ClientID:          getenv("OIDC_CLIENT_ID", ""),
		ClientSecret:      getenv("OIDC_CLIENT_SECRET", ""),
		RedirectURL:       getenv("OIDC_REDIRECT_URL", "http://localhost:8212/callback"),
		SessionSecret:     getenv("SESSION_SECRET", "change_me_session"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// issuerRewriteTransport перенаправляет запросы с хоста issuer на внутренний хост Keycloak (для Docker).
type issuerRewriteTransport struct {
	base         http.RoundTripper
	issuerHost   string
	internalHost string
}

func (t *issuerRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.internalHost != "" && req.URL != nil && req.URL.Host == t.issuerHost {
		req = req.Clone(req.Context())
		req.URL.Host = t.internalHost
		req.Host = t.internalHost
	}
	return t.base.RoundTrip(req)
}

func initOIDC(ctx context.Context, cfg appConfig) (*oidcConfig, error) {
	if cfg.DiscoveryInternal != "" {
		issuerURL, err := url.Parse(cfg.Issuer)
		if err != nil {
			return nil, err
		}
		internalURL, err := url.Parse(cfg.DiscoveryInternal)
		if err != nil {
			return nil, err
		}
		client := &http.Client{
			Transport: &issuerRewriteTransport{
				base:         http.DefaultTransport,
				issuerHost:   issuerURL.Host,
				internalHost: internalURL.Host,
			},
		}
		ctx = oidc.ClientContext(ctx, client)
	}
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, err
	}

	oidcCfg := &oidcConfig{
		Provider: provider,
		OAuth2: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
	}

	oidcCfg.Verifier = provider.Verifier(&oidc.Config{
		ClientID: cfg.ClientID,
	})

	return oidcCfg, nil
}

func initOIDCWithRetry(ctx context.Context, cfg appConfig, maxAttempts int, interval time.Duration) (*oidcConfig, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		oidcCfg, err := initOIDC(ctx, cfg)
		if err == nil {
			return oidcCfg, nil
		}
		lastErr = err
		if attempt < maxAttempts {
			log.Printf("OIDC init attempt %d/%d failed: %v; retrying in %v", attempt, maxAttempts, err, interval)
			time.Sleep(interval)
		}
	}
	return nil, lastErr
}

func main() {
	cfg := loadConfig()
	ctx := context.Background()

	if cfg.Issuer == "" {
		log.Fatal("OIDC_ISSUER is required")
	}

	oidcCfg, err := initOIDCWithRetry(ctx, cfg, 15, 3*time.Second)
	if err != nil {
		log.Fatalf("failed to init OIDC after retries: %v", err)
	}

	r := gin.Default()

	store := cookie.NewStore([]byte(cfg.SessionSecret))
	r.Use(sessions.Sessions("go_app_session", store))

	r.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		username, ok := session.Get("username").(string)
		if !ok || username == "" {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, "Привет, %s, это Go сервис", username)
	})

	r.GET("/login", func(c *gin.Context) {
		session := sessions.Default(c)
		state := "state" // для POC можно захардкодить
		session.Set("state", state)
		if err := session.Save(); err != nil {
			c.String(http.StatusInternalServerError, "failed to save session")
			return
		}

		url := oidcCfg.OAuth2.AuthCodeURL(state)
		c.Redirect(http.StatusFound, url)
	})

	r.GET("/callback", func(c *gin.Context) {
		session := sessions.Default(c)
		stateExpected, _ := session.Get("state").(string)
		state := c.Query("state")
		if state == "" || state != stateExpected {
			c.String(http.StatusBadRequest, "invalid state")
			return
		}

		code := c.Query("code")
		if code == "" {
			c.String(http.StatusBadRequest, "missing code")
			return
		}

		oauth2Token, err := oidcCfg.OAuth2.Exchange(ctx, code)
		if err != nil {
			c.String(http.StatusInternalServerError, "token exchange failed: %v", err)
			return
		}

		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			c.String(http.StatusInternalServerError, "no id_token in token response")
			return
		}

		idToken, err := oidcCfg.Verifier.Verify(ctx, rawIDToken)
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to verify id_token: %v", err)
			return
		}

		var claims struct {
			PreferredUsername string `json:"preferred_username"`
			Email             string `json:"email"`
		}
		if err := idToken.Claims(&claims); err != nil {
			c.String(http.StatusInternalServerError, "failed to parse claims: %v", err)
			return
		}

		username := claims.PreferredUsername
		if username == "" {
			username = claims.Email
		}
		if username == "" {
			username = "user"
		}

		session.Delete("state")
		session.Set("username", username)
		if err := session.Save(); err != nil {
			c.String(http.StatusInternalServerError, "failed to save session: %v", err)
			return
		}

		c.Redirect(http.StatusFound, "/")
	})

	addr := ":" + cfg.Port
	log.Printf("Starting Go app on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

