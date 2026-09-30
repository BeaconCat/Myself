<script setup lang="ts">
import { safeHref } from '../../utils/safeUrl';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { StackData } from '../types';
import type { ModProps } from './props';
import ModHead from '../parts/ModHead.vue';
import TechIcon from '../parts/TechIcon.vue';
import { TECH, techInk } from '../tech';

/**
 * 技术栈（stack）：官方品牌图标（simple-icons，按品牌色着色）或等宽字母徽标 + 名称 + 角色。
 * 徽标底与描边保持中性；近黑 / 近白的品牌色改用正文色，保证明暗两种主题都可读。
 */
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
      :href="safeHref(s.url)"
      :target="s.url ? '_blank' : undefined"
      :rel="s.url ? 'noopener' : undefined"
      class="it"
    >
      <span class="g" :class="{ ic: TECH[s.icon ?? ''] }" :style="{ '--c': techInk(s.color) || 'var(--primary)' }">
        <TechIcon v-if="TECH[s.icon ?? '']" :name="s.icon" :size="20" />
        <template v-else>{{ s.glyph || s.name.slice(0, 2) }}</template>
      </span>
      <span class="tx"><b>{{ s.name }}</b><span>{{ s.role }}</span></span>
    </component>
  </div>
</template>

<style scoped lang="scss">
.sx { display: grid; grid-template-columns: minmax(0, 1fr); align-content: space-between; flex: 1; gap: 4px; }

.it {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  padding: 6px 10px;
  margin: 0 -10px;
  border-radius: var(--r-md);
  transition: background-color var(--dur-fast);

  &:hover { background: var(--ak-sunken); }
}

.g {
  display: grid;
  place-items: center;
  flex: none;
  width: 40px;
  height: 40px;
  border-radius: var(--r-sm);
  font: 600 14px var(--ak-mono);
  /* 字形保留一点品牌色相，底与描边中性 */
  color: color-mix(in oklab, var(--c) 60%, var(--text));
  background: var(--fill);

  &.ic { color: var(--c); }
}

.tx {
  min-width: 0;

  b { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 15px; font-weight: 500; line-height: 1.35; }
  span { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 13px; color: var(--ak-text-3); }
}

/* 窄模块（4 栏）单列铺满高度；中等宽度两列；宽模块四列 */
@container (min-width: 420px) { .sx { grid-template-columns: repeat(2, minmax(0, 1fr)); column-gap: 28px; } }
@container (min-width: 680px) { .sx { grid-template-columns: repeat(4, minmax(0, 1fr)); } }
</style>
