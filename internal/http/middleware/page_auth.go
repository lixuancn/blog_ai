package middleware

import (
	"net/http"
	"net/url"

	"example.com/laneblog/internal/config"
)

func RequirePageAuth(cfg config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !Authorized(cfg, r) {
			redirectTo := "/admin/login?redirect=" + url.QueryEscape(r.URL.RequestURI())
			http.Redirect(w, r, redirectTo, http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
