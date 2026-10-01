<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { api, thumbOf, type Note, type Post } from '../../api';
import { useConfigStore } from '../../stores/config';
import LargeTitlePage from '../../components/mobile/LargeTitlePage.vue';
import CoverCarousel from '../../components/mobile/CoverCarousel.vue';
import CoverArt from '../../components/common/CoverArt.vue';
import MIcon from '../../components/mobile/MIcon.vue';
import AboutMeCard from '../../components/home/AboutMeCard.vue';
import GithubStatusCard from '../../components/home/GithubStatusCard.vue';
import type { GhStatus } from '../../components/home/format';
import { copyText, monthDay, shell, toast } from '../../components/mobile/shell';

/**
 * 移动端首页：封面卡横滑轮播 + 最新文章 + 随想横滑预览 + 底部「关于我 + GitHub」简介卡
 * （复用桌面首页的 AboutMeCard / GithubStatusCard，纵向堆叠，窄屏适配只在外层做）。
 */
const { t } = useI18n();
const router = useRouter();
const config = useConfigStore();

const hero = ref<Post[]>([]);
const latest = ref<Post[]>([]);
const notes = ref<Note[]>([]);
const loading = ref(true);
const failed = ref(false);
const postTotal = ref(0);
const noteTotal = ref(0);
const gh = ref<GhStatus | null>(null);
const ghLoading = ref(true);
const commits = computed(() => gh.value?.stats.commits ?? config.cfg.github.stats.commits);

const eyebrow = computed(() => {
  const now = new Date();
  const week = new Intl.DateTimeFormat('zh-CN', { weekday: 'long', timeZone: config.cfg.timezone }).format(now);
  return `${week} · ${now.getMonth() + 1} 月 ${now.getDate()} 日`;
});

async function load(): Promise<void> {
  failed.value = false;
  try {
    const [feed, list, feedNotes] = await Promise.all([
      api.hero(),
      api.posts({ pageSize: 4 }),
      api.notes({ pageSize: 6 }),
    ]);
    hero.value = feed.items.length ? feed.items : list.items;
    latest.value = list.items;
    notes.value = feedNotes.items;
    postTotal.value = list.total;
    noteTotal.value = feedNotes.total;
  } catch {
    failed.value = true;
  } finally {
    loading.value = false;
  }
}

/** GitHub 状态单独拉取（服务端缓存，可能较慢），不阻塞首屏内容 */
async function loadGithub(): Promise<void> {
  try {
    gh.value = await api.githubStatus<GhStatus>();
  } catch {
    gh.value = null;
  } finally {
    ghLoading.value = false;
  }
}

async function refresh(): Promise<void> {
  await Promise.all([load(), loadGithub()]);
}

onMounted(() => {
  void load();
  void loadGithub();
});

function plain(md: string): string {
  return md.replace(/!\[[^\]]*]\([^)]*\)/g, '').replace(/[#>*_`~[\]]/g, '').replace(/\(([^)]*)\)/g, '').trim();
}

function md(s: string): string {
  const { m, d } = monthDay(s);
  return `${m}月${d}日`;
}

function openPost(slug: string): void {
  void router.push(`/articles/${slug}`);
}

async function copyRss(): Promise<void> {
  const ok = await copyText(`${window.location.origin}/feed`);
  toast(ok ? t('mobile.rssCopied') : t('mobile.copyFailed'), ok ? '/feed' : '');
}

const carouselPaused = computed(() => shell.dp > 0.1 || shell.pushed || shell.tab !== 'home' || shell.searchOpen);
</script>

<template>
  <LargeTitlePage :title="config.cfg.site.title" :eyebrow="eyebrow" :refresh="refresh">
    <template #right>
      <button class="m-icbtn m-tap" :aria-label="t('mobile.rss')" @click="copyRss"><MIcon name="rss" /></button>
    </template>

    <Transition name="m-swap" mode="out-in">
      <!-- 骨架：与最终布局同几何 -->
      <div v-if="loading" key="sk" class="sk">
        <div class="m-sk sk-card" />
        <div class="sk-dots"><span class="m-sk" /><span class="m-sk" /><span class="m-sk" /></div>
        <div class="sec-h"><span class="m-sk m-sk-line" style="width: 96px; height: 20px" /></div>
        <div v-for="n in 3" :key="n" class="sk-row">
          <div class="rt">
            <span class="m-sk m-sk-line" style="width: 40%; height: 10px" />
            <span class="m-sk m-sk-line" style="width: 82%; height: 16px" />
            <span class="m-sk m-sk-line" style="width: 64%" />
          </div>
          <span class="m-sk thumb" />
        </div>
      </div>

      <div v-else key="ok">
        <div v-if="failed && !hero.length" class="m-empty">{{ t('mobile.loadFailed') }}</div>
        <CoverCarousel v-if="hero.length" class="m-in" :posts="hero" :paused="carouselPaused" @open="openPost" />

        <section v-if="latest.length" class="sec">
          <div class="sec-h m-in" style="--i: 2">
            <h2>{{ t('mobile.home.latest') }}</h2>
            <button class="more m-tap" @click="router.push('/articles')">{{ t('mobile.home.all') }}<MIcon name="chev" class="xs" /></button>
          </div>
          <div class="rows">
            <button
              v-for="(p, i) in latest"
              :key="p.id"
              class="m-arow m-in"
              :style="{ '--i': i + 3 }"
              @click="openPost(p.slug)"
            >
              <div class="rt">
                <div class="meta"><em>{{ p.tags[0] }}</em> · {{ md(p.createdAt) }}</div>
                <b>{{ p.title }}</b>
                <span class="post-excerpt ex">{{ p.excerpt }}</span>
              </div>
              <div class="thumb"><CoverArt :src="p.covers[0]" :seed="p.slug" thumb /></div>
            </button>
          </div>
        </section>

        <section v-if="notes.length" class="sec">
          <div class="sec-h m-in" style="--i: 7">
            <h2>{{ t('mobile.home.thoughts') }}</h2>
            <button class="more m-tap" @click="router.push('/thoughts')">{{ t('mobile.home.all') }}<MIcon name="chev" class="xs" /></button>
          </div>
          <div class="tp m-hscroll">
            <button
              v-for="(n, i) in notes"
              :key="n.id"
              class="tcard m-tap m-in"
              :style="{ '--i': i + 8 }"
              @click="router.push('/thoughts')"
            >
              <p>{{ plain(n.contentMd) }}</p>
              <div v-if="n.images.length" class="strip">
                <img v-for="src in n.images.slice(0, 3)" :key="src" :src="thumbOf(src)" alt="" loading="lazy" draggable="false" />
                <span v-if="n.images.length > 3" class="more">+{{ n.images.length - 3 }}</span>
              </div>
              <footer><em v-if="n.mood" class="m-mood">{{ n.mood }}</em><span>{{ md(n.createdAt) }}</span></footer>
            </button>
          </div>
        </section>

        <section class="sec me m-in" style="--i: 12">
          <AboutMeCard :posts="postTotal" :notes="noteTotal" :commits="commits" />
          <GithubStatusCard :data="gh" :loading="ghLoading" />
        </section>
      </div>
    </Transition>
  </LargeTitlePage>
</template>

<style scoped lang="scss">
/* 高密度：分区间距 28px、分区标题 24px 宋体、左右统一 16px 边距 */
.sec { margin-top: 28px; }
:root[data-style='clean'] .sec { margin-top: 40px; }

.sec-h {
  display: flex;
  align-items: baseline;
  padding: 0 16px;
  margin-bottom: 4px;

  h2 {
    font-family: var(--font-serif);
    font-size: 24px;
    font-weight: 700;
    line-height: 1.3;
  }

  .more {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: 2px;
    font-size: 14px;
    color: var(--ink);
  }
}

.rows { padding: 0; }

.tp {
  display: flex;
  gap: 10px;
  padding: 10px 16px 14px;
  scroll-snap-type: x mandatory;
  scroll-padding: 0 16px;
}

.tcard {
  scroll-snap-align: start;
  flex: none;
  width: min(280px, 76vw);
  padding: 16px;
  border-radius: var(--card-r);
  text-align: left;
  background: var(--card-bg);
  box-shadow: var(--card-shadow);
  display: flex;
  flex-direction: column;
  gap: 12px;

  p {
    font-size: 15px;
    line-height: 1.65;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
    min-height: 74px;
  }

  .strip {
    display: flex;
    gap: 4px;

    img,
    .more {
      width: 56px;
      height: 56px;
      border-radius: var(--r-sm);
      object-fit: cover;
      background: var(--fill-3);
    }

    .more {
      display: grid;
      place-items: center;
      font-size: 13px;
      color: var(--text-2);
      font-family: var(--font-mono);
    }
  }

  footer {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: var(--text-3);
    margin-top: auto;

    .m-mood { font-size: 13px; }
  }
}

/* 简洁风格：随想卡去卡面，卡间一条竖向发丝线 */
:root[data-style='clean'] {
  .tp { gap: 0; }

  .tcard {
    padding: 2px 16px 4px;
    border-radius: 0;
    box-shadow: inset 1px 0 0 var(--line);

    &:first-child { padding-left: 0; box-shadow: none; }
  }
}

/* 底部「关于我 + GitHub」：复用桌面卡片，纵向堆叠；窄屏适配只在外层用 :deep 调整 */
.me {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 0 16px;

  :deep(.me-card),
  :deep(.gh-card) {
    height: auto;
    padding: 20px;
    gap: 18px;
  }

  /* 简洁风格：两块之间一条发丝线，内容与页边对齐 */
  :root[data-style='clean'] & {
    gap: 28px;

    :deep(.me-card),
    :deep(.gh-card) { padding: 0; }

    :deep(.gh-card) { padding-top: 24px; border-radius: 0; box-shadow: inset 0 1px 0 var(--line); }
  }

  /* 身份行：头像跨两行；名字与社交按钮同一行，格言落到第二行占满剩余宽度，不再被挤成两行 */
  :deep(.me-card .head) { column-gap: 14px; row-gap: 2px; align-items: center; }
  :deep(.me-card .av) { grid-row: 1 / span 2; width: 64px; height: 64px; }
  :deep(.me-card .who) { display: contents; }
  :deep(.me-card .who b) { grid-column: 2; grid-row: 1; font-size: 24px; }
  :deep(.me-card .who q) { grid-column: 2 / -1; grid-row: 2; margin-top: 0; font-size: 14px; }
  :deep(.me-card .soc) { grid-column: 3; grid-row: 1; align-self: center; }
  :deep(.me-card .foot) { flex-wrap: wrap; }
  :deep(.me-card .btn-2nd) { margin-left: auto; }

  /* 热力图：窄屏只保留最近 22 周（从开头隐藏整列 31 x 7 格，周对齐不变），格子随宽度放大到约 11px */
  :deep(.gh-card .heatmap > .cell:nth-child(-n + 217)) { display: none; }
  :deep(.gh-card .activity li) { grid-template-columns: minmax(0, 1fr) auto; }
  :deep(.gh-card .activity .a-repo) { grid-column: 1 / -1; }
}

/* 骨架 */
.sk-card {
  margin: 4px 16px 0;
  height: min(calc(100vw - 32px), 52vh);
  border-radius: var(--r-xl);
}

.sk-dots {
  display: flex;
  justify-content: center;
  gap: 6px;
  margin: 14px 0 28px;

  span { width: 6px; height: 6px; border-radius: 999px; }
  span:first-child { width: 30px; }
}

.sk-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 16px;

  .rt {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .thumb {
    width: 76px;
    height: 76px;
    border-radius: var(--r-md);
    flex: none;
  }
}

.sk .sec-h { margin-bottom: 8px; }
</style>
