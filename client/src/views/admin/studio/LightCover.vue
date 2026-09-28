<script setup lang="ts">
import { computed } from 'vue';
import { coverForKind, defaultCover } from '../../../utils/defaultCovers';

/**
 * 后台封面位：有 src 显示图片；否则取默认封面（kind 指定固定一幅，缺省按 seed 稳定取）。
 * 默认封面见 utils/defaultCovers.ts（程序生成的抽象构图，与前台同一套）。
 */
const props = defineProps<{ src?: string; kind?: string; seed?: number | string; small?: boolean }>();

const url = computed(() => {
  if (props.src) return props.src;
  if (props.kind) return coverForKind(props.kind, props.small);
  return defaultCover(props.seed ?? 0, props.small ?? true);
});
</script>

<template>
  <div class="cv">
    <img :src="url" alt="" loading="lazy" draggable="false" />
    <slot />
  </div>
</template>

<style scoped lang="scss">
.cv {
  position: relative;
  overflow: hidden;
  isolation: isolate;
  background: var(--well-2);

  img { position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover; display: block; }
}
</style>
