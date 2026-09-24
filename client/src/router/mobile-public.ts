import type { RouteComponent } from 'vue-router';

/**
 * 前台移动端视图：路由 name → 移动端组件（懒加载）。
 * 未登记的路由在移动端回落到桌面组件。由移动端前台负责维护。
 */
export const mobilePublicViews: Record<string, () => Promise<RouteComponent>> = {};
