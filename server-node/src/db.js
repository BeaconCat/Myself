import Database from 'better-sqlite3';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { seedNotes, seedPosts } from './seed.js';

const dataDir = path.join(path.dirname(fileURLToPath(import.meta.url)), '..', 'data');
fs.mkdirSync(dataDir, { recursive: true });

export const db = new Database(path.join(dataDir, 'myself.db'));
db.pragma('journal_mode = WAL');

db.exec(`
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
`);

/* 旧库缺列时补齐（幂等迁移） */
try {
  db.exec(`ALTER TABLE notes ADD COLUMN images TEXT NOT NULL DEFAULT '[]'`);
} catch { /* 列已存在 */ }
try {
  db.exec(`ALTER TABLE posts ADD COLUMN pinned INTEGER NOT NULL DEFAULT 0`);
} catch { /* 列已存在 */ }
try {
  db.exec(`ALTER TABLE notes ADD COLUMN pinned INTEGER NOT NULL DEFAULT 0`);
} catch { /* 列已存在 */ }

/** 空表分别注入演示数据 */
export function seedIfEmpty() {
  seedNotesIfEmpty();
  const { n } = db.prepare('SELECT COUNT(*) AS n FROM posts').get();
  if (n > 0) return;
  const insert = db.prepare(`
    INSERT INTO posts (slug, title, excerpt, content_md, covers, tags, created_at)
    VALUES (@slug, @title, @excerpt, @contentMd, @covers, @tags, @createdAt)
  `);
  const tx = db.transaction((posts) => {
    for (const p of posts) {
      insert.run({
        ...p,
        covers: JSON.stringify(p.covers ?? []),
        tags: JSON.stringify(p.tags ?? []),
      });
    }
  });
  tx(seedPosts);
}

function seedNotesIfEmpty() {
  const { n } = db.prepare('SELECT COUNT(*) AS n FROM notes').get();
  if (n > 0) return;
  const insertNote = db.prepare(`
    INSERT INTO notes (content_md, mood, images, created_at)
    VALUES (@contentMd, @mood, @images, @createdAt)
  `);
  const txNotes = db.transaction((notes) => {
    for (const note of notes) {
      insertNote.run({ ...note, images: JSON.stringify(note.images ?? []) });
    }
  });
  txNotes(seedNotes);
}

/** 行 → API 对象 */
export function toPost(row, { withContent = false } = {}) {
  if (!row) return null;
  const post = {
    id: row.id,
    slug: row.slug,
    title: row.title,
    excerpt: row.excerpt,
    covers: JSON.parse(row.covers),
    tags: JSON.parse(row.tags),
    pinned: !!row.pinned,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  };
  if (withContent) post.contentMd = row.content_md;
  return post;
}
