import type { RotateChoreo } from '../types';
import { between, pose, topZ } from './shared';

const BELT = 'cubic-bezier(.5,0,.3,1)';
const IN = 'cubic-bezier(.5,0,.6,.5)';
const OUT = 'cubic-bezier(.2,.6,.3,1)';

/**
 * 传送平推：整组沿「右上 → 前 → 左下」这条对角线同时平移一格。
 * 前卡顺势滑向左下末位，途中溶解并换到底层；左下的卡继续沿带子滑出淡出，从右上端补回。
 */
export const slide: RotateChoreo = {
  meta: { id: 'slide', name: '传送平推', tag: '秩序', desc: '整组沿对角线平移一格，像传送带送来下一张。', duration: 900 },
  plan(c, g) {
    const dur = 900;
    if (c.role === 'in') {
      return { dur, kf: [pose(g, c.from, {}, { z: g.restZ(0), easing: BELT }), pose(g, 0)] };
    }
    if (c.role === 'out') {
      const Z = topZ(g);
      return {
        dur,
        kf: [
          pose(g, 0, {}, { z: Z, easing: IN }),
          pose(g, c.to, between(g, 0, c.to, 0.5), { z: Z, b: 0.8, op: 0, offset: 0.46, easing: 'linear' }),
          pose(g, c.to, between(g, 0, c.to, 0.52), { b: 0.8, op: 0, offset: 0.47, easing: OUT }),
          pose(g, c.to),
        ],
      };
    }
    // 回绕：带子方向 = 右上槽 → 左下槽；沿带子继续滑出淡出，再从另一端滑入
    const a = g.base(1);
    const b = g.base(g.len - 1);
    const vx = (b.x - a.x) * 0.35;
    const vy = (b.y - a.y) * 0.35;
    const z = g.restZ(c.to);
    return {
      dur,
      kf: [
        pose(g, c.from, {}, { z, easing: IN }),
        pose(g, c.from, { dx: vx, dy: vy }, { z, op: 0, offset: 0.42, easing: 'linear' }),
        pose(g, c.to, { dx: -vx, dy: -vy }, { z, op: 0, offset: 0.5, easing: OUT }),
        pose(g, c.to),
      ],
    };
  },
};
