<script setup lang="ts">
import type { Post } from '../../api';
import CoverArt from '../common/CoverArt.vue';
import PostMeta from './PostMeta.vue';

/** 缩略行（紧凑）：元信息 + 宋体标题 + 一行摘要 + 右侧缩略封面（与移动端 .m-arow 同构） */
withDefaults(defineProps<{ post: Post; meta?: string }>(), { meta: '' });
</script>

<template>
  <router-link :to="`/articles/${post.slug}`" class="arow">
    <div class="rt">
      <PostMeta :post="post" :extra="meta" />
      <span class="t"><span>{{ post.title }}</span></span>
      <span v-if="post.excerpt" class="post-excerpt ex">{{ post.excerpt }}</span>
    </div>
    <div class="thumb"><CoverArt :src="post.covers[0]" :seed="post.slug" thumb /></div>
  </router-link>
</template>

<style scoped lang="scss">
.arow {
  position: relative;
  display: flex;
  align-items: center;
  gap: 20px;
  margin: 0 -16px;
  padding: 14px 16px;
  border-radius: var(--r-lg);
  outline: none;
  transition: background-color var(--dur-fast);

  /* 行间 0.5px 分隔线：悬停行及其下一行的分隔线淡出 */
  & + &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 16px;
    right: 16px;
    height: 0.5px;
    background: var(--line-2);
    transition: opacity var(--dur-fast);
  }

  &:hover { background: var(--fill); }
  &:hover::before, &:hover + &::before { opacity: 0; }
  &:focus-visible { box-shadow: var(--focus); }
}

.rt {
  flex: 1;
  min-width: 0;
}

.t {
  display: block;
  margin-top: 3px;
  font-family: var(--font-serif);
  font-size: 20px;
  font-weight: 700;
  line-height: 1.45;
  color: var(--text);

  span {
    background: linear-gradient(currentColor, currentColor) 0 100% / 0 1px no-repeat;
    transition: background-size var(--dur) var(--ease-out);
  }
}

.arow:hover .t span { background-size: 100% 1px; }

.ex {
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-top: 3px;
  font-size: 15px;
  line-height: 1.65;
  color: var(--text-2);
}

.thumb {
  position: relative;
  flex: none;
  width: 120px;
  height: 76px;
  overflow: hidden;
  isolation: isolate;
  border-radius: var(--r-md);
  background: #040914;

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

:root[data-mode='light'] .thumb::after { box-shadow: inset 0 0 0 0.5px rgb(16 24 40 / 0.1); }

.arow:hover .thumb :deep(.cv) { transform: scale(1.05); }
</style>
