import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { Router } from 'express';
import multer from 'multer';
import sharp from 'sharp';
import { requireAuth } from '../auth.js';
import { db } from '../db.js';

export const UPLOAD_DIR = path.join(
  path.dirname(fileURLToPath(import.meta.url)), '..', '..', 'uploads',
);
const ORIGINALS_DIR = path.join(UPLOAD_DIR, '.originals');
fs.mkdirSync(ORIGINALS_DIR, { recursive: true });

db.exec(`
  CREATE TABLE IF NOT EXISTS media (
    name TEXT PRIMARY KEY,
    crop_json TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
  );
`);

const ALLOWED = new Set(['.png', '.jpg', '.jpeg', '.webp', '.gif']);

const storage = multer.diskStorage({
  destination: UPLOAD_DIR,
  filename: (_req, file, cb) => {
    const ext = path.extname(file.originalname).toLowerCase();
    cb(null, `${Date.now().toString(36)}-${crypto.randomBytes(4).toString('hex')}${ext}`);
  },
});

const upload = multer({
  storage,
  limits: { fileSize: 20 * 1024 * 1024, files: 20 },
  fileFilter: (_req, file, cb) => {
    cb(null, ALLOWED.has(path.extname(file.originalname).toLowerCase()));
  },
});

/** 防路径穿越 */
function safeName(name) {
  return /^[a-z0-9.-]+$/i.test(name) && !name.includes('..') ? name : null;
}

function fileInfo(name) {
  const full = path.join(UPLOAD_DIR, name);
  const stat = fs.statSync(full);
  const meta = db.prepare('SELECT crop_json, created_at FROM media WHERE name = ?').get(name);
  return {
    name,
    url: `/uploads/${name}`,
    size: stat.size,
    hasOriginal: fs.existsSync(path.join(ORIGINALS_DIR, name)),
    crop: meta?.crop_json ? JSON.parse(meta.crop_json) : null,
    createdAt: meta?.created_at ?? stat.mtime.toISOString(),
  };
}

export const mediaRouter = Router();
mediaRouter.use('/admin/media', requireAuth);

/** GET /api/v1/admin/media 素材列表 */
mediaRouter.get('/admin/media', (_req, res) => {
  const names = fs.readdirSync(UPLOAD_DIR)
    .filter((f) => ALLOWED.has(path.extname(f).toLowerCase()))
    .filter((f) => !f.startsWith('.'));
  const items = names.map(fileInfo).sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1));
  res.json(items);
});

/** POST /api/v1/admin/media 上传（单个/批量，字段 files） */
mediaRouter.post('/admin/media', upload.array('files', 20), (req, res) => {
  const files = req.files ?? [];
  const insert = db.prepare('INSERT OR IGNORE INTO media (name) VALUES (?)');
  for (const f of files) insert.run(f.filename);
  res.status(201).json(files.map((f) => fileInfo(f.filename)));
});

/**
 * POST /api/v1/admin/media/:name/crop {left,top,width,height}
 * 首次裁切时把当前文件备份为原图；此后一律从原图重裁并替换公开文件，
 * 裁切框入库，反悔时可调出原图 + 上次框继续调整。
 */
mediaRouter.post('/admin/media/:name/crop', async (req, res) => {
  const name = safeName(req.params.name);
  if (!name || !fs.existsSync(path.join(UPLOAD_DIR, name))) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  const { left, top, width, height } = req.body ?? {};
  const rect = {
    left: Math.max(0, Math.round(Number(left))),
    top: Math.max(0, Math.round(Number(top))),
    width: Math.round(Number(width)),
    height: Math.round(Number(height)),
  };
  if (!(rect.width > 0) || !(rect.height > 0)) {
    res.status(400).json({ error: 'invalid_rect' });
    return;
  }

  const publicPath = path.join(UPLOAD_DIR, name);
  const originalPath = path.join(ORIGINALS_DIR, name);
  if (!fs.existsSync(originalPath)) fs.copyFileSync(publicPath, originalPath);

  const source = sharp(originalPath);
  const info = await source.metadata();
  rect.width = Math.min(rect.width, (info.width ?? rect.width) - rect.left);
  rect.height = Math.min(rect.height, (info.height ?? rect.height) - rect.top);

  const buffer = await source.extract(rect).toBuffer();
  fs.writeFileSync(publicPath, buffer);

  db.prepare(`
    INSERT INTO media (name, crop_json) VALUES (@name, @crop)
    ON CONFLICT(name) DO UPDATE SET crop_json = excluded.crop_json
  `).run({ name, crop: JSON.stringify(rect) });

  res.json(fileInfo(name));
});

/** GET /api/v1/admin/media/:name/original 原图（重裁 UI 用） */
mediaRouter.get('/admin/media/:name/original', (req, res) => {
  const name = safeName(req.params.name);
  if (!name) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  const original = path.join(ORIGINALS_DIR, name);
  const fallback = path.join(UPLOAD_DIR, name);
  const target = fs.existsSync(original) ? original : fallback;
  if (!fs.existsSync(target)) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  res.sendFile(target);
});

/** DELETE /api/v1/admin/media/:name 删除（连同原图与裁切记录） */
mediaRouter.delete('/admin/media/:name', (req, res) => {
  const name = safeName(req.params.name);
  if (!name || !fs.existsSync(path.join(UPLOAD_DIR, name))) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  fs.rmSync(path.join(UPLOAD_DIR, name), { force: true });
  fs.rmSync(path.join(ORIGINALS_DIR, name), { force: true });
  db.prepare('DELETE FROM media WHERE name = ?').run(name);
  res.json({ ok: true });
});
