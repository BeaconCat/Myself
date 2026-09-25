import { pivot } from '../geom';
import type { RotateChoreo } from '../types';
import { RE, glide, pose, topZ } from './shared';

/**
 * 扇形翻叠：前卡以底边为轴向后倒下（顶边远离镜头），带一点扇形偏转；
 * 倒到几乎侧立、最暗的那一刻换到最底层，贴地滑到末位后以底边为轴立起。
 */
export const fan: RotateChoreo = {
  meta: { id: 'fan', name: '扇形翻叠', tag: '物理感', desc: '前卡绕底边向后倒下，侧立的一瞬插回底部，再在末位立起。', duration: 960 },
  plan(c, g) {
    if (c.role !== 'out') return glide(c, g, 820, c.role === 'in' ? 120 : 220);
    const Z = topZ(g);
    const q = { y: g.H / 2 };
    const down = 84;
    return {
      dur: 960,
      kf: [
        pose(g, 0, {}, { z: Z, easing: 'cubic-bezier(.5,0,.9,.5)' }),
        pose(g, 0, pivot(g, 0, { rx: down }, q, { rz: -6 }), { z: Z, b: 0.3, offset: 0.4, easing: RE.linear }),
        // 侧立（只剩一条暗边）时换层，贴地滑到末位
        pose(g, c.to, pivot(g, c.to, { rx: down }, q, { rz: 4 }), { b: 0.3, offset: 0.52, easing: 'cubic-bezier(.1,.5,.3,1)' }),
        pose(g, c.to),
      ],
    };
  },
};
