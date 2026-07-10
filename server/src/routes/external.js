import { Router } from 'express';
import { db } from '../db.js';
import { requireApiKey } from './apikeys.js';

/**
 * 外部发文通道（X-Api-Key 认证）：供外部 AI/脚本投稿。
 * 内容一律 Markdown。
 */
export const externalRouter = Router();
externalRouter.use('/ext', requireApiKey);

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
