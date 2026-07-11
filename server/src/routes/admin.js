import { Router } from 'express';
import { changePassword, login, requireAuth } from '../auth.js';
import { db, toPost } from '../db.js';

export const adminRouter = Router();

/** POST /api/v1/auth/login */
adminRouter.post('/auth/login', (req, res) => {
  const { username, password } = req.body ?? {};
  const token = login(String(username ?? ''), String(password ?? ''));
  if (!token) {
    res.status(401).json({ error: 'bad_credentials' });
    return;
  }
  res.json({ token });
});

/* ===== 以下全部需要管理员 Token ===== */
adminRouter.use(requireAuth);

/** PUT /api/v1/auth/password */
adminRouter.put('/auth/password', (req, res) => {
  const { oldPassword, newPassword } = req.body ?? {};
  if (typeof newPassword !== 'string' || newPassword.length < 8) {
    res.status(400).json({ error: 'weak_password' });
    return;
  }
  if (!changePassword(String(oldPassword ?? ''), newPassword)) {
    res.status(401).json({ error: 'bad_credentials' });
    return;
  }
  res.json({ ok: true });
});

/** GET /api/v1/admin/posts 全量（含草稿） */
adminRouter.get('/admin/posts', (_req, res) => {
  const rows = db.prepare('SELECT * FROM posts ORDER BY created_at DESC').all();
  res.json(rows.map((r) => ({ ...toPost(r), status: r.status })));
});

/** GET /api/v1/admin/posts/:id 编辑用详情 */
adminRouter.get('/admin/posts/:id', (req, res) => {
  const row = db.prepare('SELECT * FROM posts WHERE id = ?').get(req.params.id);
  if (!row) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  res.json({ ...toPost(row, { withContent: true }), status: row.status });
});

function validPost(body) {
  return (
    body
    && typeof body.slug === 'string' && /^[a-z0-9-]{1,80}$/.test(body.slug)
    && typeof body.title === 'string' && body.title.trim()
    && typeof body.contentMd === 'string'
  );
}

/** POST /api/v1/admin/posts 新建 */
adminRouter.post('/admin/posts', (req, res) => {
  const b = req.body;
  if (!validPost(b)) {
    res.status(400).json({ error: 'invalid_post' });
    return;
  }
  try {
    const info = db.prepare(`
      INSERT INTO posts (slug, title, excerpt, content_md, covers, tags, status, pinned)
      VALUES (@slug, @title, @excerpt, @contentMd, @covers, @tags, @status, @pinned)
    `).run({
      slug: b.slug,
      title: b.title.trim(),
      excerpt: String(b.excerpt ?? ''),
      contentMd: b.contentMd,
      covers: JSON.stringify(Array.isArray(b.covers) ? b.covers : []),
      tags: JSON.stringify(Array.isArray(b.tags) ? b.tags : []),
      status: b.status === 'draft' ? 'draft' : 'published',
      pinned: b.pinned ? 1 : 0,
    });
    res.status(201).json({ id: info.lastInsertRowid });
  } catch (err) {
    if (String(err).includes('UNIQUE')) {
      res.status(409).json({ error: 'slug_exists' });
      return;
    }
    throw err;
  }
});

/** PUT /api/v1/admin/posts/:id 更新 */
adminRouter.put('/admin/posts/:id', (req, res) => {
  const b = req.body;
  if (!validPost(b)) {
    res.status(400).json({ error: 'invalid_post' });
    return;
  }
  const info = db.prepare(`
    UPDATE posts SET
      slug = @slug, title = @title, excerpt = @excerpt, content_md = @contentMd,
      covers = @covers, tags = @tags, status = @status, pinned = @pinned,
      updated_at = datetime('now')
    WHERE id = @id
  `).run({
    id: req.params.id,
    slug: b.slug,
    title: b.title.trim(),
    excerpt: String(b.excerpt ?? ''),
    contentMd: b.contentMd,
    covers: JSON.stringify(Array.isArray(b.covers) ? b.covers : []),
    tags: JSON.stringify(Array.isArray(b.tags) ? b.tags : []),
    status: b.status === 'draft' ? 'draft' : 'published',
    pinned: b.pinned ? 1 : 0,
  });
  if (!info.changes) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  res.json({ ok: true });
});

/** DELETE /api/v1/admin/posts/:id */
adminRouter.delete('/admin/posts/:id', (req, res) => {
  const info = db.prepare('DELETE FROM posts WHERE id = ?').run(req.params.id);
  if (!info.changes) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  res.json({ ok: true });
});

/** POST /api/v1/admin/notes 发随想 */
adminRouter.post('/admin/notes', (req, res) => {
  const b = req.body;
  if (!b || typeof b.contentMd !== 'string' || !b.contentMd.trim()) {
    res.status(400).json({ error: 'invalid_note' });
    return;
  }
  const info = db.prepare(`
    INSERT INTO notes (content_md, mood, images)
    VALUES (@contentMd, @mood, @images)
  `).run({
    contentMd: b.contentMd.trim(),
    mood: String(b.mood ?? ''),
    images: JSON.stringify(Array.isArray(b.images) ? b.images.slice(0, 9) : []),
  });
  res.status(201).json({ id: info.lastInsertRowid });
});

/** PUT /api/v1/admin/notes/:id 编辑随想 */
adminRouter.put('/admin/notes/:id', (req, res) => {
  const b = req.body;
  if (!b || typeof b.contentMd !== 'string' || !b.contentMd.trim()) {
    res.status(400).json({ error: 'invalid_note' });
    return;
  }
  const info = db.prepare(`
    UPDATE notes SET content_md = @contentMd, mood = @mood, images = @images
    WHERE id = @id
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

/** DELETE /api/v1/admin/notes/:id */
adminRouter.delete('/admin/notes/:id', (req, res) => {
  const info = db.prepare('DELETE FROM notes WHERE id = ?').run(req.params.id);
  if (!info.changes) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  res.json({ ok: true });
});
