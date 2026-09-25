import type { TextChoreo } from '../types';

/**
 * 门光点字（原 08 文字部分）：旧文字熄灭；光从新卡组（门）里出来，
 * 按距离由近及远逐字点亮标题，摘要逐行被点亮。与任何卡组动效搭配都以相册中心为光源。
 *
 * 强调系统例外：这里的 textShadow 是「门光」照到字上的瞬时品牌光，只存在于入场前段，
 * 终帧回落为透明零半径（与静态样式等价），静止态文字不带任何投影。
 */
export const glow: TextChoreo = {
  meta: {
    id: 'glow',
    name: '门光点字',
    tag: '叙事',
    desc: '光从卡组里出来，按离门远近逐字点亮标题。',
    duration: 1550,
  },
  exit(e, t) {
    [e.tag, e.title, e.excerpt, e.btn].forEach((el, i) => t.a(el, [
      { opacity: 1, transform: 'translateX(0px)', filter: 'blur(0px)' },
      { opacity: 0, transform: 'translateX(-10px)', filter: 'blur(3px)' },
    ], { dur: 320, delay: i * 30, ease: 'quartIn' }));
  },
  enter(e, t) {
    const door = t.light();
    const dark = t.mode === 'dark';
    const lit = dark ? '#ffffff' : t.color('--primary');
    const text = getComputedStyle(e.title).color;
    const halo = dark ? 'rgba(255,255,255,.9)' : `rgba(${t.color('--primary-rgb')}, .55)`;
    const clear = '0 0 0px rgba(0,0,0,0), 0 0 0px rgba(0,0,0,0)';
    const ds = e.chars.map((c) => {
      const q = t.rel(c);
      return Math.hypot(q.cx - door.cx, q.cy - door.cy);
    });
    const dmin = ds.length ? Math.min(...ds) : 0;
    const dmax = ds.length ? Math.max(...ds) : 0;
    e.chars.forEach((c, i) => {
      const k = (ds[i] - dmin) / Math.max(1, dmax - dmin);
      t.a(c, [
        { opacity: 0, color: lit, textShadow: `0 0 22px ${halo}, 0 0 4px ${halo}`, filter: 'blur(5px)' },
        { opacity: 1, color: lit, textShadow: `0 0 22px ${halo}, 0 0 4px ${halo}`, filter: 'blur(0px)', offset: 0.28 },
        { opacity: 1, color: text, textShadow: clear, filter: 'blur(0px)' },
      ], { dur: 800, delay: Math.round(400 + k * 320) });
    });
    e.exLines.forEach((l, i) => t.a(l.inner, [
      { opacity: 0, textShadow: `0 0 16px ${halo}` },
      { opacity: 1, textShadow: `0 0 16px ${halo}`, offset: 0.3 },
      { opacity: 1, textShadow: '0 0 0px rgba(0,0,0,0)' },
    ], { dur: 720, delay: 680 + i * 70 }));
    t.a(e.tag, [{ opacity: 0, filter: 'brightness(2)' }, { opacity: 1, filter: 'brightness(1)' }], { dur: 600, delay: 460 });
    t.a(e.btn, [
      { opacity: 0, transform: 'translateY(8px)', filter: 'brightness(1.8)' },
      { opacity: 1, transform: 'translateY(0px)', filter: 'brightness(1)' },
    ], { dur: 640, delay: 860 });
  },
};
