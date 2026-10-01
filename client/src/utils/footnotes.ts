import type MarkdownIt from 'markdown-it';
import footnote from 'markdown-it-footnote';
import { i18n } from '../i18n';

const configured = new WeakSet<MarkdownIt>();
let serial = 0;
interface FootnoteEnv { footnoteScope?: string; footnotes?: { refs?: Record<string, number>; list?: { label?: string; count: number }[] } }
interface FootnoteMeta { id: number; subId?: number; label?: string }

/** Shared syntax, with reader links or editor-safe node attributes. */
export function configureFootnotes(md: MarkdownIt, editor = false): void {
  if (configured.has(md)) return;
  configured.add(md);
  md.use(footnote);
  const escape = md.utils.escapeHtml;
  md.core.ruler.before('block', 'myself_footnote_scope', state => {
    const env = state.env as FootnoteEnv;
    env.footnoteScope = env.footnoteScope?.replace(/[^a-z\d_-]/gi, '-') || `block-${++serial}`;
  });
  if (editor) {
    // Preserve definitions while the author is still writing their references.
    md.core.ruler.before('footnote_tail', 'myself_unused_footnotes', state => {
      const env = state.env as FootnoteEnv;
      if (!env.footnotes) return;
      env.footnotes.list ??= [];
      env.footnotes.refs ??= {};
      const used = new Set(Object.keys(env.footnotes.refs).map(key => key.slice(1)));
      for (const [index, entry] of env.footnotes.list.entries()) {
        if (entry.label) continue;
        let label = `inline-${index + 1}`;
        while (used.has(label)) label += '-inline';
        entry.label = label; used.add(label);
      }
      for (const [key, id] of Object.entries(env.footnotes.refs)) {
        if (id >= 0) continue;
        env.footnotes.refs[key] = env.footnotes.list.length;
        env.footnotes.list.push({ label: key.slice(1), count: 0 });
      }
    });
  }
  const key = (meta: FootnoteMeta, env: FootnoteEnv) => meta.label || env.footnotes?.list?.[meta.id]?.label || `inline-${meta.id + 1}`;
  const noteId = (meta: FootnoteMeta, env: FootnoteEnv) => `myself-fn-${env.footnoteScope}-${meta.id + 1}`;
  const refId = (meta: FootnoteMeta, env: FootnoteEnv) => `${noteId(meta, env)}-ref-${(meta.subId ?? 0) + 1}`;
  md.renderer.rules.footnote_ref = (tokens, index, _options, env: FootnoteEnv) => {
    const meta = tokens[index].meta as FootnoteMeta, number = meta.id + 1;
    if (editor) return `<sup data-footnote-reference="${escape(key(meta, env))}" data-number="${number}">[${number}]</sup>`;
    const label = escape(i18n.global.t('footnotes.reference', { n: number }));
    return `<sup class="footnote-ref"><a data-footnote-link role="doc-noteref" href="#${noteId(meta, env)}" id="${refId(meta, env)}" aria-label="${label}">[${number}]</a></sup>`;
  };
  md.renderer.rules.footnote_block_open = () => editor
    ? '<section data-footnote-list><ol>'
    : `<section class="footnotes" role="doc-endnotes" aria-label="${escape(i18n.global.t('footnotes.title'))}"><div class="footnotes-title" role="heading" aria-level="2">${escape(i18n.global.t('footnotes.title'))}</div><ol class="footnotes-list">`;
  md.renderer.rules.footnote_block_close = () => '</ol></section>';
  md.renderer.rules.footnote_open = (tokens, index, _options, env: FootnoteEnv) => {
    const meta = tokens[index].meta as FootnoteMeta;
    return editor
      ? `<li data-footnote-definition="${escape(key(meta, env))}" data-number="${meta.id + 1}"><div class="editor-footnote-body">`
      : `<li class="footnote-item" role="doc-endnote" id="${noteId(meta, env)}" tabindex="-1"><div class="footnote-content">`;
  };
  md.renderer.rules.footnote_close = () => '</div></li>';
  md.renderer.rules.footnote_anchor = (tokens, index, _options, env: FootnoteEnv) => {
    if (editor) return '';
    const meta = tokens[index].meta as FootnoteMeta;
    const label = escape(i18n.global.t('footnotes.back', { n: meta.id + 1, r: (meta.subId ?? 0) + 1 }));
    return ` <a data-footnote-link class="footnote-backref" role="doc-backlink" href="#${refId(meta, env)}" aria-label="${label}">↩︎</a>`;
  };
}
