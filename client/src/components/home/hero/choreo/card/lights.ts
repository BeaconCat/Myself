import type { CardEl, CardEls, ChoreoCtx } from '../types';

/**
 * 卡组光影层的统一进出：中性阴影（每张卡的 .shade）与前卡门缝光外溢（.deck-spill）。
 *
 * 编舞期间卡片会被 clip-path / 开合 / 透视接管，阴影若挂在外观层上会被裁掉，
 * 收尾 cancel 时又一次性出现（BACKLOG：着色阴影硬切）。现在阴影是独立层，
 * 只动画 opacity：离场开头淡出，入场在每张卡落定前 ~300ms 淡入，终点 = 静止态。
 */
const OUT_MS = 200;
const IN_MS = 300;
const SPILL_IN_MS = 420;

/** 离场开头：阴影与外溢光先熄 */
export function lightsOut(e: CardEls, t: ChoreoCtx, delay = 0): void {
  e.cards.forEach((c) => t.a(c.shade, [{ opacity: 1 }, { opacity: 0 }], { dur: OUT_MS, delay, ease: 'linear' }));
  t.a(e.spill, [{ opacity: 1 }, { opacity: 0 }], { dur: OUT_MS, delay, ease: 'linear' });
}

/**
 * 入场结尾：endOf(c) = 该卡入场动画结束的绝对毫秒（相对切换起点）；
 * 阴影在结束前 IN_MS 开始淡入，与卡片同时落定。外溢光跟前卡。
 */
export function lightsIn(e: CardEls, t: ChoreoCtx, endOf: (c: CardEl) => number): void {
  e.cards.forEach((c) => {
    const end = endOf(c);
    t.a(c.shade, [{ opacity: 0 }, { opacity: 1 }], { dur: IN_MS, delay: Math.max(0, end - IN_MS), ease: 'linear' });
  });
  const end = endOf(e.front);
  t.a(e.spill, [{ opacity: 0 }, { opacity: 1 }], { dur: SPILL_IN_MS, delay: Math.max(0, end - SPILL_IN_MS), ease: 'linear' });
}
