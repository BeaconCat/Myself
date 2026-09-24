<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { MilestonesData } from '../types';
import type { ModProps } from './props';
import ModHead from '../parts/ModHead.vue';

/** 历程（milestones）：竖排（最新在上）/ 横排；标记 now 的一条点亮；窄容器横排自动回落竖排 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as MilestonesData);
const { t } = useI18n();

/** 未显式标记 now 时，最后一条点亮 */
const items = computed(() => {
  const anyNow = d.value.items.some((x) => x.now);
  return d.value.items.map((x, i) => ({ ...x, now: anyNow ? !!x.now : i === d.value.items.length - 1 }));
});
const split = (date: string) => {
  const [y, ...rest] = String(date).split(/[.\-/]/);
  return { y, m: rest.join('.') };
};
</script>

<template>
  <ModHead :title="title">{{ t('aboutKit.milestones.count', { n: items.length }) }}</ModHead>
  <ol v-if="variant === 'horizontal'" class="msh">
    <li v-for="(it, i) in items" :key="i" :class="{ now: it.now }" :style="{ '--k': i }">
      <time>{{ split(it.date).y }}<small v-if="split(it.date).m">.{{ split(it.date).m }}</small></time>
      <b>{{ it.title }}</b>
      <p v-if="it.text">{{ it.text }}</p>
    </li>
  </ol>
  <ol v-else class="ms">
    <li v-for="(it, i) in [...items].reverse()" :key="i" :class="{ now: it.now }" :style="{ '--k': i }">
      <time>{{ it.date }}</time>
      <b>{{ it.title }}</b>
      <p v-if="it.text">{{ it.text }}</p>
    </li>
  </ol>
</template>

<style scoped lang="scss">
.ms {
  position: relative;
  list-style: none;
  padding-left: 26px;

  &::before {
    content: '';
    position: absolute;
    left: 5px;
    top: 8px;
    bottom: 8px;
    width: 1px;
    background: linear-gradient(rgba(var(--primary-rgb), 0.8), var(--ak-line-2) 30%, var(--ak-line-2));
  }

  li { position: relative; padding-bottom: 20px; }
  li:last-child { padding-bottom: 0; }

  li::before {
    content: '';
    position: absolute;
    left: -25px;
    top: 7px;
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: var(--ak-surface);
    box-shadow: inset 0 0 0 1.5px var(--ak-text-3);
  }

  time { font: 500 11.5px var(--ak-mono); letter-spacing: 0.04em; color: var(--ak-text-3); }
  b { display: block; margin-top: 2px; font: 700 15px/1.5 var(--font-serif); }
  p { margin-top: 2px; font-size: 13px; color: var(--text-2); }
}

.ms li.now::before, .msh li.now::before {
  background: var(--primary);
  box-shadow: 0 0 0 4px rgba(var(--primary-rgb), 0.2), 0 0 16px rgba(var(--primary-rgb), 0.9);
}

.ms li.now time, .msh li.now time { color: var(--ak-ink); }

.msh {
  position: relative;
  list-style: none;
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: 1fr;
  padding-top: 30px;

  &::before {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    top: 34px;
    height: 1px;
    background: linear-gradient(90deg, var(--ak-line-2), var(--ak-line-2) 70%, rgba(var(--primary-rgb), 0.8));
  }

  li { position: relative; padding: 22px 18px 0 0; }

  li::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: var(--ak-surface);
    box-shadow: inset 0 0 0 1.5px var(--ak-text-3);
  }

  time {
    display: block;
    margin-bottom: 8px;
    font: 700 26px/1 var(--font-serif);
    letter-spacing: -0.02em;

    small { margin-left: 2px; font: 500 12px var(--ak-mono); letter-spacing: 0; color: var(--ak-text-3); }
  }

  b { display: block; font: 700 14.5px/1.5 var(--font-serif); }
  p { margin-top: 4px; font-size: 12.5px; color: var(--text-2); }
}

.ms li, .msh li { opacity: 0; transform: translateY(8px); transition: opacity 0.6s var(--ease-out), transform 0.6s var(--ease-out); transition-delay: calc(var(--k) * 80ms + 200ms); }
.in .ms li, .in .msh li { opacity: 1; transform: none; }

@container (max-width: 600px) {
  .msh { grid-auto-flow: row; padding: 0 0 0 26px; }
  .msh::before { left: 5px; right: auto; top: 8px; bottom: 8px; width: 1px; height: auto; background: var(--ak-line-2); }
  .msh li { padding: 0 0 18px; }
  .msh li::before { left: -25px; top: 6px; }
  .msh time { font-size: 18px; }
}
</style>
