<script setup lang="ts">
/**
 * 桌面外壳线性图标（24 网格，1.6 描边，全部离线内联）。
 * 与移动端 MIcon 同一绘制风格；新增图标直接往 P 里加 path。
 */
const P: Record<string, string> = {
  search: '<circle cx="10.8" cy="10.8" r="6.3"/><path d="m15.6 15.6 4.4 4.4"/>',
  history: '<path d="M4.5 12a7.5 7.5 0 1 0 2.2-5.3L4.5 9"/><path d="M4.5 4.8V9h4.2M12 8v4.2l2.8 1.8"/>',
  enter: '<path d="M19 5.5v6.2a2.3 2.3 0 0 1-2.3 2.3H6"/><path d="m9.5 10.5-3.5 3.5 3.5 3.5"/>',
  book: '<path d="M12 6.7C10 5.3 7.2 4.7 3.8 5.1v13.4c3.4-.4 6.2.2 8.2 1.6 2-1.4 4.8-2 8.2-1.6V5.1c-3.4-.4-6.2.2-8.2 1.6z"/><path d="M12 6.7v13.4"/>',
  bubble: '<path d="M6 4.6h12a2.4 2.4 0 0 1 2.4 2.4v7.4a2.4 2.4 0 0 1-2.4 2.4h-5.3L8 20.3v-3.5H6a2.4 2.4 0 0 1-2.4-2.4V7A2.4 2.4 0 0 1 6 4.6z"/><path d="M8.3 9.4h7.4M8.3 12.4h4.6"/>',
  user: '<circle cx="12" cy="8.2" r="3.6"/><path d="M4.9 19.6c.8-3.5 3.7-5.4 7.1-5.4s6.3 1.9 7.1 5.4"/>',
  hash: '<path d="M9.5 4 8 20M16 4l-1.5 16M4.5 9h16M3.5 15h16"/>',
  rss: '<path d="M5 5.5a13.5 13.5 0 0 1 13.5 13.5M5 11a8 8 0 0 1 8 8"/><circle cx="6" cy="18" r="1.3" fill="currentColor"/>',
  github: '<path d="M9 19.5c-4 1.2-4-2-5.6-2.4M14.6 21.5v-3.2a2.8 2.8 0 0 0-.8-2.2c2.6-.3 5.3-1.3 5.3-5.8a4.5 4.5 0 0 0-1.2-3.1 4.2 4.2 0 0 0-.1-3.1s-1-.3-3.2 1.2a11 11 0 0 0-5.8 0C6.6 3.8 5.6 4.1 5.6 4.1a4.2 4.2 0 0 0-.1 3.1 4.5 4.5 0 0 0-1.2 3.1c0 4.5 2.7 5.5 5.3 5.8a2.8 2.8 0 0 0-.8 2.2v3.2"/>',
  close: '<path d="M6.5 6.5l11 11M17.5 6.5l-11 11"/>',
  sun: '<circle cx="12" cy="12" r="3.8"/><path d="M12 3v1.8M12 19.2V21M3 12h1.8M19.2 12H21M5.6 5.6l1.3 1.3M17.1 17.1l1.3 1.3M5.6 18.4l1.3-1.3M17.1 6.9l1.3-1.3"/>',
  moon: '<path d="M19.5 14.6A7.8 7.8 0 1 1 9.4 4.5a6.2 6.2 0 0 0 10.1 10.1z"/>',
  pen: '<path d="M12 20h8"/><path d="M16.3 3.8a2.1 2.1 0 0 1 3 3L7.5 18.6l-4 1 1-4z"/>',
  doc: '<path d="M6 3.5h8.5L19 8v12.5H6z"/><path d="M14 3.5V8.5h5M9 12.5h6M9 16h4"/>',
  gear: '<circle cx="12" cy="12" r="3"/><path d="M19.4 13.5a7.6 7.6 0 0 0 0-3l2-1.6-2-3.4-2.4.9a7.5 7.5 0 0 0-2.6-1.5L14 2.5h-4l-.4 2.4A7.5 7.5 0 0 0 7 6.4l-2.4-.9-2 3.4 2 1.6a7.6 7.6 0 0 0 0 3l-2 1.6 2 3.4 2.4-.9a7.5 7.5 0 0 0 2.6 1.5l.4 2.4h4l.4-2.4a7.5 7.5 0 0 0 2.6-1.5l2.4.9 2-3.4z"/>',
};

defineProps<{ name: string }>();
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <svg class="ui-ic" viewBox="0 0 24 24" aria-hidden="true" v-html="P[name] ?? ''" />
</template>

<style lang="scss">
.ui-ic {
  width: 20px;
  height: 20px;
  flex: none;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.6;
  stroke-linecap: round;
  stroke-linejoin: round;

  &.s { width: 16px; height: 16px; }
  &.xs { width: 14px; height: 14px; }
}
</style>
