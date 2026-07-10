<script setup lang="ts">
import { onMounted, ref } from 'vue';
import HeroCarousel, { type HeroItem } from '../components/home/HeroCarousel.vue';
import { placeholderCover } from '../utils/placeholder';
import { api, type Post } from '../api';
import PostZigzagList from '../components/post/PostZigzagList.vue';
import GithubStatusCard from '../components/home/GithubStatusCard.vue';
import AboutMeCard from '../components/home/AboutMeCard.vue';

const latest = ref<Post[]>([]);

onMounted(async () => {
  latest.value = (await api.posts({ pageSize: 4 })).items;
});

/* P1 接后端前的演示数据：两组；covers 1–3 张演示立体相册 */
const groups: HeroItem[][] = [
  [
    {
      title: '用 CSS Variables 打造季节主题系统',
      excerpt: '两层主题架构：深浅模式与季节色盘正交组合，四季配色一键切换，品牌荧光三色贯穿始终。',
      covers: [
        placeholderCover('#ff0032', '#7a0020', 'Spring 1'),
        placeholderCover('#ff5c7a', '#a3002a', 'Spring 2'),
        placeholderCover('#ffb3c2', '#d40029', 'Spring 3'),
      ],
      tag: '技术',
    },
    {
      title: '毛玻璃与 3D：简约的惊艳',
      excerpt: 'backdrop-filter 胶囊导航、透视卡片飞入飞出，克制的动效如何撑起高级感。',
      covers: [
        placeholderCover('#0078ff', '#00295c', 'Glass 1'),
        placeholderCover('#4d9fff', '#003d80', 'Glass 2'),
      ],
      tag: '设计',
    },
    {
      title: 'Markdown 作为博客的统一内容规范',
      excerpt: '从后台编辑到 API 发文，一切内容皆 Markdown：可移植、可版本化、对 AI 友好。',
      covers: [placeholderCover('#ffb300', '#7a5200', 'Markdown')],
      tag: '架构',
    },
  ],
  [
    {
      title: '博客后台与 APIKey 发文中心',
      excerpt: 'JWT 管理员后台负责日常编辑，APIKey 认证接口留给外部 AI 代笔——懒人的完整闭环。',
      covers: [
        placeholderCover('#2b6cb0', '#0d1f33', 'API 1'),
        placeholderCover('#5a9bd5', '#123c66', 'API 2'),
        placeholderCover('#8ec3ee', '#1e5590', 'API 3'),
      ],
      tag: '后端',
    },
    {
      title: '思源黑体与宋体的排版实践',
      excerpt: '黑体承担 UI 与正文，宋体点睛标题与文章，离线 woff2 子集化控制体积。',
      covers: [
        placeholderCover('#c78800', '#3a2800', 'Fonts 1'),
        placeholderCover('#ffd166', '#7a5200', 'Fonts 2'),
      ],
      tag: '排版',
    },
  ],
];
</script>

<template>
  <main class="page">
    <HeroCarousel :groups="groups" />

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
