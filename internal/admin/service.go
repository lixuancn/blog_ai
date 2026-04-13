package admin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"example.com/laneblog/internal/store"
)

type AppError struct {
	Status  int               `json:"-"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

type Store interface {
	Read() store.Snapshot
	Write(func(*store.Snapshot) error) error
}

type Service struct {
	store Store
	now   func() time.Time
}

type CategoryInput struct {
	Name           string `json:"name"`
	SEOTitle       string `json:"seo_title"`
	SEODescription string `json:"seo_description"`
	SEOKeywords    string `json:"seo_keywords"`
	InOut          int    `json:"in_out"`
	PID            int64  `json:"pid"`
	URL            string `json:"url"`
	Item           string `json:"item"`
}

type CategoryView struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	SEOTitle       string `json:"seo_title"`
	SEODescription string `json:"seo_description"`
	SEOKeywords    string `json:"seo_keywords"`
	InOut          int    `json:"in_out"`
	PID            int64  `json:"pid"`
	URL            string `json:"url"`
	Item           string `json:"item"`
	ParentName     string `json:"parent_name"`
	Level          int    `json:"level"`
	HasChildren    bool   `json:"has_children"`
}

type ArticleInput struct {
	MID            int64  `json:"mid"`
	Author         string `json:"author"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	SEOTitle       string `json:"seo_title"`
	SEODescription string `json:"seo_description"`
	SEOKeywords    string `json:"seo_keywords"`
	Tag            string `json:"tag"`
	Clicks         int    `json:"clicks"`
	Content        string `json:"content"`
	CTime          int64  `json:"ctime"`
	GoodNum        int    `json:"good_num"`
	BadNum         int    `json:"bad_num"`
	RecommendType  int    `json:"recommend_type"`
}

type ArticleQuery struct {
	Title         string
	CategoryID    int64
	RecommendType int
	Page          int
	PageSize      int
}

type ArticleView struct {
	ID             int64  `json:"id"`
	MID            int64  `json:"mid"`
	CategoryName   string `json:"category_name"`
	Author         string `json:"author"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	SEOTitle       string `json:"seo_title"`
	SEODescription string `json:"seo_description"`
	SEOKeywords    string `json:"seo_keywords"`
	Tag            string `json:"tag"`
	Clicks         int    `json:"clicks"`
	Content        string `json:"content"`
	CTime          int64  `json:"ctime"`
	GoodNum        int    `json:"good_num"`
	BadNum         int    `json:"bad_num"`
	RecommendType  int    `json:"recommend_type"`
}

type ArticleListResult struct {
	Items      []ArticleView `json:"items"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	Total      int           `json:"total"`
	TotalPages int           `json:"total_pages"`
}

type TagView struct {
	ID  int64  `json:"id"`
	Tag string `json:"tag"`
	Num int    `json:"num"`
}

type DashboardSummary struct {
	ArticleCount         int `json:"article_count"`
	CategoryCount        int `json:"category_count"`
	TagCount             int `json:"tag_count"`
	RecommendedSiteCount int `json:"recommended_site_count"`
	RecommendedHomeCount int `json:"recommended_home_count"`
}

type DashboardData struct {
	Summary        DashboardSummary `json:"summary"`
	RecentArticles []ArticleView    `json:"recent_articles"`
	QuickLinks     []QuickLink      `json:"quick_links"`
}

type QuickLink struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
		now:   time.Now,
	}
}

func (s *Service) ListCategories(keyword string) []CategoryView {
	snapshot := s.store.Read()
	items := buildCategoryViews(snapshot, strings.TrimSpace(keyword))
	return items
}

func (s *Service) GetCategory(id int64) (CategoryView, error) {
	snapshot := s.store.Read()
	for _, item := range buildCategoryViews(snapshot, "") {
		if item.ID == id {
			return item, nil
		}
	}
	return CategoryView{}, notFound("分类不存在")
}

func (s *Service) CreateCategory(input CategoryInput) (CategoryView, error) {
	var created store.Category
	err := s.store.Write(func(snapshot *store.Snapshot) error {
		normalized, appErr := validateCategoryInput(snapshot, 0, input)
		if appErr != nil {
			return appErr
		}

		created = store.Category{
			ID:             snapshot.NextCategoryID,
			Name:           normalized.Name,
			SEOTitle:       normalized.SEOTitle,
			SEODescription: normalized.SEODescription,
			SEOKeywords:    normalized.SEOKeywords,
			InOut:          normalized.InOut,
			PID:            normalized.PID,
			URL:            normalized.URL,
			Item:           normalized.Item,
		}
		snapshot.NextCategoryID++
		snapshot.Categories = append(snapshot.Categories, created)
		return nil
	})
	if err != nil {
		return CategoryView{}, err
	}

	return s.GetCategory(created.ID)
}

func (s *Service) UpdateCategory(id int64, input CategoryInput) (CategoryView, error) {
	err := s.store.Write(func(snapshot *store.Snapshot) error {
		index := findCategoryIndex(snapshot.Categories, id)
		if index < 0 {
			return notFound("分类不存在")
		}

		normalized, appErr := validateCategoryInput(snapshot, id, input)
		if appErr != nil {
			return appErr
		}

		snapshot.Categories[index] = store.Category{
			ID:             id,
			Name:           normalized.Name,
			SEOTitle:       normalized.SEOTitle,
			SEODescription: normalized.SEODescription,
			SEOKeywords:    normalized.SEOKeywords,
			InOut:          normalized.InOut,
			PID:            normalized.PID,
			URL:            normalized.URL,
			Item:           normalized.Item,
		}
		return nil
	})
	if err != nil {
		return CategoryView{}, err
	}

	return s.GetCategory(id)
}

func (s *Service) DeleteCategory(id int64) error {
	return s.store.Write(func(snapshot *store.Snapshot) error {
		index := findCategoryIndex(snapshot.Categories, id)
		if index < 0 {
			return notFound("分类不存在")
		}

		for _, category := range snapshot.Categories {
			if category.PID == id {
				return conflict("当前分类存在子分类，需先处理子分类")
			}
		}
		for _, article := range snapshot.Articles {
			if article.MID == id {
				return conflict("当前分类下仍有关联文章")
			}
		}

		snapshot.Categories = append(snapshot.Categories[:index], snapshot.Categories[index+1:]...)
		return nil
	})
}

func (s *Service) ListArticles(query ArticleQuery) ArticleListResult {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	snapshot := s.store.Read()
	categories := make(map[int64]store.Category, len(snapshot.Categories))
	for _, category := range snapshot.Categories {
		categories[category.ID] = category
	}

	titleKeyword := strings.ToLower(strings.TrimSpace(query.Title))
	filtered := make([]store.Article, 0, len(snapshot.Articles))
	for _, article := range snapshot.Articles {
		if titleKeyword != "" && !strings.Contains(strings.ToLower(article.Title), titleKeyword) {
			continue
		}
		if query.CategoryID > 0 && article.MID != query.CategoryID {
			continue
		}
		if query.RecommendType > 0 && article.RecommendType != query.RecommendType {
			continue
		}
		filtered = append(filtered, article)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].CTime == filtered[j].CTime {
			return filtered[i].ID > filtered[j].ID
		}
		return filtered[i].CTime > filtered[j].CTime
	})

	total := len(filtered)
	totalPages := 0
	if total > 0 {
		totalPages = (total + query.PageSize - 1) / query.PageSize
	}

	start := (query.Page - 1) * query.PageSize
	if start > total {
		start = total
	}
	end := start + query.PageSize
	if end > total {
		end = total
	}

	items := make([]ArticleView, 0, end-start)
	for _, article := range filtered[start:end] {
		items = append(items, toArticleView(article, categories[article.MID]))
	}

	return ArticleListResult{
		Items:      items,
		Page:       query.Page,
		PageSize:   query.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

func (s *Service) GetArticle(id int64) (ArticleView, error) {
	snapshot := s.store.Read()
	categories := make(map[int64]store.Category, len(snapshot.Categories))
	for _, category := range snapshot.Categories {
		categories[category.ID] = category
	}

	for _, article := range snapshot.Articles {
		if article.ID == id {
			return toArticleView(article, categories[article.MID]), nil
		}
	}

	return ArticleView{}, notFound("文章不存在")
}

func (s *Service) CreateArticle(input ArticleInput) (ArticleView, error) {
	var id int64
	err := s.store.Write(func(snapshot *store.Snapshot) error {
		normalized, tags, appErr := validateArticleInput(snapshot, input, s.now().Unix())
		if appErr != nil {
			return appErr
		}

		id = snapshot.NextArticleID
		snapshot.NextArticleID++
		snapshot.Articles = append(snapshot.Articles, store.Article{
			ID:             id,
			MID:            normalized.MID,
			Author:         normalized.Author,
			Title:          normalized.Title,
			Description:    normalized.Description,
			SEOTitle:       normalized.SEOTitle,
			SEODescription: normalized.SEODescription,
			SEOKeywords:    normalized.SEOKeywords,
			Tag:            strings.Join(tags, ","),
			Clicks:         normalized.Clicks,
			Content:        normalized.Content,
			CTime:          normalized.CTime,
			GoodNum:        normalized.GoodNum,
			BadNum:         normalized.BadNum,
			RecommendType:  normalized.RecommendType,
		})
		rebuildTags(snapshot)
		return nil
	})
	if err != nil {
		return ArticleView{}, err
	}

	return s.GetArticle(id)
}

func (s *Service) UpdateArticle(id int64, input ArticleInput) (ArticleView, error) {
	err := s.store.Write(func(snapshot *store.Snapshot) error {
		index := findArticleIndex(snapshot.Articles, id)
		if index < 0 {
			return notFound("文章不存在")
		}

		normalized, tags, appErr := validateArticleInput(snapshot, input, snapshot.Articles[index].CTime)
		if appErr != nil {
			return appErr
		}

		snapshot.Articles[index] = store.Article{
			ID:             id,
			MID:            normalized.MID,
			Author:         normalized.Author,
			Title:          normalized.Title,
			Description:    normalized.Description,
			SEOTitle:       normalized.SEOTitle,
			SEODescription: normalized.SEODescription,
			SEOKeywords:    normalized.SEOKeywords,
			Tag:            strings.Join(tags, ","),
			Clicks:         normalized.Clicks,
			Content:        normalized.Content,
			CTime:          normalized.CTime,
			GoodNum:        normalized.GoodNum,
			BadNum:         normalized.BadNum,
			RecommendType:  normalized.RecommendType,
		}
		rebuildTags(snapshot)
		return nil
	})
	if err != nil {
		return ArticleView{}, err
	}

	return s.GetArticle(id)
}

func (s *Service) DeleteArticle(id int64) error {
	return s.store.Write(func(snapshot *store.Snapshot) error {
		index := findArticleIndex(snapshot.Articles, id)
		if index < 0 {
			return notFound("文章不存在")
		}

		snapshot.Articles = append(snapshot.Articles[:index], snapshot.Articles[index+1:]...)
		rebuildTags(snapshot)
		return nil
	})
}

func (s *Service) ListTags(keyword string) []TagView {
	snapshot := s.store.Read()
	items := make([]TagView, 0, len(snapshot.Tags))
	needle := strings.ToLower(strings.TrimSpace(keyword))

	for _, tag := range snapshot.Tags {
		if needle != "" && !strings.Contains(strings.ToLower(tag.Tag), needle) {
			continue
		}
		items = append(items, TagView{
			ID:  tag.ID,
			Tag: tag.Tag,
			Num: tag.Num,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Num == items[j].Num {
			return items[i].Tag < items[j].Tag
		}
		return items[i].Num > items[j].Num
	})

	return items
}

func (s *Service) DeleteTag(id int64) error {
	return s.store.Write(func(snapshot *store.Snapshot) error {
		index := -1
		for i, tag := range snapshot.Tags {
			if tag.ID == id {
				index = i
				break
			}
		}
		if index < 0 {
			return notFound("标签不存在")
		}
		if snapshot.Tags[index].Num > 0 {
			return conflict("当前标签已被文章使用，禁止删除")
		}
		snapshot.Tags = append(snapshot.Tags[:index], snapshot.Tags[index+1:]...)
		return nil
	})
}

func (s *Service) RebuildTags() []TagView {
	_ = s.store.Write(func(snapshot *store.Snapshot) error {
		rebuildTags(snapshot)
		return nil
	})
	return s.ListTags("")
}

func (s *Service) Dashboard() DashboardData {
	snapshot := s.store.Read()
	categories := make(map[int64]store.Category, len(snapshot.Categories))
	for _, category := range snapshot.Categories {
		categories[category.ID] = category
	}

	recent := append([]store.Article(nil), snapshot.Articles...)
	sort.Slice(recent, func(i, j int) bool {
		if recent[i].CTime == recent[j].CTime {
			return recent[i].ID > recent[j].ID
		}
		return recent[i].CTime > recent[j].CTime
	})
	if len(recent) > 5 {
		recent = recent[:5]
	}

	siteCount := 0
	homeCount := 0
	for _, article := range snapshot.Articles {
		if article.RecommendType == 1 {
			siteCount++
		}
		if article.RecommendType == 2 {
			homeCount++
		}
	}

	recentViews := make([]ArticleView, 0, len(recent))
	for _, article := range recent {
		recentViews = append(recentViews, toArticleView(article, categories[article.MID]))
	}

	return DashboardData{
		Summary: DashboardSummary{
			ArticleCount:         len(snapshot.Articles),
			CategoryCount:        len(snapshot.Categories),
			TagCount:             len(snapshot.Tags),
			RecommendedSiteCount: siteCount,
			RecommendedHomeCount: homeCount,
		},
		RecentArticles: recentViews,
		QuickLinks: []QuickLink{
			{Name: "新建文章", Path: "/admin/article/create"},
			{Name: "新建分类", Path: "/admin/category/create"},
			{Name: "查看标签", Path: "/admin/tag"},
		},
	}
}

func buildCategoryViews(snapshot store.Snapshot, keyword string) []CategoryView {
	categoryMap := make(map[int64]store.Category, len(snapshot.Categories))
	children := make(map[int64][]store.Category)

	for _, category := range snapshot.Categories {
		categoryMap[category.ID] = category
		children[category.PID] = append(children[category.PID], category)
	}

	for pid := range children {
		sort.Slice(children[pid], func(i, j int) bool {
			return children[pid][i].ID < children[pid][j].ID
		})
	}

	needle := strings.ToLower(keyword)
	result := make([]CategoryView, 0, len(snapshot.Categories))
	for _, top := range children[0] {
		views, ok := collectCategoryViews(top, categoryMap, children, 0, needle)
		if ok {
			result = append(result, views...)
		}
	}
	return result
}

func collectCategoryViews(category store.Category, categoryMap map[int64]store.Category, children map[int64][]store.Category, level int, keyword string) ([]CategoryView, bool) {
	matchedSelf := keyword == "" || strings.Contains(strings.ToLower(category.Name), keyword)
	childMatched := false
	collectedChildren := make([]CategoryView, 0, len(children[category.ID]))

	for _, child := range children[category.ID] {
		if childViews, ok := collectCategoryViews(child, categoryMap, children, level+1, keyword); ok {
			childMatched = true
			collectedChildren = append(collectedChildren, childViews...)
		}
	}

	if !matchedSelf && !childMatched {
		return nil, false
	}

	parentName := ""
	if category.PID > 0 {
		parentName = categoryMap[category.PID].Name
	}

	view := CategoryView{
		ID:             category.ID,
		Name:           category.Name,
		SEOTitle:       category.SEOTitle,
		SEODescription: category.SEODescription,
		SEOKeywords:    category.SEOKeywords,
		InOut:          category.InOut,
		PID:            category.PID,
		URL:            category.URL,
		Item:           category.Item,
		ParentName:     parentName,
		Level:          level,
		HasChildren:    len(children[category.ID]) > 0,
	}

	return append([]CategoryView{view}, collectedChildren...), true
}

func toArticleView(article store.Article, category store.Category) ArticleView {
	return ArticleView{
		ID:             article.ID,
		MID:            article.MID,
		CategoryName:   category.Name,
		Author:         article.Author,
		Title:          article.Title,
		Description:    article.Description,
		SEOTitle:       article.SEOTitle,
		SEODescription: article.SEODescription,
		SEOKeywords:    article.SEOKeywords,
		Tag:            article.Tag,
		Clicks:         article.Clicks,
		Content:        article.Content,
		CTime:          article.CTime,
		GoodNum:        article.GoodNum,
		BadNum:         article.BadNum,
		RecommendType:  article.RecommendType,
	}
}

func validateCategoryInput(snapshot *store.Snapshot, currentID int64, input CategoryInput) (CategoryInput, *AppError) {
	input.Name = strings.TrimSpace(input.Name)
	input.SEOTitle = strings.TrimSpace(input.SEOTitle)
	input.SEODescription = strings.TrimSpace(input.SEODescription)
	input.SEOKeywords = strings.TrimSpace(input.SEOKeywords)
	input.URL = strings.TrimSpace(input.URL)
	input.Item = strings.TrimSpace(input.Item)

	fields := make(map[string]string)
	requireMax(fields, "name", input.Name, 50)
	requireMax(fields, "seo_title", input.SEOTitle, 100)
	requireMax(fields, "seo_description", input.SEODescription, 500)
	requireMax(fields, "seo_keywords", input.SEOKeywords, 200)
	requireMax(fields, "item", input.Item, 20)

	if input.InOut != 1 && input.InOut != 2 {
		fields["in_out"] = "链接类型仅允许 1 或 2"
	}
	if input.PID < 0 {
		fields["pid"] = "父级分类不合法"
	}
	if input.PID == currentID && currentID != 0 {
		fields["pid"] = "父级分类不能为自身"
	}

	if input.InOut == 2 {
		if input.URL == "" {
			fields["url"] = "站外分类必须填写合法链接"
		} else if _, err := validateExternalURL(input.URL); err != nil {
			fields["url"] = err.Error()
		}
	} else if input.InOut == 1 {
		input.URL = ""
	}

	if input.PID > 0 {
		parent, ok := findCategory(snapshot.Categories, input.PID)
		if !ok {
			fields["pid"] = "父级分类不存在"
		} else if parent.PID != 0 {
			fields["pid"] = "当前阶段仅支持两级分类"
		}
	}

	if currentID > 0 && input.PID > 0 {
		parentID := input.PID
		for parentID > 0 {
			if parentID == currentID {
				fields["pid"] = "不允许形成循环父子关系"
				break
			}
			parent, ok := findCategory(snapshot.Categories, parentID)
			if !ok {
				break
			}
			parentID = parent.PID
		}
	}

	if len(fields) > 0 {
		return CategoryInput{}, badRequest("分类参数校验失败", fields)
	}
	return input, nil
}

func validateArticleInput(snapshot *store.Snapshot, input ArticleInput, fallbackCTime int64) (ArticleInput, []string, *AppError) {
	input.Author = strings.TrimSpace(input.Author)
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.SEOTitle = strings.TrimSpace(input.SEOTitle)
	input.SEODescription = strings.TrimSpace(input.SEODescription)
	input.SEOKeywords = strings.TrimSpace(input.SEOKeywords)
	input.Content = strings.TrimSpace(input.Content)

	fields := make(map[string]string)
	if input.MID <= 0 {
		fields["mid"] = "请选择有效分类"
	} else {
		category, ok := findCategory(snapshot.Categories, input.MID)
		if !ok {
			fields["mid"] = "分类不存在"
		} else if category.InOut != 1 {
			fields["mid"] = "站外分类不能用于文章归类"
		}
	}

	requireMax(fields, "author", input.Author, 50)
	requireMax(fields, "title", input.Title, 100)
	requireMax(fields, "description", input.Description, 500)
	requireMax(fields, "seo_title", input.SEOTitle, 100)
	requireMax(fields, "seo_description", input.SEODescription, 500)
	requireMax(fields, "seo_keywords", input.SEOKeywords, 200)
	if input.Content == "" {
		fields["content"] = "正文不能为空"
	}
	if input.Clicks < 0 {
		fields["clicks"] = "点击数不能小于 0"
	}
	if input.GoodNum < 0 {
		fields["good_num"] = "点赞数不能小于 0"
	}
	if input.BadNum < 0 {
		fields["bad_num"] = "点踩数不能小于 0"
	}
	if input.RecommendType != 1 && input.RecommendType != 2 {
		fields["recommend_type"] = "推荐类型仅允许 1 或 2"
	}

	tags := normalizeTags(input.Tag)
	if len(tags) == 0 {
		fields["tag"] = "标签不能为空"
	}
	if len(tags) > 10 {
		fields["tag"] = "单篇文章最多 10 个标签"
	}
	if input.CTime <= 0 {
		input.CTime = fallbackCTime
	}
	if input.CTime <= 0 {
		input.CTime = time.Now().Unix()
	}

	if len(fields) > 0 {
		return ArticleInput{}, nil, badRequest("文章参数校验失败", fields)
	}
	return input, tags, nil
}

func normalizeTags(raw string) []string {
	parts := strings.Split(raw, ",")
	seen := make(map[string]struct{}, len(parts))
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result
}

func rebuildTags(snapshot *store.Snapshot) {
	counts := make(map[string]int)
	for _, article := range snapshot.Articles {
		for _, tag := range normalizeTags(article.Tag) {
			counts[tag]++
		}
	}

	existingIDs := make(map[string]int64, len(snapshot.Tags))
	for _, tag := range snapshot.Tags {
		existingIDs[tag.Tag] = tag.ID
	}

	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)

	nextTags := make([]store.Tag, 0, len(names))
	for _, name := range names {
		id := existingIDs[name]
		if id == 0 {
			id = snapshot.NextTagID
			snapshot.NextTagID++
		}
		nextTags = append(nextTags, store.Tag{
			ID:  id,
			Tag: name,
			Num: counts[name],
		})
	}

	snapshot.Tags = nextTags
}

func validateExternalURL(raw string) (*url.URL, error) {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return nil, errors.New("站外分类必须填写合法链接")
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("站外分类必须填写合法链接")
	}
	return parsed, nil
}

func requireMax(fields map[string]string, field, value string, max int) {
	if value == "" {
		fields[field] = fmt.Sprintf("%s 不能为空", field)
		return
	}
	if len([]rune(value)) > max {
		fields[field] = fmt.Sprintf("%s 长度不能超过 %d", field, max)
	}
}

func findCategory(items []store.Category, id int64) (store.Category, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return store.Category{}, false
}

func findCategoryIndex(items []store.Category, id int64) int {
	for i, item := range items {
		if item.ID == id {
			return i
		}
	}
	return -1
}

func findArticleIndex(items []store.Article, id int64) int {
	for i, item := range items {
		if item.ID == id {
			return i
		}
	}
	return -1
}

func badRequest(message string, fields map[string]string) *AppError {
	return &AppError{
		Status:  http.StatusBadRequest,
		Message: message,
		Fields:  fields,
	}
}

func notFound(message string) *AppError {
	return &AppError{
		Status:  http.StatusNotFound,
		Message: message,
	}
}

func conflict(message string) *AppError {
	return &AppError{
		Status:  http.StatusConflict,
		Message: message,
	}
}
