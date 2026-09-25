<script setup lang="ts">
import { ref } from 'vue';
import type { Post } from '../../api';
import CoverArt from '../common/CoverArt.vue';
import PostMeta from './PostMeta.vue';

/** 首篇大卡：宽封面（悬停 3D 轻倾 + 光影微放大）+ 元信息 + 宋体标题 + 摘要（与移动端 .feat 同构） */
withDefaults(defineProps<{ post: Post; ratio?: string; meta?: string }>(), { ratio: '2.1 / 1', meta: '' });

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
    <PostMeta class="fmeta" :post="post" :extra="meta" />
    <h3><span>{{ post.title }}</span></h3>
    <p v-if="post.excerpt">{{ post.excerpt }}</p>
  </router-link>
</template>

<style scoped lang="scss">
.feat {
  display: block;
  border-radius: var(--r-xl);
  outline: none;

  &:focus-visible { box-shadow: var(--focus); }
}

/* 封面容器：品牌光影底 + 0.5px 内描边；阴影为中性色 */
.cover {
  position: relative;
  overflow: hidden;
  isolation: isolate;
  border-radius: var(--r-xl);
  background: #040914;
  box-shadow: 0 30px 60px -36px rgb(0 0 0 / 0.8);
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

:root[data-mode='light'] .cover {
  box-shadow: 0 30px 60px -36px rgb(16 24 40 / 0.45);

  &::after { box-shadow: inset 0 0 0 0.5px rgb(16 24 40 / 0.1); }
}

.feat:hover .cover :deep(.cv) { transform: scale(1.035); }

.fmeta { margin-top: 20px; }

h3 {
  margin-top: 8px;
  font-family: var(--font-serif);
  font-size: 30px;
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
  max-width: 60ch;
  margin-top: 10px;
  font-size: 15px;
  line-height: 1.8;
  color: var(--text-2);
}

@media (max-width: 1100px) {
  h3 { font-size: 26px; }
}
</style>
