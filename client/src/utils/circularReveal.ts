/**
 * 基于 View Transition API 的圆形剪贴蒙版主题切换。
 * expand：新画面以 (x,y) 为圆心向外扩散覆盖旧画面。
 * contract：旧画面向 (x,y) 聚合收缩，露出下方新画面。
 * 不支持的浏览器直接执行 apply 降级。
 */
/** 揭幕圆周上的强调边：与剪贴圆同步扩散 / 收拢的主色细环 */
const RING_W = 2;

/**
 * 强调边宿主：铺满视口、尺寸不变（有独立 view-transition-name，浮在新旧两层快照之上），
 * 圆环画在里面并随动画改变大小；宿主尺寸固定才不会被过渡组拉伸变形。
 */
function mountRing(x: number, y: number): { host: HTMLElement; ring: HTMLElement } {
  const host = document.createElement('div');
  host.className = 'vt-ring-host';
  const ring = document.createElement('div');
  ring.className = 'vt-ring';
  Object.assign(ring.style, { left: `${x}px`, top: `${y}px` });
  host.appendChild(ring);
  document.body.appendChild(host);
  return { host, ring };
}

export function circularReveal(
  origin: { x: number; y: number },
  apply: () => void,
  direction: 'expand' | 'contract' = 'expand',
  /** edge：在扩散的圆周上加一道主色强调边（色盘 / 界面风格切换用；深浅切换不用） */
  opts: { edge?: boolean } = {},
): void {
  const doc = document as Document & {
    startViewTransition?: (cb: () => void) => {
      ready: Promise<void>;
      finished: Promise<void>;
    };
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
  // 供 CSS 预设新画面首帧 clip 初值，堵住动画挂载前的整屏闪现
  root.style.setProperty('--vt-x', `${x}px`);
  root.style.setProperty('--vt-y', `${y}px`);

  const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  let edge: { host: HTMLElement; ring: HTMLElement } | null = null;
  const transition = doc.startViewTransition(() => {
    apply();
    // 在新状态里挂上强调边（此时主色已是切换后的颜色）
    if (opts.edge && !reduced) edge = mountRing(x, y);
  });
  transition.ready.catch(() => edge?.host.remove());
  transition.ready.then(() => {
    if (edge) {
      const d = [`${RING_W}px`, `${radius * 2}px`];
      const size = direction === 'expand' ? d : [...d].reverse();
      edge.ring.animate(
        [
          { width: size[0], height: size[0], opacity: 1 },
          { opacity: 1, offset: 0.82 },
          { width: size[1], height: size[1], opacity: 0 },
        ],
        { duration: 650, easing: 'cubic-bezier(0.65, 0, 0.35, 1)', fill: 'forwards' },
      );
    }
    const grow = [`circle(0px at ${x}px ${y}px)`, `circle(${radius}px at ${x}px ${y}px)`];
    const anim = root.animate(
      { clipPath: direction === 'expand' ? grow : [...grow].reverse() },
      {
        duration: 650,
        easing: 'cubic-bezier(0.65, 0, 0.35, 1)',
        fill: 'forwards',
        pseudoElement:
          direction === 'expand'
            ? '::view-transition-new(root)'
            : '::view-transition-old(root)',
      },
    );
    // forwards 动画会残留在 root 上，污染下一次过渡的同名伪元素（表现为闪屏）。
    // 必须等整个过渡结束（伪元素已销毁）再 cancel，过早清理会在收尾帧回闪。
    transition.finished
      .catch(() => undefined)
      .then(() => {
        anim.cancel();
        root.classList.remove('vt-contract');
        edge?.host.remove();
      });
  });
}
