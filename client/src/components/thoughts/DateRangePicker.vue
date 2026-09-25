<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import ContentIcon from '../post/ContentIcon.vue';
import { isoDay, ymdOf } from '../post/content';
import MonthCalendar from './MonthCalendar.vue';

/**
 * 日期筛选：按钮显示当前范围（无筛选 =「全部日期」；有筛选 = 抬升 + 轻染，尾部 × 一键清除）；
 * 弹层 = 快捷区间 chips + 月历区间点选（先点起点再点终点，同一天点两次即单日）。
 */
const props = defineProps<{ from: string; to: string; marks: Set<string>; /** 打开时默认显示的月份（YYYY-MM） */ defaultMonth?: string }>();
const emit = defineEmits<{
  apply: [from: string, to: string];
  quick: [days: number];
  clear: [];
  /** 月历翻到某月：父级据此补取该月「有随想的日子」 */
  month: [ym: string];
}>();
const { t } = useI18n();

const open = ref(false);
const pending = ref('');
const month = ref((props.to || isoDay(new Date())).slice(0, 7));
const root = ref<HTMLElement | null>(null);

const has = computed(() => !!(props.from || props.to));

function md(s: string): string {
  const { m, d } = ymdOf(s);
  return `${m}月${d}日`;
}

const label = computed(() => {
  if (!has.value) return t('content.thoughts.allDates');
  if (props.from === props.to || !props.to) return md(props.from || props.to);
  return `${md(props.from)} – ${md(props.to)}`;
});

/** 快捷区间：当前是否正好等于「最近 n 天」 */
function quickOn(days: number): boolean {
  if (!has.value) return false;
  const end = isoDay(new Date());
  const start = isoDay(new Date(Date.now() - (days - 1) * 864e5));
  return props.from === start && props.to === end;
}

function toggle(): void {
  open.value = !open.value;
  pending.value = '';
  if (open.value) month.value = (props.to || props.defaultMonth || isoDay(new Date())).slice(0, 7);
}

function pick(day: string): void {
  if (!pending.value) {
    pending.value = day;
    return;
  }
  const [a, b] = pending.value <= day ? [pending.value, day] : [day, pending.value];
  pending.value = '';
  open.value = false;
  emit('apply', a, b);
}

function quick(days: number): void {
  open.value = false;
  emit('quick', days);
}

function clear(): void {
  open.value = false;
  pending.value = '';
  emit('clear');
}

watch(month, (ym) => emit('month', ym), { immediate: true });

function onDoc(e: MouseEvent): void {
  if (open.value && root.value && !root.value.contains(e.target as Node)) open.value = false;
}

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape') open.value = false;
}

onMounted(() => {
  document.addEventListener('pointerdown', onDoc);
  window.addEventListener('keydown', onKey);
});
onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onDoc);
  window.removeEventListener('keydown', onKey);
});
</script>

<template>
  <div ref="root" class="drp">
    <button type="button" class="datebtn" :class="{ has, open }" :aria-expanded="open" @click="toggle">
      <ContentIcon name="cal" size="s" />
      <span>{{ label }}</span>
      <span v-if="has" class="x" role="button" :aria-label="t('content.thoughts.clearDate')" @click.stop="clear">
        <ContentIcon name="x" size="xs" />
      </span>
    </button>

    <Transition name="pop">
      <div v-if="open" class="pop" role="dialog">
        <div class="quick">
          <button type="button" class="qc" :class="{ on: quickOn(1) }" @click="quick(1)">{{ t('thoughts.today') }}</button>
          <button type="button" class="qc" :class="{ on: quickOn(7) }" @click="quick(7)">{{ t('thoughts.last7') }}</button>
          <button type="button" class="qc" :class="{ on: quickOn(30) }" @click="quick(30)">{{ t('thoughts.last30') }}</button>
        </div>
        <MonthCalendar
          v-model:month="month"
          :marks="marks"
          :from="from"
          :to="to"
          :pending="pending"
          @pick="pick"
        />
        <div class="foot">
          <span>{{ pending ? md(pending) + ' –' : t('content.thoughts.rangeHint') }}</span>
          <button v-if="has || pending" type="button" class="clr" @click="clear">{{ t('content.thoughts.clearDate') }}</button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped lang="scss">
.drp { position: relative; }

.datebtn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 40px;
  padding: 0 14px;
  border: 0;
  border-radius: var(--r-md);
  background: var(--fill);
  box-shadow: inset 0 0 0 1px var(--line);
  font-size: 13.5px;
  white-space: nowrap;
  color: var(--text-2);
  transition: background-color var(--dur-fast), color var(--dur-fast), box-shadow var(--dur-fast),
    transform var(--dur-fast) var(--ease-spring);

  &:hover, &.open { background: var(--fill-2); color: var(--text); }
  &:active { transform: scale(0.97); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }

  &.has {
    background: var(--lift);
    color: var(--lift-fg);
    box-shadow: var(--lift-shadow);
  }

  .x {
    display: grid;
    place-items: center;
    width: 20px;
    height: 20px;
    margin-right: -6px;
    border-radius: 50%;
    color: var(--text-3);
    transition: background-color var(--dur-fast), color var(--dur-fast);

    &:hover { background: var(--fill-2); color: var(--text); }
  }
}

:root[data-mode='light'] .datebtn:not(.has) { background: color-mix(in oklab, var(--bg) 40%, white); }

.pop {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  z-index: 80;
  width: 300px;
  padding: 16px;
  border-radius: var(--r-lg);
  background: color-mix(in oklab, var(--surface) 92%, transparent);
  backdrop-filter: blur(24px) saturate(170%);
  -webkit-backdrop-filter: blur(24px) saturate(170%);
  box-shadow: var(--shadow-pop);
  transform-origin: top right;
}

:root[data-mode='light'] .pop { background: rgb(255 255 255 / 0.94); }

.pop-enter-active { transition: opacity 0.2s var(--ease-out), transform 0.35s var(--ease-spring); }
.pop-leave-active { transition: opacity 0.16s var(--ease-out), transform 0.2s var(--ease-out); }

.pop-enter-from,
.pop-leave-to {
  opacity: 0;
  transform: translateY(-6px) scale(0.97);
}

.quick {
  display: flex;
  gap: 6px;
  margin-bottom: 14px;
}

.qc {
  flex: 1;
  height: 30px;
  border: 0;
  border-radius: var(--r-pill);
  background: none;
  font-size: 12.5px;
  color: var(--text-2);
  box-shadow: inset 0 0 0 1px var(--line);
  transition: background-color var(--dur-fast), color var(--dur-fast), box-shadow var(--dur-fast),
    transform var(--dur-fast) var(--ease-spring);

  &:hover { background: var(--fill); color: var(--text); box-shadow: inset 0 0 0 1px var(--line-2); }
  &:active { transform: scale(0.96); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
  &.on { background: var(--lift); color: var(--lift-fg); font-weight: 500; box-shadow: var(--lift-shadow); }
}

.foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 12px;
  padding-top: 12px;
  box-shadow: inset 0 0.5px 0 var(--line);
  font-size: 12px;
  color: var(--text-3);
}

.clr {
  border: 0;
  background: none;
  font-size: 12px;
  color: var(--ink);

  &:hover { text-decoration: underline; text-underline-offset: 3px; }
}
</style>
