<script setup lang="ts">
import type { AboutModule } from '../../stores/config';

/** 语言占比：堆叠比例条 + 图例 */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <h2 class="block-title">占比</h2>
  <div class="langs card-box">
    <div class="lang-bar">
      <div
        v-for="item in mod.data.items"
        :key="item.name"
        class="lang-seg"
        :style="{ width: `${item.percent}%`, background: item.color }"
        :title="`${item.name} ${item.percent}%`"
      />
    </div>
    <div class="lang-legend">
      <span v-for="item in mod.data.items" :key="item.name">
        <i :style="{ background: item.color }" />{{ item.name }} {{ item.percent }}%
      </span>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use './shared';

.lang-bar {
  display: flex;
  height: 16px;
  border-radius: 999px;
  overflow: hidden;
  margin-bottom: 14px;
}

.lang-seg { transition: width var(--dur-slow) var(--ease-out); }

.lang-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  font-size: 12.5px;
  color: var(--text-2);

  i {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 3px;
    margin-right: 6px;
    vertical-align: -1px;
  }
}
</style>
