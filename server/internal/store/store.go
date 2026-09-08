// Package store 封装 SQLite 连接、建表迁移与行模型转换。
package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// DB 是全局数据库句柄的薄封装。
type DB struct {
	*sql.DB
}

// Open 打开（或创建）data/myself.db 并执行幂等建表迁移。
func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)",
		filepath.ToSlash(filepath.Join(dataDir, "myself.db")))
	raw, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite 单写者：限制单连接避免 database is locked。
	raw.SetMaxOpenConns(1)
	db := &DB{raw}
	if err := db.migrate(); err != nil {
		raw.Close()
		return nil, err
	}
	return db, nil
}

func (db *DB) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS posts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  slug TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  excerpt TEXT NOT NULL DEFAULT '',
  content_md TEXT NOT NULL,
  covers TEXT NOT NULL DEFAULT '[]',
  tags TEXT NOT NULL DEFAULT '[]',
  status TEXT NOT NULL DEFAULT 'published',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS notes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  content_md TEXT NOT NULL,
  mood TEXT NOT NULL DEFAULT '',
  images TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS api_keys (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  key_hash TEXT NOT NULL UNIQUE,
  prefix TEXT NOT NULL,
  last_used_at TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS media (
  name TEXT PRIMARY KEY,
  crop_json TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);`
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	// 旧库缺列时补齐（列已存在则忽略报错）。
	for _, stmt := range []string{
		`ALTER TABLE notes ADD COLUMN images TEXT NOT NULL DEFAULT '[]'`,
		`ALTER TABLE posts ADD COLUMN pinned INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE notes ADD COLUMN pinned INTEGER NOT NULL DEFAULT 0`,
	} {
		if _, err := db.Exec(stmt); err != nil && !strings.Contains(err.Error(), "duplicate column") {
			return err
		}
	}
	return nil
}

// GetSetting 读取 settings 表单值；不存在返回空串。
func (db *DB) GetSetting(key string) (string, error) {
	var v string
	err := db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&v)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return v, err
}

// SetSetting 写入或覆盖 settings 表单值。
func (db *DB) SetSetting(key, value string) error {
	_, err := db.Exec(`INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

// Post 是文章的 API 表示。
type Post struct {
	ID        int64    `json:"id"`
	Slug      string   `json:"slug"`
	Title     string   `json:"title"`
	Excerpt   string   `json:"excerpt"`
	Covers    []string `json:"covers"`
	Tags      []string `json:"tags"`
	Pinned    bool     `json:"pinned"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
	ContentMd *string  `json:"contentMd,omitempty"`
	Status    string   `json:"status,omitempty"`
}

// PostRow 是 posts 表的一行。
type PostRow struct {
	ID        int64
	Slug      string
	Title     string
	Excerpt   string
	ContentMd string
	Covers    string
	Tags      string
	Status    string
	Pinned    int64
	CreatedAt string
	UpdatedAt string
}

const postColumns = `id, slug, title, excerpt, content_md, covers, tags, status, pinned, created_at, updated_at`

// PostOpts 控制行 → API 对象的字段。
type PostOpts struct {
	WithContent bool
	WithStatus  bool
}

// ToPost 行 → API 对象。
func (r PostRow) ToPost(o PostOpts) Post {
	p := Post{
		ID:        r.ID,
		Slug:      r.Slug,
		Title:     r.Title,
		Excerpt:   r.Excerpt,
		Covers:    ParseStrings(r.Covers),
		Tags:      ParseStrings(r.Tags),
		Pinned:    r.Pinned != 0,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if o.WithContent {
		md := r.ContentMd
		p.ContentMd = &md
	}
	if o.WithStatus {
		p.Status = r.Status
	}
	return p
}

type scanner interface {
	Scan(dest ...any) error
}

func scanPost(s scanner) (PostRow, error) {
	var r PostRow
	err := s.Scan(&r.ID, &r.Slug, &r.Title, &r.Excerpt, &r.ContentMd, &r.Covers, &r.Tags,
		&r.Status, &r.Pinned, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

// QueryPosts 执行 posts 查询（where/order/limit 由调用方拼接，参数占位）。
func (db *DB) QueryPosts(tail string, args ...any) ([]PostRow, error) {
	rows, err := db.Query(`SELECT `+postColumns+` FROM posts `+tail, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PostRow
	for rows.Next() {
		r, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetPost 按条件取单篇；不存在返回 (nil, nil)。
func (db *DB) GetPost(tail string, args ...any) (*PostRow, error) {
	r, err := scanPost(db.QueryRow(`SELECT `+postColumns+` FROM posts `+tail, args...))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// Note 是随想的 API 表示。
type Note struct {
	ID        int64    `json:"id"`
	ContentMd string   `json:"contentMd"`
	Mood      string   `json:"mood"`
	Images    []string `json:"images"`
	Pinned    *bool    `json:"pinned,omitempty"`
	CreatedAt string   `json:"createdAt"`
}

// QueryNotes 执行 notes 查询；withPinned 控制是否输出 pinned 字段。
func (db *DB) QueryNotes(tail string, withPinned bool, args ...any) ([]Note, error) {
	rows, err := db.Query(`SELECT id, content_md, mood, images, pinned, created_at FROM notes `+tail, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Note
	for rows.Next() {
		var n Note
		var images string
		var pinned int64
		if err := rows.Scan(&n.ID, &n.ContentMd, &n.Mood, &images, &pinned, &n.CreatedAt); err != nil {
			return nil, err
		}
		n.Images = ParseStrings(images)
		if withPinned {
			b := pinned != 0
			n.Pinned = &b
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// ParseStrings 解析 JSON 字符串数组；失败返回空数组。
func ParseStrings(raw string) []string {
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil || out == nil {
		return []string{}
	}
	return out
}

// JSONStrings 序列化字符串数组（nil → "[]"）。
func JSONStrings(v []string) string {
	if v == nil {
		v = []string{}
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// IsUniqueErr 判断是否唯一约束冲突。
func IsUniqueErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE")
}
