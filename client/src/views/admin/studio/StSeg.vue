<script setup lang="ts" generic="T extends string | number">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import SIcon from './SIcon.vue';

/** 分段控件：抬升 + 轻染的选中块在项之间 morph（前缘 .30s ease-out，后缘 .46s spring） */
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
const knob = ref({ l: 0, r: 0, toL: false, ready: false });

function measure(): void {
  const box = root.value;
  const el = box?.querySelector<HTMLElement>('button.on');
  if (!box || !el) return;
  const l = el.offsetLeft;
  knob.value = {
    l,
    r: box.clientWidth - l - el.offsetWidth,
    toL: l < knob.value.l,
    ready: knob.value.ready || el.offsetWidth > 0,
  };
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
    <span class="k" :class="{ ready: knob.ready, 'to-l': knob.toL }" :style="{ left: `${knob.l}px`, right: `${knob.r}px` }" />
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
  border-radius: var(--r-sm);
  padding: 3px;
  box-shadow: 0 0 0 1px var(--line) inset;
  flex: none;

  button {
    position: relative;
    z-index: 1;
    height: 30px;
    padding: 0 14px;
    border-radius: var(--r-xs);
    font-size: 13px;
    color: var(--st-ink-3);
    display: inline-flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
    transition: color var(--dur) var(--ease-out), transform var(--dur-fast) var(--ease-spring);

    &:hover { color: var(--st-ink-2); }
    &:active { transform: scale(0.96); }
    &:focus-visible { outline: 0; box-shadow: var(--focus); }
    &.on { color: var(--lift-fg); font-weight: 500; }
    &.on .st-ic { color: var(--ink); }
  }

  &.icon button { padding: 0 10px; }

  .n { font-size: 11.5px; opacity: 0.7; }
}

.k {
  position: absolute;
  top: 3px;
  bottom: 3px;
  border-radius: var(--r-xs);
  background: var(--lift);
  box-shadow: var(--lift-shadow);
  opacity: 0;

  /* 向右：右缘领先、左缘拖尾；向左反之 */
  &.ready {
    opacity: 1;
    transition: right 0.3s var(--ease-out), left 0.46s var(--ease-spring) 0.05s, background-color var(--dur);
  }

  &.ready.to-l { transition: left 0.3s var(--ease-out), right 0.46s var(--ease-spring) 0.05s, background-color var(--dur); }
}
</style>
