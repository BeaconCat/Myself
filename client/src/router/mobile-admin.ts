import type { RouteComponent } from 'vue-router';

/**
 * 后台移动端视图：路由 name → 移动端组件（懒加载）。
 * 未登记的后台子路由在移动端回落到桌面组件（MobileAdminLayout 为其套原生导航）。由移动端后台负责维护。
 *
 * - 内容页同时承接 admin-posts / admin-notes（分段与路由同步）与 admin-write-note（打开写随想 composer）
 * - 「我的」挂在 admin-settings；完整站点设置用 /admin/settings?full=1 打开桌面设置组件
 */
const content = () => import('../views/mobile/admin/MobileAdminContent.vue');

export const mobileAdminViews: Record<string, () => Promise<RouteComponent>> = {
  'admin-today': () => import('../views/mobile/admin/MobileAdminToday.vue'),
  'admin-posts': content,
  'admin-notes': content,
  'admin-write-note': content,
  'admin-write-post': () => import('../views/mobile/admin/MobileAdminWritePost.vue'),
  'admin-media': () => import('../views/mobile/admin/MobileAdminMedia.vue'),
  'admin-about': () => import('../views/mobile/admin/MobileAdminAbout.vue'),
  'admin-settings': () => import('../views/mobile/admin/MobileAdminMe.vue'),
};

/** 移动端登录页（/admin/login 不在 /admin 子路由内，由集成方在 router/index.ts 接入） */
export const mobileAdminLogin = (): Promise<RouteComponent> => import('../views/mobile/admin/MobileLogin.vue');
