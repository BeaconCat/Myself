<script setup lang="ts">
import type { AboutModule } from '../../stores/config';

/** 历程时间线 */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <h2 class="block-title">历程</h2>
  <div class="milestones">
    <div v-for="(m, i) in mod.data.items" :key="i" class="milestone">
      <span class="year">{{ m.year }}</span>
      <span class="dot" aria-hidden="true" />
      <p>{{ m.text }}</p>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use './shared';

.milestones { display: flex; flex-direction: column; }

.milestone {
  display: flex;
  align-items: baseline;
  gap: 18px;
  padding: 16px 4px;
  border-bottom: 1px solid var(--border);
  transition: background var(--dur-fast);

  &:hover {
    background: rgba(var(--primary-rgb), 0.04);
    .dot { transform: scale(1.4); box-shadow: 0 0 10px rgba(var(--primary-rgb), 0.6); }
  }

  .year {
    font-family: var(--font-serif);
    font-size: 20px;
    font-weight: 700;
    color: var(--primary);
    flex-shrink: 0;
    width: 72px;
  }

  .dot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: var(--primary);
    flex-shrink: 0;
    align-self: center;
    transition: all var(--dur-fast) var(--ease-spring);
  }

  p { font-size: 14.5px; line-height: 1.8; color: var(--text-2); }
}
</style>
