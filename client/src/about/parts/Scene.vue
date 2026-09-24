<script setup lang="ts">
/** CSS 光影构图（无真实图片时的封面）；src 存在时显示图片 */
defineProps<{ scene?: string; src?: string; alt?: string }>();
</script>

<template>
  <img v-if="src" class="ak-scene-img" :src="src" :alt="alt ?? ''" loading="lazy" draggable="false" />
  <span v-else class="ak-scene" :class="`sc-${scene || 'door'}`" aria-hidden="true" />
</template>
