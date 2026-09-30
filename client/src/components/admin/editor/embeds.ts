import { Node, VueNodeViewRenderer, mergeAttributes } from '@tiptap/vue-3';
import Image from '@tiptap/extension-image';
import { joinSize, splitSize, type MediaAlign } from '../../../utils/mediaSize';
import ImageView from './ImageView.vue';
import type MarkdownIt from 'markdown-it';
import {
  defaultGallery, embedsPlugin, galleryToMarkdown, mediaToMarkdown, sanitizeGallery, safeMediaSrc, MEDIA_KINDS,
  type GalleryData, type MediaData, type MediaKind,
} from '../../../utils/embeds';
import MediaEmbedView from './MediaEmbedView.vue';
import GalleryView from './GalleryView.vue';

/**
 * 正文嵌入节点（Tiptap）：媒体（视频 / 音频 / 文件 / 压缩包）与拼图。
 * Markdown 读写走 tiptap-markdown 的 storage.markdown：解析时挂同一个 markdown-it 插件（utils/embeds.ts，editor 模式），
 * 序列化成 `@[kind](src "title")` 与 `::: gallery` 容器块，前台用同一插件（site 模式）渲染。
 * 交互（换文件、编辑拼图）由宿主通过 options 注入回调，节点视图调用。
 */

type MarkdownSerializerState = { write: (s: string) => void; closeBlock: (n: unknown) => void };

const parseSetup = (md: MarkdownIt): void => {
  md.use(embedsPlugin, { mode: 'editor' });
};

interface SizeAttrs { width: number | null; height: number | null; align: MediaAlign }

/** 尺寸 / 对齐属性：从地址片段（#w=&h=&a=）解析，节点里单独存，序列化时再拼回地址 */
function sizeAttrs(srcAttr: string) {
  const from = (el: HTMLElement) => splitSize(el.getAttribute(srcAttr) ?? '');
  return {
    width: { default: null, parseHTML: (el: HTMLElement) => from(el).w, renderHTML: () => ({}) },
    height: { default: null, parseHTML: (el: HTMLElement) => from(el).h, renderHTML: () => ({}) },
    align: { default: '', parseHTML: (el: HTMLElement) => from(el).align, renderHTML: () => ({}) },
  };
}

/**
 * 正文图片：可拖动改尺寸、可对齐（节点视图 ImageView）。
 * Markdown 仍是标准 `![说明](地址)`，尺寸写在地址片段里（见 utils/mediaSize.ts）。
 */
export const ResizableImage = Image.extend({
  draggable: true,

  addAttributes() {
    return {
      ...this.parent?.(),
      src: { default: null, parseHTML: (el: HTMLElement) => splitSize(el.getAttribute('src') ?? '').src },
      ...sizeAttrs('src'),
    };
  },

  addNodeView() {
    return VueNodeViewRenderer(ImageView);
  },

  addStorage() {
    return {
      markdown: {
        serialize(state: MarkdownSerializerState, node: { attrs: { src: string; alt: string | null; title: string | null } & SizeAttrs }) {
          const a = node.attrs;
          const alt = (a.alt ?? '').replace(/[[\]\\]/g, '\\$&');
          const src = joinSize(a.src, { w: a.width, h: a.height, align: a.align }).replace(/[\s()]/g, encodeURIComponent);
          const title = a.title ? ` "${a.title.replace(/"/g, '\\"')}"` : '';
          state.write(`![${alt}](${src}${title})`);
        },
        parse: {},
      },
    };
  },
});

export interface EmbedHooks {
  /** 换一个素材（按当前类型筛选）；回调拿到新的 src / 标题 */
  replaceMedia?: (kind: MediaKind, done: (next: Partial<MediaData>) => void) => void;
  /** 打开拼图编辑器 */
  editGallery?: (data: GalleryData, done: (next: GalleryData) => void) => void;
}

export const MediaEmbed = Node.create<EmbedHooks>({
  name: 'mediaEmbed',
  group: 'block',
  atom: true,
  draggable: true,
  selectable: true,

  addOptions() {
    return {};
  },

  addAttributes() {
    return {
      // 属性可能来自粘贴的外部 HTML：类型取枚举、地址须安全，否则置空
      kind: { default: 'file', parseHTML: (el) => ((MEDIA_KINDS as string[]).includes(el.getAttribute('data-kind') ?? '') ? el.getAttribute('data-kind') : 'file') },
      src: {
        default: '',
        parseHTML: (el) => {
          const src = splitSize(el.getAttribute('data-src') ?? '').src;
          return safeMediaSrc(src) ? src : '';
        },
      },
      title: { default: '', parseHTML: (el) => el.getAttribute('data-title') ?? '' },
      ...sizeAttrs('data-src'),
    };
  },

  parseHTML() {
    return [{ tag: 'div[data-embed="media"]' }];
  },

  renderHTML({ node, HTMLAttributes }) {
    return ['div', mergeAttributes(HTMLAttributes, {
      'data-embed': 'media', 'data-kind': node.attrs.kind, 'data-src': node.attrs.src, 'data-title': node.attrs.title,
    })];
  },

  addNodeView() {
    return VueNodeViewRenderer(MediaEmbedView);
  },

  addStorage() {
    return {
      markdown: {
        serialize(state: MarkdownSerializerState, node: { attrs: MediaData & SizeAttrs }) {
          const a = node.attrs;
          state.write(mediaToMarkdown({ ...a, src: joinSize(a.src, { w: a.width, h: a.height, align: a.align }) }));
          state.closeBlock(node);
        },
        parse: { setup: parseSetup },
      },
    };
  },
});

export const Gallery = Node.create<EmbedHooks>({
  name: 'gallery',
  group: 'block',
  atom: true,
  draggable: true,
  selectable: true,

  addOptions() {
    return {};
  },

  addAttributes() {
    return {
      data: {
        default: defaultGallery(),
        parseHTML: (el) => {
          try {
            return sanitizeGallery(JSON.parse(el.getAttribute('data-json') ?? '{}'));
          } catch {
            return defaultGallery();
          }
        },
        renderHTML: (attrs) => ({ 'data-json': JSON.stringify(attrs.data) }),
      },
    };
  },

  parseHTML() {
    return [{ tag: 'div[data-embed="gallery"]' }];
  },

  renderHTML({ HTMLAttributes }) {
    return ['div', mergeAttributes(HTMLAttributes, { 'data-embed': 'gallery' })];
  },

  addNodeView() {
    return VueNodeViewRenderer(GalleryView);
  },

  addStorage() {
    return {
      markdown: {
        serialize(state: MarkdownSerializerState, node: { attrs: { data: GalleryData } }) {
          state.write(galleryToMarkdown(node.attrs.data));
          state.closeBlock(node);
        },
        parse: { setup: parseSetup },
      },
    };
  },
});
