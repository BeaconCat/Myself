<script setup lang="ts">
import type { AboutModule } from '../../stores/config';

/** 技术栈卡片 */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <h2 class="block-title">技术栈</h2>
  <div class="stack">
    <div v-for="s in mod.data.items" :key="s.name" class="stack-item">
      <strong>{{ s.name }}</strong>
      <span>{{ s.role }}</span>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use './shared';

.stack {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 12px;
}

.stack-item {
  text-align: center;
  padding: 18px 10px;
  background: var(--surface);
  border: 1px solid var(--border);
  transition: transform var(--dur-fast) var(--ease-out), border-color var(--dur-fast);

  strong { display: block; font-family: var(--font-serif); font-size: 16px; }
  span { display: block; margin-top: 6px; font-size: 12px; color: var(--text-2); }
  &:hover { transform: scale(1.05); border-color: var(--primary); }
}
</style>
