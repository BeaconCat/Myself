<script setup lang="ts">
import { computed } from 'vue';
import type { MottoData } from '../types';
import type { ModProps } from './props';

/** 格言（motto）：closing = 页尾整行宋体大字逐字模糊浮现 + 下方一道自中心展开的落款线；card = 卡片变体 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as MottoData);
const chars = computed(() => [...(d.value.text ?? '')]);
</script>

<template>
  <div class="mo" :class="{ 'card-v': variant === 'card' }" :style="{ '--n': chars.length }">
    <p><span v-for="(ch, i) in chars" :key="i" :style="{ '--k': i }">{{ ch }}</span></p>
    <small v-if="d.sign">{{ d.sign }}</small>
  </div>
</template>

<style scoped lang="scss">
.mo {
  position: relative;
  padding: 56px 0 72px;
  text-align: center;

  /*
   * 收尾落款线：一道中性细线自中心向两侧展开，中段一小截墨色加粗作为「落笔」。
   * 全部取语义 token（--line-2 / --ak-ink），深浅模式自动适配，不再使用辉光。
   */
  &::before,
  &::after {
    content: '';
    position: absolute;
    left: 50%;
    bottom: 36px;
    translate: -50% 0;
    transform: scaleX(0);
    transition: transform 1.1s var(--ease-out);
    transition-delay: calc(var(--n, 8) * 70ms + 250ms);
  }

  &::before {
    width: min(420px, 62%);
    height: 1px;
    background: linear-gradient(90deg, transparent, var(--line-2) 30%, var(--line-2) 70%, transparent);
  }

  &::after {
    width: 44px;
    height: 3px;
    margin-bottom: -1px;
    border-radius: var(--r-pill);
    background: var(--ak-ink);
    transition-duration: 0.7s;
    transition-delay: calc(var(--n, 8) * 70ms + 150ms);
  }

  p { position: relative; font: 700 clamp(30px, 5.6cqi, 64px)/1.2 var(--font-serif); letter-spacing: 0.06em; }

  p span {
    display: inline-block;
    opacity: 0;
    filter: blur(10px);
    transform: translateY(12px);
    transition: opacity 0.9s var(--ease-out), filter 0.9s var(--ease-out), transform 0.9s var(--ease-out);
    transition-delay: calc(var(--k) * 70ms + 150ms);
  }

  small { position: relative; display: block; margin-top: 18px; font: 500 13px var(--ak-mono); letter-spacing: 0.2em; color: var(--ak-text-3); }
}

.in .mo p span { opacity: 1; filter: none; transform: none; }
.in .mo::before,
.in .mo::after { transform: scaleX(1); }

.mo.card-v {
  padding: 8px 0 10px;
  text-align: left;

  /* 卡片形态不带落款线 */
  &::after, &::before { display: none; }
  p { font-size: 26px; letter-spacing: 0.04em; }
}

@media (prefers-reduced-motion: reduce) {
  .mo p span { opacity: 1; filter: none; transform: none; }
  .mo::before, .mo::after { transform: none; transition: none; }
}
</style>
