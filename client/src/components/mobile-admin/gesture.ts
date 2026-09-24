/** 移动后台手势工具：主轴锁定拖拽 + 速度采样 + 橡皮筋阻尼。 */

export const clamp = (v: number, a: number, b: number): number => Math.max(a, Math.min(b, v));

/** 橡皮筋阻尼：越拉越紧（iOS 近似），d 为阻尼尺度 */
export const rubber = (x: number, d = 60): number => d * (1 - 1 / ((x / d) * 0.55 + 1));

export const prefersReducedMotion = (): boolean =>
  window.matchMedia('(prefers-reduced-motion: reduce)').matches;

export interface DragHandlers<C> {
  /** 按下：返回上下文开始跟踪，返回 null 放弃 */
  down: (e: PointerEvent) => C | null;
  /** 锁定主轴：'x' | 'y'，非本轴时放弃 */
  axis?: 'x' | 'y';
  begin?: (ctx: C) => void;
  move: (ctx: C, dx: number, dy: number, e: PointerEvent) => void;
  /** 抬起：vx/vy 为 px/ms */
  end: (ctx: C, vx: number, vy: number, dx: number, dy: number) => void;
}

let suppressUntil = 0;
/** 拖拽结束后吞掉紧随的 click，避免误触 */
if (typeof document !== 'undefined') {
  document.addEventListener(
    'click',
    (e) => {
      if (performance.now() < suppressUntil) {
        e.stopPropagation();
        e.preventDefault();
      }
    },
    true,
  );
}

/** 绑定拖拽，返回解绑函数 */
export function bindDrag<C>(el: HTMLElement, o: DragHandlers<C>): () => void {
  const onDown = (e: PointerEvent): void => {
    if (e.button !== 0) return;
    const ctx = o.down(e);
    if (!ctx) return;
    const x0 = e.clientX;
    const y0 = e.clientY;
    let lock: 'x' | 'y' | null = null;
    const hist = [{ t: performance.now(), x: x0, y: y0 }];

    const mv = (ev: PointerEvent): void => {
      const dx = ev.clientX - x0;
      const dy = ev.clientY - y0;
      if (!lock) {
        if (Math.hypot(dx, dy) < 6) return;
        lock = Math.abs(dx) > Math.abs(dy) ? 'x' : 'y';
        if (o.axis && o.axis !== lock) {
          off();
          return;
        }
        o.begin?.(ctx);
      }
      hist.push({ t: performance.now(), x: ev.clientX, y: ev.clientY });
      if (hist.length > 6) hist.shift();
      o.move(ctx, dx, dy, ev);
    };
    const up = (ev: PointerEvent): void => {
      off();
      if (!lock) return;
      const a = hist[0];
      const b = hist[hist.length - 1];
      const dt = Math.max(16, b.t - a.t);
      const stale = performance.now() - b.t > 90;
      o.end(ctx, stale ? 0 : (b.x - a.x) / dt, stale ? 0 : (b.y - a.y) / dt, ev.clientX - x0, ev.clientY - y0);
      suppressUntil = performance.now() + 80;
    };
    const off = (): void => {
      window.removeEventListener('pointermove', mv);
      window.removeEventListener('pointerup', up);
      window.removeEventListener('pointercancel', up);
    };
    window.addEventListener('pointermove', mv);
    window.addEventListener('pointerup', up);
    window.addEventListener('pointercancel', up);
  };
  el.addEventListener('pointerdown', onDown);
  return () => el.removeEventListener('pointerdown', onDown);
}

/** 轻触觉反馈（支持的设备） */
export function haptic(ms = 8): void {
  try {
    navigator.vibrate?.(ms);
  } catch {
    /* 不支持时静默 */
  }
}
