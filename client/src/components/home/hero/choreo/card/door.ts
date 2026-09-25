import { fade } from '../engine';
import { bezier, doorFan, poly } from '../geom';
import type { CardChoreo, ChoreoCtx } from '../types';
import { lightsIn, lightsOut } from './lights';

/**
 * 构造一扇门：preserve-3d 的门扇 = 正面（半幅画面 + 渐暗层）+ 自由边侧面（厚度）+ 背面（朝向门后光源，随开门渐亮）。
 * 门扇不做 opacity / filter / clip（它们会强制拍平 3D），淡出与变暗只作用在子面上。
 */
function buildLeaf(t: ChoreoCtx, card: HTMLElement, sheet: HTMLElement, side: 'l' | 'r') {
  const leaf = t.make(card, `hc-leaf hc-leaf-${side}`);
  const face = document.createElement('div');
  face.className = 'hc-leaf-face';
  const pic = document.createElement('div');
  pic.className = 'hc-leaf-pic';
  for (const node of Array.from(sheet.children)) pic.append(node.cloneNode(true));
  const shade = document.createElement('div');
  shade.className = 'hc-leaf-shade';
  face.append(pic, shade);
  const edge = document.createElement('div');
  edge.className = 'hc-leaf-edge';
  const back = document.createElement('div');
  back.className = 'hc-leaf-back';
  const lit = document.createElement('div');
  lit.className = 'hc-leaf-lit';
  back.append(lit);
  leaf.append(back, edge, face);
  return { leaf, face, shade, edge, back, lit };
}

/* ===== 开门角度曲线：门扇与地面光扇共用同一组采样，保证逐帧对齐 ===== */
const OPEN_MS = 820;
const MAX_DEG = 114;
const N = 28;
const quartInOut = bezier(0.76, 0, 0.24, 1);

/** 进度 p（0–1）→ 开门角：先裂开一道缝（前 20% 到 9°），再加速转出 */
function angleAt(p: number): number {
  const e = quartInOut(p);
  return e < 0.2 ? (9 * e) / 0.2 : 9 + ((MAX_DEG - 9) * (e - 0.2)) / 0.8;
}

/** 门扇（正面 / 厚度 / 背面）与地面光扇共用的淡出窗口：LEAF_MS 内从 LEAF_FADE_AT 开始线性淡出 */
const LEAF_MS = OPEN_MS + 120;
const LEAF_FADE_AT = 0.82;

const OFFSETS = Array.from({ length: N + 1 }, (_, i) => i / N);

/**
 * 地面光扇：一块与卡片同属门框 3D 空间的「地面」平面（卡片底边向观者方向放平），
 * 按门扇实时角度裁出受光梯形；两侧边 = 门扇底边，越过自由边后沿门后光源射线铺开；
 * 门扇外侧各有一道投影切断光扇。平面挂在前卡变换层下，透视与门扇同一套（卡片实际位置即几何来源）。
 */
function buildFan(t: ChoreoCtx, card: HTMLElement): void {
  const w = card.offsetWidth;
  const L = w * 1.3;
  const plane = t.make(card, 'hc-fan', {
    left: `${-w}px`,
    width: `${w * 3}px`,
    height: `${L}px`,
  });
  const lit = document.createElement('div');
  lit.className = 'hc-fan-lit';
  const sl = document.createElement('div');
  sl.className = 'hc-fan-sh';
  const sr = document.createElement('div');
  sr.className = 'hc-fan-sh';
  plane.append(sl, sr, lit);
  const ext = w * 1.25;
  const frames = OFFSETS.map((o) => doorFan(w, w, angleAt(o), ext));
  t.a(lit, frames.map((f, i) => ({ clipPath: poly(f.lit), offset: OFFSETS[i] })), { dur: OPEN_MS, ease: 'linear' });
  t.a(sl, frames.map((f, i) => ({ clipPath: poly(f.sl), offset: OFFSETS[i] })), { dur: OPEN_MS, ease: 'linear' });
  t.a(sr, frames.map((f, i) => ({ clipPath: poly(f.sr), offset: OFFSETS[i] })), { dur: OPEN_MS, ease: 'linear' });
  // 光扇强度：门缝裂开即有光，全开时最亮；与门扇同一时间窗、同一段淡出（门扇离场即熄灭）
  t.a(plane, [
    { opacity: 0 },
    { opacity: 0.55, offset: 0.12 },
    { opacity: 1, offset: 0.45 },
    { opacity: 1, offset: LEAF_FADE_AT },
    { opacity: 0 },
  ], { dur: LEAF_MS, ease: 'linear' });
}

/**
 * 推门见光（原 08 卡片部分，按定稿改为向外开）：
 * 旧封面从中线分成两扇门，以外缘为轴朝观者方向转出（门扇有厚度，背面朝向门后光源被照亮）；
 * 门后是下一篇（不过曝、无内部白光）；光从门洞底边投到地面，形成与门扇底边相接、随开门同步展开的光扇。
 */
export const door: CardChoreo = {
  meta: {
    id: 'door',
    name: '推门见光',
    tag: '叙事',
    desc: '封面像两扇门朝你打开，门洞底边投出一道随门展开的地面光扇。',
    duration: 1480,
  },
  exit(e, t) {
    lightsOut(e, t);
    const f = e.front;
    // 消失点在卡片中心：card 层 perspective 属性（filter 置空避免拍平）
    t.set(f.el, {
      perspective: `${Math.round(e.album.offsetWidth * 2.4)}px`,
      perspectiveOrigin: '50% 50%',
      filter: 'none',
    });
    t.set(f.sheet, { opacity: '0' });
    buildFan(t, f.el);
    (['l', 'r'] as const).forEach((side) => {
      const dir = side === 'l' ? -1 : 1;
      const p = buildLeaf(t, f.el, f.sheet, side);
      // 与光扇共用角度采样（linear 播放采样帧 = 同一条 quartInOut 曲线）
      t.a(p.leaf, OFFSETS.map((o) => ({ transform: `rotateY(${(dir * angleAt(o)).toFixed(2)}deg)`, offset: o })), { dur: OPEN_MS, ease: 'linear' });
      t.a(p.shade, [{ opacity: 0 }, { opacity: 0.2, offset: 0.2 }, { opacity: 0.72 }], { dur: OPEN_MS, ease: 'quartIn' });
      // 背面朝向门后光源：转过 90° 露出时已被照亮，靠门洞一侧最亮
      t.a(p.lit, [{ opacity: 0 }, { opacity: 0, offset: 0.55 }, { opacity: 1, offset: 0.8 }, { opacity: 0.85 }], { dur: OPEN_MS, ease: 'linear' });
      [p.face, p.edge, p.back].forEach((el) => t.a(el, [
        { opacity: 1 },
        { opacity: 1, offset: LEAF_FADE_AT },
        { opacity: 0 },
      ], { dur: LEAF_MS, ease: 'linear' }));
    });
    e.backs.forEach((c, i) => t.a(c.el, [
      { filter: e.F(c.b), opacity: 1 },
      { filter: e.F(c.b * 0.15), opacity: 0 },
    ], { dur: 420, delay: 60 + i * 40, ease: 'quartIn' }));
    t.a(e.bg, fade(1, 0), { dur: 420, ease: 'quartIn' });
  },
  enter(e, t) {
    const f = e.front;
    // 门后不再过曝（不要内部白光）：新封面从略暗、略远处随开门自然推近到静止态
    t.a(f.el, [
      { filter: e.F(0.72), transform: e.T(0, { ds: 0.95 }) },
      { filter: e.F(1), transform: e.T(0) },
    ], { dur: 1100, delay: 60, ease: 'quartInOut' });
    e.backs.forEach((c, i) => t.a(c.el, [
      { transform: e.T(c.slot, { dz: -80 }), filter: e.F(c.b * 0.6), opacity: 0 },
      { transform: e.T(c.slot), filter: e.F(c.b), opacity: 1 },
    ], { dur: 760, delay: 640 + i * 80 }));
    t.a(e.bg, fade(0, 1), { dur: 1000, delay: 280 });
    lightsIn(e, t, (c) => (c.slot === 0 ? 1160 : 1400 + (c.slot - 1) * 80));
  },
};
