<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue';
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

const dialog = useDialogStore();

/**
 * 所见即所得编辑器：对外始终以 Markdown 交换（统一内容规范），
 * 内部用 Tiptap 富文本编辑。支持标题/列表/待办/表格/图片/拼图/代码/引用。
 */
const props = defineProps<{ modelValue: string; placeholder?: string }>();
const emit = defineEmits<{ 'update:modelValue': [value: string] }>();

let applyingExternal = false;

/** tiptap-markdown 未提供 Storage 类型增强 */
function currentMarkdown(): string {
  return (editor.storage as unknown as { markdown: { getMarkdown: () => string } })
    .markdown.getMarkdown();
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
});

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

/* ===== 工具栏动作 ===== */
const fileInput = ref<HTMLInputElement | null>(null);
const collageInput = ref<HTMLInputElement | null>(null);

async function setLink(): Promise<void> {
  const prev = editor.getAttributes('link').href as string | undefined;
  const url = await dialog.prompt({ title: '链接地址', inputValue: prev ?? 'https://' });
  if (url === null) return;
  if (!url) {
    editor.chain().focus().unsetLink().run();
    return;
  }
  editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run();
}

async function insertImageUrl(): Promise<void> {
  const url = await dialog.prompt({
    title: '插入图片',
    message: '可先在素材库上传后复制链接。',
    placeholder: '/uploads/…',
  });
  if (url) editor.chain().focus().setImage({ src: url }).run();
}

/** 上传并插入单图 */
async function uploadInsert(e: Event): Promise<void> {
  const input = e.target as HTMLInputElement;
  if (!input.files?.length) return;
  const uploaded = await adminApi.uploadMedia([...input.files]);
  input.value = '';
  const chain = editor.chain().focus();
  for (const item of uploaded) chain.setImage({ src: item.url });
  chain.run();
}

/** 拼图：多图并排插入同一段落（前台按行内连续图片渲染成宫格） */
async function uploadCollage(e: Event): Promise<void> {
  const input = e.target as HTMLInputElement;
  if (!input.files?.length) return;
  const uploaded = await adminApi.uploadMedia([...input.files]);
  input.value = '';
  const chain = editor.chain().focus();
  uploaded.forEach((item, i) => {
    if (i > 0) chain.insertContent(' ');
    chain.setImage({ src: item.url });
  });
  chain.run();
}

const TOOLBAR = [
  { icon: 'B', title: '加粗', run: () => editor.chain().focus().toggleBold().run(), active: () => editor.isActive('bold') },
  { icon: 'I', title: '斜体', run: () => editor.chain().focus().toggleItalic().run(), active: () => editor.isActive('italic') },
  { icon: 'S', title: '删除线', run: () => editor.chain().focus().toggleStrike().run(), active: () => editor.isActive('strike') },
  { icon: '<>', title: '行内代码', run: () => editor.chain().focus().toggleCode().run(), active: () => editor.isActive('code') },
  { divider: true },
  { icon: 'H1', title: '一级标题', run: () => editor.chain().focus().toggleHeading({ level: 1 }).run(), active: () => editor.isActive('heading', { level: 1 }) },
  { icon: 'H2', title: '二级标题', run: () => editor.chain().focus().toggleHeading({ level: 2 }).run(), active: () => editor.isActive('heading', { level: 2 }) },
  { icon: 'H3', title: '三级标题', run: () => editor.chain().focus().toggleHeading({ level: 3 }).run(), active: () => editor.isActive('heading', { level: 3 }) },
  { divider: true },
  { icon: '•', title: '无序列表', run: () => editor.chain().focus().toggleBulletList().run(), active: () => editor.isActive('bulletList') },
  { icon: '1.', title: '有序列表', run: () => editor.chain().focus().toggleOrderedList().run(), active: () => editor.isActive('orderedList') },
  { icon: '☐', title: '待办清单', run: () => editor.chain().focus().toggleTaskList().run(), active: () => editor.isActive('taskList') },
  { icon: '❝', title: '引用', run: () => editor.chain().focus().toggleBlockquote().run(), active: () => editor.isActive('blockquote') },
  { icon: '{ }', title: '代码块', run: () => editor.chain().focus().toggleCodeBlock().run(), active: () => editor.isActive('codeBlock') },
  { icon: '—', title: '分隔线', run: () => editor.chain().focus().setHorizontalRule().run(), active: () => false },
  { divider: true },
  { icon: '⌗', title: '插入表格 3×3', run: () => editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run(), active: () => editor.isActive('table') },
  { icon: '+行', title: '下方加行', run: () => editor.chain().focus().addRowAfter().run(), active: () => false, needTable: true },
  { icon: '+列', title: '右侧加列', run: () => editor.chain().focus().addColumnAfter().run(), active: () => false, needTable: true },
  { icon: '-行', title: '删除行', run: () => editor.chain().focus().deleteRow().run(), active: () => false, needTable: true },
  { icon: '-列', title: '删除列', run: () => editor.chain().focus().deleteColumn().run(), active: () => false, needTable: true },
  { icon: '×表', title: '删除表格', run: () => editor.chain().focus().deleteTable().run(), active: () => false, needTable: true },
  { divider: true },
  { icon: '🔗', title: '链接', run: setLink, active: () => editor.isActive('link') },
  { icon: '⎌', title: '撤销', run: () => editor.chain().focus().undo().run(), active: () => false },
  { icon: '⎌⃗', title: '重做', run: () => editor.chain().focus().redo().run(), active: () => false },
] as const;
</script>

<template>
  <div class="rich">
    <!-- 工具栏 -->
    <div class="toolbar">
      <template v-for="(item, i) in TOOLBAR" :key="i">
        <span v-if="'divider' in item && item.divider" class="divider" />
        <button
          v-else
          type="button"
          class="tool"
          :class="{ on: 'active' in item && item.active(), dim: 'needTable' in item && item.needTable && !editor.isActive('table') }"
          :title="'title' in item ? item.title : ''"
          @click="'run' in item && item.run()"
        >{{ 'icon' in item ? item.icon : '' }}</button>
      </template>
      <span class="divider" />
      <button type="button" class="tool" title="插入图片链接" @click="insertImageUrl">图链</button>
      <button type="button" class="tool" title="上传并插入图片" @click="fileInput?.click()">传图</button>
      <button type="button" class="tool" title="上传多图插入拼图" @click="collageInput?.click()">拼图</button>
      <input ref="fileInput" type="file" accept="image/*" multiple hidden @change="uploadInsert" />
      <input ref="collageInput" type="file" accept="image/*" multiple hidden @change="uploadCollage" />
    </div>

    <!-- 编辑区（所见即所得） -->
    <EditorContent class="content" :editor="editor" />
  </div>
</template>

<style scoped lang="scss">
.rich {
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--surface);
  overflow: hidden;

  &:focus-within { border-color: var(--primary); }
}

.toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 2px;
  padding: 8px 10px;
  border-bottom: 1px solid var(--border);
  background: var(--surface-2);
  position: sticky;
  top: 0;
  z-index: 5;
}

.tool {
  min-width: 30px;
  height: 28px;
  padding: 0 8px;
  border: none;
  border-radius: 7px;
  background: none;
  color: var(--text-2);
  font-size: 12.5px;
  font-weight: 700;
  font-family: inherit;
  transition: all var(--dur-fast);

  &:hover { background: var(--surface); color: var(--text); }

  &.on {
    background: rgba(var(--primary-rgb), 0.14);
    color: var(--primary);
  }

  &.dim { opacity: 0.35; }
}

.divider {
  width: 1px;
  height: 18px;
  background: var(--border);
  margin: 0 6px;
}

.content {
  :deep(.ProseMirror) {
    min-height: 52vh;
    padding: 20px 24px;
    outline: none;
    font-size: 15px;
    line-height: 1.9;

    > * + * { margin-top: 0.6em; }

    h1, h2, h3 { font-family: var(--font-serif); line-height: 1.4; }
    h1 { font-size: 28px; }
    h2 { font-size: 22px; padding-left: 12px; border-left: 4px solid var(--primary); }
    h3 { font-size: 18px; }

    ul, ol { padding-left: 26px; }

    ul[data-type='taskList'] {
      list-style: none;
      padding-left: 4px;

      li {
        display: flex;
        gap: 8px;
        align-items: flex-start;

        > label { margin-top: 4px; }
        input[type='checkbox'] { accent-color: var(--primary); }
        > div { flex: 1; }

        &[data-checked='true'] > div {
          color: var(--text-2);
          text-decoration: line-through;
        }
      }
    }

    blockquote {
      border-left: 4px solid rgba(var(--primary-rgb), 0.5);
      padding: 6px 14px;
      color: var(--text-2);
      background: var(--surface-2);
      border-radius: 0 8px 8px 0;
    }

    code {
      font-family: Consolas, 'Courier New', monospace;
      font-size: 0.88em;
      background: var(--surface-2);
      padding: 2px 6px;
      border-radius: 6px;
    }

    pre {
      background: var(--surface-2);
      border-radius: 8px;
      padding: 14px;
      overflow-x: auto;

      code { background: none; padding: 0; }
    }

    table {
      border-collapse: collapse;
      width: 100%;
      table-layout: fixed;

      th, td {
        border: 1px solid var(--border);
        padding: 7px 12px;
        vertical-align: top;
        position: relative;
      }

      th { background: var(--surface-2); font-weight: 700; }

      .selectedCell::after {
        content: '';
        position: absolute;
        inset: 0;
        background: rgba(var(--primary-rgb), 0.12);
        pointer-events: none;
      }

      .column-resize-handle {
        position: absolute;
        right: -2px;
        top: 0;
        bottom: 0;
        width: 4px;
        background: var(--primary);
        cursor: col-resize;
      }
    }

    img {
      max-width: 100%;
      border-radius: 8px;

      &.ProseMirror-selectednode { outline: 2px solid var(--primary); }
    }

    hr {
      border: none;
      border-top: 1px solid var(--border);
      margin: 18px 0;
    }

    p.is-editor-empty:first-child::before {
      content: attr(data-placeholder);
      color: var(--text-2);
      float: left;
      height: 0;
      pointer-events: none;
    }
  }
}
</style>
