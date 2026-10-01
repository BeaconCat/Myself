package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestPublicationColumnsPreserveLegacyAndSnapshots(t *testing.T) {
	dir := t.TempDir()
	old, err := sql.Open("sqlite", filepath.Join(dir, "myself.db"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = old.Exec(`CREATE TABLE posts(id INTEGER PRIMARY KEY,slug TEXT UNIQUE,title TEXT,excerpt TEXT,content_md TEXT,covers TEXT,tags TEXT,status TEXT,created_at TEXT,updated_at TEXT);
 CREATE TABLE notes(id INTEGER PRIMARY KEY,content_md TEXT,mood TEXT,created_at TEXT);
 INSERT INTO posts VALUES(1,'legacy','旧文章','','旧正文','[]','[]','published','2020-01-02 03:04:05','2020-01-02 03:04:05');
 INSERT INTO notes VALUES(1,'旧随想','','2020-01-02 03:04:05');`)
	old.Close()
	if err != nil {
		t.Fatal(err)
	}
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.migrate(); err != nil {
		t.Fatal(err)
	}
	post, err := db.GetPost(`WHERE id = 1`)
	if err != nil {
		t.Fatal(err)
	}
	if post.ContentMd != "旧正文" || post.Status != "published" || post.PublishAt != "" {
		t.Fatal(post)
	}
	notes, err := db.QueryNotes(`WHERE id = 1`, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 || notes[0].Status != "published" || notes[0].Views != 0 || notes[0].CreatedAt != "2020-01-02 03:04:05" {
		t.Fatal(notes)
	}
	if _, err := db.Exec(`UPDATE notes SET status='scheduled',publish_at='2099-01-02 03:04:05',views=42 WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	snapshotDir := t.TempDir()
	if err := db.SnapshotTo(context.Background(), filepath.Join(snapshotDir, "myself.db")); err != nil {
		t.Fatal(err)
	}
	restored, err := Open(snapshotDir)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	notes, err = restored.QueryNotes(`WHERE id = 1`, true)
	if err != nil {
		t.Fatal(err)
	}
	if notes[0].ContentMd != "旧随想" || notes[0].Status != "scheduled" || notes[0].PublishAt != "2099-01-02T03:04:05Z" || notes[0].Views != 42 {
		t.Fatal(notes)
	}
}
