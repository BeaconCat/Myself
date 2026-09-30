<script setup lang="ts">
import { computed, ref } from 'vue';

/**
 * 可拖动改尺寸的外框（编辑器里的图片 / 视频 / 音频）：右缘、下缘、右下角三个拖柄。
 * - 默认自由变换（宽高各自改变，内容 object-fit: cover 填满）；按住 Shift 锁定当前比例；
 * - axis = 'x' 只改宽（音频条）；居中对齐时右缘拖动按两倍改宽，手感与所见一致；
 * - 双击拖柄恢复原始尺寸。拖动中实时预览，松手才 emit('change') 写回节点。
 * 外框把宽度与比例写成 CSS 变量，插槽里的媒体按 `width: 100%; aspect-ratio: var(--rz-ar)` 跟随。
 */
const props = withDefaults(defineProps<{
  w: number | null;
  h: number | null;
  selected: boolean;
  axis?: 'both' | 'x';
  centered?: boolean;
}>(), { axis: 'both', centered: false });
const emit = defineEmits<{ change: [size: { w: number | null; h: number | null }] }>();

const box = ref<HTMLElement | null>(null);
const live = ref<{ w: number; h: number | null } | null>(null);

const cur = computed(() => live.value ?? (props.w ? { w: props.w, h: props.h } : null));
const style = computed(() => {
  const c = cur.value;
  if (!c) return {};
  return { width: `${c.w}px`, '--rz-ar': c.h ? `${c.w} / ${c.h}` : 'auto' };
});

type Dir = 'e' | 's' | 'se';
let drag: { dir: Dir; x: number; y: number; w0: number; h0: number; ratio: number; max: number } | null = null;

function start(e: PointerEvent, dir: Dir): void {
  const el = box.value;
  if (!el || e.button !== 0) return;
  e.preventDefault();
  e.stopPropagation();
  const r = el.getBoundingClientRect();
  const host = el.closest('.ProseMirror') as HTMLElement | null;
  drag = { dir, x: e.clientX, y: e.clientY, w0: r.width, h0: r.height, ratio: r.width / Math.max(1, r.height), max: host?.clientWidth ?? 2000 };
  (e.target as HTMLElement).setPointerCapture(e.pointerId);
}

function move(e: PointerEvent): void {
  if (!drag) return;
  const k = props.centered ? 2 : 1;
  const dx = (e.clientX - drag.x) * k;
  const dy = e.clientY - drag.y;
  let w = drag.w0;
  let h = drag.h0;
  if (drag.dir !== 's') w = drag.w0 + dx;
  if (drag.dir !== 'e' && props.axis === 'both') h = drag.h0 + dy;
  if (e.shiftKey && props.axis === 'both') {
    // 按比例：以主拖动方向为准
    if (drag.dir === 's') w = h * drag.ratio;
    else h = w / drag.ratio;
  }
  w = Math.round(Math.min(drag.max, Math.max(48, w)));
  h = Math.round(Math.min(2400, Math.max(32, h)));
  live.value = { w, h: props.axis === 'x' ? null : h };
}

function end(): void {
  if (!drag) return;
  drag = null;
  if (live.value) emit('change', { ...live.value });
  live.value = null;
}

function reset(): void {
  emit('change', { w: null, h: null });
}
</script>

<template>
  <span ref="box" class="rz" :class="{ sel: selected, active: !!live, sized: !!cur }" :style="style">
    <slot />
    <template v-if="selected || live">
      <i class="hd e" @pointerdown="start($event, 'e')" @pointermove="move" @pointerup="end" @pointercancel="end" @dblclick="reset" />
      <template v-if="axis === 'both'">
        <i class="hd s" @pointerdown="start($event, 's')" @pointermove="move" @pointerup="end" @pointercancel="end" @dblclick="reset" />
        <i class="hd se" @pointerdown="start($event, 'se')" @pointermove="move" @pointerup="end" @pointercancel="end" @dblclick="reset" />
      </template>
      <span v-if="live" class="dim">{{ live.w }}<template v-if="live.h"> × {{ live.h }}</template></span>
    </template>
  </span>
</template>

<style scoped lang="scss">
.rz {
  position: relative;
  display: inline-block;
  max-width: 100%;
  vertical-align: bottom;
  line-height: 0;

  &.sel { outline: 2px solid var(--ink, var(--primary)); outline-offset: 2px; border-radius: var(--r-xs); }
  &.active { user-select: none; }
}

.hd {
  position: absolute;
  z-index: 4;
  background: var(--paper, #fff);
  box-shadow: 0 0 0 1.5px var(--ink, var(--primary)), 0 1px 4px rgba(0, 0, 0, 0.2);
  border-radius: var(--r-pill);
  touch-action: none;

  &.e { top: 50%; right: -6px; width: 8px; height: 32px; translate: 0 -50%; cursor: ew-resize; }
  &.s { left: 50%; bottom: -6px; width: 32px; height: 8px; translate: -50% 0; cursor: ns-resize; }
  &.se { right: -7px; bottom: -7px; width: 12px; height: 12px; cursor: nwse-resize; }
}

.dim {
  position: absolute;
  right: 8px;
  bottom: 8px;
  z-index: 4;
  padding: 3px 8px;
  border-radius: var(--r-xs);
  font: 500 12px/16px var(--font-mono);
  color: #fff;
  background: rgba(0, 0, 0, 0.6);
  pointer-events: none;
}
</style>
