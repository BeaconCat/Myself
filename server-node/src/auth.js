import crypto from 'node:crypto';
import jwt from 'jsonwebtoken';
import { db } from './db.js';

const TOKEN_TTL = '30d';

/** 设置表：管理员凭据与 JWT 密钥（首次启动自动初始化） */
db.exec(`
  CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
  );
`);

function getSetting(key) {
  return db.prepare('SELECT value FROM settings WHERE key = ?').get(key)?.value;
}

function setSetting(key, value) {
  db.prepare(
    'INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value',
  ).run(key, value);
}

export function hashPassword(password) {
  const salt = crypto.randomBytes(16).toString('hex');
  const hash = crypto.scryptSync(password, salt, 64).toString('hex');
  return `${salt}:${hash}`;
}

export function verifyPassword(password, stored) {
  const [salt, hash] = String(stored).split(':');
  if (!salt || !hash) return false;
  const candidate = crypto.scryptSync(password, salt, 64);
  return crypto.timingSafeEqual(candidate, Buffer.from(hash, 'hex'));
}

/** 初始化默认管理员与密钥 */
export function initAuth() {
  if (!getSetting('jwt_secret')) {
    setSetting('jwt_secret', crypto.randomBytes(32).toString('hex'));
  }
  if (!getSetting('admin_username')) {
    const defaultPassword = 'myself-admin';
    setSetting('admin_username', 'admin');
    setSetting('admin_password', hashPassword(defaultPassword));
    console.log(
      `[myself-server] admin account initialized -> admin / ${defaultPassword} (请尽快在后台修改密码)`,
    );
  }
}

export function login(username, password) {
  if (username !== getSetting('admin_username')) return null;
  if (!verifyPassword(password, getSetting('admin_password'))) return null;
  return jwt.sign({ sub: username, role: 'admin' }, getSetting('jwt_secret'), {
    expiresIn: TOKEN_TTL,
  });
}

export function changePassword(oldPassword, newPassword) {
  if (!verifyPassword(oldPassword, getSetting('admin_password'))) return false;
  setSetting('admin_password', hashPassword(newPassword));
  return true;
}

/** Express 中间件：Bearer Token 校验 */
export function requireAuth(req, res, next) {
  const header = req.headers.authorization ?? '';
  const token = header.startsWith('Bearer ') ? header.slice(7) : '';
  try {
    req.admin = jwt.verify(token, getSetting('jwt_secret'));
    next();
  } catch {
    res.status(401).json({ error: 'unauthorized' });
  }
}
