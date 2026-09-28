<script setup lang="ts">
/* 正文宋体常规字重（全站只预载 700；详情分包内按需加载，unicode-range 切片） */
import '@fontsource/noto-serif-sc/400.css';
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { api, type Post } from '../api';
import CoverArt from '../components/common/CoverArt.vue';
import CoverAccordion from '../components/post/CoverAccordion.vue';
import ArticleToc from '../components/post/ArticleToc.vue';
import ContentIcon from '../components/post/ContentIcon.vue';
import ContentToast, { showToast } from '../components/post/ContentToast.vue';
import { copyText, dotted, readMinutes, wordCount } from '../components/post/content';
import { useConfigStore } from '../stores/config';
import { useLoadingStore } from '../stores/loading';
import { renderWithToc, type TocItem } from '../utils/markdown';

/**
 * 桌面文章详情：封面带（多封面手风琴 / 光影构成）→ 标签·日期·时长 → 宋体标题 → 导语 → 署名
 * → 正文 → 描边标签 → 上一篇 / 下一篇；右栏阅读卡（统计条 + 进度条 + 目录 2px 信号线），顶部 2px 阅读进度；
 * 窄于 960 时右栏收起为底部阅读浮条（与移动端同构）。
 */
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const config = useConfigStore();

const NAV_H = 64;
const post = ref<Post | null>(null);
const prev = ref<Post | null>(null);
const next = ref<Post | null>(null);
const loading = ref(true);
const notFound = ref(false);
const proseEl = ref<HTMLElement | null>(null);

const BIG_KEY = 'myself.bigType';
const bigType = ref(localStorage.getItem(BIG_KEY) === '1');

/** 正文首个一级标题与文章标题重复时去掉（标题已单独排版） */
const rendered = computed(() => {
  const src = post.value?.contentMd ?? '';
  const title = post.value?.title.trim() ?? '';
  const body = src.replace(/^\s*#\s+(.+)\n+/, (m, h: string) => (h.trim() === title ? '' : m));
  return body ? renderWithToc(body) : { html: '', toc: [] as TocItem[] };
});
const minutes = computed(() => readMinutes(post.value?.contentMd ?? ''));
const words = computed(() => wordCount(post.value?.contentMd ?? ''));
const avatar = computed(() => config.cfg.about.avatar || '/favicon-256.png');

/* ---------- 数据 ---------- */
let seq = 0;

async function load(slug: string): Promise<void> {
  const my = ++seq;
  loading.value = true;
  notFound.value = false;
  prev.value = null;
  next.value = null;
  try {
    post.value = await api.post(slug);
  } catch {
    if (my === seq) {
      post.value = null;
      notFound.value = true;
    }
  } finally {
    if (my === seq) loading.value = false;
  }
  if (my !== seq || !post.value) return;
  await nextTick();
  onScroll();
  scrollToHash();
  void loadSiblings(slug, my);
}

/** 相邻文章：按时间倒序翻页定位（上一篇 = 更早，下一篇 = 更新） */
async function loadSiblings(slug: string, my: number): Promise<void> {
  let before: Post | null = null;
  for (let page = 1; page <= 10; page++) {
    const res = await api.posts({ page, pageSize: 50 }).catch(() => null);
    if (!res || my !== seq) return;
    const idx = res.items.findIndex((p) => p.slug === slug);
    if (idx >= 0) {
      next.value = idx > 0 ? res.items[idx - 1] : before;
      if (idx + 1 < res.items.length) prev.value = res.items[idx + 1];
      else if (page * 50 < res.total) {
        const more = await api.posts({ page: page + 1, pageSize: 50 }).catch(() => null);
        if (my === seq) prev.value = more?.items[0] ?? null;
      }
      return;
    }
    before = res.items[res.items.length - 1] ?? null;
    if (page * 50 >= res.total) return;
  }
}

/* ---------- 滚动：阅读进度 / 当前节 ---------- */
const rp = ref(0);
const activeId = ref('');

function onScroll(): void {
  const el = proseEl.value;
  if (!el) return;
  const r = el.getBoundingClientRect();
  const span = r.height - window.innerHeight + NAV_H + 120;
  rp.value = Math.min(1, Math.max(0, (NAV_H + 40 - r.top) / Math.max(1, span)));
  let cur = '';
  for (const h of rendered.value.toc) {
    const node = document.getElementById(h.id);
    if (node && node.getBoundingClientRect().top <= NAV_H + 80) cur = h.id;
    else break;
  }
  activeId.value = cur || rendered.value.toc[0]?.id || '';
}

const left = computed(() =>
  rp.value > 0.98
    ? t('content.article.done')
    : t('content.article.left', { n: Math.max(1, Math.ceil(minutes.value * (1 - rp.value))) }),
);

function scrollToId(id: string, smooth = true): void {
  const node = document.getElementById(id);
  if (!node) return;
  const top = node.getBoundingClientRect().top + window.scrollY - NAV_H - 24;
  window.scrollTo({ top, behavior: smooth ? 'smooth' : 'auto' });
  history.replaceState(history.state, '', `#${encodeURIComponent(id)}`);
}

function scrollToHash(): void {
  const id = decodeURIComponent(route.hash.slice(1));
  if (id) window.setTimeout(() => scrollToId(id, false), 60);
}

function toTop(): void {
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

/* ---------- 动作 ---------- */
function toggleType(): void {
  bigType.value = !bigType.value;
  localStorage.setItem(BIG_KEY, bigType.value ? '1' : '0');
  showToast(bigType.value ? t('content.article.bigType') : t('content.article.normalType'));
  void nextTick(onScroll);
}

async function copyLink(): Promise<void> {
  const ok = await copyText(window.location.href.split('#')[0]);
  showToast(ok ? t('content.article.linkCopied') : t('content.article.copyFailed'));
}

/** 正文内锚点平滑滚动；站内链接走路由 */
function onProseClick(e: MouseEvent): void {
  const a = (e.target as HTMLElement).closest('a');
  if (!a) return;
  const href = a.getAttribute('href') ?? '';
  if (href.startsWith('#')) {
    e.preventDefault();
    scrollToId(decodeURIComponent(href.slice(1)));
  } else if (/^\/(?!\/)/.test(href) && !href.startsWith('/uploads') && !href.startsWith('/feed')) {
    e.preventDefault();
    void router.push(href);
  }
}

/* ---------- 窄屏阅读浮条 + 目录弹层 ---------- */
const dockToc = ref(false);

function onDocClick(e: MouseEvent): void {
  if (!dockToc.value) return;
  if (!(e.target as HTMLElement).closest('.dock')) dockToc.value = false;
}

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape') dockToc.value = false;
}

function dockJump(id: string): void {
  dockToc.value = false;
  scrollToId(id);
}

/* ---------- 生命周期 ---------- */
watch(
  () => route.params.slug,
  (slug, old) => {
    if (typeof slug !== 'string') return;
    if (old !== undefined) window.scrollTo({ top: 0 });
    void load(slug);
  },
);

onMounted(async () => {
  window.addEventListener('scroll', onScroll, { passive: true });
  window.addEventListener('resize', onScroll);
  document.addEventListener('click', onDocClick);
  window.addEventListener('keydown', onKey);
  const slug = route.params.slug;
  if (typeof slug !== 'string') return;
  const release = useLoadingStore().holdRoute();
  try {
    await load(slug);
  } finally {
    release();
  }
});

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll);
  window.removeEventListener('resize', onScroll);
  document.removeEventListener('click', onDocClick);
  window.removeEventListener('keydown', onKey);
});

const C = 2 * Math.PI * 9;
</script>

<template>
  <main class="dt" :style="{ '--rp': rp }">
    <!-- 顶部 2px 阅读进度 -->
    <div class="prog" :class="{ on: !!post && rp > 0.002 }"><i /></div>

    <!-- 骨架：与真实布局同形 -->
    <div v-if="loading" class="skel" aria-hidden="true">
      <div class="band sk" />
      <div class="grid">
        <div>
          <div class="head">
            <span class="sk sk-line" style="width: 220px" />
            <span class="sk sk-line" style="width: 72%; height: 46px; margin-top: 18px" />
            <span class="sk sk-line" style="width: 94%; margin-top: 22px" />
            <span class="sk sk-line" style="width: 64%; margin-top: 10px" />
            <div class="by"><span class="sk sk-av" /><span class="sk sk-line" style="width: 140px" /></div>
          </div>
          <div class="sk-prose">
            <span v-for="(w, i) in [96, 90, 98, 72, 0, 40, 94, 88, 97, 66]" :key="i" class="sk sk-line" :class="{ gap: !w }" :style="{ width: `${w}%` }" />
          </div>
        </div>
        <div class="sk-toc">
          <span class="sk sk-line" style="width: 40px" />
          <span v-for="n in 4" :key="n" class="sk sk-line" :style="{ width: `${[80, 60, 70, 50][n - 1]}%`, marginTop: '14px' }" />
        </div>
      </div>
    </div>

    <template v-else-if="post">
      <div class="band rise">
        <div class="cover">
          <CoverAccordion v-if="post.covers.length > 1" :images="post.covers" />
          <CoverArt v-else :src="post.covers[0]" :seed="post.slug" />
        </div>
        <div class="crumbs">
          <router-link to="/articles" class="cbtn" :aria-label="t('content.article.back')" :title="t('content.article.back')">
            <ContentIcon name="arrowL" />
          </router-link>
          <div class="r">
            <button class="cbtn" :class="{ on: bigType }" :aria-label="t('content.article.fontSize')" :title="t('content.article.fontSize')" @click="toggleType">
              <ContentIcon name="type" />
            </button>
            <button class="cbtn" :aria-label="t('content.article.share')" :title="t('content.article.share')" @click="copyLink">
              <ContentIcon name="share" />
            </button>
          </div>
        </div>
      </div>

      <div class="grid">
        <div class="col">
          <header class="head">
            <div class="meta rise" style="--i: 1">
              <router-link v-if="post.tags[0]" class="tag" :to="{ path: '/articles', query: { tag: post.tags[0] } }">{{ post.tags[0] }}</router-link>
              <time>{{ dotted(post.createdAt) }}</time>
              <span class="dotsep" />
              <span>{{ t('content.article.minutes', { n: minutes }) }}</span>
            </div>
            <h1 class="rise" style="--i: 2">{{ post.title }}</h1>
            <p v-if="post.excerpt" class="lede rise" style="--i: 3">{{ post.excerpt }}</p>
            <div class="by rise" style="--i: 4">
              <span class="av"><img :src="avatar" alt="" draggable="false" /></span>
              <div>
                <b>{{ config.cfg.about.name }}</b>
                <small>{{ config.cfg.about.motto }}</small>
              </div>
              <span class="grow" />
              <span class="words">{{ t('content.article.words', { n: words }) }}</span>
            </div>
          </header>

          <!-- eslint-disable-next-line vue/no-v-html -->
          <article ref="proseEl" class="prose rise" :class="{ big: bigType }" style="--i: 5" @click="onProseClick" v-html="rendered.html" />

          <footer class="end rise" style="--i: 6">
            <div class="row1">
              <div class="tags">
                <router-link v-for="tag in post.tags" :key="tag" class="tag-o" :to="{ path: '/articles', query: { tag } }">{{ tag }}</router-link>
              </div>
              <button class="btn-secondary" @click="copyLink"><ContentIcon name="link" size="s" />{{ t('content.article.copyLink') }}</button>
            </div>
            <nav v-if="prev || next" class="pn">
              <router-link v-if="prev" :to="`/articles/${prev.slug}`" class="pv">
                <div class="th"><CoverArt :src="prev.covers[0]" :seed="prev.slug" thumb /></div>
                <div class="tx">
                  <small><ContentIcon name="arrowL" size="xs" />{{ t('content.article.prev') }}</small>
                  <b>{{ prev.title }}</b>
                </div>
              </router-link>
              <span v-else />
              <router-link v-if="next" :to="`/articles/${next.slug}`" class="nx">
                <div class="th"><CoverArt :src="next.covers[0]" :seed="next.slug" thumb /></div>
                <div class="tx">
                  <small>{{ t('content.article.next') }}<ContentIcon name="arrowR" size="xs" /></small>
                  <b>{{ next.title }}</b>
                </div>
              </router-link>
            </nav>
          </footer>
        </div>

        <ArticleToc
          class="side rise"
          style="--i: 3"
          :items="rendered.toc"
          :active="activeId"
          :progress="rp"
          :left="left"
          :words="words"
          :minutes="minutes"
          @jump="scrollToId"
          @top="toTop"
        />
      </div>

      <!-- 窄屏阅读浮条：进度环 + 剩余时间（回顶）| 目录 -->
      <div class="dock" :class="{ open: dockToc }">
        <Transition name="dpop">
          <div v-if="dockToc && rendered.toc.length" class="dpop">
            <button
              v-for="it in rendered.toc"
              :key="it.id"
              :class="{ on: it.id === activeId, l3: it.level === 3 }"
              @click="dockJump(it.id)"
            >
              {{ it.text }}
            </button>
          </div>
        </Transition>
        <button class="dk" @click="toTop">
          <svg class="ring" viewBox="0 0 24 24" aria-hidden="true">
            <circle class="bgc" cx="12" cy="12" r="9" />
            <circle class="fgc" cx="12" cy="12" r="9" :style="{ strokeDasharray: C, strokeDashoffset: C * (1 - rp) }" />
          </svg>
          <b>{{ Math.round(rp * 100) }}%</b>
          <span>{{ left }}</span>
        </button>
        <template v-if="rendered.toc.length">
          <span class="sep" />
          <button class="dk" :class="{ on: dockToc }" @click.stop="dockToc = !dockToc">
            <ContentIcon name="list" size="s" />{{ t('article.toc') }}
          </button>
        </template>
      </div>
    </template>

    <div v-else-if="notFound" class="nf rise-stagger">
      <p>{{ t('article.notFound') }}</p>
      <router-link to="/articles" class="btn-secondary"><ContentIcon name="arrowL" size="s" />{{ t('content.article.back') }}</router-link>
    </div>

    <ContentToast />
  </main>
</template>

<style scoped lang="scss">
.dt {
  --nav-h: 64px;
  --col: 720px;
  --toc: 300px;
  --gap: 48px;
  --outer: min(calc(var(--col) + var(--toc) + var(--gap)), calc(100% - 80px));
  padding: var(--nav-h) 0 64px;
  min-height: 100vh;
}

/* ---------- 顶部阅读进度 ---------- */
.prog {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 200;
  height: 2px;
  opacity: 0;
  pointer-events: none;
  transition: opacity var(--dur);

  i {
    position: absolute;
    inset: 0;
    border-radius: 0 var(--r-pill) var(--r-pill) 0;
    background: var(--ink);
    transform: scaleX(var(--rp, 0));
    transform-origin: left;
  }

  &.on { opacity: 1; }
}

/* ---------- 封面带 ---------- */
.band {
  position: relative;
  width: var(--outer);
  height: 280px;
  margin: 20px auto 0;
  border-radius: var(--r-lg);
}

.cover {
  position: absolute;
  inset: 0;
  overflow: hidden;
  isolation: isolate;
  border-radius: inherit;
  background: #040914;

  :deep(.cv) { transform: scale(1.04); }

  &::after {
    content: '';
    position: absolute;
    inset: 0;
    z-index: 3;
    border-radius: inherit;
    box-shadow: inset 0 0 0 0.5px rgb(255 255 255 / 0.08);
    pointer-events: none;
  }
}

:root[data-mode='light'] .cover::after { box-shadow: inset 0 0 0 0.5px rgb(16 24 40 / 0.1); }

.crumbs {
  position: absolute;
  top: 18px;
  left: 18px;
  right: 18px;
  z-index: 4;
  display: flex;
  justify-content: space-between;
  pointer-events: none;

  .r { display: flex; gap: 8px; }
}

/* 封面上的圆形按钮：深色毛玻璃（封面恒为深色，与模式无关） */
.cbtn {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  border: 0;
  border-radius: 50%;
  color: #e8ecf5;
  background: rgb(8 12 22 / 0.45);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  box-shadow: inset 0 0 0 0.5px rgb(255 255 255 / 0.12);
  pointer-events: auto;
  transition: background-color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  &:hover { background: rgb(8 12 22 / 0.62); }
  &:active { transform: scale(0.94); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
  &.on { background: rgb(255 255 255 / 0.2); }
}

/* ---------- 双栏 ---------- */
.grid {
  display: grid;
  grid-template-columns: minmax(0, var(--col)) var(--toc);
  gap: var(--gap);
  width: var(--outer);
  margin-inline: auto;
}

.head {
  padding: 28px 0 0;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0 12px;
  font-size: 13px;
  color: var(--text-3);
}

.dotsep {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: currentColor;
  opacity: 0.7;
}

.tag {
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
}

h1 {
  margin-top: 10px;
  font-family: var(--font-serif);
  font-size: 40px;
  font-weight: 900;
  line-height: 1.25;
  color: var(--text);
  text-wrap: balance;
}

.lede {
  margin-top: 12px;
  font-family: var(--font-serif);
  font-size: 18px;
  font-weight: 400;
  line-height: 1.8;
  color: var(--text-2);
}

.by {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 20px;
  padding-bottom: 20px;
  box-shadow: inset 0 -0.5px 0 var(--line-2);
  font-size: 15px;

  b { display: block; font-weight: 500; color: var(--text); line-height: 1.5; }
  small { display: block; font-size: 13px; color: var(--text-3); }
  .grow { flex: 1; }
  .words { font-size: 13px; color: var(--text-3); }
}

.av {
  flex: none;
  width: 40px;
  height: 40px;
  overflow: hidden;
  border-radius: 50%;
  background: #060b16;
  box-shadow: 0 0 0 0.5px rgb(255 255 255 / 0.14);

  img { display: block; width: 100%; height: 100%; object-fit: cover; }
}

/* ---------- 正文排版 ---------- */
.prose {
  padding-top: 26px;
  font-family: var(--font-serif);
  font-size: 17.5px;
  font-weight: 400;
  line-height: 1.95;
  color: color-mix(in oklab, var(--text) 90%, var(--bg));
  overflow-wrap: anywhere;
  transition: font-size var(--dur) var(--ease-out);

  &.big { font-size: 19.5px; }

  :deep(h1),
  :deep(h2),
  :deep(h3),
  :deep(h4) {
    font-family: var(--font-serif);
    font-weight: 700;
    color: var(--text);
    scroll-margin-top: 90px;
  }

  :deep(h1) { margin: 48px 0 14px; font-size: 30px; line-height: 1.4; }
  :deep(h2) { margin: 48px 0 12px; font-size: 25px; line-height: 1.45; }
  :deep(h3) { margin: 32px 0 8px; font-size: 20px; line-height: 1.5; }
  :deep(h4) { margin: 24px 0 6px; font-size: 17.5px; }
  > :deep(:first-child) { margin-top: 0; }

  :deep(p) { margin: 0 0 18px; }
  :deep(strong) { color: var(--text); font-weight: 700; }
  :deep(hr) { height: 0.5px; margin: 40px 0; border: 0; background: var(--line-2); }

  :deep(a) {
    color: var(--ink);
    text-decoration: underline;
    text-decoration-color: color-mix(in oklab, var(--ink) 35%, transparent);
    text-decoration-thickness: 1px;
    text-underline-offset: 4px;
    transition: text-decoration-color var(--dur-fast);

    &:hover { text-decoration-color: currentColor; }
  }

  :deep(code:not(pre code)) {
    padding: 2px 6px;
    border-radius: var(--r-xs);
    background: var(--fill-2);
    font-family: var(--font-mono);
    font-size: 0.8em;
    color: var(--text);
  }

  :deep(pre) {
    margin: 8px 0 24px;
    padding: 20px 22px;
    overflow-x: auto;
    border-radius: var(--r-md);
    background: var(--hl-bg);
    box-shadow: inset 0 0 0 0.5px var(--line);
    font-size: 13px;
    line-height: 1.8;
    color: color-mix(in oklab, var(--text) 86%, var(--bg));

    code { background: none; padding: 0; font-size: inherit; }
  }

  :deep(blockquote) {
    margin: 26px 0;
    padding: 2px 0 2px 22px;
    background: linear-gradient(var(--ink), var(--ink)) left / 2px 100% no-repeat;
    font-size: 1.08em;
    font-weight: 600;
    line-height: 1.8;
    color: var(--text);

    p:last-child { margin-bottom: 0; }
  }

  :deep(ul),
  :deep(ol) { margin: 0 0 18px; list-style: none; }

  :deep(li) { position: relative; margin-bottom: 8px; padding-left: 22px; }
  :deep(li > ul), :deep(li > ol) { margin: 8px 0 0; }

  :deep(ul > li::before) {
    content: '';
    position: absolute;
    left: 5px;
    top: 0.85em;
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--text-3);
  }

  :deep(ol) { counter-reset: li; }
  :deep(ol > li) { counter-increment: li; }

  :deep(ol > li::before) {
    content: counter(li);
    position: absolute;
    left: 0;
    top: 0.15em;
    font-family: var(--font-mono);
    font-size: 0.76em;
    color: var(--text-3);
  }

  /* 待办清单：去圆点，复选框着 ink */
  :deep(li.task-list-item) { padding-left: 0; }
  :deep(li.task-list-item::before) { display: none; }
  :deep(.task-list-item input) { margin-right: 10px; accent-color: var(--ink); }

  :deep(table) {
    width: 100%;
    margin: 6px 0 24px;
    border-collapse: collapse;
    font-family: var(--font-sans);
    font-size: 14px;
    line-height: 1.7;
  }

  :deep(th),
  :deep(td) {
    padding: 10px 14px 10px 0;
    text-align: left;
    box-shadow: inset 0 -0.5px 0 var(--line-2);
  }

  :deep(th) {
    font-size: 12.5px;
    font-weight: 500;
    letter-spacing: 0.04em;
    color: var(--text-3);
  }

  :deep(img) {
    display: block;
    max-width: 100%;
    margin: 8px auto;
    border-radius: var(--r-md);
  }

  /* 拼图：同段落多张图并排成宫格 */
  :deep(p:has(img + img)) {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 6px;

    img {
      width: 100%;
      height: 100%;
      margin: 0;
      aspect-ratio: 4 / 3;
      object-fit: cover;
      border-radius: var(--r-sm);
    }
  }
}

/* ---------- 文末 ---------- */
.end {
  margin-top: 36px;
  padding-top: 22px;
  box-shadow: inset 0 0.5px 0 var(--line-2);
}

.row1 {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.tags { display: flex; flex-wrap: wrap; gap: 8px; }

/* 需要点击的标签：极轻描边胶囊 */
.tag-o {
  display: inline-flex;
  align-items: center;
  padding: 5px 12px;
  border-radius: var(--r-pill);
  font-size: 13px;
  color: var(--text-2);
  box-shadow: inset 0 0 0 1px var(--line-2);
  transition: color var(--dur-fast), box-shadow var(--dur-fast), background-color var(--dur-fast),
    transform var(--dur-fast) var(--ease-spring);

  &::before {
    content: '#';
    margin-right: 3px;
    font-family: var(--font-mono);
    font-size: 0.9em;
    color: var(--text-3);
  }

  &:hover { color: var(--text); background: var(--fill); box-shadow: inset 0 0 0 1px var(--line-2); }
  &:active { transform: scale(0.96); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

.btn-secondary {
  display: inline-flex;
  flex: none;
  align-items: center;
  gap: 7px;
  height: 38px;
  padding: 0 16px;
  border: 0;
  border-radius: var(--r-pill);
  background: var(--fill-2);
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
  box-shadow: inset 0 0 0 0.5px var(--line);
  transition: background-color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  &:hover { background: var(--fill-3); }
  &:active { transform: scale(0.97); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

:root[data-mode='light'] .btn-secondary {
  background: #fff;
  box-shadow: 0 0 0 0.5px var(--line-2), 0 1px 2px rgb(16 24 40 / 0.06);

  &:hover { background: color-mix(in oklab, var(--bg) 50%, white); }
}

.pn {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--card-gap);
  margin-top: 22px;

  a {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px;
    border-radius: var(--card-r);
    background: var(--card-bg);
    box-shadow: var(--card-shadow);
    transition: transform var(--dur) var(--ease-out), box-shadow var(--dur) var(--ease-out);

    &:hover { transform: translateY(var(--card-rise)); box-shadow: var(--card-shadow-hover); }
    &:focus-visible { outline: none; box-shadow: var(--card-shadow), var(--focus); }
  }

  .nx { flex-direction: row-reverse; text-align: right; }
  .nx small { justify-content: flex-end; }

  .th {
    position: relative;
    flex: none;
    width: 76px;
    height: 76px;
    overflow: hidden;
    isolation: isolate;
    border-radius: var(--r-md);
    background: #040914;
  }

  .tx { min-width: 0; }

  small {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 13px;
    color: var(--text-3);

    :deep(.ci) { transition: transform var(--dur) var(--ease-spring); }
  }

  b {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    margin-top: 4px;
    font-family: var(--font-serif);
    font-size: 19px;
    font-weight: 700;
    line-height: 1.45;
    color: var(--text);
  }

  .pv:hover small :deep(.ci) { transform: translateX(-3px); }
  .nx:hover small :deep(.ci) { transform: translateX(3px); }
}

/* 简洁风格：上一篇 / 下一篇去卡面，顶部一条发丝线，悬停只让缩略图微放大 */
:root[data-style='clean'] .pn a {
  padding: 18px 0 0;
  border-radius: 0;
  box-shadow: inset 0 1px 0 var(--line);

  &:hover { box-shadow: inset 0 1px 0 var(--line-2); }
  &:hover b { color: var(--ink); }
  &:focus-visible { box-shadow: var(--focus); }
}

:root[data-mode='dark']:not([data-style='clean']) .pn a {
  background: linear-gradient(180deg, color-mix(in oklab, var(--surface) 90%, white), var(--surface) 70%);
}

/* ---------- 窄屏阅读浮条 ---------- */
.dock {
  position: fixed;
  left: 50%;
  bottom: 24px;
  z-index: 50;
  display: none;
  align-items: center;
  gap: 4px;
  height: 48px;
  padding: 0 6px;
  border-radius: var(--r-pill);
  background: color-mix(in oklab, var(--bg) 78%, transparent);
  backdrop-filter: blur(22px) saturate(180%);
  -webkit-backdrop-filter: blur(22px) saturate(180%);
  box-shadow: inset 0 0 0 0.5px var(--line), var(--shadow-pop);
  font-size: 13px;
  color: var(--text-2);
  transform: translateX(-50%);

  .sep { width: 0.5px; height: 20px; background: var(--line-2); }
}

:root[data-mode='light'] .dock { background: rgb(255 255 255 / 0.78); }

.dk {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 36px;
  padding: 0 12px;
  border: 0;
  border-radius: var(--r-pill);
  background: none;
  color: inherit;
  font-size: 13px;
  transition: background-color var(--dur-fast), color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  b { font-family: var(--font-mono); font-weight: 500; color: var(--text); }

  &:hover { background: var(--fill-2); color: var(--text); }
  &:active { transform: scale(0.97); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
  &.on { background: var(--lift); color: var(--lift-fg); box-shadow: var(--lift-shadow); }

  .ring {
    width: 22px;
    height: 22px;
    transform: rotate(-90deg);

    circle { fill: none; stroke-width: 2.6; }
    .bgc { stroke: var(--fill-3); }
    .fgc { stroke: var(--ink); stroke-linecap: round; transition: stroke-dashoffset 0.15s linear; }
  }
}

.dpop {
  position: absolute;
  left: 50%;
  bottom: calc(100% + 10px);
  width: 300px;
  max-height: 50vh;
  padding: 8px;
  overflow-y: auto;
  border-radius: var(--r-lg);
  background: color-mix(in oklab, var(--surface) 94%, transparent);
  backdrop-filter: blur(24px) saturate(170%);
  -webkit-backdrop-filter: blur(24px) saturate(170%);
  box-shadow: var(--shadow-pop);
  transform: translateX(-50%);
  transform-origin: bottom center;

  button {
    display: block;
    width: 100%;
    padding: 8px 12px;
    border: 0;
    border-radius: var(--r-sm);
    background: none;
    text-align: left;
    font-size: 13.5px;
    line-height: 1.5;
    color: var(--text-2);
    transition: background-color var(--dur-fast), color var(--dur-fast);

    &.l3 { padding-left: 26px; font-size: 13px; }
    &:hover { background: var(--fill); color: var(--text); }
    &.on { background: var(--lift); color: var(--lift-fg); box-shadow: var(--lift-shadow); font-weight: 500; }
    &:focus-visible { outline: none; box-shadow: var(--focus); }
  }
}

:root[data-mode='light'] .dpop { background: rgb(255 255 255 / 0.94); }

.dpop-enter-active { transition: opacity 0.2s var(--ease-out), transform 0.35s var(--ease-spring); }
.dpop-leave-active { transition: opacity 0.16s var(--ease-out), transform 0.2s var(--ease-out); }

.dpop-enter-from,
.dpop-leave-to {
  opacity: 0;
  transform: translate(-50%, 8px) scale(0.97);
}

/* ---------- 未找到 ---------- */
.nf {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  padding: 160px 0 40px;
  color: var(--text-3);
}

/* ---------- 骨架 ---------- */
.skel {
  .band { border-radius: var(--r-xl); }
  .by { box-shadow: none; }
}

.sk-av { width: 40px; height: 40px; border-radius: 50%; }
.sk-prose { padding-top: 34px; --lh: 1.95em; font-size: 17.5px; }
.sk-prose .gap { visibility: hidden; }
.sk-toc { margin-top: 28px; height: 360px; border-radius: var(--card-r); background: var(--card-bg); box-shadow: var(--card-shadow); padding: var(--card-pad); font-size: 13px; }

/* ---------- 响应式 ---------- */
@media (max-width: 1300px) {
  .dt { --gap: 40px; --toc: 280px; }
}

/* 中宽（如 1024）：保留右栏阅读卡，正文栏收窄 */
@media (max-width: 1179px) {
  .dt {
    --toc: 272px;
    --gap: 32px;
    --outer: calc(100% - 64px);
  }
}

@media (max-width: 959px) {
  .dt {
    --toc: 0px;
    --gap: 0px;
  }

  .grid {
    grid-template-columns: minmax(0, 1fr);
    width: min(700px, calc(100% - 64px));
  }

  .band { width: min(860px, calc(100% - 64px)); height: 260px; }
  .side, .sk-toc { display: none; }
  .dock { display: flex; }
}

@media (max-width: 800px) {
  h1 { font-size: 36px; }
  .pn { grid-template-columns: 1fr; }
}
</style>
