/**
 * 正文嵌入块（Markdown 扩展语法），编辑器与前台共用同一套解析：
 *
 * 1. 媒体：独占一行
 *      @[video](/uploads/x.mp4 "标题")      视频播放器
 *      @[audio](/uploads/x.mp3 "标题")      音频播放器
 *      @[file](/uploads/x.pdf "名称")       文件卡片（下载）
 *      @[archive](/uploads/x.zip "名称")    压缩包卡片（下载 + 在线预览目录）
 *
 * 2. 拼图：容器块，参数可省略
 *      ::: gallery layout=grid cols=3 ratio=4:3 gap=m
 *      ![说明](/uploads/a.webp)
 *      ![说明](/uploads/b.webp)
 *      图注（非图片的行合并为图注）
 *      :::
 *
 * 地址只放行站内路径（/uploads/…）与 https 外链；其余整块丢弃。
 * mode = 'editor' 时输出带 data-* 的占位元素，交给 Tiptap 节点解析；mode = 'site' 输出前台最终 HTML。
 */
import type MarkdownIt from 'markdown-it';
import type StateBlock from 'markdown-it/lib/rules_block/state_block.mjs';
import { iconMarkup } from './lucide';
import { splitSize } from './mediaSize';
import { Archive, Download, FileText, Music, SquareArrowOutUpRight } from 'lucide';

export type MediaKind = 'video' | 'audio' | 'file' | 'archive';
export const MEDIA_KINDS: MediaKind[] = ['video', 'audio', 'file', 'archive'];

export type GalleryLayout = 'grid' | 'mosaic' | 'row' | 'masonry';
export type GalleryRatio = 'auto' | '1:1' | '4:3' | '3:2' | '16:9' | '3:4';
export type GalleryGap = 'none' | 's' | 'm' | 'l';

export interface GalleryImage { src: string; alt: string }
export interface GalleryData {
  images: GalleryImage[];
  layout: GalleryLayout;
  cols: number;
  ratio: GalleryRatio;
  gap: GalleryGap;
  caption: string;
}

export interface MediaData { kind: MediaKind; src: string; title: string }

export const GALLERY_LAYOUTS: GalleryLayout[] = ['grid', 'mosaic', 'row', 'masonry'];
export const GALLERY_RATIOS: GalleryRatio[] = ['auto', '1:1', '4:3', '3:2', '16:9', '3:4'];
export const GALLERY_GAPS: GalleryGap[] = ['none', 's', 'm', 'l'];
export const GAP_PX: Record<GalleryGap, number> = { none: 0, s: 4, m: 8, l: 14 };

export function defaultGallery(images: GalleryImage[] = []): GalleryData {
  return { images, layout: 'grid', cols: images.length === 2 ? 2 : 3, ratio: '4:3', gap: 'm', caption: '' };
}

/** 站内路径或 https 外链才可用 */
export function safeMediaSrc(src: string): boolean {
  const s = src.trim();
  // 站内路径：单个 / 开头，且第二个字符不是 / 或 \（浏览器会把 /\host 当成 //host）
  return /^\/(?![/\\])/.test(s) || /^https:\/\//i.test(s);
}

/** 媒体数据白名单化：类型只取枚举值，地址须安全，标题转成字符串（来自粘贴的 HTML 时不可信） */
export function sanitizeMedia(raw: Partial<Record<keyof MediaData, unknown>>): MediaData {
  const kind = (MEDIA_KINDS as string[]).includes(String(raw.kind)) ? (raw.kind as MediaKind) : 'file';
  const src = typeof raw.src === 'string' && safeMediaSrc(raw.src) ? raw.src.trim() : '';
  return { kind, src, title: typeof raw.title === 'string' ? raw.title.slice(0, 200) : '' };
}

/** 拼图数据白名单化：布局 / 比例 / 间距取枚举值，列数取 1–6 的整数，图片地址须安全 */
export function sanitizeGallery(raw: unknown): GalleryData {
  const r = (raw && typeof raw === 'object' ? raw : {}) as Record<string, unknown>;
  const images = (Array.isArray(r.images) ? r.images : [])
    .map((im) => (im && typeof im === 'object' ? im : {}) as Record<string, unknown>)
    .filter((im) => typeof im.src === 'string' && safeMediaSrc(im.src))
    .slice(0, 60)
    .map((im) => ({ src: String(im.src).trim(), alt: typeof im.alt === 'string' ? im.alt.slice(0, 200) : '' }));
  const pick = <T extends string>(v: unknown, list: readonly T[], def: T): T => ((list as readonly string[]).includes(String(v)) ? (v as T) : def);
  const cols = Math.round(Number(r.cols));
  return {
    images,
    layout: pick(r.layout, GALLERY_LAYOUTS, 'grid'),
    cols: Number.isFinite(cols) ? Math.min(6, Math.max(1, cols)) : 3,
    ratio: pick(r.ratio, GALLERY_RATIOS, '4:3'),
    gap: pick(r.gap, GALLERY_GAPS, 'm'),
    caption: typeof r.caption === 'string' ? r.caption.slice(0, 300) : '',
  };
}

function esc(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

/* ---------- 序列化（编辑器 → Markdown） ---------- */

function quote(title: string): string {
  const t = title.replace(/"/g, '”').replace(/[\r\n]+/g, ' ').trim();
  return t ? ` "${t}"` : '';
}

export function mediaToMarkdown(d: MediaData): string {
  return `@[${d.kind}](${d.src.replace(/\s/g, '%20')}${quote(d.title)})`;
}

export function galleryToMarkdown(d: GalleryData): string {
  const params = [`layout=${d.layout}`, `cols=${d.cols}`, `ratio=${d.ratio}`, `gap=${d.gap}`].join(' ');
  const lines = d.images.map((im) => `![${im.alt.replace(/[[\]\r\n]/g, ' ').trim()}](${im.src.replace(/\s/g, '%20')})`);
  const cap = d.caption.replace(/[\r\n]+/g, ' ').trim();
  return [`::: gallery ${params}`, ...lines, ...(cap ? [cap] : []), ':::'].join('\n');
}

/* ---------- 解析 ---------- */

const MEDIA_RE = /^@\[(video|audio|file|archive)\]\((\S+?)(?:\s+"([^"]*)")?\)\s*$/;
const IMG_RE = /^!\[([^\]]*)\]\((\S+?)(?:\s+"[^"]*")?\)\s*$/;
const GALLERY_OPEN = /^:::\s*gallery\b(.*)$/;
const FENCE_CLOSE = /^:::\s*$/;

export function parseGalleryParams(raw: string): Omit<GalleryData, 'images' | 'caption'> {
  const out = { layout: 'grid' as GalleryLayout, cols: 3, ratio: '4:3' as GalleryRatio, gap: 'm' as GalleryGap };
  for (const m of raw.matchAll(/(\w+)=(\S+)/g)) {
    const [, k, v] = m;
    if (k === 'layout' && (GALLERY_LAYOUTS as string[]).includes(v)) out.layout = v as GalleryLayout;
    if (k === 'cols') out.cols = Math.min(6, Math.max(1, Number(v) || 3));
    if (k === 'ratio' && (GALLERY_RATIOS as string[]).includes(v)) out.ratio = v as GalleryRatio;
    if (k === 'gap' && (GALLERY_GAPS as string[]).includes(v)) out.gap = v as GalleryGap;
  }
  return out;
}

function lineAt(state: StateBlock, line: number): string {
  return state.src.slice(state.bMarks[line] + state.tShift[line], state.eMarks[line]);
}

/* ---------- 前台渲染 ---------- */

const svg = (node: Parameters<typeof iconMarkup>[0], cls: string): string =>
  `<svg class="${cls}" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${iconMarkup(node)}</svg>`;

function extOf(src: string): string {
  const m = /\.([a-z0-9]{1,10})(?:[?#].*)?$/i.exec(src);
  return m ? m[1].toLowerCase() : '';
}

function nameOf(src: string): string {
  try {
    return decodeURIComponent(src.split(/[?#]/)[0].split('/').pop() ?? src);
  } catch {
    return src;
  }
}

export interface EmbedLabels { download: string; preview: string; open: string }

export function renderMediaHtml(input: MediaData, labels: EmbedLabels): string {
  const raw = sanitizeMedia(input);
  if (!raw.src) return '';
  // 尺寸 / 对齐写在地址片段里：视频按宽高比缩放，音频只定宽，文件卡片整行
  const sz = splitSize(raw.src);
  const d = { ...raw, src: sz.src };
  const src = esc(d.src);
  const al = sz.align ? ` al-${sz.align}` : '';
  const external = !d.src.startsWith('/');
  const ref = external ? ' referrerpolicy="no-referrer"' : '';
  const title = esc(d.title);
  if (d.kind === 'video') {
    const fig = sz.w ? ` style="width:${sz.w}px;max-width:100%"` : '';
    const vid = sz.w && sz.h ? ` style="aspect-ratio:${sz.w} / ${sz.h};object-fit:cover"` : '';
    return `<figure class="md-media md-video${al}"${fig}><video src="${src}" controls playsinline preload="metadata"${ref}${vid}></video>${title ? `<figcaption>${title}</figcaption>` : ''}</figure>`;
  }
  if (d.kind === 'audio') {
    const fig = sz.w ? ` style="width:${sz.w}px;max-width:100%"` : '';
    return `<figure class="md-media md-audio${al}"${fig}><span class="ic">${svg(Music, 'md-i')}</span><div class="bd"><b>${title || esc(nameOf(d.src))}</b><audio src="${src}" controls preload="metadata"${ref}></audio></div></figure>`;
  }
  const ext = extOf(d.src);
  const name = title || esc(nameOf(d.src));
  const icon = d.kind === 'archive' ? svg(Archive, 'md-i') : svg(FileText, 'md-i');
  const preview = d.kind === 'archive' && !external && ext === 'zip'
    ? `<button type="button" class="md-file-act" data-archive="${src}" data-title="${name}">${svg(SquareArrowOutUpRight, 'md-i sm')}${esc(labels.preview)}</button>`
    : '';
  return `<div class="md-media md-file" data-kind="${esc(d.kind)}"><span class="ic">${icon}${ext ? `<em>${esc(ext)}</em>` : ''}</span><div class="bd"><b>${name}</b><small>${esc(ext.toUpperCase())}</small></div>${preview}<a class="md-file-act" href="${src}" download${external ? ' target="_blank" rel="noopener noreferrer nofollow"' : ''}>${svg(Download, 'md-i sm')}${esc(labels.download)}</a></div>`;
}

export function renderGalleryHtml(input: GalleryData): string {
  const d = sanitizeGallery(input);
  const n = d.images.length;
  if (!n) return '';
  const ratio = d.ratio === 'auto' ? 'auto' : d.ratio.replace(':', ' / ');
  const style = `--cols:${Math.min(d.cols, n)};--gap:${GAP_PX[d.gap]}px;--ratio:${ratio}`;
  const imgs = d.images
    .map((im, i) => {
      const ext = !im.src.startsWith('/');
      return `<div class="gi" style="--i:${i}"><img src="${esc(im.src)}" alt="${esc(im.alt)}" loading="lazy" decoding="async"${ext ? ' referrerpolicy="no-referrer"' : ''} /></div>`;
    })
    .join('');
  return `<figure class="md-gallery" data-layout="${d.layout}" data-n="${Math.min(n, 9)}" style="${style}"><div class="gg">${imgs}</div>${d.caption ? `<figcaption>${esc(d.caption)}</figcaption>` : ''}</figure>`;
}

/* ---------- markdown-it 插件 ---------- */

export interface EmbedPluginOptions {
  mode: 'editor' | 'site';
  labels?: () => EmbedLabels;
}

export function embedsPlugin(md: MarkdownIt, opts: EmbedPluginOptions): void {
  const flag = md as unknown as { __embeds?: boolean };
  if (flag.__embeds) return;
  flag.__embeds = true;

  md.block.ruler.before('paragraph', 'embed_media', (state, startLine, _endLine, silent) => {
    const m = MEDIA_RE.exec(lineAt(state, startLine));
    if (!m || !safeMediaSrc(m[2])) return false;
    if (silent) return true;
    const token = state.push('embed_media', '', 0);
    token.meta = { kind: m[1] as MediaKind, src: m[2], title: m[3] ?? '' } satisfies MediaData;
    token.map = [startLine, startLine + 1];
    state.line = startLine + 1;
    return true;
  });

  md.block.ruler.before('fence', 'embed_gallery', (state, startLine, endLine, silent) => {
    const open = GALLERY_OPEN.exec(lineAt(state, startLine));
    if (!open) return false;
    let line = startLine + 1;
    const images: GalleryImage[] = [];
    const caption: string[] = [];
    let closed = false;
    for (; line < endLine; line++) {
      const text = lineAt(state, line).trim();
      if (FENCE_CLOSE.test(text)) {
        closed = true;
        break;
      }
      if (!text) continue;
      const im = IMG_RE.exec(text);
      if (im) {
        if (safeMediaSrc(im[2])) images.push({ alt: im[1], src: im[2] });
      } else caption.push(text);
    }
    if (!closed) return false;
    if (silent) return true;
    const token = state.push('embed_gallery', '', 0);
    token.meta = { ...parseGalleryParams(open[1]), images, caption: caption.join(' ') } satisfies GalleryData;
    token.map = [startLine, line + 1];
    state.line = line + 1;
    return true;
  });

  md.renderer.rules.embed_media = (tokens, idx) => {
    const d = tokens[idx].meta as MediaData;
    if (opts.mode === 'editor') {
      return `<div data-embed="media" data-kind="${esc(d.kind)}" data-src="${esc(d.src)}" data-title="${esc(d.title)}"></div>\n`;
    }
    return `${renderMediaHtml(d, opts.labels?.() ?? { download: 'Download', preview: 'Preview', open: 'Open' })}\n`;
  };

  md.renderer.rules.embed_gallery = (tokens, idx) => {
    const d = tokens[idx].meta as GalleryData;
    if (opts.mode === 'editor') return `<div data-embed="gallery" data-json="${esc(JSON.stringify(d))}"></div>\n`;
    return `${renderGalleryHtml(d)}\n`;
  };
}
