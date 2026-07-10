<script setup lang="ts">
import HeroCarousel, { type HeroItem } from '../components/home/HeroCarousel.vue';
import { placeholderCover } from '../utils/placeholder';

/* P1 接后端前的演示数据：两组，每组三张卡 */
const groups: HeroItem[][] = [
  [
    {
      title: '用 CSS Variables 打造季节主题系统',
      excerpt: '两层主题架构：深浅模式与季节色盘正交组合，四季配色一键切换，品牌荧光三色贯穿始终。',
      cover: placeholderCover('#ff0032', '#7a0020', 'Spring'),
      tag: '技术',
    },
    {
      title: '毛玻璃与 3D：简约的惊艳',
      excerpt: 'backdrop-filter 胶囊导航、透视卡片飞入飞出，克制的动效如何撑起高级感。',
      cover: placeholderCover('#0078ff', '#00295c', 'Glass'),
      tag: '设计',
    },
    {
      title: 'Markdown 作为博客的统一内容规范',
      excerpt: '从后台编辑到 API 发文，一切内容皆 Markdown：可移植、可版本化、对 AI 友好。',
      cover: placeholderCover('#ffb300', '#7a5200', 'Markdown'),
      tag: '架构',
    },
  ],
  [
    {
      title: '博客后台与 APIKey 发文中心',
      excerpt: 'JWT 管理员后台负责日常编辑，APIKey 认证接口留给外部 AI 代笔——懒人的完整闭环。',
      cover: placeholderCover('#2b6cb0', '#0d1f33', 'API'),
      tag: '后端',
    },
    {
      title: '思源黑体与宋体的排版实践',
      excerpt: '黑体承担 UI 与正文，宋体点睛标题与文章，离线 woff2 子集化控制体积。',
      cover: placeholderCover('#c78800', '#3a2800', 'Fonts'),
      tag: '排版',
    },
  ],
];
</script>

<template>
  <main class="page">
    <HeroCarousel :groups="groups" />

    <section v-reveal class="teaser">
      <h2>最新文章</h2>
      <p>P1 阶段接入文章列表。</p>
    </section>
  </main>
</template>

<style scoped lang="scss">
.page {
  max-width: 1180px;
  margin: 0 auto;
  padding: 110px 24px 80px;
}

.teaser {
  margin-top: 64px;
  padding: 32px;
  border-radius: var(--radius-lg);
  background: var(--surface);
  border: 1px solid var(--border);

  h2 { margin-bottom: 8px; }
  p { color: var(--text-2); }
}

@media (max-width: 768px) {
  .page { padding-top: 88px; }
}
</style>
