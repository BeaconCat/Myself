import type { RotateChoreo } from '../types';
import { RE, glide, pose, topZ } from './shared';

/**
 * 抬起溶解：前卡朝镜头微微抬起并淡出（交叉溶解露出正在前移的下一张），
 * 在完全透明的那一帧换到最底层，再从纵深处于末位淡入归队。
 */
export const lift: RotateChoreo = {
  meta: { id: 'lift', name: '抬起溶解', tag: '推荐', desc: '前卡朝你抬起淡出，下一张从身后接上；旧卡从纵深处回到末位。', duration: 900 },
  plan(c, g) {
    if (c.role !== 'out') return glide(c, g, 820, c.role === 'in' ? 60 : 150);
    const Z = topZ(g);
    return {
      dur: 900,
      kf: [
        pose(g, 0, {}, { z: Z, easing: RE.in }),
        pose(g, 0, { dz: 70 * g.k, dy: -8 * g.k, ds: 1.03 }, { z: Z, b: 1.06, op: 0, offset: 0.4, easing: RE.linear }),
        // 完全透明时换层，并瞬移到末位后方
        pose(g, c.to, { dz: -70 * g.k, ds: 0.96 }, { b: g.bright(c.to) * 0.6, op: 0, offset: 0.41, easing: RE.out }),
        pose(g, c.to),
      ],
    };
  },
};
