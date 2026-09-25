import type { CardChoreo } from '../types';
import { lightsIn, lightsOut } from './lights';

/**
 * 纵深推拉（原 02 卡片部分）：镜头后撤——旧卡组沿 Z 轴退入暗处；
 * 再推近——新卡组从镜头前方失焦落位，暗角随推拉收紧再放开。
 */
export const dolly: CardChoreo = {
  meta: {
    id: 'dolly',
    name: '纵深推拉',
    tag: '纵深',
    desc: '旧卡组退入暗处，新卡组从镜头前方失焦落位。',
    duration: 1470,
  },
  exit(e, t) {
    lightsOut(e, t);
    const n = e.cards.length;
    e.cards.forEach((c, k) => t.a(c.el, [
      { transform: e.T(c.slot), filter: e.F(c.b, 0), opacity: 1 },
      { transform: e.T(c.slot, { dz: -560 }), filter: e.F(c.b * 0.5, 6), opacity: 0 },
    ], { dur: 460, delay: (n - 1 - k) * 50, ease: 'quartIn' }));
    t.a(e.bg, [{ opacity: 1, transform: 'scale(1)' }, { opacity: 0, transform: 'scale(.75)' }], { dur: 460, ease: 'quartIn' });
    t.a(t.fx.vig, [{ opacity: 0 }, { opacity: 0.85, offset: 0.42 }, { opacity: 0 }], { dur: 1150, ease: 'quartInOut' });
  },
  enter(e, t) {
    e.cards.forEach((c, k) => t.a(c.el, [
      { transform: e.T(c.slot, { dz: 520 }), filter: e.F(c.b * 1.3, 16), opacity: 0 },
      { opacity: 1, offset: 0.28 },
      { transform: e.T(c.slot), filter: e.F(c.b, 0), opacity: 1 },
    ], { dur: 950, delay: 380 + k * 70 }));
    t.a(e.bg, [{ opacity: 0, transform: 'scale(1.3)' }, { opacity: 1, transform: 'scale(1)' }], { dur: 1050, delay: 380 });
    lightsIn(e, t, (c) => 1330 + c.slot * 70);
  },
};
