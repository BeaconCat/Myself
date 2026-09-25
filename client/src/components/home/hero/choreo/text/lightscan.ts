import { MASK, fade, inkBox, wipe } from '../engine';
import type { TextChoreo } from '../types';

/**
 * 门缝光扫（原 01 文字部分）：旧文字逐行被 mask 擦除；
 * 新标题逐行被一束光扫过即显现（行内 glint 与 mask 边缘同速），摘要逐行跟随。
 */
export const lightscan: TextChoreo = {
  meta: {
    id: 'lightscan',
    name: '门缝光扫',
    tag: '品牌感',
    desc: '一束光从门缝扫过标题，文字按行被照亮。',
    duration: 1500,
  },
  exit(e, t) {
    e.lines.forEach((l, i) => t.a(l.inner, wipe(MASK.hide), { dur: 380, delay: i * 60, ease: 'quartIn' }));
    e.exLines.forEach((l, i) => t.a(l.inner, wipe(MASK.hide), { dur: 320, delay: 60 + i * 40, ease: 'quartIn' }));
    t.a(e.tag, fade(1, 0), { dur: 240, ease: 'quartIn' });
    t.a(e.btn, fade(1, 0), { dur: 240, delay: 80, ease: 'quartIn' });
  },
  enter(e, t) {
    e.lines.forEach((l, i) => {
      const d = 600 + i * 110;
      const ib = inkBox(l.inner);
      const lw = ib.width;
      t.a(l.inner, wipe(MASK.show), { dur: 720, delay: d, ease: 'quartInOut' });
      const track = t.make(l.ln, 'hc-glint-track', { left: `${ib.left}px`, width: `${lw}px` });
      const g = t.make(track, 'hc-glint');
      // 起点帧透明：等待期间短行的扫光会探进行内，不能提前露出
      const x0 = -0.08 * lw - 32;
      const x1 = 1.08 * lw - 32;
      t.a(g, [
        { transform: `translateX(${x0}px)`, opacity: 0 },
        { transform: `translateX(${x0 + (x1 - x0) * 0.04}px)`, opacity: 1, offset: 0.06 },
        { transform: `translateX(${x1}px)`, opacity: 1, offset: 0.85 },
        { transform: `translateX(${x1}px)`, opacity: 0 },
      ], { dur: 720, delay: d, ease: 'quartInOut' });
    });
    e.exLines.forEach((l, i) => t.a(l.inner, wipe(MASK.show), { dur: 560, delay: 800 + i * 60, ease: 'quartInOut' }));
    t.a(e.tag, [{ opacity: 0, transform: 'translateY(6px)' }, { opacity: 1, transform: 'translateY(0px)' }], { dur: 500, delay: 540 });
    t.a(e.btn, [{ opacity: 0, transform: 'translateY(10px)' }, { opacity: 1, transform: 'translateY(0px)' }], { dur: 560, delay: 900 });
  },
};
