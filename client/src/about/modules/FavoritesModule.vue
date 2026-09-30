<script setup lang="ts">
import { computed, ref } from 'vue';
import type { FavoritesData } from '../types';
import type { ModProps } from './props';
import ModHead from '../parts/ModHead.vue';

/** 喜好（favorites）：分组 Tab 切换，条目带作者 / 年份 / 一句评语，切换时逐条入场 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as FavoritesData);
const tab = ref(0);
const group = computed(() => d.value.groups[Math.min(tab.value, d.value.groups.length - 1)]);
</script>

<template>
  <ModHead :title="title" />
  <div v-if="d.groups.length > 1" class="ak-seg fv-tabs">
    <button v-for="(g, i) in d.groups" :key="i" :class="{ on: i === tab }" @click="tab = i">{{ g.title }}</button>
  </div>
  <ul v-if="group" :key="tab" class="fv-list">
    <li v-for="(it, k) in group.items" :key="k" :style="{ '--k': k }">
      <span class="n">{{ String(k + 1).padStart(2, '0') }}</span>
      <div>
        <b>{{ it.name }}</b>
        <span v-if="it.by || it.note">{{ [it.by, it.note].filter(Boolean).join(' · ') }}</span>
      </div>
      <time>{{ it.year }}</time>
    </li>
  </ul>
</template>

<style scoped lang="scss">
.fv-tabs { align-self: flex-start; margin-bottom: 12px; }

/* 条目按正常间距自上而下排列，不撑满模块高度（撑满会在少量条目时拉得很散） */
.fv-list {
  list-style: none;
  display: flex;
  flex-direction: column;

  li {
    display: grid;
    grid-template-columns: 24px minmax(0, 1fr) auto;
    gap: 10px;
    align-items: baseline;
    padding: 12px 0;
    border-top: 1px solid var(--ak-line);
    animation: ak-fade-up 0.45s var(--ease-out) both;
    animation-delay: calc(var(--k) * 50ms);
  }

  li:first-child { border-top: 0; }

  .n { font: 400 12.5px var(--ak-mono); color: var(--ak-text-3); }
  b { font: 700 16px var(--font-serif); }
  div span { display: block; margin-top: 2px; font-size: 13px; color: var(--ak-text-3); }
  time { font: 400 12.5px var(--ak-mono); color: var(--ak-text-3); }
}
</style>
