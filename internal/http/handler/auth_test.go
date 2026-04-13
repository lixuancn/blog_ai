package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/laneblog/internal/config"
)

func TestAuthHandlerLogin(t *testing.T) {
	cfg := config.Config{
		Auth: config.AuthConfig{
			Username:     "lane",
			PasswordHash: hashPassword("correct-password"),
			Token:        "test-token",
			CookieName:   "admin_token",
		},
	}

	handler := NewAuthHandler(cfg)

	t.Run("success", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"username": "lane",
			"password": "correct-password",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handler.Login(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		if len(rec.Result().Cookies()) == 0 {
			t.Fatal("expected auth cookie to be set")
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"username": "lane",
			"password": "wrong",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handler.Login(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}
