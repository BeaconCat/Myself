<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { SkillbarsData } from '../types';
import type { ModProps } from './props';
import ModHead from '../parts/ModHead.vue';

/** 技能刻度（skillbars）：20 格刻度尺逐格点亮、末格发光；dots 变体为 10 点阵（支持半点） */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as SkillbarsData);
const { t } = useI18n();

const clamp = (n: number) => Math.max(0, Math.min(100, Number(n) || 0));
const tier = (level: number, custom?: string) =>
  custom || t(level >= 80 ? 'aboutKit.skillbars.expert' : level >= 60 ? 'aboutKit.skillbars.skilled' : 'aboutKit.skillbars.novice');

const rows = computed(() =>
  d.value.items.map((it) => {
    const level = clamp(it.level);
    return { name: it.name, level, tier: tier(level, it.tier), ticks: Math.round(level / 5), dots: level / 10 };
  }),
);
</script>

<template>
  <ModHead :title="title">0 — 100</ModHead>
  <ul class="sb">
    <li v-for="r in rows" :key="r.name" class="sb-row">
      <span class="nm">{{ r.name }}<span v-if="variant !== 'dots'" class="sb-tier">{{ r.tier }}</span></span>
      <div v-if="variant === 'dots'" class="dots">
        <i
          v-for="k in 10"
          :key="k"
          :style="{ '--k': k - 1 }"
          :class="{ f: k <= r.dots, h: k > r.dots && k - 0.5 <= r.dots }"
        />
      </div>
      <div v-else class="ruler">
        <i
          v-for="k in 20"
          :key="k"
          :style="{ '--k': k - 1 }"
          :class="{ f: k <= r.ticks, tip: k === r.ticks }"
        />
      </div>
      <b>{{ r.level }}</b>
    </li>
  </ul>
  <div v-if="variant !== 'dots'" class="sb-scale">
    <span /><div><span>0</span><span>25</span><span>50</span><span>75</span><span>100</span></div><span />
  </div>
</template>

<style scoped lang="scss">
.sb { list-style: none; display: flex; flex-direction: column; gap: 16px; }

.sb-row {
  display: grid;
  grid-template-columns: 56px minmax(0, 1fr) 34px;
  gap: 14px;
  align-items: center;

  .nm { font-size: 13.5px; line-height: 1.35; }
  > b { font: 500 12.5px var(--ak-mono); text-align: right; color: var(--text-2); }
}

.sb-tier { display: block; font: 400 10.5px var(--ak-mono); color: var(--ak-text-3); }

.ruler {
  position: relative;
  display: grid;
  grid-template-columns: repeat(20, 1fr);
  align-items: end;
  height: 22px;

  i {
    justify-self: center;
    width: 2px;
    height: 9px;
    border-radius: 1px;
    background: var(--ak-line-2);
    transition: background 0.25s var(--ease-out), height 0.3s var(--ease-spring), box-shadow 0.3s;
    transition-delay: calc(var(--k) * 30ms + 200ms);

    &:nth-child(5n) { height: 14px; }
  }
}

.in .ruler i.f { background: var(--primary); height: 14px; }
.in .ruler i.f:nth-child(5n) { height: 19px; }
.in .ruler i.tip { height: 22px !important; box-shadow: 0 0 10px rgba(var(--primary-rgb), 0.9); }

.sb-scale {
  display: grid;
  grid-template-columns: 56px minmax(0, 1fr) 34px;
  gap: 14px;
  margin-top: 6px;
  font: 400 10px var(--ak-mono);
  color: var(--ak-text-3);

  div { display: flex; justify-content: space-between; }
}

.dots {
  display: grid;
  grid-template-columns: repeat(10, minmax(0, 15px));
  justify-content: space-between;
  gap: 5px;

  i {
    aspect-ratio: 1;
    border-radius: 50%;
    background: var(--ak-sunken);
    box-shadow: inset 0 0 0 1px var(--ak-line-2);
    transition: background 0.3s, transform 0.4s var(--ease-spring);
    transition-delay: calc(var(--k) * 50ms + 200ms);
  }
}

.in .dots i.f { background: var(--primary); box-shadow: none; }
.in .dots i.h { background: linear-gradient(90deg, var(--primary) 50%, var(--ak-sunken) 50%); }
</style>
