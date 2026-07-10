import { createRouter, createWebHistory } from 'vue-router';
import { useLoadingStore } from '../stores/loading';

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: () => import('../views/HomeView.vue') },
    { path: '/articles', name: 'articles', component: () => import('../views/ArticlesView.vue') },
    { path: '/articles/:slug', name: 'article', component: () => import('../views/ArticleView.vue') },
    { path: '/thoughts', name: 'thoughts', component: () => import('../views/ThoughtsView.vue') },
    { path: '/archive', redirect: '/thoughts' },
    { path: '/about', name: 'about', component: () => import('../views/AboutView.vue') },
    {
      path: '/admin/login',
      name: 'admin-login',
      component: () => import('../views/admin/AdminLoginView.vue'),
      meta: { bare: true },
    },
    {
      path: '/admin',
      component: () => import('../views/admin/AdminLayout.vue'),
      meta: { admin: true, bare: true },
      children: [
        { path: '', redirect: '/admin/posts' },
        { path: 'posts', component: () => import('../views/admin/AdminPostsView.vue') },
        { path: 'posts/:id', component: () => import('../views/admin/AdminPostEditView.vue') },
        { path: 'notes', component: () => import('../views/admin/AdminNotesView.vue') },
        { path: 'media', component: () => import('../views/admin/AdminMediaView.vue') },
        { path: 'apikeys', component: () => import('../views/admin/AdminApiKeysView.vue') },
        { path: 'settings', component: () => import('../views/admin/AdminSettingsView.vue') },
      ],
    },
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
  if (!from.name) return; // 首屏由 AppLoading 负责
  shownAt = performance.now();
  useLoadingStore().startRoute();
  await new Promise((resolve) => window.setTimeout(resolve, COVER_MS));
});

router.afterEach((_to, from) => {
  if (!from.name) return;
  const loading = useLoadingStore();
  const remain = Math.max(0, MIN_SHOW_MS - (performance.now() - shownAt));
  window.setTimeout(() => loading.finishRoute(), remain);
});
