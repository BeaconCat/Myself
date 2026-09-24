<script setup lang="ts">
import { computed } from 'vue';
import { ICONS } from './icons';

/** Studio 线性图标 */
const props = withDefaults(defineProps<{ name: string; size?: number }>(), { size: 18 });
const body = computed(() => ICONS[props.name] ?? '');
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <svg class="st-ic" :width="size" :height="size" viewBox="0 0 24 24" aria-hidden="true" v-html="body" />
</template>
