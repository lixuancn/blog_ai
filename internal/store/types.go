package store

type Category struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	SEOTitle       string `json:"seo_title"`
	SEODescription string `json:"seo_description"`
	SEOKeywords    string `json:"seo_keywords"`
	InOut          int    `json:"in_out"`
	PID            int64  `json:"pid"`
	URL            string `json:"url"`
	Item           string `json:"item"`
}

type Article struct {
	ID             int64  `json:"id"`
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

type Tag struct {
	ID  int64  `json:"id"`
	Tag string `json:"tag"`
	Num int    `json:"num"`
}

type Snapshot struct {
	NextCategoryID int64      `json:"next_category_id"`
	NextArticleID  int64      `json:"next_article_id"`
	NextTagID      int64      `json:"next_tag_id"`
	Categories     []Category `json:"categories"`
	Articles       []Article  `json:"articles"`
	Tags           []Tag      `json:"tags"`
}

func NormalizeSnapshot(snapshot *Snapshot) {
	var maxCategoryID int64
	var maxArticleID int64
	var maxTagID int64

	for _, category := range snapshot.Categories {
		if category.ID > maxCategoryID {
			maxCategoryID = category.ID
		}
	}

	for _, article := range snapshot.Articles {
		if article.ID > maxArticleID {
			maxArticleID = article.ID
		}
	}

	for _, tag := range snapshot.Tags {
		if tag.ID > maxTagID {
			maxTagID = tag.ID
		}
	}

	if snapshot.NextCategoryID <= maxCategoryID {
		snapshot.NextCategoryID = maxCategoryID + 1
	}
	if snapshot.NextArticleID <= maxArticleID {
		snapshot.NextArticleID = maxArticleID + 1
	}
	if snapshot.NextTagID <= maxTagID {
		snapshot.NextTagID = maxTagID + 1
	}
	if snapshot.NextCategoryID == 0 {
		snapshot.NextCategoryID = 1
	}
	if snapshot.NextArticleID == 0 {
		snapshot.NextArticleID = 1
	}
	if snapshot.NextTagID == 0 {
		snapshot.NextTagID = 1
	}
}

func CloneSnapshot(snapshot Snapshot) Snapshot {
	cloned := Snapshot{
		NextCategoryID: snapshot.NextCategoryID,
		NextArticleID:  snapshot.NextArticleID,
		NextTagID:      snapshot.NextTagID,
	}

	if len(snapshot.Categories) > 0 {
		cloned.Categories = append([]Category(nil), snapshot.Categories...)
	}
	if len(snapshot.Articles) > 0 {
		cloned.Articles = append([]Article(nil), snapshot.Articles...)
	}
	if len(snapshot.Tags) > 0 {
		cloned.Tags = append([]Tag(nil), snapshot.Tags...)
	}

	return cloned
}
