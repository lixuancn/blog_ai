package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"example.com/laneblog/internal/admin"
	"example.com/laneblog/internal/http/response"
)

type AdminHandler struct {
	service *admin.Service
}

func NewAdminHandler(service *admin.Service) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.Dashboard())
}

func (h *AdminHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	items := h.service.ListCategories(r.URL.Query().Get("name"))
	response.JSON(w, http.StatusOK, map[string]any{
		"items": items,
		"total": len(items),
	})
}

func (h *AdminHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	item, err := h.service.GetCategory(id)
	if err != nil {
		writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *AdminHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var input admin.CategoryInput
	if !decodeJSON(w, r, &input) {
		return
	}

	item, err := h.service.CreateCategory(input)
	if err != nil {
		writeError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, item)
}

func (h *AdminHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	var input admin.CategoryInput
	if !decodeJSON(w, r, &input) {
		return
	}

	item, err := h.service.UpdateCategory(id, input)
	if err != nil {
		writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *AdminHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteCategory(id); err != nil {
		writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"message": "category deleted",
	})
}

func (h *AdminHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	query := admin.ArticleQuery{
		Title:         r.URL.Query().Get("title"),
		CategoryID:    parseQueryInt64(r, "mid"),
		RecommendType: int(parseQueryInt64(r, "recommend_type")),
		Page:          int(parseQueryInt64(r, "page")),
		PageSize:      int(parseQueryInt64(r, "page_size")),
	}

	response.JSON(w, http.StatusOK, h.service.ListArticles(query))
}

func (h *AdminHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	item, err := h.service.GetArticle(id)
	if err != nil {
		writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *AdminHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	var input admin.ArticleInput
	if !decodeJSON(w, r, &input) {
		return
	}

	item, err := h.service.CreateArticle(input)
	if err != nil {
		writeError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, item)
}

func (h *AdminHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	var input admin.ArticleInput
	if !decodeJSON(w, r, &input) {
		return
	}

	item, err := h.service.UpdateArticle(id, input)
	if err != nil {
		writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *AdminHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteArticle(id); err != nil {
		writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"message": "article deleted",
	})
}

func (h *AdminHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	items := h.service.ListTags(r.URL.Query().Get("name"))
	response.JSON(w, http.StatusOK, map[string]any{
		"items": items,
		"total": len(items),
	})
}

func (h *AdminHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteTag(id); err != nil {
		writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"message": "tag deleted",
	})
}

func (h *AdminHandler) RebuildTags(w http.ResponseWriter, r *http.Request) {
	items := h.service.RebuildTags()
	response.JSON(w, http.StatusOK, map[string]any{
		"message": "tags rebuilt",
		"items":   items,
		"total":   len(items),
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func parseID(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	value := r.PathValue(key)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func parseQueryInt64(r *http.Request, key string) int64 {
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

func writeError(w http.ResponseWriter, err error) {
	var appErr *admin.AppError
	if ok := errorAs(err, &appErr); ok {
		payload := map[string]any{
			"error": appErr.Message,
		}
		if len(appErr.Fields) > 0 {
			payload["fields"] = appErr.Fields
		}
		response.JSON(w, appErr.Status, payload)
		return
	}

	response.Error(w, http.StatusInternalServerError, "internal server error")
}

func errorAs(err error, target **admin.AppError) bool {
	appErr, ok := err.(*admin.AppError)
	if !ok {
		return false
	}
	*target = appErr
	return true
}
