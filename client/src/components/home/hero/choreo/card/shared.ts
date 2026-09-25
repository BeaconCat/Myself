import { fade } from '../engine';
import { tf } from '../geom';
import type { CardChoreo, CardEls, ChoreoCtx } from '../types';
import { parallax } from './parallax';
import { lightsIn, lightsOut } from './lights';

/** FLIP：把相册布局盒映射到第 i 个缩略格（同一变换函数表，rotateY 归零） */
function flipTo(t: ChoreoCtx, e: CardEls, i: number): { s: number; tf: string } | null {
  const thumb = t.nav?.thumb(i);
  if (!thumb) return null;
  const A = t.rel(e.album);
  const th = t.rel(thumb);
  const s = th.w / A.w;
  return { s, tf: tf({ ry: 0, x: th.cx - A.cx, y: th.cy - A.cy, z: 0, sx: s, sy: th.h / A.h }) };
}

/** 缩略格圆角（px，随 --r-base） */
function thumbRadius(t: ChoreoCtx, i: number): number {
  const el = t.nav?.thumb(i);
  return el ? parseFloat(getComputedStyle(el).borderTopLeftRadius) || 0 : 0;
}

/**
 * 共享元素变形（原 06 卡片部分）：旧前卡缩回下方缩略导航里属于它的那一格，
 * 新文章的缩略格同时放大成前卡，两者在空中交错。缩略导航由卡组 chrome 常驻渲染。
 */
export const shared: CardChoreo = {
  meta: {
    id: 'shared',
    name: '共享元素变形',
    tag: '产品感',
    desc: '旧前卡缩回缩略导航，新文章的缩略格放大成前卡。',
    duration: 1470,
  },
  chrome: 'rail',
  exit(e, t) {
    const m = flipTo(t, e, e.idx);
    if (!m) {
      parallax.exit(e, t);
      return;
    }
    lightsOut(e, t);
    const f = e.front;
    // 飞行卡压在缩略导航之上
    t.set(f.el, { zIndex: '48' });
    t.a(f.el, [
      { transform: e.T(0), opacity: 1 },
      { opacity: 1, offset: 0.86 },
      { transform: m.tf, opacity: 0 },
    ], { dur: 600, ease: 'quartInOut' });
    t.a(f.sheet, [{ borderRadius: e.radius }, { borderRadius: `${thumbRadius(t, e.idx) / m.s}px` }], { dur: 600, ease: 'quartInOut' });
    // 终点省略 opacity = 隐式回到底层值（此时已非选中格，静止为 .72），收尾 cancel 无跳变
    t.a(t.nav?.thumb(e.idx), [{ opacity: 0, offset: 0 }, { opacity: 0, offset: 0.84 }], { dur: 620, ease: 'linear' });
    e.backs.forEach((c, i) => t.a(c.el, [
      { transform: e.T(c.slot), opacity: 1, filter: e.F(c.b) },
      { transform: e.T(0, { ds: 0.9 }), opacity: 0, filter: e.F(c.b * 0.5) },
    ], { dur: 320, delay: i * 40, ease: 'quartIn' }));
    t.a(e.bg, fade(1, 0), { dur: 400, ease: 'quartIn' });
  },
  enter(e, t) {
    const m = flipTo(t, e, e.idx);
    if (!m || !t.nav) {
      parallax.enter(e, t);
      return;
    }
    const f = e.front;
    t.set(f.el, { zIndex: '49' });
    t.a(f.el, [{ transform: m.tf, filter: e.F(1) }, { transform: e.T(0), filter: e.F(1) }], { dur: 920, delay: 160, ease: 'spring' });
    t.a(f.sheet, [{ borderRadius: `${thumbRadius(t, e.idx) / m.s}px` }, { borderRadius: e.radius }], { dur: 920, delay: 160, ease: 'spring' });
    t.a(t.nav.thumb(e.idx), [{ opacity: 0 }, { opacity: 0, offset: 0.7 }, { opacity: 1 }], { dur: 1100, ease: 'linear' });
    e.backs.forEach((c, i) => t.a(c.el, [
      { transform: e.T(0, { ds: 0.9 }), opacity: 0, filter: e.F(c.b * 0.5) },
      { transform: e.T(c.slot), opacity: 1, filter: e.F(c.b) },
    ], { dur: 640, delay: 760 + i * 70 }));
    t.a(e.bg, [{ opacity: 0, transform: 'scale(.4)' }, { opacity: 1, transform: 'scale(1)' }], { dur: 1000, delay: 220 });
    lightsIn(e, t, (c) => (c.slot === 0 ? 1080 : 1400 + (c.slot - 1) * 70));
    const { ind, indX, prevIdx } = t.nav;
    if (ind && prevIdx !== e.idx) {
      t.a(ind, [
        { transform: `translateX(${indX(prevIdx)}px)` },
        { transform: `translateX(${indX(e.idx)}px)` },
      ], { dur: 700, delay: 100, ease: 'quartInOut' });
    }
  },
};
