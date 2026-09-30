package store

import (
	"database/sql"
	"strconv"
)

// API 调用日志：外部通道（X-Api-Key）每次请求一行，含无效 / 缺失 Key 的尝试。
// 只记元数据（时间、Key、方法、路径、状态、耗时、IP、UA），从不记录 Key 明文与请求体。

const apiLogsSchema = `
CREATE TABLE IF NOT EXISTS api_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  key_id INTEGER,
  key_name TEXT NOT NULL DEFAULT '',
  key_prefix TEXT NOT NULL DEFAULT '',
  method TEXT NOT NULL,
  path TEXT NOT NULL,
  status INTEGER NOT NULL,
  duration_ms INTEGER NOT NULL DEFAULT 0,
  ip TEXT NOT NULL DEFAULT '',
  user_agent TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_api_logs_key ON api_logs(key_id, id);
`

// 日志保留：最多 APILogKeep 行，且不超过 APILogDays 天。
const (
	APILogKeep = 5000
	APILogDays = 90
	// apiLogPruneEvery 每写入这么多行顺带清理一次，摊薄开销
	apiLogPruneEvery = 50
)

// APILog 一条调用记录（KeyID 为空表示 Key 缺失或无效）。
type APILog struct {
	ID        int64  `json:"id"`
	At        string `json:"at"`
	KeyID     *int64 `json:"keyId"`
	KeyName   string `json:"keyName"`
	KeyPrefix string `json:"keyPrefix"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Ms        int64  `json:"ms"`
	IP        string `json:"ip"`
	UA        string `json:"ua"`
}

// InsertAPILog 写入一条调用记录，并按行数顺带清理过期 / 超额的旧记录。
func (db *DB) InsertAPILog(l APILog) error {
	var keyID any
	if l.KeyID != nil {
		keyID = *l.KeyID
	}
	res, err := db.Exec(`INSERT INTO api_logs (key_id, key_name, key_prefix, method, path, status, duration_ms, ip, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		keyID, l.KeyName, l.KeyPrefix, l.Method, l.Path, l.Status, l.Ms, l.IP, l.UA)
	if err != nil {
		return err
	}
	if id, err := res.LastInsertId(); err == nil && id%apiLogPruneEvery == 0 {
		return db.PruneAPILogs(id)
	}
	return nil
}

// PruneAPILogs 删除 id 落在最近 APILogKeep 行之外、或早于 APILogDays 天的记录。
func (db *DB) PruneAPILogs(lastID int64) error {
	_, err := db.Exec(`DELETE FROM api_logs WHERE id <= ? OR created_at < datetime('now', ?)`,
		lastID-APILogKeep, "-"+strconv.Itoa(APILogDays)+" days")
	return err
}

// QueryAPILogs 按条件（where 以 WHERE 开头或为空）分页取记录，新到旧。
func (db *DB) QueryAPILogs(where string, args []any, limit, offset int) ([]APILog, error) {
	rows, err := db.Query(`SELECT id, created_at, key_id, key_name, key_prefix, method, path, status, duration_ms, ip, user_agent
		FROM api_logs `+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []APILog{}
	for rows.Next() {
		var l APILog
		var keyID sql.NullInt64
		if err := rows.Scan(&l.ID, &l.At, &keyID, &l.KeyName, &l.KeyPrefix, &l.Method, &l.Path, &l.Status, &l.Ms, &l.IP, &l.UA); err != nil {
			return nil, err
		}
		if keyID.Valid {
			id := keyID.Int64
			l.KeyID = &id
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// APIKeyUsage 各 Key 在保留窗口内的调用数与失败数。
type APIKeyUsage struct {
	Calls  int
	Errors int
}

// APILogUsage 按 Key 汇总调用数。
func (db *DB) APILogUsage() (map[int64]APIKeyUsage, error) {
	rows, err := db.Query(`SELECT key_id, COUNT(*), SUM(status >= 400) FROM api_logs WHERE key_id IS NOT NULL GROUP BY key_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]APIKeyUsage{}
	for rows.Next() {
		var id int64
		var u APIKeyUsage
		if err := rows.Scan(&id, &u.Calls, &u.Errors); err != nil {
			return nil, err
		}
		out[id] = u
	}
	return out, rows.Err()
}
