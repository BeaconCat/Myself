import type { RotateChoreo } from '../types';
import { RE, glide, pose, topZ } from './shared';

/**
 * 洗牌抽出：前卡向右上方抽离牌堆（微转、稍缩），下一张随后前移补位；
 * 抽到最远处虚化的那一刻换到最底层，再从牌堆后方落回末位（透明度回升与落位同步）。
 */
export const shuffle: RotateChoreo = {
  meta: { id: 'shuffle', name: '洗牌抽出', tag: '手感', desc: '前卡向侧上方抽出，再从牌堆后方落回末位，像手里洗了一次牌。', duration: 980 },
  plan(c, g) {
    if (c.role !== 'out') return glide(c, g, 760, c.role === 'in' ? 200 : 300);
    const Z = topZ(g);
    const out = { px: 0.56 * g.W, py: -0.16 * g.W, pr: 8, ds: 0.86 };
    return {
      dur: 980,
      kf: [
        pose(g, 0, {}, { z: Z, easing: 'cubic-bezier(.3,0,.2,1)' }),
        pose(g, 0, out, { z: Z, b: 0.78, offset: 0.32, easing: RE.in }),
        // 抽到最远处稍一停顿、虚化，在最淡的那一帧插回牌堆后面（与补位卡仍有少量重叠，靠溶解遮住换层）
        pose(g, 0, { ...out, px: out.px + 8 * g.k, py: out.py - 4 * g.k }, { z: Z, b: 0.7, op: 0.12, offset: 0.42, easing: RE.linear }),
        pose(g, 0, { ...out, px: out.px + 8 * g.k, py: out.py - 4 * g.k }, { z: g.restZ(c.to), b: 0.7, op: 0.12, offset: 0.43, easing: RE.inOut }),
        pose(g, c.to),
      ],
    };
  },
};
