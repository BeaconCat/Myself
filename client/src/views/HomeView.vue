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
    <HeroCarousel v-if="heroItems.length" :items="heroItems" :photo-ms="heroInterval" />

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
