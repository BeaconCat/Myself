<script setup lang="ts">
import { computed } from 'vue';
import type { MottoData } from '../types';
import type { ModProps } from './props';

/** 格言（motto）：closing = 页尾整行宋体大字逐字模糊浮现 + 下方一道光；card = 卡片变体 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as MottoData);
const chars = computed(() => [...(d.value.text ?? '')]);
</script>

<template>
  <div class="mo" :class="{ 'card-v': variant === 'card' }">
    <p><span v-for="(ch, i) in chars" :key="i" :style="{ '--k': i }">{{ ch }}</span></p>
    <small v-if="d.sign">{{ d.sign }}</small>
  </div>
</template>

<style scoped lang="scss">
.mo {
  position: relative;
  overflow: hidden;
  padding: 90px 0 110px;
  text-align: center;

  &::before {
    content: '';
    position: absolute;
    left: 50%;
    bottom: 0;
    width: 70%;
    height: 70%;
    translate: -50% 0;
    background: radial-gradient(50% 60% at 50% 100%, rgba(var(--primary-rgb), 0.28), transparent 70%);
    filter: blur(10px);
  }

  &::after {
    content: '';
    position: absolute;
    left: 50%;
    bottom: 0;
    width: 2px;
    height: 120px;
    translate: -50% 0;
    border-radius: 2px;
    background: linear-gradient(transparent, #fff);
    box-shadow: 0 0 18px 3px rgba(var(--primary-rgb), 0.7);
  }

  p { position: relative; font: 700 clamp(30px, 6.5cqi, 76px)/1.2 var(--font-serif); letter-spacing: 0.06em; }

  p span {
    display: inline-block;
    opacity: 0;
    filter: blur(10px);
    transform: translateY(12px);
    transition: opacity 0.9s var(--ease-out), filter 0.9s var(--ease-out), transform 0.9s var(--ease-out);
    transition-delay: calc(var(--k) * 70ms + 150ms);
  }

  small { position: relative; display: block; margin-top: 22px; font: 500 12px var(--ak-mono); letter-spacing: 0.2em; color: var(--ak-text-3); }
}

:root[data-mode='light'] .mo::after { background: linear-gradient(transparent, var(--primary)); }

.in .mo p span { opacity: 1; filter: none; transform: none; }

.mo.card-v {
  padding: 8px 0 10px;
  text-align: left;

  &::after { display: none; }
  &::before { width: 90%; height: 90%; left: 60%; }
  p { font-size: 26px; letter-spacing: 0.04em; }
}

@media (prefers-reduced-motion: reduce) {
  .mo p span { opacity: 1; filter: none; transform: none; }
}
</style>
