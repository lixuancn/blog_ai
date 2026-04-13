package frontend

import (
	"errors"
	"html"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"example.com/laneblog/internal/store"
)

const (
	siteTitle           = "LaneBlog"
	siteSubtitle        = "每一个没有起舞的日子都是在辜负生命。"
	homeSEOTitle        = "LaneBlog | 技术博客"
	homeSEODescription  = "LaneBlog 分享编程、后端开发与技术实践，专注内容阅读与沉淀。"
	homeSEOKeywords     = "LaneBlog,博客,Go,后端开发,技术实践"
	defaultRecommendNum = 8
	defaultHomePageSize = 10
)

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

type Store interface {
	Read() store.Snapshot
	Write(func(*store.Snapshot) error) error
}

type AppError struct {
	Status  int    `json:"-"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

type SEO struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
}

type SiteInfo struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

type NavItem struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Href     string    `json:"href"`
	External bool      `json:"external"`
	Children []NavItem `json:"children,omitempty"`
}

type CategoryInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Href        string `json:"href"`
}

type ArticleSummary struct {
	ID             int64    `json:"id"`
	MID            int64    `json:"mid"`
	Title          string   `json:"title"`
	Author         string   `json:"author"`
	CategoryName   string   `json:"category_name"`
	CategoryHref   string   `json:"category_href"`
	Description    string   `json:"description"`
	ContentPreview string   `json:"content_preview"`
	Tags           []string `json:"tags"`
	CTime          int64    `json:"ctime"`
	Link           string   `json:"link"`
}

type RecommendItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Link  string `json:"link"`
}

type ListingPageData struct {
	SEO             SEO              `json:"seo"`
	Site            SiteInfo         `json:"site"`
	Navigation      []NavItem        `json:"navigation"`
	HeroTitle       string           `json:"hero_title"`
	HeroSubtitle    string           `json:"hero_subtitle"`
	Articles        []ArticleSummary `json:"articles"`
	Recommendations []RecommendItem  `json:"recommendations"`
	CurrentCategory *CategoryInfo    `json:"current_category,omitempty"`
	Pagination      Pagination       `json:"pagination"`
}

type Pagination struct {
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	Total      int              `json:"total"`
	TotalPages int              `json:"total_pages"`
	PrevPage   int              `json:"prev_page"`
	NextPage   int              `json:"next_page"`
	HasPrev    bool             `json:"has_prev"`
	HasNext    bool             `json:"has_next"`
	Pages      []int            `json:"pages"`
	Items      []PaginationItem `json:"items"`
}

type PaginationItem struct {
	Page       int    `json:"page"`
	Label      string `json:"label"`
	Active     bool   `json:"active"`
	IsEllipsis bool   `json:"is_ellipsis"`
}

type ArticleDetail struct {
	ID           int64    `json:"id"`
	MID          int64    `json:"mid"`
	Title        string   `json:"title"`
	Author       string   `json:"author"`
	CategoryName string   `json:"category_name"`
	CategoryHref string   `json:"category_href"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	Content      string   `json:"content"`
	CTime        int64    `json:"ctime"`
	Clicks       int      `json:"clicks"`
	GoodNum      int      `json:"good_num"`
	BadNum       int      `json:"bad_num"`
}

type ArticlePageData struct {
	SEO             SEO             `json:"seo"`
	Site            SiteInfo        `json:"site"`
	Navigation      []NavItem       `json:"navigation"`
	HeroTitle       string          `json:"hero_title"`
	HeroSubtitle    string          `json:"hero_subtitle"`
	Article         ArticleDetail   `json:"article"`
	Recommendations []RecommendItem `json:"recommendations"`
}

type VoteResult struct {
	ID      int64 `json:"id"`
	GoodNum int   `json:"good_num"`
	BadNum  int   `json:"bad_num"`
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) HomePageData(page int) ListingPageData {
	snapshot := s.store.Read()
	categories := categoryMap(snapshot.Categories)
	articles, pagination := buildListingArticles(snapshot.Articles, categories, 0, page, defaultHomePageSize)
	return ListingPageData{
		SEO:             SEO{Title: homeSEOTitle, Description: homeSEODescription, Keywords: homeSEOKeywords},
		Site:            defaultSiteInfo(),
		Navigation:      buildNavigation(snapshot.Categories),
		HeroTitle:       siteTitle,
		HeroSubtitle:    siteSubtitle,
		Articles:        articles,
		Recommendations: buildSiteRecommendations(snapshot.Articles),
		Pagination:      pagination,
	}
}

func (s *Service) CategoryPageData(id int64) (ListingPageData, error) {
	snapshot := s.store.Read()
	category, ok := findCategory(snapshot.Categories, id)
	if !ok {
		return ListingPageData{}, notFound("分类不存在")
	}

	categories := categoryMap(snapshot.Categories)
	heroSubtitle := firstNonEmpty(category.SEODescription, siteSubtitle)
	return ListingPageData{
		SEO: SEO{
			Title:       firstNonEmpty(category.SEOTitle, category.Name+" | "+siteTitle),
			Description: firstNonEmpty(category.SEODescription, "查看 "+category.Name+" 分类下的文章内容。"),
			Keywords:    firstNonEmpty(category.SEOKeywords, joinKeywords(category.Name, siteTitle)),
		},
		Site:            defaultSiteInfo(),
		Navigation:      buildNavigation(snapshot.Categories),
		HeroTitle:       category.Name,
		HeroSubtitle:    heroSubtitle,
		Articles:        buildListingArticlesAll(snapshot.Articles, categories, id),
		Recommendations: buildCategoryRecommendations(snapshot.Articles, id, 0),
		CurrentCategory: &CategoryInfo{ID: category.ID, Name: category.Name, Description: heroSubtitle, Href: categoryHref(category.ID)},
	}, nil
}

func (s *Service) ArticlePageData(id int64) (ArticlePageData, error) {
	if err := s.incrementClicks(id); err != nil {
		return ArticlePageData{}, err
	}

	snapshot := s.store.Read()
	article, ok := findArticle(snapshot.Articles, id)
	if !ok {
		return ArticlePageData{}, notFound("文章不存在")
	}

	categories := categoryMap(snapshot.Categories)
	category := categories[article.MID]
	return ArticlePageData{
		SEO: SEO{
			Title:       firstNonEmpty(article.SEOTitle, article.Title+" | "+siteTitle),
			Description: firstNonEmpty(article.SEODescription, article.Description, truncateRunes(plainContent(article.Content), 120)),
			Keywords:    firstNonEmpty(article.SEOKeywords, joinKeywords(article.Tag, category.Name, siteTitle)),
		},
		Site:         defaultSiteInfo(),
		Navigation:   buildNavigation(snapshot.Categories),
		HeroTitle:    article.Title,
		HeroSubtitle: firstNonEmpty(article.Description, siteSubtitle),
		Article: ArticleDetail{
			ID:           article.ID,
			MID:          article.MID,
			Title:        article.Title,
			Author:       article.Author,
			CategoryName: category.Name,
			CategoryHref: categoryHref(article.MID),
			Description:  article.Description,
			Tags:         splitTags(article.Tag),
			Content:      article.Content,
			CTime:        article.CTime,
			Clicks:       article.Clicks,
			GoodNum:      article.GoodNum,
			BadNum:       article.BadNum,
		},
		Recommendations: buildCategoryRecommendations(snapshot.Articles, article.MID, article.ID),
	}, nil
}

func (s *Service) ArticleData(id int64) (ArticlePageData, error) {
	return s.ArticlePageData(id)
}

func (s *Service) LikeArticle(id int64) (VoteResult, error) {
	return s.voteArticle(id, true)
}

func (s *Service) DislikeArticle(id int64) (VoteResult, error) {
	return s.voteArticle(id, false)
}

func (s *Service) incrementClicks(id int64) error {
	return s.store.Write(func(snapshot *store.Snapshot) error {
		index := articleIndex(snapshot.Articles, id)
		if index < 0 {
			return notFound("文章不存在")
		}
		snapshot.Articles[index].Clicks++
		return nil
	})
}

func (s *Service) voteArticle(id int64, good bool) (VoteResult, error) {
	result := VoteResult{}
	err := s.store.Write(func(snapshot *store.Snapshot) error {
		index := articleIndex(snapshot.Articles, id)
		if index < 0 {
			return notFound("文章不存在")
		}
		if good {
			snapshot.Articles[index].GoodNum++
		} else {
			snapshot.Articles[index].BadNum++
		}
		result = VoteResult{
			ID:      snapshot.Articles[index].ID,
			GoodNum: snapshot.Articles[index].GoodNum,
			BadNum:  snapshot.Articles[index].BadNum,
		}
		return nil
	})
	if err != nil {
		return VoteResult{}, err
	}
	return result, nil
}

func buildNavigation(categories []store.Category) []NavItem {
	tops := make([]store.Category, 0)
	childMap := make(map[int64][]store.Category)
	for _, category := range categories {
		if category.PID == 0 {
			tops = append(tops, category)
			continue
		}
		childMap[category.PID] = append(childMap[category.PID], category)
	}

	sort.Slice(tops, func(i, j int) bool { return tops[i].ID < tops[j].ID })
	for id := range childMap {
		sort.Slice(childMap[id], func(i, j int) bool { return childMap[id][i].ID < childMap[id][j].ID })
	}

	nav := make([]NavItem, 0, len(tops)+1)
	nav = append(nav, NavItem{ID: 0, Name: "首页", Href: "/"})
	for _, top := range tops {
		item := NavItem{
			ID:       top.ID,
			Name:     top.Name,
			Href:     resolveCategoryHref(top),
			External: top.InOut == 2,
		}
		for _, child := range childMap[top.ID] {
			item.Children = append(item.Children, NavItem{
				ID:       child.ID,
				Name:     child.Name,
				Href:     resolveCategoryHref(child),
				External: child.InOut == 2,
			})
		}
		nav = append(nav, item)
	}
	return nav
}

func buildListingArticlesAll(articles []store.Article, categories map[int64]store.Category, categoryID int64) []ArticleSummary {
	sorted := filterAndSortArticles(articles, func(article store.Article) bool {
		return categoryID == 0 || article.MID == categoryID
	})
	return toArticleSummaries(sorted, categories)
}

func buildListingArticles(articles []store.Article, categories map[int64]store.Category, categoryID int64, page, pageSize int) ([]ArticleSummary, Pagination) {
	sorted := filterAndSortArticles(articles, func(article store.Article) bool {
		return categoryID == 0 || article.MID == categoryID
	})
	paged, pagination := paginateArticles(sorted, page, pageSize)
	return toArticleSummaries(paged, categories), pagination
}

func toArticleSummaries(articles []store.Article, categories map[int64]store.Category) []ArticleSummary {
	items := make([]ArticleSummary, 0, len(articles))
	for _, article := range articles {
		category := categories[article.MID]
		items = append(items, ArticleSummary{
			ID:             article.ID,
			MID:            article.MID,
			Title:          article.Title,
			Author:         article.Author,
			CategoryName:   category.Name,
			CategoryHref:   categoryHref(article.MID),
			Description:    article.Description,
			ContentPreview: truncateRunes(plainContent(article.Content), 100),
			Tags:           splitTags(article.Tag),
			CTime:          article.CTime,
			Link:           articleHref(article.ID),
		})
	}
	return items
}

func paginateArticles(articles []store.Article, page, pageSize int) ([]store.Article, Pagination) {
	if pageSize <= 0 {
		pageSize = defaultHomePageSize
	}
	total := len(articles)
	totalPages := 1
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	if page <= 0 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	pagination := Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		PrevPage:   maxInt(page-1, 1),
		NextPage:   minInt(page+1, totalPages),
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
	}
	pagination.Pages = make([]int, 0, totalPages)
	for i := 1; i <= totalPages; i++ {
		pagination.Pages = append(pagination.Pages, i)
	}
	pagination.Items = buildPaginationItems(page, totalPages)
	return articles[start:end], pagination
}

func buildPaginationItems(currentPage, totalPages int) []PaginationItem {
	if totalPages <= 0 {
		return []PaginationItem{{Page: 1, Label: "1", Active: true}}
	}

	addPage := func(items *[]PaginationItem, page int, active bool) {
		*items = append(*items, PaginationItem{
			Page:   page,
			Label:  strconv.Itoa(page),
			Active: active,
		})
	}
	addEllipsis := func(items *[]PaginationItem) {
		if len(*items) == 0 || (*items)[len(*items)-1].IsEllipsis {
			return
		}
		*items = append(*items, PaginationItem{Label: "...", IsEllipsis: true})
	}

	items := make([]PaginationItem, 0, minInt(totalPages, 7))
	if totalPages <= 7 {
		for page := 1; page <= totalPages; page++ {
			addPage(&items, page, page == currentPage)
		}
		return items
	}

	addPage(&items, 1, currentPage == 1)
	switch {
	case currentPage <= 4:
		for page := 2; page <= 5; page++ {
			addPage(&items, page, page == currentPage)
		}
		addEllipsis(&items)
	case currentPage >= totalPages-3:
		addEllipsis(&items)
		for page := totalPages - 4; page <= totalPages-1; page++ {
			addPage(&items, page, page == currentPage)
		}
	default:
		addEllipsis(&items)
		for page := currentPage - 1; page <= currentPage+1; page++ {
			addPage(&items, page, page == currentPage)
		}
		addEllipsis(&items)
	}
	addPage(&items, totalPages, currentPage == totalPages)
	return items
}

func buildSiteRecommendations(articles []store.Article) []RecommendItem {
	sorted := filterAndSortArticles(articles, func(article store.Article) bool {
		return article.RecommendType == 1
	})
	return toRecommendItems(sorted, 0, defaultRecommendNum)
}

func buildCategoryRecommendations(articles []store.Article, categoryID, excludeID int64) []RecommendItem {
	primary := filterAndSortArticles(articles, func(article store.Article) bool {
		return article.MID == categoryID && article.ID != excludeID && article.RecommendType == 2
	})
	fallback := filterAndSortArticles(articles, func(article store.Article) bool {
		return article.MID == categoryID && article.ID != excludeID && article.RecommendType != 2
	})

	combined := append([]store.Article(nil), primary...)
	for _, article := range fallback {
		if len(combined) >= defaultRecommendNum {
			break
		}
		combined = append(combined, article)
	}
	return toRecommendItems(combined, 0, defaultRecommendNum)
}

func filterAndSortArticles(articles []store.Article, keep func(store.Article) bool) []store.Article {
	filtered := make([]store.Article, 0, len(articles))
	for _, article := range articles {
		if keep(article) {
			filtered = append(filtered, article)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].CTime == filtered[j].CTime {
			return filtered[i].ID > filtered[j].ID
		}
		return filtered[i].CTime > filtered[j].CTime
	})
	return filtered
}

func toRecommendItems(articles []store.Article, skip, limit int) []RecommendItem {
	if skip > len(articles) {
		skip = len(articles)
	}
	articles = articles[skip:]
	if limit > 0 && len(articles) > limit {
		articles = articles[:limit]
	}
	items := make([]RecommendItem, 0, len(articles))
	for _, article := range articles {
		items = append(items, RecommendItem{ID: article.ID, Title: article.Title, Link: articleHref(article.ID)})
	}
	return items
}

func defaultSiteInfo() SiteInfo {
	return SiteInfo{Title: siteTitle, Subtitle: siteSubtitle}
}

func categoryMap(categories []store.Category) map[int64]store.Category {
	items := make(map[int64]store.Category, len(categories))
	for _, category := range categories {
		items[category.ID] = category
	}
	return items
}

func findCategory(categories []store.Category, id int64) (store.Category, bool) {
	for _, category := range categories {
		if category.ID == id {
			return category, true
		}
	}
	return store.Category{}, false
}

func findArticle(articles []store.Article, id int64) (store.Article, bool) {
	for _, article := range articles {
		if article.ID == id {
			return article, true
		}
	}
	return store.Article{}, false
}

func articleIndex(articles []store.Article, id int64) int {
	for index, article := range articles {
		if article.ID == id {
			return index
		}
	}
	return -1
}

func resolveCategoryHref(category store.Category) string {
	if category.InOut == 2 && strings.TrimSpace(category.URL) != "" {
		return strings.TrimSpace(category.URL)
	}
	return categoryHref(category.ID)
}

func articleHref(id int64) string {
	return "/article/" + strconv.FormatInt(id, 10)
}

func categoryHref(id int64) string {
	return "/category/" + strconv.FormatInt(id, 10)
}

func splitTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		tags = append(tags, tag)
	}
	return tags
}

func plainContent(content string) string {
	withoutTags := htmlTagPattern.ReplaceAllString(content, " ")
	unescaped := html.UnescapeString(withoutTags)
	return strings.Join(strings.Fields(unescaped), " ")
}

func truncateRunes(value string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" || limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return strings.TrimSpace(string(runes[:limit])) + "..."
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func joinKeywords(values ...string) string {
	keywords := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			keyword := strings.TrimSpace(part)
			if keyword == "" {
				continue
			}
			lower := strings.ToLower(keyword)
			if _, ok := seen[lower]; ok {
				continue
			}
			seen[lower] = struct{}{}
			keywords = append(keywords, keyword)
		}
	}
	return strings.Join(keywords, ",")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func notFound(message string) *AppError {
	return &AppError{Status: 404, Message: message}
}

func ErrorAs(err error, target **AppError) bool {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		return false
	}
	*target = appErr
	return true
}
