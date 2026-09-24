import { fade } from '../engine';
import type { CardChoreo } from '../types';

/**
 * 门缝开合（原 01 卡片部分，按定稿修改离场）：
 * 离场——最前面的卡片以左边缘为轴向屏幕内开，后面两张以右边缘为轴向内开（对开的两扇门）；
 * 入场——新前卡中线裂开一道竖光并向两侧推开，画面曝光回落，后卡随后被光照出。
 * 卡片透视取 card 层的 perspective 属性（消失点在卡片中心），sheet 只做 rotateY。
 */
export const hinge: CardChoreo = {
  meta: {
    id: 'hinge',
    name: '门缝开合',
    tag: '品牌感',
    desc: '旧卡像门向内开合，新封面从一道竖光里被推开。',
    duration: 1500,
  },
  exit(e, t) {
    const persp = `${Math.round(e.album.offsetWidth * 2.2)}px`;
    const f = e.front;
    t.set(f.el, { perspective: persp });
    t.set(f.sheet, { transformOrigin: '0% 50%' });
    t.a(f.sheet, [
      { transform: 'rotateY(0deg)', filter: 'brightness(1)', opacity: 1 },
      { transform: 'rotateY(78deg)', filter: 'brightness(.3)', opacity: 1, offset: 0.82 },
      { transform: 'rotateY(90deg)', filter: 'brightness(.15)', opacity: 0 },
    ], { dur: 480, ease: 'quartIn' });
    e.backs.forEach((c, i) => {
      t.set(c.el, { perspective: persp });
      t.set(c.sheet, { transformOrigin: '100% 50%' });
      t.a(c.sheet, [
        { transform: 'rotateY(0deg)', filter: 'brightness(1)', opacity: 1 },
        { transform: 'rotateY(-78deg)', filter: 'brightness(.3)', opacity: 1, offset: 0.82 },
        { transform: 'rotateY(-90deg)', filter: 'brightness(.15)', opacity: 0 },
      ], { dur: 440, delay: 50 + i * 50, ease: 'quartIn' });
    });
    t.a(e.bg, fade(1, 0), { dur: 420, ease: 'quartIn' });
  },
  enter(e, t) {
    const f = e.front;
    const W = e.album.offsetWidth;
    // 光缝：两条边从中线向两侧推开（card 内，继承 3D 姿态）；外层横移与 clip-path 同曲线同时长
    [-1, 1].forEach((dir) => {
      const edge = t.make(f.el, 'hc-slit-edge');
      const line = t.make(edge, 'hc-slit-line');
      t.a(edge, [{ transform: 'translateX(0px)' }, { transform: `translateX(${(dir * W) / 2}px)` }], { dur: 700, delay: 540 });
      t.a(line, [
        { opacity: 0, transform: 'scaleY(0)' },
        { opacity: 1, transform: 'scaleY(1)', offset: 0.2 },
        { opacity: 1, transform: 'scaleY(1)', offset: 0.45 },
        { opacity: 0, transform: 'scaleY(1)' },
      ], { dur: 800, delay: 380, ease: 'linear' });
    });
    t.a(f.sheet, [
      { clipPath: 'inset(0% 50% 0% 50% round 18px)' },
      { clipPath: 'inset(0% 0% 0% 0% round 18px)' },
    ], { dur: 700, delay: 540 });
    t.a(f.cv, [
      { transform: 'scale(1.14)', filter: 'brightness(1.8)' },
      { transform: 'scale(1)', filter: 'brightness(1)' },
    ], { dur: 900, delay: 540 });
    e.backs.forEach((c, i) => t.a(c.el, [
      { transform: e.T(0, { ds: 0.9 }), filter: e.F(c.b * 0.3), opacity: 0 },
      { transform: e.T(c.slot), filter: e.F(c.b), opacity: 1 },
    ], { dur: 620, delay: 800 + i * 70 }));
    const r = t.rel(e.album);
    t.set(t.fx.flare, {
      left: `${r.cx - r.w * 0.55}px`,
      top: `${r.cy - r.h * 0.6}px`,
      width: `${r.w * 1.1}px`,
      height: `${r.h * 1.2}px`,
    });
    t.a(t.fx.flare, [
      { opacity: 0, transform: 'scaleX(.15)' },
      { opacity: 0.95, transform: 'scaleX(.6)', offset: 0.3 },
      { opacity: 0, transform: 'scaleX(1.1)' },
    ], { dur: 1000, delay: 420, ease: 'brand' });
    t.a(e.bg, fade(0, 1), { dur: 900, delay: 480 });
  },
};
