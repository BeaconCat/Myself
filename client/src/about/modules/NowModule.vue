<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { NowData } from '../types';
import type { ModProps } from './props';
import ModHead from '../parts/ModHead.vue';
import { ago, useClock } from '../useClock';

/** 此刻（now）：类别 + 一句话 + 注释 + 可选进度；底部显示更新时间 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as NowData);
const { t } = useI18n();
const now = useClock();
</script>

<template>
  <ModHead :title="title">
    <template v-if="d.updatedAt">{{ t('aboutKit.now.updated', { when: ago(d.updatedAt, t, now) }) }}</template>
  </ModHead>
  <ul class="nw" :class="{ cols: variant === 'cols' }">
    <li v-for="(it, i) in d.items" :key="i">
      <span class="k">{{ it.kind }}</span>
      <div class="v">
        {{ it.text }}
        <small v-if="it.note">{{ it.note }}</small>
        <span v-if="typeof it.progress === 'number'" class="prog"><span class="tr"><i :style="{ '--w': `${it.progress}%`, '--k': i }" /></span><b>{{ it.progress }}%</b></span>
      </div>
    </li>
  </ul>
  <div v-if="d.updatedAt" class="nw-foot"><span>/now</span><span>{{ d.updatedAt.replaceAll('-', '.') }}</span></div>
</template>

<style scoped lang="scss">
.nw { list-style: none; display: flex; flex-direction: column; margin-bottom: 16px; }

.nw li {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr);
  gap: 12px;
  align-items: baseline;
  padding: 13px 0;
  border-top: 1px solid var(--ak-line);

  &:first-child { border-top: 0; padding-top: 0; }
}

.k { font: 500 13px var(--font-sans); color: var(--ak-ink); }

.v {
  font-size: 15px;
  font-weight: 500;
  line-height: 1.55;

  small { display: block; margin-top: 3px; font-size: 13px; font-weight: 400; color: var(--ak-text-3); }
}

/* 进度：撑满正文宽度 + 右侧等宽百分比 */
.prog {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 9px;

  .tr { flex: 1; height: 5px; overflow: hidden; border-radius: var(--r-pill); background: var(--fill-2); }
  i { display: block; width: 0; height: 100%; border-radius: inherit; background: var(--ink); transition: width 1.2s var(--ease-out); transition-delay: calc(var(--k) * 90ms + 250ms); }
  b { font: 500 12.5px var(--ak-mono); color: var(--text-2); font-variant-numeric: tabular-nums; }
}

.in .prog i { width: var(--w); }

.nw-foot {
  display: flex;
  justify-content: space-between;
  margin-top: auto;
  padding-top: 12px;
  border-top: 1px solid var(--ak-line);
  font: 400 12.5px var(--ak-mono);
  color: var(--ak-text-3);
}

.nw.cols {
  display: grid;
  grid-template-columns: 1fr 1fr;
  column-gap: 28px;

  li:nth-child(2) { border-top: 0; padding-top: 0; }
}

@container (max-width: 520px) {
  .nw.cols { grid-template-columns: 1fr; }
  .nw.cols li:nth-child(2) { border-top: 1px solid var(--ak-line); padding-top: 13px; }
}
</style>
