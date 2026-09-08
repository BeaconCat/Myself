import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { ZipArchive } from 'archiver';
import { Router } from 'express';
import { requireAuth } from '../auth.js';
import { getConfig } from '../config.js';
import { UPLOAD_DIR } from './media.js';

const ROOT = path.join(path.dirname(fileURLToPath(import.meta.url)), '..', '..');
const DATA_DIR = path.join(ROOT, 'data');
const BACKUP_DIR = path.join(ROOT, 'backups');
fs.mkdirSync(BACKUP_DIR, { recursive: true });

function timestamp() {
  return new Date().toISOString().replace(/[:T]/g, '-').slice(0, 19);
}

/** 全站备份：数据库（文章/随想/用户/配置） + 上传素材（含原图备份） */
export function createBackup() {
  return new Promise((resolve, reject) => {
    const name = `backup-${timestamp()}.zip`;
    const out = fs.createWriteStream(path.join(BACKUP_DIR, name));
    const archive = new ZipArchive({ zlib: { level: 6 } });
    out.on('close', () => resolve(name));
    archive.on('error', reject);
    archive.pipe(out);
    if (fs.existsSync(DATA_DIR)) archive.directory(DATA_DIR, 'data');
    if (fs.existsSync(UPLOAD_DIR)) archive.directory(UPLOAD_DIR, 'uploads');
    void archive.finalize();
  });
}

/** 自动备份调度：间隔小时数取自站点配置，0 关闭；每分钟核对一次 */
let lastAutoBackup = 0;

export function startAutoBackup() {
  setInterval(() => {
    const hours = Number(getConfig().backup?.autoHours) || 0;
    if (hours <= 0) return;
    const due = lastAutoBackup + hours * 3.6e6;
    if (Date.now() >= due) {
      lastAutoBackup = Date.now();
      createBackup().catch((err) => console.error('[auto-backup]', err));
    }
  }, 60_000);
}

export const dataRouter = Router();
dataRouter.use('/admin/backups', requireAuth);

/** GET /api/v1/admin/backups 备份列表 */
dataRouter.get('/admin/backups', (_req, res) => {
  const items = fs.readdirSync(BACKUP_DIR)
    .filter((f) => f.endsWith('.zip'))
    .map((name) => {
      const stat = fs.statSync(path.join(BACKUP_DIR, name));
      return { name, size: stat.size, createdAt: stat.mtime.toISOString() };
    })
    .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1));
  res.json(items);
});

/** POST /api/v1/admin/backups 立即备份 */
dataRouter.post('/admin/backups', async (_req, res) => {
  const name = await createBackup();
  res.status(201).json({ name });
});

/** GET /api/v1/admin/backups/:name 下载 */
dataRouter.get('/admin/backups/:name', (req, res) => {
  const name = req.params.name;
  if (!/^backup-[\d-]+\.zip$/.test(name)) {
    res.status(400).json({ error: 'bad_name' });
    return;
  }
  const full = path.join(BACKUP_DIR, name);
  if (!fs.existsSync(full)) {
    res.status(404).json({ error: 'not_found' });
    return;
  }
  res.download(full);
});

/** DELETE /api/v1/admin/backups/:name */
dataRouter.delete('/admin/backups/:name', (req, res) => {
  const name = req.params.name;
  if (!/^backup-[\d-]+\.zip$/.test(name)) {
    res.status(400).json({ error: 'bad_name' });
    return;
  }
  fs.rmSync(path.join(BACKUP_DIR, name), { force: true });
  res.json({ ok: true });
});
