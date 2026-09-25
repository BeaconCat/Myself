<script lang="ts">
/** 服务端渐变占位图（/api/v1/img/...）视同无图：改走光影构成，避免纯色块 */
export function isArtUrl(url: string | undefined): boolean {
  return !url || url.startsWith('/api/v1/img/');
}
</script>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { thumbOf } from '../../api';

/**
 * 封面：有真实封面用图片（可选缩略图），没有时按 seed 稳定选一幅 CSS 光影构成
 * （门缝光 / 光谱扇 / 书页台灯 / 光束信号环；pool="all" 时追加地平线 / 窗格，供随想配图），
 * 与品牌「门、光、书页」母题一致。
 * 注：本组件是全站唯一允许 blur / glow 的地方（品牌封面光影），控件不得借用。
 */
const props = withDefaults(
  defineProps<{ src?: string; seed: string; thumb?: boolean; pool?: 'post' | 'all' }>(),
  { src: '', thumb: false, pool: 'post' },
);

const POST_VARIANTS = ['door', 'bands', 'pages', 'beam'] as const;
const ALL_VARIANTS = [...POST_VARIANTS, 'dusk', 'pane'] as const;

const variant = computed(() => {
  const list = props.pool === 'all' ? ALL_VARIANTS : POST_VARIANTS;
  let h = 0;
  for (const ch of props.seed) h = (h * 31 + ch.charCodeAt(0)) >>> 0;
  return list[h % list.length];
});

const url = computed(() => {
  if (isArtUrl(props.src)) return '';
  return props.thumb ? thumbOf(props.src) : props.src;
});
const loaded = ref(false);
watch(url, () => { loaded.value = false; });
</script>

<template>
  <div v-if="url" class="cv cv-img" :class="{ loaded }">
    <img :src="url" alt="" draggable="false" decoding="async" @load="loaded = true" />
  </div>
  <div v-else class="cv" :class="`cv-${variant}`" aria-hidden="true">
    <template v-if="variant === 'door'"><i class="halo" /><i class="floor" /><i class="l" /><i class="r" /></template>
    <template v-else-if="variant === 'bands'"><i class="fan" /><i class="floor" /><i class="horizon" /><i class="core" /></template>
    <template v-else-if="variant === 'pages'"><i class="lamp" /><i class="pg p2" /><i class="pg p1" /></template>
    <template v-else-if="variant === 'beam'"><i class="rings" /><i class="ray" /><i class="gap" /></template>
    <template v-else-if="variant === 'dusk'"><i class="sun" /><i class="refl" /></template>
    <template v-else><i class="wedge" /><i class="slit" /></template>
  </div>
</template>

<style scoped lang="scss">
.cv {
  position: absolute;
  inset: 0;
  overflow: hidden;
  background: #040914;
  isolation: isolate;
  container-type: size;
  /* 父级悬停时可对 .cv 做轻微放大 */
  transition: transform var(--dur-slow) var(--ease-out);

  i { position: absolute; display: block; }
}

.cv-img {
  background: var(--m-fill-2, var(--fill-2));

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    opacity: 0;
    transform: scale(1.04);
    transition: opacity 0.5s var(--ease-out), transform 0.8s var(--ease-out);
  }

  &.loaded img {
    opacity: 1;
    transform: none;
  }
}

/* 门：左白页 + 右钴蓝页 + 地面透出的光 */
.cv-door {
  background: radial-gradient(70% 48% at 50% 80%, #16306a 0%, transparent 70%), linear-gradient(#060d1d, #02050d);

  .halo { left: 25%; right: 25%; top: 12%; bottom: 30%; background: radial-gradient(closest-side, rgba(90, 140, 255, 0.35), transparent); filter: blur(4cqw); }
  .l, .r { top: 20%; height: 50%; width: 21%; border-radius: 3.5cqw; }
  .l { left: 29%; background: linear-gradient(180deg, #fff, #dde5f4); transform: perspective(70cqw) rotateY(34deg); transform-origin: left center; box-shadow: 0 0 10cqw rgba(190, 210, 255, 0.3); }
  .r { right: 29%; background: linear-gradient(200deg, #4094ff, #0b58e6 50%, #0535a0); transform: perspective(70cqw) rotateY(-34deg); transform-origin: right center; }
  .floor { left: 12%; right: -14%; top: 69%; bottom: -6%; background: linear-gradient(180deg, rgba(240, 244, 255, 0.95), rgba(150, 170, 232, 0.5) 42%, rgba(50, 70, 150, 0) 92%); clip-path: polygon(35% 0, 57% 0, 100% 100%, 0 100%); filter: blur(1.1cqw); }
}

/* 光谱扇：四季四束光 */
.cv-bands {
  background: radial-gradient(90% 60% at 50% 104%, #0f2250, transparent 70%), #02060f;

  .fan { inset: -12%; background: conic-gradient(from -44deg at 50% 100%, transparent 0 6deg, rgba(0, 200, 83, 0.9) 12deg, transparent 20deg 26deg, rgba(255, 0, 50, 0.85) 32deg, transparent 40deg 46deg, rgba(255, 179, 0, 0.9) 52deg, transparent 60deg 66deg, rgba(0, 120, 255, 0.95) 72deg, transparent 80deg 360deg); filter: blur(2.6cqw); mix-blend-mode: screen; }
  .core { left: 38%; right: 38%; bottom: 9%; height: 8%; background: radial-gradient(closest-side, #fff, rgba(255, 255, 255, 0)); filter: blur(1cqw); }
  .horizon { left: -5%; right: -5%; top: 86%; height: 0.8cqw; background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.85) 50%, transparent); filter: blur(0.4cqw); }
  .floor { left: 0; right: 0; top: 87%; bottom: 0; background: linear-gradient(180deg, rgba(40, 70, 140, 0.55), #02060f); }
}

/* 书页：台灯光锥照在两页纸上 */
.cv-pages {
  background: radial-gradient(60% 50% at 62% 22%, rgba(255, 179, 0, 0.28), transparent 64%), linear-gradient(#0b1428, #03070f);

  .lamp { top: -12%; left: 40%; width: 48%; height: 92%; background: linear-gradient(180deg, rgba(255, 214, 140, 0.55), rgba(255, 200, 120, 0) 90%); clip-path: polygon(44% 0, 56% 0, 100% 100%, 0 100%); filter: blur(2.2cqw); }
  .pg { left: 18%; top: 24%; width: 62%; height: 40%; border-radius: 1.6cqw; transform: perspective(90cqw) rotateX(54deg) rotateZ(-15deg); box-shadow: 0 6cqw 10cqw rgba(0, 0, 0, 0.6); }
  .p2 { background: linear-gradient(160deg, #cbd5ea, #8fa1c6); transform: perspective(90cqw) rotateX(54deg) rotateZ(-3deg) translate(8%, -14%); }
  .p1 { background: repeating-linear-gradient(180deg, transparent 0 8%, rgba(20, 30, 55, 0.2) 8% 9.2%), linear-gradient(160deg, #fff, #e2e7f1); background-clip: padding-box; border: 3cqw solid transparent; }
}

/* 门缝光束 + 信号环 */
.cv-beam {
  background: radial-gradient(60% 50% at 30% 58%, #0a2250, transparent 70%), #02060e;

  .rings { inset: -10%; background: repeating-radial-gradient(circle at 34% 58%, transparent 0 10cqw, rgba(0, 120, 255, 0.5) 10cqw 10.8cqw, transparent 10.8cqw 16cqw, rgba(255, 0, 50, 0.4) 16cqw 16.6cqw, transparent 16.6cqw 22cqw); mask: radial-gradient(circle at 34% 58%, #000 18%, transparent 62%); }
  .ray { left: 30%; top: 30%; width: 110%; height: 56%; background: linear-gradient(90deg, rgba(220, 232, 255, 0.8), rgba(0, 120, 255, 0.18) 55%, transparent 85%); clip-path: polygon(0 42%, 100% 0, 100% 100%, 0 58%); filter: blur(1.6cqw); }
  .gap { left: 28%; top: 12%; width: 5%; height: 46%; border-radius: 1.2cqw; background: linear-gradient(#fff, #b8d0ff); box-shadow: 0 0 6cqw 1.5cqw rgba(120, 170, 255, 0.7), 0 0 22cqw 5cqw rgba(0, 120, 255, 0.3); }
}

/* 地平线：落日与水面倒影（随想配图） */
.cv-dusk {
  background: linear-gradient(180deg, #0a1530 0, #1a2d5a 44%, #f0a45b 55%, #0b1a33 55.4%, #081226 100%);

  .sun { left: 50%; top: 55%; width: 26%; aspect-ratio: 1; translate: -50% -50%; border-radius: 50%; background: radial-gradient(circle, #fff5dc, #ffb300 40%, rgba(255, 120, 40, 0) 70%); clip-path: inset(0 0 50% 0); }
  .refl { left: 44%; right: 44%; top: 57%; bottom: 8%; background: repeating-linear-gradient(180deg, rgba(255, 200, 120, 0.8) 0 2px, transparent 2px 7px); mask: linear-gradient(#000, transparent); filter: blur(0.6px); }
}

/* 窗格：暗室里一道暖光缝 */
.cv-pane {
  background: radial-gradient(70% 55% at 50% 100%, rgba(255, 170, 60, 0.22), transparent 70%), linear-gradient(180deg, #120d14, #07060c);

  .slit { left: 47%; width: 6%; top: 10%; height: 62%; border-radius: 0.8cqw; background: linear-gradient(180deg, #fff3d6, #ffc15a 55%, #ff9d2e); box-shadow: 0 0 5cqw 1cqw rgba(255, 180, 80, 0.55), 0 0 20cqw 4cqw rgba(255, 150, 40, 0.22); }
  .wedge { left: -10%; right: -10%; top: 72%; bottom: -4%; background: linear-gradient(180deg, rgba(255, 226, 170, 0.85), rgba(255, 170, 60, 0.28) 45%, transparent 90%); clip-path: polygon(47% 0, 53% 0, 82% 100%, 18% 100%); filter: blur(0.8cqw); }
}
</style>
