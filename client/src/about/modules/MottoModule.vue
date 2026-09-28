<script setup lang="ts">
import { computed } from 'vue';
import type { MottoData } from '../types';
import type { ModProps } from './props';

/**
 * 格言（motto）：closing = 页尾整行宋体大字逐字模糊浮现 + 可选收尾装饰（data.flourish）；card = 卡片变体（不带装饰）。
 * 收尾装饰：line 落款线 / horizon 地平线 / ink 墨痕 / seal 印章 / quote 引号 / none。
 * 装饰在文字浮现完成后接续入场（--n = 字数），颜色全部取语义 token，深浅模式自动适配。
 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as MottoData);
const chars = computed(() => [...(d.value.text ?? '')]);
const fx = computed(() => (props.variant === 'card' ? 'none' : d.value.flourish ?? 'line'));
const seal = computed(() => [...(d.value.seal?.trim() || chars.value[0] || '')].slice(0, 4));
</script>

<template>
  <div class="mo" :class="[`fx-${fx}`, { 'card-v': variant === 'card' }]" :style="{ '--n': chars.length }">
    <span v-if="fx === 'quote'" class="mo-quote" aria-hidden="true">&ldquo;</span>
    <p>
      <span class="mo-line">
        <span v-for="(ch, i) in chars" :key="i" class="ch" :style="{ '--k': i }">{{ ch }}</span>
        <span v-if="fx === 'seal' && seal.length" class="mo-seal" :class="`n${seal.length}`" aria-hidden="true">
          <b v-for="(c, i) in seal" :key="i">{{ c }}</b>
        </span>
      </span>
    </p>
    <svg v-if="fx === 'ink'" class="mo-ink" viewBox="0 0 320 22" preserveAspectRatio="none" aria-hidden="true">
      <path d="M2 14.5C38 9.8 84 8.2 132 8.6c52 .4 101 2.6 150 4.3 13 .5 26 1.2 36 2.3-9 2.6-24 3.4-39 3.1-49-1.1-98-3.9-147-3.9-47 0-89 1.8-127 4.3-3 .2-4.4-3.1-1-3.6z" />
    </svg>
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

/* ---------- ink 墨痕：一笔毛笔横画自左向右写出，墨色取 --ak-ink ---------- */
.fx-ink { padding-bottom: 64px; }

.mo-ink {
  display: block;
  width: min(360px, 56%);
  height: 26px;
  margin: 12px auto 0;
  fill: var(--ak-ink);
  opacity: 0.85;
  clip-path: inset(0 100% 0 0);
  transition: clip-path 0.9s cubic-bezier(0.55, 0.1, 0.25, 1) var(--after);
}

.in .mo-ink { clip-path: inset(0 0 0 0); }

/* ---------- seal 印章：句末钤一方朱印（品牌红 --accent-red），盖下时由大到小落定；悬挂在句末，不参与居中 ---------- */
.mo-line { position: relative; }

.mo-seal {
  position: absolute;
  left: 100%;
  bottom: 0.14em;
  display: inline-grid;
  grid-template-columns: 1fr;
  place-items: center;
  width: 0.72em;
  height: 0.72em;
  margin-left: 0.22em;
  border-radius: 0.08em;
  background: var(--accent-red);
  color: #fff;
  font-size: 1em;
  letter-spacing: 0;
  opacity: 0;
  transform: rotate(-6deg) scale(1.6);
  transition: opacity 0.25s ease-out var(--after), transform 0.45s var(--ease-spring) var(--after);

  b { font: 700 0.46em/1 var(--font-serif); }

  /* 二至四字：竖排两列，右列先读 */
  &.n2 b { font-size: 0.3em; }
  &.n2 { grid-template-columns: 1fr; gap: 0.02em; }
  &.n3, &.n4 { grid-template-columns: 1fr 1fr; grid-auto-flow: column; grid-template-rows: 1fr 1fr; direction: rtl; padding: 0.05em; }
  &.n3 b, &.n4 b { font-size: 0.26em; }
}

.in .mo-seal { opacity: 0.92; transform: rotate(-6deg); }

/* ---------- quote 引号：一枚超大宋体引号淡淡压在句首上方 ---------- */
.fx-quote { padding-top: 76px; }

.mo-quote {
  position: absolute;
  left: 50%;
  top: 10px;
  translate: -50% 0;
  font: 700 150px/1 var(--font-serif);
  color: var(--line-2);
  pointer-events: none;
  opacity: 0;
  transform: translateY(-14px);
  transition: opacity 0.9s var(--ease-out) 0.1s, transform 0.9s var(--ease-out) 0.1s;
}

.in .mo-quote { opacity: 1; transform: none; }

/* ---------- card 卡片变体：左对齐小字号，不带装饰 ---------- */
.mo.card-v {
  padding: 8px 0 10px;
  text-align: left;

  p { font-size: 26px; letter-spacing: 0.04em; }
}

.fx-none { padding-bottom: 56px; }

@media (prefers-reduced-motion: reduce) {
  .mo *, .mo::before, .mo::after, .mo *::before { transition: none !important; }
  .mo .ch, .mo-quote { opacity: 1; filter: none; transform: none; }
  .fx-line::before, .fx-line::after, .mo-horizon i, .mo-horizon::before { transform: none; opacity: 1; }
  .mo-ink { clip-path: none; }
  .mo-seal { opacity: 0.92; transform: rotate(-6deg); }
}
</style>
