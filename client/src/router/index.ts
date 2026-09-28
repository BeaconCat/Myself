import {
  createRouter,
  createWebHistory,
  isNavigationFailure,
  NavigationFailureType,
  type RouteComponent,
  type RouteLocationNormalized,
  type RouteRecordRaw,
} from 'vue-router';
import { MOBILE_QUERY } from '../composables/useDevice';
import { useLoadingStore } from '../stores/loading';
import { configLoaded, useConfigStore } from '../stores/config';

import { adminChildren } from './admin';
import { mobileAdminLogin, mobileAdminViews } from './mobile-admin';
import { mobilePublicViews } from './mobile-public';

type Lazy = () => Promise<RouteComponent>;

/** 桌面组件 + 可选移动端组件 → 命名视图（default / mobile），移动端缺省回落桌面。 */
function views(name: string, desktop: Lazy, mobileMap: Record<string, Lazy>) {
  return { default: desktop, mobile: mobileMap[name] ?? desktop };
}

function page(path: string, name: string, desktop: Lazy): RouteRecordRaw {
  return { path, name, components: views(name, desktop, mobilePublicViews) };
}

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    page('/', 'home', () => import('../views/HomeView.vue')),
    page('/articles', 'articles', () => import('../views/ArticlesView.vue')),
    page('/articles/:slug', 'article', () => import('../views/ArticleView.vue')),
    page('/thoughts', 'thoughts', () => import('../views/ThoughtsView.vue')),
    page('/about', 'about', () => import('../views/AboutView.vue')),
    { path: '/archive', redirect: '/thoughts' },
    {
      path: '/setup',
      name: 'setup',
      component: () => import('../views/SetupView.vue'),
      meta: { bare: true },
    },
    {
      path: '/admin/login',
      name: 'admin-login',
      component: () => import('../views/admin/AdminLoginRoot.vue'),
      meta: { bare: true },
    },
    /* 旧写作入口兼容 */
    { path: '/write/post', redirect: (to) => ({ path: '/admin/write/post', query: to.query }) },
    { path: '/write/note', redirect: '/admin/write/note' },
    {
      path: '/admin',
      component: () => import('../views/admin/AdminRoot.vue'),
      meta: { admin: true, bare: true },
      children: adminChildren.map((c) => ({
        path: c.path,
        name: c.name,
        components: views(c.name, c.component, mobileAdminViews),
      })),
    },
    { path: '/admin/posts/:id', redirect: (to) => ({ path: '/admin/write/post', query: { id: String(to.params.id) } }) },
  ],
  scrollBehavior: () => ({ top: 0, behavior: 'smooth' }),
});

/**
 * 路由 Loading 时序（ms 以点击为 0）：
 *   0      startRoute：遮罩进场（模糊层 0–300，纯色层 120–670），同时并行预取目标页分包
 *   700    遮罩盖满 且 分包就绪 → 放行导航（旧页卸载、新页挂载都发生在遮罩背后）
 *   ≥900   目标页已挂载 + 数据门闩清零 + 最短展示到 → 下一帧揭幕
 *   揭幕   遮罩开始下滑退场（700ms），同帧放开新页入场动画（reveal / rise / Hero）
 * 任一条件卡住 6s 兜底强制揭幕。
 */
const reduced = (): boolean => window.matchMedia('(prefers-reduced-motion: reduce)').matches;
const COVER_MS = 700;
const MIN_SHOW_MS = 900;
/** 预取兜底：网络极慢时不无限等待（路由自身仍会继续加载） */
const PRELOAD_CAP_MS = 6000;
let shownAt = 0;
/** 本次导航是否在播遮罩（beforeEach 判定，afterEach 沿用，避免两处判断不一致） */
let covering = false;
/** 一次性跳过标记：移动端左边缘右滑返回已给出连续反馈，提交后的导航不再盖遮罩 */
let skipCoverUntil = 0;

/** 由手势返回在触发导航前调用：紧随其后的一次导航不播遮罩 */
export function skipNextRouteCover(): void {
  skipCoverUntil = performance.now() + 1500;
}

/** 是否为需要播放遮罩的站内跳转（首屏、后台内部标签、同页 query/hash 变化都不播；桌面与移动端一致） */
function playsCover(to: RouteLocationNormalized, from: RouteLocationNormalized): boolean {
  if (!from.matched.length) return false;
  if (to.meta.admin && from.meta.admin) return false;
  return to.path !== from.path;
}

/** 路由懒加载函数（非已解析组件对象） */
function isLazy(c: unknown): c is Lazy {
  return typeof c === 'function' && !('__vccOpts' in c) && !('props' in c) && !('displayName' in c);
}

/**
 * 并行预取目标页分包：路由记录上的懒加载组件（按当前设备取命名视图），
 * 以及外壳内部 defineAsyncComponent 包装的布局（后台布局 / 登录页），
 * 与遮罩进场同时下载，盖满时通常已就绪。
 */
function preload(to: RouteLocationNormalized): Promise<unknown> {
  const mobile = window.matchMedia(MOBILE_QUERY).matches;
  const jobs: Promise<unknown>[] = [];
  for (const rec of to.matched) {
    const comps = rec.components ?? {};
    const c = (mobile ? comps.mobile : undefined) ?? comps.default;
    if (isLazy(c)) jobs.push(c());
  }
  if (to.meta.admin) {
    jobs.push(mobile ? import('../views/mobile/admin/MobileAdminLayout.vue') : import('../views/admin/AdminLayout.vue'));
  }
  if (to.name === 'admin-login') {
    jobs.push(mobile ? mobileAdminLogin() : import('../views/admin/AdminLoginView.vue'));
  }
  if (!jobs.length) return Promise.resolve();
  const settled = Promise.allSettled(jobs);
  return Promise.race([settled, new Promise((r) => window.setTimeout(r, PRELOAD_CAP_MS))]);
}

/**
 * 目标页挂载探针：叶子路由记录已有组件实例（router-view 挂载后登记），
 * 或外壳自报已挂载（移动端外壳不经 router-view 渲染 tab / 详情）。
 */
export function routeMounted(to: RouteLocationNormalized): boolean {
  if (useLoadingStore().pageMounted) return true;
  const leaf = to.matched[to.matched.length - 1];
  if (!leaf) return true;
  return Object.values(leaf.instances).some(Boolean);
}

router.beforeEach(async (to, from) => {
  // 首次启动：未初始化时一律进入初始化流程；已初始化后 /setup 只保留改密模式
  if (!useConfigStore().loaded) await configLoaded;
  const needsSetup = !!useConfigStore().cfg.needsSetup;
  if (needsSetup && to.name !== 'setup') return { name: 'setup' };
  if (!needsSetup && to.name === 'setup' && to.query.change !== '1') return { path: '/' };
  // 后台鉴权
  if (to.meta.admin && !localStorage.getItem('myself.token')) {
    return { path: '/admin/login' };
  }
  const skip = performance.now() < skipCoverUntil;
  skipCoverUntil = 0;
  covering = !skip && playsCover(to, from);
  const ready = preload(to);
  if (!covering) {
    await ready;
    return;
  }
  shownAt = performance.now();
  useLoadingStore().startRoute();
  const cover = reduced() ? 0 : COVER_MS;
  await Promise.all([ready, new Promise((resolve) => window.setTimeout(resolve, cover))]);
  if (import.meta.env.DEV) performance.mark('route:covered');
});

router.afterEach((to, _from, failure) => {
  const loading = useLoadingStore();
  // 未播遮罩（或上一次播遮罩的导航被本次不播遮罩的导航取代时仍需收尾）
  if (!covering && !loading.routeLoading) return;
  // 被更新的导航取代：由新导航负责揭幕
  if (isNavigationFailure(failure, NavigationFailureType.cancelled)) return;
  if (failure) {
    // 导航被拦下（守卫返回 false 等）：停在原页，直接揭幕
    loading.finishRoute();
    return;
  }
  const min = reduced() ? 200 : MIN_SHOW_MS;
  const remain = Math.max(0, min - (performance.now() - shownAt));
  loading.scheduleFinish(performance.now() + remain, () => routeMounted(to));
});
