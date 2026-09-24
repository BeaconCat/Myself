<script setup lang="ts">
import { computed } from 'vue';

/**
 * 无封面时的光影构成（品牌母题：门、光、书页），纯 CSS。
 * kind 缺省时按 seed 稳定取一种；src 存在时直接显示图片。
 */
const KINDS = ['door', 'pages', 'band', 'dawn', 'night', 'blind', 'paper'] as const;
type Kind = (typeof KINDS)[number];

const props = defineProps<{ src?: string; kind?: Kind; seed?: number | string; color?: string }>();

const k = computed<Kind>(() => {
  if (props.kind) return props.kind;
  const s = String(props.seed ?? 0);
  let h = 0;
  for (const ch of s) h = (h * 31 + ch.charCodeAt(0)) >>> 0;
  return KINDS[h % KINDS.length];
});
</script>

<template>
  <div class="cv" :class="src ? 'img' : `cv-${k}`" :style="color ? { '--c': color } : undefined">
    <img v-if="src" :src="src" alt="" loading="lazy" draggable="false" />
    <template v-else-if="k === 'door'"><span class="glow" /><span class="beam"><i /></span><span class="slit" /><span class="noise" /></template>
    <template v-else-if="k === 'pages'"><span class="beam"><i /></span><span class="pg l" /><span class="pg r" /><span class="noise" /></template>
    <template v-else-if="k === 'band'"><span class="b3" /><span class="b1" /><span class="b2" /><span class="noise" /></template>
    <template v-else-if="k === 'dawn'"><span class="rays" /><span class="sun" /><span class="hz" /><span class="noise" /></template>
    <template v-else-if="k === 'paper'"><span class="warm" /><span class="sheet s1" /><span class="sheet s2" /><span class="shade" /><span class="noise" /></template>
    <template v-else-if="k === 'blind'"><span class="bl" /><span class="leaf" /><span class="sill" /><span class="noise" /></template>
    <template v-else><span class="st" /><span class="hill" /><span class="win" /><span class="noise" /></template>
    <slot />
  </div>
</template>

<style scoped lang="scss">
.cv {
  position: relative;
  overflow: hidden;
  isolation: isolate;
  background: #0a0f1a;
  --c: var(--primary);

  > span { position: absolute; pointer-events: none; }

  &.img {
    background: var(--well-2);

    img { position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover; display: block; }
  }
}

.noise {
  inset: 0;
  opacity: 0.22;
  mix-blend-mode: overlay;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='160' height='160'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.9' numOctaves='2' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");
}

/* 门缝光 */
.cv-door {
  background: radial-gradient(90% 70% at 50% 108%, color-mix(in oklab, var(--c) 42%, #0a0f1a), #060910 70%);

  .glow { left: 30%; right: 30%; top: 5%; height: 85%; background: radial-gradient(closest-side, color-mix(in oklab, var(--c) 55%, transparent), transparent); filter: blur(18px); opacity: 0.8; }
  .slit { left: 46.5%; width: 7%; top: 16%; height: 52%; border-radius: 3px 3px 1px 1px; background: linear-gradient(#fff, color-mix(in oklab, var(--c) 30%, #fff)); box-shadow: 0 0 18px 2px color-mix(in oklab, var(--c) 70%, #fff), 0 0 60px 10px color-mix(in oklab, var(--c) 40%, transparent); }
  .beam { left: 0; right: 0; top: 68%; bottom: 0; filter: blur(5px); }
  .beam i { position: absolute; inset: 0; clip-path: polygon(46.5% 0, 53.5% 0, 88% 100%, 22% 100%); background: linear-gradient(rgba(255, 255, 255, 0.95), color-mix(in oklab, var(--c) 45%, transparent) 55%, transparent); }
}

/* 双页门 */
.cv-pages {
  background: radial-gradient(80% 60% at 60% 110%, color-mix(in oklab, var(--c) 30%, #0a0f1a), #05080e 65%);

  .pg { top: 18%; height: 56%; width: 17%; border-radius: 10px 6px 6px 10px; }
  .pg.l { left: 31%; background: linear-gradient(100deg, #fff 60%, #e6e8ee); transform: perspective(300px) rotateY(24deg); box-shadow: 0 0 40px -6px rgba(255, 255, 255, 0.4); }
  .pg.r { left: 52%; background: linear-gradient(160deg, color-mix(in oklab, var(--c) 80%, #fff), var(--c) 45%, color-mix(in oklab, var(--c) 70%, #000)); transform: perspective(300px) rotateY(-24deg); border-radius: 6px 10px 10px 6px; }
  .beam { left: 0; right: 0; top: 70%; bottom: 0; filter: blur(7px); }
  .beam i { position: absolute; inset: 0; clip-path: polygon(36% 0, 50% 0, 100% 100%, 10% 100%); background: linear-gradient(rgba(230, 236, 255, 0.85), color-mix(in oklab, var(--c) 40%, transparent) 60%, transparent); }
}

/* 光带 */
.cv-band {
  background: linear-gradient(160deg, #0d1322, #05070c);

  .b1 { left: -20%; right: -20%; top: 40%; height: 26%; transform: rotate(-24deg); background: linear-gradient(90deg, transparent, color-mix(in oklab, var(--c) 70%, #fff) 45%, #fff 52%, color-mix(in oklab, var(--c) 70%, transparent) 60%, transparent); filter: blur(16px); opacity: 0.9; }
  .b2 { left: -20%; right: -20%; top: 62%; height: 7%; transform: rotate(-24deg); background: linear-gradient(90deg, transparent 20%, rgba(255, 255, 255, 0.9) 50%, transparent 80%); filter: blur(3px); opacity: 0.7; }
  .b3 { inset: 0; background: radial-gradient(60% 50% at 80% 20%, color-mix(in oklab, var(--c) 30%, transparent), transparent); }
}

/* 晨光 */
.cv-dawn {
  background: linear-gradient(180deg, #0b1224 0%, color-mix(in oklab, var(--c) 40%, #1a1030) 58%, color-mix(in oklab, #ffb300 70%, var(--c)) 74%, #120d10 74.4%, #07070b);

  .sun { left: 50%; top: 74%; width: 46%; aspect-ratio: 1; transform: translate(-50%, -50%); border-radius: 50%; background: radial-gradient(closest-side, #fff, color-mix(in oklab, #ffb300 70%, #fff) 30%, color-mix(in oklab, #ffb300 60%, transparent) 55%, transparent); filter: blur(4px); }
  .rays { inset: 0; background: repeating-conic-gradient(from 0deg at 50% 74%, rgba(255, 255, 255, 0.07) 0 4deg, transparent 4deg 11deg); mask: radial-gradient(60% 60% at 50% 74%, #000, transparent); }
  .hz { left: 0; right: 0; top: 74%; bottom: 0; background: linear-gradient(color-mix(in oklab, #ffb300 40%, transparent), transparent 50%); opacity: 0.5; }
}

/* 窗光下的纸 */
.cv-paper {
  background: linear-gradient(135deg, #efe7da, #dcd0bd);

  .sheet { background: #fbf9f5; border-radius: 3px; box-shadow: 0 1px 1px rgba(80, 60, 30, 0.1), 0 12px 24px -8px rgba(80, 60, 30, 0.35); }
  .s1 { left: 26%; top: 18%; width: 36%; height: 70%; transform: rotate(-7deg); }
  .s2 { left: 40%; top: 14%; width: 36%; height: 70%; transform: rotate(4deg); background: linear-gradient(#fdfcf9, #f4efe6); }
  .s2::before { content: ''; position: absolute; left: 14%; right: 14%; top: 16%; height: 62%; background: repeating-linear-gradient(transparent 0 9px, rgba(80, 60, 30, 0.12) 9px 10px); }
  .shade { inset: -20%; background: repeating-linear-gradient(115deg, rgba(50, 30, 0, 0) 0 22px, rgba(50, 30, 0, 0.16) 22px 36px); filter: blur(5px); mix-blend-mode: multiply; }
  .warm { inset: 0; background: radial-gradient(70% 60% at 20% 10%, color-mix(in oklab, #ffb300 40%, transparent), transparent); mix-blend-mode: soft-light; }
}

/* 百叶窗光 */
.cv-blind {
  background: linear-gradient(165deg, color-mix(in oklab, var(--c) 12%, #c9b99d), color-mix(in oklab, var(--c) 16%, #8a7a62));

  .bl { inset: -30%; background: repeating-linear-gradient(-34deg, transparent 0 18px, rgba(255, 236, 196, 0.6) 18px 34px); filter: blur(1.5px); mask: radial-gradient(50% 55% at 62% 42%, #000 35%, transparent 78%); mix-blend-mode: screen; }
  .leaf { left: -4%; bottom: -12%; width: 44%; height: 80%; background: radial-gradient(40% 30% at 40% 30%, rgba(40, 34, 20, 0.35), transparent 70%), radial-gradient(30% 25% at 70% 60%, rgba(40, 34, 20, 0.3), transparent 70%), radial-gradient(35% 25% at 30% 75%, rgba(40, 34, 20, 0.3), transparent 70%); filter: blur(4px); }
  .sill { left: 0; right: 0; bottom: 0; height: 18%; background: linear-gradient(rgba(255, 250, 240, 0), rgba(60, 45, 25, 0.18)); }
}

/* 地平线夜窗 */
.cv-night {
  background: radial-gradient(120% 80% at 50% 120%, color-mix(in oklab, var(--c) 35%, #0b1220), #05070d 60%);

  .st { inset: 0; background-image: radial-gradient(1px 1px at 20% 30%, #fff, transparent), radial-gradient(1px 1px at 70% 20%, #fff, transparent), radial-gradient(1.5px 1.5px at 45% 12%, #fff, transparent), radial-gradient(1px 1px at 85% 45%, #fff, transparent), radial-gradient(1px 1px at 10% 60%, #fff, transparent), radial-gradient(1px 1px at 60% 50%, rgba(255, 255, 255, 0.7), transparent); opacity: 0.8; }
  .win { left: 58%; top: 52%; width: 9%; height: 22%; background: linear-gradient(#fff5d6, color-mix(in oklab, #ffb300 70%, #fff)); box-shadow: 0 0 30px 6px color-mix(in oklab, #ffb300 50%, transparent); border-radius: 2px; }
  .hill { left: -10%; right: -10%; top: 66%; bottom: -40%; border-radius: 50% 50% 0 0; background: #04060b; }
}
</style>
