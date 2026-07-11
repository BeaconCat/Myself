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
    },
    finishRoute() {
      this.routeLoading = false;
    },
  },
});
