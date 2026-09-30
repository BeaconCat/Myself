<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
import { Check } from 'lucide';
import Icon from './Icon.vue';
import { highlightParts, type MatchedOption, type UiOption } from './listbox';

/**
 * 下拉浮层（Select / Combobox 共用）：Teleport 到 body，按锚点 fixed 定位（下方放不下时翻到上方），
 * 滚动 / 缩放时跟随；点外部关闭。锚点位于后台 .studio 内时浮层同样挂 .studio，取到后台 token。
 * 键盘由宿主组件处理，这里只负责渲染、高亮与把当前项滚入可视区。
 */
const props = withDefaults(
  defineProps<{
    open: boolean;
    anchor: HTMLElement | null;
    items: MatchedOption[];
    active: number;
    selected?: string;
    empty?: string;
    minWidth?: number;
    listId?: string;
  }>(),
  { selected: undefined, empty: '', minWidth: 0, listId: undefined },
);
const emit = defineEmits<{ pick: [option: UiOption]; hover: [index: number]; close: [] }>();

const pop = ref<HTMLElement | null>(null);
const list = ref<HTMLElement | null>(null);
const pos = ref({ top: 0, left: 0, width: 0, maxH: 320, up: false });
const studio = ref(false);

function place(): void {
  const a = props.anchor;
  if (!a) return;
  const r = a.getBoundingClientRect();
  const gap = 6;
  const below = window.innerHeight - r.bottom - gap - 12;
  const above = r.top - gap - 12;
  const up = below < 220 && above > below;
  const width = Math.max(r.width, props.minWidth);
  const left = Math.min(Math.max(8, r.left), window.innerWidth - width - 8);
  pos.value = {
    top: up ? r.top - gap : r.bottom + gap,
    left,
    width,
    maxH: Math.max(120, Math.min(320, up ? above : below)),
    up,
  };
}

let raf = 0;
function schedule(): void {
  cancelAnimationFrame(raf);
  raf = requestAnimationFrame(place);
}

function onDown(e: PointerEvent): void {
  const t = e.target as Node;
  if (pop.value?.contains(t) || props.anchor?.contains(t)) return;
  emit('close');
}

function bind(on: boolean): void {
  const fn = on ? window.addEventListener : window.removeEventListener;
  fn('scroll', schedule, true);
  fn('resize', schedule);
  (on ? document.addEventListener : document.removeEventListener)('pointerdown', onDown, true);
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      studio.value = !!props.anchor?.closest('.studio');
      place();
      bind(true);
    } else bind(false);
  },
  { immediate: true },
);
onBeforeUnmount(() => {
  bind(false);
  cancelAnimationFrame(raf);
});

/* 当前项滚入可视区 */
watch(
  () => [props.active, props.open] as const,
  async ([i, open]) => {
    if (!open || i < 0) return;
    await nextTick();
    const el = list.value?.children[i] as HTMLElement | undefined;
    el?.scrollIntoView({ block: 'nearest' });
  },
);
/* 选项数量变化（筛选）时浮层高度变化，翻转位置需重算 */
watch(() => props.items.length, () => props.open && schedule());

const style = computed(() => ({
  left: `${pos.value.left}px`,
  width: `${pos.value.width}px`,
  maxHeight: `${pos.value.maxH}px`,
  ...(pos.value.up ? { bottom: `${window.innerHeight - pos.value.top}px` } : { top: `${pos.value.top}px` }),
}));

const optId = (i: number): string | undefined => (props.listId ? `${props.listId}-${i}` : undefined);
</script>

<template>
  <Teleport to="body">
    <Transition name="ui-lb">
      <div
        v-if="open"
        ref="pop"
        class="ui-lb"
        :class="{ studio, up: pos.up }"
        :style="style"
      >
        <slot name="header" />
        <ul :id="listId" ref="list" class="ui-lb-list" role="listbox" @mousedown.prevent>
          <li
            v-for="(m, i) in items"
            :id="optId(i)"
            :key="m.option.value"
            role="option"
            class="ui-lb-opt"
            :class="{ active: i === active, on: m.option.value === selected, off: m.option.disabled }"
            :aria-selected="m.option.value === selected"
            :aria-disabled="m.option.disabled || undefined"
            @mousemove="i !== active && emit('hover', i)"
            @click="!m.option.disabled && emit('pick', m.option)"
          >
            <span v-if="$slots.icon" class="ui-lb-ic"><slot name="icon" :option="m.option" /></span>
            <span class="ui-lb-label">
              <template v-for="(p, k) in highlightParts(m.option.label, m.hits)" :key="k">
                <mark v-if="p.hit">{{ p.text }}</mark><template v-else>{{ p.text }}</template>
              </template>
            </span>
            <span v-if="m.option.hint" class="ui-lb-hint">{{ m.option.hint }}</span>
            <Icon v-if="m.option.value === selected" class="ui-lb-tick" :icon="Check" :size="15" :stroke="2.2" />
          </li>
        </ul>
        <p v-if="!items.length && empty" class="ui-lb-empty">{{ empty }}</p>
      </div>
    </Transition>
  </Teleport>
</template>

<style lang="scss">
/* 浮层 Teleport 到 body：不用 scoped，类名统一 ui-lb- 前缀 */
.ui-lb {
  position: fixed;
  z-index: 1300;
  display: flex;
  flex-direction: column;
  min-width: 140px;
  padding: 4px;
  overflow: hidden;
  border-radius: var(--r-md);
  background: var(--paper, var(--surface, #fff));
  color: var(--text);
  box-shadow: var(--sh-pop, 0 0 0 1px var(--line-2), 0 18px 44px -14px rgb(0 0 0 / 0.35));
  transform-origin: top center;
  font-family: var(--font-sans);

  &.up { transform-origin: bottom center; }
}

.ui-lb-list {
  flex: 1;
  min-height: 0;
  margin: 0;
  padding: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  list-style: none;
}

.ui-lb-opt {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 34px;
  padding: 6px 10px;
  border-radius: var(--r-sm);
  font-size: 13.5px;
  line-height: 1.3;
  cursor: pointer;
  transition: background-color var(--dur-fast), color var(--dur-fast);

  &.active { background: var(--hover, var(--fill)); }
  &.on { font-weight: 600; }
  &.off { opacity: 0.4; cursor: default; }

  mark {
    padding: 0;
    border-radius: 2px;
    background: color-mix(in oklab, var(--ink) 16%, transparent);
    color: inherit;
    font-weight: 600;
  }
}

.ui-lb-ic {
  display: grid;
  place-items: center;
  flex: none;
  min-width: 22px;
  height: 22px;
  color: var(--text-2);
}

.ui-lb-label { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ui-lb-hint { flex: none; font: 12px var(--font-mono, monospace); color: var(--text-2); }

.ui-lb-tick { color: var(--ink); }

.ui-lb-empty { margin: 0; padding: 12px 10px; font-size: 13px; color: var(--text-2); text-align: center; }

.ui-lb-enter-active { transition: opacity var(--dur-fast) var(--ease-out), transform var(--dur) var(--ease-spring); }
.ui-lb-leave-active { transition: opacity var(--dur-fast) var(--ease-out), transform var(--dur-fast) var(--ease-out); }
.ui-lb-enter-from, .ui-lb-leave-to { opacity: 0; transform: translateY(-4px) scale(0.97); }
.ui-lb.up.ui-lb-enter-from, .ui-lb.up.ui-lb-leave-to { transform: translateY(4px) scale(0.97); }

@media (prefers-reduced-motion: reduce) {
  .ui-lb-enter-active, .ui-lb-leave-active { transition: opacity var(--dur-fast); }
  .ui-lb-enter-from, .ui-lb-leave-to, .ui-lb.up.ui-lb-enter-from, .ui-lb.up.ui-lb-leave-to { transform: none; }
}
</style>
