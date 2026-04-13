package admin

import (
	"strconv"
	"testing"
	"time"

	"example.com/laneblog/internal/store"
)

func TestServiceCategoryDeleteConstraint(t *testing.T) {
	svc := NewService(store.NewMemoryStore())

	parent, err := svc.CreateCategory(CategoryInput{
		Name:           "后端开发",
		SEOTitle:       "后端开发",
		SEODescription: "后端开发描述",
		SEOKeywords:    "go,backend",
		InOut:          1,
		PID:            0,
		Item:           "backend",
	})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}

	_, err = svc.CreateCategory(CategoryInput{
		Name:           "Go",
		SEOTitle:       "Go",
		SEODescription: "Go 描述",
		SEOKeywords:    "go",
		InOut:          1,
		PID:            parent.ID,
		Item:           "go",
	})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	err = svc.DeleteCategory(parent.ID)
	if err == nil {
		t.Fatal("expected delete category with child to fail")
	}

	appErr, ok := err.(*AppError)
	if !ok || appErr.Status != 409 {
		t.Fatalf("expected conflict error, got %#v", err)
	}
}

func TestServiceArticleTagSyncAndDashboard(t *testing.T) {
	svc := NewService(store.NewMemoryStore())
	svc.now = func() time.Time {
		return time.Unix(1710000000, 0)
	}

	category, err := svc.CreateCategory(CategoryInput{
		Name:           "Go",
		SEOTitle:       "Go",
		SEODescription: "Go 分类",
		SEOKeywords:    "go",
		InOut:          1,
		PID:            0,
		Item:           "go",
	})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	article, err := svc.CreateArticle(ArticleInput{
		MID:            category.ID,
		Author:         "lane",
		Title:          "Go 后台开发实践",
		Description:    "文章摘要",
		SEOTitle:       "Go 后台开发实践",
		SEODescription: "SEO 描述",
		SEOKeywords:    "go,admin",
		Tag:            "Go, Backend, Go",
		Content:        "content",
		RecommendType:  1,
	})
	if err != nil {
		t.Fatalf("create article: %v", err)
	}

	tags := svc.ListTags("")
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags after create, got %d", len(tags))
	}

	updated, err := svc.UpdateArticle(article.ID, ArticleInput{
		MID:            category.ID,
		Author:         "lane",
		Title:          "Go 后台开发实践（更新）",
		Description:    "文章摘要",
		SEOTitle:       "Go 后台开发实践",
		SEODescription: "SEO 描述",
		SEOKeywords:    "go,admin",
		Tag:            "Go, API",
		Content:        "content",
		RecommendType:  2,
		CTime:          article.CTime,
	})
	if err != nil {
		t.Fatalf("update article: %v", err)
	}

	if updated.RecommendType != 2 {
		t.Fatalf("expected recommend type 2, got %d", updated.RecommendType)
	}

	tags = svc.ListTags("")
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags after update, got %d", len(tags))
	}
	if tags[0].Tag != "API" && tags[1].Tag != "API" {
		t.Fatalf("expected tags to contain API, got %#v", tags)
	}

	dashboard := svc.Dashboard()
	if dashboard.Summary.ArticleCount != 1 {
		t.Fatalf("expected article count 1, got %d", dashboard.Summary.ArticleCount)
	}
	if dashboard.Summary.CategoryCount != 1 {
		t.Fatalf("expected category count 1, got %d", dashboard.Summary.CategoryCount)
	}
	if dashboard.Summary.TagCount != 2 {
		t.Fatalf("expected tag count 2, got %d", dashboard.Summary.TagCount)
	}
	if dashboard.Summary.RecommendedHomeCount != 1 || dashboard.Summary.RecommendedSiteCount != 0 {
		t.Fatalf("unexpected recommend summary: %#v", dashboard.Summary)
	}
	if len(dashboard.RecentArticles) != 1 {
		t.Fatalf("expected 1 recent article, got %d", len(dashboard.RecentArticles))
	}

	if err := svc.DeleteArticle(article.ID); err != nil {
		t.Fatalf("delete article: %v", err)
	}
	if len(svc.ListTags("")) != 0 {
		t.Fatal("expected tags to be cleared after deleting last article")
	}
}

func TestServiceListArticlesWithFiltersAndPagination(t *testing.T) {
	svc := NewService(store.NewMemoryStore())

	category, err := svc.CreateCategory(CategoryInput{
		Name:           "数据库",
		SEOTitle:       "数据库",
		SEODescription: "数据库分类",
		SEOKeywords:    "mysql",
		InOut:          1,
		PID:            0,
		Item:           "db",
	})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	for i := 1; i <= 3; i++ {
		_, err := svc.CreateArticle(ArticleInput{
			MID:            category.ID,
			Author:         "lane",
			Title:          "MySQL 优化记录 " + strconv.Itoa(i),
			Description:    "文章摘要",
			SEOTitle:       "MySQL 优化记录",
			SEODescription: "SEO 描述",
			SEOKeywords:    "mysql",
			Tag:            "MySQL",
			Content:        "content",
			RecommendType:  1,
			CTime:          int64(i),
		})
		if err != nil {
			t.Fatalf("create article %d: %v", i, err)
		}
	}

	page := svc.ListArticles(ArticleQuery{
		Title:         "mysql",
		CategoryID:    category.ID,
		RecommendType: 1,
		Page:          2,
		PageSize:      2,
	})

	if page.Total != 3 {
		t.Fatalf("expected total 3, got %d", page.Total)
	}
	if page.TotalPages != 2 {
		t.Fatalf("expected total pages 2, got %d", page.TotalPages)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 article on second page, got %d", len(page.Items))
	}
}
