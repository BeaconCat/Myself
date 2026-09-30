import { defineStore } from 'pinia';

/**
 * 全局加载状态（首屏 AppLoading + 路由 RouteLoading）与「揭幕」交接。
 *
 * 阶段划分：
 * - covered：遮罩完全盖住，页面在背后挂载、取数；页面入场动画按住在首帧（`curtain` getter）。
 * - 揭幕开始（reveal）：遮罩开始退场的同一帧放开页面入场动画 —— 退场与入场重叠，不串行。
 * - overlayVisible：遮罩仍有像素在屏上（含退场动画），仅供需要「完全离屏」语义的场景。
 *
 * 揭幕条件（缺一不可，另有 6s 兜底）：目标页已挂载（mounted 探针）、数据门闩清零、
 * 最短展示时间已到；满足后再等一帧确保首帧已绘制才揭幕。
 */

/** 揭幕兜底：任何条件卡住超过该时长都强制揭幕，避免白屏死等 */
const REVEAL_TIMEOUT_MS = 6000;
/** 首屏：进度条补满到 100% 的过渡时长，之后才允许收缩退场 */
const BOOT_BAR_MS = 250;

/** 揭幕开始事件名：window 上派发，detail 为 'boot' | 'route' */
export const REVEAL_EVENT = 'myself:reveal';

type Probe = () => boolean;

/** 开发期时间线打点（performance.mark），便于 Playwright / DevTools 对照 */
function mark(name: string): void {
  if (import.meta.env.DEV) performance.mark(name);
}

function emitReveal(kind: 'boot' | 'route'): void {
  mark(`${kind}:reveal`);
  window.dispatchEvent(new CustomEvent(REVEAL_EVENT, { detail: kind }));
}

let routeProbe: Probe = () => true;
let routeDeadline = 0;
let routeLoop = 0;
let bootLoop = 0;

export const useLoadingStore = defineStore('loading', {
  state: () => ({
    bootProgress: 0,
    bootDone: false,
    bootDoneAt: 0,
    /** 首屏遮罩完全盖住（收缩退场开始即为 false，页面同帧入场） */
    bootCovered: true,
    /** 首屏遮罩仍在屏上（含收缩退场动画期间） */
    bootOverlayVisible: true,
    /** 路由遮罩盖住阶段；变为 false 即揭幕开始（RouteLoading 同帧开始下滑退场） */
    routeLoading: false,
    /** 本次换页不播遮罩、改用轻量渐入（标准 / 简约档位下的栏目内切换） */
    softNav: false,
    /** 路由遮罩仍在屏上（含下滑退场动画期间） */
    routeOverlayVisible: false,
    /** 页面就绪门闩：目标页数据未就绪时揭幕等待，避免揭开后内容才闪出 */
    routePending: 0,
    routeMinUntil: 0,
    /** 外壳自报「目标页已挂载」（移动端外壳不经 router-view 渲染页面时使用） */
    pageMounted: false,
    /** 揭幕计数：每次揭幕开始 +1，可 watch 作为统一事件 */
    revealSeq: 0,
  }),
  getters: {
    /** 幕布盖住：页面入场动画 / Hero 编舞时钟按住在首帧；揭幕开始即放开 */
    curtain: (s): boolean => s.bootCovered || s.routeLoading,
  },
  actions: {
    setBootProgress(p: number) {
      this.bootProgress = Math.min(100, Math.max(0, p));
    },
    finishBoot() {
      this.bootProgress = 100;
      this.bootDone = true;
      this.bootDoneAt = performance.now();
      mark('boot:done');
    },
    /**
     * 首屏揭幕：bootDone 后等待目标页挂载 + 门闩清零（+ 进度条补满），下一帧揭幕。
     * @param isMounted 目标页挂载探针
     */
    revealBoot(isMounted: Probe) {
      if (!this.bootCovered || bootLoop) return;
      const deadline = performance.now() + REVEAL_TIMEOUT_MS;
      const tick = () => {
        const now = performance.now();
        const ready =
          now >= deadline ||
          (isMounted() && this.routePending === 0 && now >= this.bootDoneAt + BOOT_BAR_MS);
        if (!ready) {
          bootLoop = requestAnimationFrame(tick);
          return;
        }
        // 条件满足的这一帧会先把就绪页面绘制出来，下一帧再揭幕
        bootLoop = requestAnimationFrame(() => {
          bootLoop = 0;
          this.bootCovered = false;
          this.revealSeq += 1;
          emitReveal('boot');
        });
      };
      bootLoop = requestAnimationFrame(tick);
    },
    startRoute() {
      cancelAnimationFrame(routeLoop);
      routeLoop = 0;
      this.routeLoading = true;
      this.routeOverlayVisible = true;
      this.routePending = 0;
      this.pageMounted = false;
      mark('route:start');
    },
    /** 页面调用：数据加载期间按住揭幕（首屏与路由遮罩均生效），返回释放函数 */
    holdRoute(): () => void {
      if (!this.routeLoading && !this.bootCovered) return () => undefined;
      this.routePending += 1;
      let released = false;
      return () => {
        if (released) return;
        released = true;
        this.routePending = Math.max(0, this.routePending - 1);
        if (this.routePending === 0) mark('route:released');
      };
    },
    /** 外壳报告目标页已挂载 */
    markPageMounted() {
      if (!this.pageMounted) mark('page:mounted');
      this.pageMounted = true;
    },
    /**
     * 路由 afterEach 调用：开始逐帧查验揭幕条件。
     * @param minUntil 最短展示截止（performance.now 时间）
     * @param isMounted 目标页挂载探针（门闩在页面 onMounted 同步注册，挂载后查验不会漏）
     */
    scheduleFinish(minUntil: number, isMounted: Probe) {
      this.routeMinUntil = minUntil;
      routeProbe = isMounted;
      routeDeadline = performance.now() + REVEAL_TIMEOUT_MS;
      cancelAnimationFrame(routeLoop);
      let mountedSeen = false;
      const tick = () => {
        if (!this.routeLoading) {
          routeLoop = 0;
          return;
        }
        const now = performance.now();
        const mounted = routeProbe();
        if (mounted && !mountedSeen) {
          mountedSeen = true;
          mark('route:mounted');
        }
        const ready = now >= routeDeadline || (mounted && this.routePending === 0 && now >= this.routeMinUntil);
        if (!ready) {
          routeLoop = requestAnimationFrame(tick);
          return;
        }
        routeLoop = requestAnimationFrame(() => {
          routeLoop = 0;
          this.finishRoute();
        });
      };
      routeLoop = requestAnimationFrame(tick);
    },
    /** 揭幕开始：遮罩开始退场，同帧放开页面入场 */
    finishRoute() {
      if (!this.routeLoading) return;
      cancelAnimationFrame(routeLoop);
      routeLoop = 0;
      this.routeLoading = false;
      this.revealSeq += 1;
      emitReveal('route');
    },
  },
});
