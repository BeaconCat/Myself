<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { ChevronLeft, ChevronRight } from 'lucide';
import Icon from '../../../components/ui/Icon.vue';
import './i18n';

/**
 * Studio 分页条：左侧「第 a–b 条 / 共 n 条」，右侧 上一页 · 页码（省略号折叠）· 下一页，可选每页条数。
 * 页码槽位数固定（首尾 + 当前页两侧 siblings 个），翻页时按钮不跳动；选中块在页码间滑动。
 * 用法：<StPager v-model:page="page" v-model:page-size="size" :total="total" :page-sizes="[20, 50, 100]" />
 */
const props = withDefaults(
  defineProps<{
    total: number;
    /** 当前页两侧各显示几个页码 */
    siblings?: number;
    /** 给出时显示每页条数选择 */
    pageSizes?: number[];
    /** 数据加载中：禁用按钮，避免连点 */
    busy?: boolean;
  }>(),
  { siblings: 1, pageSizes: undefined, busy: false },
);
const page = defineModel<number>('page', { required: true });
const pageSize = defineModel<number>('pageSize', { default: 20 });
const { t } = useI18n();

const pages = computed(() => Math.max(1, Math.ceil(props.total / Math.max(1, pageSize.value))));
const from = computed(() => (props.total ? (page.value - 1) * pageSize.value + 1 : 0));
const to = computed(() => Math.min(props.total, page.value * pageSize.value));

type Slot = number | 'gap-l' | 'gap-r';
/** 页码槽：总页数少时全列，否则 1 … 当前±siblings … 末页，槽位总数恒定 */
const slots = computed<Slot[]>(() => {
  const n = pages.value;
  const s = props.siblings;
  const width = s * 2 + 5;
  if (n <= width) return Array.from({ length: n }, (_, i) => i + 1);
  const cur = Math.min(Math.max(page.value, 1), n);
  const left = Math.max(cur - s, 2);
  const right = Math.min(cur + s, n - 1);
  const showL = left > 3;
  const showR = right < n - 2;
  if (!showL) return [...Array.from({ length: width - 2 }, (_, i) => i + 1), 'gap-r', n];
  if (!showR) return [1, 'gap-l', ...Array.from({ length: width - 2 }, (_, i) => n - (width - 3) + i)];
  return [1, 'gap-l', ...Array.from({ length: right - left + 1 }, (_, i) => left + i), 'gap-r', n];
});

function go(p: number): void {
  const next = Math.min(Math.max(1, p), pages.value);
  if (next !== page.value && !props.busy) page.value = next;
}

/** 省略号：向对应方向跳过一组 */
function jump(gap: 'gap-l' | 'gap-r'): void {
  const step = props.siblings * 2 + 1;
  go(gap === 'gap-l' ? page.value - step : page.value + step);
}

function setSize(e: Event): void {
  const n = Number((e.target as HTMLSelectElement).value);
  if (!n || n === pageSize.value) return;
  // 换每页条数时保持当前首条仍在可视页内
  const first = (page.value - 1) * pageSize.value;
  pageSize.value = n;
  page.value = Math.floor(first / n) + 1;
}

/* 选中块：量出当前页按钮位置，left / width 过渡 */
const nums = ref<HTMLElement | null>(null);
const knob = ref({ x: 0, w: 0, ready: false });
function measure(): void {
  const el = nums.value?.querySelector<HTMLElement>('button[aria-current="page"]');
  if (!el) {
    knob.value = { ...knob.value, ready: false };
    return;
  }
  knob.value = { x: el.offsetLeft, w: el.offsetWidth, ready: knob.value.ready || el.offsetWidth > 0 };
}
let ro: ResizeObserver | null = null;
onMounted(() => {
  void nextTick(measure);
  ro = new ResizeObserver(measure);
  if (nums.value) ro.observe(nums.value);
});
onBeforeUnmount(() => ro?.disconnect());
watch([page, slots], () => void nextTick(measure));

/* 总数变少（删除 / 筛选）导致当前页越界时回到末页 */
watch(pages, (n) => {
  if (page.value > n) page.value = n;
});
</script>

<template>
  <nav v-if="total > 0" class="pager" :class="{ busy }" :aria-label="t('studio.pager.label')">
    <span class="sum">
      <span class="mono">{{ from }}–{{ to }}</span>
      <span>{{ t('studio.pager.total', { n: total }) }}</span>
    </span>
    <span class="sp" />
    <label v-if="pageSizes?.length" class="size">
      <span class="vh">{{ t('studio.pager.pageSize') }}</span>
      <select :value="pageSize" @change="setSize">
        <option v-for="n in pageSizes" :key="n" :value="n">{{ t('studio.pager.perPage', { n }) }}</option>
      </select>
    </label>
    <div v-if="pages > 1" class="nav">
      <button type="button" class="arrow" :disabled="page <= 1 || busy" :aria-label="t('studio.pager.prev')" @click="go(page - 1)">
        <Icon :icon="ChevronLeft" :size="16" />
      </button>
      <div ref="nums" class="nums">
        <span class="k" :class="{ ready: knob.ready }" :style="{ transform: `translateX(${knob.x}px)`, width: `${knob.w}px` }" />
        <template v-for="(s, i) in slots" :key="i">
          <button
            v-if="typeof s === 'string'"
            type="button"
            class="gap"
            :disabled="busy"
            :aria-label="s === 'gap-l' ? t('studio.pager.back', { n: siblings * 2 + 1 }) : t('studio.pager.forward', { n: siblings * 2 + 1 })"
            @click="jump(s)"
          >
            …
          </button>
          <button
            v-else
            type="button"
            class="mono"
            :aria-current="s === page ? 'page' : undefined"
            :aria-label="t('studio.pager.page', { n: s })"
            :disabled="busy && s !== page"
            @click="go(s)"
          >
            {{ s }}
          </button>
        </template>
      </div>
      <button type="button" class="arrow" :disabled="page >= pages || busy" :aria-label="t('studio.pager.next')" @click="go(page + 1)">
        <Icon :icon="ChevronRight" :size="16" />
      </button>
    </div>
  </nav>
</template>

<style scoped lang="scss">
.pager {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  padding: 14px 4px 0;
  font-size: 13px;
  color: var(--st-ink-3);
}

.sum {
  display: inline-flex;
  align-items: baseline;
  gap: 8px;

  .mono { font-size: 12.5px; color: var(--st-ink-2); }
}

.sp { flex: 1; }

.vh { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }

.size select {
  height: 32px;
  padding: 0 12px;
  border: 0;
  border-radius: var(--r-sm);
  background: var(--well);
  box-shadow: 0 0 0 1px var(--line) inset;
  color: var(--st-ink-2);
  font-size: 13px;
  cursor: pointer;
  transition: box-shadow var(--dur-fast);

  &:hover { box-shadow: 0 0 0 1px var(--line-2) inset; }
  &:focus-visible { box-shadow: var(--focus); }
}

.nav {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px;
  border-radius: var(--r-sm);
  background: var(--well);
  box-shadow: 0 0 0 1px var(--line) inset;
}

.nums {
  position: relative;
  display: inline-flex;
  gap: 2px;
}

button {
  position: relative;
  z-index: 1;
  min-width: 30px;
  height: 30px;
  padding: 0 8px;
  border-radius: var(--r-xs);
  display: inline-grid;
  place-items: center;
  font-size: 13px;
  color: var(--st-ink-3);
  transition: color var(--dur) var(--ease-out), background var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  &:hover:not(:disabled):not([aria-current]) { background: var(--hover); color: var(--st-ink); }
  &:active:not(:disabled) { transform: scale(0.94); }
  &:focus-visible { outline: 0; box-shadow: var(--focus); }
  &:disabled { cursor: default; opacity: 0.4; }
  &[aria-current='page'] { color: var(--lift-fg); font-weight: 600; opacity: 1; }
}

.arrow { padding: 0; width: 30px; }
.gap { letter-spacing: 0.1em; color: var(--st-ink-4); }

/* 选中块：抬升 + 轻染，在页码间滑动 */
.k {
  position: absolute;
  top: 0;
  left: 0;
  height: 30px;
  border-radius: var(--r-xs);
  background: var(--lift);
  box-shadow: var(--lift-shadow);
  opacity: 0;
  pointer-events: none;

  &.ready {
    opacity: 1;
    transition: transform 0.38s var(--ease-spring), width 0.3s var(--ease-out), opacity var(--dur-fast);
  }
}

.busy .nums { opacity: 0.75; }

@media (max-width: 767px) {
  .pager { flex-wrap: wrap; gap: 10px; }
  .sum { flex: 1; flex-wrap: wrap; gap: 4px 8px; }
  .sp { display: none; }
  .nav { margin-left: auto; max-width: 100%; }
}

@media (max-width: 360px) {
  .nav { gap: 2px; }
  button { min-width: 25px; padding: 0 5px; }
  .arrow { width: 25px; }
}

@media (prefers-reduced-motion: reduce) {
  .k.ready, button { transition: none; }
}
</style>
