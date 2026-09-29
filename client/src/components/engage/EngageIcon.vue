<script setup lang="ts">
/** 互动线性图标（24 网格，1.6 线宽，与全站图标同一画法）；filled 时喜欢等图标实心 */
defineProps<{ name: string; filled?: boolean; size?: number }>();

const PATHS: Record<string, string> = {
  like: 'M12 20s-7-4.35-7-10a4 4 0 0 1 7-2.65A4 4 0 0 1 19 10c0 5.65-7 10-7 10Z',
  spark: 'M12 3.5l1.9 5.1a2 2 0 0 0 1.2 1.2L20.5 12l-5.4 1.9a2 2 0 0 0-1.2 1.2L12 20.5l-1.9-5.4a2 2 0 0 0-1.2-1.2L3.5 12l5.4-1.9a2 2 0 0 0 1.2-1.2Z',
  smile: 'M12 20.5a8.5 8.5 0 1 0 0-17 8.5 8.5 0 0 0 0 17ZM8.5 14c.9 1.3 2.1 2 3.5 2s2.6-.7 3.5-2M9 9.8h.01M15 9.8h.01',
  resonate: 'M12 13.2a1.2 1.2 0 1 0 0-2.4 1.2 1.2 0 0 0 0 2.4ZM8.3 8.3a5.2 5.2 0 0 0 0 7.4M15.7 8.3a5.2 5.2 0 0 1 0 7.4M5.5 5.5a9.2 9.2 0 0 0 0 13M18.5 5.5a9.2 9.2 0 0 1 0 13',
  comment: 'M20 12.5c0 4-3.6 7-8 7-1.1 0-2.2-.2-3.2-.6L4.5 20l1.2-3.4A6.8 6.8 0 0 1 4 12.5c0-4 3.6-7 8-7s8 3 8 7Z',
  react: 'M12 20.5a8.5 8.5 0 1 1 8.2-10.7M8.5 14c.9 1.3 2.1 2 3.5 2s2.6-.7 3.5-2M9 9.8h.01M15 9.8h.01M19 14v6M16 17h6',
  share: 'M12 15V3.5M7.5 8 12 3.5 16.5 8M5 12.5v6a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-6',
  copy: 'M9 9h10v10H9zM5 15V5h10',
  check: 'M5 12.5l4.5 4.5L19 7',
};
</script>

<template>
  <svg
    class="eg-i"
    :class="{ filled }"
    viewBox="0 0 24 24"
    :width="size ?? 18"
    :height="size ?? 18"
    fill="none"
    stroke="currentColor"
    stroke-width="1.6"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <path :d="PATHS[name] ?? ''" />
  </svg>
</template>

<style scoped>
.eg-i { display: block; flex: none; transition: transform var(--dur-fast) var(--ease-spring); }
.eg-i.filled path { fill: currentColor; }
</style>
