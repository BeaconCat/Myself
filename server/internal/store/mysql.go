package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// DatabaseConfig is private runtime configuration, never public site settings.
type DatabaseConfig struct {
	Driver string      `json:"driver"`
	MySQL  MySQLConfig `json:"mysql,omitempty"`
}

type MySQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
	TLS      string `json:"tls"`
}

func OpenConfigured(dataDir string, config DatabaseConfig) (*DB, error) {
	switch config.Driver {
	case "", "sqlite":
		return Open(dataDir)
	case "mysql":
		return openMySQL(config.MySQL)
	default:
		return nil, errors.New("database driver must be sqlite or mysql")
	}
}

func openMySQL(config MySQLConfig) (*DB, error) {
	if config.Host == "" || config.Name == "" || config.User == "" {
		return nil, errors.New("MySQL host, database name and user are required")
	}
	if config.Port == 0 {
		config.Port = 3306
	}
	if config.Port < 1 || config.Port > 65535 {
		return nil, errors.New("invalid MySQL port")
	}
	if config.TLS != "" && config.TLS != "false" && config.TLS != "true" {
		return nil, errors.New("MySQL tls must be true or false")
	}
	cfg := mysql.NewConfig()
	cfg.User, cfg.Passwd, cfg.DBName = config.User, config.Password, config.Name
	cfg.Net, cfg.Addr = "tcp", net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	cfg.TLSConfig = config.TLS
	if err := cfg.Apply(mysql.Charset("utf8mb4", "utf8mb4_bin")); err != nil {
		return nil, err
	}
	cfg.Params = map[string]string{"time_zone": "'+00:00'", "sql_mode": "'STRICT_TRANS_TABLES,ONLY_FULL_GROUP_BY,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,NO_BACKSLASH_ESCAPES,PIPES_AS_CONCAT'"}
	cfg.ClientFoundRows = true // SQLite UPDATE counts matched rows, including unchanged values.
	cfg.Timeout, cfg.ReadTimeout, cfg.WriteTimeout = 10*time.Second, 2*time.Minute, 2*time.Minute
	connector, err := mysql.NewConnector(cfg)
	if err != nil {
		return nil, errors.New("invalid MySQL connection options")
	}
	raw := sql.OpenDB(connector)
	raw.SetMaxOpenConns(10)
	raw.SetMaxIdleConns(5)
	raw.SetConnMaxLifetime(3 * time.Minute)
	db := &DB{DB: raw, dialect: "mysql"}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := raw.PingContext(ctx); err != nil {
		raw.Close()
		return nil, fmt.Errorf("connect MySQL: %w", err)
	}
	if err := db.migrateMySQL(); err != nil {
		raw.Close()
		return nil, fmt.Errorf("migrate MySQL: %w", err)
	}
	return db, nil
}

func (db *DB) Driver() string {
	if db.dialect == "mysql" {
		return "mysql"
	}
	return "sqlite"
}

// IgnoreUpdate is used when merging virtual media folders with an existing path.
func (db *DB) IgnoreUpdate(query string) string {
	if db.Driver() == "mysql" {
		return strings.Replace(query, "UPDATE OR IGNORE", "UPDATE IGNORE", 1)
	}
	return query
}

// Upsert handles only the dialect-specific conflict clause. Identifiers come from code.
func (db *DB) Upsert(table string, columns, keys, updates []string, args ...any) (sql.Result, error) {
	quote := func(s string) string { return "`" + strings.ReplaceAll(s, "`", "``") + "`" }
	quoted := func(names []string) string {
		out := make([]string, len(names))
		for i, name := range names {
			out[i] = quote(name)
		}
		return strings.Join(out, ", ")
	}
	query := "INSERT INTO " + quote(table) + " (" + quoted(columns) + ") VALUES (" + strings.TrimSuffix(strings.Repeat("?,", len(columns)), ",") + ")"
	var set []string
	if db.Driver() == "mysql" {
		query += " AS incoming ON DUPLICATE KEY UPDATE "
		for _, c := range updates {
			set = append(set, quote(c)+" = incoming."+quote(c))
		}
		if len(set) == 0 {
			set = append(set, quote(keys[0])+" = "+quote(table)+"."+quote(keys[0]))
		}
	} else {
		query += " ON CONFLICT (" + quoted(keys) + ") DO "
		if len(updates) == 0 {
			return db.Exec(query+"NOTHING", args...)
		}
		query += "UPDATE SET "
		for _, c := range updates {
			set = append(set, quote(c)+" = excluded."+quote(c))
		}
	}
	return db.Exec(query+strings.Join(set, ", "), args...)
}

func (db *DB) migrateMySQL() error {
	// The schema is additive. MySQL tables use InnoDB for transactional restores.
	for _, statement := range strings.Split(mysqlSchema, ";") {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := db.Exec(statement + " ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin"); err != nil {
			return err
		}
	}
	return nil
}

const mysqlSchema = `
CREATE TABLE IF NOT EXISTS posts (
 id BIGINT PRIMARY KEY AUTO_INCREMENT, slug VARCHAR(80) NOT NULL UNIQUE,
 title TEXT NOT NULL, excerpt TEXT NOT NULL DEFAULT (''), content_md LONGTEXT NOT NULL,
 covers LONGTEXT NOT NULL DEFAULT ('[]'), tags LONGTEXT NOT NULL DEFAULT ('[]'),
 status VARCHAR(20) NOT NULL DEFAULT 'published', pinned BIGINT NOT NULL DEFAULT 0,
 source_key BIGINT, author_id BIGINT, hidden BIGINT NOT NULL DEFAULT 0,
 created_at VARCHAR(32) NOT NULL DEFAULT (CURRENT_TIMESTAMP), updated_at VARCHAR(32) NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
CREATE TABLE IF NOT EXISTS notes (
 id BIGINT PRIMARY KEY AUTO_INCREMENT, content_md LONGTEXT NOT NULL, mood TEXT NOT NULL DEFAULT (''),
 images LONGTEXT NOT NULL DEFAULT ('[]'), pinned BIGINT NOT NULL DEFAULT 0, source_key BIGINT,
 hidden BIGINT NOT NULL DEFAULT 0, created_at VARCHAR(32) NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
CREATE TABLE IF NOT EXISTS settings (
 ` + "`key`" + ` VARCHAR(191) PRIMARY KEY, value LONGTEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS api_keys (
 id BIGINT PRIMARY KEY AUTO_INCREMENT, name TEXT NOT NULL, key_hash VARCHAR(64) NOT NULL UNIQUE,
 prefix VARCHAR(64) NOT NULL, scope VARCHAR(20) NOT NULL DEFAULT 'full',
 last_used_at VARCHAR(32), created_at VARCHAR(32) NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
CREATE TABLE IF NOT EXISTS media (
 name VARCHAR(255) PRIMARY KEY, crop_json LONGTEXT, sha256 VARCHAR(64),
 title TEXT NOT NULL DEFAULT (''), folder VARCHAR(512) NOT NULL DEFAULT '',
 compressed_from VARCHAR(255) NOT NULL DEFAULT '', compressed_before BIGINT NOT NULL DEFAULT 0,
 compressed_at VARCHAR(32), created_at VARCHAR(32) NOT NULL DEFAULT (CURRENT_TIMESTAMP),
 INDEX idx_media_sha (sha256)
);
CREATE TABLE IF NOT EXISTS media_folders (
 path VARCHAR(512) PRIMARY KEY, created_at VARCHAR(32) NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
CREATE TABLE IF NOT EXISTS users (
 id BIGINT PRIMARY KEY AUTO_INCREMENT,
 login VARCHAR(191) COLLATE utf8mb4_0900_as_ci NOT NULL UNIQUE,
 email VARCHAR(254) COLLATE utf8mb4_0900_as_ci UNIQUE,
 name TEXT NOT NULL DEFAULT (''), avatar TEXT NOT NULL DEFAULT (''), avatar_pending TEXT NOT NULL DEFAULT (''),
 role VARCHAR(20) NOT NULL DEFAULT 'reader', status VARCHAR(20) NOT NULL DEFAULT 'active',
 password_hash TEXT NOT NULL DEFAULT (''), github_id BIGINT UNIQUE,
 email_verified BIGINT NOT NULL DEFAULT 0, token_version BIGINT NOT NULL DEFAULT 0, must_change BIGINT NOT NULL DEFAULT 0,
 created_at VARCHAR(32) NOT NULL DEFAULT (CURRENT_TIMESTAMP), last_active_at VARCHAR(32), login_changed_at VARCHAR(32)
);
CREATE TABLE IF NOT EXISTS user_tokens (
 id BIGINT PRIMARY KEY AUTO_INCREMENT, kind VARCHAR(32) NOT NULL, token_hash VARCHAR(64) NOT NULL UNIQUE,
 user_id BIGINT, role VARCHAR(20) NOT NULL DEFAULT '', email VARCHAR(254) NOT NULL DEFAULT '',
 note TEXT NOT NULL DEFAULT (''), expires_at VARCHAR(32) NOT NULL, used_at VARCHAR(32), used_by BIGINT,
 created_at VARCHAR(32) NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
CREATE TABLE IF NOT EXISTS comments (
 id BIGINT PRIMARY KEY AUTO_INCREMENT, target VARCHAR(20) NOT NULL, target_id BIGINT NOT NULL DEFAULT 0,
 parent_id BIGINT, user_id BIGINT, guest_name TEXT NOT NULL DEFAULT (''), body LONGTEXT NOT NULL,
 status VARCHAR(20) NOT NULL DEFAULT 'pending', ip_hash VARCHAR(64) NOT NULL DEFAULT '',
 created_at VARCHAR(32) NOT NULL DEFAULT (CURRENT_TIMESTAMP),
 INDEX idx_comments_target (target, target_id, status), INDEX idx_comments_user (user_id)
);
CREATE TABLE IF NOT EXISTS reactions (
 id BIGINT PRIMARY KEY AUTO_INCREMENT, target VARCHAR(20) NOT NULL, target_id BIGINT NOT NULL,
 kind VARCHAR(32) NOT NULL, voter VARCHAR(128) NOT NULL, created_at VARCHAR(32) NOT NULL DEFAULT (CURRENT_TIMESTAMP),
 UNIQUE (target, target_id, kind, voter)
);
CREATE TABLE IF NOT EXISTS api_logs (
 id BIGINT PRIMARY KEY AUTO_INCREMENT, created_at VARCHAR(32) NOT NULL DEFAULT (CURRENT_TIMESTAMP),
 key_id BIGINT, key_name TEXT NOT NULL DEFAULT (''), key_prefix VARCHAR(64) NOT NULL DEFAULT '',
 method VARCHAR(16) NOT NULL, path TEXT NOT NULL, status BIGINT NOT NULL,
 duration_ms BIGINT NOT NULL DEFAULT 0, ip VARCHAR(64) NOT NULL DEFAULT '', user_agent TEXT NOT NULL DEFAULT (''),
 INDEX idx_api_logs_key (key_id, id)
);`
