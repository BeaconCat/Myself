import crypto from 'node:crypto';
import { Router } from 'express';
import { requireAuth } from '../auth.js';
import { db } from '../db.js';

db.exec(`
  CREATE TABLE IF NOT EXISTS api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    prefix TEXT NOT NULL,
    last_used_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
  );
`);

function hashKey(key) {
  return crypto.createHash('sha256').update(key).digest('hex');
}

/** X-Api-Key 认证中间件（外部 AI 发文通道） */
export function requireApiKey(req, res, next) {
  const key = req.headers['x-api-key'];
  if (typeof key !== 'string' || !key) {
    res.status(401).json({ error: 'missing_api_key' });
    return;
  }
  const row = db
    .prepare('SELECT id FROM api_keys WHERE key_hash = ?')
    .get(hashKey(key));
  if (!row) {
    res.status(401).json({ error: 'invalid_api_key' });
    return;
  }
  db.prepare(`UPDATE api_keys SET last_used_at = datetime('now') WHERE id = ?`).run(row.id);
  next();
}

export const apiKeysRouter = Router();
// 注意：只保护本模块路径。router.use(requireAuth) 会拦下挂载点上所有请求（包括登录）
apiKeysRouter.use('/admin/apikeys', requireAuth);

/** GET /api/v1/admin/apikeys */
apiKeysRouter.get('/admin/apikeys', (_req, res) => {
  const rows = db
    .prepare('SELECT id, name, prefix, last_used_at, created_at FROM api_keys ORDER BY id DESC')
    .all();
  res.json(rows.map((r) => ({
    id: r.id,
    name: r.name,
    prefix: r.prefix,
    lastUsedAt: r.last_used_at,
    createdAt: r.created_at,
  })));
});

/** POST /api/v1/admin/apikeys 创建（明文只返回这一次） */
apiKeysRouter.post('/admin/apikeys', (req, res) => {
  const name = String(req.body?.name ?? '').trim();
  if (!name) {
    res.status(400).json({ error: 'name_required' });
    return;
  }
  const key = `myk_${crypto.randomBytes(24).toString('hex')}`;
  const prefix = key.slice(0, 12);
  const info = db
    .prepare('INSERT INTO api_keys (name, key_hash, prefix) VALUES (?, ?, ?)')
    .run(name, hashKey(key), prefix);
  res.status(201).json({ id: info.lastInsertRowid, name, prefix, key });
});

/** DELETE /api/v1/admin/apikeys/:id 吊销 */
apiKeysRouter.delete('/admin/apikeys/:id', (req, res) => {
  const info = db.prepare('DELETE FROM api_keys WHERE id = ?').run(req.params.id);
  if (!info.changes) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  res.json({ ok: true });
});
