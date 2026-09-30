import { onBeforeUnmount, watchEffect, type Ref } from 'vue';
import { useConfigStore } from '../stores/config';

/** 页面标题：「页面名 · 站点名」；页面名为空时只用站点名，离开页面恢复站点名 */
export function usePageTitle(title: Ref<string>): void {
  const config = useConfigStore();
  watchEffect(() => {
    const site = config.cfg.site.title;
    document.title = title.value ? `${title.value} · ${site}` : site;
  });
  onBeforeUnmount(() => {
    document.title = config.cfg.site.title;
  });
}
