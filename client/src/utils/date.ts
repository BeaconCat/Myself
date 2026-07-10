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

export function formatDateTime(s: string): string {
  return new Intl.DateTimeFormat('zh-CN', {
    timeZone: useConfigStore().cfg.timezone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(parseUtc(s)).replaceAll('/', '-');
}
