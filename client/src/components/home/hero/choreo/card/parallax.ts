import { fade } from '../engine';
import type { CardChoreo } from '../types';
import { lightsIn, lightsOut } from './lights';

/** 轻视差换位（原 07 卡片部分）：卡片只做极轻的位移换位，画面层反向视差——最克制。 */
export const parallax: CardChoreo = {
  meta: {
    id: 'parallax',
    name: '轻视差换位',
    tag: '克制',
    desc: '卡片轻移换位，画面层反向视差，不抢文字的戏。',
    duration: 1380,
  },
  exit(e, t) {
    lightsOut(e, t);
    e.cards.forEach((c) => {
      t.a(c.el, [
        { transform: e.T(c.slot), opacity: 1 },
        { transform: e.T(c.slot, { dx: -24 }), opacity: 0 },
      ], { dur: 440, ease: 'quartIn' });
      t.a(c.cv, [{ transform: 'translateX(0px)' }, { transform: 'translateX(16px)' }], { dur: 440, ease: 'quartIn' });
    });
    t.a(e.bg, fade(1, 0), { dur: 500, ease: 'quartIn' });
  },
  enter(e, t) {
    e.cards.forEach((c, k) => {
      t.a(c.el, [
        { transform: e.T(c.slot, { dx: 28 }), opacity: 0 },
        { transform: e.T(c.slot), opacity: 1 },
      ], { dur: 900, delay: 260 + k * 60 });
      t.a(c.cv, [
        { transform: 'translateX(-18px) scale(1.05)' },
        { transform: 'translateX(0px) scale(1)' },
      ], { dur: 1000, delay: 260 + k * 60 });
    });
    t.a(e.bg, fade(0, 1), { dur: 800, delay: 200 });
    lightsIn(e, t, (c) => 1160 + c.slot * 60);
  },
};
