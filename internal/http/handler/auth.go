package handler

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"example.com/laneblog/internal/config"
	"example.com/laneblog/internal/http/response"
)

type AuthHandler struct {
	cfg config.Config
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAuthHandler(cfg config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "username and password are required")
		return
	}

	if req.Username != h.cfg.Auth.Username || hashPassword(req.Password) != h.cfg.Auth.PasswordHash {
		response.Error(w, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.Auth.CookieName,
		Value:    h.cfg.Auth.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	response.JSON(w, http.StatusOK, map[string]any{
		"message": "login success",
		"token":   h.cfg.Auth.Token,
		"user": map[string]string{
			"username": h.cfg.Auth.Username,
		},
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{
		"username": h.cfg.Auth.Username,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.Auth.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	response.JSON(w, http.StatusOK, map[string]any{
		"message": "logout success",
	})
}

func hashPassword(raw string) string {
	md5Sum := md5.Sum([]byte(raw))
	md5Hex := hex.EncodeToString(md5Sum[:])

	sha1Sum := sha1.Sum([]byte(md5Hex))
	return hex.EncodeToString(sha1Sum[:])
}
