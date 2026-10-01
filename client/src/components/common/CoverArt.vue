<script lang="ts">
/** 服务端渐变占位图（/api/v1/img/...）视同无图：改用默认封面，避免纯色块 */
export function isArtUrl(url: string | undefined): boolean {
  return !url || url.startsWith('/api/v1/img/');
}
</script>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { thumbOf } from '../../api';
import { defaultCover } from '../../utils/defaultCovers';

/**
 * 封面：有真实封面用图片（thumb 时走按需缩略图）；没有时按 seed 稳定取一幅默认封面
 * （预生成的抽象构图，thumb 时用 640px 小图）。图片加载完成后淡入。
 * pool 为历史参数（随想配图占位），现与文章共用同一组默认封面。
 */
const props = withDefaults(
  defineProps<{ src?: string; seed: string; thumb?: boolean; pool?: 'post' | 'all' }>(),
  { src: '', thumb: false, pool: 'post' },
);

const url = computed(() => {
  if (isArtUrl(props.src)) return defaultCover(props.seed, props.thumb);
  return props.thumb ? thumbOf(props.src) : props.src;
});
const loaded = ref(false);
watch(url, () => { loaded.value = false; });
</script>

<template>
  <div class="cv cv-img" :class="{ loaded }">
    <img :src="url" alt="" draggable="false" decoding="async" @load="loaded = true" />
  </div>
</template>

<style scoped lang="scss">
.cv {
  position: absolute;
  inset: 0;
  overflow: hidden;
  isolation: isolate;
  /* 父级悬停时可对 .cv 做轻微放大 */
  transition: transform var(--dur-slow) var(--ease-out);
}

.cv-img {
  background: var(--m-fill-2, var(--fill-2));

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    opacity: 0;
    transform: scale(1.04);
    transition: opacity 0.5s var(--ease-out), transform 0.8s var(--ease-out);
  }

  &.loaded img {
    opacity: 1;
    transform: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .cv-img img { transition: none; }
}
</style>
