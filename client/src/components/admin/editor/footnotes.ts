import { Node, mergeAttributes } from '@tiptap/vue-3';
import { Plugin } from '@tiptap/pm/state';
import type { Node as ProseNode } from '@tiptap/pm/model';
import type MarkdownIt from 'markdown-it';
import { configureFootnotes } from '../../../utils/footnotes';
import { i18n } from '../../../i18n';

interface SerializerState {
  write(value: string): void;
  closeBlock(node: ProseNode): void;
  renderContent(node: ProseNode): void;
  wrapBlock(delimiter: string, first: string, node: ProseNode, render: () => void): void;
}
const attributes = (attribute: string) => ({
  label: { default: '', parseHTML: (el: HTMLElement) => el.getAttribute(attribute) ?? '', renderHTML: () => ({}) },
  number: { default: 1, parseHTML: (el: HTMLElement) => Number(el.getAttribute('data-number')) || 1, renderHTML: () => ({}) },
});

export const FootnoteReference = Node.create({
  name: 'footnoteReference', group: 'inline', inline: true, atom: true, priority: 1000,
  addAttributes() { return attributes('data-footnote-reference'); },
  parseHTML() { return [{ tag: 'sup[data-footnote-reference]' }]; },
  renderHTML({ node, HTMLAttributes }) {
    return ['sup', mergeAttributes(HTMLAttributes, { 'data-footnote-reference': node.attrs.label, class: 'editor-footnote-ref' }), `[${node.attrs.number}]`];
  },
  addStorage() {
    return { markdown: {
      serialize(state: SerializerState, node: ProseNode) { state.write(`[^${node.attrs.label}]`); },
      parse: { setup(md: MarkdownIt) { configureFootnotes(md, true); } },
    } };
  },
  addProseMirrorPlugins() {
    return [new Plugin({ appendTransaction(transactions, _old, state) {
      if (!transactions.some(tr => tr.docChanged)) return null;
      const numbers = new Map<string, number>(), nodes: { node: ProseNode; pos: number }[] = [];
      state.doc.descendants((node, pos) => {
        if (node.type.name === 'footnoteReference') {
          if (!numbers.has(node.attrs.label)) numbers.set(node.attrs.label, numbers.size + 1);
          nodes.push({ node, pos });
        } else if (node.type.name === 'footnoteDefinition') nodes.push({ node, pos });
      });
      for (const { node } of nodes) if (!numbers.has(node.attrs.label)) numbers.set(node.attrs.label, numbers.size + 1);
      const tr = state.tr;
      for (const { node, pos } of nodes) {
        const number = numbers.get(node.attrs.label);
        if (node.attrs.number !== number) tr.setNodeMarkup(pos, undefined, { ...node.attrs, number });
      }
      return tr.docChanged ? tr : null;
    } })];
  },
});

export const FootnoteDefinition = Node.create({
  name: 'footnoteDefinition', content: 'block+', defining: true, priority: 1000,
  addAttributes() { return attributes('data-footnote-definition'); },
  parseHTML() { return [{ tag: 'li[data-footnote-definition]', contentElement: '.editor-footnote-body' }]; },
  renderHTML({ node, HTMLAttributes }) {
    return ['li', mergeAttributes(HTMLAttributes, { 'data-footnote-definition': node.attrs.label, class: 'editor-footnote-item', value: node.attrs.number, 'data-number': node.attrs.number }),
      ['span', { class: 'editor-footnote-label', contenteditable: 'false' }, `[^${node.attrs.label}]`],
      ['div', { class: 'editor-footnote-body' }, 0]];
  },
  addStorage() { return { markdown: {
    serialize(state: SerializerState, node: ProseNode) {
      state.wrapBlock('    ', `[^${node.attrs.label}]: `, node, () => state.renderContent(node));
    },
  } }; },
});

export const FootnoteList = Node.create({
  name: 'footnoteList', group: 'block', content: 'footnoteDefinition+', defining: true, priority: 1000,
  parseHTML() { return [{ tag: 'section[data-footnote-list]', contentElement: 'ol' }]; },
  renderHTML({ HTMLAttributes }) {
    return ['section', mergeAttributes(HTMLAttributes, { 'data-footnote-list': '', class: 'editor-footnotes' }),
      ['div', { class: 'footnotes-title', contenteditable: 'false' }, i18n.global.t('footnotes.title')], ['ol', {}, 0]];
  },
  addStorage() { return { markdown: {
    serialize(state: SerializerState, node: ProseNode) { state.renderContent(node); state.closeBlock(node); },
  } }; },
});
