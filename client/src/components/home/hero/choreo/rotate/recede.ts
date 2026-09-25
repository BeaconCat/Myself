import type { RotateChoreo } from '../types';
import { RE, glide, pose, topZ } from './shared';

/**
 * 退入暗处：前卡沿视线向深处沉下去、变暗、淡出（像退到舞台暗处），
 * 下一张同时前移补位；旧卡在最暗且透明时换到最底层，从末位后方亮起。
 */
export const recede: RotateChoreo = {
  meta: { id: 'recede', name: '退入暗处', tag: '纵深', desc: '前卡沉向深处变暗消失，下一张前移补位，旧卡在末位重新亮起。', duration: 940 },
  plan(c, g) {
    if (c.role !== 'out') return glide(c, g, 860, c.role === 'in' ? 0 : 110);
    const Z = topZ(g);
    return {
      dur: 940,
      kf: [
        pose(g, 0, {}, { z: Z, easing: RE.in }),
        pose(g, 0, { dz: -220 * g.k, dy: 14 * g.k }, { z: Z, b: 0.3, op: 0, offset: 0.46, easing: RE.linear }),
        pose(g, c.to, { dz: -140 * g.k }, { b: 0.3, op: 0, offset: 0.47, easing: RE.out }),
        pose(g, c.to),
      ],
    };
  },
};
