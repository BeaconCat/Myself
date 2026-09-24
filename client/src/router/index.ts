import { createRouter, createWebHistory, type RouteComponent, type RouteRecordRaw } from 'vue-router';
import { useLoadingStore } from '../stores/loading';

import { adminChildren } from './admin';
import { mobileAdminViews } from './mobile-admin';
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
      path: '/admin/login',
      name: 'admin-login',
      component: () => import('../views/admin/AdminLoginView.vue'),
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
 * 路由 Loading：先播覆盖动画，纯色层完整盖住全屏后才放行导航
 * （组件加载/渲染全部发生在覆盖层背后），退场前最短展示 900ms。
 */
const COVER_MS = 700;
const MIN_SHOW_MS = 900;
let shownAt = 0;

router.beforeEach(async (to, from) => {
  // 后台鉴权
  if (to.meta.admin && !localStorage.getItem('myself.token')) {
    return { path: '/admin/login' };
  }
  // 首屏由 AppLoading 负责；后台子路由无 name，用 matched 判断是否已在站内
  if (!from.matched.length) return;
  // 控制中心内部标签页切换不播 loading
  if (to.meta.admin && from.meta.admin) return;
  shownAt = performance.now();
  useLoadingStore().startRoute();
  await new Promise((resolve) => window.setTimeout(resolve, COVER_MS));
});

router.afterEach((to, from) => {
  if (!from.matched.length) return;
  if (to.meta.admin && from.meta.admin) return;
  const loading = useLoadingStore();
  const remain = Math.max(0, MIN_SHOW_MS - (performance.now() - shownAt));
  // 揭幕 = 最短展示时间到 且 目标页数据门闩全部释放
  loading.scheduleFinish(performance.now() + remain);
});
