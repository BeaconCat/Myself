import { pivot } from '../geom';
import type { RotateChoreo } from '../types';
import { RE, glide, pose, topZ } from './shared';

/**
 * 翻页：前卡以左边为轴像书页一样翻走（右缘远离镜头），翻到侧立时换层；
 * 在末位以同一左边为轴从侧立翻回，像翻过去的一页重新归入书册。
 */
export const flip: RotateChoreo = {
  meta: { id: 'flip', name: '翻页', tag: '书卷气', desc: '前卡以左边为轴翻走，侧立时隐入末位，再以同一轴翻回归队。', duration: 960 },
  plan(c, g) {
    if (c.role !== 'out') return glide(c, g, 800, c.role === 'in' ? 140 : 240);
    const Z = topZ(g);
    const q = { x: -g.W / 2 };
    const turn = 104;
    return {
      dur: 960,
      kf: [
        pose(g, 0, {}, { z: Z, easing: 'cubic-bezier(.45,0,.85,.45)' }),
        pose(g, 0, pivot(g, 0, { ry: turn }, q), { z: Z, b: 0.35, op: 0, offset: 0.42, easing: RE.linear }),
        pose(g, c.to, pivot(g, c.to, { ry: turn }, q), { b: 0.35, op: 0, offset: 0.46, easing: RE.linear }),
        pose(g, c.to, pivot(g, c.to, { ry: turn * 0.8 }, q), { b: 0.45, op: 1, offset: 0.52, easing: RE.out }),
        pose(g, c.to),
      ],
    };
  },
};
