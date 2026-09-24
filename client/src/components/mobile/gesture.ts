/**
 * 通用拖拽：主轴锁定 + 末段速度采样（px/ms）。
 * down 返回上下文即接管本次手势，返回 null 放弃；锁定轴与 axis 不符时自动放弃（交给原生滚动）。
 */
export interface DragOptions<C> {
  axis?: 'x' | 'y';
  down: (e: PointerEvent) => C | null;
  begin?: (ctx: C, dx: number, dy: number) => void;
  move: (ctx: C, dx: number, dy: number, e: PointerEvent) => void;
  end: (ctx: C, vx: number, vy: number, dx: number, dy: number) => void;
  /** 捕获阶段监听（边缘手势需先于子元素拿到事件） */
  capture?: boolean;
  /** 起手即阻止冒泡（嵌套拖拽容器） */
  stop?: boolean;
}

let suppressUntil = 0;
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

/** 拖动结束后吞掉紧随的一次 click，避免松手误触 */
export function suppressClick(): void {
  suppressUntil = performance.now() + 80;
}

export function attachDrag<C>(el: HTMLElement, o: DragOptions<C>): () => void {
  const onDown = (e: PointerEvent) => {
    if (e.button !== 0) return;
    const ctx = o.down(e);
    if (!ctx) return;
    if (o.stop) e.stopPropagation();
    const x0 = e.clientX;
    const y0 = e.clientY;
    let lock: 'x' | 'y' | null = null;
    const hist = [{ t: performance.now(), x: x0, y: y0 }];

    const mv = (ev: PointerEvent) => {
      if (ev.pointerId !== e.pointerId) return;
      const dx = ev.clientX - x0;
      const dy = ev.clientY - y0;
      if (!lock) {
        if (Math.hypot(dx, dy) < 6) return;
        lock = Math.abs(dx) > Math.abs(dy) ? 'x' : 'y';
        if (o.axis && o.axis !== lock) {
          off();
          return;
        }
        o.begin?.(ctx, dx, dy);
      }
      hist.push({ t: performance.now(), x: ev.clientX, y: ev.clientY });
      if (hist.length > 6) hist.shift();
      if (ev.cancelable) ev.preventDefault();
      o.move(ctx, dx, dy, ev);
    };
    const up = (ev: PointerEvent) => {
      if (ev.pointerId !== e.pointerId) return;
      off();
      if (!lock) return;
      const a = hist[0];
      const b = hist[hist.length - 1];
      const dt = Math.max(16, b.t - a.t);
      const stale = performance.now() - b.t > 120;
      o.end(
        ctx,
        stale ? 0 : (b.x - a.x) / dt,
        stale ? 0 : (b.y - a.y) / dt,
        ev.clientX - x0,
        ev.clientY - y0,
      );
      suppressClick();
    };
    const off = () => {
      window.removeEventListener('pointermove', mv);
      window.removeEventListener('pointerup', up);
      window.removeEventListener('pointercancel', up);
    };
    window.addEventListener('pointermove', mv, { passive: false });
    window.addEventListener('pointerup', up);
    window.addEventListener('pointercancel', up);
  };
  el.addEventListener('pointerdown', onDown, { capture: !!o.capture });
  return () => el.removeEventListener('pointerdown', onDown, { capture: !!o.capture });
}

/** 橡皮筋阻尼：越拉越紧，d 为软上限 */
export function rubber(x: number, d = 60): number {
  return d * (1 - 1 / ((x / d) * 0.55 + 1));
}

export function clamp(v: number, a: number, b: number): number {
  return Math.max(a, Math.min(b, v));
}

export const reducedMotion = (): boolean =>
  typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
