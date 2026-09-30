<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { api, type Post, type Tag } from '../api';
import PostFeature from '../components/post/PostFeature.vue';
import PostRow from '../components/post/PostRow.vue';
import ContentIcon from '../components/post/ContentIcon.vue';
import { dotted, isStuck, ymdOf } from '../components/post/content';
import { useConfigStore } from '../stores/config';
import { useLoadingStore } from '../stores/loading';

/**
 * 桌面文章列表（高密度）：标题与统计副标题同一行；搜索 + 标签 chips 吸顶（毛玻璃）；
 * 8 / 4 两栏——左首篇封面卡 + 紧凑缩略行（按月分组，触底续载），
 * 右栏吸顶概览卡：统计条（篇数 / 标签 / 最近更新）+ 可点击的标签分布条形图。
 * ?tag= 驱动筛选（详情页标签、⌘K 可深链）。
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

/* 右栏概览：统计条 + 标签分布（条长按最大计数归一） */
const stats = computed(() => [
  { k: t('dense.articles.posts'), v: allTotal.value ? String(allTotal.value) : '' },
  { k: t('dense.articles.tags'), v: tags.value.length ? String(tags.value.length) : '' },
  { k: t('dense.articles.updated'), v: latestAt.value ? dotted(latestAt.value).slice(5) : '' },
]);
const tagMax = computed(() => Math.max(1, ...tags.value.map((tg) => tg.count)));
let seq = 0;

function params(page: number) {
  return { page, pageSize: PAGE_SIZE, tag: activeTag.value || undefined, q: keyword.value.trim() || undefined, pinnedFirst: true };
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
    if (!activeTag.value && !keyword.value.trim()) allTotal.value = res.total;
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
  if (latestAt.value) return;
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
    // 置顶文章（接口排在最前）单独成组；其余按站点时区的年月分组
    const { y, m } = ymdOf(p.createdAt);
    const key = p.pinned ? 'pinned' : `${y}-${m}`;
    let g = out[out.length - 1];
    if (!g || g.key !== key) {
      g = { key, label: p.pinned ? t('content.articles.pinned') : t('content.articles.month', { y, m }), count: 0, items: [] };
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
/** 刷新恢复滚动位置、入场动画结束时不触发 scroll：主动补判（同随想页） */
function recheck(): void {
  requestAnimationFrame(onScroll);
  window.setTimeout(onScroll, 700);
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
  filterbar.value?.addEventListener('animationend', onScroll);
  recheck();
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
        <p v-else><span class="sk sk-line" style="width: 220px; height: 12px" /></p>
      </header>
    </div>

    <!-- 搜索 + 标签 chips：吸顶毛玻璃条 -->
    <div ref="filterbar" class="filterbar" :class="{ stuck }">
      <div class="wrap in">
        <label class="field">
          <ContentIcon name="search" />
          <input
            ref="input"
            v-model="keyword"
            type="search"
            :placeholder="t('articles.searchPlaceholder')"
            :aria-label="t('a11y.searchArticles')"
            @input="onSearch"
            @keydown.esc="($event.target as HTMLInputElement).blur()"
          />
          <kbd aria-hidden="true">/</kbd>
        </label>
        <div class="chips">
          <button class="chip" :class="{ on: !activeTag }" :aria-pressed="!activeTag" :aria-label="allTotal ? t('a11y.withCount', { label: t('articles.all'), n: allTotal }) : undefined" @click="pickTag('')">
            {{ t('articles.all') }}<em v-if="allTotal">{{ allTotal }}</em>
          </button>
          <button
            v-for="tg in tags"
            :key="tg.name"
            class="chip"
            :class="{ on: activeTag === tg.name }"
            :aria-pressed="activeTag === tg.name"
            :aria-label="t('a11y.withCount', { label: tg.name, n: tg.count })"
            @click="pickTag(tg.name)"
          >
            {{ tg.name }}<em>{{ tg.count }}</em>
          </button>
        </div>
      </div>
    </div>

    <div class="wrap body">
      <div class="main">
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

      <!-- 右栏概览：统计条 + 标签分布 -->
      <aside class="side rise">
        <section class="panel">
          <div class="stats">
            <div v-for="s in stats" :key="s.k" class="stat">
              <strong v-if="s.v">{{ s.v }}</strong>
              <span v-else class="sk num-sk" />
              <small>{{ s.k }}</small>
            </div>
          </div>
          <template v-if="tags.length">
            <h6>{{ t('dense.articles.tagDist') }}<small>{{ t('content.thoughts.moodKinds', { n: tags.length }) }}</small></h6>
            <div class="dist">
              <button
                v-for="tg in tags"
                :key="tg.name"
                type="button"
                :class="{ on: activeTag === tg.name }"
                :aria-pressed="activeTag === tg.name"
                :aria-label="t('a11y.withCount', { label: tg.name, n: tg.count })"
                @click="pickTag(activeTag === tg.name ? '' : tg.name)"
              >
                <span class="nm">{{ tg.name }}</span>
                <span class="track"><i :style="{ width: `${(tg.count / tagMax) * 100}%` }" /></span>
                <em>{{ tg.count }}</em>
              </button>
            </div>
          </template>
        </section>
      </aside>
    </div>
  </main>
</template>

<style scoped lang="scss">
.articles {
  --nav-h: 64px;
  /* 浮动导航之下留出页首空间 */
  padding: 64px 0 64px;
  min-height: 100vh;
}

.wrap {
  width: min(1200px, calc(100% - 80px));
  margin-inline: auto;
}

/* ---------- 页首：标题与统计副标题同一行，基线对齐 ---------- */
.ph {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 6px 18px;
  padding: 36px 0 14px;

  h1 {
    font-family: var(--font-serif);
    font-size: 40px;
    font-weight: 900;
    line-height: 1.2;
    letter-spacing: 0.01em;
    color: var(--text);
  }

  p {
    font-size: 15px;
    line-height: 1.7;
    color: var(--text-2);

    .sk-line { display: inline-block; }
  }
}

/* ---------- 8 / 4 两栏 ---------- */
.body {
  display: grid;
  grid-template-columns: minmax(0, 8fr) minmax(0, 4fr);
  gap: 40px;
  align-items: start;
}

.main { min-width: 0; }

/* 右栏概览卡：统计条 + 标签分布；吸顶于筛选条之下 */
.side {
  position: sticky;
  top: calc(var(--nav-h) + 76px);
  padding-top: 14px;
}

.panel {
  display: flex;
  flex-direction: column;
  padding: var(--card-pad);
  border-radius: var(--card-r);
  background: var(--card-bg);
  box-shadow: var(--card-shadow);

  h6 {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin: 22px 0 10px;
    font-family: var(--font-serif);
    font-size: 20px;
    font-weight: 700;
    color: var(--text);

    small { font-family: var(--font-sans); font-size: 13px; font-weight: 400; color: var(--text-3); }
  }
}

.stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--stat-gap);
  padding: var(--statbar-pad);
  border-radius: var(--statbar-r);
  background: var(--statbar-bg);
  box-shadow: var(--statbar-shadow);
}

.stat {
  min-width: 0;
  padding: var(--stat-pad);

  & + & { box-shadow: var(--stat-sep); }

  strong {
    display: block;
    font-family: var(--font-mono);
    font-size: 30px;
    line-height: 1.1;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--text);
    white-space: nowrap;
  }

  small {
    display: block;
    margin-top: 6px;
    font-size: 12.5px;
    color: var(--text-3);
    white-space: nowrap;
  }
}

.num-sk { display: block; width: 60%; height: 33px; border-radius: var(--r-xs); }

/* 简洁风格：页首更舒展；右栏概览去卡面，只留左侧一条发丝线与主栏分隔 */
:root[data-style='clean'] {
  .ph { padding: 52px 0 22px; }
  .side { padding: 14px 0 0 28px; box-shadow: inset 1px 0 0 var(--line); }
  .grp-h { padding-top: 30px; }
  .feat { margin-bottom: 20px; }
}

/* 标签分布：名称 | 撑满剩余宽度的条 | 计数；整行可点筛选，选中 = 抬升 + 轻染 */
.dist {
  display: flex;
  flex-direction: column;
  gap: 2px;

  button {
    display: grid;
    grid-template-columns: minmax(0, 72px) minmax(0, 1fr) 28px;
    align-items: center;
    gap: 12px;
    height: 38px;
    margin: 0 -10px;
    padding: 0 10px;
    border: 0;
    border-radius: var(--r-sm);
    background: none;
    font-size: 14px;
    color: var(--text-2);
    text-align: left;
    transition: background-color var(--dur-fast), color var(--dur-fast), box-shadow var(--dur-fast);

    &:hover { background: var(--fill); color: var(--text); }
    &:focus-visible { outline: none; box-shadow: var(--focus); }

    &.on {
      background: var(--lift);
      color: var(--lift-fg);
      box-shadow: var(--lift-shadow);

      i { background: var(--ink); }
    }
  }

  .nm {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;

    &::before {
      content: '#';
      margin-right: 3px;
      font-family: var(--font-mono);
      font-size: 0.92em;
      color: var(--text-3);
    }
  }

  .track {
    height: 8px;
    border-radius: var(--r-pill);
    background: var(--fill-2);
    overflow: hidden;
  }

  i {
    display: block;
    height: 100%;
    min-width: 8px;
    border-radius: inherit;
    background: color-mix(in oklab, var(--primary) 70%, var(--fill-2));
    transition: width var(--dur-slow) var(--ease-out), background-color var(--dur);
  }

  em {
    font-style: normal;
    font-family: var(--font-mono);
    font-size: 13px;
    text-align: right;
    color: var(--text-3);
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
    /* 吸顶条是通栏色带：实底。无模糊渲染的环境里半透明会让下方配图 / 正文的轮廓透出来 */
    background: var(--bg);
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
  box-shadow: inset 0 0 0 1px var(--line-2);
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
    font-size: 12px;
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
  padding: 22px 0 12px;
  font-size: 13px;
  letter-spacing: 0.06em;
  color: var(--text-3);

  b { font-weight: 500; color: var(--text-2); }
}

.feat { margin-bottom: 14px; }

.nores {
  padding: 72px 0;
  text-align: center;
  color: var(--text-3);
}

.list-end {
  padding: 28px 0 0;
  text-align: center;
  font-size: 12px;
  letter-spacing: 0.12em;
  color: var(--text-3);
}

.sentinel { height: 1px; }

/* ---------- 骨架 ---------- */
.skel .grp-h { align-items: center; }

.sk-cover {
  aspect-ratio: 2.6 / 1;
  border-radius: var(--r-lg);
}

.sk-rows { margin-top: 20px; }
.sk-rows.more { margin-top: 0; }

.sk-row {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 14px 0;

  & + & { box-shadow: inset 0 0.5px 0 var(--line); }

  .rt { flex: 1; min-width: 0; }
}

.sk-thumb {
  flex: none;
  width: 120px;
  height: 76px;
  border-radius: var(--r-md);
}

@media (max-width: 1100px) {
  .wrap { width: calc(100% - 64px); }
  .body { gap: 28px; }
  .field { width: 220px; }
  .stat strong { font-size: 26px; }
}

@media (max-width: 900px) {
  .body { grid-template-columns: minmax(0, 1fr); }
  .side { display: none; }
}
</style>
