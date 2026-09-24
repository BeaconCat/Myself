import type { RouteComponent } from 'vue-router';

/**
 * 后台移动端视图：路由 name → 移动端组件（懒加载）。
 * 未登记的后台子路由在移动端回落到桌面组件。由移动端后台负责维护。
 */
export const mobileAdminViews: Record<string, () => Promise<RouteComponent>> = {};
