import { fade, inkBox } from '../engine';
import type { TextChoreo } from '../types';

/**
 * 字幕排印（原 07 文字部分）：旧标题逐字向随机方向散开并失焦；
 * 一根发丝线在首行基线上画出，新字沿同一基线逐个立起，摘要逐行跟随。
 */
export const caption: TextChoreo = {
  meta: {
    id: 'caption',
    name: '字幕排印',
    tag: '排印',
    desc: '旧字逐个散开，新字沿一根发丝基线依次立起。',
    duration: 1400,
  },
  exit(e, t) {
    e.chars.forEach((c) => {
      const a = t.rng() * Math.PI * 2;
      const d = 12 + t.rng() * 26;
      const r = (t.rng() - 0.5) * 40;
      t.a(c, [
        { opacity: 1, transform: 'translate(0px, 0px) rotate(0deg)', filter: 'blur(0px)' },
        {
          opacity: 0,
          transform: `translate(${(Math.cos(a) * d).toFixed(1)}px, ${(Math.sin(a) * d).toFixed(1)}px) rotate(${r.toFixed(1)}deg)`,
          filter: 'blur(6px)',
        },
      ], { dur: 380, delay: Math.round(t.rng() * 100), ease: 'quartIn' });
    });
    e.exLines.forEach((l, i) => t.a(l.inner, [
      { opacity: 1, transform: 'translateY(0px)', filter: 'blur(0px)' },
      { opacity: 0, transform: 'translateY(-6px)', filter: 'blur(4px)' },
    ], { dur: 300, delay: 40 + i * 30, ease: 'quartIn' }));
    t.a(e.tag, fade(1, 0), { dur: 240, ease: 'quartIn' });
    t.a(e.btn, fade(1, 0), { dur: 240, delay: 60, ease: 'quartIn' });
  },
  enter(e, t) {
    const l0 = e.lines[0];
    if (l0) {
      const ib = inkBox(l0.inner);
      const bl = t.make(l0.ln, 'hc-baseline', { left: `${ib.left}px`, width: `${ib.width}px` });
      t.a(bl, [
        { transform: 'scaleX(0)', transformOrigin: '0% 50%', opacity: 1 },
        { transform: 'scaleX(1)', transformOrigin: '0% 50%', opacity: 1, offset: 0.42 },
        { transform: 'scaleX(1)', transformOrigin: '100% 50%', opacity: 1, offset: 0.43 },
        { transform: 'scaleX(0)', transformOrigin: '100% 50%', opacity: 1 },
      ], { dur: 980, delay: 240, ease: 'quartInOut' });
    }
    // 字符 stagger 按总字数收敛：长标题也不会拖过 ~1.4s
    const step = Math.min(26, 420 / Math.max(1, e.chars.length));
    e.chars.forEach((c, i) => t.a(c, [
      { opacity: 0, transform: 'translateY(.42em)' },
      { opacity: 1, transform: 'translateY(0em)' },
    ], { dur: 720, delay: Math.round(320 + i * step) }));
    const tail = Math.round(320 + e.chars.length * step * 0.5);
    e.exLines.forEach((l, i) => t.a(l.inner, [
      { opacity: 0, transform: 'translateY(12px)' },
      { opacity: 1, transform: 'translateY(0px)' },
    ], { dur: 640, delay: tail + i * 70 }));
    t.a(e.tag, [{ opacity: 0, transform: 'translateX(-8px)' }, { opacity: 1, transform: 'translateX(0px)' }], { dur: 520, delay: 280 });
    t.a(e.btn, [{ opacity: 0, transform: 'translateY(8px)' }, { opacity: 1, transform: 'translateY(0px)' }], { dur: 560, delay: tail + 140 });
  },
};
