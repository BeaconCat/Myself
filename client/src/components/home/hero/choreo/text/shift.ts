import type { TextChoreo } from '../types';

/** 交叉位移（原 06 文字部分）：旧文字上移淡出，新文字自下方逐行升入——最安静的产品感。 */
export const shift: TextChoreo = {
  meta: {
    id: 'shift',
    name: '交叉位移',
    tag: '克制',
    desc: '旧文字上移淡出，新文字自下方逐行升入，安静利落。',
    duration: 1200,
  },
  exit(e, t) {
    e.lines.forEach((l, i) => t.a(l.inner, [
      { opacity: 1, transform: 'translateY(0px)' },
      { opacity: 0, transform: 'translateY(-14px)' },
    ], { dur: 300, delay: i * 40, ease: 'quartIn' }));
    [e.tag, e.excerpt, e.btn].forEach((el, i) => t.a(el, [
      { opacity: 1, transform: 'translateY(0px)' },
      { opacity: 0, transform: 'translateY(-10px)' },
    ], { dur: 260, delay: 30 + i * 30, ease: 'quartIn' }));
  },
  enter(e, t) {
    e.lines.forEach((l, i) => t.a(l.inner, [
      { opacity: 0, transform: 'translateY(18px)' },
      { opacity: 1, transform: 'translateY(0px)' },
    ], { dur: 700, delay: 380 + i * 60 }));
    e.exLines.forEach((l, i) => t.a(l.inner, [
      { opacity: 0, transform: 'translateY(12px)' },
      { opacity: 1, transform: 'translateY(0px)' },
    ], { dur: 640, delay: 480 + i * 50 }));
    t.a(e.tag, [{ opacity: 0, transform: 'translateY(8px)' }, { opacity: 1, transform: 'translateY(0px)' }], { dur: 500, delay: 360 });
    t.a(e.btn, [{ opacity: 0, transform: 'translateY(10px)' }, { opacity: 1, transform: 'translateY(0px)' }], { dur: 560, delay: 640 });
  },
};
