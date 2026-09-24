import { onBeforeUnmount, ref, type Ref } from 'vue';

/** 移动端断点：窄于此宽度切换到独立的移动端外壳（不是响应式缩放）。 */
export const MOBILE_QUERY = '(max-width: 767px)';

const mql = typeof window !== 'undefined' ? window.matchMedia(MOBILE_QUERY) : null;
const isMobile = ref(mql?.matches ?? false);
mql?.addEventListener('change', (e) => { isMobile.value = e.matches; });

/** 全局共享的设备形态；外壳层据此选择桌面或移动端组件树。 */
export function useDevice(): { isMobile: Ref<boolean> } {
  onBeforeUnmount(() => { /* 共享单例，无需解绑 */ });
  return { isMobile };
}
