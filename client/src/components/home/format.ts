import { formatDate } from '../../utils/date';

/** 首页共用的小工具：日期拆月日、Markdown 转纯文本预览、GitHub 状态类型 */

export function monthDay(s: string): { m: number; d: number } {
  const [, m, d] = formatDate(s).split('-');
  return { m: Number(m), d: Number(d) };
}

/** Markdown → 单行纯文本（随想预览用；去图片、链接地址与标记符） */
export function plainText(md: string): string {
  return md
    .replace(/!\[[^\]]*]\([^)]*\)/g, '')
    .replace(/\[([^\]]*)]\([^)]*\)/g, '$1')
    .replace(/```[\s\S]*?```/g, '')
    .replace(/[#>*_`~]/g, '')
    .replace(/\s+/g, ' ')
    .trim();
}

/** 按字符串稳定取模（无封面时挑选 CSS 光影封面） */
export function hashOf(seed: string): number {
  let h = 0;
  for (const ch of seed) h = (h * 31 + ch.charCodeAt(0)) >>> 0;
  return h;
}

export interface GhActivity {
  type: string;
  repo: string;
  text: string;
  time: string;
}

export interface HeatDay {
  date: string;
  count: number;
  level: number;
}

export interface GhStatus {
  stats: { repos: number; stars: number; followers: number; commits: number };
  activities: GhActivity[];
  heatmap?: HeatDay[];
}
