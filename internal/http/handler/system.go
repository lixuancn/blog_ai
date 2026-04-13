package handler

import (
	"net/http"
	"time"

	"example.com/laneblog/internal/config"
	"example.com/laneblog/internal/http/response"
)

type SystemHandler struct {
	cfg config.Config
}

func NewSystemHandler(cfg config.Config) *SystemHandler {
	return &SystemHandler{cfg: cfg}
}

func (h *SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": h.cfg.App.Name,
		"env":     h.cfg.App.Env,
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (h *SystemHandler) Ready(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{
		"status": "ready",
	})
}
