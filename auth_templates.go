package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
)

func authTemplate(strategy string) string {
	switch strategy {
	case "JWT + Refresh Token":
		return `package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct { secret []byte }

func New() Service {
	secret := os.Getenv("AUTH_SECRET")
	if secret == "" { secret = "change-me-in-production" }
	return Service{secret: []byte(secret)}
}

func (service Service) IssueAccessToken(subject string) (string, error) {
	return service.issue(subject, 15*time.Minute, "access")
}

func (service Service) IssueRefreshToken(subject string) (string, error) {
	return service.issue(subject, 7*24*time.Hour, "refresh")
}

func (service Service) issue(subject string, lifetime time.Duration, tokenType string) (string, error) {
	claims := jwt.MapClaims{"sub": subject, "type": tokenType, "exp": time.Now().Add(lifetime).Unix()}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(service.secret)
}

func (service Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		header := request.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") { http.Error(writer, "missing bearer token", http.StatusUnauthorized); return }
		if _, err := jwt.Parse(strings.TrimPrefix(header, "Bearer "), func(token *jwt.Token) (any, error) { return service.secret, nil }); err != nil {
			http.Error(writer, "invalid token", http.StatusUnauthorized); return
		}
		next.ServeHTTP(writer, request)
	})
}

func (service Service) LoginHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost { http.Error(writer, "method not allowed", http.StatusMethodNotAllowed); return }
	subject := request.FormValue("subject")
	if subject == "" { http.Error(writer, "subject is required", http.StatusBadRequest); return }
	access, err := service.IssueAccessToken(subject); if err != nil { http.Error(writer, "token issue failed", http.StatusInternalServerError); return }
	refresh, err := service.IssueRefreshToken(subject); if err != nil { http.Error(writer, "token issue failed", http.StatusInternalServerError); return }
	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "{\"access_token\":%q,\"refresh_token\":%q}", access, refresh)
}
`
	case "OAuth2 (Google / GitHub via Goth)":
		return `package auth

import (
	"net/http"
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

func ConfigureOAuth() {
	if os.Getenv("GOOGLE_KEY") != "" { goth.UseProviders(google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), os.Getenv("OAUTH_CALLBACK_URL"), "email", "profile")) }
	if os.Getenv("GITHUB_KEY") != "" { goth.UseProviders(github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), os.Getenv("OAUTH_CALLBACK_URL"), "user:email")) }
}

func BeginOAuth(writer http.ResponseWriter, request *http.Request) { gothic.BeginAuthHandler(writer, request) }
func CompleteOAuth(writer http.ResponseWriter, request *http.Request) { user, err := gothic.CompleteUserAuth(writer, request); if err != nil { http.Error(writer, err.Error(), http.StatusUnauthorized); return }; http.Redirect(writer, request, "/?user="+user.Email, http.StatusFound) }
`
	default:
		return `package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
)

type Session struct { secret []byte }

func New() Session { secret := os.Getenv("SESSION_SECRET"); if secret == "" { secret = "change-me-in-production" }; return Session{[]byte(secret)} }

func (session Session) Set(writer http.ResponseWriter, subject string) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(subject))
	mac := hmac.New(sha256.New, session.secret); mac.Write([]byte(payload))
	value := payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	http.SetCookie(writer, &http.Cookie{Name: "session", Value: value, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
}

func (session Session) Subject(request *http.Request) (string, bool) {
	cookie, err := request.Cookie("session"); if err != nil { return "", false }
	parts := strings.Split(cookie.Value, "."); if len(parts) != 2 { return "", false }
	mac := hmac.New(sha256.New, session.secret); mac.Write([]byte(parts[0]))
	signature, err := base64.RawURLEncoding.DecodeString(parts[1]); if err != nil || !hmac.Equal(signature, mac.Sum(nil)) { return "", false }
	payload, err := base64.RawURLEncoding.DecodeString(parts[0]); return string(payload), err == nil
}

func (session Session) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, ok := session.Subject(request); !ok { http.Error(writer, "unauthorized", http.StatusUnauthorized); return }
		next.ServeHTTP(writer, request)
	})
}

func (session Session) LoginHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost { http.Error(writer, "method not allowed", http.StatusMethodNotAllowed); return }
	subject := request.FormValue("subject")
	if subject == "" { http.Error(writer, "subject is required", http.StatusBadRequest); return }
	session.Set(writer, subject)
	writer.WriteHeader(http.StatusNoContent)
}
`
	}
}

func pageTemplate(strategy string) string {
	switch strategy {
	case "Go Templ + HTMX (native Go server-side rendering)":
		return `<!doctype html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><script src="https://unpkg.com/htmx.org@2.0.4"></script><link rel="stylesheet" href="/static/css/app.css"></head><body>
<main><h1>{{.ProjectName}}</h1><button hx-get="/health" hx-target="#status">Check server</button><p id="status">Ready.</p></main>
</body></html>
`
	case "Go + Inertia.js (Go monolith with React/Vue)":
		return `<div id="app"></div>
<script type="module" src="/frontend/main.jsx"></script>
`
	default:
		return `<!doctype html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script><link rel="stylesheet" href="/static/css/app.css"></head><body>
<main x-data="{ open: false }"><h1>{{.ProjectName}}</h1><button @click="open = !open">Toggle</button><span x-show="open">Ready.</span></main>
</body></html>
`
	}
}

func render(source string, options Options) (string, error) {
	tmpl, err := template.New("file").Parse(source)
	if err != nil {
		return "", err
	}
	var output strings.Builder
	if err := tmpl.Execute(&output, options); err != nil {
		return "", err
	}
	return output.String(), nil
}

func validProjectName(value string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(value)
}

func fatal(message string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", message)
	os.Exit(1)
}
