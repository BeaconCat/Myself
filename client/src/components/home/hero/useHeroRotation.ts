import { computed, onBeforeUnmount, onMounted, ref, watch, type Ref } from 'vue';
import { useLoadingStore } from '../../../stores/loading';
import type { HeroItem, HeroPhase } from './types';

/** enter 保持时长必须 ≥ 卡片动画 0.75s + 末卡级联 0.2s，过早摘类会尾段跳变闪烁 */
const ENTER_MS = 1000;
/** 出场启动后多久切入下一条（重叠期：出场透明度已归零但位移未播完） */
const ITEM_SWAP_MS = 380;

/**
 * 轮播状态机：条目/照片索引、入退场阶段、悬停与 lightbox 暂停。
 * 自动步进不走 JS 定时器，由进度条 animationend（onFillEnd）驱动，与视觉天然同步。
 */
export function useHeroRotation(items: Ref<HeroItem[]>) {
  const itemIndex = ref(0);
  const photoIndex = ref(0);
  const phase = ref<HeroPhase>('enter');

  const hovering = ref(false);
  const lightboxOn = ref(false);

  const item = computed(() => items.value[itemIndex.value] ?? items.value[0]);
  const covers = computed(() => (item.value?.covers ?? []).slice(0, 3));
  const paused = computed(() => hovering.value || lightboxOn.value);

  let enterTimer = 0;

  /**
   * enter → idle 的计时要等「幕布」揭开才起跑：
   * loading 覆盖期间 CSS 动画被全局暂停，但 JS 定时器照走，
   * 若不等待，揭幕时入场动画类已被摘掉，卡片动效直接跳终态。
   */
  const loadingStore = useLoadingStore();

  function curtainDown(): boolean {
    return loadingStore.bootOverlayVisible || loadingStore.routeOverlayVisible;
  }

  function settle(): void {
    window.clearTimeout(enterTimer);
    if (curtainDown()) {
      const stop = watch(
        () => curtainDown(),
        (down) => {
          if (down) return;
          stop();
          settle();
        },
      );
      return;
    }
    enterTimer = window.setTimeout(() => {
      if (phase.value === 'enter') phase.value = 'idle';
    }, ENTER_MS);
  }

  function swapToItem(next: number): void {
    if (next === itemIndex.value || phase.value === 'out' || lightboxOn.value) return;
    phase.value = 'out';
    // 出/入场重叠：出场透明度前 30% 已归零，中途即切数据开始入场，衔接更流畅
    window.setTimeout(() => {
      itemIndex.value = (next + items.value.length) % items.value.length;
      photoIndex.value = 0;
      phase.value = 'enter';
      settle();
    }, ITEM_SWAP_MS);
  }

  /** 自动步进由进度条驱动：当前段填满（animationend）才前进，与视觉天然同步 */
  function onFillEnd(): void {
    if (phase.value === 'out' || lightboxOn.value) return;
    if (photoIndex.value < covers.value.length - 1) {
      photoIndex.value += 1;
    } else {
      swapToItem(itemIndex.value + 1);
    }
  }

  function stepPhoto(delta: number): void {
    const len = covers.value.length;
    if (len < 2) return;
    photoIndex.value = (photoIndex.value + delta + len) % len;
  }

  onMounted(settle);
  onBeforeUnmount(() => window.clearTimeout(enterTimer));

  return {
    itemIndex,
    photoIndex,
    phase,
    hovering,
    lightboxOn,
    paused,
    item,
    covers,
    swapToItem,
    onFillEnd,
    stepPhoto,
  };
}
