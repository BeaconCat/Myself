<script lang="ts">
/** CSS 光影封面种类（门、门缝、光带、四季、光谱弧、书页、信号、钥匙孔、光栅） */
export const COVER_KINDS = ['door', 'slit', 'beams', 'season', 'arcs', 'page', 'signal', 'key', 'grid'] as const;
export type CoverKind = (typeof COVER_KINDS)[number];

const PARTS: Record<CoverKind, string[]> = {
  door: ['halo', 'floor', 'leaf-l', 'leaf-r', 'gap'],
  slit: ['wall', 'fan', 'line', 'dust'],
  beams: ['b b3', 'b b1', 'b b2', 'win', 'pool'],
  season: ['o o1', 'o o2', 'o o3', 'o o4', 'disc', 'half'],
  arcs: ['a a3', 'a a2', 'a a1', 'sun'],
  page: ['ray', 'shadow', 'pg pl', 'pg pr', 'spine', 'mark'],
  signal: ['arc r4', 'arc r3', 'arc r2', 'arc r1', 'tower', 'lamp'],
  key: ['spill', 'plate', 'beam', 'hole', 'slot'],
  grid: ['floor', 'sun', 'horizon', 'pkt'],
};

/**
 * 封面源 → CSS 光影种类：
 * `css:<kind>` 显式指定；无图时的 SVG 渐变占位（data:image/svg+xml）按内容哈希稳定映射，
 * 真实图片返回 null。
 */
export function coverKindOf(src: string): CoverKind | null {
  if (src.startsWith('css:')) {
    const k = src.slice(4) as CoverKind;
    return COVER_KINDS.includes(k) ? k : 'door';
  }
  if (!src || src.startsWith('data:image/svg+xml')) {
    let h = 0;
    for (let i = 0; i < src.length; i++) h = (h * 31 + src.charCodeAt(i)) | 0;
    return COVER_KINDS[Math.abs(h) % COVER_KINDS.length];
  }
  return null;
}
</script>

<script setup lang="ts">
const props = defineProps<{ kind: CoverKind }>();
</script>

<template>
  <div class="art" :class="`cv-${props.kind}`" aria-hidden="true">
    <i v-for="p in PARTS[props.kind]" :key="p" :class="p" />
  </div>
</template>

<style scoped lang="scss">
/* 按 % 定位：卡片、缩略图、lightbox 同一份构成等比适配 */
.art {
  position: absolute;
  inset: 0;
  overflow: hidden;
  background: radial-gradient(120% 90% at 50% 112%, #0d2247 0%, #050c1a 58%, #02050c 100%);

  i { position: absolute; display: block; }
}

/* 门：左白页 + 右钴蓝页 + 地面透光（呼应 logo） */
.cv-door {
  .halo { inset: 8% 22% 18%; background: radial-gradient(closest-side, rgba(90, 150, 255, 0.35), transparent); filter: blur(8px); }
  .leaf-l { left: 30%; top: 15%; width: 19%; height: 63%; border-radius: 3%; background: linear-gradient(180deg, #fff, #e6ecf6 60%, #b8c6de); clip-path: polygon(0 3%, 100% 11%, 100% 89%, 0 100%); }
  .leaf-r { left: 51%; top: 15%; width: 19%; height: 63%; background: linear-gradient(165deg, #3a95ff, #0b5ee0 55%, #0842a8); clip-path: polygon(0 11%, 100% 0, 100% 97%, 0 89%); }
  .gap { left: 49%; top: 22%; width: 2%; height: 50%; background: linear-gradient(180deg, rgba(255, 255, 255, 0.2), rgba(200, 225, 255, 0.95)); filter: blur(2px); }
  .floor { left: 6%; right: 6%; top: 70%; bottom: -4%; background: linear-gradient(180deg, rgba(225, 238, 255, 0.9), rgba(120, 170, 255, 0.28) 55%, transparent); clip-path: polygon(44% 0, 56% 0, 100% 100%, 0 100%); filter: blur(5px); }
}

/* 门缝：黑暗中一道竖光 + 光扇 */
.cv-slit {
  background: radial-gradient(60% 80% at 50% 50%, #0a1830 0%, #03060d 70%);

  .wall { inset: 0; background: radial-gradient(18% 55% at 50% 42%, rgba(80, 140, 255, 0.28), transparent 70%); }
  .line { left: 49.6%; top: 12%; width: 0.8%; height: 60%; background: linear-gradient(180deg, rgba(255, 255, 255, 0.3), #fff 30%, #fff); box-shadow: 0 0 10px 2px rgba(170, 205, 255, 0.9), 0 0 42px 10px rgba(0, 120, 255, 0.55); border-radius: 2px; }
  .fan { left: 0; right: 0; top: 71%; bottom: 0; background: linear-gradient(180deg, rgba(235, 244, 255, 0.85), rgba(80, 140, 255, 0.2) 60%, transparent); clip-path: polygon(49.3% 0, 50.7% 0, 88% 100%, 12% 100%); filter: blur(3px); }
  .dust { inset: 0; background-image: radial-gradient(1px 1px at 46% 30%, rgba(255, 255, 255, 0.7), transparent), radial-gradient(1px 1px at 55% 50%, rgba(255, 255, 255, 0.5), transparent), radial-gradient(1.5px 1.5px at 52% 22%, rgba(255, 255, 255, 0.6), transparent), radial-gradient(1px 1px at 44% 60%, rgba(255, 255, 255, 0.4), transparent); }
}

/* 体积光带：高窗斜射 */
.cv-beams {
  background: linear-gradient(160deg, #0b1a33 0%, #050b18 60%, #03060d 100%);

  .win { left: 10%; top: 6%; width: 16%; height: 22%; background: linear-gradient(135deg, #fff, #cfe0ff); box-shadow: 0 0 30px 8px rgba(160, 200, 255, 0.5); transform: skewY(-6deg); }
  .b { left: -10%; top: 10%; width: 150%; height: 13%; transform-origin: 0 50%; background: linear-gradient(90deg, rgba(230, 240, 255, 0.55), rgba(120, 170, 255, 0.12) 60%, transparent); filter: blur(7px); mix-blend-mode: screen; }
  .b1 { transform: rotate(28deg); }
  .b2 { transform: rotate(35deg); top: 16%; height: 9%; opacity: 0.7; }
  .b3 { transform: rotate(22deg); top: 4%; height: 7%; opacity: 0.55; }
  .pool { right: 4%; bottom: 6%; width: 58%; height: 26%; border-radius: 50%; background: radial-gradient(closest-side, rgba(210, 228, 255, 0.55), rgba(0, 120, 255, 0.12) 60%, transparent); filter: blur(4px); }
}

/* 四季光晕 */
.cv-season {
  background: radial-gradient(90% 90% at 50% 50%, #0d1424, #04060c);

  .o { width: 46%; aspect-ratio: 1; border-radius: 50%; filter: blur(16px); mix-blend-mode: screen; opacity: 0.85; }
  .o1 { left: 12%; top: 12%; background: radial-gradient(closest-side, #00c853, transparent); }
  .o2 { left: 42%; top: 6%; background: radial-gradient(closest-side, #ff0032, transparent); }
  .o3 { left: 18%; top: 44%; background: radial-gradient(closest-side, #ffb300, transparent); }
  .o4 { left: 46%; top: 40%; background: radial-gradient(closest-side, #0078ff, transparent); }
  .disc { left: 30%; top: 17%; width: 40%; aspect-ratio: 1; border-radius: 50%; background: radial-gradient(circle at 35% 30%, rgba(255, 255, 255, 0.35), rgba(255, 255, 255, 0.04) 45%, rgba(0, 0, 0, 0.25)); box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.28), 0 20px 40px rgba(0, 0, 0, 0.4); }
  .half { left: 30%; top: 17%; width: 40%; aspect-ratio: 1; border-radius: 50%; background: linear-gradient(90deg, rgba(255, 255, 255, 0.18) 50%, rgba(0, 0, 0, 0.35) 50%); mix-blend-mode: overlay; }
}

/* 光谱弧 */
.cv-arcs {
  background: radial-gradient(80% 90% at 50% 100%, #10213f, #03060d 70%);

  .a { left: 50%; top: 100%; border-radius: 50%; border-style: solid; border-color: transparent; }
  .a1 { width: 44%; aspect-ratio: 1; border-width: 5px; border-top-color: #ff0032; border-left-color: #ff0032; filter: drop-shadow(0 0 8px rgba(255, 0, 50, 0.6)); transform: translate(-50%, -50%) rotate(8deg); }
  .a2 { width: 66%; aspect-ratio: 1; border-width: 5px; border-top-color: #ffb300; border-right-color: #ffb300; filter: drop-shadow(0 0 8px rgba(255, 179, 0, 0.55)); transform: translate(-50%, -50%) rotate(-12deg); }
  .a3 { width: 90%; aspect-ratio: 1; border-width: 5px; border-top-color: #0078ff; border-left-color: #0078ff; filter: drop-shadow(0 0 10px rgba(0, 120, 255, 0.7)); transform: translate(-50%, -50%) rotate(30deg); }
  .sun { left: 50%; bottom: -6%; width: 22%; aspect-ratio: 1; border-radius: 50%; transform: translateX(-50%); background: radial-gradient(closest-side, #fff, rgba(255, 255, 255, 0.5) 40%, rgba(0, 120, 255, 0.2) 70%, transparent); }
}

/* 书页：摊开的书 + 斜射一道光 */
.cv-page {
  background: radial-gradient(90% 80% at 50% 40%, #13213b, #05090f 75%);

  .pg { top: 20%; width: 34%; height: 60%; background: linear-gradient(90deg, #cfd8e6, #f4f7fb 70%, #e6ebf3); }
  .pg::after { content: ''; position: absolute; inset: 14% 12% 18%; background: repeating-linear-gradient(180deg, transparent 0 7px, rgba(40, 60, 100, 0.2) 7px 8px); }
  .pl { left: 16%; transform: perspective(500px) rotateY(16deg); transform-origin: 100% 50%; }
  .pr { left: 50%; transform: perspective(500px) rotateY(-16deg); transform-origin: 0 50%; background: linear-gradient(270deg, #cfd8e6, #f4f7fb 70%, #e9eef5); }
  .spine { left: 49.4%; top: 19%; width: 1.2%; height: 62%; background: linear-gradient(90deg, rgba(0, 0, 0, 0.35), rgba(0, 0, 0, 0.05), rgba(0, 0, 0, 0.35)); }
  .ray { left: -20%; top: -30%; width: 60%; height: 170%; transform: rotate(-28deg); background: linear-gradient(90deg, transparent, rgba(200, 225, 255, 0.22) 45%, rgba(255, 255, 255, 0.35) 50%, rgba(200, 225, 255, 0.22) 55%, transparent); filter: blur(6px); mix-blend-mode: screen; }
  .mark { left: 60%; top: 17%; width: 3%; height: 26%; background: linear-gradient(180deg, #0078ff, #0050c0); clip-path: polygon(0 0, 100% 0, 100% 100%, 50% 86%, 0 100%); }
  .shadow { left: 12%; right: 12%; top: 78%; height: 12%; border-radius: 50%; background: radial-gradient(closest-side, rgba(0, 0, 0, 0.6), transparent); }
}

/* 信号塔：红蓝同心弧 + 灯 */
.cv-signal {
  background: radial-gradient(80% 80% at 50% 70%, #0f1e3a, #03060d 72%);

  .arc { top: 50%; left: 50%; border-radius: 50%; border: 5px solid transparent; transform: translate(-50%, -50%); }
  .r1 { width: 26%; aspect-ratio: 1; border-left-color: #ff3355; border-right-color: #ff3355; filter: drop-shadow(0 0 6px rgba(255, 0, 50, 0.6)); }
  .r2 { width: 44%; aspect-ratio: 1; border-left-color: #d42a45; border-right-color: #d42a45; opacity: 0.9; }
  .r3 { width: 64%; aspect-ratio: 1; border-left-color: #1a6fff; border-right-color: #1a6fff; filter: drop-shadow(0 0 8px rgba(0, 120, 255, 0.6)); }
  .r4 { width: 86%; aspect-ratio: 1; border-left-color: #0a4fc0; border-right-color: #0a4fc0; opacity: 0.7; }
  .tower { left: 45%; top: 52%; width: 10%; height: 44%; background: linear-gradient(90deg, #8594ad, #c9d3e3 50%, #8594ad); clip-path: polygon(30% 0, 70% 0, 100% 100%, 0 100%); }
  .lamp { left: 50%; top: 50%; width: 9%; aspect-ratio: 1; border-radius: 50%; transform: translate(-50%, -50%); background: radial-gradient(circle, #fff, #ffe6a0 45%, rgba(255, 179, 0, 0.2) 70%, transparent); box-shadow: 0 0 30px 10px rgba(255, 200, 80, 0.35); }
}

/* 钥匙孔：APIKey 的光 */
.cv-key {
  background: radial-gradient(70% 70% at 50% 45%, #111e36, #04070d 70%);

  .plate { left: 33%; top: 12%; width: 34%; height: 76%; border-radius: 14px; background: linear-gradient(160deg, #1d2a42, #0b1220); box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.12), 0 20px 40px rgba(0, 0, 0, 0.5); }
  .spill { left: 20%; top: 18%; width: 60%; height: 60%; background: radial-gradient(closest-side, rgba(0, 120, 255, 0.35), transparent); filter: blur(10px); }
  .hole { left: 50%; top: 32%; width: 10%; aspect-ratio: 1; border-radius: 50%; transform: translateX(-50%); background: #fff; box-shadow: 0 0 14px 3px rgba(190, 215, 255, 0.9), 0 0 50px 14px rgba(0, 120, 255, 0.5); }
  .slot { left: 50%; top: 42%; width: 7%; height: 20%; transform: translateX(-50%); background: linear-gradient(180deg, #fff, #dbe8ff); clip-path: polygon(30% 0, 70% 0, 100% 100%, 0 100%); box-shadow: 0 0 20px rgba(190, 215, 255, 0.8); }
  .beam { left: 30%; right: 30%; top: 62%; bottom: 0; background: linear-gradient(180deg, rgba(220, 235, 255, 0.5), transparent); clip-path: polygon(45% 0, 55% 0, 100% 100%, 0 100%); filter: blur(4px); }
}

/* 光栅地平线：数据在光里流动 */
.cv-grid {
  background: linear-gradient(180deg, #03060d 0%, #081328 52%, #0a1a38 53%, #03060d 100%);

  .floor { left: -30%; right: -30%; top: 53%; height: 90%; transform: perspective(220px) rotateX(62deg); transform-origin: 50% 0; background-image: linear-gradient(rgba(0, 120, 255, 0.55) 1px, transparent 1px), linear-gradient(90deg, rgba(0, 120, 255, 0.55) 1px, transparent 1px); background-size: 7% 12%; mask-image: linear-gradient(180deg, #000, transparent 70%); }
  .horizon { left: 0; right: 0; top: 52%; height: 2%; background: linear-gradient(90deg, transparent, #7fb3ff 30%, #fff 50%, #7fb3ff 70%, transparent); filter: blur(1px); box-shadow: 0 0 30px 6px rgba(0, 120, 255, 0.5); }
  .sun { left: 50%; top: 52%; width: 38%; aspect-ratio: 2; border-radius: 999px 999px 0 0; transform: translate(-50%, -100%); background: linear-gradient(180deg, #ff0032, #ff7a1a 60%, #ffb300); mask-image: repeating-linear-gradient(180deg, #000 0 9%, transparent 9% 12%); opacity: 0.9; }
  .pkt { left: 62%; top: 70%; width: 5%; height: 3%; background: #fff; box-shadow: 0 0 12px 4px rgba(120, 180, 255, 0.9); transform: skewX(-30deg); }
}
</style>
