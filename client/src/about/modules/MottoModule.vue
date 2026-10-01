<script setup lang="ts">
import { computed } from 'vue';
import { MOTTO_FLOURISHES, type MottoData } from '../types';
import type { ModProps } from './props';

/**
 * 格言（motto）：closing = 页尾整行宋体大字逐字模糊浮现 + 可选收尾装饰（data.flourish）；card = 卡片变体（不带装饰）。
 * 收尾装饰：line 落款线 / horizon 地平线 / quote 引号（包住整句两端）/ none。
 * 装饰在文字浮现完成后接续入场（--n = 字数），颜色全部取语义 token，深浅模式自动适配。
 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as MottoData);
const chars = computed(() => [...(d.value.text ?? '')]);
const fx = computed(() => {
  if (props.variant === 'card') return 'none';
  const f = d.value.flourish ?? 'line';
  return (MOTTO_FLOURISHES as string[]).includes(f) ? f : 'line';
});
</script>

<template>
  <div class="mo" :class="[`fx-${fx}`, { 'card-v': variant === 'card' }]" :style="{ '--n': chars.length }">
    <p>
      <span class="mo-line">
        <span v-if="fx === 'quote'" class="mo-q open" aria-hidden="true">&ldquo;</span>
        <span class="mo-text"><span v-for="(ch, i) in chars" :key="i" class="ch" :style="{ '--k': i }">{{ ch }}</span></span>
        <span v-if="fx === 'quote'" class="mo-q close" aria-hidden="true">&rdquo;</span>
      </span>
    </p>
    <span v-if="fx === 'horizon'" class="mo-horizon" aria-hidden="true"><i /></span>
    <small v-if="d.sign">{{ d.sign }}</small>
  </div>
</template>

<style scoped lang="scss">
.mo {
  --after: calc(var(--n, 8) * 70ms + 250ms);

  position: relative;
  padding: 56px 0 72px;
  text-align: center;

  p { position: relative; font: 700 clamp(30px, 5.6cqi, 64px)/1.2 var(--font-serif); letter-spacing: 0.06em; }

  .ch {
    display: inline-block;
    opacity: 0;
    filter: blur(10px);
    transform: translateY(12px);
    transition: opacity 0.9s var(--ease-out), filter 0.9s var(--ease-out), transform 0.9s var(--ease-out);
    transition-delay: calc(var(--k) * 70ms + 150ms);
  }

  small { position: relative; display: block; margin-top: 18px; font: 500 13px var(--ak-mono); letter-spacing: 0.2em; color: var(--ak-text-3); }
}

.in .mo .ch { opacity: 1; filter: none; transform: none; }

/* ---------- line 落款线：中性细线自中心向两侧展开，中段一小截墨色加粗作「落笔」 ---------- */
.fx-line {
  &::before,
  &::after {
    content: '';
    position: absolute;
    left: 50%;
    bottom: 36px;
    translate: -50% 0;
    transform: scaleX(0);
    transition: transform 1.1s var(--ease-out) var(--after);
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
    transition-delay: calc(var(--after) - 100ms);
  }
}

.in .fx-line::before,
.in .fx-line::after { transform: scaleX(1); }

/* ---------- horizon 地平线：一道通栏细线，线上方升起一片极淡的主色晨光（只在线上方，不外溢发光） ---------- */
.mo-horizon {
  position: absolute;
  left: 50%;
  bottom: 36px;
  width: min(760px, 92%);
  height: 150px;
  translate: -50% 0;
  pointer-events: none;

  /* 晨光：椭圆底边贴线，自下而上升起 */
  &::before {
    content: '';
    position: absolute;
    inset: 0;
    background: radial-gradient(50% 100% at 50% 100%, rgba(var(--primary-rgb), var(--horizon-a, 0.11)), transparent 100%);
    transform: scaleY(0);
    transform-origin: 50% 100%;
    opacity: 0;
    transition: transform 1.4s var(--ease-out) var(--after), opacity 1.4s var(--ease-out) var(--after);
  }

  /* 地平线：两端渐隐 */
  i {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    height: 1px;
    background: linear-gradient(90deg, transparent, var(--line-2) 22%, color-mix(in oklab, var(--ak-ink) 55%, var(--line-2)) 50%, var(--line-2) 78%, transparent);
    transform: scaleX(0);
    transition: transform 1.2s var(--ease-out) calc(var(--after) - 150ms);
  }
}

.fx-horizon p { z-index: 1; }
:root[data-mode='dark'] .mo-horizon { --horizon-a: 0.2; }

.in .mo-horizon {
  &::before { transform: none; opacity: 1; }
  i { transform: none; }
}

/* Quote gutters participate in layout, so wrapped text never runs under a quote. */
.mo-line { position: relative; }
.fx-quote .mo-line { display: inline-grid; grid-template-columns: auto minmax(0, 1fr) auto; gap: .08em; max-width: 100%; align-items: center; }
.fx-quote .mo-text { min-width: 0; overflow-wrap: anywhere; }

.mo-q {
  display: block;
  font: 700 1.7em/1 var(--font-serif);
  color: var(--line-2);
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.9s var(--ease-out), transform 0.9s var(--ease-out);

  &.open { align-self: start; transform: translate(-10px, -6px); transition-delay: 0.1s; }
  &.close { align-self: end; position: relative; top: .35em; transform: translate(10px, 6px); transition-delay: var(--after); }
}

.in .mo-q { opacity: 1; transform: none; }

/* ---------- card 卡片变体：左对齐小字号，不带装饰 ---------- */
.mo.card-v {
  padding: 8px 0 10px;
  text-align: left;

  p { font-size: 26px; letter-spacing: 0.04em; }
}

.fx-none { padding-bottom: 56px; }

@media (prefers-reduced-motion: reduce) {
  .mo *, .mo::before, .mo::after, .mo *::before { transition: none !important; }
  .mo .ch, .mo-q { opacity: 1; filter: none; transform: none; }
  .fx-line::before, .fx-line::after, .mo-horizon i, .mo-horizon::before { transform: none; opacity: 1; }
}
</style>
