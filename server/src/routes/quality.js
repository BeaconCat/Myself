import fs from 'node:fs';
import path from 'node:path';
import { Router } from 'express';
import sharp from 'sharp';
import { requireAuth } from '../auth.js';
import { db } from '../db.js';
import { UPLOAD_DIR } from './media.js';

const ORIGINALS_DIR = path.join(UPLOAD_DIR, '.originals');

const COMPRESSIBLE = new Set(['.png', '.jpg', '.jpeg', '.webp']);

/** 站内 URL 引用替换：压缩改名（png→webp）后同步修正文章/随想里的链接 */
function replaceUrlRefs(oldUrl, newUrl) {
  const like = `%${oldUrl}%`;
  for (const [table, cols] of [
    ['posts', ['covers', 'content_md']],
    ['notes', ['images', 'content_md']],
  ]) {
    for (const col of cols) {
      db.prepare(
        `UPDATE ${table} SET ${col} = REPLACE(${col}, @oldUrl, @newUrl) WHERE ${col} LIKE @like`,
      ).run({ oldUrl, newUrl, like });
    }
  }
}

export const qualityRouter = Router();
qualityRouter.use('/admin/quality', requireAuth);

/** GET /api/v1/admin/quality/scan 扫描可压缩图片 */
qualityRouter.get('/admin/quality/scan', async (_req, res) => {
  const names = fs.readdirSync(UPLOAD_DIR)
    .filter((f) => !f.startsWith('.'))
    .filter((f) => COMPRESSIBLE.has(path.extname(f).toLowerCase()));

  const items = [];
  for (const name of names) {
    const full = path.join(UPLOAD_DIR, name);
    const stat = fs.statSync(full);
    const ext = path.extname(name).toLowerCase();
    let meta = {};
    try {
      meta = await sharp(full).metadata();
    } catch { continue; }
    items.push({
      name,
      url: `/uploads/${name}`,
      size: stat.size,
      format: ext.slice(1),
      width: meta.width ?? 0,
      height: meta.height ?? 0,
      hasAlpha: !!meta.hasAlpha,
      /** png 转 webp、jpg/webp 重压均有收益空间 */
      compressible: true,
    });
  }
  res.json(items.sort((a, b) => b.size - a.size));
});

/**
 * POST /api/v1/admin/quality/compress { names: string[], quality: 1-100 }
 * PNG → WebP（保留 alpha，文件名改 .webp，站内引用自动替换）；
 * JPG/WebP → 原格式重压。原图备份到 .originals 保留后路。
 */
qualityRouter.post('/admin/quality/compress', async (req, res) => {
  const names = Array.isArray(req.body?.names) ? req.body.names : [];
  const quality = Math.min(100, Math.max(1, Number(req.body?.quality) || 80));
  const results = [];

  for (const rawName of names) {
    const name = String(rawName);
    if (name.includes('..') || name.includes('/') || name.includes('\\')) continue;
    const full = path.join(UPLOAD_DIR, name);
    if (!fs.existsSync(full)) continue;
    const ext = path.extname(name).toLowerCase();
    if (!COMPRESSIBLE.has(ext)) continue;

    const before = fs.statSync(full).size;
    const backup = path.join(ORIGINALS_DIR, name);
    if (!fs.existsSync(backup)) fs.copyFileSync(full, backup);

    try {
      if (ext === '.png') {
        const newName = `${name.slice(0, -ext.length)}.webp`;
        const buffer = await sharp(full).webp({ quality, alphaQuality: 100 }).toBuffer();
        fs.writeFileSync(path.join(UPLOAD_DIR, newName), buffer);
        fs.rmSync(full, { force: true });
        db.prepare(`
          INSERT INTO media (name) VALUES (@newName)
          ON CONFLICT(name) DO NOTHING
        `).run({ newName });
        db.prepare('DELETE FROM media WHERE name = ?').run(name);
        replaceUrlRefs(`/uploads/${name}`, `/uploads/${newName}`);
        results.push({ name, newName, before, after: buffer.length });
      } else {
        const pipeline = sharp(full);
        const buffer = ext === '.webp'
          ? await pipeline.webp({ quality }).toBuffer()
          : await pipeline.jpeg({ quality, mozjpeg: true }).toBuffer();
        // 只有确实变小才替换
        if (buffer.length < before) fs.writeFileSync(full, buffer);
        results.push({ name, newName: name, before, after: Math.min(buffer.length, before) });
      }
    } catch (err) {
      results.push({ name, error: String(err?.message ?? err) });
    }
  }
  res.json(results);
});
