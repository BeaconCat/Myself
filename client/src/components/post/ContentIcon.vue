<script setup lang="ts">
/**
 * 内容区线性图标（24 网格，1.5 描边，currentColor），全部内联离线。
 * 与设计稿 round3 的符号表一致；size：s = 16，xs = 14，默认 18。
 */
defineProps<{ name: string; size?: 's' | 'xs' }>();
</script>

<script lang="ts">
const P: Record<string, string> = {
  search: '<circle cx="11" cy="11" r="6.5"/><path d="m16 16 4 4"/>',
  bubble: '<path d="M5 5h14a1 1 0 0 1 1 1v9.5a1 1 0 0 1-1 1h-7.5L7 20v-3.5H5a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1zM8 9.5h8M8 12.5h5"/>',
  image: '<rect x="3.5" y="4.5" width="17" height="15" rx="2"/><circle cx="9" cy="9.5" r="1.5"/><path d="m4 17 5-4.5 4 3.5 3-2.5 4 3.5"/>',
  cal: '<rect x="4" y="5.5" width="16" height="14.5" rx="2"/><path d="M4 10h16M8.5 3.5v3.5M15.5 3.5v3.5"/>',
  arrowL: '<path d="M19 12H5m5-5-5 5 5 5"/>',
  arrowR: '<path d="M5 12h14m-5-5 5 5-5 5"/>',
  chevL: '<path d="m14 6-6 6 6 6"/>',
  chevR: '<path d="m10 6 6 6-6 6"/>',
  share: '<path d="M12 15V4m-4 4 4-4 4 4M6 11v8a1 1 0 0 0 1 1h10a1 1 0 0 0 1-1v-8"/>',
  link: '<path d="M10 14a4 4 0 0 0 5.7 0l3-3a4 4 0 0 0-5.7-5.7l-1 1M14 10a4 4 0 0 0-5.7 0l-3 3a4 4 0 0 0 5.7 5.7l1-1"/>',
  type: '<path d="M3.5 18 8 6l4.5 12M5.2 14h5.6M14.5 18l3-7.5 3 7.5M15.5 15.5h4"/>',
  list: '<path d="M9 7h11M9 12h11M9 17h11"/><circle cx="4.8" cy="7" r=".6"/><circle cx="4.8" cy="12" r=".6"/><circle cx="4.8" cy="17" r=".6"/>',
  up: '<path d="M12 19V5m-6 6 6-6 6 6"/>',
  x: '<path d="M6 6l12 12M18 6 6 18"/>',
  check: '<path d="m5 12.5 4.5 4.5L19 7.5"/>',
  copy: '<rect x="8.5" y="8.5" width="11" height="11" rx="2"/><path d="M15.5 8.5V6a1.5 1.5 0 0 0-1.5-1.5H6A1.5 1.5 0 0 0 4.5 6v8A1.5 1.5 0 0 0 6 15.5h2.5"/>',
};
export default {};
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <svg class="ci" :class="size" viewBox="0 0 24 24" aria-hidden="true" v-html="P[name]" />
</template>

<style scoped>
.ci {
  width: 18px;
  height: 18px;
  flex: none;
  display: block;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.5;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.ci.s { width: 16px; height: 16px; }
.ci.xs { width: 14px; height: 14px; }
</style>
