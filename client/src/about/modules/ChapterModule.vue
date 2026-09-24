<script setup lang="ts">
import { computed } from 'vue';
import type { ChapterData } from '../types';
import type { ModProps } from './props';

/** 章节（chapter）：编号 + 渐显细线 + 宋体大标题 + 副题，负责页面节奏 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as ChapterData);
</script>

<template>
  <div class="ch">
    <div class="ch-top">
      <span class="ch-no">{{ d.no || no }}</span>
      <span class="ch-rule" />
      <span v-if="d.meta" class="ch-meta">{{ d.meta }}</span>
    </div>
    <div class="ch-row">
      <h2>{{ d.title }}</h2>
      <p v-if="d.subtitle">{{ d.subtitle }}</p>
    </div>
  </div>
</template>

<style scoped lang="scss">
.ch { padding: 84px 0 4px; }

.ch-top { display: flex; align-items: center; gap: 16px; margin-bottom: 16px; }
.ch-no { font: 500 12px/1 var(--ak-mono); letter-spacing: 0.12em; color: var(--ak-ink); }

.ch-rule {
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, var(--ak-line-2), var(--ak-line) 60%, transparent);
  transform: scaleX(0);
  transform-origin: left;
  transition: transform 1.4s 0.2s var(--ease-out);
}

.in .ch-rule { transform: none; }

.ch-meta { font: 400 11.5px var(--ak-mono); color: var(--ak-text-3); }

.ch-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 24px;

  h2 { font: 700 38px/1.1 var(--font-serif); letter-spacing: 0.02em; }
  p { font-size: 14px; color: var(--ak-text-3); }
}

@container (max-width: 560px) {
  .ch { padding: 40px 0 0; }
  .ch-row { flex-direction: column; gap: 8px; }
  .ch-row h2 { font-size: 30px; }
}
</style>
