package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// ApplicationTables is the restore allowlist; unrelated MySQL tables are never touched.
var ApplicationTables = []string{"posts", "notes", "settings", "api_keys", "media", "media_folders", "users", "user_tokens", "comments", "reactions", "api_logs"}

type Queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func quoteColumn(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func (db *DB) Columns(ctx context.Context, q Queryer, table string) ([]string, error) {
	query := `SELECT name FROM pragma_table_info(?)`
	if db.Driver() == "mysql" {
		query = `SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION`
	}
	rows, err := q.QueryContext(ctx, query, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	return columns, rows.Err()
}

// CopyRows transfers values, never SQL from a backup. Shared columns preserve old backups.
func CopyRows(ctx context.Context, src Queryer, dst *sql.Tx, table string, columns []string) error {
	if len(columns) == 0 {
		return nil
	}
	quoted := make([]string, len(columns))
	for i, col := range columns {
		quoted[i] = quoteColumn(col)
	}
	cols := strings.Join(quoted, ",")
	rows, err := src.QueryContext(ctx, "SELECT "+cols+" FROM "+quoteColumn(table))
	if err != nil {
		return err
	}
	defer rows.Close()
	stmt, err := dst.PrepareContext(ctx, "INSERT INTO "+quoteColumn(table)+" ("+cols+") VALUES ("+strings.TrimSuffix(strings.Repeat("?,", len(columns)), ",")+")")
	if err != nil {
		return err
	}
	defer stmt.Close()
	values := make([]any, len(columns))
	pointers := make([]any, len(columns))
	for i := range values {
		pointers[i] = &values[i]
	}
	for rows.Next() {
		if err := rows.Scan(pointers...); err != nil {
			return err
		}
		for i, v := range values {
			if bytes, ok := v.([]byte); ok {
				values[i] = string(bytes)
			}
		}
		if _, err := stmt.ExecContext(ctx, values...); err != nil {
			return fmt.Errorf("copy %s: %w", table, err)
		}
	}
	return rows.Err()
}

// SnapshotTo writes the same SQLite interchange format for both database backends.
// SQLite VACUUM INTO and MySQL REPEATABLE READ each provide a consistent DB snapshot.
func (db *DB) SnapshotTo(ctx context.Context, filename string) error {
	if db.Driver() == "sqlite" {
		_, err := db.ExecContext(ctx, `VACUUM INTO ?`, filepath.ToSlash(filename))
		return err
	}
	dest, err := openSQLiteFile(filename)
	if err != nil {
		return err
	}
	defer dest.Close()
	srcTx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return err
	}
	defer srcTx.Rollback()
	tx, err := dest.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, table := range ApplicationTables {
		columns, err := db.Columns(ctx, srcTx, table)
		if err != nil {
			return err
		}
		if err := CopyRows(ctx, srcTx, tx, table, columns); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	_, err = dest.ExecContext(ctx, `PRAGMA wal_checkpoint(TRUNCATE)`)
	return err
}

// RestoreSnapshot loads a SQLite backup into MySQL in one transaction.
// The caller can atomically swap uploads immediately before committing.
func (db *DB) RestoreSnapshot(ctx context.Context, filename string, beforeCommit func() error) error {
	src, err := sql.Open("sqlite", "file:"+filepath.ToSlash(filename)+"?mode=ro&_pragma=query_only(1)")
	if err != nil {
		return err
	}
	defer src.Close()
	var n int
	if err := src.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='posts'`).Scan(&n); err != nil || n != 1 {
		return errors.New("invalid_backup")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	source := &DB{DB: src}
	for _, table := range ApplicationTables {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+quoteColumn(table)); err != nil {
			return err
		}
		from, err := source.Columns(ctx, src, table)
		if err != nil {
			return err
		}
		to, err := db.Columns(ctx, tx, table)
		if err != nil {
			return err
		}
		var common []string
		for _, a := range to {
			for _, b := range from {
				if a == b {
					common = append(common, a)
				}
			}
		}
		if err := CopyRows(ctx, src, tx, table, common); err != nil {
			return err
		}
	}
	if err := beforeCommit(); err != nil {
		return err
	}
	return tx.Commit()
}
