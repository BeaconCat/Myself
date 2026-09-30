<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, shallowRef, watch } from 'vue';
import { Editor, EditorContent } from '@tiptap/vue-3';
import StarterKit from '@tiptap/starter-kit';
import { Table, TableRow, TableHeader, TableCell } from '@tiptap/extension-table';
import TaskList from '@tiptap/extension-task-list';
import TaskItem from '@tiptap/extension-task-item';
import Placeholder from '@tiptap/extension-placeholder';
import { Markdown } from 'tiptap-markdown';
import { ExternalLink, Grid2x2X, PanelTop, Pencil, Unlink } from 'lucide';
import type { MediaItem, MediaKindName } from '../../api';
import { useI18n } from 'vue-i18n';
import '../../views/admin/studio/i18n';
import SIcon from '../../views/admin/studio/SIcon.vue';
import StModal from '../../views/admin/studio/StModal.vue';
import MediaLibraryModal from '../../views/admin/studio/MediaLibraryModal.vue';
import Icon from '../ui/Icon.vue';
import GalleryEditor from './editor/GalleryEditor.vue';
import { Gallery, MediaEmbed, ResizableImage } from './editor/embeds';
import { defaultGallery, type GalleryData, type MediaData, type MediaKind } from '../../utils/embeds';
import { mediaKind } from '../../utils/mediaKind';

/**
 * 所见即所得编辑器：对外始终以 Markdown 交换（统一内容规范），内部用 Tiptap 富文本编辑。
 * - 默认：自带工具栏；lite：精简工具栏（随想）；bare：不渲染工具栏，由外部通过 expose 的命令驱动
 *   （写文章页的浮动胶囊工具条）。
 * - 插图 / 媒体 / 拼图统一走素材库模态框（可就地上传、可填外链）；拼图进拼图编辑器。
 * - 光标在链接上时浮出链接气泡（打开 / 编辑 / 移除），在表格里时浮出表格工具条（增删行列、表头、删表）。
 */
const props = defineProps<{
  modelValue: string;
  placeholder?: string;
  lite?: boolean;
  bare?: boolean;
}>();
const emit = defineEmits<{ 'update:modelValue': [value: string]; typing: [] }>();

const { t } = useI18n();
let applyingExternal = false;
/** 每次事务 +1：外部工具条据此刷新激活态 */
const version = ref(0);

function currentMarkdown(): string {
  return (editor.storage as unknown as { markdown: { getMarkdown: () => string } }).markdown.getMarkdown();
}

/* ===== 素材库 / 拼图编辑器（节点视图经 hooks 调起） ===== */
const lib = reactive({
  open: false,
  title: '',
  accept: ['image'] as MediaKindName[],
  multiple: false,
  done: (_items: MediaItem[]) => {},
});
function openLibrary(opts: { title: string; accept: MediaKindName[]; multiple: boolean; done: (items: MediaItem[]) => void }): void {
  Object.assign(lib, opts, { open: true });
}

const gal = reactive({ open: false, value: null as GalleryData | null, done: (_d: GalleryData) => {} });

const hooks = {
  replaceMedia: (kind: MediaKind, done: (next: Partial<MediaData>) => void) => {
    const accept: MediaKindName[] = kind === 'archive' || kind === 'file' ? ['archive', 'file'] : [kind];
    openLibrary({
      title: t('studio.embed.replaceTitle'),
      accept,
      multiple: false,
      done: ([m]) => m && done({ src: m.url, title: m.title || '' }),
    });
  },
  editGallery: (data: GalleryData, done: (next: GalleryData) => void) => {
    Object.assign(gal, { open: true, value: data, done });
  },
};

const editor = new Editor({
  extensions: [
    StarterKit.configure({ link: { openOnClick: false, autolink: true, defaultProtocol: 'https' } }),
    Table.configure({ resizable: true, cellMinWidth: 60 }),
    TableRow,
    // 单元格只容纳一段行内内容（与 GFM 表格一致）：不能再嵌表格 / 列表等块，换行用 Shift+Enter
    TableHeader.extend({ content: 'paragraph' }),
    TableCell.extend({ content: 'paragraph' }),
    TaskList,
    TaskItem.configure({ nested: true }),
    ResizableImage.configure({ inline: true, allowBase64: false }),
    MediaEmbed.configure(hooks),
    Gallery.configure(hooks),
    Placeholder.configure({ placeholder: () => props.placeholder ?? '' }),
    Markdown.configure({ html: false, linkify: true, breaks: false }),
  ],
  content: props.modelValue,
  // 编辑区的可访问名称与角色（contenteditable 默认没有名字）
  editorProps: { attributes: { 'aria-label': t('studio.a11y.editor'), role: 'textbox', 'aria-multiline': 'true' } },
  onUpdate: () => {
    if (applyingExternal) return;
    emit('update:modelValue', currentMarkdown());
  },
  onTransaction: () => {
    version.value += 1;
  },
});
const editorRef = shallowRef(editor);

watch(
  () => props.modelValue,
  (value) => {
    if (value === currentMarkdown()) return;
    applyingExternal = true;
    editor.commands.setContent(value);
    applyingExternal = false;
  },
);

onBeforeUnmount(() => editor.destroy());

/* ===== 插入 ===== */
/** 图片：素材库多选，逐张插入（行内图片） */
function pickImage(): void {
  openLibrary({
    title: t('studio.embed.imageTitle'),
    accept: ['image'],
    multiple: true,
    done: (items) => {
      const chain = editor.chain().focus();
      items.forEach((m) => chain.setImage({ src: m.url, alt: m.title?.replace(/\.[a-z0-9]+$/i, '') ?? '' }));
      chain.run();
    },
  });
}

/** 视频 / 音频 / 压缩包 / 文件：素材库多选，按类型插入媒体节点（图片走「插入图片」，这里不列） */
function pickMedia(): void {
  openLibrary({
    title: t('studio.embed.mediaTitle'),
    accept: ['video', 'audio', 'archive', 'file'],
    multiple: true,
    done: (items) => {
      const nodes = items.map((m) => {
        const k = mediaKind(m);
        if (k === 'image') return { type: 'image', attrs: { src: m.url } };
        return { type: 'mediaEmbed', attrs: { kind: k, src: m.url, title: m.title || '' } };
      });
      editor.chain().focus().insertContent(nodes).run();
    },
  });
}

/** 拼图：打开拼图编辑器，保存后插入 */
function pickCollage(): void {
  Object.assign(gal, {
    open: true,
    value: defaultGallery(),
    done: (d: GalleryData) => editor.chain().focus().insertContent({ type: 'gallery', attrs: { data: d } }).run(),
  });
}

async function insertImageUrl(): Promise<void> {
  pickImage();
}

/* ===== 链接：气泡 + 编辑框 ===== */
const root = ref<HTMLElement | null>(null);
const linkDlg = reactive({ open: false, text: '', url: '', editing: false, error: '' });

/** 规范化链接：补 https://；只放行 http(s) / mailto / tel / 站内路径 / 锚点 */
function normalizeUrl(raw: string): string | null {
  const s = raw.trim();
  if (!s) return '';
  if (/^(https?:|mailto:|tel:)/i.test(s) || s.startsWith('/') || s.startsWith('#')) return s;
  if (/^[a-z][a-z0-9+.-]*:/i.test(s)) return null;
  if (/^[\w-]+(\.[\w-]+)+(:\d+)?(\/|$|\?|#)/.test(s)) return `https://${s}`;
  return null;
}

function setLink(): void {
  const { from, to, empty } = editor.state.selection;
  const inLink = editor.isActive('link');
  if (inLink) editor.chain().extendMarkRange('link').run();
  const sel = editor.state.selection;
  linkDlg.text = editor.state.doc.textBetween(sel.from, sel.to, ' ') || (empty ? '' : editor.state.doc.textBetween(from, to, ' '));
  linkDlg.url = (editor.getAttributes('link').href as string | undefined) ?? '';
  linkDlg.editing = inLink;
  linkDlg.error = '';
  linkDlg.open = true;
}

function applyLink(): void {
  const url = normalizeUrl(linkDlg.url);
  if (url === null) {
    linkDlg.error = t('studio.link.bad');
    return;
  }
  const chain = editor.chain().focus();
  if (!url) {
    chain.extendMarkRange('link').unsetLink().run();
    linkDlg.open = false;
    return;
  }
  const text = linkDlg.text.trim() || url;
  const { empty } = editor.state.selection;
  const current = editor.state.doc.textBetween(editor.state.selection.from, editor.state.selection.to, ' ');
  if (empty || text !== current) {
    chain.insertContent({ type: 'text', text, marks: [{ type: 'link', attrs: { href: url } }] }).run();
  } else {
    chain.extendMarkRange('link').setLink({ href: url }).run();
  }
  linkDlg.open = false;
}

function unlink(): void {
  editor.chain().focus().extendMarkRange('link').unsetLink().run();
}

function openHref(): void {
  const href = editor.getAttributes('link').href as string | undefined;
  if (href) window.open(href, '_blank', 'noopener,noreferrer');
}

/** 浮层定位：相对编辑器根节点 */
function relRect(r: DOMRect): { x: number; y: number; w: number; h: number } {
  const box = root.value?.getBoundingClientRect();
  return { x: r.left - (box?.left ?? 0), y: r.top - (box?.top ?? 0), w: r.width, h: r.height };
}

const bubble = computed(() => {
  void version.value;
  if (!editor.isFocused || !editor.isActive('link') || !root.value) return null;
  const href = (editor.getAttributes('link').href as string | undefined) ?? '';
  const pos = editor.view.coordsAtPos(editor.state.selection.from);
  const r = relRect(new DOMRect(pos.left, pos.top, 0, pos.bottom - pos.top));
  return { href, x: r.x, y: r.y + r.h + 8 };
});

/* ===== 表格工具条 ===== */
const tableBar = computed(() => {
  void version.value;
  if (!editor.isActive('table') || !root.value) return null;
  const dom = editor.view.domAtPos(editor.state.selection.from).node as HTMLElement;
  const el = (dom.nodeType === 1 ? dom : dom.parentElement)?.closest('table');
  if (!el) return null;
  const r = relRect(el.getBoundingClientRect());
  return { x: r.x, y: r.y - 44 };
});

/**
 * 表格工具条：按「行 / 列」分组，每个按钮直接写明动作（纯方向图标难以理解）；
 * 表头是开关，删除表格单列并标危险色。
 */
const TABLE_GROUPS = computed(() => [
  {
    label: t('studio.table.row'),
    items: [
      { text: t('studio.table.above'), title: t('studio.table.rowBefore'), run: () => editor.chain().focus().addRowBefore().run() },
      { text: t('studio.table.below'), title: t('studio.table.rowAfter'), run: () => editor.chain().focus().addRowAfter().run() },
      { text: t('studio.table.del'), title: t('studio.table.rowDelete'), run: () => editor.chain().focus().deleteRow().run() },
    ],
  },
  {
    label: t('studio.table.col'),
    items: [
      { text: t('studio.table.left'), title: t('studio.table.colBefore'), run: () => editor.chain().focus().addColumnBefore().run() },
      { text: t('studio.table.right'), title: t('studio.table.colAfter'), run: () => editor.chain().focus().addColumnAfter().run() },
      { text: t('studio.table.del'), title: t('studio.table.colDelete'), run: () => editor.chain().focus().deleteColumn().run() },
    ],
  },
]);
const headerOn = computed(() => {
  void version.value;
  return editor.isActive('tableHeader');
});

/** 插入表格：光标已在表格里时不插（不允许表格套表格） */
function insertTable(): void {
  if (editor.isActive('table')) return;
  editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run();
}

/* ===== 自带工具栏（非 bare） ===== */
type Cmd = { icon: string; title: string; run: () => unknown; active?: () => boolean; disabled?: () => boolean } | { divider: true };

const c = () => editor.chain().focus();
const TOOLBAR: Cmd[] = [
  { icon: 'bold', title: t('studio.editor.bold'), run: () => c().toggleBold().run(), active: () => editor.isActive('bold') },
  { icon: 'italic', title: t('studio.editor.italic'), run: () => c().toggleItalic().run(), active: () => editor.isActive('italic') },
  { icon: 'strike', title: t('studio.editor.strike'), run: () => c().toggleStrike().run(), active: () => editor.isActive('strike') },
  { icon: 'code', title: t('studio.editor.code'), run: () => c().toggleCode().run(), active: () => editor.isActive('code') },
  { divider: true },
  { icon: 'ul', title: t('studio.editor.ul'), run: () => c().toggleBulletList().run(), active: () => editor.isActive('bulletList') },
  { icon: 'ol', title: t('studio.editor.ol'), run: () => c().toggleOrderedList().run(), active: () => editor.isActive('orderedList') },
  { icon: 'task', title: t('studio.editor.task'), run: () => c().toggleTaskList().run(), active: () => editor.isActive('taskList') },
  { icon: 'quote', title: t('studio.editor.quote'), run: () => c().toggleBlockquote().run(), active: () => editor.isActive('blockquote') },
  { divider: true },
  { icon: 'link', title: t('studio.editor.link'), run: setLink, active: () => editor.isActive('link') },
  { icon: 'undo', title: t('studio.editor.undo'), run: () => c().undo().run() },
  { icon: 'redo', title: t('studio.editor.redo'), run: () => c().redo().run() },
];
const FULL_EXTRA: Cmd[] = [
  { divider: true },
  { icon: 'codeBlock', title: t('studio.editor.codeBlock'), run: () => c().toggleCodeBlock().run(), active: () => editor.isActive('codeBlock') },
  { icon: 'table', title: t('studio.editor.table'), run: insertTable, disabled: () => editor.isActive('table') },
  { icon: 'hr', title: t('studio.editor.hr'), run: () => c().setHorizontalRule().run() },
  { icon: 'image', title: t('studio.editor.image'), run: pickImage },
  { icon: 'collage', title: t('studio.editor.collage'), run: pickCollage },
  { icon: 'media', title: t('studio.editor.media'), run: pickMedia },
];
const toolbar = props.lite ? TOOLBAR : [...TOOLBAR, ...FULL_EXTRA];

function onLibPick(items: MediaItem[]): void {
  lib.done(items);
}

defineExpose({
  editor: editorRef,
  version,
  setLink,
  insertImageUrl,
  pickImage,
  pickCollage,
  pickMedia,
  insertTable,
  focus: () => editor.commands.focus(),
});
</script>

<template>
  <div ref="root" class="rich" :class="{ lite, bare }">
    <div v-if="!bare" class="toolbar" :data-v="version">
      <template v-for="(item, i) in toolbar" :key="i">
        <span v-if="'divider' in item" class="divider" />
        <button
          v-else
          type="button"
          class="tool"
          :class="{ on: item.active?.() }"
          :title="item.disabled?.() ? t('studio.write.tool.tableNested') : item.title"
          :disabled="item.disabled?.()"
          @click="item.run()"
        ><SIcon :name="item.icon" :size="16" /></button>
      </template>
    </div>
    <EditorContent class="content" :editor="editor" @keydown="emit('typing')" />

    <!-- 链接气泡 -->
    <Transition name="rb-pop">
      <div v-if="bubble" class="rb link-bubble" :style="{ left: `${bubble.x}px`, top: `${bubble.y}px` }" @mousedown.prevent>
        <span class="href" :title="bubble.href">{{ bubble.href }}</span>
        <button type="button" :title="t('studio.link.open')" @click="openHref"><Icon :icon="ExternalLink" :size="15" /></button>
        <button type="button" :title="t('studio.link.edit')" @click="setLink"><Icon :icon="Pencil" :size="15" /></button>
        <button type="button" :title="t('studio.link.remove')" @click="unlink"><Icon :icon="Unlink" :size="15" /></button>
      </div>
    </Transition>

    <!-- 表格工具条 -->
    <Transition name="rb-pop">
      <div v-if="tableBar" class="rb table-bar" role="toolbar" :aria-label="t('studio.a11y.tableBar')" :style="{ left: `${tableBar.x}px`, top: `${tableBar.y}px` }" @mousedown.prevent>
        <template v-for="g in TABLE_GROUPS" :key="g.label">
          <span class="grp-l">{{ g.label }}</span>
          <button v-for="it in g.items" :key="it.title" type="button" class="txt" :title="it.title" :aria-label="it.title" @click="it.run()">{{ it.text }}</button>
          <span class="sep" />
        </template>
        <button type="button" class="txt" :class="{ on: headerOn }" :title="t('studio.table.header')" :aria-pressed="headerOn" @click="editor.chain().focus().toggleHeaderRow().run()">
          <Icon :icon="PanelTop" :size="14" />{{ t('studio.table.headerShort') }}
        </button>
        <button type="button" class="txt danger" :title="t('studio.table.remove')" @click="editor.chain().focus().deleteTable().run()">
          <Icon :icon="Grid2x2X" :size="14" />{{ t('studio.table.removeShort') }}
        </button>
      </div>
    </Transition>

    <MediaLibraryModal
      :open="lib.open"
      :title="lib.title"
      :accept="lib.accept"
      :multiple="lib.multiple"
      @pick="onLibPick"
      @close="lib.open = false"
    />
    <GalleryEditor :open="gal.open" :value="gal.value" @save="(d) => gal.done(d)" @close="gal.open = false" />

    <StModal :open="linkDlg.open" panel-class="link-dlg" @close="linkDlg.open = false">
      <h3>{{ linkDlg.editing ? t('studio.link.editTitle') : t('studio.link.addTitle') }}</h3>
      <div class="st-flabel">{{ t('studio.link.text') }}</div>
      <label class="st-field"><input v-model="linkDlg.text" :placeholder="t('studio.link.textPh')" @keydown.enter.prevent="applyLink" /></label>
      <div class="st-flabel">{{ t('studio.link.url') }}</div>
      <label class="st-field" :class="{ bad: linkDlg.error }">
        <input v-model="linkDlg.url" placeholder="https://" autofocus @keydown.enter.prevent="applyLink" @input="linkDlg.error = ''" />
      </label>
      <p class="hint" :class="{ err: linkDlg.error }">{{ linkDlg.error || t('studio.link.hint') }}</p>
      <div class="acts">
        <button v-if="linkDlg.editing" type="button" class="st-btn g" @click="linkDlg.url = ''; applyLink()">{{ t('studio.link.remove') }}</button>
        <span class="sp" />
        <button type="button" class="st-btn g" @click="linkDlg.open = false">{{ t('studio.cancel') }}</button>
        <button type="button" class="st-btn p" @click="applyLink">{{ t('studio.editor.ok') }}</button>
      </div>
    </StModal>
  </div>
</template>

<style scoped lang="scss">
.rich {
  position: relative;
  border-radius: var(--r-sm);
  background: var(--surface);
  box-shadow: 0 0 0 1px var(--border);

  &:focus-within { box-shadow: 0 0 0 1px color-mix(in oklab, var(--ink) 70%, transparent), 0 0 0 3px color-mix(in oklab, var(--ink) 18%, transparent); }

  &.bare {
    border-radius: 0;
    background: none;
    box-shadow: none;
  }
}

.toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 2px;
  padding: 6px 8px;
  border-bottom: 1px solid var(--border);
  border-radius: var(--r-sm) var(--r-sm) 0 0;
  background: var(--surface-2);
  position: sticky;
  top: 0;
  z-index: 5;
}

.tool {
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  border: none;
  border-radius: var(--r-xs);
  background: none;
  color: var(--text-2);
  cursor: pointer;
  transition: all var(--dur-fast);

  &:hover { background: var(--surface); color: var(--text); }
  &.on { background: var(--lift); box-shadow: var(--lift-shadow); color: var(--lift-fg); }
  &:disabled { opacity: 0.35; cursor: not-allowed; background: none; }
}

.divider {
  width: 1px;
  height: 18px;
  background: var(--border);
  margin: 0 4px;
}

.rich.lite {
  box-shadow: none;
  background: none;

  .toolbar { border-radius: var(--r-sm); border: 1px solid var(--border); margin-bottom: 4px; }
  .content :deep(.ProseMirror) { min-height: 140px; padding: 12px 4px; font-size: 16px; }
}

/* 默认 / 精简模式的正文排版 */
.rich:not(.bare) .content :deep(.ProseMirror) {
  min-height: 52vh;
  padding: 20px 24px;
  font-size: 15px;
  line-height: 1.9;
}

.content :deep(.ProseMirror) {
  outline: none;

  > * + * { margin-top: 0.6em; }

  h1, h2, h3 { font-family: var(--font-serif); line-height: 1.4; }
  ul, ol { padding-left: 26px; }

  /* 待办：复选框与首行文字垂直居中（高度取一行的行高） */
  ul[data-type='taskList'] {
    list-style: none;
    padding-left: 2px;

    li {
      display: flex;
      gap: 10px;
      align-items: flex-start;

      > label {
        flex: none;
        height: 1lh;
        display: flex;
        align-items: center;
        margin: 0;
        user-select: none;
      }

      > div { flex: 1; min-width: 0; }
      > div > p { margin: 0; }

      &[data-checked='true'] > div { color: var(--text-2); text-decoration: line-through; }
    }

    input[type='checkbox'] {
      appearance: none;
      width: 17px;
      height: 17px;
      margin: 0;
      display: grid;
      place-items: center;
      border-radius: var(--r-xs);
      cursor: pointer;
      background: var(--paper, var(--surface));
      box-shadow: 0 0 0 1.5px var(--line-3, var(--border)) inset;
      transition: background var(--dur-fast), box-shadow var(--dur-fast);

      /* 勾：两条边框旋转而成（CSS 形状，非图标） */
      &::after {
        content: '';
        width: 4px;
        height: 8px;
        margin-top: -2px;
        border: solid var(--on-solid, #fff);
        border-width: 0 2px 2px 0;
        transform: rotate(45deg) scale(0);
        transition: transform var(--dur-fast) var(--ease-spring);
      }

      &:checked { background: var(--solid); box-shadow: 0 0 0 1.5px var(--solid) inset; }
      &:checked::after { transform: rotate(45deg) scale(1); }
      &:focus-visible { outline: 2px solid var(--ink); outline-offset: 2px; }
    }
  }

  code {
    font-family: ui-monospace, Consolas, monospace;
    font-size: 0.86em;
    background: var(--surface-2);
    padding: 2px 6px;
    border-radius: var(--r-xs);
  }

  pre {
    background: var(--code-bg, #1b1a1f);
    color: #e6e3dc;
    border-radius: var(--r-sm);
    padding: 16px 18px;
    overflow-x: auto;
    font-size: 14px;
    line-height: 1.7;

    code { background: none; padding: 0; color: inherit; }
  }

  /* 表格：固定布局等分列宽（拖动列宽后按拖动结果），拖柄只在悬停列边时出现 */
  .tableWrapper { overflow-x: auto; margin: 1.2em 0; }

  table {
    border-collapse: collapse;
    width: 100%;
    table-layout: fixed;
    margin: 0;

    th, td {
      border: 1px solid var(--line-2, var(--border));
      padding: 8px 12px;
      vertical-align: top;
      position: relative;
      min-width: 60px;

      > p { margin: 0; }
    }

    th { background: var(--well, var(--surface-2)); font-weight: 600; text-align: left; }

    .selectedCell::after {
      content: '';
      position: absolute;
      inset: 0;
      background: var(--tint);
      pointer-events: none;
    }

    .column-resize-handle {
      position: absolute;
      right: -1px;
      top: 0;
      bottom: 0;
      width: 2px;
      background: color-mix(in oklab, var(--ink) 70%, transparent);
      pointer-events: none;
    }
  }

  &.resize-cursor { cursor: col-resize; }

  img {
    max-width: 100%;
    border-radius: var(--r-sm);

    &.ProseMirror-selectednode { outline: 2px solid var(--ink); outline-offset: 2px; }
  }

  hr { border: none; border-top: 1px solid var(--border); margin: 2em 0; }

  a { color: var(--ink); text-decoration: underline; text-underline-offset: 3px; }

  p.is-editor-empty:first-child::before {
    content: attr(data-placeholder);
    color: var(--st-ink-4, var(--text-2));
    float: left;
    height: 0;
    pointer-events: none;
  }
}

/* 默认模式下的引用与标题 */
.rich:not(.bare) .content :deep(.ProseMirror) {
  h1 { font-size: 28px; }
  h2 { font-size: 22px; }
  h3 { font-size: 18px; }

  blockquote {
    border-left: 3px solid var(--line-2);
    padding: 6px 14px;
    color: var(--text-2);
    background: var(--surface-2);
    border-radius: 0 var(--r-xs) var(--r-xs) 0;
  }
}

/* ---------- 浮层：链接气泡 / 表格工具条 ---------- */
.rb {
  position: absolute;
  z-index: 20;
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 4px;
  border-radius: var(--r-sm);
  background: var(--paper, var(--surface));
  box-shadow: 0 0 0 1px var(--line-2, var(--border)), var(--shadow-pop, 0 8px 24px rgba(0, 0, 0, 0.12));

  button {
    height: 30px;
    min-width: 30px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 3px;
    padding: 0 6px;
    border: 0;
    border-radius: var(--r-xs);
    cursor: pointer;
    color: var(--st-ink-2, var(--text-2));
    background: none;

    small { font-size: 11.5px; }

    /* 表格工具条：文字按钮 */
    &.txt { padding: 0 8px; font: 500 12.5px var(--font-sans); white-space: nowrap; }
    &.on { color: var(--ink, var(--primary)); background: var(--tint, var(--surface-2)); }
    &:hover { color: var(--st-ink, var(--text)); background: var(--hover, var(--surface-2)); }
    &.danger:hover { color: #fff; background: color-mix(in oklab, var(--red, #ff0032) 85%, black); }
  }

  .sep { width: 1px; height: 18px; margin: 0 3px; background: var(--line-2, var(--border)); }

  .grp-l {
    padding: 0 4px 0 6px;
    font-size: 11.5px;
    font-weight: 600;
    color: var(--st-ink-4, var(--text-2));
    letter-spacing: 0.08em;
  }
}

.link-bubble .href {
  max-width: 280px;
  padding: 0 8px;
  font: 12.5px var(--font-mono);
  color: var(--st-ink-2, var(--text-2));
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rb-pop-enter-active, .rb-pop-leave-active { transition: opacity var(--dur-fast), transform var(--dur-fast) var(--ease-out); }
.rb-pop-enter-from, .rb-pop-leave-to { opacity: 0; transform: translateY(4px); }

:global(.st-modal.link-dlg) { width: min(480px, calc(100vw - 32px)); }
:global(.st-modal.link-dlg .st-flabel) { margin: 12px 0 6px; }
:global(.st-modal.link-dlg .st-field.bad) { box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--red) 70%, transparent) inset; }
:global(.st-modal.link-dlg .hint) { margin: 8px 0 0; font-size: 12.5px; color: var(--st-ink-3); }
:global(.st-modal.link-dlg .hint.err) { color: var(--red); }
:global(.st-modal.link-dlg .acts) { display: flex; gap: 8px; margin-top: 18px; }
:global(.st-modal.link-dlg .acts .sp) { flex: 1; }
</style>
