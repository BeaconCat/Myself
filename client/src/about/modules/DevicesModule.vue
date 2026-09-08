<script setup lang="ts">
import type { AboutModule } from '../../stores/config';

/** 装备清单 */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <h2 class="block-title">装备</h2>
  <div class="devices">
    <div v-for="(d, i) in mod.data.items" :key="i" class="device card-box">
      <strong>{{ d.name }}</strong>
      <span>{{ d.desc }}</span>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use './shared';

.devices {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

.device {
  display: flex;
  flex-direction: column;
  gap: 6px;
  transition: transform var(--dur-fast) var(--ease-out), border-color var(--dur-fast);

  strong { font-size: 15px; }
  span { font-size: 12.5px; color: var(--text-2); }
  &:hover { transform: scale(1.03); border-color: var(--primary); }
}
</style>
