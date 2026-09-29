import type { IconNode } from 'lucide';

/**
 * 全站标准图标：取自离线安装的 Lucide（24 网格线性图标，按需引入可摇树）。
 * 各图标组件（SIcon / MaIcon / MIcon / UiIcon / ContentIcon / EngageIcon / KitIcon）维护自己的
 * 「名字 → Lucide 图标」映射，经这里转成 SVG 内部片段，外层 <svg> 统一 fill:none + stroke:currentColor。
 * 品牌图标另见 about/brands.ts（simple-icons）。
 */
const esc = (v: unknown): string => String(v).replace(/&/g, '&amp;').replace(/"/g, '&quot;');

export function iconMarkup(node: IconNode): string {
  return node
    .map(([tag, attrs]) => `<${tag} ${Object.entries(attrs).map(([k, v]) => `${k}="${esc(v)}"`).join(' ')}/>`)
    .join('');
}

/** 名字 → Lucide 图标 的映射整体转成 名字 → SVG 片段 */
export function iconSet<K extends string>(map: Record<K, IconNode>): Record<K, string> {
  const out = {} as Record<K, string>;
  for (const k of Object.keys(map) as K[]) out[k] = iconMarkup(map[k]);
  return out;
}
