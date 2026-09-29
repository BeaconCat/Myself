package store

import (
	"database/sql"
	"errors"
	"strings"
)

// 用户系统：users（管理员 / 作者 / 读者）+ user_tokens（邮箱验证、重置密码、邀请的一次性令牌）+ comments。
// 管理员也是 users 里的一行（role=admin）；历史版本存在 settings 里的单管理员由 auth.Init 迁移进来。

const usersSchema = `
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  login TEXT NOT NULL UNIQUE COLLATE NOCASE,
  email TEXT UNIQUE COLLATE NOCASE,
  name TEXT NOT NULL DEFAULT '',
  avatar TEXT NOT NULL DEFAULT '',
  role TEXT NOT NULL DEFAULT 'reader',
  status TEXT NOT NULL DEFAULT 'active',
  password_hash TEXT NOT NULL DEFAULT '',
  github_id INTEGER UNIQUE,
  email_verified INTEGER NOT NULL DEFAULT 0,
  token_version INTEGER NOT NULL DEFAULT 0,
  must_change INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  last_active_at TEXT
);
CREATE TABLE IF NOT EXISTS user_tokens (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  kind TEXT NOT NULL,
  token_hash TEXT NOT NULL UNIQUE,
  user_id INTEGER,
  role TEXT NOT NULL DEFAULT '',
  email TEXT NOT NULL DEFAULT '',
  note TEXT NOT NULL DEFAULT '',
  expires_at TEXT NOT NULL,
  used_at TEXT,
  used_by INTEGER,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS comments (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  target TEXT NOT NULL,
  target_id INTEGER NOT NULL DEFAULT 0,
  parent_id INTEGER,
  user_id INTEGER,
  guest_name TEXT NOT NULL DEFAULT '',
  body TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  ip_hash TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_comments_target ON comments(target, target_id, status);
CREATE INDEX IF NOT EXISTS idx_comments_user ON comments(user_id);
`

// 角色与状态。
const (
	RoleAdmin  = "admin"
	RoleAuthor = "author"
	RoleReader = "reader"

	StatusActive   = "active"
	StatusPending  = "pending"  // 待邮箱验证
	StatusDisabled = "disabled" // 管理员停用
)

// User 用户行（含口令哈希，勿直接下发）。
type User struct {
	ID            int64
	Login         string
	Email         string
	Name          string
	Avatar        string
	Role          string
	Status        string
	PasswordHash  string
	GitHubID      int64
	EmailVerified bool
	TokenVersion  int
	MustChange    bool
	CreatedAt     string
	LastActiveAt  string
}

const userColumns = `id, login, COALESCE(email, ''), name, avatar, role, status, password_hash, COALESCE(github_id, 0),
	email_verified, token_version, must_change, created_at, COALESCE(last_active_at, '')`

func scanUser(s scanner) (*User, error) {
	var u User
	err := s.Scan(&u.ID, &u.Login, &u.Email, &u.Name, &u.Avatar, &u.Role, &u.Status, &u.PasswordHash, &u.GitHubID,
		&u.EmailVerified, &u.TokenVersion, &u.MustChange, &u.CreatedAt, &u.LastActiveAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UserBy 按条件取一个用户；不存在返回 nil。where 为常量 SQL 片段。
func (db *DB) UserBy(where string, args ...any) (*User, error) {
	return scanUser(db.QueryRow(`SELECT `+userColumns+` FROM users WHERE `+where+` LIMIT 1`, args...))
}

// UserByID 按 id 取用户。
func (db *DB) UserByID(id int64) (*User, error) { return db.UserBy(`id = ?`, id) }

// UserByLogin 按登录名或邮箱（不区分大小写）取用户。
func (db *DB) UserByLogin(identifier string) (*User, error) {
	id := strings.TrimSpace(identifier)
	return db.UserBy(`login = ? OR (email IS NOT NULL AND email = ?)`, id, id)
}

// ListUsers 列出用户（管理员视图），附评论数。
func (db *DB) ListUsers() ([]User, map[int64]int, error) {
	rows, err := db.Query(`SELECT ` + userColumns + ` FROM users ORDER BY CASE role WHEN 'admin' THEN 0 WHEN 'author' THEN 1 ELSE 2 END, id`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, *u)
	}
	counts := map[int64]int{}
	crows, err := db.Query(`SELECT user_id, COUNT(*) FROM comments WHERE user_id IS NOT NULL GROUP BY user_id`)
	if err != nil {
		return nil, nil, err
	}
	defer crows.Close()
	for crows.Next() {
		var id int64
		var n int
		if err := crows.Scan(&id, &n); err != nil {
			return nil, nil, err
		}
		counts[id] = n
	}
	return out, counts, rows.Err()
}

// NewUser 创建用户的字段。
type NewUser struct {
	Login, Email, Name, Avatar, Role, Status, PasswordHash string
	GitHubID                                                int64
	EmailVerified                                           bool
}

// CreateUser 插入用户，返回 id；登录名 / 邮箱 / GitHub 冲突返回 IsUniqueErr 可识别的错误。
func (db *DB) CreateUser(n NewUser) (int64, error) {
	var email, gh any
	if n.Email != "" {
		email = n.Email
	}
	if n.GitHubID != 0 {
		gh = n.GitHubID
	}
	res, err := db.Exec(`INSERT INTO users (login, email, name, avatar, role, status, password_hash, github_id, email_verified)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.Login, email, n.Name, n.Avatar, n.Role, n.Status, n.PasswordHash, gh, n.EmailVerified)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// CountUsers 按角色计数（role 为空则全部）。
func (db *DB) CountUsers(role string) (int, error) {
	var n int
	var err error
	if role == "" {
		err = db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	} else {
		err = db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = ?`, role).Scan(&n)
	}
	return n, err
}

// TouchUser 记录最近活跃（5 分钟内不重复写）。
func (db *DB) TouchUser(id int64) {
	_, _ = db.Exec(`UPDATE users SET last_active_at = datetime('now')
		WHERE id = ? AND (last_active_at IS NULL OR last_active_at < datetime('now', '-5 minutes'))`, id)
}

// AttachAuthors 为文章填上协作作者署名（管理员写的不署名，前台用站点身份）。
func (db *DB) AttachAuthors(posts []Post) error {
	ids := map[int64]bool{}
	for _, p := range posts {
		if p.AuthorID > 0 {
			ids[p.AuthorID] = true
		}
	}
	if len(ids) == 0 {
		return nil
	}
	authors := map[int64]*PostAuthor{}
	for id := range ids {
		u, err := db.UserByID(id)
		if err != nil {
			return err
		}
		if u != nil && u.Role != RoleAdmin {
			authors[id] = &PostAuthor{ID: u.ID, Name: u.Name, Avatar: u.Avatar}
		}
	}
	for i := range posts {
		posts[i].Author = authors[posts[i].AuthorID]
	}
	return nil
}
