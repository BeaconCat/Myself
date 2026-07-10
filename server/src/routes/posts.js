import { Router } from 'express';
import { db, toPost } from '../db.js';

export const postsRouter = Router();

/** GET /api/v1/posts?page=&pageSize=&tag=&q= 文章列表（分页 + 标签过滤 + 关键词搜索） */
postsRouter.get('/posts', (req, res) => {
  const page = Math.max(1, Number(req.query.page) || 1);
  const pageSize = Math.min(50, Math.max(1, Number(req.query.pageSize) || 10));
  const tag = typeof req.query.tag === 'string' ? req.query.tag : '';
  const q = typeof req.query.q === 'string' ? req.query.q.trim() : '';

  const where = ["status = 'published'"];
  const params = {};
  if (tag) {
    // tags 为 JSON 数组文本，用引号包裹精确匹配单个标签
    where.push(`tags LIKE @tagPattern`);
    params.tagPattern = `%"${tag.replaceAll('"', '')}"%`;
  }
  if (q) {
    where.push(`(title LIKE @q OR excerpt LIKE @q OR content_md LIKE @q)`);
    params.q = `%${q}%`;
  }
  const whereSql = where.join(' AND ');

  const { total } = db
    .prepare(`SELECT COUNT(*) AS total FROM posts WHERE ${whereSql}`)
    .get(params);
  const rows = db
    .prepare(
      `SELECT * FROM posts WHERE ${whereSql}
       ORDER BY created_at DESC LIMIT @limit OFFSET @offset`,
    )
    .all({ ...params, limit: pageSize, offset: (page - 1) * pageSize });

  res.json({
    items: rows.map((r) => toPost(r)),
    page,
    pageSize,
    total,
  });
});

/** GET /api/v1/posts/:slug 文章详情（含 Markdown 正文） */
postsRouter.get('/posts/:slug', (req, res) => {
  const row = db
    .prepare(`SELECT * FROM posts WHERE slug = ? AND status = 'published'`)
    .get(req.params.slug);
  if (!row) {
    res.status(404).json({ error: 'post_not_found' });
    return;
  }
  res.json(toPost(row, { withContent: true }));
});

/** GET /api/v1/tags 标签及计数 */
postsRouter.get('/tags', (_req, res) => {
  const rows = db
    .prepare(`SELECT tags FROM posts WHERE status = 'published'`)
    .all();
  const counts = new Map();
  for (const { tags } of rows) {
    for (const tag of JSON.parse(tags)) {
      counts.set(tag, (counts.get(tag) ?? 0) + 1);
    }
  }
  res.json([...counts.entries()].map(([name, count]) => ({ name, count })));
});

/** GET /api/v1/notes?page=&pageSize= 随想信息流（时间倒序） */
postsRouter.get('/notes', (req, res) => {
  const page = Math.max(1, Number(req.query.page) || 1);
  const pageSize = Math.min(50, Math.max(1, Number(req.query.pageSize) || 20));
  const { total } = db.prepare('SELECT COUNT(*) AS total FROM notes').get();
  const rows = db
    .prepare('SELECT * FROM notes ORDER BY created_at DESC LIMIT @limit OFFSET @offset')
    .all({ limit: pageSize, offset: (page - 1) * pageSize });
  res.json({
    items: rows.map((r) => ({
      id: r.id,
      contentMd: r.content_md,
      mood: r.mood,
      images: JSON.parse(r.images ?? '[]'),
      createdAt: r.created_at,
    })),
    page,
    pageSize,
    total,
  });
});

/** GET /api/v1/img/:from/:to/:label 本地渐变占位图（演示配图，无外部资源） */
postsRouter.get('/img/:from/:to/:label', (req, res) => {
  const hex = /^[0-9a-fA-F]{6}$/;
  const { from, to } = req.params;
  if (!hex.test(from) || !hex.test(to)) {
    res.status(400).json({ error: 'bad_color' });
    return;
  }
  const label = String(req.params.label).slice(0, 24).replace(/[<>&"']/g, '');
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="900" height="900">
  <defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
    <stop offset="0" stop-color="#${from}"/><stop offset="1" stop-color="#${to}"/>
  </linearGradient></defs>
  <rect width="900" height="900" fill="url(#g)"/>
  <text x="50%" y="52%" text-anchor="middle" font-family="sans-serif" font-size="52"
    fill="rgba(255,255,255,.85)" font-weight="700">${label}</text>
</svg>`;
  res.type('image/svg+xml').setHeader('Cache-Control', 'public, max-age=86400').send(svg);
});

/** GET /api/v1/archive 按年-月归档 */
postsRouter.get('/archive', (_req, res) => {
  const rows = db
    .prepare(
      `SELECT * FROM posts WHERE status = 'published' ORDER BY created_at DESC`,
    )
    .all();
  const groups = new Map();
  for (const row of rows) {
    const key = row.created_at.slice(0, 7); // YYYY-MM
    if (!groups.has(key)) groups.set(key, []);
    groups.get(key).push(toPost(row));
  }
  res.json([...groups.entries()].map(([month, items]) => ({ month, items })));
});
