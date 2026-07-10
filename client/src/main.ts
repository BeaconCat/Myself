import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import { router } from './router';
import { i18n } from './i18n';
import { vReveal } from './directives/reveal';
import { useThemeStore } from './stores/theme';
import { useLoadingStore } from './stores/loading';
import './styles/base.scss';
import './styles/motion.scss';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);
app.use(i18n);
app.directive('reveal', vReveal);

useThemeStore().init();

app.mount('#app');

/* 首屏加载进度：mount → router 就绪 → 资源加载完成 */
const loading = useLoadingStore();
loading.setBootProgress(25);
router.isReady().then(() => loading.setBootProgress(65));
if (document.readyState === 'complete') {
  loading.finishBoot();
} else {
  window.addEventListener('load', () => loading.finishBoot(), { once: true });
}
