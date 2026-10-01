import { useConfigStore } from '../stores/config';

/** SQLite datetime('now') 为 UTC；按配置时区格式化展示 */
function parseUtc(s: string): Date {
  return new Date(`${s.replace(' ', 'T')}Z`);
}

export function formatDate(s: string): string {
  return new Intl.DateTimeFormat('zh-CN', {
    timeZone: useConfigStore().cfg.timezone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(parseUtc(s)).replaceAll('/', '-');
}

export function formatDateTime(s: string, withSeconds = false): string {
  return new Intl.DateTimeFormat('zh-CN', {
    timeZone: useConfigStore().cfg.timezone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    ...(withSeconds ? { second: '2-digit' as const } : {}),
    hour12: false,
  }).format(parseUtc(s)).replaceAll('/', '-');
}

/** 站点时区的今天（YYYY-MM-DD）：替代 toISOString().slice(0, 10)（那是 UTC 日期，东八区零点到八点会差一天） */
export function siteToday(): string {
  return new Intl.DateTimeFormat('en-CA', {
    timeZone: useConfigStore().cfg.timezone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(new Date());
}
