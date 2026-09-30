<script setup lang="ts">
import { computed } from 'vue';
import { TECH } from '../tech';

/** 技术栈品牌图标（simple-icons 实心路径）：颜色取 currentColor，由外层按品牌色着色；未知键不渲染 */
const props = withDefaults(defineProps<{ name?: string; size?: number }>(), { name: '', size: 18 });
const tech = computed(() => TECH[props.name]);
</script>

<template>
  <svg v-if="tech" class="ak-tech" viewBox="0 0 24 24" :width="size" :height="size" fill="currentColor" aria-hidden="true">
    <path :d="tech.path" />
  </svg>
</template>

<style>
svg.ak-tech { flex: none; fill: currentColor; stroke: none; }
</style>
