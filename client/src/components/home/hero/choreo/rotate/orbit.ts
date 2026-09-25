import type { RotateChoreo } from '../types';
import { between, pose, topZ } from './shared';

const ARC = 'cubic-bezier(.45,0,.55,1)';

/**
 * 环绕轨道：三张卡沿同一椭圆轨道同向转动一位。前卡从左侧外圈绕向后方，晋升卡从右侧外圈绕到前方，
 * 两者在轨道两端错开；前卡在外圈最远、半透明的那一刻换到后方。
 */
export const orbit: RotateChoreo = {
  meta: { id: 'orbit', name: '环绕轨道', tag: '空间感', desc: '三张卡沿椭圆轨道同时转动一位，前后自然交替。', duration: 980 },
  plan(c, g) {
    const dur = 980;
    const W = g.W;
    if (c.role === 'out') {
      const Z = topZ(g);
      const mid = between(g, 0, c.to, 0.5, { dx: -0.3 * W, dz: -40 * g.k, ry: 10 });
      return {
        dur,
        kf: [
          pose(g, 0, {}, { z: Z, easing: ARC }),
          pose(g, c.to, mid, { z: Z, b: 0.7, op: 0.25, offset: 0.5, easing: 'linear' }),
          pose(g, c.to, mid, { b: 0.7, op: 0.25, offset: 0.51, easing: ARC }),
          pose(g, c.to),
        ],
      };
    }
    const bulge = c.role === 'in' ? { dx: 0.26 * W, dz: 30 * g.k, ry: -8 } : { dx: 0.06 * W, dy: 0.12 * W, dz: -60 * g.k };
    const z = g.restZ(c.to);
    const b = (g.bright(c.from) + g.bright(c.to)) / 2;
    return {
      dur,
      kf: [
        pose(g, c.from, {}, { z, easing: ARC }),
        pose(g, c.to, between(g, c.from, c.to, 0.5, bulge), { z, b, offset: 0.5, easing: ARC }),
        pose(g, c.to),
      ],
    };
  },
};
