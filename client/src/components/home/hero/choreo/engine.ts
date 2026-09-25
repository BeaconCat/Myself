/**
 * Hero 编舞引擎：Web Animations API 的薄封装。
 *
 * 一次切换 = 一个 ChoreoRun。文字与卡组的出场 / 入场在同一个同步任务里注册，
 * 共享同一时刻作为时钟原点（delay 均为相对切换起点的绝对毫秒），各自时长可不同；
 * run.done 在全部动画结束后 resolve。所有 enter 关键帧终点 = 静止态，
 * 收尾时 cancel() 即无缝回到 CSS / inline 样式控制，不残留 fill。
 */

/** 欠阻尼弹簧 → CSS linear()；ζ 越小过冲越大，t=1 时已收敛到 0.2% 内 */
function springLinear(zeta: number): string {
  if (typeof CSS === 'undefined' || !CSS.supports('transition-timing-function', 'linear(0, 1)')) {
    return 'cubic-bezier(.2,.8,.3,1.2)';
  }
  const om = Math.log(500) / zeta;
  const wd = om * Math.sqrt(1 - zeta * zeta);
  const pts: number[] = [];
  const N = 60;
  for (let i = 0; i <= N; i++) {
    const x = i / N;
    const v = 1 - Math.exp(-zeta * om * x) * (Math.cos(wd * x) + ((zeta * om) / wd) * Math.sin(wd * x));
    pts.push(i === N ? 1 : +v.toFixed(4));
  }
  return `linear(${pts.join(', ')})`;
}

/**
 * 缓动表：
 * expoOut 入场主力（前 30% 走完 ~80% 路程）；quartIn 出场主力（平滑起步持续加速）；
 * quartInOut 穿越画面的位移；spring / springPop 物理落位（~8% / ~23% 过冲）。
 */
export const EASE = {
  expoOut: 'cubic-bezier(.16,1,.3,1)',
  quartIn: 'cubic-bezier(.5,0,.75,0)',
  quartInOut: 'cubic-bezier(.76,0,.24,1)',
  spring: springLinear(0.62),
  springPop: springLinear(0.42),
  brand: 'cubic-bezier(.2,.8,.3,1)',
  linear: 'linear',
} as const;

export type EaseName = keyof typeof EASE;

export interface AnimOpts {
  dur: number;
  delay?: number;
  ease?: EaseName;
}

export interface Rect {
  x: number;
  y: number;
  w: number;
  h: number;
  cx: number;
  cy: number;
}

type Styles = Record<string, string>;

/** 可被编舞驱动的一次切换：收集动画 / 临时样式 / 临时元素，统一收尾 */
export class ChoreoRun {
  readonly anims: Animation[] = [];
  private restores: (() => void)[] = [];
  private temps: HTMLElement[] = [];
  private speed: number;
  private paused = false;
  private closed = false;
  private settleCbs: (() => void)[] = [];

  constructor(speed = 1) {
    this.speed = speed;
  }

  /** 注册一条动画；delay 为相对切换起点的绝对毫秒 */
  a(el: Element | null | undefined, kf: Keyframe[], o: AnimOpts): Animation | null {
    if (!el || this.closed) return null;
    const anim = el.animate(kf, {
      duration: o.dur,
      delay: o.delay ?? 0,
      easing: EASE[o.ease ?? 'expoOut'],
      fill: 'both',
    });
    anim.playbackRate = this.speed;
    if (this.paused) anim.pause();
    this.anims.push(anim);
    return anim;
  }

  /** 临时改 inline 样式，收尾时还原 */
  set(el: HTMLElement | null | undefined, styles: Styles): void {
    if (!el) return;
    const style = el.style as unknown as Styles;
    const prev: Styles = {};
    for (const k of Object.keys(styles)) {
      prev[k] = style[k];
      style[k] = styles[k];
    }
    this.restores.push(() => Object.assign(el.style, prev));
  }

  /** 临时元素（收尾移除） */
  make(parent: Element, cls: string, styles: Styles = {}, node?: HTMLElement): HTMLElement {
    const e = node ?? document.createElement('div');
    if (cls) e.classList.add(...cls.split(' '));
    Object.assign(e.style, styles);
    parent.append(e);
    this.temps.push(e);
    return e;
  }

  /** 全部动画结束（被 finish() 提前结束同样 resolve） */
  get done(): Promise<void> {
    return Promise.all(this.anims.map((a) => a.finished.catch(() => undefined))).then(() => undefined);
  }

  setSpeed(v: number): void {
    this.speed = v;
    for (const a of this.anims) {
      if (a.playState === 'running') a.updatePlaybackRate(v);
      else a.playbackRate = v;
    }
  }

  pause(): void {
    this.paused = true;
    this.anims.forEach((a) => a.pause());
  }

  play(): void {
    this.paused = false;
    // 已到终点的动画不能直接 play()：会自动倒带重播，改为定格终态
    this.anims.forEach((a) => {
      const end = Number(a.effect?.getComputedTiming().endTime ?? 0);
      if (Number(a.currentTime ?? 0) >= end) a.finish();
      else a.play();
    });
  }

  /** 最早一条动画的起点（ms） */
  get firstDelay(): number {
    const ds = this.anims.map((a) => Number(a.effect?.getTiming().delay ?? 0));
    return ds.length ? Math.min(...ds) : 0;
  }

  /** 整体快进到 ms（共享时钟：所有动画同一时刻） */
  seek(ms: number): void {
    this.anims.forEach((a) => {
      a.currentTime = ms;
    });
  }

  /** 收尾回调（在 cleanup 中触发，仅一次） */
  onSettle(cb: () => void): void {
    if (this.closed) cb();
    else this.settleCbs.push(cb);
  }

  /** 跳到终态并收尾（被新的切换打断时用） */
  finish(): void {
    if (this.closed) return;
    this.anims.forEach((a) => {
      try {
        a.finish();
      } catch {
        /* 无限时长或已取消：忽略 */
      }
    });
    this.cleanup();
  }

  cleanup(): void {
    if (this.closed) return;
    this.closed = true;
    this.anims.forEach((a) => a.cancel());
    this.temps.forEach((e) => e.remove());
    this.restores.reverse().forEach((f) => f());
    this.settleCbs.splice(0).forEach((f) => f());
  }
}

/** mulberry32：可复现的伪随机（散字方向等） */
export function rng(seed: number): () => number {
  let a = seed;
  return () => {
    a |= 0;
    a = (a + 0x6d2b79f5) | 0;
    let x = Math.imul(a ^ (a >>> 15), 1 | a);
    x = (x + Math.imul(x ^ (x >>> 7), 61 | x)) ^ x;
    return ((x ^ (x >>> 14)) >>> 0) / 4294967296;
  };
}

export const fade = (from: number, to: number): Keyframe[] => [{ opacity: from }, { opacity: to }];

/**
 * 行级 mask 擦除（mask 宽 200%，no-repeat：图像外 = 透明 = 隐藏）。
 * show：可见区从左向右长出；hide：隐藏区从左向右吞没。终点帧与静止态等价，cancel 无跳变。
 */
export const MASK = {
  show: { img: 'linear-gradient(90deg, #000 50%, transparent 58%)', from: '116%', to: '0%' },
  hide: { img: 'linear-gradient(90deg, transparent 42%, #000 50%)', from: '100%', to: '-16%' },
};

export function wipe(m: { img: string; from: string; to: string }): Keyframe[] {
  const base = {
    maskImage: m.img,
    maskSize: '200% 100%',
    maskRepeat: 'no-repeat',
    webkitMaskImage: m.img,
    webkitMaskSize: '200% 100%',
    webkitMaskRepeat: 'no-repeat',
  };
  return [
    { ...base, maskPosition: `${m.from} 0`, webkitMaskPosition: `${m.from} 0` },
    { ...base, maskPosition: `${m.to} 0`, webkitMaskPosition: `${m.to} 0` },
  ];
}

/** 行内容盒（去掉遮罩安全区 padding）：扫光轨道 / 基线按真实字宽定位 */
export function inkBox(inner: HTMLElement): { left: number; width: number } {
  const cs = getComputedStyle(inner);
  const pl = parseFloat(cs.paddingLeft) || 0;
  const pr = parseFloat(cs.paddingRight) || 0;
  return { left: inner.offsetLeft + pl, width: inner.offsetWidth - pl - pr };
}
