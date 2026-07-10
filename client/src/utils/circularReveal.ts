/**
 * 基于 View Transition API 的圆形剪贴蒙版主题切换。
 * expand：新画面以 (x,y) 为圆心向外扩散覆盖旧画面。
 * contract：旧画面向 (x,y) 聚合收缩，露出下方新画面。
 * 不支持的浏览器直接执行 apply 降级。
 */
export function circularReveal(
  origin: { x: number; y: number },
  apply: () => void,
  direction: 'expand' | 'contract' = 'expand',
): void {
  const doc = document as Document & {
    startViewTransition?: (cb: () => void) => { ready: Promise<void> };
  };

  if (!doc.startViewTransition) {
    apply();
    return;
  }

  const { x, y } = origin;
  const radius = Math.hypot(
    Math.max(x, window.innerWidth - x),
    Math.max(y, window.innerHeight - y),
  );

  const root = document.documentElement;
  root.classList.toggle('vt-contract', direction === 'contract');

  const transition = doc.startViewTransition(apply);
  transition.ready.then(() => {
    const grow = [`circle(0px at ${x}px ${y}px)`, `circle(${radius}px at ${x}px ${y}px)`];
    root.animate(
      { clipPath: direction === 'expand' ? grow : [...grow].reverse() },
      {
        duration: 650,
        easing: 'cubic-bezier(0.65, 0, 0.35, 1)',
        pseudoElement:
          direction === 'expand'
            ? '::view-transition-new(root)'
            : '::view-transition-old(root)',
      },
    );
  });
}
