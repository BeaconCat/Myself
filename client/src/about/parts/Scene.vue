<script setup lang="ts">
import { computed } from 'vue';
import { LEGACY_SCENES, SCENES } from '../icons';
import { coverUrl } from '../../utils/defaultCovers';

/** 模块配图：有 src 显示图片；否则按 scene 取一幅默认封面（兼容旧版 CSS 光影场景名） */
const props = defineProps<{ scene?: string; src?: string; alt?: string; small?: boolean }>();

const url = computed(() => {
  if (props.src) return props.src;
  const s = props.scene ?? '';
  const id = (SCENES as readonly string[]).includes(s) ? s : LEGACY_SCENES[s] ?? '05';
  return coverUrl(id, props.small);
});
</script>

<template>
  <img class="ak-scene-img" :src="url" :alt="src ? alt ?? '' : ''" loading="lazy" draggable="false" />
</template>
