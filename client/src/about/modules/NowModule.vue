<script setup lang="ts">
import type { AboutModule } from '../../stores/config';

/** 正在做（Now 页理念） */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <h2 class="block-title">现在</h2>
  <ul class="now card-box">
    <li v-for="(item, i) in mod.data.items" :key="i">
      <span class="pulse" aria-hidden="true" />{{ item }}
    </li>
  </ul>
</template>

<style scoped lang="scss">
@use './shared';

.now {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 12px;

  li {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 14.5px;
  }
}

.pulse {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: var(--primary);
  flex-shrink: 0;
  animation: pulse-glow 1.8s ease-in-out infinite;
}

@keyframes pulse-glow {
  0%, 100% { box-shadow: 0 0 0 0 rgba(var(--primary-rgb), 0.5); }
  50% { box-shadow: 0 0 0 6px rgba(var(--primary-rgb), 0); }
}
</style>
