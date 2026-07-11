import { defineStore } from 'pinia';

/**
 * 全局加载状态。
 * boot：首屏加载（AppLoading，进度条 + clip 收缩退场）
 * route：路由切换（RouteLoading，纯色下切覆盖 → 下滑裁切退场）
 */
export const useLoadingStore = defineStore('loading', {
  state: () => ({
    bootProgress: 0,
    bootDone: false,
    /** 首屏遮罩仍在屏上（含收缩退场动画期间）：页面动画保持暂停 */
    bootOverlayVisible: true,
    routeLoading: false,
    /** 路由遮罩仍在屏上（含下滑退场动画期间）：动画等它完全消失再播，避免同帧竞争卡顿 */
    routeOverlayVisible: false,
    /** 页面就绪门闩：目标页数据未就绪时揭幕等待，避免揭开后内容才闪出 */
    routePending: 0,
    routeMinUntil: 0,
  }),
  actions: {
    setBootProgress(p: number) {
      this.bootProgress = Math.min(100, Math.max(0, p));
    },
    finishBoot() {
      this.bootProgress = 100;
      this.bootDone = true;
    },
    startRoute() {
      this.routeLoading = true;
      this.routeOverlayVisible = true;
      this.routePending = 0;
    },
    /** 页面调用：数据加载期间按住揭幕，返回释放函数 */
    holdRoute(): () => void {
      if (!this.routeLoading) return () => undefined;
      this.routePending += 1;
      let released = false;
      return () => {
        if (released) return;
        released = true;
        this.routePending = Math.max(0, this.routePending - 1);
        this.tryFinishRoute();
      };
    },
    /** 路由 afterEach 设定最短展示截止后调用 */
    scheduleFinish(minUntil: number) {
      this.routeMinUntil = minUntil;
      // 留一帧给目标页面注册门闩，再开始查验
      window.setTimeout(() => this.tryFinishRoute(), 60);
    },
    tryFinishRoute() {
      if (!this.routeLoading) return;
      if (this.routePending > 0) return;
      const wait = this.routeMinUntil - performance.now();
      if (wait > 0) {
        window.setTimeout(() => this.tryFinishRoute(), wait);
        return;
      }
      this.finishRoute();
    },
    finishRoute() {
      this.routeLoading = false;
    },
  },
});
