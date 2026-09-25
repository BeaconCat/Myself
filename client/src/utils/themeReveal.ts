/**
 * 可叠加的圆形主题揭幕（不依赖 View Transition）。
 *
 * 原理：切换前把当前页面克隆成一层「冻结旧主题」的快照盖在最上面，再立刻把新主题应用到真实页面；
 * 快照上从点击处挖一个不断扩大的圆洞，洞里露出下面已经是新主题的真实页面。
 *
 * 连续切换：每次点击都生成一层新快照（冻结的是此刻真实页面的主题），插在已有快照的**下面**，
 * 同时给所有快照追加这次的新洞。于是：
 *   - 两个圆都没覆盖的地方 → 最早的快照（最初的主题）；
 *   - 只在旧圆内 → 中间主题；
 *   - 在新圆内 → 真实页面（最新主题）。
 * 视觉上就是「新蒙版从新的点击处扩散，盖住旧蒙版」。任一层的洞扩到全屏即移除该层。
 *
 * 快照层 pointer-events: none，扩散期间页面仍可点击（可继续切换）。
 * 减弱动效时直接切换。
 */

interface Hole {
  x: number;
  y: number;
  max: number;
  start: number;
}

interface Layer {
  el: HTMLElement;
  holes: Hole[];
}

const DURATION = 720;
/** 近似 cubic-bezier(.2,.8,.3,1)：起步快、收尾柔 */
const ease = (t: number): number => 1 - (1 - t) ** 3.4;

const layers: Layer[] = [];
let host: HTMLElement | null = null;
let raf = 0;

function ensureHost(): HTMLElement {
  if (host?.isConnected) return host;
  host = document.createElement('div');
  host.className = 'theme-reveal-host';
  Object.assign(host.style, {
    position: 'fixed',
    inset: '0',
    zIndex: '2147483000',
    pointerEvents: 'none',
  });
  document.body.appendChild(host);
  return host;
}

/** 读取元素上所有自定义属性的计算值（var() 已被替换，可直接冻结） */
function customProps(el: Element): [string, string][] {
  const cs = getComputedStyle(el);
  const out: [string, string][] = [];
  for (let i = 0; i < cs.length; i++) {
    const name = cs[i];
    if (name.startsWith('--')) out.push([name, cs.getPropertyValue(name)]);
  }
  return out;
}

/** 克隆当前页面并冻结当前主题 */
function snapshot(): HTMLElement {
  const app = document.getElementById('app') ?? document.body;
  const root = document.documentElement;

  const layer = document.createElement('div');
  layer.className = 'theme-reveal-layer';
  const rootCs = getComputedStyle(root);
  const bodyBg = getComputedStyle(document.body).backgroundColor;
  Object.assign(layer.style, {
    position: 'absolute',
    inset: '0',
    overflow: 'hidden',
    background: bodyBg !== 'rgba(0, 0, 0, 0)' ? bodyBg : rootCs.backgroundColor,
    colorScheme: rootCs.colorScheme,
    color: getComputedStyle(document.body).color,
  });
  for (const [k, v] of customProps(root)) layer.style.setProperty(k, v);

  const clone = app.cloneNode(true) as HTMLElement;
  clone.removeAttribute('id');
  Object.assign(clone.style, {
    position: 'absolute',
    top: `${-window.scrollY}px`,
    left: `${-window.scrollX}px`,
    width: `${root.clientWidth}px`,
    margin: '0',
  });
  layer.appendChild(clone);

  // 逐节点对齐原树与克隆树：冻结作用域变量（.studio 等在 [data-mode] 下重定义变量的容器）、
  // 表单当前值、滚动位置（滚动需挂载后设置）
  const src = [app, ...app.querySelectorAll('*')];
  const dst = [clone, ...clone.querySelectorAll('*')];
  const scrolls: [Element, number, number][] = [];
  for (let i = 0; i < src.length && i < dst.length; i++) {
    const s = src[i];
    const d = dst[i] as HTMLElement;
    if (s.classList.contains('studio') || s.classList.contains('app-shell')) {
      for (const [k, v] of customProps(s)) d.style.setProperty(k, v);
    }
    if (s instanceof HTMLInputElement || s instanceof HTMLTextAreaElement || s instanceof HTMLSelectElement) {
      (d as HTMLInputElement).value = s.value;
    }
    if (s.scrollTop || s.scrollLeft) scrolls.push([d, s.scrollTop, s.scrollLeft]);
  }

  ensureHost().prepend(layer);
  for (const [d, top, left] of scrolls) {
    d.scrollTop = top;
    d.scrollLeft = left;
  }
  return layer;
}

function maskFor(holes: Hole[], now: number): string {
  return holes
    .map((h) => {
      const r = ease(Math.min(1, (now - h.start) / DURATION)) * h.max;
      return `radial-gradient(circle at ${h.x}px ${h.y}px, transparent ${r}px, #000 ${r + 1}px)`;
    })
    .join(', ');
}

/** 按当前时间刷新所有快照层的蒙版，移除已被完全揭开的层；返回是否仍有层 */
function paint(): boolean {
  const now = performance.now();
  for (let i = layers.length - 1; i >= 0; i--) {
    const layer = layers[i];
    if (layer.holes.some((h) => now - h.start >= DURATION)) {
      layer.el.remove();
      layers.splice(i, 1);
      continue;
    }
    const mask = maskFor(layer.holes, now);
    const st = layer.el.style as CSSStyleDeclaration & { webkitMaskImage: string; webkitMaskComposite: string };
    st.maskImage = mask;
    st.webkitMaskImage = mask;
  }
  return layers.length > 0;
}

function tick(): void {
  if (paint()) {
    raf = requestAnimationFrame(tick);
    return;
  }
  raf = 0;
  host?.remove();
  host = null;
  document.documentElement.classList.remove('theme-revealing');
}

/**
 * 从 origin 处圆形扩散切换主题。apply 负责真正修改主题（同步执行）。
 */
export function themeReveal(origin: { x: number; y: number }, apply: () => void): void {
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
    apply();
    return;
  }
  const { x, y } = origin;
  const max = Math.hypot(Math.max(x, window.innerWidth - x), Math.max(y, window.innerHeight - y)) + 2;
  const hole: Hole = { x, y, max, start: performance.now() };

  const el = snapshot();
  const st = el.style as CSSStyleDeclaration & { webkitMaskComposite: string };
  st.maskComposite = 'intersect';
  st.webkitMaskComposite = 'source-in';
  layers.unshift({ el, holes: [] });
  for (const layer of layers) layer.holes.push(hole);

  // 揭幕期间真实页面关闭颜色过渡：洞里直接是新主题的终态，避免圆内再「渐变一次」
  document.documentElement.classList.add('theme-revealing');
  apply();
  paint();
  if (!raf) raf = requestAnimationFrame(tick);
}

/** 取事件的揭幕原点：鼠标点击用指针位置，键盘触发用按钮中心 */
export function revealOrigin(e?: Event): { x: number; y: number } {
  const me = e as MouseEvent | undefined;
  if (me && (me.clientX || me.clientY)) return { x: me.clientX, y: me.clientY };
  const el = (e?.currentTarget ?? e?.target) as Element | null;
  if (el?.getBoundingClientRect) {
    const r = el.getBoundingClientRect();
    return { x: r.left + r.width / 2, y: r.top + r.height / 2 };
  }
  return { x: window.innerWidth / 2, y: window.innerHeight / 2 };
}
