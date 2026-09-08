<script setup lang="ts">
import type { AboutModule } from '../../stores/config';

/** 技能条形图：0–100 熟练度 */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <h2 class="block-title">能力图谱</h2>
  <div class="bars card-box">
    <div v-for="item in mod.data.items" :key="item.name" class="bar-row">
      <span class="bar-name">{{ item.name }}</span>
      <div class="bar-track">
        <div class="bar-fill" :style="{ width: `${Math.min(100, Math.max(0, item.level))}%` }" />
      </div>
      <span class="bar-val">{{ item.level }}</span>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use './shared';

.bars { display: flex; flex-direction: column; gap: 14px; }

.bar-row {
  display: flex;
  align-items: center;
  gap: 14px;
}

.bar-name { width: 90px; font-size: 13.5px; font-weight: 600; flex-shrink: 0; }

.bar-track {
  flex: 1;
  height: 10px;
  border-radius: 999px;
  background: var(--surface-2);
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--primary), var(--primary-deep));
  box-shadow: 0 0 8px rgba(var(--primary-rgb), 0.4);
  transition: width var(--dur-slow) var(--ease-out);
}

.bar-val {
  width: 34px;
  text-align: right;
  font-size: 12px;
  color: var(--text-2);
  font-variant-numeric: tabular-nums;
}
</style>
