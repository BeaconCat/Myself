<script setup lang="ts">
import { computed } from 'vue';
import type { IconNode } from 'lucide';
import { iconMarkup } from '../../utils/lucide';

/**
 * 通用标准图标：直接传入 Lucide 图标（import { ArrowLeft } from 'lucide'），按需引入、离线。
 * 用于组件里的零散图标；成套的具名图标仍走各自的图标组件（SIcon / MIcon / UiIcon …）。
 */
const props = withDefaults(defineProps<{ icon: IconNode; size?: number | string; stroke?: number | string }>(), {
  size: 18,
  stroke: 1.8,
});
const body = computed(() => iconMarkup(props.icon));
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <svg
    class="l-ic"
    viewBox="0 0 24 24"
    :width="size"
    :height="size"
    fill="none"
    stroke="currentColor"
    :stroke-width="stroke"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
    v-html="body"
  />
</template>

<style>
.l-ic { flex: none; }
</style>
