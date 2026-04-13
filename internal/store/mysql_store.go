package store

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLStore struct {
	mu sync.Mutex
	db *sql.DB
}

func NewMySQLStore(dsn string) (*MySQLStore, error) {
	db, err := sql.Open("mysql", sanitizeMySQLDSN(dsn))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &MySQLStore{db: db}, nil
}

func (s *MySQLStore) Read() Snapshot {
	snapshot, err := s.readSnapshot()
	if err != nil {
		panic(fmt.Errorf("mysql store read failed: %w", err))
	}
	return snapshot
}

func (s *MySQLStore) Write(fn func(*Snapshot) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, err := s.readSnapshot()
	if err != nil {
		return err
	}

	next := CloneSnapshot(current)
	if err := fn(&next); err != nil {
		return err
	}
	NormalizeSnapshot(&next)
	return s.persist(next)
}

func (s *MySQLStore) readSnapshot() (Snapshot, error) {
	categories, err := s.readCategories()
	if err != nil {
		return Snapshot{}, err
	}
	articles, err := s.readArticles()
	if err != nil {
		return Snapshot{}, err
	}
	tags, err := s.readTags()
	if err != nil {
		return Snapshot{}, err
	}

	snapshot := Snapshot{
		Categories: categories,
		Articles:   articles,
		Tags:       tags,
	}
	NormalizeSnapshot(&snapshot)
	return snapshot, nil
}

func (s *MySQLStore) readCategories() ([]Category, error) {
	rows, err := s.db.Query(`
		SELECT id, name, seo_title, seo_description, seo_keywords, in_out, pid, url, item
		FROM info_menu
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Category, 0)
	for rows.Next() {
		var item Category
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.SEOTitle,
			&item.SEODescription,
			&item.SEOKeywords,
			&item.InOut,
			&item.PID,
			&item.URL,
			&item.Item,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *MySQLStore) readArticles() ([]Article, error) {
	rows, err := s.db.Query(`
		SELECT id, mid, author, title, description, seo_title, seo_description, seo_keywords,
		       tag, clicks, content, ctime, good_num, bad_num, recommend_type
		FROM info_article
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Article, 0)
	for rows.Next() {
		var item Article
		if err := rows.Scan(
			&item.ID,
			&item.MID,
			&item.Author,
			&item.Title,
			&item.Description,
			&item.SEOTitle,
			&item.SEODescription,
			&item.SEOKeywords,
			&item.Tag,
			&item.Clicks,
			&item.Content,
			&item.CTime,
			&item.GoodNum,
			&item.BadNum,
			&item.RecommendType,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *MySQLStore) readTags() ([]Tag, error) {
	rows, err := s.db.Query(`
		SELECT id, tag, num
		FROM info_tag
		ORDER BY num DESC, tag ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Tag, 0)
	for rows.Next() {
		var item Tag
		if err := rows.Scan(&item.ID, &item.Tag, &item.Num); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *MySQLStore) persist(snapshot Snapshot) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = replaceCategories(tx, snapshot.Categories); err != nil {
		return err
	}
	if err = replaceArticles(tx, snapshot.Articles); err != nil {
		return err
	}
	if err = replaceTags(tx, snapshot.Tags); err != nil {
		return err
	}

	return tx.Commit()
}

func replaceCategories(tx *sql.Tx, items []Category) error {
	if _, err := tx.Exec(`DELETE FROM info_menu`); err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(`
		INSERT INTO info_menu (id, name, seo_title, seo_description, seo_keywords, in_out, pid, url, item)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	for _, item := range items {
		if _, err := stmt.Exec(
			item.ID, item.Name, item.SEOTitle, item.SEODescription, item.SEOKeywords,
			item.InOut, item.PID, item.URL, item.Item,
		); err != nil {
			return err
		}
	}
	return resetAutoIncrement(tx, "info_menu", nextIDFromCategories(items))
}

func replaceArticles(tx *sql.Tx, items []Article) error {
	if _, err := tx.Exec(`DELETE FROM info_article`); err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(`
		INSERT INTO info_article (
			id, mid, author, title, description, seo_title, seo_description, seo_keywords,
			tag, clicks, content, ctime, good_num, bad_num, recommend_type
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	for _, item := range items {
		if _, err := stmt.Exec(
			item.ID, item.MID, item.Author, item.Title, item.Description, item.SEOTitle,
			item.SEODescription, item.SEOKeywords, item.Tag, item.Clicks, item.Content,
			item.CTime, item.GoodNum, item.BadNum, item.RecommendType,
		); err != nil {
			return err
		}
	}
	return resetAutoIncrement(tx, "info_article", nextIDFromArticles(items))
}

func replaceTags(tx *sql.Tx, items []Tag) error {
	if _, err := tx.Exec(`DELETE FROM info_tag`); err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(`
		INSERT INTO info_tag (id, tag, num)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	for _, item := range items {
		if _, err := stmt.Exec(item.ID, item.Tag, item.Num); err != nil {
			return err
		}
	}
	return resetAutoIncrement(tx, "info_tag", nextIDFromTags(items))
}

func resetAutoIncrement(tx *sql.Tx, table string, nextID int64) error {
	if nextID <= 0 {
		nextID = 1
	}
	query := fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = %d", table, nextID)
	_, err := tx.Exec(query)
	return err
}

func nextIDFromCategories(items []Category) int64 {
	var maxID int64
	for _, item := range items {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func nextIDFromArticles(items []Article) int64 {
	var maxID int64
	for _, item := range items {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func nextIDFromTags(items []Tag) int64 {
	var maxID int64
	for _, item := range items {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func (s *MySQLStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func sanitizeMySQLDSN(dsn string) string {
	return strings.TrimSpace(dsn)
}
