<script setup lang="ts">
import MarkdownIt from 'markdown-it';
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { api, type Note } from '../api';
import ImageViewer from '../components/media/ImageViewer.vue';
import { useConfigStore } from '../stores/config';
import { useLoadingStore } from '../stores/loading';

const { t } = useI18n();
const config = useConfigStore();
const notes = ref<Note[]>([]);

/* 帖文 / 媒体 双视图 + 搜索 + 时间筛选 + 分段加载（20/页，滚动续载） */
const tab = ref<'posts' | 'media'>('posts');
const keyword = ref('');
const loading = ref(false);
const page = ref(1);
const total = ref(0);
const PAGE_SIZE = 20;

/* 时间筛选 */
const dateOpen = ref(false);
const dateFrom = ref('');
const dateTo = ref('');

const hasDateFilter = computed(() => !!(dateFrom.value || dateTo.value));

function applyQuickRange(days: number): void {
  const end = new Date();
  const start = new Date(Date.now() - (days - 1) * 864e5);
  const fmt = (d: Date) => d.toISOString().slice(0, 10);
  dateFrom.value = fmt(start);
  dateTo.value = fmt(end);
  void reload();
}

function clearDate(): void {
  dateFrom.value = '';
  dateTo.value = '';
  void reload();
}

function queryParams(nextPage: number) {
  return {
    page: nextPage,
    pageSize: PAGE_SIZE,
    q: keyword.value || undefined,
    media: tab.value === 'media' || undefined,
    from: dateFrom.value || undefined,
    to: dateTo.value || undefined,
  };
}

async function reload(): Promise<void> {
  loading.value = true;
  try {
    const res = await api.notes(queryParams(1));
    notes.value = res.items;
    total.value = res.total;
    page.value = 1;
  } finally {
    loading.value = false;
  }
}

const loadingMore = ref(false);
const hasMore = computed(() => notes.value.length < total.value);

async function loadMore(): Promise<void> {
  if (loadingMore.value || loading.value || !hasMore.value) return;
  loadingMore.value = true;
  try {
    const res = await api.notes(queryParams(page.value + 1));
    notes.value = [...notes.value, ...res.items];
    total.value = res.total;
    page.value += 1;
  } finally {
    loadingMore.value = false;
  }
}

/* 触底哨兵 */
const sentinel = ref<HTMLElement | null>(null);
let observer: IntersectionObserver | null = null;

onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      if (entries.some((e) => e.isIntersecting)) void loadMore();
    },
    { rootMargin: '400px' },
  );
  if (sentinel.value) observer.observe(sentinel.value);
});

onBeforeUnmount(() => observer?.disconnect());

function switchTab(next: 'posts' | 'media'): void {
  if (tab.value === next) return;
  tab.value = next;
  void reload();
}

let debounce = 0;
function onSearch(): void {
  window.clearTimeout(debounce);
  debounce = window.setTimeout(() => void reload(), 300);
}

/** 媒体视图：铺平所有配图（X 风媒体墙） */
interface MediaCell {
  src: string;
  note: Note;
  index: number;
}

const mediaCells = computed<MediaCell[]>(() =>
  notes.value.flatMap((note) =>
    note.images.map((src, index) => ({ src, note, index })),
  ),
);

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
  const release = useLoadingStore().holdRoute();
  try {
    await reload();
  } finally {
    release();
  }
});
onBeforeUnmount(() => window.clearTimeout(debounce));
</script>

<template>
  <main class="page">
    <h1 v-reveal class="page-title">{{ t('nav.thoughts') }}</h1>
    <p v-reveal class="page-sub">{{ config.cfg.thoughts.subtitle }}</p>

    <!-- 工具栏：帖文/媒体胶囊 + 搜索 -->
    <div v-reveal class="toolbar">
      <div class="tabs">
        <button class="tab" :class="{ on: tab === 'posts' }" @click="switchTab('posts')">
          {{ t('thoughts.tabPosts') }}
        </button>
        <button class="tab" :class="{ on: tab === 'media' }" @click="switchTab('media')">
          {{ t('thoughts.tabMedia') }}
        </button>
      </div>
      <div class="search">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round">
          <circle cx="11" cy="11" r="7" />
          <path d="M20 20l-3.8-3.8" />
        </svg>
        <input
          v-model="keyword"
          type="search"
          :placeholder="t('thoughts.searchPlaceholder')"
          @input="onSearch"
        />
      </div>

      <!-- 时间选择器 -->
      <div class="date-wrap">
        <button
          class="date-btn"
          :class="{ on: hasDateFilter || dateOpen }"
          :title="t('thoughts.dateFilter')"
          @click="dateOpen = !dateOpen"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="5" width="18" height="16" rx="2" />
            <path d="M8 3v4M16 3v4M3 10h18" />
          </svg>
        </button>

        <transition name="pop">
          <div v-if="dateOpen" class="date-pop">
            <div class="dp-quick">
              <button @click="applyQuickRange(1)">{{ t('thoughts.today') }}</button>
              <button @click="applyQuickRange(7)">{{ t('thoughts.last7') }}</button>
              <button @click="applyQuickRange(30)">{{ t('thoughts.last30') }}</button>
              <button class="clear" @click="clearDate">{{ t('thoughts.clearDate') }}</button>
            </div>
            <div class="dp-range">
              <label>
                <span>{{ t('thoughts.dateFrom') }}</span>
                <input v-model="dateFrom" type="date" @change="reload" />
              </label>
              <label>
                <span>{{ t('thoughts.dateTo') }}</span>
                <input v-model="dateTo" type="date" @change="reload" />
              </label>
            </div>
          </div>
        </transition>
      </div>
    </div>

    <!-- 媒体墙（X 风） -->
    <div v-if="tab === 'media'" class="media-wall" :class="{ loading }">
      <button
        v-for="cell in mediaCells"
        :key="`${cell.note.id}-${cell.index}`"
        class="media-cell"
        @click="openViewer(cell.note.images, cell.index)"
      >
        <img :src="cell.src" loading="lazy" alt="" />
      </button>
      <p v-if="!loading && !mediaCells.length" class="empty">{{ t('thoughts.mediaEmpty') }}</p>
    </div>

    <!-- X 风信息流 -->
    <div v-else class="feed" :class="{ loading }">
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

    <!-- 触底续载 + 到底标注 -->
    <div ref="sentinel" class="sentinel" aria-hidden="true" />
    <p v-if="loadingMore" class="more-hint">{{ t('thoughts.loadingMore') }}</p>
    <p v-else-if="!hasMore && notes.length" class="end-text">{{ config.cfg.site.listEndText }}</p>

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

/* 工具栏 */
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 18px;
}

.tabs {
  display: flex;
  gap: 2px;
  padding: 4px;
  border-radius: 999px;
  background: var(--surface);
  border: 1px solid var(--border);
}

.tab {
  padding: 7px 20px;
  border: none;
  border-radius: 999px;
  background: none;
  color: var(--text-2);
  font-size: 13px;
  font-weight: 700;
  transition: all var(--dur-fast) var(--ease-out);

  &.on {
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    color: #fff;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.35);
    box-shadow: 0 2px 10px rgba(var(--primary-rgb), 0.45);
  }
}

.search {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  padding: 9px 14px;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--surface);
  transition: border-color var(--dur-fast), box-shadow var(--dur-fast);

  svg { width: 16px; height: 16px; color: var(--text-2); flex-shrink: 0; }

  input {
    flex: 1;
    min-width: 0;
    border: none;
    outline: none;
    background: none;
    color: var(--text);
    font-size: 14px;
    font-family: inherit;

    &::placeholder { color: var(--text-2); }
  }

  &:focus-within {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(var(--primary-rgb), 0.15);
  }
}

/* 媒体墙 */
.media-wall {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 4px;
  border-radius: var(--radius);
  overflow: hidden;
  transition: opacity var(--dur-fast);

  &.loading { opacity: 0.55; }
}

.media-cell {
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

  &:hover img { transform: scale(1.06); filter: brightness(1.06); }
}

.empty {
  grid-column: 1 / -1;
  color: var(--text-2);
  text-align: center;
  padding: 48px 0;
}

.feed {
  border-top: 1px solid var(--border);
  transition: opacity var(--dur-fast);

  &.loading { opacity: 0.55; }
}

/* 时间选择器 */
.date-wrap { position: relative; }

.date-btn {
  width: 40px;
  height: 40px;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text-2);
  display: grid;
  place-items: center;
  transition: all var(--dur-fast) var(--ease-out);

  svg { width: 17px; height: 17px; }

  &:hover { border-color: var(--primary); color: var(--primary); transform: scale(1.06); }

  &.on {
    border-color: rgba(var(--primary-rgb), 0.5);
    color: var(--primary);
    background: rgba(var(--primary-rgb), 0.08);
  }
}

.date-pop {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  z-index: 30;
  width: 260px;
  padding: 14px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--shadow), 0 16px 40px -12px rgba(var(--primary-rgb), 0.25);
}

.pop-enter-active, .pop-leave-active { transition: opacity var(--dur-fast), transform var(--dur-fast) var(--ease-out); }
.pop-enter-from, .pop-leave-to { opacity: 0; transform: translateY(-8px) scale(0.96); }

.dp-quick {
  display: flex;
  gap: 6px;
  margin-bottom: 12px;

  button {
    flex: 1;
    padding: 6px 0;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: none;
    color: var(--text-2);
    font-size: 12px;
    font-weight: 600;
    transition: all var(--dur-fast);

    &:hover { border-color: var(--primary); color: var(--primary); }
    &.clear:hover { border-color: var(--accent-red); color: var(--accent-red); }
  }
}

.dp-range {
  display: flex;
  flex-direction: column;
  gap: 10px;

  label {
    display: flex;
    align-items: center;
    gap: 10px;

    span { font-size: 12px; color: var(--text-2); width: 28px; flex-shrink: 0; }

    input {
      flex: 1;
      padding: 7px 10px;
      border-radius: 8px;
      border: 1px solid var(--border);
      background: var(--bg);
      color: var(--text);
      font-size: 13px;
      font-family: inherit;
      outline: none;

      &:focus { border-color: var(--primary); }
    }
  }
}

/* 触底与到底 */
.sentinel { height: 1px; }

.more-hint, .end-text {
  text-align: center;
  padding: 22px 0;
  font-size: 13px;
  color: var(--text-2);
}

.end-text { font-family: var(--font-serif); letter-spacing: 0.1em; }

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
