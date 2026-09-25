import type { TextChoreo } from '../types';

/**
 * 对焦推拉（原 02 文字部分）：旧标题缩远失焦；新标题由虚到实、字距收拢——一次对焦。
 * 入场不做 scale：带缩放动画的合成层会按最大缩放栅格化，终帧与静止态字形会有细微差异。
 */
export const dolly: TextChoreo = {
  meta: {
    id: 'dolly',
    name: '对焦推拉',
    tag: '纵深',
    desc: '标题由虚到实、字距收拢，像镜头完成一次对焦。',
    duration: 1500,
  },
  exit(e, t) {
    e.lines.forEach((l, i) => t.a(l.inner, [
      { opacity: 1, transform: 'scale(1)', filter: 'blur(0px)' },
      { opacity: 0, transform: 'scale(.92)', filter: 'blur(10px)' },
    ], { dur: 360, delay: i * 40, ease: 'quartIn' }));
    [e.tag, e.excerpt, e.btn].forEach((el, i) => t.a(el, [
      { opacity: 1, filter: 'blur(0px)' },
      { opacity: 0, filter: 'blur(6px)' },
    ], { dur: 300, delay: 30 + i * 30, ease: 'quartIn' }));
  },
  enter(e, t) {
    // 终点字距取静止态计算值（标题自带 0.005em），收尾交还 DOM 时不跳
    const rest = getComputedStyle(e.title).letterSpacing;
    e.lines.forEach((l, i) => t.a(l.inner, [
      { letterSpacing: '.3em', filter: 'blur(14px)', opacity: 0 },
      { opacity: 1, offset: 0.35 },
      // 失焦先于字距收拢结束：终段只剩亚像素级的字距 / 缩放收敛，交还 DOM 时栅格一致
      { filter: 'blur(0px)', offset: 0.82 },
      { letterSpacing: rest === 'normal' ? '0px' : rest, filter: 'blur(0px)', opacity: 1 },
    ], { dur: 980, delay: 430 + i * 70 }));
    e.exLines.forEach((l, i) => t.a(l.inner, [
      { opacity: 0, filter: 'blur(8px)', transform: 'translateY(8px)' },
      { opacity: 1, filter: 'blur(0px)', transform: 'translateY(0px)' },
    ], { dur: 700, delay: 640 + i * 60 }));
    t.a(e.tag, [{ opacity: 0, filter: 'blur(6px)' }, { opacity: 1, filter: 'blur(0px)' }], { dur: 500, delay: 420 });
    t.a(e.btn, [
      { opacity: 0, filter: 'blur(6px)', transform: 'translateY(6px)' },
      { opacity: 1, filter: 'blur(0px)', transform: 'translateY(0px)' },
    ], { dur: 640, delay: 780 });
  },
};
