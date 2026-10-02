<script setup lang="ts">
import { useIdentity } from '../../about/useIdentity';
/* 正文宋体常规字重（全站只预载了 700；仅移动详情分包内按需加载，unicode-range 切片） */
import '@fontsource/noto-serif-sc/400.css';
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { api, type Post } from '../../api';
import { renderWithToc, type TocItem } from '../../utils/markdown';
import CoverArt, { isArtUrl } from '../../components/common/CoverArt.vue';
import MediaViewer from '../../components/mobile/MediaViewer.vue';
import { defaultCover } from '../../utils/defaultCovers';
import MIcon from '../../components/mobile/MIcon.vue';
import BottomSheet from '../../components/mobile/BottomSheet.vue';
import EngageBar from '../../components/engage/EngageBar.vue';
import CommentSection from '../../components/engage/CommentSection.vue';
import { copyText, monthDay, readMinutes, toast } from '../../components/mobile/shell';
import IdentityName from '../../components/common/IdentityName.vue';

/**
 * 移动端文章详情（沉浸阅读）：底栏隐藏；顶部返回条随滚动变毛玻璃并浮现标题 + 阅读进度线；
 * 封面视差；底部浮条 = 进度环 + 剩余时间 + 目录按钮（下滑隐藏、上滑浮现）；目录为可拖拽 BottomSheet。
 */
const props = defineProps<{ slug: string }>();
const emit = defineEmits<{ back: [] }>();
const { t } = useI18n();
const router = useRouter();

const post = ref<Post | null>(null);
const next = ref<Post | null>(null);
const loading = ref(true);
const notFound = ref(false);
const scroller = ref<HTMLElement | null>(null);
const proseEl = ref<HTMLElement | null>(null);

const BIG_KEY = 'myself.m.bigType';
const bigType = ref(localStorage.getItem(BIG_KEY) === '1');

/** 正文首个一级标题与文章标题重复时去掉（详情页已单独排版标题） */
const rendered = computed(() => {
  const src = post.value?.contentMd ?? '';
  const title = post.value?.title.trim() ?? '';
  const body = src.replace(/^\s*#\s+(.+)\n+/, (m, h: string) => (h.trim() === title ? '' : m));
  return body ? renderWithToc(body, `post-${post.value?.id}`) : { html: '', toc: [] as TocItem[] };
});
const minutes = computed(() => readMinutes(post.value?.contentMd ?? ''));

onMounted(async () => {
  try {
    const [p, list] = await Promise.all([api.post(props.slug), api.posts({ pageSize: 30 }).catch(() => null)]);
    post.value = p;
    if (list) {
      const idx = list.items.findIndex((x) => x.slug === p.slug);
      const cand = idx >= 0 ? list.items[idx + 1] ?? list.items[0] : list.items[0];
      next.value = cand && cand.slug !== p.slug ? cand : null;
    }
  } catch {
    notFound.value = true;
  } finally {
    loading.value = false;
  }
  await nextTick();
  onScroll();
});

/* ---------- 滚动：视差 / 顶栏实化 / 进度 / 浮条显隐 ---------- */
const y = ref(0);
const rp = ref(0);
const dockHide = ref(false);
let lastY = 0;

const solid = computed(() => y.value > 250);
const heroY = computed(() => Math.min(Math.max(y.value, -200), 420));
const left = computed(() =>
  rp.value > 0.98 ? t('mobile.article.done') : t('mobile.article.left', { n: Math.max(1, Math.ceil(minutes.value * (1 - rp.value))) }),
);

function onScroll(): void {
  const el = scroller.value;
  if (!el) return;
  const top = el.scrollTop;
  const max = el.scrollHeight - el.clientHeight;
  y.value = top;
  rp.value = Math.min(1, Math.max(0, top / Math.max(1, max - 120)));
  const dy = top - lastY;
  lastY = top;
  if (dy > 6 && top > 400) dockHide.value = true;
  else if (dy < -6 || top < 400 || rp.value > 0.97) dockHide.value = false;
}

/* ---------- 目录 ---------- */
const tocOpen = ref(false);
const curIdx = ref(0);

function headingTop(id: string): number {
  const sc = scroller.value;
  const el = proseEl.value?.querySelector<HTMLElement>(`#${CSS.escape(id)}`);
  if (!sc || !el) return 0;
  return el.getBoundingClientRect().top - sc.getBoundingClientRect().top + sc.scrollTop;
}

function openToc(): void {
  const probe = (scroller.value?.scrollTop ?? 0) + 140;
  let cur = 0;
  rendered.value.toc.forEach((h, k) => { if (headingTop(h.id) <= probe) cur = k; });
  curIdx.value = cur;
  tocOpen.value = true;
}

function jump(item: TocItem): void {
  tocOpen.value = false;
  const top = headingTop(item.id) - 76;
  window.setTimeout(() => scroller.value?.scrollTo({ top, behavior: 'smooth' }), 120);
}

const tocMeta = computed(() => t('mobile.article.tocMeta', { n: minutes.value, s: rendered.value.toc.length }));
const perSection = computed(() => Math.max(1, Math.round(minutes.value / Math.max(1, rendered.value.toc.length))));

/* ---------- 动作 ---------- */
function toggleType(): void {
  bigType.value = !bigType.value;
  localStorage.setItem(BIG_KEY, bigType.value ? '1' : '0');
  toast(bigType.value ? t('mobile.article.bigType') : t('mobile.article.normalType'));
}

/** 互动栏的评论按钮：平滑滚到评论区（评论关闭时评论区不渲染，什么也不做） */
function toComments(): void {
  document.getElementById('comments')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

async function share(): Promise<void> {
  const ok = await copyText(window.location.href);
  toast(ok ? t('mobile.linkCopied') : t('mobile.copyFailed'));
}

function toTop(): void {
  scroller.value?.scrollTo({ top: 0, behavior: 'smooth' });
}

function date(s: string): string {
  const { y: yy, m, d } = monthDay(s);
  return `${yy}.${String(m).padStart(2, '0')}.${String(d).padStart(2, '0')}`;
}

const coverButton = ref<HTMLElement | null>(null);
const coverOpen = ref(false);
const coverIndex = ref(0);
const coverItems = computed(() => {
  const p = post.value;
  if (!p) return [];
  return (p.covers.length ? p.covers : ['']).map((src) => ({
    src: isArtUrl(src) ? defaultCover(p.slug) : src,
    caption: p.title,
  }));
});
const coverOrigin = (index: number) => index === 0 ? coverButton.value : null;
function openCover(): void {
  coverIndex.value = 0;
  coverOpen.value = true;
}

/* 正文内站内链接走路由，不整页刷新 */
function onProseClick(e: MouseEvent): void {
  const a = (e.target as HTMLElement).closest('a');
  if (!a) return;
  const href = a.getAttribute('href') ?? '';
  if (href.startsWith('#')) {
    e.preventDefault();
    const top = headingTop(decodeURIComponent(href.slice(1))) - 76;
    scroller.value?.scrollTo({ top, behavior: 'smooth' });
  } else if (/^\/(?!\/)/.test(href) && !href.startsWith('/uploads') && !href.startsWith('/feed')) {
    e.preventDefault();
    void router.push(href);
  }
}

onBeforeUnmount(() => { tocOpen.value = false; coverOpen.value = false; });

const me = useIdentity();
const byline = computed(() => (post.value?.author
  ? { avatar: post.value.author.avatar || '/favicon-256.png', name: post.value.author.name, alias: '', sign: '' }
  : { avatar: me.avatar.value, name: me.name.value, alias: me.alias.value, sign: me.sign.value }));
</script>

<template>
  <div class="art" :class="{ solid, 'dock-hide': dockHide, big: bigType }" :style="{ '--rp': rp }">
    <div ref="scroller" class="dt-scroll" @scroll.passive="onScroll">
      <Transition name="m-swap" mode="out-in">
        <!-- 骨架：封面 + 标题 + 段落 -->
        <div v-if="loading" key="sk" class="sk">
          <div class="m-sk sk-hero" />
          <div class="sk-head">
            <span class="m-sk m-sk-line" style="width: 46%; height: 12px" />
            <span class="m-sk m-sk-line" style="width: 88%; height: 26px; margin-top: 16px" />
            <span class="m-sk m-sk-line" style="width: 60%; height: 26px; margin-top: 10px" />
            <span class="m-sk m-sk-line" style="width: 94%; margin-top: 20px" />
            <span class="m-sk m-sk-line" style="width: 76%; margin-top: 10px" />
            <div class="sk-by"><span class="m-sk sk-av" /><span class="m-sk m-sk-line" style="width: 120px" /></div>
            <span v-for="n in 6" :key="n" class="m-sk m-sk-line" :style="{ width: `${[96, 90, 98, 70, 92, 84][n - 1]}%`, marginTop: '16px' }" />
          </div>
        </div>

        <div v-else-if="post" key="ok">
          <header class="dt-hero">
            <button ref="coverButton" type="button" class="hero-art" :aria-label="t('a11y.viewImageOf', { label: post.title })" :style="{ transform: `translateY(${heroY * 0.45}px) scale(1.02)` }" @click="openCover">
              <CoverArt :src="post.covers[0]" :seed="post.slug" />
            </button>
            <div class="fade" />
          </header>
          <div class="dt-head">
            <div class="dt-meta m-in" style="--i: 0">
              <span v-if="post.tags[0]" class="m-mood">{{ post.tags[0] }}</span>
              <span>{{ date(post.createdAt) }} · {{ t('mobile.article.minutes', { n: minutes }) }}</span>
            </div>
            <h1 class="m-in" style="--i: 1">{{ post.title }}</h1>
            <p v-if="post.excerpt" class="post-excerpt lede m-in" style="--i: 2">{{ post.excerpt }}</p>
            <div class="by m-in" style="--i: 3">
              <span class="av"><img class="m-avatar" :src="byline.avatar" alt="" draggable="false" /></span>
              <div><b><IdentityName :name="byline.name" :alias="byline.alias" /></b><template v-if="byline.sign"><br /><small>{{ byline.sign }}</small></template></div>
            </div>
          </div>
          <!-- eslint-disable-next-line vue/no-v-html -->
          <article ref="proseEl" class="prose markdown-content m-in" style="--i: 4" @click="onProseClick" v-html="rendered.html" />
          <footer class="dt-end m-in" style="--i: 5">
            <div class="tags">
              <button
                v-for="tag in post.tags"
                :key="tag"
                class="m-tagp m-tap"
                @click="router.push({ path: '/articles', query: { tag } })"
              >
                {{ tag }}
              </button>
            </div>
            <button v-if="next" class="next m-tap" @click="router.push(`/articles/${next.slug}`)">
              <div class="nt"><small>{{ t('mobile.article.next') }}</small><b>{{ next.title }}</b></div>
              <div class="thumb"><CoverArt :src="next.covers[0]" :seed="next.slug" thumb /></div>
            </button>
          </footer>
          <div class="dt-engage m-in" style="--i: 6">
            <EngageBar target="post" :id="post.id" :link="`/articles/${post.slug}`" :text="post.title" big @comment="toComments" />
          </div>
          <div class="dt-comments">
            <CommentSection target="post" :comment-key="post.slug" :engage-id="post.id" />
          </div>
        </div>

        <div v-else key="nf" class="m-empty nf">{{ t('mobile.article.notFound') }}</div>
      </Transition>
    </div>

    <!-- 顶部返回条 -->
    <div class="dt-top" :class="{ plain: !post && !loading }">
      <div class="bg" />
      <div class="prog" />
      <button class="m-icbtn m-tap" :aria-label="t('mobile.back')" @click="emit('back')"><MIcon name="back" /></button>
      <span class="ttl">{{ post?.title }}</span>
      <button class="m-icbtn m-tap" :aria-label="t('mobile.article.fontSize')" @click="toggleType"><MIcon name="type" /></button>
      <button class="m-icbtn m-tap" :aria-label="t('mobile.share')" @click="share"><MIcon name="share" /></button>
    </div>

    <!-- 底部阅读浮条 -->
    <div v-if="post" class="dock m-glass">
      <button class="dk m-tap" @click="toTop">
        <svg class="ring" viewBox="0 0 24 24"><circle class="bg" cx="12" cy="12" r="9" /><circle class="fg" cx="12" cy="12" r="9" /></svg>
        <span class="pct">{{ Math.round(rp * 100) }}%</span>
        <small>{{ left }}</small>
      </button>
      <template v-if="rendered.toc.length">
        <span class="sep" />
        <button class="dk m-tap" @click="openToc"><MIcon name="list" class="s" />{{ t('mobile.article.toc') }}</button>
      </template>
    </div>

    <MediaViewer v-model:open="coverOpen" v-model:index="coverIndex" :items="coverItems" :origin="coverOrigin" />

    <BottomSheet v-model:open="tocOpen" :title="t('mobile.article.toc')" :meta="tocMeta">
      <div class="toc">
        <button
          v-for="(h, k) in rendered.toc"
          :key="h.id"
          class="m-tap"
          :class="{ cur: k === curIdx, done: k < curIdx, sub: h.level === 3 }"
          @click="jump(h)"
        >
          <i />{{ h.text }}
          <small>{{ k < curIdx ? t('mobile.article.read') : k === curIdx ? t('mobile.article.reading') : `${perSection} min` }}</small>
        </button>
      </div>
    </BottomSheet>
  </div>
</template>

<style scoped lang="scss">
.art {
  position: absolute;
  inset: 0;
  background: var(--bg);
}

.dt-scroll {
  position: absolute;
  inset: 0;
  overflow-y: auto;
  overflow-x: hidden;
  overscroll-behavior-y: contain;
  scrollbar-width: none;
  touch-action: pan-y pinch-zoom;

  &::-webkit-scrollbar { display: none; }
}

/* 封面视差 */
.dt-hero {
  position: relative;
  /* 高密度：封面收一档（原 min(400px, 50vh)） */
  height: min(340px, 42vh);
  overflow: hidden;
}

.hero-art {
  position: absolute;
  inset: 0;
  will-change: transform;
  width: 100%;
  height: 100%;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: zoom-in;

  &:focus-visible { outline: 2px solid var(--ink); outline-offset: -3px; }
}

.fade {
  position: absolute;
  inset: auto 0 0 0;
  height: 220px;
  background: linear-gradient(180deg, transparent, color-mix(in oklab, var(--bg) 70%, transparent) 55%, var(--bg));
  pointer-events: none;
}

.dt-head {
  position: relative;
  margin-top: -110px;
  padding: 0 18px;

  h1 {
    margin-top: 12px;
    font-family: var(--font-serif);
    font-size: 30px;
    line-height: 1.34;
    font-weight: 700;
    letter-spacing: 0.01em;
    text-wrap: balance;
  }
}

.dt-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--text-2);

  .m-mood {
    font-size: inherit;
    font-weight: 500;
    color: var(--text);
  }
}

.lede {
  margin-top: 12px;
  font-family: var(--font-serif);
  font-size: 16px;
  line-height: 1.8;
  color: var(--text-2);
}

.by {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 14px;
  padding: 12px 0 16px;
  border-bottom: 0.5px solid var(--line);
  font-size: 14px;

  .av {
    width: 32px;
    height: 32px;
    flex: none;
  }

  small {
    color: var(--text-3);
    font-size: 13px;
  }
}

/* 正文排版 */
.prose { padding: 26px 22px 0; :deep(img) { cursor: zoom-in; } }

/* 底部留白让给最后一块（评论区），避开阅读浮条 */
.dt-end {
  padding: 18px 22px 0;

  .tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
}

.dt-engage { display: flex; justify-content: center; padding: 26px 22px 0; }
.dt-comments { padding: 0 22px calc(var(--m-safe-b) + 120px); }

.next {
  display: flex;
  align-items: center;
  gap: 14px;
  width: 100%;
  margin-top: 26px;
  padding: 14px;
  border-radius: var(--r-xl);
  text-align: left;
  background: var(--card-bg);
  box-shadow: var(--card-shadow);

  /* 简洁风格：去卡面，上方一条发丝线 */
  :root[data-style='clean'] & {
    padding: 16px 0 0;
    border-radius: 0;
    box-shadow: inset 0 1px 0 var(--line);
  }

  .nt { flex: 1; min-width: 0; }

  small {
    font-size: 12px;
    color: var(--text-3);
  }

  b {
    display: block;
    font-family: var(--font-serif);
    font-size: 16.5px;
    margin-top: 3px;
  }

  .thumb {
    position: relative;
    width: 62px;
    height: 62px;
    border-radius: var(--r-lg);
    overflow: hidden;
    flex: none;
  }
}

.nf { padding-top: calc(var(--m-nav-h) + 80px); }

/* 顶部返回条 */
.dt-top {
  position: absolute;
  z-index: 5;
  top: 0;
  left: 0;
  right: 0;
  height: var(--m-nav-h);
  padding: calc(var(--m-safe-t) + 6px) 14px 0;
  display: flex;
  align-items: center;
  gap: 8px;

  .bg {
    position: absolute;
    inset: 0;
    background: var(--m-glass-2);
    backdrop-filter: blur(24px) saturate(180%);
    -webkit-backdrop-filter: blur(24px) saturate(180%);
    box-shadow: 0 0.5px 0 var(--line-2);
    opacity: 0;
    transition: opacity var(--dur) var(--ease-out);
  }

  .m-icbtn {
    position: relative;
    background: rgba(8, 14, 28, 0.38);
    color: #fff;
    backdrop-filter: blur(14px);
    -webkit-backdrop-filter: blur(14px);
    box-shadow: inset 0 0 0 0.5px rgba(255, 255, 255, 0.18);
    transition: background var(--dur), color var(--dur), transform var(--dur-fast) var(--ease-spring);
  }

  .ttl {
    position: relative;
    flex: 1;
    text-align: center;
    font-size: 15px;
    font-weight: 600;
    opacity: 0;
    transform: translateY(6px);
    transition: all var(--dur) var(--ease-out);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .prog {
    position: absolute;
    left: 0;
    bottom: 0;
    height: 2px;
    width: 100%;
    transform: scaleX(var(--rp, 0));
    transform-origin: left center;
    background: var(--ink);
    opacity: 0;
    transition: opacity var(--dur);
  }
}

.solid .dt-top,
.dt-top.plain {
  .bg { opacity: 1; }

  .ttl {
    opacity: 1;
    transform: none;
  }

  .m-icbtn {
    background: var(--fill-2);
    color: var(--text);
    box-shadow: none;
  }
}

.solid .dt-top .prog { opacity: 1; }

/* 底部阅读浮条 */
.dock {
  position: absolute;
  z-index: 6;
  left: 50%;
  bottom: max(20px, calc(var(--m-safe-b) + 4px));
  display: flex;
  align-items: center;
  height: 52px;
  padding: 0 6px;
  border-radius: 999px;
  transform: translateX(-50%);
  transition: transform 0.5s var(--ease-spring), opacity var(--dur);
  white-space: nowrap;
  box-shadow: inset 0 0 0 0.5px var(--m-glass-line), inset 0 1px 0 var(--m-glass-hi), var(--m-bar-shadow);
  animation: dock-in 0.6s var(--ease-spring) 0.35s backwards;

  .dock-hide & { transform: translateX(-50%) translateY(110px); }
}

@keyframes dock-in {
  from { transform: translateX(-50%) translateY(90px); }
}

.dk {
  display: flex;
  align-items: center;
  gap: 9px;
  height: 40px;
  padding: 0 14px;
  border-radius: 999px;
  font-size: 14px;

  &:active { background: var(--fill-3); }

  small {
    font-size: 12.5px;
    color: var(--text-3);
  }
}

.ring {
  width: 24px;
  height: 24px;
  transform: rotate(-90deg);

  circle {
    fill: none;
    stroke-width: 2.6;
  }

  .bg { stroke: var(--fill-3); }

  .fg {
    stroke: var(--ink);
    stroke-linecap: round;
    stroke-dasharray: 56.55;
    stroke-dashoffset: calc(56.55 * (1 - var(--rp, 0)));
  }
}

.pct {
  font-family: var(--m-font-mono);
  font-size: 13px;
  font-weight: 500;
  min-width: 34px;
}

.sep {
  width: 0.5px;
  height: 22px;
  background: var(--line-2);
}

/* 目录 sheet */
.toc {
  position: relative;
  padding: 0 12px;

  button {
    position: relative;
    width: 100%;
    display: flex;
    align-items: center;
    gap: 14px;
    min-height: 52px;
    padding: 0 14px;
    border-radius: var(--r-md);
    text-align: left;
    font-size: 15.5px;
    color: var(--text-2);

    &.sub {
      padding-left: 34px;
      font-size: 14.5px;
    }

    i {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: var(--line-2);
      flex: none;
      transition: all var(--dur) var(--ease-spring);
    }

    small {
      margin-left: auto;
      padding-left: 12px;
      font-size: 12px;
      color: var(--text-3);
      font-family: var(--m-font-mono);
      font-weight: 400;
      flex: none;
    }

    &.done i { background: var(--text-3); }

    &.cur {
      background: var(--lift);
      box-shadow: var(--lift-shadow);
      color: var(--lift-fg);
      font-weight: 600;

      i { background: var(--ink); }
    }
  }
}

/* 骨架 */
.sk-hero {
  height: min(340px, 42vh);
  border-radius: 0;
}

.sk-head {
  padding: 24px 22px 0;

  .m-sk-line { display: block; }
}

.sk-by {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 24px 0 12px;
}

.sk-av {
  width: 32px;
  height: 32px;
  border-radius: 50%;
}
</style>
