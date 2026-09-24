<script setup lang="ts" generic="T extends string | number">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import SIcon from './SIcon.vue';

/** 分段控件：滑块跟随选中项（弹簧曲线） */
interface SegOption<V> {
  value: V;
  label?: string;
  icon?: string;
  title?: string;
  count?: number | string;
}

const props = defineProps<{ options: SegOption<T>[]; iconOnly?: boolean }>();
const model = defineModel<T>({ required: true });

const root = ref<HTMLElement | null>(null);
const knob = ref({ x: 0, w: 0, ready: false });

function measure(): void {
  const el = root.value?.querySelector<HTMLElement>('button.on');
  if (!el) return;
  knob.value = { x: el.offsetLeft, w: el.offsetWidth, ready: knob.value.ready || el.offsetWidth > 0 };
}

let ro: ResizeObserver | null = null;
onMounted(() => {
  void nextTick(measure);
  void document.fonts?.ready.then(measure);
  ro = new ResizeObserver(measure);
  if (root.value) ro.observe(root.value);
});
onBeforeUnmount(() => ro?.disconnect());
watch([model, () => props.options], () => void nextTick(measure));
</script>

<template>
  <span ref="root" class="seg" :class="{ icon: iconOnly }">
    <span class="k" :class="{ ready: knob.ready }" :style="{ width: `${knob.w}px`, transform: `translateX(${knob.x}px)` }" />
    <button
      v-for="o in options"
      :key="String(o.value)"
      type="button"
      :class="{ on: o.value === model }"
      :title="o.title"
      @click="model = o.value"
    >
      <SIcon v-if="o.icon" :name="o.icon" :size="16" />
      <span v-if="o.label">{{ o.label }}</span>
      <span v-if="o.count !== undefined" class="mono n">{{ o.count }}</span>
    </button>
  </span>
</template>

<style scoped lang="scss">
.seg {
  position: relative;
  display: inline-flex;
  background: var(--well);
  border-radius: 11px;
  padding: 3px;
  box-shadow: 0 0 0 1px var(--line) inset;
  flex: none;

  button {
    position: relative;
    z-index: 1;
    height: 30px;
    padding: 0 14px;
    border-radius: 9px;
    font-size: 13px;
    color: var(--ink-3);
    display: inline-flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
    transition: color var(--dur);

    &:hover { color: var(--ink-2); }
    &.on { color: var(--ink); font-weight: 500; }
  }

  &.icon button { padding: 0 10px; }

  .n { font-size: 11.5px; opacity: 0.7; }
}

.k {
  position: absolute;
  top: 3px;
  left: 0;
  height: 30px;
  border-radius: 9px;
  background: var(--paper);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08), 0 0 0 1px var(--line);
  opacity: 0;

  &.ready {
    opacity: 1;
    transition: transform var(--dur) var(--ease-spring), width var(--dur) var(--ease-spring);
  }
}

:global(:root[data-mode='dark'] .studio .seg .k) { background: var(--well-2); }
</style>
