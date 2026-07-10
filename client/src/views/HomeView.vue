<script setup lang="ts">
import { onMounted, ref } from 'vue';
import HeroCarousel, { type HeroItem } from '../components/home/HeroCarousel.vue';
import { placeholderCover } from '../utils/placeholder';
import { api, type Post } from '../api';

const latest = ref<Post[]>([]);

/** 无封面时按标签色生成占位图 */
const fallbackPalette = [
  ['#ff0032', '#7a0020'],
  ['#ffb300', '#7a5200'],
  ['#0078ff', '#00295c'],
  ['#2b6cb0', '#0d1f33'],
];

function coverOf(post: Post, index: number): string {
  if (post.covers.length) return post.covers[0];
  const [from, to] = fallbackPalette[index % fallbackPalette.length];
  return placeholderCover(from, to, post.tags[0] ?? 'Post');
}

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
      <router-link
        v-for="(post, i) in latest"
        :key="post.slug"
        v-reveal
        :to="`/articles/${post.slug}`"
        class="zig"
        :class="{ flip: i % 2 === 1 }"
      >
        <div class="zig-text">
          <span class="z-date">{{ post.createdAt.slice(0, 10) }}</span>
          <h3>{{ post.title }}</h3>
          <p>{{ post.excerpt }}</p>
          <div class="z-tags">
            <span v-for="tag in post.tags" :key="tag">{{ tag }}</span>
          </div>
        </div>
        <div class="zig-media">
          <img :src="coverOf(post, i)" :alt="post.title" loading="lazy" draggable="false" />
        </div>
      </router-link>
    </section>

    <!-- 预留扩展板块：后续接 GitHub Status（热力图 / 最新动态 / commit），数据源走后端配置 -->
    <section v-reveal class="widgets">
      <div class="widget-placeholder">
        <span class="w-label">GitHub Status</span>
        <span class="w-hint">板块预留 · 后端配置接入热力图与最新动态</span>
      </div>
    </section>
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

.zig {
  display: grid;
  grid-template-columns: 1.15fr 1fr;
  gap: 40px;
  align-items: center;
  padding: 34px 28px;
  position: relative;
  transition: background var(--dur-fast);

  /* 偶数条翻转：图左文右 */
  &.flip .zig-text { order: 2; }
  &.flip .zig-media { order: 1; }

  /* 悬停直角边框：四角括号式描边 */
  &::before,
  &::after {
    content: '';
    position: absolute;
    width: 26px;
    height: 26px;
    opacity: 0;
    transition: opacity var(--dur-fast) ease, transform var(--dur) var(--ease-out);
    pointer-events: none;
  }

  &::before {
    top: 10px;
    left: 10px;
    border-top: 2px solid var(--primary);
    border-left: 2px solid var(--primary);
    transform: translate(8px, 8px);
  }

  &::after {
    bottom: 10px;
    right: 10px;
    border-bottom: 2px solid var(--primary);
    border-right: 2px solid var(--primary);
    transform: translate(-8px, -8px);
  }

  &:hover {
    background: rgba(var(--primary-rgb), 0.04);

    &::before, &::after { opacity: 1; transform: none; }
    .zig-media img { transform: scale(1.04); }
    h3 { color: var(--primary); }
  }
}

.zig-text {
  min-width: 0;

  .z-date {
    font-size: 13px;
    color: var(--text-2);
    font-variant-numeric: tabular-nums;
  }

  h3 {
    font-size: clamp(20px, 2.4vw, 26px);
    margin: 10px 0 12px;
    transition: color var(--dur-fast);
  }

  p {
    font-size: 14px;
    line-height: 1.8;
    color: var(--text-2);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
}

.z-tags {
  display: flex;
  gap: 8px;
  margin-top: 14px;

  span {
    font-size: 11px;
    font-weight: 600;
    padding: 2px 10px;
    background: rgba(var(--primary-rgb), 0.1);
    color: var(--primary);
  }
}

.zig-media {
  overflow: hidden;
  aspect-ratio: 16 / 9;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    transition: transform var(--dur-slow) var(--ease-out);
  }
}

/* ===== 预留扩展板块 ===== */
.widgets {
  margin-top: 56px;
}

.widget-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 44px 24px;
  border: 1px dashed var(--border);
  color: var(--text-2);

  .w-label {
    font-family: var(--font-serif);
    font-size: 17px;
    font-weight: 700;
    color: var(--text);
  }

  .w-hint { font-size: 13px; }
}

@media (max-width: 768px) {
  .page { padding-top: 88px; }

  .zig {
    grid-template-columns: 1fr;
    gap: 18px;
    padding: 22px 14px;

    /* 移动端统一图上文下 */
    &.flip .zig-text { order: 2; }
    &.flip .zig-media { order: 1; }
    .zig-text { order: 2; }
    .zig-media { order: 1; }
  }
}
</style>
