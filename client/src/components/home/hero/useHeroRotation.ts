import { computed, ref, type Ref } from 'vue';
import { useLoadingStore } from '../../../stores/loading';
import type { HeroItem } from './types';

/**
 * 轮播状态机：条目 / 照片索引、暂停来源（悬停、lightbox、幕布）。
 * 自动步进不走 JS 定时器：由导航进度条（WAAPI）填满回调 onFillEnd 驱动，与视觉天然同步；
 * 条目切换本身交给 swap（由 HeroCarousel 以 useHeroChoreo 编排）。
 */
export function useHeroRotation(items: Ref<HeroItem[]>, swap: (next: number) => void) {
  const itemIndex = ref(0);
  const photoIndex = ref(0);
  const hovering = ref(false);
  const lightboxOn = ref(false);

  const item = computed(() => items.value[itemIndex.value] ?? items.value[0]);
  const covers = computed(() => (item.value?.covers ?? []).slice(0, 3));

  /** 幕布（首屏 / 路由 loading）仍在屏上：动效与计时都等揭幕 */
  const loadingStore = useLoadingStore();
  const curtainDown = computed(() => loadingStore.bootOverlayVisible || loadingStore.routeOverlayVisible);

  function wrap(i: number): number {
    const n = items.value.length || 1;
    return ((i % n) + n) % n;
  }

  /** 进度条填满：组内还有下一张则轮转，否则切下一篇 */
  function onFillEnd(): void {
    if (lightboxOn.value) return;
    if (photoIndex.value < covers.value.length - 1) photoIndex.value += 1;
    else swap(wrap(itemIndex.value + 1));
  }

  function stepPhoto(delta: number): void {
    const len = covers.value.length;
    if (len < 2) return;
    photoIndex.value = (photoIndex.value + delta + len) % len;
  }

  return {
    itemIndex,
    photoIndex,
    hovering,
    lightboxOn,
    curtainDown,
    item,
    covers,
    wrap,
    onFillEnd,
    stepPhoto,
  };
}
