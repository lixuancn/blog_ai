package handler

import (
	"embed"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"example.com/laneblog/internal/frontend"
	"example.com/laneblog/internal/http/response"
)

//go:embed fronttemplates/*.tmpl
var frontTemplateFS embed.FS

type FrontHandler struct {
	service *frontend.Service
	tmpl    *template.Template
}

func NewFrontHandler(service *frontend.Service) *FrontHandler {
	funcs := template.FuncMap{
		"formatUnix": formatUnix,
		"safeHTML": func(value string) template.HTML {
			return template.HTML(value)
		},
		"join": strings.Join,
	}
	tmpl := template.Must(template.New("front").Funcs(funcs).ParseFS(frontTemplateFS, "fronttemplates/*.tmpl"))
	return &FrontHandler{service: service, tmpl: tmpl}
}

func (h *FrontHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	h.render(w, "home_page", h.service.HomePageData(parseFrontQueryInt(r, "page", 1)))
}

func (h *FrontHandler) CategoryPage(w http.ResponseWriter, r *http.Request) {
	id, ok := parseFrontID(w, r, "id")
	if !ok {
		return
	}
	data, err := h.service.CategoryPageData(id)
	if err != nil {
		writeFrontHTMLError(w, err)
		return
	}
	h.render(w, "category_page", data)
}

func (h *FrontHandler) ArticlePage(w http.ResponseWriter, r *http.Request) {
	id, ok := parseFrontID(w, r, "id")
	if !ok {
		return
	}
	data, err := h.service.ArticlePageData(id)
	if err != nil {
		writeFrontHTMLError(w, err)
		return
	}
	h.render(w, "article_page", data)
}

func (h *FrontHandler) HomeAPI(w http.ResponseWriter, r *http.Request) {
	categoryID := parseFrontQueryInt64(r, "mid")
	if categoryID > 0 {
		data, err := h.service.CategoryPageData(categoryID)
		if err != nil {
			writeFrontJSONError(w, err)
			return
		}
		response.JSON(w, http.StatusOK, data)
		return
	}
	response.JSON(w, http.StatusOK, h.service.HomePageData(parseFrontQueryInt(r, "page", 1)))
}

func (h *FrontHandler) ArticleAPI(w http.ResponseWriter, r *http.Request) {
	id, ok := parseFrontID(w, r, "id")
	if !ok {
		return
	}
	data, err := h.service.ArticleData(id)
	if err != nil {
		writeFrontJSONError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, data)
}

func (h *FrontHandler) LikeArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := parseFrontID(w, r, "id")
	if !ok {
		return
	}
	data, err := h.service.LikeArticle(id)
	if err != nil {
		writeFrontJSONError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, data)
}

func (h *FrontHandler) DislikeArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := parseFrontID(w, r, "id")
	if !ok {
		return
	}
	data, err := h.service.DislikeArticle(id)
	if err != nil {
		writeFrontJSONError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, data)
}

func (h *FrontHandler) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "render page failed", http.StatusInternalServerError)
	}
}

func parseFrontID(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	value := r.PathValue(key)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func parseFrontQueryInt64(r *http.Request, key string) int64 {
	value := r.URL.Query().Get(key)
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func parseFrontQueryInt(r *http.Request, key string, fallback int) int {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func writeFrontJSONError(w http.ResponseWriter, err error) {
	var appErr *frontend.AppError
	if frontend.ErrorAs(err, &appErr) {
		response.JSON(w, appErr.Status, map[string]any{"error": appErr.Message})
		return
	}
	response.Error(w, http.StatusInternalServerError, "internal server error")
}

func writeFrontHTMLError(w http.ResponseWriter, err error) {
	var appErr *frontend.AppError
	if frontend.ErrorAs(err, &appErr) {
		http.Error(w, appErr.Message, appErr.Status)
		return
	}
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func formatUnix(value int64) string {
	if value <= 0 {
		return "-"
	}
	return time.Unix(value, 0).Format("2006-01-02 15:04")
}
