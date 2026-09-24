<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';

/** 等宽数字滚动计数：进入视口后从 0 缓动到 value（尊重 reduced-motion） */
const props = defineProps<{ value: number }>();

const el = ref<HTMLElement | null>(null);
const shown = ref(0);
let started = false;
let raf = 0;
let io: IntersectionObserver | null = null;

const reduce = () => window.matchMedia('(prefers-reduced-motion: reduce)').matches;

function run(to: number): void {
  cancelAnimationFrame(raf);
  if (reduce()) { shown.value = to; return; }
  const from = shown.value;
  const t0 = performance.now();
  const dur = 1300 + Math.min(900, Math.abs(to - from) * 2);
  const step = (t: number) => {
    const p = Math.min(1, (t - t0) / dur);
    shown.value = Math.round(from + (to - from) * (1 - Math.pow(1 - p, 4)));
    if (p < 1) raf = requestAnimationFrame(step);
  };
  raf = requestAnimationFrame(step);
}

onMounted(() => {
  io = new IntersectionObserver((es) => {
    if (es.some((e) => e.isIntersecting)) {
      started = true;
      run(props.value);
      io?.disconnect();
    }
  }, { threshold: 0.2 });
  if (el.value) io.observe(el.value);
});

watch(() => props.value, (v) => { if (started) run(v); });

onBeforeUnmount(() => { io?.disconnect(); cancelAnimationFrame(raf); });
</script>

<template>
  <span ref="el">{{ shown.toLocaleString() }}</span>
</template>
