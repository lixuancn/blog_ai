package frontend

import (
	"strconv"
	"strings"
	"testing"

	"example.com/laneblog/internal/store"
)

func TestServiceHomePageDataBuildsNavigationAndPreview(t *testing.T) {
	repo := store.NewMemoryStore()
	err := repo.Write(func(snapshot *store.Snapshot) error {
		snapshot.Categories = []store.Category{
			{ID: 1, Name: "Go", InOut: 1, PID: 0},
			{ID: 2, Name: "后端", InOut: 1, PID: 1},
			{ID: 3, Name: "GitHub", InOut: 2, PID: 0, URL: "https://github.com"},
		}
		snapshot.Articles = []store.Article{
			{ID: 1, MID: 1, Author: "lane", Title: "旧文章", Description: "旧摘要", Tag: "Go,后端", Content: "<p>旧内容</p>", CTime: 100, RecommendType: 1},
			{ID: 2, MID: 1, Author: "lane", Title: "新文章", Description: "新摘要", Tag: "Go,实践", Content: "<p>你好，世界</p>" + strings.Repeat("测", 110), CTime: 200, RecommendType: 1},
			{ID: 3, MID: 2, Author: "lane", Title: "普通文章", Description: "普通摘要", Tag: "后端", Content: "<div>内容</div>", CTime: 150, RecommendType: 2},
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed store: %v", err)
	}

	svc := NewService(repo)
	data := svc.HomePageData(1)

	if data.HeroTitle != siteTitle || data.HeroSubtitle != siteSubtitle {
		t.Fatalf("unexpected hero: %#v", data)
	}
	if len(data.Navigation) != 3 {
		t.Fatalf("expected home + 2 top nav items, got %d", len(data.Navigation))
	}
	if data.Navigation[1].Name != "Go" || len(data.Navigation[1].Children) != 1 || data.Navigation[1].Children[0].Name != "后端" {
		t.Fatalf("unexpected navigation tree: %#v", data.Navigation)
	}
	if !data.Navigation[2].External || data.Navigation[2].Href != "https://github.com" {
		t.Fatalf("expected external nav item, got %#v", data.Navigation[2])
	}
	if len(data.Articles) != 3 || data.Articles[0].Title != "新文章" {
		t.Fatalf("unexpected article order: %#v", data.Articles)
	}
	if !strings.HasSuffix(data.Articles[0].ContentPreview, "...") {
		t.Fatalf("expected truncated preview, got %q", data.Articles[0].ContentPreview)
	}
	if strings.Contains(data.Articles[0].ContentPreview, "<p>") {
		t.Fatalf("expected preview without html tags, got %q", data.Articles[0].ContentPreview)
	}
	if len(data.Recommendations) != 2 || data.Recommendations[0].Title != "新文章" || data.Recommendations[1].Title != "旧文章" {
		t.Fatalf("unexpected site recommendations: %#v", data.Recommendations)
	}
	if data.Pagination.Page != 1 || data.Pagination.TotalPages != 1 || data.Pagination.Total != 3 {
		t.Fatalf("unexpected pagination: %#v", data.Pagination)
	}
}

func TestServiceHomePageDataPaginatesArticles(t *testing.T) {
	repo := store.NewMemoryStore()
	err := repo.Write(func(snapshot *store.Snapshot) error {
		snapshot.Categories = []store.Category{{ID: 1, Name: "Go", InOut: 1, PID: 0}}
		for i := 1; i <= 12; i++ {
			snapshot.Articles = append(snapshot.Articles, store.Article{
				ID:          int64(i),
				MID:         1,
				Author:      "lane",
				Title:       "文章" + strconv.Itoa(i),
				Description: "摘要",
				Content:     "正文",
				CTime:       int64(i),
			})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed store: %v", err)
	}

	svc := NewService(repo)
	page2 := svc.HomePageData(2)
	if page2.Pagination.Page != 2 || page2.Pagination.TotalPages != 2 || !page2.Pagination.HasPrev || page2.Pagination.HasNext {
		t.Fatalf("unexpected page2 pagination: %#v", page2.Pagination)
	}
	if len(page2.Articles) != 2 {
		t.Fatalf("expected 2 articles on page 2, got %d", len(page2.Articles))
	}
	if page2.Articles[0].Title != "文章2" || page2.Articles[1].Title != "文章1" {
		t.Fatalf("unexpected page2 article order: %#v", page2.Articles)
	}
	if len(page2.Pagination.Items) != 2 {
		t.Fatalf("expected compact pagination items for 2 pages, got %#v", page2.Pagination.Items)
	}
}

func TestBuildPaginationItemsCompactsLongPageList(t *testing.T) {
	items := buildPaginationItems(10, 11)
	labels := make([]string, 0, len(items))
	for _, item := range items {
		labels = append(labels, item.Label)
	}
	got := strings.Join(labels, ",")
	want := "1,...,7,8,9,10,11"
	if got != want {
		t.Fatalf("unexpected compact pagination items: got %s want %s", got, want)
	}
}

func TestServiceArticlePageDataIncrementsClicksAndVotes(t *testing.T) {
	repo := store.NewMemoryStore()
	err := repo.Write(func(snapshot *store.Snapshot) error {
		snapshot.Categories = []store.Category{{ID: 1, Name: "Go", InOut: 1, PID: 0}}
		snapshot.Articles = []store.Article{
			{ID: 1, MID: 1, Author: "lane", Title: "当前文章", Description: "当前摘要", Tag: "Go,SSR", Content: "<p>正文</p>", CTime: 100, Clicks: 9, GoodNum: 1, BadNum: 2, RecommendType: 1},
			{ID: 2, MID: 1, Author: "lane", Title: "同类推荐", Description: "推荐", Tag: "Go", Content: "推荐", CTime: 120, RecommendType: 2},
			{ID: 3, MID: 1, Author: "lane", Title: "同类最新", Description: "最新", Tag: "Go", Content: "最新", CTime: 110, RecommendType: 1},
			{ID: 4, MID: 1, Author: "lane", Title: "更旧文章", Description: "更旧", Tag: "Go", Content: "更旧", CTime: 90, RecommendType: 1},
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed store: %v", err)
	}

	svc := NewService(repo)
	data, err := svc.ArticlePageData(1)
	if err != nil {
		t.Fatalf("article page data: %v", err)
	}
	if data.Article.Clicks != 10 {
		t.Fatalf("expected clicks 10, got %d", data.Article.Clicks)
	}
	if len(data.Recommendations) < 2 || data.Recommendations[0].Title != "同类推荐" || data.Recommendations[1].Title != "同类最新" {
		t.Fatalf("unexpected recommendations: %#v", data.Recommendations)
	}

	vote, err := svc.LikeArticle(1)
	if err != nil {
		t.Fatalf("like article: %v", err)
	}
	if vote.GoodNum != 2 || vote.BadNum != 2 {
		t.Fatalf("unexpected like result: %#v", vote)
	}

	vote, err = svc.DislikeArticle(1)
	if err != nil {
		t.Fatalf("dislike article: %v", err)
	}
	if vote.GoodNum != 2 || vote.BadNum != 3 {
		t.Fatalf("unexpected dislike result: %#v", vote)
	}
}
