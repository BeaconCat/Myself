<script setup lang="ts">
import { computed } from 'vue';
import type { SkillsData } from '../types';
import type { ModProps } from './props';
import ModHead from '../parts/ModHead.vue';

/** 技能（skills）：分组的极轻描边胶囊；每组可标一个主技能（前置 4px 主色圆点） */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as SkillsData);
</script>

<template>
  <ModHead :title="title" />
  <div class="skl">
    <div v-for="(g, gi) in d.groups" :key="gi" class="skl-g">
      <h5>{{ g.title }}<small>{{ String(g.items.length).padStart(2, '0') }}</small></h5>
      <div>
        <span v-for="it in g.items" :key="it" class="ak-chip" :class="{ star: it === g.star }">{{ it }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.skl { display: flex; flex-direction: column; gap: 18px; }

.skl-g {
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

.ak-chip { cursor: default; }

@container (min-width: 560px) { .skl { display: grid; grid-template-columns: repeat(3, 1fr); gap: 26px; } }
</style>
