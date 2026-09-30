/**
 * 正文图片 / 视频 / 音频的尺寸与对齐，写在地址的片段里：`/uploads/a.webp#w=480&h=320&a=center`。
 * - 片段不会发往服务器，其它 Markdown 渲染器照常显示原图，不破坏可移植性；
 * - w / h 为设计宽高（像素），只给 w 时按原比例；h 只在自由变换后出现（前台用 aspect-ratio 自适应缩放）；
 * - a 为对齐：left / center / right；缺省为行内（图片）或整行（视频 / 音频）。
 */
export type MediaAlign = '' | 'left' | 'center' | 'right';

export interface MediaSize {
  w: number | null;
  h: number | null;
  align: MediaAlign;
}

const ALIGNS: MediaAlign[] = ['left', 'center', 'right'];

/** 拆出干净地址与尺寸；其它片段（如视频的 #t=）原样保留在地址里 */
export function splitSize(raw: string): { src: string } & MediaSize {
  const i = raw.indexOf('#');
  const out: { src: string } & MediaSize = { src: raw, w: null, h: null, align: '' };
  if (i < 0) return out;
  const params = new URLSearchParams(raw.slice(i + 1));
  if (!params.has('w') && !params.has('h') && !params.has('a')) return out;
  const num = (k: string): number | null => {
    const n = Math.round(Number(params.get(k)));
    return Number.isFinite(n) && n >= 16 && n <= 4000 ? n : null;
  };
  out.w = num('w');
  out.h = out.w ? num('h') : null;
  const a = params.get('a') as MediaAlign;
  out.align = ALIGNS.includes(a) ? a : '';
  params.delete('w');
  params.delete('h');
  params.delete('a');
  const rest = params.toString();
  out.src = raw.slice(0, i) + (rest ? `#${rest}` : '');
  return out;
}

/** 把尺寸写回地址片段（没有尺寸与对齐时返回原地址） */
export function joinSize(src: string, s: MediaSize): string {
  const parts: string[] = [];
  if (s.w) parts.push(`w=${Math.round(s.w)}`);
  if (s.w && s.h) parts.push(`h=${Math.round(s.h)}`);
  if (s.align) parts.push(`a=${s.align}`);
  if (!parts.length) return src;
  return `${src}${src.includes('#') ? '&' : '#'}${parts.join('&')}`;
}

/** 前台 / 节点视图共用的内联样式：宽度封顶 100%，自由变换过的按宽高比缩放（object-fit: cover） */
export function sizeStyle(s: MediaSize): string {
  const out: string[] = [];
  if (s.w) out.push(`width:${s.w}px`, 'max-width:100%');
  if (s.w && s.h) out.push(`aspect-ratio:${s.w} / ${s.h}`, 'height:auto', 'object-fit:cover');
  return out.join(';');
}
