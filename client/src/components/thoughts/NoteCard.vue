<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Note } from '../../api';
import CoverArt from '../common/CoverArt.vue';
import EngageBar from '../engage/EngageBar.vue';
import { useRouter } from 'vue-router';
import { Pin } from 'lucide';
import Icon from '../ui/Icon.vue';
import { formatDateTime } from '../../utils/date';
import { render as renderMarkdown } from '../../utils/markdown';
import IdentityName from '../common/IdentityName.vue';

/**
 * 单条随想：头像 + 头部（名字 · @handle · 时间 · 右侧「# 心情」可点）+ Markdown 正文 + 配图宫格 + 互动栏。
 * 列表里点卡片空白处进入详情页（/thoughts/:id）；detail 模式为详情页本体，字号放大、不再跳转。
 * 配图 1 / 2 / 3 / 4 / 5–9 张对应单图、双拼、三拼、四宫格、九宫格；无图随想只有正文。
 * 服务端占位图与缺图统一走 CoverArt 光影构成。
 */
const props = defineProps<{ note: Note; avatar: string; name: string; alias?: string; handle: string; detail?: boolean }>();
const emit = defineEmits<{
  open: [images: string[], index: number, rect: DOMRect];
  mood: [mood: string];
}>();
const { t } = useI18n();
const router = useRouter();
const link = computed(() => `/thoughts/${props.note.id}`);

/** 点卡片空白处进入详情（链接、按钮、配图、选中文字时不跳） */
function onCardClick(e: MouseEvent): void {
  if (props.detail) return;
  if ((e.target as HTMLElement).closest('a, button, .igrid')) return;
  if (window.getSelection()?.toString()) return;
  void router.push(link.value);
}

const html = computed(() => renderMarkdown(props.note.contentMd, `note-${props.note.id}`));
const grid = computed(() => {
  const n = props.note.images.length;
  return n >= 5 ? 'n9' : `n${n}`;
});

function openAt(e: MouseEvent, i: number): void {
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
  emit('open', props.note.images, i, rect);
}
</script>

<template>
  <article class="post" :class="{ detail, link: !detail }" @click="onCardClick">
    <span class="av"><img :src="avatar" alt="" draggable="false" /></span>
    <div class="main">
      <header class="hd">
        <b><IdentityName :name="name" :alias="alias" /></b>
        <span v-if="handle" class="handle">@{{ handle }}</span>
        <span class="dotsep" />
        <time :datetime="`${note.createdAt.replace(' ', 'T')}Z`" :title="formatDateTime(note.createdAt, true)">{{ formatDateTime(note.createdAt, true) }}</time>
        <span v-if="note.pinned" class="pin"><Icon :icon="Pin" :size="12" :stroke="2" />{{ t('noteDetail.pinned') }}</span>
        <button v-if="note.mood" type="button" class="tag" @click="emit('mood', note.mood)">{{ note.mood }}</button>
      </header>
      <!-- eslint-disable-next-line vue/no-v-html -->
      <div class="body markdown-content" v-html="html" />

      <div v-if="note.images.length" class="igrid" :class="grid">
        <button
          v-for="(src, i) in note.images"
          :key="`${i}-${src}`"
          type="button"
          class="cell"
          :aria-label="t('a11y.viewImage', { n: i + 1, total: note.images.length })"
          @click="openAt($event, i)"
        >
          <CoverArt :src="src" :seed="`${note.id}-${i}`" pool="all" thumb />
        </button>
      </div>

      <EngageBar class="act" target="note" :id="note.id" :link="link" :text="note.contentMd" :big="detail" @comment="detail || router.push(link + '#comments')" />
    </div>
  </article>
</template>

<style scoped lang="scss">
.post {
  position: relative;
  isolation: isolate;
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr);
  gap: 16px;
  padding: 18px 0 10px;

  & + &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 60px;
    right: 0;
    height: 0.5px;
    background: var(--line-2);
  }
}

.av {
  width: 44px;
  height: 44px;
  overflow: hidden;
  border-radius: 50%;
  background: #060b16;
  box-shadow: 0 0 0 0.5px rgb(255 255 255 / 0.14);

  img { display: block; width: 100%; height: 100%; object-fit: cover; }
}

.main { min-width: 0; }

.hd {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  min-width: 0;
  font-size: 13px;
  color: var(--text-3);

  b { font-size: 15px; font-weight: 500; color: var(--text); }
  time { white-space: nowrap; }
  .handle { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
}

@media (max-width: 767px) {
  .hd time { flex-basis: 100%; order: 1; font-size: 12px; }
  .hd .dotsep { display: none; }
}

.dotsep {
  flex: none;
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: currentColor;
  opacity: 0.6;
}

/* 心情：「# 名称」纯文字，可点击按心情搜索 */
/* 置顶：主色轻染小胶囊 */
.pin {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 1px 8px 1px 6px;
  border-radius: var(--r-pill);
  background: color-mix(in oklab, var(--ink) 12%, transparent);
  color: var(--ink);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

.tag {
  flex: none;
  margin-left: auto;
  padding: 2px 4px;
  border: 0;
  border-radius: var(--r-xs);
  background: none;
  font-size: 13px;
  color: var(--text-2);
  transition: color var(--dur-fast);

  &::before {
    content: '#';
    margin-right: 3px;
    font-family: var(--font-mono);
    font-size: 0.92em;
    color: var(--text-3);
  }

  &:hover { color: var(--ink); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

.body { margin-top: 8px; user-select: text; }

/* ---------- 配图宫格 ---------- */
.igrid {
  display: grid;
  gap: 4px;
  max-width: 540px;
  margin-top: 12px;
  overflow: hidden;
  isolation: isolate;
  border-radius: var(--r-lg);

  &.n1 { grid-template-columns: 1fr; max-width: 600px; }
  &.n1 .cell { aspect-ratio: 2 / 1; }
  &.n2 { grid-template-columns: 1fr 1fr; }
  &.n3 { grid-template-columns: repeat(3, 1fr); }
  &.n4 { grid-template-columns: 1fr 1fr; max-width: 480px; }
  &.n9 { grid-template-columns: repeat(3, 1fr); max-width: 520px; }
}

.cell {
  position: relative;
  aspect-ratio: 1;
  overflow: hidden;
  padding: 0;
  border: 0;
  background: #040914;
  cursor: zoom-in;

  &:hover :deep(.cv) { transform: scale(1.05); }
  &:focus-visible { outline: none; box-shadow: inset 0 0 0 2px var(--ink); }
}

/* ---------- 互动栏 ---------- */
.act { margin-top: 6px; }

/* 列表：整卡可点进入详情，悬停时底色轻染（左右渐隐） */
.post.link {
  cursor: pointer;

  &::after {
    content: '';
    position: absolute;
    inset: 0 -16px;
    z-index: -1;
    border-radius: var(--r-md);
    background: linear-gradient(90deg, transparent, var(--fill) 10%, var(--fill) 90%, transparent);
    opacity: 0;
    transition: opacity var(--dur-fast);
  }

  &:hover::after { opacity: 0.6; }
}

/* 详情：正文放大 */
.post.detail {
  .body { font-size: 18px; line-height: 1.85; }
}
</style>
