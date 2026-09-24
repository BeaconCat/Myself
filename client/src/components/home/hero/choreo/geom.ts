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
  };
}
