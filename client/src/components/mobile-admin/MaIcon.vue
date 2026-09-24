<script setup lang="ts">
/** 线性图标：内联 SVG（离线），选中态由父级给 .on 后填充 .f 面。 */
import { computed } from 'vue';
import { ICONS, type IconName } from './icons';

const props = withDefaults(defineProps<{ name: IconName; size?: number }>(), { size: 24 });
const body = computed(() => ICONS[props.name]);
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <svg class="ma-ic" :width="size" :height="size" viewBox="0 0 24 24" aria-hidden="true" v-html="body" />
</template>

<style lang="scss">
.ma-ic {
  flex: none;
  display: block;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.5;
  stroke-linecap: round;
  stroke-linejoin: round;

  .f { fill: transparent; transition: fill var(--dur) var(--ease-out); }
  .d { transition: stroke var(--dur) var(--ease-out); }
}
</style>
