/** Studio 格式化工具：体积、时间、节气、问候 */
import { i18n } from '../../../i18n';
import './i18n';

const tr = (k: string, v?: Record<string, unknown>): string => (v ? i18n.global.t(`studio.time.${k}`, v) : i18n.global.t(`studio.time.${k}`));

export function formatSize(bytes: number): string {
  if (bytes >= 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GB`;
  if (bytes >= 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
  // 0 字节显示 0；有内容但不足 1 KB 时按 1 KB 显示
  if (bytes <= 0) return '0 KB';
  return `${Math.max(1, Math.round(bytes / 1024))} KB`;
}

/** 体积拆成 [数值, 单位]，用于大号数字 + 小单位排版 */
export function sizeParts(bytes: number): [string, string] {
  const [n, u] = formatSize(bytes).split(' ');
  return [n, u];
}

export function formatNumber(n: number): string {
  return n.toLocaleString('en-US');
}

/**
 * 后端时间：SQLite datetime('now')（UTC，"YYYY-MM-DD HH:MM:SS"）或 RFC3339。
 * 统一解析为 Date。
 */
export function parseTime(s: string | null | undefined): Date | null {
  if (!s) return null;
  const iso = /[zZ]|[+-]\d\d:?\d\d$/.test(s) ? s : `${s.replace(' ', 'T')}Z`;
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? null : d;
}

const pad = (n: number) => String(n).padStart(2, '0');

export function ymd(d: Date): string {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

export function dateText(s: string | null | undefined): string {
  const d = parseTime(s);
  return d ? ymd(d) : '';
}

export function dateTimeText(s: string | null | undefined): string {
  const d = parseTime(s);
  return d ? `${ymd(d)} ${pad(d.getHours())}:${pad(d.getMinutes())}` : '';
}

/** 相对时间：刚刚 / x 分钟前 / x 小时前 / 昨天 / x 天前 / 日期 */
export function relTime(s: string | null | undefined, now = new Date()): string {
  const d = parseTime(s);
  if (!d) return '';
  const sec = Math.round((now.getTime() - d.getTime()) / 1000);
  if (sec < 60) return tr('justNow');
  if (sec < 3600) return tr('minAgo', { n: Math.floor(sec / 60) });
  if (sec < 86400) return tr('hourAgo', { n: Math.floor(sec / 3600) });
  const days = Math.floor((startOfDay(now).getTime() - startOfDay(d).getTime()) / 86400000);
  if (days <= 1) return tr('yesterday');
  if (days < 30) return tr('dayAgo', { n: days });
  return d.getFullYear() === now.getFullYear() ? ymd(d).slice(5) : ymd(d);
}

export function startOfDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate());
}

export const WEEKDAYS = tr('weekdays').split(',');
export const WEEK_SHORT = tr('weekShort').split(',');

/* ===== 二十四节气（21 世纪通式：[Y×D + C] − L） ===== */
const TERMS: [string, number][] = [
  ['小寒', 5.4055], ['大寒', 20.12], ['立春', 3.87], ['雨水', 18.73],
  ['惊蛰', 5.63], ['春分', 20.646], ['清明', 4.81], ['谷雨', 20.1],
  ['立夏', 5.52], ['小满', 21.04], ['芒种', 5.678], ['夏至', 21.37],
  ['小暑', 7.108], ['大暑', 22.83], ['立秋', 7.5], ['处暑', 23.13],
  ['白露', 7.646], ['秋分', 23.042], ['寒露', 8.318], ['霜降', 23.438],
  ['立冬', 7.438], ['小雪', 22.36], ['大雪', 7.18], ['冬至', 21.94],
];

function termDate(year: number, index: number): Date {
  const y = year % 100;
  const month = Math.floor(index / 2);
  const leapBase = index < 4 ? y - 1 : y;
  const day = Math.floor(y * 0.2422 + TERMS[index][1]) - Math.floor(leapBase / 4);
  return new Date(year, month, day);
}

/** 当前所处节气：{ name, days }，days=0 表示今天正是该节气 */
export function solarTerm(now = new Date()): { name: string; days: number } {
  const today = startOfDay(now);
  let best = { name: TERMS[23][0], date: termDate(now.getFullYear() - 1, 23) };
  for (let i = 0; i < 24; i += 1) {
    const d = termDate(now.getFullYear(), i);
    if (d.getTime() <= today.getTime()) best = { name: TERMS[i][0], date: d };
  }
  return { name: best.name, days: Math.round((today.getTime() - best.date.getTime()) / 86400000) };
}

export function greeting(hour: number): string {
  if (hour < 5) return tr('lateNight');
  if (hour < 11) return tr('morning');
  if (hour < 14) return tr('noon');
  if (hour < 18) return tr('afternoon');
  return tr('evening');
}

/** 首字（头像占位） */
export function initial(name: string): string {
  return (name.trim()[0] ?? 'M').toUpperCase();
}

/** Markdown → 纯文本摘要（时间线 / 卡片） */
export function plainText(md: string): string {
  return md
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/!\[[^\]]*\]\([^)]*\)/g, ' ')
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/[#>*_`~|-]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

/** 中文字数（去空白） */
export function wordCount(md: string): number {
  return plainText(md).replace(/\s/g, '').length;
}
