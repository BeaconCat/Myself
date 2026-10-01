<script setup lang="ts">
import { ref } from 'vue';
import type { Post } from '../../api';
import CoverArt from '../common/CoverArt.vue';
import PostMeta from './PostMeta.vue';

/**
 * 首篇封面卡：封面（悬停 3D 轻倾 + 光影微放大）与元信息 / 宋体标题 / 摘要同在一张卡内；
 * 封面较旧版收一档（2.1:1 → 2.6:1），标题与摘要更醒目。
 */
withDefaults(defineProps<{ post: Post; ratio?: string; meta?: string }>(), { ratio: '2.6 / 1', meta: '' });

const cover = ref<HTMLElement | null>(null);
const reduced = window.matchMedia('(prefers-reduced-motion: reduce)');
let raf = 0;

function onMove(e: MouseEvent): void {
  const el = cover.value;
  if (!el || reduced.matches) return;
  cancelAnimationFrame(raf);
  raf = requestAnimationFrame(() => {
    const r = el.getBoundingClientRect();
    const x = (e.clientX - r.left) / r.width - 0.5;
    const y = (e.clientY - r.top) / r.height - 0.5;
    el.style.transform = `perspective(1400px) rotateX(${(-y * 3).toFixed(2)}deg) rotateY(${(x * 3).toFixed(2)}deg)`;
  });
}

function onLeave(): void {
  cancelAnimationFrame(raf);
  if (cover.value) cover.value.style.transform = '';
}
</script>

<template>
  <router-link :to="`/articles/${post.slug}`" class="feat" @mousemove="onMove" @mouseleave="onLeave">
    <div ref="cover" class="cover" :style="{ aspectRatio: ratio }">
      <CoverArt :src="post.covers[0]" :seed="post.slug" />
    </div>
    <div class="info">
      <PostMeta :post="post" :extra="meta" />
      <h3><span>{{ post.title }}</span></h3>
      <p v-if="post.excerpt" class="post-excerpt">{{ post.excerpt }}</p>
    </div>
  </router-link>
</template>

<style scoped lang="scss">
.feat {
  display: block;
  border-radius: var(--card-r);
  background: var(--card-bg);
  box-shadow: var(--card-shadow);
  outline: none;
  transition: box-shadow var(--dur) var(--ease-out);

  &:hover { box-shadow: var(--card-shadow-hover); }
  &:focus-visible { box-shadow: var(--card-shadow), var(--focus); }
}

.info { padding: 18px var(--card-pad) 22px; }

/* 简洁风格：封面四角圆、无卡底，信息块直接落在页面底色上 */
:root[data-style='clean'] {
  .info { padding-bottom: 6px; }
  .cover { border-radius: var(--r-lg); }
}

/* 封面容器：品牌光影底 + 0.5px 内描边；阴影为中性色 */
.cover {
  position: relative;
  overflow: hidden;
  isolation: isolate;
  border-radius: var(--r-lg) var(--r-lg) 0 0;
  background: #040914;
  transition: transform var(--dur) var(--ease-out), box-shadow var(--dur) var(--ease-out);

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

.feat:hover .cover :deep(.cv) { transform: scale(1.035); }

h3 {
  margin-top: 6px;
  font-family: var(--font-serif);
  font-size: 26px;
  font-weight: 700;
  line-height: 1.35;
  color: var(--text);
  text-wrap: balance;

  /* 悬停下划线从左生长 */
  span {
    background: linear-gradient(currentColor, currentColor) 0 100% / 0 1.5px no-repeat;
    transition: background-size var(--dur-slow) var(--ease-out);
  }
}

.feat:hover h3 span { background-size: 100% 1.5px; }

p {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-top: 6px;
  font-size: 15px;
  line-height: 1.75;
  color: var(--text-2);
}

@media (max-width: 1100px) {
  h3 { font-size: 24px; }
}
</style>
