<script setup lang="ts">
/** 进度环：value 0..1；indeterminate 时转圈（上传中）。 */
import { computed } from 'vue';

const props = withDefaults(defineProps<{ value?: number; size?: number; stroke?: number; indeterminate?: boolean }>(), {
  value: 0,
  size: 34,
  stroke: 3,
  indeterminate: false,
});
const r = computed(() => (props.size - props.stroke) / 2);
const c = computed(() => 2 * Math.PI * r.value);
const offset = computed(() => c.value * (1 - Math.max(0, Math.min(1, props.value))));
</script>

<template>
  <svg
    class="ma-ring"
    :class="{ spin: indeterminate }"
    :width="size"
    :height="size"
    :viewBox="`0 0 ${size} ${size}`"
    role="progressbar"
    :aria-valuenow="Math.round(value * 100)"
    aria-valuemin="0"
    aria-valuemax="100"
  >
    <circle class="bg" :cx="size / 2" :cy="size / 2" :r="r" :stroke-width="stroke" />
    <circle
      class="fg"
      :cx="size / 2"
      :cy="size / 2"
      :r="r"
      :stroke-width="stroke"
      :stroke-dasharray="c"
      :stroke-dashoffset="indeterminate ? c * 0.72 : offset"
    />
  </svg>
</template>

<style scoped lang="scss">
.ma-ring {
  transform: rotate(-90deg);

  circle { fill: none; }
  .bg { stroke: var(--ring-bg, rgba(255, 255, 255, 0.25)); }

  .fg {
    stroke: var(--ring-fg, #fff);
    stroke-linecap: round;
    transition: stroke-dashoffset 0.35s var(--ease-out);
  }

  &.spin { animation: ma-ring-spin 0.9s linear infinite; }
}

@keyframes ma-ring-spin {
  from { transform: rotate(-90deg); }
  to { transform: rotate(270deg); }
}
</style>
