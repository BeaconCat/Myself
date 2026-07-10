import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import { router } from './router';
import { i18n } from './i18n';
import { vReveal } from './directives/reveal';
import { useThemeStore } from './stores/theme';
import { useLoadingStore } from './stores/loading';
import { useConfigStore } from './stores/config';
import './styles/base.scss';
import './styles/motion.scss';
import './styles/admin.scss';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);
app.use(i18n);
app.directive('reveal', vReveal);

/* 站点配置就绪后再初始化主题（默认色盘/自动切换依赖配置），首帧仍用回退配置渲染 */
const configStore = useConfigStore();
useThemeStore().init();
const configReady = configStore.load().then(() => useThemeStore().init());

app.mount('#app');

/* 首屏进度 = 真实加载事件完成占比：DOM 解析 / 路由(首屏组件)就绪 / 字体就绪 / 全部资源 load */
const loading = useLoadingStore();

const domReady = new Promise<void>((resolve) => {
  if (document.readyState !== 'loading') resolve();
  else document.addEventListener('DOMContentLoaded', () => resolve(), { once: true });
});
const windowLoaded = new Promise<void>((resolve) => {
  if (document.readyState === 'complete') resolve();
  else window.addEventListener('load', () => resolve(), { once: true });
});

const bootTasks: Promise<unknown>[] = [
  domReady,
  router.isReady(),
  document.fonts.ready,
  windowLoaded,
  configReady,
];

let completed = 0;
for (const task of bootTasks) {
  task.then(() => {
    completed += 1;
    loading.setBootProgress((completed / bootTasks.length) * 100);
  });
}
Promise.all(bootTasks).then(() => loading.finishBoot());
