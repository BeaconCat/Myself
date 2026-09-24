<script setup lang="ts">
import { computed } from 'vue';
import type { SkillsData } from '../types';
import type { ModProps } from './props';
import ModHead from '../parts/ModHead.vue';

/** 技能（skills）：分组键帽标签 —— 按下有真实行程；每组可标一个主技能高亮 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as SkillsData);
</script>

<template>
  <ModHead :title="title" />
  <div class="sk">
    <div v-for="(g, gi) in d.groups" :key="gi" class="sk-g">
      <h5>{{ g.title }}<small>{{ String(g.items.length).padStart(2, '0') }}</small></h5>
      <div>
        <span v-for="it in g.items" :key="it" class="key" :class="{ star: it === g.star }">{{ it }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.sk { display: flex; flex-direction: column; gap: 18px; }

.sk-g {
  h5 {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: 10px;
    font: 700 14px var(--font-serif);

    small { font: 400 11px var(--ak-mono); color: var(--ak-text-3); }
  }

  > div { display: flex; flex-wrap: wrap; gap: 6px; }
}

.key {
  display: inline-flex;
  align-items: center;
  height: 30px;
  padding: 0 12px;
  border-radius: 9px;
  font-size: 13px;
  background: var(--ak-surface-hi);
  border: 1px solid var(--ak-line-2);
  box-shadow: 0 2px 0 var(--ak-line-2);
  cursor: default;
  transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast), color var(--dur-fast), border-color var(--dur-fast);

  &:hover { transform: translateY(2px); box-shadow: 0 0 0 var(--ak-line-2); color: var(--ak-ink); border-color: rgba(var(--primary-rgb), 0.45); }
  &.star { color: var(--ak-ink); border-color: rgba(var(--primary-rgb), 0.35); background: rgba(var(--primary-rgb), 0.08); }
}

@container (min-width: 560px) { .sk { display: grid; grid-template-columns: repeat(3, 1fr); gap: 26px; } }
</style>
