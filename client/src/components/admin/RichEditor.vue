<script setup lang="ts">
import { onBeforeUnmount, ref, shallowRef, watch } from 'vue';
import { Editor, EditorContent } from '@tiptap/vue-3';
import StarterKit from '@tiptap/starter-kit';
import { Table, TableRow, TableHeader, TableCell } from '@tiptap/extension-table';
import TaskList from '@tiptap/extension-task-list';
import TaskItem from '@tiptap/extension-task-item';
import Image from '@tiptap/extension-image';
import Placeholder from '@tiptap/extension-placeholder';
import { Markdown } from 'tiptap-markdown';
import { adminApi } from '../../api';
import { useDialogStore } from '../../stores/dialog';
import { useI18n } from 'vue-i18n';
import '../../views/admin/studio/i18n';
import SIcon from '../../views/admin/studio/SIcon.vue';

/**
 * 所见即所得编辑器：对外始终以 Markdown 交换（统一内容规范），内部用 Tiptap 富文本编辑。
 * - 默认：自带工具栏；lite：精简工具栏（随想）；bare：不渲染工具栏，由外部通过 expose 的命令驱动
 *   （写文章页的浮动胶囊工具条）。
 */
const props = defineProps<{
  modelValue: string;
  placeholder?: string;
  lite?: boolean;
  bare?: boolean;
}>();
const emit = defineEmits<{ 'update:modelValue': [value: string]; typing: [] }>();

const dialog = useDialogStore();
const { t } = useI18n();
let applyingExternal = false;
/** 每次事务 +1：外部工具条据此刷新激活态 */
const version = ref(0);

function currentMarkdown(): string {
  return (editor.storage as unknown as { markdown: { getMarkdown: () => string } }).markdown.getMarkdown();
}

const editor = new Editor({
  extensions: [
    StarterKit.configure({ link: { openOnClick: false } }),
    Table.configure({ resizable: true }),
    TableRow,
    TableHeader,
    TableCell,
    TaskList,
    TaskItem.configure({ nested: true }),
    Image.configure({ inline: true, allowBase64: false }),
    Placeholder.configure({ placeholder: () => props.placeholder ?? '' }),
    Markdown.configure({ html: false, linkify: true, breaks: false }),
  ],
  content: props.modelValue,
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

/* ===== 命令 ===== */
async function setLink(): Promise<void> {
  const prev = editor.getAttributes('link').href as string | undefined;
  const url = await dialog.prompt({ title: t('studio.editor.linkTitle'), inputValue: prev ?? 'https://', confirmText: t('studio.editor.ok') });
  if (url === null) return;
  if (!url) {
    editor.chain().focus().unsetLink().run();
    return;
  }
  editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run();
}

async function insertImageUrl(): Promise<void> {
  const url = await dialog.prompt({ title: t('studio.editor.imageTitle'), message: t('studio.editor.imageMsg'), placeholder: '/uploads/…' });
  if (url) editor.chain().focus().setImage({ src: url }).run();
}

/** 上传并插入；collage=true 时多图并排插入同一段落（前台渲染成宫格） */
async function uploadInsert(files: File[], collage = false): Promise<void> {
  const list = files.filter((f) => f.type.startsWith('image/'));
  if (!list.length) return;
  const uploaded = await adminApi.uploadMedia(list);
  const chain = editor.chain().focus();
  uploaded.forEach((item, i) => {
    if (collage && i > 0) chain.insertContent(' ');
    chain.setImage({ src: item.url });
  });
  chain.run();
}

const fileInput = ref<HTMLInputElement | null>(null);
const collageInput = ref<HTMLInputElement | null>(null);

function onPick(e: Event, collage: boolean): void {
  const input = e.target as HTMLInputElement;
  if (input.files?.length) void uploadInsert([...input.files], collage);
  input.value = '';
}

type Cmd = { icon: string; title: string; run: () => unknown; active?: () => boolean; table?: boolean } | { divider: true };

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
  { icon: 'table', title: t('studio.editor.table'), run: () => c().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run(), active: () => editor.isActive('table') },
  { icon: 'hr', title: t('studio.editor.hr'), run: () => c().setHorizontalRule().run() },
  { icon: 'image', title: t('studio.editor.image'), run: () => fileInput.value?.click() },
  { icon: 'collage', title: t('studio.editor.collage'), run: () => collageInput.value?.click() },
];
const toolbar = props.lite ? TOOLBAR : [...TOOLBAR, ...FULL_EXTRA];

defineExpose({
  editor: editorRef,
  version,
  setLink,
  insertImageUrl,
  pickImage: () => fileInput.value?.click(),
  pickCollage: () => collageInput.value?.click(),
  focus: () => editor.commands.focus(),
});
</script>

<template>
  <div class="rich" :class="{ lite, bare }">
    <div v-if="!bare" class="toolbar" :data-v="version">
      <template v-for="(item, i) in toolbar" :key="i">
        <span v-if="'divider' in item" class="divider" />
        <button
          v-else
          type="button"
          class="tool"
          :class="{ on: item.active?.() }"
          :title="item.title"
          @click="item.run()"
        ><SIcon :name="item.icon" :size="16" /></button>
      </template>
    </div>
    <input ref="fileInput" type="file" accept="image/*" multiple hidden @change="onPick($event, false)" />
    <input ref="collageInput" type="file" accept="image/*" multiple hidden @change="onPick($event, true)" />
    <EditorContent class="content" :editor="editor" @keydown="emit('typing')" />
  </div>
</template>

<style scoped lang="scss">
.rich {
  border-radius: var(--r-sm);
  background: var(--surface);
  box-shadow: 0 0 0 1px var(--border);
  overflow: hidden;

  &:focus-within { box-shadow: 0 0 0 1px color-mix(in oklab, var(--ink) 70%, transparent), 0 0 0 3px color-mix(in oklab, var(--ink) 18%, transparent); }

  &.bare {
    border-radius: 0;
    background: none;
    box-shadow: none;
    overflow: visible;
  }
}

.toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 2px;
  padding: 6px 8px;
  border-bottom: 1px solid var(--border);
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

  ul[data-type='taskList'] {
    list-style: none;
    padding-left: 4px;

    li {
      display: flex;
      gap: 8px;
      align-items: flex-start;

      > label { margin-top: 0.35em; }
      input[type='checkbox'] { accent-color: var(--solid); }
      > div { flex: 1; }

      &[data-checked='true'] > div { color: var(--text-2); text-decoration: line-through; }
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

  table {
    border-collapse: collapse;
    width: 100%;
    table-layout: fixed;

    th, td { border: 1px solid var(--border); padding: 7px 12px; vertical-align: top; position: relative; }
    th { background: var(--surface-2); font-weight: 600; }

    .selectedCell::after {
      content: '';
      position: absolute;
      inset: 0;
      background: var(--tint);
      pointer-events: none;
    }

    .column-resize-handle {
      position: absolute;
      right: -2px;
      top: 0;
      bottom: 0;
      width: 4px;
      background: var(--ink);
      cursor: col-resize;
    }
  }

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
</style>
