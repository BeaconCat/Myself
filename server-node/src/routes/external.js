import { Router } from 'express';
import { db, toPost } from '../db.js';
import { requireApiKey } from './apikeys.js';

/**
 * 外部通道（X-Api-Key 认证）：供外部 AI/脚本全托管内容。
 * 覆盖文章/随想的增删改查，内容一律 Markdown。
 */
export const externalRouter = Router();
externalRouter.use('/ext', requireApiKey);

/** GET /api/v1/ext/posts?status=all|published|draft 列表（含草稿） */
externalRouter.get('/ext/posts', (req, res) => {
  const status = String(req.query.status ?? 'all');
  const where = status === 'all' ? '1=1' : 'status = @status';
  const rows = db
    .prepare(`SELECT * FROM posts WHERE ${where} ORDER BY created_at DESC`)
    .all({ status });
  res.json(rows.map((r) => ({ ...toPost(r), status: r.status })));
});

/** GET /api/v1/ext/posts/:id 详情（含 Markdown 正文） */
externalRouter.get('/ext/posts/:id', (req, res) => {
  const row = db.prepare('SELECT * FROM posts WHERE id = ?').get(req.params.id);
  if (!row) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  res.json({ ...toPost(row, { withContent: true }), status: row.status });
});

/** POST /api/v1/ext/posts 投稿文章（默认草稿，可显式 published） */
externalRouter.post('/ext/posts', (req, res) => {
  const b = req.body ?? {};
  if (
    typeof b.slug !== 'string' || !/^[a-z0-9-]{1,80}$/.test(b.slug)
    || typeof b.title !== 'string' || !b.title.trim()
    || typeof b.contentMd !== 'string'
  ) {
    res.status(400).json({ error: 'invalid_post' });
    return;
  }
  try {
    const info = db.prepare(`
      INSERT INTO posts (slug, title, excerpt, content_md, covers, tags, status)
      VALUES (@slug, @title, @excerpt, @contentMd, @covers, @tags, @status)
    `).run({
      slug: b.slug,
      title: b.title.trim(),
      excerpt: String(b.excerpt ?? ''),
      contentMd: b.contentMd,
      covers: JSON.stringify(Array.isArray(b.covers) ? b.covers.slice(0, 3) : []),
      tags: JSON.stringify(Array.isArray(b.tags) ? b.tags : []),
      status: b.status === 'published' ? 'published' : 'draft',
    });
    res.status(201).json({ id: info.lastInsertRowid, slug: b.slug });
  } catch (err) {
    if (String(err).includes('UNIQUE')) {
      res.status(409).json({ error: 'slug_exists' });
      return;
    }
    throw err;
  }
});

/** PUT /api/v1/ext/posts/:id 更新文章 */
externalRouter.put('/ext/posts/:id', (req, res) => {
  const b = req.body ?? {};
  if (
    typeof b.title !== 'string' || !b.title.trim()
    || typeof b.contentMd !== 'string'
  ) {
    res.status(400).json({ error: 'invalid_post' });
    return;
  }
  const info = db.prepare(`
    UPDATE posts SET
      title = @title, excerpt = @excerpt, content_md = @contentMd,
      covers = @covers, tags = @tags, status = @status, updated_at = datetime('now')
    WHERE id = @id
  `).run({
    id: req.params.id,
    title: b.title.trim(),
    excerpt: String(b.excerpt ?? ''),
    contentMd: b.contentMd,
    covers: JSON.stringify(Array.isArray(b.covers) ? b.covers.slice(0, 3) : []),
    tags: JSON.stringify(Array.isArray(b.tags) ? b.tags : []),
    status: b.status === 'published' ? 'published' : 'draft',
  });
  if (!info.changes) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  res.json({ ok: true });
});

/** DELETE /api/v1/ext/posts/:id */
externalRouter.delete('/ext/posts/:id', (req, res) => {
  const info = db.prepare('DELETE FROM posts WHERE id = ?').run(req.params.id);
  if (!info.changes) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  res.json({ ok: true });
});

/** GET /api/v1/ext/notes 随想列表 */
externalRouter.get('/ext/notes', (req, res) => {
  const pageSize = Math.min(100, Math.max(1, Number(req.query.pageSize) || 50));
  const rows = db
    .prepare('SELECT * FROM notes ORDER BY created_at DESC LIMIT ?')
    .all(pageSize);
  res.json(rows.map((r) => ({
    id: r.id,
    contentMd: r.content_md,
    mood: r.mood,
    images: JSON.parse(r.images ?? '[]'),
    createdAt: r.created_at,
  })));
});

/** PUT /api/v1/ext/notes/:id 更新随想 */
externalRouter.put('/ext/notes/:id', (req, res) => {
  const b = req.body ?? {};
  if (typeof b.contentMd !== 'string' || !b.contentMd.trim()) {
    res.status(400).json({ error: 'invalid_note' });
    return;
  }
  const info = db.prepare(`
    UPDATE notes SET content_md = @contentMd, mood = @mood, images = @images WHERE id = @id
  `).run({
    id: req.params.id,
    contentMd: b.contentMd.trim(),
    mood: String(b.mood ?? ''),
    images: JSON.stringify(Array.isArray(b.images) ? b.images.slice(0, 9) : []),
  });
  if (!info.changes) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  res.json({ ok: true });
});

/** DELETE /api/v1/ext/notes/:id */
externalRouter.delete('/ext/notes/:id', (req, res) => {
  const info = db.prepare('DELETE FROM notes WHERE id = ?').run(req.params.id);
  if (!info.changes) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  res.json({ ok: true });
});

/** POST /api/v1/ext/notes 投稿随想 */
externalRouter.post('/ext/notes', (req, res) => {
  const b = req.body ?? {};
  if (typeof b.contentMd !== 'string' || !b.contentMd.trim()) {
    res.status(400).json({ error: 'invalid_note' });
    return;
  }
  const info = db.prepare(`
    INSERT INTO notes (content_md, mood, images) VALUES (@contentMd, @mood, @images)
  `).run({
    contentMd: b.contentMd.trim(),
    mood: String(b.mood ?? ''),
    images: JSON.stringify(Array.isArray(b.images) ? b.images.slice(0, 9) : []),
  });
  res.status(201).json({ id: info.lastInsertRowid });
});
