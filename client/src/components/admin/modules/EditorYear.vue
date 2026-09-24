<script setup lang="ts">
/* eslint-disable vue/no-mutating-props */
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { YearData, YearEntry } from '../../../about/types';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 年度回顾：多年份切换；每年 四个大数 / 12 个月数值（留空 = 未到）/ 大事月份 / 高光 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<YearData>(() => props.mod);
const { t } = useI18n();

const years = computed(() => Object.keys(d.value.years).sort().reverse());
const picked = ref('');
const year = computed(() => (years.value.includes(picked.value) ? picked.value : years.value[0] ?? ''));
const Y = computed<YearEntry | undefined>(() => d.value.years[year.value]);

function addYear(): void {
  const next = String(Math.max(new Date().getFullYear(), ...years.value.map(Number).filter(Boolean)) + (years.value.includes(String(new Date().getFullYear())) ? 1 : 0));
  d.value.years[next] = {
    sub: '',
    nums: [['文章', 0], ['随想', 0], ['提交', 0], ['读完', 0]],
    months: Array.from({ length: 12 }, () => null),
    pins: [],
    highlights: [],
  };
  picked.value = next;
}

function removeYear(): void {
  if (!year.value) return;
  delete d.value.years[year.value];
  picked.value = '';
}

function setMonth(i: number, v: string): void {
  if (!Y.value) return;
  Y.value.months[i] = v.trim() === '' ? null : Number(v);
}

function togglePin(i: number): void {
  if (!Y.value) return;
  const s = new Set(Y.value.pins);
  if (s.has(i)) s.delete(i);
  else s.add(i);
  Y.value.pins = [...s].sort((a, b) => a - b);
}
</script>

<template>
  <div class="ed">
    <div class="grid2">
      <label><span>{{ t('aboutKit.ed.metric') }}</span><input v-model="d.metric" class="a-input" type="text" /></label>
    </div>
    <div class="line">
      <div class="seg">
        <button v-for="y in years" :key="y" type="button" :class="{ on: y === year }" @click="picked = y">{{ y }}</button>
      </div>
      <button type="button" class="a-btn ghost sm" @click="addYear">{{ t('aboutKit.ed.addYear') }}</button>
      <button v-if="year" type="button" class="a-btn ghost sm danger" @click="removeYear">{{ t('aboutKit.ed.removeYear', { y: year }) }}</button>
    </div>

    <template v-if="Y">
      <label><span>{{ t('aboutKit.ed.yearSub') }}</span><input v-model="Y.sub" class="a-input" type="text" /></label>
      <div class="sub-title">{{ t('aboutKit.ed.nums') }}</div>
      <div class="grid4">
        <div v-for="(pair, i) in Y.nums" :key="i" class="line">
          <input v-model="pair[0]" class="a-input flex-in" type="text" />
          <input v-model.number="pair[1]" class="a-input w-num" type="number" min="0" />
        </div>
      </div>
      <div class="sub-title">{{ t('aboutKit.ed.months') }}</div>
      <div class="months">
        <div v-for="(v, i) in Y.months" :key="i" class="mo">
          <span>{{ t('aboutKit.github.month', { m: i + 1 }) }}</span>
          <input class="a-input" type="number" min="0" :value="v ?? ''" :placeholder="t('aboutKit.ed.future')" @change="setMonth(i, ($event.target as HTMLInputElement).value)" />
          <button type="button" class="pin" :class="{ on: Y.pins.includes(i) }" :title="t('aboutKit.ed.pin')" @click="togglePin(i)" />
        </div>
      </div>
      <p class="hint">{{ t('aboutKit.ed.monthsHint') }}</p>
      <div class="sub-title">{{ t('aboutKit.ed.highlights') }}</div>
      <EdList v-slot="{ item }" :items="Y.highlights" :make="() => ['', ''] as [string, string]" compact>
        <div class="line">
          <input v-model="item[0]" class="a-input w-num" type="text" :placeholder="t('aboutKit.ed.month')" />
          <input v-model="item[1]" class="a-input flex-in" type="text" />
        </div>
      </EdList>
    </template>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.danger { color: var(--accent-red) !important; }

.months { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 8px; }

.mo {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;

  span { font-size: 11px; color: var(--text-2); }
  .a-input { padding: 6px 8px; text-align: center; }
}

.pin {
  width: 12px;
  height: 12px;
  padding: 0;
  border: 1.5px solid color-mix(in oklab, var(--text) 25%, transparent);
  border-radius: 50%;
  background: none;

  &.on { background: #ffb300; border-color: #ffb300; box-shadow: 0 0 6px #ffb300; }
}

@media (max-width: 720px) { .months { grid-template-columns: repeat(4, minmax(0, 1fr)); } }
</style>
