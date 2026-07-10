import Database from 'better-sqlite3';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { seedPosts } from './seed.js';

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
`);

/** 空库时注入演示文章 */
export function seedIfEmpty() {
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
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  };
  if (withContent) post.contentMd = row.content_md;
  return post;
}
