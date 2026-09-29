<script setup lang="ts">
import { computed } from 'vue';
import { ICONS } from '../icons';
import { BRANDS } from '../brands';

/**
 * 关于页图标：社交平台名（about/brands.ts）渲染 simple-icons 实心品牌图标，
 * 其余取 about/icons.ts 线性图标；未知名称回落为 link。
 */
const props = withDefaults(defineProps<{ name: string; size?: number }>(), { size: 18 });
const brand = computed(() => BRANDS[props.name]);
const inner = computed(() => ICONS[props.name] ?? ICONS.link);
</script>

<template>
  <svg v-if="brand" class="ak-i ak-brand" viewBox="0 0 24 24" :width="size * 0.9" :height="size * 0.9" aria-hidden="true">
    <path :d="brand.path" />
  </svg>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <svg v-else class="ak-i" viewBox="0 0 24 24" :width="size" :height="size" aria-hidden="true" v-html="inner" />
</template>

<style>
/* 基础线性样式不依赖 .ak 作用域：后台身份页、移动端名片等模块流之外的位置同样正确 */
svg.ak-i {
  flex: none;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.6;
  stroke-linecap: round;
  stroke-linejoin: round;
}

/* 品牌图标为实心路径：覆盖线性图标的描边样式（任何作用域下都生效） */
svg.ak-i.ak-brand {
  fill: currentColor !important;
  stroke: none !important;
}
</style>
