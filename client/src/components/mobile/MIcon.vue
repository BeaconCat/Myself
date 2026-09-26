<script setup lang="ts">
/**
 * 移动端线性图标（24 网格，1.5 描边）。.f = 选中态填充面，.d = 选中态镂空细节。
 * 与站内其余内联 SVG 图标同一风格，全部离线。
 */
const P: Record<string, string> = {
  home: '<path class="f" d="M4 10.4 12 4l8 6.4V19a1.2 1.2 0 0 1-1.2 1.2H15v-5.2H9v5.2H5.2A1.2 1.2 0 0 1 4 19z"/>',
  book: '<path class="f" d="M12 6.7C10 5.3 7.2 4.7 3.8 5.1v13.4c3.4-.4 6.2.2 8.2 1.6 2-1.4 4.8-2 8.2-1.6V5.1c-3.4-.4-6.2.2-8.2 1.6z"/><path class="d" d="M12 6.7v13.4"/>',
  bubble: '<path class="f" d="M6 4.6h12a2.4 2.4 0 0 1 2.4 2.4v7.4a2.4 2.4 0 0 1-2.4 2.4h-5.3L8 20.3v-3.5H6a2.4 2.4 0 0 1-2.4-2.4V7A2.4 2.4 0 0 1 6 4.6z"/><path class="d" d="M8.3 9.4h7.4M8.3 12.4h4.6"/>',
  user: '<circle class="f" cx="12" cy="8.2" r="3.6"/><path class="f" d="M4.9 19.6c.8-3.5 3.7-5.4 7.1-5.4s6.3 1.9 7.1 5.4z"/>',
  search: '<circle class="f" cx="10.8" cy="10.8" r="6.3"/><path d="m15.6 15.6 4.4 4.4"/>',
  image: '<rect class="f" x="3.6" y="4.6" width="16.8" height="14.8" rx="2.6"/><path class="d" d="m4.2 17 4.6-4.6a1.4 1.4 0 0 1 2 0L16 17.6M13.6 15l1.8-1.8a1.4 1.4 0 0 1 2 0l2.6 2.6"/><circle class="d" cx="15.4" cy="8.8" r="1.4"/>',
  back: '<path d="M14.5 5.5 8 12l6.5 6.5"/>',
  chev: '<path d="m9.5 6 6 6-6 6"/>',
  close: '<path d="M6.5 6.5l11 11M17.5 6.5l-11 11"/>',
  share: '<path d="M12 3.8v11M8 7.6l4-3.8 4 3.8"/><path d="M6.5 11H6a1.6 1.6 0 0 0-1.6 1.6v6a1.6 1.6 0 0 0 1.6 1.6h12a1.6 1.6 0 0 0 1.6-1.6v-6A1.6 1.6 0 0 0 18 11h-.5"/>',
  list: '<path d="M9 6.5h11M9 12h11M9 17.5h11"/><circle cx="4.8" cy="6.5" r=".9"/><circle cx="4.8" cy="12" r=".9"/><circle cx="4.8" cy="17.5" r=".9"/>',
  type: '<path d="M3.5 18.5 8 6.5l4.5 12M5.2 14h5.6M14.5 18.5l3-8 3 8M15.4 16h4.2"/>',
  rss: '<path d="M5 5.5a13.5 13.5 0 0 1 13.5 13.5M5 11a8 8 0 0 1 8 8"/><circle cx="6" cy="18" r="1.3"/>',
  github: '<path d="M9 19.5c-4 1.2-4-2-5.6-2.4M14.6 21.5v-3.2a2.8 2.8 0 0 0-.8-2.2c2.6-.3 5.3-1.3 5.3-5.8a4.5 4.5 0 0 0-1.2-3.1 4.2 4.2 0 0 0-.1-3.1s-1-.3-3.2 1.2a11 11 0 0 0-5.8 0C6.6 3.8 5.6 4.1 5.6 4.1a4.2 4.2 0 0 0-.1 3.1 4.5 4.5 0 0 0-1.2 3.1c0 4.5 2.7 5.5 5.3 5.8a2.8 2.8 0 0 0-.8 2.2v3.2"/>',
  info: '<circle cx="12" cy="12" r="8.4"/><path d="M12 11v5.2M12 7.9v.1"/>',
  lock: '<rect x="5" y="10.5" width="14" height="9.5" rx="2.4"/><path d="M8.2 10.5V8a3.8 3.8 0 0 1 7.6 0v2.5"/>',
  sun: '<circle cx="12" cy="12" r="3.8"/><path d="M12 3v1.8M12 19.2V21M3 12h1.8M19.2 12H21M5.6 5.6l1.3 1.3M17.1 17.1l1.3 1.3M5.6 18.4l1.3-1.3M17.1 6.9l1.3-1.3"/>',
  moon: '<path d="M19.5 14.6A7.8 7.8 0 1 1 9.4 4.5a6.2 6.2 0 0 0 10.1 10.1z"/>',
  clock: '<circle cx="12" cy="12" r="8.4"/><path d="M12 7.6V12l3 1.8"/>',
  tag: '<path d="M3.8 12.6V5.4a1.6 1.6 0 0 1 1.6-1.6h7.2l7.6 7.6a1.6 1.6 0 0 1 0 2.3l-6.9 6.9a1.6 1.6 0 0 1-2.3 0z"/><circle cx="8.4" cy="8.4" r="1.3"/>',
  check: '<path d="m5.5 12.5 4.2 4.2 8.8-9.2"/>',
  link: '<path d="M10 14a4 4 0 0 0 5.7 0l3-3a4 4 0 0 0-5.7-5.7l-1 1M14 10a4 4 0 0 0-5.7 0l-3 3a4 4 0 0 0 5.7 5.7l1-1"/>',
  cards: '<rect x="4" y="4.5" width="7" height="6.5" rx="1.6"/><rect x="13" y="4.5" width="7" height="6.5" rx="1.6"/><rect x="4" y="13" width="7" height="6.5" rx="1.6"/><rect x="13" y="13" width="7" height="6.5" rx="1.6"/>',
  clean: '<path d="M4 6.5h16M4 12h10M4 17.5h13"/>',
  arrowUp: '<path d="M12 19V5M6.5 10.5 12 5l5.5 5.5"/>',
  history: '<path d="M4.5 12a7.5 7.5 0 1 0 2.2-5.3L4.5 9"/><path d="M4.5 4.8V9h4.2M12 8v4.2l2.8 1.8"/>',
};

defineProps<{ name: string }>();
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <svg class="m-ic" viewBox="0 0 24 24" aria-hidden="true" v-html="P[name] ?? ''" />
</template>

<style lang="scss">
.m-ic {
  width: 24px;
  height: 24px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.5;
  stroke-linecap: round;
  stroke-linejoin: round;
  flex: none;

  .f { fill: transparent; transition: fill var(--dur) var(--ease-out); }
  .d { transition: stroke var(--dur) var(--ease-out); }

  &.s { width: 18px; height: 18px; }
  &.xs { width: 15px; height: 15px; }
}
</style>
