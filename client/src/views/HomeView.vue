<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import HeroCarousel, { type HeroItem } from '../components/home/HeroCarousel.vue';
import { COVER_KINDS } from '../components/home/hero/CoverArt.vue';
import { api, type Note, type Post } from '../api';
import { useLoadingStore } from '../stores/loading';
import { useConfigStore } from '../stores/config';
import LatestPosts from '../components/home/LatestPosts.vue';
import ThoughtsStrip from '../components/home/ThoughtsStrip.vue';
import AboutMeCard from '../components/home/AboutMeCard.vue';
import GithubStatusCard from '../components/home/GithubStatusCard.vue';
import { hashOf, monthDay, type GhStatus } from '../components/home/format';

/**
 * 首页：Hero（高度随内容，约 min(78vh, 720px)）→ 最新文章（首篇大卡 + 缩略列表）
 * → 随想横向预览 → 关于小卡 + GitHub。
 * Hero 数据按住路由揭幕；其余板块各自加载，期间渲染同形骨架，到达后 rise-stagger 入场。
 */
const { t } = useI18n();
const config = useConfigStore();

const heroItems = ref<HeroItem[]>([]);
const heroInterval = ref(3000);
const heroLoading = ref(true);

const latest = ref<Post[]>([]);
const postTotal = ref(0);
const postsLoading = ref(true);

const notes = ref<Note[]>([]);
const noteTotal = ref(0);
const notesLoading = ref(true);

const gh = ref<GhStatus | null>(null);
const ghLoading = ref(true);

const commits = computed(() => gh.value?.stats.commits ?? config.cfg.github.stats.commits);

/** 无封面：按 slug 稳定挑一幅 CSS 光影封面（css:<kind>） */
function coversOf(post: Post): string[] {
  if (post.covers.length) return post.covers.slice(0, 3);
  return [`css:${COVER_KINDS[hashOf(post.slug || post.title) % COVER_KINDS.length]}`];
}

async function loadHero(): Promise<void> {
  // 数据就绪前按住路由揭幕，避免揭开后轮播才闪现
  const release = useLoadingStore().holdRoute();
  try {
    const feed = await api.hero();
    heroInterval.value = feed.intervalMs;
    heroItems.value = feed.items.map((post) => ({
      title: post.title,
      excerpt: post.excerpt,
      covers: coversOf(post),
      tag: post.tags[0] ?? t('home.posts'),
      date: t('home.md', monthDay(post.createdAt)),
      slug: post.slug,
    }));
  } catch {
    heroItems.value = [];
  } finally {
    heroLoading.value = false;
    release();
  }
}

async function loadPosts(): Promise<void> {
  try {
    const list = await api.posts({ pageSize: 4 });
    latest.value = list.items;
    postTotal.value = list.total;
  } catch {
    latest.value = [];
  } finally {
    postsLoading.value = false;
  }
}

async function loadNotes(): Promise<void> {
  try {
    const list = await api.notes({ pageSize: 6 });
    notes.value = list.items;
    noteTotal.value = list.total;
  } catch {
    notes.value = [];
  } finally {
    notesLoading.value = false;
  }
}

async function loadGithub(): Promise<void> {
  try {
    gh.value = await api.githubStatus<GhStatus>();
  } catch {
    gh.value = null;
  } finally {
    ghLoading.value = false;
  }
}

onMounted(() => {
  void loadHero();
  void loadPosts();
  void loadNotes();
  void loadGithub();
});
</script>

<template>
  <main class="page">
    <!-- Hero：无外框，舞台高度随内容；数据未到时渲染同形骨架 -->
    <div class="hero-wrap" :class="{ empty: !heroLoading && !heroItems.length }">
      <HeroCarousel v-if="heroItems.length" :items="heroItems" :photo-ms="heroInterval" />
      <div v-else-if="heroLoading" class="hero-sk" aria-hidden="true">
        <div class="txt">
          <span class="sk sk-line" style="width: 120px" />
          <span class="sk sk-line title" style="width: 86%" />
          <span class="sk sk-line title" style="width: 58%" />
          <span class="sk sk-line" style="width: 92%; margin-top: 12px" />
          <span class="sk sk-line" style="width: 70%" />
          <div class="cta"><span class="sk" /><span class="sk" /></div>
        </div>
        <div class="sk deck" />
      </div>
    </div>

    <LatestPosts
      v-if="postsLoading || latest.length"
      class="block"
      :posts="latest"
      :total="postTotal"
      :loading="postsLoading"
    />

    <ThoughtsStrip
      v-if="notesLoading || notes.length"
      class="block wide"
      :notes="notes"
      :total="noteTotal"
      :loading="notesLoading"
    />

    <div v-reveal class="block me rise-stagger">
      <AboutMeCard :posts="postTotal" :notes="noteTotal" :commits="commits" :loading="postsLoading || notesLoading" />
      <GithubStatusCard :data="gh" :loading="ghLoading" />
    </div>
  </main>
</template>

<style scoped lang="scss">
.page {
  max-width: 1248px;
  margin: 0 auto;
  padding: 64px 24px 80px;
}

/* ===== Hero 舞台：高度随内容（约 min(78vh, 720px)），内容垂直居中，无外框 ===== */
.hero-wrap {
  height: min(78vh, 720px);
  min-height: 540px;
  display: grid;
  align-items: center;

  &.empty {
    height: auto;
    min-height: 0;
  }
}

.hero-sk {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.05fr);
  gap: 56px;
  align-items: center;

  .txt {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .sk-line {
    display: block;
    height: 12px;
  }

  .title { height: 46px; }

  .cta {
    display: flex;
    gap: 10px;
    margin-top: 18px;

    span {
      width: 128px;
      height: 46px;
      border-radius: var(--r-pill);
    }

    span + span { width: 104px; }
  }

  .deck {
    justify-self: center;
    width: min(100%, clamp(430px, 36vw, 580px));
    aspect-ratio: 4 / 3;
    border-radius: var(--r-lg);
  }
}

.block { margin-top: 96px; }
.hero-wrap + .block { margin-top: 24px; }

.me {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-top: 72px;
}

@media (max-width: 900px) {
  .hero-wrap {
    height: auto;
    min-height: 0;
    padding: 24px 0 8px;
  }

  .hero-sk {
    grid-template-columns: 1fr;
    gap: 26px;

    .deck { width: min(100% - 92px, 300px); }
  }

  .block { margin-top: 72px; }
}

@media (max-width: 768px) {
  .page { padding-top: 72px; }
}
</style>
