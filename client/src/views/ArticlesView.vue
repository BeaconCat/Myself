<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { api, type Post, type Tag } from '../api';
import PostFeature from '../components/post/PostFeature.vue';
import PostRow from '../components/post/PostRow.vue';
import ContentIcon from '../components/post/ContentIcon.vue';
import { isStuck, ymdOf } from '../components/post/content';
import { useConfigStore } from '../stores/config';
import { useLoadingStore } from '../stores/loading';

/**
 * 桌面文章列表：大标题 + 统计副标题；搜索 + 标签 chips 吸顶（毛玻璃）；
 * 首篇大卡 + 其余缩略行，按月分组；触底续载。?tag= 驱动筛选（详情页标签、⌘K 可深链）。
 */
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const config = useConfigStore();

const PAGE_SIZE = 20;
const posts = ref<Post[]>([]);
const tags = ref<Tag[]>([]);
const total = ref(0);
const pageNo = ref(1);
/** 首屏：同形骨架 */
const booting = ref(true);
/** 筛选切换中：旧列表压暗 */
const switching = ref(false);
const loadingMore = ref(false);
const failed = ref(false);
/** 列表重建序号：换筛选后 key 变化触发逐项 rise */
const listSeq = ref(0);

const keyword = ref('');
const activeTag = computed(() => (typeof route.query.tag === 'string' ? route.query.tag : ''));

/* 统计：全站篇数 + 最近更新日期（与筛选无关） */
const allTotal = ref(0);
const latestAt = ref('');
const sub = computed(() => {
  if (!latestAt.value) return allTotal.value || booting.value ? '' : t('content.articles.subEmpty');
  const { m, d } = ymdOf(latestAt.value);
  return t('content.articles.sub', { n: allTotal.value, m, d });
});

const hasMore = computed(() => posts.value.length < total.value);
let seq = 0;

function params(page: number) {
  return { page, pageSize: PAGE_SIZE, tag: activeTag.value || undefined, q: keyword.value.trim() || undefined };
}

async function reload(): Promise<void> {
  const my = ++seq;
  failed.value = false;
  if (!booting.value) switching.value = true;
  try {
    const res = await api.posts(params(1));
    if (my !== seq) return;
    posts.value = res.items;
    total.value = res.total;
    pageNo.value = 1;
    listSeq.value += 1;
    if (!activeTag.value && !keyword.value.trim()) {
      allTotal.value = res.total;
      latestAt.value = res.items[0]?.createdAt ?? '';
    }
  } catch {
    if (my === seq) failed.value = true;
  } finally {
    if (my === seq) {
      booting.value = false;
      switching.value = false;
    }
  }
}

async function loadStats(): Promise<void> {
  if (allTotal.value) return;
  const all = await api.posts({ pageSize: 1 }).catch(() => null);
  if (!all) return;
  allTotal.value = all.total;
  latestAt.value = all.items[0]?.createdAt ?? '';
}

async function loadMore(): Promise<void> {
  if (booting.value || switching.value || loadingMore.value || !hasMore.value) return;
  loadingMore.value = true;
  const my = seq;
  try {
    const res = await api.posts(params(pageNo.value + 1));
    if (my !== seq) return;
    posts.value = [...posts.value, ...res.items];
    total.value = res.total;
    pageNo.value += 1;
  } finally {
    loadingMore.value = false;
  }
}

function pickTag(name: string): void {
  if (name === activeTag.value) return;
  void router.replace({ query: name ? { ...route.query, tag: name } : {} });
}

watch(activeTag, () => {
  keepFilterInView();
  void reload();
});

let debounce = 0;
function onSearch(): void {
  window.clearTimeout(debounce);
  debounce = window.setTimeout(() => {
    keepFilterInView();
    void reload();
  }, 280);
}

/** 已滚过标题区时，换筛选后停在吸顶条处，而不是被短列表弹回页首 */
const filterbar = ref<HTMLElement | null>(null);
const head = ref<HTMLElement | null>(null);
const NAV_H = 64;
function keepFilterInView(): void {
  if (!stuck.value || !head.value) return;
  const top = head.value.getBoundingClientRect().bottom + window.scrollY - NAV_H;
  void nextTick(() => window.scrollTo({ top: Math.max(0, top) }));
}

/* 按月分组：首篇大卡，其余缩略行 */
interface Group { key: string; label: string; count: number; items: { post: Post; feat: boolean }[] }
const groups = computed<Group[]>(() => {
  const out: Group[] = [];
  posts.value.forEach((p, i) => {
    const key = p.createdAt.slice(0, 7);
    let g = out[out.length - 1];
    if (!g || g.key !== key) {
      const { y, m } = ymdOf(p.createdAt);
      g = { key, label: t('content.articles.month', { y, m }), count: 0, items: [] };
      out.push(g);
    }
    g.count += 1;
    g.items.push({ post: p, feat: i === 0 });
  });
  return out;
});

const emptyText = computed(() => {
  const q = keyword.value.trim();
  if (failed.value) return t('content.articles.loadFailed');
  if (q) return t('content.articles.noResult', { q });
  if (activeTag.value) return t('content.articles.emptyTag');
  return t('articles.empty');
});

/* 吸顶态：贴住导航下缘时显出毛玻璃底 */
const stuck = ref(false);
function onScroll(): void {
  stuck.value = isStuck(filterbar.value, NAV_H);
}

/* 「/」聚焦搜索 */
const input = ref<HTMLInputElement | null>(null);
function onKey(e: KeyboardEvent): void {
  if (e.key !== '/' || e.metaKey || e.ctrlKey || e.altKey) return;
  const el = document.activeElement as HTMLElement | null;
  if (el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.isContentEditable)) return;
  e.preventDefault();
  input.value?.focus();
}

/* 触底哨兵 */
const sentinel = ref<HTMLElement | null>(null);
let io: IntersectionObserver | null = null;

onMounted(async () => {
  window.addEventListener('scroll', onScroll, { passive: true });
  window.addEventListener('keydown', onKey);
  io = new IntersectionObserver((es) => { if (es.some((e) => e.isIntersecting)) void loadMore(); }, { rootMargin: '400px' });
  if (sentinel.value) io.observe(sentinel.value);
  const release = useLoadingStore().holdRoute();
  try {
    const [, tagList] = await Promise.all([reload(), api.tags().catch(() => [] as Tag[])]);
    tags.value = tagList;
    await loadStats();
  } finally {
    release();
  }
  onScroll();
});

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll);
  window.removeEventListener('keydown', onKey);
  io?.disconnect();
  window.clearTimeout(debounce);
});
</script>

<template>
  <main class="articles">
    <div class="wrap">
      <header ref="head" class="ph rise-stagger">
        <h1>{{ t('nav.articles') }}</h1>
        <p v-if="sub">{{ sub }}</p>
        <p v-else><span class="sk sk-line" style="width: 220px" /></p>
      </header>
    </div>

    <!-- 搜索 + 标签 chips：吸顶毛玻璃条 -->
    <div ref="filterbar" class="filterbar" :class="{ stuck }">
      <div class="wrap in">
        <label class="field">
          <ContentIcon name="search" size="s" />
          <input
            ref="input"
            v-model="keyword"
            type="search"
            :placeholder="t('articles.searchPlaceholder')"
            @input="onSearch"
            @keydown.esc="($event.target as HTMLInputElement).blur()"
          />
          <kbd>/</kbd>
        </label>
        <div class="chips">
          <button class="chip" :class="{ on: !activeTag }" @click="pickTag('')">
            {{ t('articles.all') }}<em v-if="allTotal">{{ allTotal }}</em>
          </button>
          <button
            v-for="tg in tags"
            :key="tg.name"
            class="chip"
            :class="{ on: activeTag === tg.name }"
            @click="pickTag(tg.name)"
          >
            {{ tg.name }}<em>{{ tg.count }}</em>
          </button>
        </div>
      </div>
    </div>

    <div class="wrap">
      <!-- 首屏骨架：与真实列表同形（分组头 / 大卡 / 缩略行） -->
      <div v-if="booting" class="skel" aria-hidden="true">
        <div class="grp-h"><span class="sk sk-line" style="width: 96px" /><span class="sk sk-line" style="width: 36px" /></div>
        <div class="sk sk-cover" />
        <span class="sk sk-line" style="width: 180px; margin-top: 22px" />
        <span class="sk sk-line" style="width: 62%; height: 30px; margin-top: 12px" />
        <span class="sk sk-line" style="width: 84%; margin-top: 14px" />
        <div class="sk-rows">
          <div v-for="n in 3" :key="n" class="sk-row">
            <div class="rt">
              <span class="sk sk-line" style="width: 160px" />
              <span class="sk sk-line" style="width: 54%; height: 20px; margin-top: 10px" />
              <span class="sk sk-line" style="width: 88%; margin-top: 10px" />
            </div>
            <span class="sk sk-thumb" />
          </div>
        </div>
      </div>

      <div v-else :key="listSeq" class="list" :class="{ dim: switching }">
        <p v-if="!posts.length" class="nores rise">{{ emptyText }}</p>
        <section v-for="g in groups" :key="g.key" class="grp rise-stagger">
          <div class="grp-h"><b>{{ g.label }}</b><span>{{ t('content.articles.monthCount', { n: g.count }) }}</span></div>
          <template v-for="it in g.items" :key="it.post.slug">
            <PostFeature v-if="it.feat" class="feat" :post="it.post" />
            <PostRow v-else :post="it.post" />
          </template>
        </section>

        <div v-if="loadingMore" class="sk-rows more" aria-hidden="true">
          <div v-for="n in 2" :key="n" class="sk-row">
            <div class="rt">
              <span class="sk sk-line" style="width: 160px" />
              <span class="sk sk-line" style="width: 54%; height: 20px; margin-top: 10px" />
              <span class="sk sk-line" style="width: 88%; margin-top: 10px" />
            </div>
            <span class="sk sk-thumb" />
          </div>
        </div>
        <p v-else-if="!hasMore && posts.length" class="list-end">{{ config.cfg.site.listEndText }}</p>
      </div>

      <div ref="sentinel" class="sentinel" aria-hidden="true" />
    </div>
  </main>
</template>

<style scoped lang="scss">
.articles {
  --nav-h: 64px;
  /* 浮动导航之下留出页首空间（导航 64 + 页首 64） */
  padding: 64px 0 96px;
  min-height: 100vh;
}

.wrap {
  width: min(880px, calc(100% - 80px));
  margin-inline: auto;
}

/* ---------- 页首 ---------- */
.ph {
  padding: 64px 0 28px;

  h1 {
    font-family: var(--font-serif);
    font-size: 52px;
    font-weight: 900;
    line-height: 1.12;
    letter-spacing: 0.01em;
    color: var(--text);
  }

  p {
    min-height: 1.7em;
    margin-top: 12px;
    font-size: 15px;
    line-height: 1.7;
    color: var(--text-2);
  }
}

/* ---------- 吸顶筛选条 ---------- */
.filterbar {
  position: sticky;
  top: var(--nav-h);
  z-index: 30;
  margin: 0 0 8px;
  padding: 12px 0;

  /* 毛玻璃底向上铺到视口顶，与导航连成一整片；未吸顶时隐藏 */
  &::before {
    content: '';
    position: absolute;
    inset: calc(var(--nav-h) * -1) 0 0;
    background: color-mix(in oklab, var(--bg) 88%, transparent);
    backdrop-filter: blur(20px) saturate(170%);
    -webkit-backdrop-filter: blur(20px) saturate(170%);
    box-shadow: inset 0 -0.5px 0 var(--line);
    opacity: 0;
    transition: opacity var(--dur) var(--ease-out);
    pointer-events: none;
  }

  &.stuck::before { opacity: 1; }

  .in {
    position: relative;
    display: flex;
    align-items: center;
    gap: 16px;
  }
}

.field {
  display: flex;
  flex: none;
  align-items: center;
  gap: 9px;
  width: 280px;
  height: 40px;
  padding: 0 10px 0 13px;
  border-radius: var(--r-md);
  background: var(--fill);
  box-shadow: inset 0 0 0 1px var(--line);
  color: var(--text-3);
  cursor: text;
  transition: background-color var(--dur-fast), box-shadow var(--dur-fast), color var(--dur-fast);

  input {
    flex: 1;
    width: 0;
    min-width: 0;
    border: 0;
    outline: none;
    background: none;
    font: inherit;
    font-size: 14px;
    color: var(--text);
    caret-color: var(--ink);

    &::placeholder { color: var(--text-3); }
    &::-webkit-search-cancel-button { display: none; }
  }

  &:hover { box-shadow: inset 0 0 0 1px var(--line-2); }

  &:focus-within {
    background: var(--elev);
    color: var(--text-2);
    box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--ink) 70%, transparent),
      0 0 0 3px color-mix(in oklab, var(--ink) 18%, transparent);

    kbd { opacity: 0; }
  }
}

:root[data-mode='light'] .field:not(:focus-within) { background: color-mix(in oklab, var(--bg) 40%, white); }

kbd {
  display: inline-grid;
  place-items: center;
  min-width: 20px;
  height: 20px;
  padding: 0 5px;
  border-radius: var(--r-xs);
  font-family: var(--font-mono);
  font-size: 11px;
  line-height: 1;
  color: var(--text-2);
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line-2), inset 0 -1px 0 var(--line);
  transition: opacity var(--dur-fast);
}

.chips {
  display: flex;
  flex: 1;
  min-width: 0;
  gap: 8px;
  padding: 2px;
  overflow-x: auto;
  scrollbar-width: none;

  &::-webkit-scrollbar { display: none; }
}

/* chip：选中 = 抬升 + 轻染 + 前置 4px 主色圆点 */
.chip {
  position: relative;
  display: inline-flex;
  flex: none;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 14px;
  border: 0;
  border-radius: var(--r-pill);
  background: transparent;
  font-size: 13.5px;
  color: var(--text-2);
  box-shadow: inset 0 0 0 1px var(--line);
  transition: background-color var(--dur-fast), color var(--dur-fast), box-shadow var(--dur-fast),
    transform var(--dur-fast) var(--ease-spring);

  &::before {
    content: '';
    width: 0;
    height: 4px;
    margin-right: -6px;
    border-radius: 50%;
    background: var(--ink);
    opacity: 0;
    transition: width var(--dur) var(--ease-spring), margin var(--dur) var(--ease-spring), opacity var(--dur);
  }

  em {
    font-style: normal;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-3);
  }

  &:hover { color: var(--text); background: var(--fill); box-shadow: inset 0 0 0 1px var(--line-2); }
  &:active { transform: scale(0.96); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }

  &.on {
    background: var(--lift);
    color: var(--lift-fg);
    font-weight: 500;
    box-shadow: var(--lift-shadow);

    &::before { width: 4px; margin-right: 0; opacity: 1; }
    em { color: var(--text-2); }
  }
}

/* ---------- 列表 ---------- */
.list {
  transition: opacity 0.26s ease;

  &.dim { opacity: 0.5; pointer-events: none; }
}

.grp-h {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding: 30px 0 16px;
  font-size: 13px;
  letter-spacing: 0.06em;
  color: var(--text-3);

  b { font-weight: 500; color: var(--text-2); }
}

.feat { margin-bottom: 36px; }

.nores {
  padding: 72px 0;
  text-align: center;
  color: var(--text-3);
}

.list-end {
  padding: 36px 0 0;
  text-align: center;
  font-size: 12px;
  letter-spacing: 0.12em;
  color: var(--text-3);
}

.sentinel { height: 1px; }

/* ---------- 骨架 ---------- */
.skel .grp-h { align-items: center; }

.sk-cover {
  aspect-ratio: 2.1 / 1;
  border-radius: var(--r-xl);
}

.sk-rows { margin-top: 36px; }
.sk-rows.more { margin-top: 0; }

.sk-row {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 22px 0;

  & + & { box-shadow: inset 0 0.5px 0 var(--line); }

  .rt { flex: 1; min-width: 0; }
}

.sk-thumb {
  flex: none;
  width: 128px;
  height: 96px;
  border-radius: var(--r-md);
}

@media (max-width: 1100px) {
  .wrap { width: calc(100% - 64px); max-width: 880px; }
  .field { width: 220px; }
}
</style>
