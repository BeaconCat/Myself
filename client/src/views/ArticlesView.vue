<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { api, type Post, type Tag } from '../api';
import PostZigzagList from '../components/post/PostZigzagList.vue';
import { useConfigStore } from '../stores/config';

const { t } = useI18n();
const config = useConfigStore();

const posts = ref<Post[]>([]);
const tags = ref<Tag[]>([]);
const activeTag = ref('');
const keyword = ref('');
const loading = ref(false);
/** 旧列表整体淡出阶段 */
const exiting = ref(false);
/** key 重建触发逐条浮入 */
const listSeq = ref(0);

const page = ref(1);
const total = ref(0);
const PAGE_SIZE = 20;

function queryParams(nextPage: number) {
  return {
    page: nextPage,
    pageSize: PAGE_SIZE,
    tag: activeTag.value || undefined,
    q: keyword.value || undefined,
  };
}

/**
 * 切换筛选的丝滑序列：变暗加载 → 数据就绪 → 旧列表整体淡出 → 新列表逐条浮入
 */
async function load(): Promise<void> {
  loading.value = true;
  try {
    const res = await api.posts(queryParams(1));
    exiting.value = true;
    await new Promise((resolve) => window.setTimeout(resolve, 260));
    posts.value = res.items;
    total.value = res.total;
    page.value = 1;
    listSeq.value += 1;
    await nextTick();
    exiting.value = false;
  } finally {
    loading.value = false;
  }
}

function pickTag(tag: string): void {
  if (activeTag.value === tag && !keyword.value) return;
  activeTag.value = tag;
  void load();
}

let debounce = 0;
function onSearch(): void {
  window.clearTimeout(debounce);
  debounce = window.setTimeout(() => void load(), 300);
}

/* 分段加载：滚动触底追加，20/页 */
const loadingMore = ref(false);
const hasMore = computed(() => posts.value.length < total.value);

async function loadMore(): Promise<void> {
  if (loadingMore.value || loading.value || exiting.value || !hasMore.value) return;
  loadingMore.value = true;
  try {
    const res = await api.posts(queryParams(page.value + 1));
    posts.value = [...posts.value, ...res.items];
    total.value = res.total;
    page.value += 1;
  } finally {
    loadingMore.value = false;
  }
}

const sentinel = ref<HTMLElement | null>(null);
let observer: IntersectionObserver | null = null;

onMounted(async () => {
  observer = new IntersectionObserver(
    (entries) => {
      if (entries.some((e) => e.isIntersecting)) void loadMore();
    },
    { rootMargin: '400px' },
  );
  if (sentinel.value) observer.observe(sentinel.value);
  await load();
  tags.value = await api.tags();
});

onBeforeUnmount(() => {
  observer?.disconnect();
  window.clearTimeout(debounce);
});
</script>

<template>
  <main class="page">
    <h1 v-reveal class="page-title">{{ t('nav.articles') }}</h1>

    <!-- 工具栏：搜索 + 标签筛选一行 -->
    <div v-reveal class="toolbar">
      <div class="search">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round">
          <circle cx="11" cy="11" r="7" />
          <path d="M20 20l-3.8-3.8" />
        </svg>
        <input
          v-model="keyword"
          type="search"
          :placeholder="t('articles.searchPlaceholder')"
          @input="onSearch"
        />
      </div>

      <div class="tag-scroll">
        <button class="tag-chip" :class="{ on: !activeTag }" @click="pickTag('')">
          {{ t('articles.all') }}
        </button>
        <button
          v-for="tag in tags"
          :key="tag.name"
          class="tag-chip"
          :class="{ on: activeTag === tag.name }"
          @click="pickTag(tag.name)"
        >{{ tag.name }}<i>{{ tag.count }}</i></button>
      </div>
    </div>

    <!-- 列表：加载变暗 → 整体淡出 → 新列表逐条浮入 -->
    <div class="list-wrap" :class="{ loading, exiting }">
      <PostZigzagList :key="listSeq" :posts="posts" :alternate="false" compact />
    </div>

    <div ref="sentinel" class="sentinel" aria-hidden="true" />
    <p v-if="loadingMore" class="more-hint">{{ t('thoughts.loadingMore') }}</p>
    <p v-else-if="!loading && !hasMore && posts.length" class="end-text">
      {{ config.cfg.site.listEndText }}
    </p>
    <p v-if="!loading && !posts.length" class="empty">{{ t('articles.empty') }}</p>
  </main>
</template>

<style scoped lang="scss">
.page {
  max-width: 1080px;
  margin: 0 auto;
  padding: 110px 24px 80px;
}

.page-title {
  font-size: clamp(30px, 4vw, 42px);
  margin-bottom: 26px;
}

/* 工具栏 */
.toolbar {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 20px;
}

.search {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  width: min(300px, 42vw);
  padding: 9px 14px;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--surface);
  transition: border-color var(--dur-fast), box-shadow var(--dur-fast);

  svg {
    width: 17px;
    height: 17px;
    color: var(--text-2);
    flex-shrink: 0;
  }

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

.tag-scroll {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  scrollbar-width: none;
  /* 发光完整余量：上下 14px、左右 12px */
  padding: 14px 12px;
  margin: -14px -12px;

  &::-webkit-scrollbar { display: none; }
}

.tag-chip {
  flex-shrink: 0;
  font-size: 13px;
  font-weight: 600;
  padding: 7px 16px;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text-2);
  transition: all var(--dur-fast) var(--ease-out);

  i {
    font-style: normal;
    margin-left: 6px;
    font-size: 11px;
    opacity: 0.7;
  }

  &:hover { border-color: var(--primary); color: var(--primary); transform: scale(1.06); }

  &.on {
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    border-color: transparent;
    color: #fff;
    box-shadow: 0 2px 10px rgba(var(--primary-rgb), 0.45);
  }
}

/* 切换序列：变暗（加载中）→ 整体淡出（数据就绪）→ 新列表浮入 */
.list-wrap {
  transition: opacity 0.26s ease;

  &.loading { opacity: 0.55; }
  &.exiting { opacity: 0; }
}

.sentinel { height: 1px; }

.more-hint, .end-text {
  text-align: center;
  padding: 22px 0;
  font-size: 13px;
  color: var(--text-2);
}

.end-text { font-family: var(--font-serif); letter-spacing: 0.1em; }

.empty {
  color: var(--text-2);
  text-align: center;
  padding: 48px 0;
}

@media (max-width: 768px) {
  .page { padding-top: 88px; }

  .toolbar { flex-direction: column; align-items: stretch; }
  .search { width: 100%; }
}
</style>
