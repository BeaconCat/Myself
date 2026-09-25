<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { StackData } from '../types';
import type { ModProps } from './props';
import ModHead from '../parts/ModHead.vue';

/** 技术栈（stack）：等宽字母徽标 + 名称 + 角色；品牌色只作 13% 淡底，不抢主色 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as StackData);
const { t } = useI18n();
</script>

<template>
  <ModHead :title="title">{{ t('aboutKit.stack.meta') }}</ModHead>
  <div class="sx">
    <component
      :is="s.url ? 'a' : 'div'"
      v-for="s in d.items"
      :key="s.name"
      :href="s.url || undefined"
      :target="s.url ? '_blank' : undefined"
      :rel="s.url ? 'noopener' : undefined"
      class="it"
    >
      <span class="g" :style="{ '--c': s.color || 'var(--primary)' }">{{ s.glyph || s.name.slice(0, 2) }}</span>
      <span class="tx"><b>{{ s.name }}</b><span>{{ s.role }}</span></span>
    </component>
  </div>
</template>

<style scoped lang="scss">
.sx { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; }

.it {
  display: flex;
  align-items: center;
  gap: 11px;
  min-width: 0;
  padding: 10px;
  border-radius: var(--r-md);
  transition: background-color var(--dur-fast), box-shadow var(--dur-fast);

  &:hover { background: var(--ak-sunken); box-shadow: inset 0 0 0 1px var(--ak-line); }
}

.g {
  display: grid;
  place-items: center;
  flex: none;
  width: 36px;
  height: 36px;
  border-radius: var(--r-sm);
  font: 600 12.5px var(--ak-mono);
  /* 字形保留一点品牌色相，底与描边中性 */
  color: color-mix(in oklab, var(--c) 60%, var(--text));
  background: var(--fill);
  box-shadow: inset 0 0 0 1px var(--ak-line);
}

.tx {
  min-width: 0;

  b { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 13.5px; font-weight: 500; line-height: 1.3; }
  span { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 11.5px; color: var(--ak-text-3); }
}

@container (min-width: 560px) { .sx { grid-template-columns: repeat(4, minmax(0, 1fr)); } }
</style>
