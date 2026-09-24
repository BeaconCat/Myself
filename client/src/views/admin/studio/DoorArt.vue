<script setup lang="ts">
/** 空状态插画：一扇虚掩的门，门缝下透出主色的光 */
const uid = `dg-${Math.random().toString(36).slice(2, 8)}`;
</script>

<template>
  <svg class="door-art" viewBox="0 0 96 96" width="96" height="96" fill="none" aria-hidden="true">
    <defs>
      <linearGradient :id="uid" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0" stop-color="var(--primary)" stop-opacity=".35" />
        <stop offset="1" stop-color="var(--primary)" stop-opacity="0" />
      </linearGradient>
    </defs>
    <path class="beam" d="M38 70 L58 70 L80 92 L16 92Z" :fill="`url(#${uid})`" />
    <rect x="30" y="14" width="36" height="56" rx="4" stroke="var(--ink-4)" stroke-width="1.5" />
    <path d="M48 14v56" stroke="var(--line-3)" stroke-width="1.5" />
    <rect class="slit" x="44" y="18" width="8" height="52" fill="var(--primary)" opacity=".18" />
    <path d="M20 70h56" stroke="var(--ink-4)" stroke-width="1.5" stroke-linecap="round" />
  </svg>
</template>

<style scoped>
.door-art .beam { transform-origin: 48px 70px; animation: beam 3.6s var(--ease-out) infinite alternate; }
.door-art .slit { animation: slit 3.6s var(--ease-out) infinite alternate; }

@keyframes beam { from { opacity: 0.55; transform: scaleX(0.86); } }
@keyframes slit { from { opacity: 0.1; } to { opacity: 0.26; } }
</style>
