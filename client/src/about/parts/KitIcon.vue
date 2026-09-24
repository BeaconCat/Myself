<script setup lang="ts">
import { computed } from 'vue';
import { ICONS } from '../icons';

/** 关于页线性图标：name 取自 about/icons.ts，未知名称回落为 link */
const props = withDefaults(defineProps<{ name: string; size?: number }>(), { size: 18 });
const inner = computed(() => ICONS[props.name] ?? ICONS.link);
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <svg class="ak-i" viewBox="0 0 24 24" :width="size" :height="size" aria-hidden="true" v-html="inner" />
</template>
