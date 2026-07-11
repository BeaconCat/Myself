<script setup lang="ts">
import MarkdownIt from 'markdown-it';
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { api, type Note } from '../api';
import ImageViewer from '../components/media/ImageViewer.vue';
import { useConfigStore } from '../stores/config';

const { t } = useI18n();
const config = useConfigStore();
const notes = ref<Note[]>([]);

/* Lightbox 状态 */
const viewerImages = ref<string[]>([]);
const viewerIndex = ref(0);
const viewerOpen = ref(false);

function openViewer(images: string[], index: number): void {
  viewerImages.value = images;
  viewerIndex.value = index;
  viewerOpen.value = true;
}

/** 拼图布局类：1 单图 / 2 双拼 / 3 三拼 / 4 四宫格 / ≥5 三列宫格 */
function gridClass(n: number): string {
  if (n === 1) return 'g1';
  if (n === 2) return 'g2';
  if (n === 3) return 'g3';
  if (n === 4) return 'g4';
  return 'gn';
}

const md = new MarkdownIt({ linkify: true });

function render(note: Note): string {
  return md.render(note.contentMd);
}

/** 相对时间：今天/昨天内显示口语化，更早显示日期 */
function timeOf(s: string): string {
  const then = new Date(s.replace(' ', 'T'));
  const diffMs = Date.now() - then.getTime();
  const hours = Math.floor(diffMs / 3.6e6);
  if (hours < 1) return t('thoughts.justNow');
  if (hours < 24) return t('thoughts.hoursAgo', { n: hours });
  const days = Math.floor(hours / 24);
  if (days < 7) return t('thoughts.daysAgo', { n: days });
  return s.slice(0, 10);
}

onMounted(async () => {
  notes.value = (await api.notes()).items;
});
</script>

<template>
  <main class="page">
    <h1 v-reveal class="page-title">{{ t('nav.thoughts') }}</h1>
    <p v-reveal class="page-sub">{{ config.cfg.thoughts.subtitle }}</p>

    <!-- X 风信息流 -->
    <div class="feed">
      <article v-for="note in notes" :key="note.id" v-reveal class="tweet">
        <img class="avatar" src="/favicon-64.png" alt="" draggable="false" />
        <div class="tweet-main">
          <header class="tweet-head">
            <strong>BeaconCat</strong>
            <span class="handle">{{ '@myself' }}</span>
            <span class="sep">·</span>
            <time>{{ timeOf(note.createdAt) }}</time>
            <span v-if="note.mood" class="mood">{{ note.mood }}</span>
          </header>
          <!-- 短内容 Markdown -->
          <div class="tweet-body markdown-mini" v-html="render(note)" />

          <!-- 配图拼图 -->
          <div
            v-if="note.images.length"
            class="pics"
            :class="gridClass(note.images.length)"
          >
            <button
              v-for="(src, i) in note.images"
              :key="src"
              class="pic"
              @click="openViewer(note.images, i)"
            >
              <img :src="src" alt="" loading="lazy" draggable="false" />
            </button>
          </div>
        </div>
      </article>
    </div>

    <ImageViewer
      v-if="viewerOpen"
      :images="viewerImages"
      :start-index="viewerIndex"
      @close="viewerOpen = false"
    />
  </main>
</template>

<style scoped lang="scss">
.page {
  max-width: 640px;
  margin: 0 auto;
  padding: 110px 24px 80px;
}

.page-title {
  font-size: clamp(30px, 4vw, 42px);
}

.page-sub {
  color: var(--text-2);
  font-size: 14px;
  margin: 8px 0 30px;
}

.feed {
  border-top: 1px solid var(--border);
}

.tweet {
  display: flex;
  gap: 14px;
  padding: 20px 12px;
  border-bottom: 1px solid var(--border);
  transition: background var(--dur-fast);

  &:hover { background: rgba(var(--primary-rgb), 0.04); }
}

.avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  flex-shrink: 0;
  border: 1px solid var(--border);
  background: var(--surface-2);
}

.tweet-main {
  flex: 1;
  min-width: 0;
}

.tweet-head {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  flex-wrap: wrap;

  strong { font-weight: 700; }

  .handle, .sep, time { color: var(--text-2); font-size: 13px; }

  .mood {
    margin-left: auto;
    font-size: 11px;
    font-weight: 600;
    padding: 2px 10px;
    background: rgba(var(--primary-rgb), 0.1);
    color: var(--primary);
  }
}

.tweet-body {
  margin-top: 6px;
  font-size: 15px;
  line-height: 1.8;

  :deep(p) { margin: 4px 0; }

  :deep(code) {
    font-family: Consolas, 'Courier New', monospace;
    font-size: 0.88em;
    background: var(--surface-2);
    padding: 2px 6px;
    border-radius: 6px;
  }

  :deep(strong) { color: var(--primary); }

  :deep(a) {
    color: var(--primary);
    border-bottom: 1px solid rgba(var(--primary-rgb), 0.35);
  }
}

/* ===== 拼图 ===== */
.pics {
  display: grid;
  gap: 4px;
  margin-top: 12px;
  border-radius: var(--radius);
  overflow: hidden;
  max-width: 480px;

  /* 单图：自然比例，限高 */
  &.g1 {
    grid-template-columns: 1fr;

    .pic { aspect-ratio: 16 / 10; }
  }

  &.g2 { grid-template-columns: repeat(2, 1fr); }
  &.g3 { grid-template-columns: repeat(3, 1fr); }
  &.g4 {
    grid-template-columns: repeat(2, 1fr);
    max-width: 380px;
  }
  /* 5–9 图：三列宫格 */
  &.gn { grid-template-columns: repeat(3, 1fr); }
}

.pic {
  position: relative;
  border: none;
  padding: 0;
  background: var(--surface-2);
  aspect-ratio: 1;
  overflow: hidden;
  cursor: zoom-in;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    transition: transform var(--dur) var(--ease-out), filter var(--dur-fast);
  }

  &:hover img {
    transform: scale(1.06);
    filter: brightness(1.06);
  }
}

@media (max-width: 768px) {
  .page { padding-top: 88px; }
}
</style>
