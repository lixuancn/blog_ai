package middleware

import (
	"net/http"
	"strings"

	"example.com/laneblog/internal/config"
	"example.com/laneblog/internal/http/response"
)

func RequireAuth(cfg config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !Authorized(cfg, r) {
			response.Error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Authorized(cfg config.Config, r *http.Request) bool {
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		if cookie, err := r.Cookie(cfg.Auth.CookieName); err == nil {
			token = cookie.Value
		}
	}
	return token != "" && token == cfg.Auth.Token
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
