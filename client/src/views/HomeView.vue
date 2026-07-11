<script setup lang="ts">
import { onMounted, ref } from 'vue';
import HeroCarousel, { type HeroItem } from '../components/home/HeroCarousel.vue';
import { placeholderCover } from '../utils/placeholder';
import { api, type Post } from '../api';
import { useLoadingStore } from '../stores/loading';
import PostZigzagList from '../components/post/PostZigzagList.vue';
import GithubStatusCard from '../components/home/GithubStatusCard.vue';
import AboutMeCard from '../components/home/AboutMeCard.vue';

/* 轮播对接真实文章（规则由后台配置：最新 n 条 + 置顶策略） */
const heroItems = ref<HeroItem[]>([]);
const heroInterval = ref(3000);
const latest = ref<Post[]>([]);

const fallbackPalette = [
  ['#ff0032', '#7a0020'],
  ['#ffb300', '#7a5200'],
  ['#0078ff', '#00295c'],
  ['#00c853', '#00512a'],
];

function coversOf(post: Post, index: number): string[] {
  if (post.covers.length) return post.covers.slice(0, 3);
  const [from, to] = fallbackPalette[index % fallbackPalette.length];
  return [placeholderCover(from, to, post.tags[0] ?? 'Post')];
}

onMounted(async () => {
  // 数据就绪前按住路由揭幕，避免揭开后轮播才闪现
  const release = useLoadingStore().holdRoute();
  try {
    const [feed, list] = await Promise.all([
      api.hero(),
      api.posts({ pageSize: 4 }),
    ]);
    heroInterval.value = feed.intervalMs;
    heroItems.value = feed.items.map((post, i) => ({
      title: post.title,
      excerpt: post.excerpt,
      covers: coversOf(post, i),
      tag: post.tags[0] ?? '文章',
      slug: post.slug,
    }));
    latest.value = list.items;
  } finally {
    release();
  }
});
</script>

<template>
  <main class="page">
    <!-- 常驻背景壳：数据未就绪也有占位，杜绝揭幕后底色闪现 -->
    <div class="hero-shell">
      <HeroCarousel v-if="heroItems.length" :items="heroItems" :photo-ms="heroInterval" />
    </div>

    <!-- 最新四条：交错图文，无边框平铺，悬停直角框 -->
    <section class="latest">
      <h2 v-reveal class="section-title">最新文章</h2>
      <PostZigzagList :posts="latest" />
    </section>

    <!-- 扩展板块：GitHub Status + 关于我（数据后续走后端配置） -->
    <div class="widgets">
      <GithubStatusCard />
      <AboutMeCard />
    </div>
  </main>
</template>

<style scoped lang="scss">
.page {
  max-width: 1180px;
  margin: 0 auto;
  padding: 110px 24px 80px;
}

/* ===== 轮播背景壳（常驻占位） ===== */
.hero-shell {
  --hero-w: min(1400px, 100vw - 48px);
  width: var(--hero-w);
  margin-inline: calc((var(--hero-w) - 100%) / -2);
  min-height: 62vh;
  padding: 36px 56px;
  border-radius: 24px;
  background:
    radial-gradient(560px 300px at 82% 24%, rgba(var(--primary-rgb), 0.08), transparent 65%),
    radial-gradient(480px 260px at 12% 80%, rgba(var(--primary-rgb), 0.05), transparent 65%),
    linear-gradient(180deg, rgba(var(--primary-rgb), 0.045), rgba(var(--primary-rgb), 0.015));
  border: 1px solid rgba(var(--primary-rgb), 0.08);
}

@media (max-width: 900px) {
  .hero-shell {
    min-height: auto;
    padding: 20px 16px;
  }
}

/* ===== 最新文章：交错图文平铺 ===== */
.latest {
  margin-top: 80px;
}

.section-title {
  font-size: clamp(24px, 3vw, 32px);
  margin-bottom: 8px;
}

/* ===== 扩展板块 ===== */
.widgets {
  margin-top: 56px;
  display: flex;
  flex-direction: column;
  gap: 28px;
}

@media (max-width: 768px) {
  .page { padding-top: 88px; }
}
</style>
