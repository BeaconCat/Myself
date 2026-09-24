/** 移动后台格式化工具：时间解析、相对时间、体积、Markdown 纯文本摘要、slug。 */

/** 兼容 SQLite `YYYY-MM-DD HH:MM:SS`（UTC）与 ISO 字符串 */
export function parseTime(s: string): Date {
  if (!s) return new Date(NaN);
  if (/[TZ]|[+-]\d\d:?\d\d$/.test(s)) return new Date(s);
  return new Date(`${s.replace(' ', 'T')}Z`);
}

type T = (key: string, named?: Record<string, unknown>) => string;

/** 相对时间：刚刚 / n 分钟前 / n 小时前 / 昨天 / M 月 D 日 */
export function relTime(s: string, t: T, now = Date.now()): string {
  const d = parseTime(s);
  const ms = now - d.getTime();
  if (!Number.isFinite(ms)) return '';
  const min = Math.floor(ms / 60000);
  if (min < 1) return t('mobileAdmin.time.now');
  if (min < 60) return t('mobileAdmin.time.min', { n: min });
  const h = Math.floor(min / 60);
  if (h < 24) return t('mobileAdmin.time.hour', { n: h });
  if (h < 48) return t('mobileAdmin.time.yesterday');
  const nowY = new Date(now).getFullYear();
  if (d.getFullYear() !== nowY) return t('mobileAdmin.time.ymd', { y: d.getFullYear(), m: d.getMonth() + 1, d: d.getDate() });
  return t('mobileAdmin.time.md', { m: d.getMonth() + 1, d: d.getDate() });
}

export function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(bytes < 10 * 1024 * 1024 ? 1 : 0)} MB`;
  return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GB`;
}

/** Markdown → 单行纯文本（列表/摘要用） */
export function mdPlain(md: string): string {
  return md
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/!\[[^\]]*\]\([^)]*\)/g, ' ')
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/[#>*_`~-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

/** 标题 → slug（仅 ascii；中文标题回落时间戳） */
export function toSlug(title: string): string {
  const s = title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 60);
  return s || `post-${Date.now().toString(36)}`;
}

export const SLUG_RE = /^[a-z0-9-]{1,80}$/;

/** 按亮度决定主色上的文字颜色（秋季黄等浅主色用深字） */
export function onColor(hex: string): string {
  const v = hex.replace('#', '');
  if (v.length < 6) return '#fff';
  const [r, g, b] = [0, 2, 4].map((i) => parseInt(v.slice(i, i + 2), 16) / 255);
  const lin = (c: number) => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4);
  const L = 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b);
  return L > 0.45 ? '#141414' : '#ffffff';
}
