package http

import (
	stdhttp "net/http"

	"example.com/laneblog/internal/admin"
	"example.com/laneblog/internal/config"
	"example.com/laneblog/internal/frontend"
	"example.com/laneblog/internal/http/handler"
	"example.com/laneblog/internal/http/middleware"
)

func NewRouter(cfg config.Config, adminService *admin.Service, frontService *frontend.Service) stdhttp.Handler {
	mux := stdhttp.NewServeMux()

	systemHandler := handler.NewSystemHandler(cfg)
	authHandler := handler.NewAuthHandler(cfg)
	adminHandler := handler.NewAdminHandler(adminService)
	adminUI := handler.NewAdminUIHandler(cfg)
	frontHandler := handler.NewFrontHandler(frontService)

	mux.HandleFunc("GET /healthz", systemHandler.Health)
	mux.HandleFunc("GET /readyz", systemHandler.Ready)

	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)
	mux.Handle("GET /api/v1/auth/me", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(authHandler.Me)))

	mux.Handle("GET /api/v1/admin/dashboard", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.Dashboard)))
	mux.Handle("GET /api/v1/admin/categories", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.ListCategories)))
	mux.Handle("POST /api/v1/admin/categories", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.CreateCategory)))
	mux.Handle("GET /api/v1/admin/categories/{id}", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.GetCategory)))
	mux.Handle("PUT /api/v1/admin/categories/{id}", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.UpdateCategory)))
	mux.Handle("DELETE /api/v1/admin/categories/{id}", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.DeleteCategory)))

	mux.Handle("GET /api/v1/admin/articles", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.ListArticles)))
	mux.Handle("POST /api/v1/admin/articles", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.CreateArticle)))
	mux.Handle("GET /api/v1/admin/articles/{id}", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.GetArticle)))
	mux.Handle("PUT /api/v1/admin/articles/{id}", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.UpdateArticle)))
	mux.Handle("DELETE /api/v1/admin/articles/{id}", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.DeleteArticle)))

	mux.Handle("GET /api/v1/admin/tags", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.ListTags)))
	mux.Handle("DELETE /api/v1/admin/tags/{id}", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.DeleteTag)))
	mux.Handle("POST /api/v1/admin/tags/rebuild", middleware.RequireAuth(cfg, stdhttp.HandlerFunc(adminHandler.RebuildTags)))

	mux.HandleFunc("GET /api/v1/front/home", frontHandler.HomeAPI)
	mux.HandleFunc("GET /api/v1/front/articles/{id}", frontHandler.ArticleAPI)
	mux.HandleFunc("POST /api/v1/front/articles/{id}/good", frontHandler.LikeArticle)
	mux.HandleFunc("POST /api/v1/front/articles/{id}/bad", frontHandler.DislikeArticle)

	mux.HandleFunc("GET /", frontHandler.HomePage)
	mux.HandleFunc("GET /article/{id}", frontHandler.ArticlePage)
	mux.HandleFunc("GET /category/{id}", frontHandler.CategoryPage)
	mux.HandleFunc("GET /assets/front.css", serveFrontCSS)
	mux.HandleFunc("GET /assets/front.js", serveFrontJS)

	mux.HandleFunc("GET /admin/login", adminUI.LoginPage)
	mux.Handle("GET /admin", middleware.RequirePageAuth(cfg, stdhttp.HandlerFunc(adminUI.DashboardPage)))
	mux.Handle("GET /admin/category", middleware.RequirePageAuth(cfg, stdhttp.HandlerFunc(adminUI.CategoryListPage)))
	mux.Handle("GET /admin/category/create", middleware.RequirePageAuth(cfg, stdhttp.HandlerFunc(adminUI.CategoryCreatePage)))
	mux.Handle("GET /admin/category/edit/{id}", middleware.RequirePageAuth(cfg, stdhttp.HandlerFunc(adminUI.CategoryEditPage)))
	mux.Handle("GET /admin/article", middleware.RequirePageAuth(cfg, stdhttp.HandlerFunc(adminUI.ArticleListPage)))
	mux.Handle("GET /admin/article/create", middleware.RequirePageAuth(cfg, stdhttp.HandlerFunc(adminUI.ArticleCreatePage)))
	mux.Handle("GET /admin/article/edit/{id}", middleware.RequirePageAuth(cfg, stdhttp.HandlerFunc(adminUI.ArticleEditPage)))
	mux.Handle("GET /admin/tag", middleware.RequirePageAuth(cfg, stdhttp.HandlerFunc(adminUI.TagListPage)))
	mux.HandleFunc("GET /admin/assets/app.css", serveAdminCSS)
	mux.HandleFunc("GET /admin/assets/app.js", serveAdminJS)

	return middleware.Recovery(middleware.Logging(mux))
}
