import { onBeforeUnmount, shallowRef, watch, type Ref } from 'vue';
import { ChoreoRun, rng, type Rect } from './choreo/engine';
import { REDUCED_CARD, REDUCED_TEXT } from './choreo';
import type {
  CardChoreo,
  CardEls,
  ChoreoCtx,
  FxEls,
  NavEls,
  TextChoreo,
  TextEls,
} from './choreo/types';

export function prefersReducedMotion(): boolean {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/** 一次切换的全部参与者；old* 缺省 = 首次入场（只有 enter） */
export interface SwapParts {
  text: TextChoreo;
  card: CardChoreo;
  oldText?: TextEls | null;
  newText: TextEls;
  oldDeck?: CardEls | null;
  newDeck: CardEls;
  nav: NavEls | null;
}

interface Options {
  /** Hero 舞台根节点（rel 坐标系） */
  host: Ref<HTMLElement | null>;
  fx: Ref<FxEls | null>;
  speed: Ref<number>;
  /** false 时冻结进行中的切换（混搭器暂停 / 幕布未揭开） */
  playing: Ref<boolean>;
}

/**
 * 编舞调度：文字与卡组各取一套编舞，在同一同步任务里注册 exit / enter
 * （共享时钟原点，各自时长可不同），run.done = 两者全部结束。
 * 新的切换到来时先把进行中的一次直接跳到终态并收尾。
 */
export function useHeroChoreo(opts: Options) {
  const running = shallowRef<ChoreoRun | null>(null);
  let seed = 7;

  function makeCtx(run: ChoreoRun, p: SwapParts): ChoreoCtx {
    const host = opts.host.value!;
    const hostRect = host.getBoundingClientRect();
    // 外层缩放（混搭器舞台）下，bounding rect 与布局像素不一致：统一换算回布局像素
    const k = host.offsetWidth ? hostRect.width / host.offsetWidth : 1;
    const rel = (el: Element): Rect => {
      const r = el.getBoundingClientRect();
      const x = (r.left - hostRect.left) / k;
      const y = (r.top - hostRect.top) / k;
      const w = r.width / k;
      const h = r.height / k;
      return { x, y, w, h, cx: x + w / 2, cy: y + h / 2 };
    };
    const root = document.documentElement;
    return {
      a: run.a.bind(run),
      set: run.set.bind(run),
      make: run.make.bind(run),
      rel,
      rng: rng((seed += 13)),
      mode: root.dataset.mode === 'light' ? 'light' : 'dark',
      mobile: p.newDeck.mobile,
      W: host.offsetWidth,
      H: host.offsetHeight,
      fx: opts.fx.value!,
      nav: p.nav,
      light: () => rel(p.newDeck.album),
      color: (name) => getComputedStyle(root).getPropertyValue(name).trim(),
    };
  }

  /** 立即结束进行中的切换（跳终态 + 收尾） */
  function finish(): void {
    const run = running.value;
    running.value = null;
    run?.finish();
  }

  function play(p: SwapParts): ChoreoRun {
    finish();
    const run = new ChoreoRun(opts.speed.value);
    const reduced = prefersReducedMotion();
    const text = reduced ? REDUCED_TEXT : p.text;
    const card = reduced ? REDUCED_CARD : p.card;
    const t = makeCtx(run, p);
    if (p.oldText) text.exit(p.oldText, t);
    if (p.oldDeck) card.exit(p.oldDeck, t);
    text.enter(p.newText, t);
    card.enter(p.newDeck, t);
    if (!opts.playing.value) run.pause();
    running.value = run;
    void run.done.then(() => {
      if (running.value !== run) return;
      running.value = null;
      run.cleanup();
    });
    return run;
  }

  watch(opts.speed, (v) => running.value?.setSpeed(v));
  watch(opts.playing, (on) => {
    if (on) running.value?.play();
    else running.value?.pause();
  });

  onBeforeUnmount(finish);

  return { play, finish, running };
}
