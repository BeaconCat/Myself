import type { SlotMods } from '../geom';
import type { RotateCard, RotateGeo, RotatePlan } from '../types';

/**
 * 组内切换公共件：静止帧、晋升卡的默认滑行、以及后退方向的时间倒放。
 */

/** 分段缓动（只用 cubic-bezier，便于倒放） */
export const RE = {
  /** 品牌缓出：晋升卡主力，无回弹 */
  out: 'cubic-bezier(.2,.8,.3,1)',
  /** 平滑缓入：离场前半段 */
  in: 'cubic-bezier(.45,0,.8,.4)',
  /** 进出对称：穿越画面 */
  inOut: 'cubic-bezier(.6,0,.3,1)',
  /** 均匀：交换帧前后的瞬时段 */
  linear: 'linear',
} as const;

/** out 卡交换之前压住全组的层级 */
export const topZ = (g: RotateGeo): number => g.restZ(0) + 5;

/** 槽位静止帧（可带修饰与额外属性）；z 缺省 = 该槽位静止层级 */
export function pose(
  g: RotateGeo,
  slot: number,
  m: SlotMods = {},
  o: { b?: number; op?: number; z?: number; offset?: number; easing?: string } = {},
): Keyframe {
  const k: Keyframe = {
    transform: g.T(slot, m),
    filter: g.F(o.b ?? g.bright(slot)),
    opacity: o.op ?? 1,
    zIndex: o.z ?? g.restZ(slot),
  };
  if (o.offset !== undefined) k.offset = o.offset;
  if (o.easing) k.easing = o.easing;
  return k;
}

/** 两槽位之间的插值修饰（t=0 在 from，t=1 在 to），附加 bulge 位移 */
export function between(g: RotateGeo, from: number, to: number, t: number, bulge: SlotMods = {}): SlotMods {
  const a = g.base(from);
  const b = g.base(to);
  return {
    ...bulge,
    dx: (a.x - b.x) * (1 - t) + (bulge.dx ?? 0),
    dy: (a.y - b.y) * (1 - t) + (bulge.dy ?? 0),
    dz: (a.z - b.z) * (1 - t) + (bulge.dz ?? 0),
    ds: ((a.s / b.s) * (1 - t) + t) * (bulge.ds ?? 1),
  };
}

/** 晋升 / 后排前移：沿直线滑向新槽位，层级全程保持新槽位值（始终低于 out 卡的 topZ） */
export function glide(c: RotateCard, g: RotateGeo, dur: number, delay = 0, easing: string = RE.out): RotatePlan {
  const z = g.restZ(c.to);
  return {
    dur,
    delay,
    kf: [pose(g, c.from, {}, { z, easing }), pose(g, c.to)],
  };
}

/* ===== 倒放（后退方向） ===== */

function reverseEase(e: string | undefined): string {
  const m = /cubic-bezier\(\s*([-\d.]+)\s*,\s*([-\d.]+)\s*,\s*([-\d.]+)\s*,\s*([-\d.]+)\s*\)/.exec(e ?? '');
  if (!m) return 'linear';
  const [x1, y1, x2, y2] = m.slice(1).map(Number);
  return `cubic-bezier(${+(1 - x2).toFixed(3)},${+(1 - y2).toFixed(3)},${+(1 - x1).toFixed(3)},${+(1 - y1).toFixed(3)})`;
}

/** 补全隐式 offset（首 0、末 1、中间均分） */
function withOffsets(kf: Keyframe[]): number[] {
  const out: (number | null)[] = kf.map((k, i) => (typeof k.offset === 'number' ? k.offset : i === 0 ? 0 : i === kf.length - 1 ? 1 : null));
  for (let i = 1; i < out.length; i++) {
    if (out[i] !== null) continue;
    let j = i;
    while (out[j] === null) j++;
    const a = out[i - 1]!;
    const b = out[j]!;
    for (let m = i; m < j; m++) out[m] = a + ((b - a) * (m - i + 1)) / (j - i + 1);
  }
  return out as number[];
}

/**
 * 时间倒放：关键帧逆序、offset 取补、每段缓动取反（cubic-bezier 绕中心对称），
 * delay 在整组时间窗内镜像。用于把「前进一位」的方案变成「后退一位」。
 */
export function reversePlan(p: RotatePlan, window: number): RotatePlan {
  const offs = withOffsets(p.kf);
  const n = p.kf.length;
  const kf: Keyframe[] = [];
  for (let i = n - 1; i >= 0; i--) {
    const { offset: _o, easing: _e, ...props } = p.kf[i];
    const k: Keyframe = { ...props, offset: +(1 - offs[i]).toFixed(4) };
    // 新序列中从第 i 帧出发的一段 = 原序列 i-1 → i 那段倒放
    if (i > 0) k.easing = reverseEase(p.kf[i - 1].easing as string | undefined);
    kf.push(k);
  }
  return { kf, dur: p.dur, delay: window - (p.delay ?? 0) - p.dur };
}
