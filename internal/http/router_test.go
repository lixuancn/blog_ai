package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"example.com/laneblog/internal/admin"
	"example.com/laneblog/internal/config"
	"example.com/laneblog/internal/frontend"
	"example.com/laneblog/internal/store"
)

func TestAdminRouterCRUDFlow(t *testing.T) {
	repo := store.NewMemoryStore()

	cfg := config.Config{
		Auth: config.AuthConfig{
			Token:      "test-token",
			CookieName: "admin_token",
			Username:   "lane",
		},
	}

	router := NewRouter(cfg, admin.NewService(repo), frontend.NewService(repo))

	categoryBody := mustJSON(t, map[string]any{
		"name":            "Go",
		"seo_title":       "Go",
		"seo_description": "Go 分类",
		"seo_keywords":    "go",
		"in_out":          1,
		"pid":             0,
		"item":            "go",
	})

	categoryRec := performJSONRequest(router, http.MethodPost, "/api/v1/admin/categories", categoryBody, cfg.Auth.Token)
	if categoryRec.Code != http.StatusCreated {
		t.Fatalf("expected create category status %d, got %d, body=%s", http.StatusCreated, categoryRec.Code, categoryRec.Body.String())
	}

	var category struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(categoryRec.Body.Bytes(), &category); err != nil {
		t.Fatalf("decode category: %v", err)
	}

	articleBody := mustJSON(t, map[string]any{
		"mid":             category.ID,
		"author":          "lane",
		"title":           "Go 后台开发实践",
		"description":     "文章摘要",
		"seo_title":       "Go 后台开发实践",
		"seo_description": "SEO 描述",
		"seo_keywords":    "go,admin",
		"tag":             "Go,Backend",
		"content":         "content",
		"recommend_type":  1,
		"clicks":          0,
		"good_num":        0,
		"bad_num":         0,
	})

	articleRec := performJSONRequest(router, http.MethodPost, "/api/v1/admin/articles", articleBody, cfg.Auth.Token)
	if articleRec.Code != http.StatusCreated {
		t.Fatalf("expected create article status %d, got %d, body=%s", http.StatusCreated, articleRec.Code, articleRec.Body.String())
	}

	dashboardRec := performJSONRequest(router, http.MethodGet, "/api/v1/admin/dashboard", nil, cfg.Auth.Token)
	if dashboardRec.Code != http.StatusOK {
		t.Fatalf("expected dashboard status %d, got %d", http.StatusOK, dashboardRec.Code)
	}

	var dashboard struct {
		Summary struct {
			ArticleCount  int `json:"article_count"`
			CategoryCount int `json:"category_count"`
			TagCount      int `json:"tag_count"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(dashboardRec.Body.Bytes(), &dashboard); err != nil {
		t.Fatalf("decode dashboard: %v", err)
	}
	if dashboard.Summary.ArticleCount != 1 || dashboard.Summary.CategoryCount != 1 || dashboard.Summary.TagCount != 2 {
		t.Fatalf("unexpected dashboard summary: %s", dashboardRec.Body.String())
	}

	tagsRec := performJSONRequest(router, http.MethodGet, "/api/v1/admin/tags", nil, cfg.Auth.Token)
	if tagsRec.Code != http.StatusOK {
		t.Fatalf("expected tags status %d, got %d", http.StatusOK, tagsRec.Code)
	}

	deleteCategoryRec := performJSONRequest(router, http.MethodDelete, "/api/v1/admin/categories/"+strconv.FormatInt(category.ID, 10), nil, cfg.Auth.Token)
	if deleteCategoryRec.Code != http.StatusConflict {
		t.Fatalf("expected delete category conflict %d, got %d, body=%s", http.StatusConflict, deleteCategoryRec.Code, deleteCategoryRec.Body.String())
	}
}

func TestAdminPagesRequireLoginRedirect(t *testing.T) {
	repo := store.NewMemoryStore()

	cfg := config.Config{
		Auth: config.AuthConfig{
			Token:      "test-token",
			CookieName: "admin_token",
			Username:   "lane",
		},
	}

	router := NewRouter(cfg, admin.NewService(repo), frontend.NewService(repo))
	req := httptest.NewRequest(http.MethodGet, "/admin/article", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected redirect status %d, got %d", http.StatusFound, rec.Code)
	}
	if location := rec.Header().Get("Location"); !strings.HasPrefix(location, "/admin/login?redirect=") {
		t.Fatalf("expected redirect to login, got %s", location)
	}
}

func TestAdminPagesRenderWithCookie(t *testing.T) {
	repo := store.NewMemoryStore()

	cfg := config.Config{
		Auth: config.AuthConfig{
			Token:      "test-token",
			CookieName: "admin_token",
			Username:   "lane",
		},
	}

	router := NewRouter(cfg, admin.NewService(repo), frontend.NewService(repo))
	req := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, req)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected login page status %d, got %d", http.StatusOK, loginRec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/admin/article/create", nil)
	req.AddCookie(&http.Cookie{Name: cfg.Auth.CookieName, Value: cfg.Auth.Token})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected page status %d, got %d", http.StatusOK, rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "LaneBlog Admin") || !strings.Contains(body, "新建文章") {
		t.Fatalf("unexpected page body: %s", body)
	}

	assetReq := httptest.NewRequest(http.MethodGet, "/admin/assets/app.css", nil)
	assetRec := httptest.NewRecorder()
	router.ServeHTTP(assetRec, assetReq)
	if assetRec.Code != http.StatusOK {
		t.Fatalf("expected asset status %d, got %d", http.StatusOK, assetRec.Code)
	}
}

func TestFrontPagesAndAPIs(t *testing.T) {
	repo := store.NewMemoryStore()
	if err := repo.Write(func(snapshot *store.Snapshot) error {
		snapshot.Categories = []store.Category{
			{ID: 1, Name: "Go", InOut: 1, PID: 0},
			{ID: 2, Name: "教程", InOut: 1, PID: 1},
			{ID: 3, Name: "GitHub", InOut: 2, PID: 0, URL: "https://github.com"},
		}
		snapshot.Articles = []store.Article{
			{ID: 1, MID: 1, Author: "lane", Title: "第12篇", Description: "摘要12", Tag: "Go,SSR", Content: "<p>正文12</p>", CTime: 1200, Clicks: 5, GoodNum: 1, BadNum: 0, RecommendType: 1},
			{ID: 2, MID: 1, Author: "lane", Title: "第11篇", Description: "摘要11", Tag: "Go", Content: "<p>正文11</p>", CTime: 1100, Clicks: 0, GoodNum: 0, BadNum: 0, RecommendType: 2},
			{ID: 3, MID: 1, Author: "lane", Title: "第10篇", Description: "摘要10", Tag: "Go", Content: "<p>正文10</p>", CTime: 1000},
			{ID: 4, MID: 1, Author: "lane", Title: "第9篇", Description: "摘要9", Tag: "Go", Content: "<p>正文9</p>", CTime: 900},
			{ID: 5, MID: 1, Author: "lane", Title: "第8篇", Description: "摘要8", Tag: "Go", Content: "<p>正文8</p>", CTime: 800},
			{ID: 6, MID: 1, Author: "lane", Title: "第7篇", Description: "摘要7", Tag: "Go", Content: "<p>正文7</p>", CTime: 700},
			{ID: 7, MID: 1, Author: "lane", Title: "第6篇", Description: "摘要6", Tag: "Go", Content: "<p>正文6</p>", CTime: 600},
			{ID: 8, MID: 1, Author: "lane", Title: "第5篇", Description: "摘要5", Tag: "Go", Content: "<p>正文5</p>", CTime: 500},
			{ID: 9, MID: 1, Author: "lane", Title: "第4篇", Description: "摘要4", Tag: "Go", Content: "<p>正文4</p>", CTime: 400},
			{ID: 10, MID: 1, Author: "lane", Title: "第3篇", Description: "摘要3", Tag: "Go", Content: "<p>正文3</p>", CTime: 300},
			{ID: 11, MID: 1, Author: "lane", Title: "第2篇", Description: "摘要2", Tag: "Go", Content: "<p>正文2</p>", CTime: 200},
			{ID: 12, MID: 1, Author: "lane", Title: "第1篇", Description: "摘要1", Tag: "Go", Content: "<p>正文1</p>", CTime: 100},
		}
		return nil
	}); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	cfg := config.Config{Auth: config.AuthConfig{Token: "test-token", CookieName: "admin_token", Username: "lane"}}
	router := NewRouter(cfg, admin.NewService(repo), frontend.NewService(repo))

	homeReq := httptest.NewRequest(http.MethodGet, "/", nil)
	homeRec := httptest.NewRecorder()
	router.ServeHTTP(homeRec, homeReq)
	if homeRec.Code != http.StatusOK {
		t.Fatalf("expected home status %d, got %d", http.StatusOK, homeRec.Code)
	}
	homeBody := homeRec.Body.String()
	if !strings.Contains(homeBody, "LaneBlog") || !strings.Contains(homeBody, "每一个没有起舞的日子都是在辜负生命。") {
		t.Fatalf("unexpected home body: %s", homeBody)
	}
	if strings.Index(homeBody, "第12篇") > strings.Index(homeBody, "第11篇") {
		t.Fatalf("expected articles sorted by time desc, body=%s", homeBody)
	}
	if strings.Contains(homeBody, "第2篇") || !strings.Contains(homeBody, "第3篇") {
		t.Fatalf("expected first page to contain top 10 articles only, body=%s", homeBody)
	}
	if !strings.Contains(homeBody, `<meta name="description"`) || !strings.Contains(homeBody, `href="https://github.com"`) {
		t.Fatalf("expected seo and nav output, body=%s", homeBody)
	}
	if !strings.Contains(homeBody, `href="/?page=2"`) {
		t.Fatalf("expected pagination link in home body, body=%s", homeBody)
	}

	homePage2Req := httptest.NewRequest(http.MethodGet, "/?page=2", nil)
	homePage2Rec := httptest.NewRecorder()
	router.ServeHTTP(homePage2Rec, homePage2Req)
	if homePage2Rec.Code != http.StatusOK {
		t.Fatalf("expected page2 home status %d, got %d", http.StatusOK, homePage2Rec.Code)
	}
	homePage2Body := homePage2Rec.Body.String()
	if !strings.Contains(homePage2Body, "第2篇") || !strings.Contains(homePage2Body, "第1篇") || strings.Contains(homePage2Body, "第3篇") {
		t.Fatalf("unexpected page2 home body: %s", homePage2Body)
	}

	articleReq := httptest.NewRequest(http.MethodGet, "/article/1", nil)
	articleRec := httptest.NewRecorder()
	router.ServeHTTP(articleRec, articleReq)
	if articleRec.Code != http.StatusOK {
		t.Fatalf("expected article status %d, got %d", http.StatusOK, articleRec.Code)
	}
	articleBody := articleRec.Body.String()
	if !strings.Contains(articleBody, "点击：<strong data-clicks-num>6</strong>") {
		t.Fatalf("expected clicks incremented in article page, body=%s", articleBody)
	}

	voteRec := performJSONRequest(router, http.MethodPost, "/api/v1/front/articles/1/good", []byte(`{}`), "")
	if voteRec.Code != http.StatusOK {
		t.Fatalf("expected good api status %d, got %d, body=%s", http.StatusOK, voteRec.Code, voteRec.Body.String())
	}
	var vote struct {
		GoodNum int `json:"good_num"`
		BadNum  int `json:"bad_num"`
	}
	if err := json.Unmarshal(voteRec.Body.Bytes(), &vote); err != nil {
		t.Fatalf("decode vote: %v", err)
	}
	if vote.GoodNum != 2 || vote.BadNum != 0 {
		t.Fatalf("unexpected vote result: %#v", vote)
	}

	apiRec := performJSONRequest(router, http.MethodGet, "/api/v1/front/home", nil, "")
	if apiRec.Code != http.StatusOK {
		t.Fatalf("expected home api status %d, got %d", http.StatusOK, apiRec.Code)
	}

	assetReq := httptest.NewRequest(http.MethodGet, "/assets/front.css", nil)
	assetRec := httptest.NewRecorder()
	router.ServeHTTP(assetRec, assetReq)
	if assetRec.Code != http.StatusOK {
		t.Fatalf("expected front asset status %d, got %d", http.StatusOK, assetRec.Code)
	}
}

func performJSONRequest(handler http.Handler, method, target string, body []byte, token string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, target, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func mustJSON(t *testing.T, payload map[string]any) []byte {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return raw
}
