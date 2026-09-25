<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import ContentIcon from '../post/ContentIcon.vue';
import { isoDay } from '../post/content';

/**
 * 月历（周一起始）：有随想的日子标 3px 小点；选中 = 抬升 + 轻染（点转 --ink）；今天细描边。
 * 区间模式下起止之间铺一层 --fill-2 连续底。
 */
const props = withDefaults(
  defineProps<{
    marks: Set<string>;
    from?: string;
    to?: string;
    /** 区间模式：已点起点、等待终点 */
    pending?: string;
    /** 无随想的日子不可点（右栏单日筛选） */
    onlyMarked?: boolean;
  }>(),
  { from: '', to: '', pending: '', onlyMarked: false },
);
const month = defineModel<string>('month', { required: true });
const emit = defineEmits<{ pick: [day: string] }>();
const { t } = useI18n();

const today = isoDay(new Date());
const weekdays = computed(() => t('content.thoughts.weekdays').split(','));

const title = computed(() => {
  const [y, m] = month.value.split('-').map(Number);
  return t('content.thoughts.monthTitle', { y, m });
});

const canNext = computed(() => month.value < today.slice(0, 7));

interface Cell { key: string; day: number; iso: string }
const cells = computed<Cell[]>(() => {
  const [y, m] = month.value.split('-').map(Number);
  const lead = (new Date(y, m - 1, 1).getDay() + 6) % 7;
  const last = new Date(y, m, 0).getDate();
  const out: Cell[] = [];
  for (let i = 0; i < lead; i++) out.push({ key: `b${i}`, day: 0, iso: '' });
  for (let d = 1; d <= last; d++) {
    out.push({ key: `d${d}`, day: d, iso: `${month.value}-${String(d).padStart(2, '0')}` });
  }
  return out;
});

function shift(delta: number): void {
  const [y, m] = month.value.split('-').map(Number);
  const d = new Date(y, m - 1 + delta, 1);
  month.value = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
}

function cls(c: Cell) {
  const lo = props.pending || props.from;
  const hi = props.pending ? '' : props.to;
  const edge = c.iso === lo || c.iso === hi;
  return {
    has: props.marks.has(c.iso),
    on: edge,
    today: c.iso === today,
    mid: !!lo && !!hi && lo !== hi && c.iso > lo && c.iso < hi,
    start: !!hi && lo !== hi && c.iso === lo,
    end: !!hi && lo !== hi && c.iso === hi,
  };
}

function disabled(c: Cell): boolean {
  if (c.iso > today) return true;
  return props.onlyMarked && !props.marks.has(c.iso);
}
</script>

<template>
  <div class="mcal">
    <div class="cal-h">
      <b>{{ title }}</b>
      <div class="nv">
        <button type="button" :aria-label="t('content.thoughts.prevMonth')" @click="shift(-1)"><ContentIcon name="chevL" size="xs" /></button>
        <button type="button" :aria-label="t('content.thoughts.nextMonth')" :disabled="!canNext" @click="shift(1)"><ContentIcon name="chevR" size="xs" /></button>
      </div>
    </div>
    <div class="cal">
      <span v-for="w in weekdays" :key="w" class="w">{{ w }}</span>
      <template v-for="c in cells" :key="c.key">
        <span v-if="!c.day" />
        <button v-else type="button" :class="cls(c)" :disabled="disabled(c)" @click="emit('pick', c.iso)">{{ c.day }}</button>
      </template>
    </div>
  </div>
</template>

<style scoped lang="scss">
.cal-h {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;

  b { font-size: 14px; font-weight: 500; color: var(--text); }

  .nv { display: flex; gap: 2px; }

  button {
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    border: 0;
    border-radius: 50%;
    background: none;
    color: var(--text-2);
    transition: background-color var(--dur-fast), color var(--dur-fast);

    &:hover { background: var(--fill-2); color: var(--text); }
    &:disabled { opacity: 0.35; pointer-events: none; }
    &:focus-visible { outline: none; box-shadow: var(--focus); }
  }
}

.cal {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  row-gap: 2px;
  text-align: center;

  .w {
    padding-bottom: 6px;
    font-size: 11px;
    color: var(--text-3);
  }

  button {
    position: relative;
    height: 34px;
    border: 0;
    border-radius: var(--r-sm);
    background: none;
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-3);
    transition: background-color var(--dur-fast), color var(--dur-fast), box-shadow var(--dur-fast);

    &.has { color: var(--text); }

    /* 有随想：底部 3px 小点 */
    &.has::after {
      content: '';
      position: absolute;
      left: 50%;
      bottom: 4px;
      width: 3px;
      height: 3px;
      margin-left: -1.5px;
      border-radius: 50%;
      background: var(--text-3);
    }

    &:hover { background: var(--fill-2); color: var(--text); }
    &.today { box-shadow: inset 0 0 0 1px var(--line-2); }

    /* 区间中段：连续浅底（去掉圆角让它连成条） */
    &.mid { border-radius: 0; background: var(--fill-2); color: var(--text); }
    &.start { border-top-right-radius: 0; border-bottom-right-radius: 0; }
    &.end { border-top-left-radius: 0; border-bottom-left-radius: 0; }

    &.on {
      background: var(--lift);
      color: var(--lift-fg);
      box-shadow: var(--lift-shadow);

      &::after { background: var(--ink); }
    }

    &:focus-visible { outline: none; box-shadow: var(--focus); }
    &:disabled { opacity: 0.35; pointer-events: none; }
  }
}
</style>
