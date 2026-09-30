/**
 * 全站唯一的 Markdown 渲染管线：markdown-it + 任务列表 + 代码高亮（highlight.js，离线）
 * + 标题锚点。文章、随想、后台预览共用，避免各处配置漂移。
 */
import MarkdownIt from 'markdown-it';
import type StateCore from 'markdown-it/lib/rules_core/state_core.mjs';
import type Token from 'markdown-it/lib/token.mjs';
// @ts-expect-error 无类型声明的离线插件
import taskLists from 'markdown-it-task-lists';
import hljs from 'highlight.js/lib/core';
import bash from 'highlight.js/lib/languages/bash';
import c from 'highlight.js/lib/languages/c';
import cpp from 'highlight.js/lib/languages/cpp';
import css from 'highlight.js/lib/languages/css';
import diff from 'highlight.js/lib/languages/diff';
import go from 'highlight.js/lib/languages/go';
import ini from 'highlight.js/lib/languages/ini';
import java from 'highlight.js/lib/languages/java';
import javascript from 'highlight.js/lib/languages/javascript';
import json from 'highlight.js/lib/languages/json';
import kotlin from 'highlight.js/lib/languages/kotlin';
import markdown from 'highlight.js/lib/languages/markdown';
import python from 'highlight.js/lib/languages/python';
import rust from 'highlight.js/lib/languages/rust';
import scss from 'highlight.js/lib/languages/scss';
import shell from 'highlight.js/lib/languages/shell';
import sql from 'highlight.js/lib/languages/sql';
import swift from 'highlight.js/lib/languages/swift';
import typescript from 'highlight.js/lib/languages/typescript';
import xml from 'highlight.js/lib/languages/xml';
import yaml from 'highlight.js/lib/languages/yaml';
import { i18n } from '../i18n';
import { ArrowUpRight } from 'lucide';
import { iconMarkup } from './lucide';
import { embedsPlugin } from './embeds';
import { sizeStyle, splitSize } from './mediaSize';

const LANGS: Record<string, Parameters<typeof hljs.registerLanguage>[1]> = {
  bash, c, cpp, css, diff, go, ini, java, javascript, json, kotlin, markdown,
  python, rust, scss, shell, sql, swift, typescript, xml, yaml,
};
for (const [name, def] of Object.entries(LANGS)) hljs.registerLanguage(name, def);
hljs.registerAliases(['js', 'mjs', 'cjs'], { languageName: 'javascript' });
hljs.registerAliases(['ts'], { languageName: 'typescript' });
hljs.registerAliases(['sh', 'zsh'], { languageName: 'bash' });
hljs.registerAliases(['html', 'vue', 'svg'], { languageName: 'xml' });
hljs.registerAliases(['yml'], { languageName: 'yaml' });
hljs.registerAliases(['py'], { languageName: 'python' });
hljs.registerAliases(['kt'], { languageName: 'kotlin' });
hljs.registerAliases(['rs'], { languageName: 'rust' });
hljs.registerAliases(['md'], { languageName: 'markdown' });

/** 目录项：level 为 2/3，id 与渲染出的标题锚点一致 */
export interface TocItem {
  id: string;
  text: string;
  level: number;
}

function slugify(text: string, used: Map<string, number>): string {
  const base = text
    .toLowerCase()
    .trim()
    .replace(/[\s]+/g, '-')
    .replace(/[^\p{L}\p{N}-]/gu, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '') || 'section';
  const n = used.get(base) ?? 0;
  used.set(base, n + 1);
  return n ? `${base}-${n}` : base;
}

function highlight(code: string, lang: string): string {
  const language = lang && hljs.getLanguage(lang) ? lang : '';
  const html = language
    ? hljs.highlight(code, { language, ignoreIllegals: true }).value
    : escapeHtml(code);
  const label = language || 'text';
  // 一键复制：按钮由全局点击委托处理（utils/codeCopy.ts）
  const copy = escapeHtml(i18n.global.t('content.article.copyCode'));
  return `<pre class="hljs" data-lang="${label}"><button type="button" class="code-copy" aria-label="${copy}" title="${copy}"></button><code class="language-${label}">${html}</code></pre>`;
}

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function createRenderer(): MarkdownIt {
  const md: MarkdownIt = new MarkdownIt({ linkify: true, highlight }).use(taskLists);

  /* 嵌入块：拼图、视频 / 音频 / 文件 / 压缩包（语法见 utils/embeds.ts） */
  md.use(embedsPlugin, {
    mode: 'site',
    labels: () => ({
      download: i18n.global.t('content.embed.download'),
      preview: i18n.global.t('content.embed.preview'),
      open: i18n.global.t('content.embed.open'),
    }),
  });

  /* 标题锚点：渲染时收集 TOC 到 env.toc */
  md.core.ruler.push('heading_anchor', (state: StateCore) => {
    const env = state.env as { toc?: TocItem[]; used?: Map<string, number> };
    env.used ??= new Map();
    const { tokens } = state;
    for (let i = 0; i < tokens.length; i++) {
      const open = tokens[i];
      if (open.type !== 'heading_open') continue;
      const inline = tokens[i + 1];
      const text = inline?.children
        ?.filter((tk: Token) => tk.type === 'text' || tk.type === 'code_inline')
        .map((tk: Token) => tk.content)
        .join('') ?? '';
      const id = slugify(text, env.used);
      open.attrSet('id', id);
      const level = Number(open.tag.slice(1));
      if (env.toc && (level === 2 || level === 3)) env.toc.push({ id, text, level });
    }
  });

  /*
   * 图片：只放行站内路径与 https 外链（html:false 已杜绝原始 HTML，validateLink 已挡 javascript: 等协议）；
   * 外链图片不带来源页（referrerpolicy=no-referrer），懒加载、异步解码。其余地址整张图片丢弃。
   */
  const defaultImage = md.renderer.rules.image!;
  md.renderer.rules.image = (tokens, idx, options, env, self) => {
    const token = tokens[idx];
    // 地址片段里的尺寸 / 对齐（编辑器拖动改尺寸写入，见 utils/mediaSize.ts）
    const sized = splitSize((token.attrGet('src') ?? '').trim());
    const src = sized.src;
    token.attrSet('src', src);
    const style = sizeStyle(sized);
    if (style) token.attrSet('style', style);
    if (sized.align) token.attrJoin('class', `md-img al-${sized.align}`);
    const local = /^\/(?![/\\])/.test(src) || src.startsWith('data:image/');
    if (!local && !/^https:\/\//i.test(src)) return '';
    if (!local) token.attrSet('referrerpolicy', 'no-referrer');
    token.attrSet('loading', 'lazy');
    token.attrSet('decoding', 'async');
    return defaultImage(tokens, idx, options, env, self);
  };

  /* 链接：外链新开标签页且不回传 opener / 来源页 */
  const defaultLinkOpen = md.renderer.rules.link_open ?? ((tokens, idx, options, _env, self) => self.renderToken(tokens, idx, options));
  md.renderer.rules.link_open = (tokens, idx, options, env, self) => {
    const href = tokens[idx].attrGet('href') ?? '';
    if (/^https?:\/\//i.test(href)) {
      tokens[idx].attrSet('target', '_blank');
      tokens[idx].attrSet('rel', 'noopener noreferrer nofollow');
      tokens[idx].attrJoin('class', 'ext');
      tokens[idx].meta = { ...(tokens[idx].meta ?? {}), external: true };
    }
    return defaultLinkOpen(tokens, idx, options, env, self);
  };

  /* 外链末尾一枚小箭头（Lucide ArrowUpRight），提示会离开本站；自动识别出的裸链接同样处理 */
  const extIcon = `<svg class="ext-i" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${iconMarkup(ArrowUpRight)}</svg>`;
  const defaultLinkClose = md.renderer.rules.link_close ?? ((tokens, idx, options, _env, self) => self.renderToken(tokens, idx, options));
  md.renderer.rules.link_close = (tokens, idx, options, env, self) => {
    let open = idx - 1;
    while (open >= 0 && tokens[open].type !== 'link_open') open--;
    const ext = open >= 0 && (tokens[open].meta as { external?: boolean } | null)?.external;
    return (ext ? extIcon : '') + defaultLinkClose(tokens, idx, options, env, self);
  };
  return md;
}

export const md = createRenderer();

/** 渲染正文并同时返回目录（仅收集 h2/h3） */
export function renderWithToc(source: string): { html: string; toc: TocItem[] } {
  const env = { toc: [] as TocItem[] };
  const html = md.render(source, env);
  return { html, toc: env.toc };
}

/** 普通渲染（随想、预览） */
export function render(source: string): string {
  return md.render(source, {});
}
