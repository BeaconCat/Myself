import { createRouter, createWebHistory } from 'vue-router';
import { useLoadingStore } from '../stores/loading';

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: () => import('../views/HomeView.vue') },
    { path: '/articles', name: 'articles', component: () => import('../views/ArticlesView.vue') },
    { path: '/archive', name: 'archive', component: () => import('../views/ArchiveView.vue') },
    { path: '/about', name: 'about', component: () => import('../views/AboutView.vue') },
  ],
  scrollBehavior: () => ({ top: 0, behavior: 'smooth' }),
});

/** 路由 Loading：进场动画演完（含数据/组件加载）再退场，最短展示 900ms */
const MIN_SHOW_MS = 900;
let shownAt = 0;

router.beforeEach((_to, from) => {
  if (!from.name) return; // 首屏由 AppLoading 负责
  shownAt = performance.now();
  useLoadingStore().startRoute();
});

router.afterEach((_to, from) => {
  if (!from.name) return;
  const loading = useLoadingStore();
  const remain = Math.max(0, MIN_SHOW_MS - (performance.now() - shownAt));
  window.setTimeout(() => loading.finishRoute(), remain);
});
