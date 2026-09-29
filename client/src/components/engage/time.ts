import { i18n } from '../../i18n';

/** 服务端 UTC 时间（'YYYY-MM-DD HH:MM:SS'）→ 口语化相对时间：刚刚 / n 分钟前 / n 小时前 / n 天前 / M月D日 */
export function relTime(utc: string): string {
  const t = i18n.global.t;
  const ts = new Date(`${utc.replace(' ', 'T')}Z`).getTime();
  const diff = Date.now() - ts;
  const min = Math.floor(diff / 6e4);
  if (min < 1) return t('comments.justNow');
  if (min < 60) return t('comments.minAgo', { n: min });
  const h = Math.floor(min / 60);
  if (h < 24) return t('comments.hourAgo', { n: h });
  const d = Math.floor(h / 24);
  if (d < 7) return t('comments.dayAgo', { n: d });
  const dt = new Date(ts);
  return dt.getFullYear() === new Date().getFullYear()
    ? `${dt.getMonth() + 1}月${dt.getDate()}日`
    : `${dt.getFullYear()}年${dt.getMonth() + 1}月${dt.getDate()}日`;
}
