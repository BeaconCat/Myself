/**
 * 卡组几何：0 前排 / 1 右上后排 / 2 左下后排。
 *
 * 所有关键帧与静止态共用同一「规范变换函数表」，保证逐函数插值、不走矩阵分解：
 *   translate(px,py) rotate(pr) | perspective rotateY rotateX translate3d rotateZ scale
 * 前两项为屏幕空间（甩出 / 抛物线），其后为卡片自身 3D 姿态。
 * 偏移量按相册实际宽度 / 设计基准宽度等比缩放。
 */

export const PERSP = 1300;

interface Slot {
  x: number;
  y: number;
  z: number;
  s: number;
  /** 静止亮度 */
  b: number;
}

const GEOM: Record<'desk' | 'mobile', { base: number; angle: number; slots: Slot[] }> = {
  desk: {
    base: 540,
    angle: -15,
    slots: [
      { x: -18, y: 0, z: 0, s: 0.86, b: 1 },
      { x: 76, y: -46, z: -90, s: 0.8, b: 0.8 },
      { x: -140, y: 96, z: -90, s: 0.74, b: 0.72 },
    ],
  },
  mobile: {
    base: 300,
    angle: -10,
    slots: [
      { x: -8, y: 0, z: 0, s: 0.86, b: 1 },
      { x: 44, y: -26, z: -60, s: 0.8, b: 0.8 },
      { x: -68, y: 54, z: -60, s: 0.74, b: 0.72 },
    ],
  },
};

export interface TfParts {
  px?: number;
  py?: number;
  pr?: number;
  ry?: number;
  rx?: number;
  x?: number;
  y?: number;
  z?: number;
  rz?: number;
  sx?: number;
  sy?: number;
}

const n = (v?: number): number => +(v ?? 0).toFixed(3);

export function tf(o: TfParts): string {
  return `translate(${n(o.px)}px, ${n(o.py)}px) rotate(${n(o.pr)}deg) perspective(${PERSP}px) `
    + `rotateY(${n(o.ry)}deg) rotateX(${n(o.rx)}deg) translate3d(${n(o.x)}px, ${n(o.y)}px, ${n(o.z)}px) `
    + `rotateZ(${n(o.rz)}deg) scale(${n(o.sx ?? 1)}, ${n(o.sy ?? 1)})`;
}

/** 槽位修饰：dx/dy/dz 位移，ds 缩放倍率，ry/rx/rz 旋转增量（ryAbs 覆盖绝对角），px/py/pr 屏幕空间 */
export interface SlotMods {
  dx?: number;
  dy?: number;
  dz?: number;
  ds?: number;
  ry?: number;
  ryAbs?: number;
  rx?: number;
  rz?: number;
  px?: number;
  py?: number;
  pr?: number;
}

export interface DeckGeom {
  mobile: boolean;
  /** 相册宽度 / 设计基准宽度 */
  k: number;
  /** 槽位变换（带修饰） */
  T: (slot: number, m?: SlotMods) => string;
  /** 规范滤镜函数表：brightness saturate blur（同序才能逐项插值） */
  F: (b?: number, blur?: number, sat?: number) => string;
  /** 槽位静止亮度 */
  bright: (slot: number) => number;
  /** 槽位静止层级 */
  restZ: (slot: number) => number;
  /** 槽位基准（已按 k 缩放的位移 + 缩放 + 基准 rotateY） */
  base: (slot: number) => { x: number; y: number; z: number; s: number; ry: number };
}

export function F(b = 1, blur = 0, sat = 1): string {
  return `brightness(${+b.toFixed(3)}) saturate(${sat}) blur(${blur}px)`;
}

export function makeGeom(width: number, mobile: boolean): DeckGeom {
  const g = mobile ? GEOM.mobile : GEOM.desk;
  const k = width > 0 ? width / g.base : 1;
  const slotOf = (i: number): Slot => g.slots[Math.min(i, g.slots.length - 1)];
  return {
    mobile,
    k,
    T(slot, m = {}) {
      const b = slotOf(slot);
      const ds = m.ds ?? 1;
      return tf({
        px: m.px,
        py: m.py,
        pr: m.pr,
        ry: m.ryAbs ?? g.angle + (m.ry ?? 0),
        rx: m.rx,
        x: b.x * k + (m.dx ?? 0),
        y: b.y * k + (m.dy ?? 0),
        z: b.z * k + (m.dz ?? 0),
        rz: m.rz,
        sx: b.s * ds,
        sy: b.s * ds,
      });
    },
    F,
    bright: (slot) => slotOf(slot).b,
    restZ: (slot) => 30 - slot * 10,
    base: (slot) => {
      const b = slotOf(slot);
      return { x: b.x * k, y: b.y * k, z: b.z * k, s: b.s, ry: g.angle };
    },
  };
}

/* ===== 几何工具：投影 / 枢轴 / 缓动采样 ===== */

/**
 * 把卡片局部点（以卡片中心为原点，未缩放像素）经规范变换投影到相册坐标（中心为原点）。
 * 与浏览器同一矩阵（DOMMatrix 解析 transform 字符串），用于按真实姿态摆放控件 / 光效。
 */
export function project(transform: string, x: number, y: number, z = 0): { x: number; y: number } {
  const p = new DOMMatrix(transform).transformPoint(new DOMPoint(x, y, z, 1));
  const w = p.w || 1;
  return { x: p.x / w, y: p.y / w };
}

/** 一组槽位变换在相册坐标中的水平包围（未变换相册盒宽 W、高 H，原点为相册中心） */
export function hBounds(transforms: string[], W: number, H: number): { l: number; r: number } {
  let l = Infinity;
  let r = -Infinity;
  for (const t of transforms) {
    for (const [x, y] of [[-W / 2, -H / 2], [W / 2, -H / 2], [W / 2, H / 2], [-W / 2, H / 2]]) {
      const p = project(t, x, y);
      l = Math.min(l, p.x);
      r = Math.max(r, p.x);
    }
  }
  return { l, r };
}

const rad = (d: number): number => (d * Math.PI) / 180;

/**
 * 枢轴补偿：让卡片绕自身局部点 q（已缩放像素，中心为原点）额外旋转 ry / rx 度，
 * 而 q 在屏幕上保持不动。规范变换是 Ry·Rx·(T + q)，所以 d = Rx(-rx)·Ry(-ry)·v - v，v = T + q。
 * 返回可直接喂给 T() 的修饰（dx/dy/dz + 旋转增量）。
 */
export function pivot(
  g: DeckGeom,
  slot: number,
  rot: { ry?: number; rx?: number },
  q: { x?: number; y?: number },
  extra: SlotMods = {},
): SlotMods {
  const b = g.base(slot);
  const ds = extra.ds ?? 1;
  let x = b.x + (extra.dx ?? 0) + (q.x ?? 0) * b.s * ds;
  let y = b.y + (extra.dy ?? 0) + (q.y ?? 0) * b.s * ds;
  let z = b.z + (extra.dz ?? 0);
  const v = { x, y, z };
  // Ry(-a)：x' = x cos a - z sin a；z' = x sin a + z cos a
  const a = rad(rot.ry ?? 0);
  [x, z] = [x * Math.cos(a) - z * Math.sin(a), x * Math.sin(a) + z * Math.cos(a)];
  // Rx(-c)：y' = y cos c + z sin c；z' = -y sin c + z cos c
  const c = rad(rot.rx ?? 0);
  [y, z] = [y * Math.cos(c) + z * Math.sin(c), -y * Math.sin(c) + z * Math.cos(c)];
  return {
    ...extra,
    dx: (extra.dx ?? 0) + x - v.x,
    dy: (extra.dy ?? 0) + y - v.y,
    dz: (extra.dz ?? 0) + z - v.z,
    ry: (extra.ry ?? 0) + (rot.ry ?? 0),
    rx: (extra.rx ?? 0) + (rot.rx ?? 0),
  };
}

/** cubic-bezier 求值（二分反解 x），用于把缓动离散成与关键帧同步的采样 */
export function bezier(x1: number, y1: number, x2: number, y2: number): (t: number) => number {
  const f = (a: number, b: number, u: number): number => 3 * a * u * (1 - u) ** 2 + 3 * b * u * u * (1 - u) + u ** 3;
  return (t) => {
    if (t <= 0) return 0;
    if (t >= 1) return 1;
    let lo = 0;
    let hi = 1;
    for (let i = 0; i < 30; i++) {
      const m = (lo + hi) / 2;
      if (f(x1, x2, m) < t) lo = m;
      else hi = m;
    }
    return f(y1, y2, (lo + hi) / 2);
  };
}

type Pt = [number, number];

/**
 * 推门见光的地面光扇（门洞所在地面平面内的坐标，px）：
 * 门洞底边 = (ox, 0)–(ox + w, 0)，y 轴朝向观者；光源在门洞后方 (ox + w/2, -depth)。
 * 两扇门以外缘为轴向观者转开 θ 度，门扇底边就是光扇的两条侧边；越过门扇自由边后沿光源射线继续铺开。
 * 返回：lit 受光区（6 点，θ 连续变化时点数不变可插值）、sl / sr 左右门扇投影（4 点）。
 */
export function doorFan(w: number, ox: number, theta: number, ext: number, shExt = ext * 0.4, depth = w * 1.1): { lit: Pt[]; sl: Pt[]; sr: Pt[] } {
  const hw = w / 2;
  const c = Math.cos(rad(theta));
  const s = Math.sin(rad(theta));
  const S: Pt = [ox + hw, -depth];
  const BL: Pt = [ox, 0];
  const BR: Pt = [ox + w, 0];
  const FL: Pt = [ox + hw * c, hw * s];
  const FR: Pt = [ox + w - hw * c, hw * s];
  const ray = (P: Pt, len = ext): Pt => {
    const dx = P[0] - S[0];
    const dy = P[1] - S[1];
    const L = Math.hypot(dx, dy) || 1;
    return [P[0] + (dx / L) * len, P[1] + (dy / L) * len];
  };
  return {
    lit: [BL, BR, FR, ray(FR), ray(FL), FL],
    sl: [BL, FL, ray(FL, shExt), ray(BL, shExt)],
    sr: [BR, FR, ray(FR, shExt), ray(BR, shExt)],
  };
}

export const poly = (pts: Pt[]): string => `polygon(${pts.map(([x, y]) => `${x.toFixed(1)}px ${y.toFixed(1)}px`).join(', ')})`;
