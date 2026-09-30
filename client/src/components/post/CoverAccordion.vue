<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useConfigStore } from '../../stores/config';
import ImageViewer, { type OriginRect } from '../media/ImageViewer.vue';

/**
 * 封面挤压手风琴（≤3 张）：|AAAA|B|C| → |A|BBBB|C| → |A|B|CCCC|
 * - 未悬浮时自动轮播（默认 10s，走 covers.expandMs 配置），悬浮暂停
 * - 位移与缩放同帧：段用绝对定位，left/width 同曲线过渡
 * - 展开段底边的倒计时条（贴底、与段左右对齐，右锚定向右收缩），条走完驱动切换；暂停时条滑出
 * - 同屏多行初始化错峰：每行顺延 1s；用户一旦手动切换即退出错峰
 * - 点未展开段=展开；点展开段=FLIP 飞出 Lightbox
 */
const props = defineProps<{ images: string[] }>();

const config = useConfigStore();
const { t } = useI18n();

const active = ref(0);
const hovering = ref(false);
const viewerOpen = ref(false);
const viewerRect = ref<OriginRect | undefined>();
/** 用户已手动干预：不再应用初始化错峰延迟 */
const userTouched = ref(false);

const list = computed(() => props.images.slice(0, 3));
const expandMs = computed(() => Math.max(1500, Number(config.cfg.covers?.expandMs) || 10000));
const paused = computed(() => viewerOpen.value || hovering.value);
/** 轮播中（条显示条件） */
const rolling = computed(() => list.value.length > 1 && !paused.value);

/* 初始化错峰：同屏连续轮播行依次 +1s，卸载让位 */
const stagger = useStagger();
const initialDelayMs = computed(() => (userTouched.value ? 0 : stagger.delayMs));

function onCountdownEnd(): void {
  if (paused.value || list.value.length < 2) return;
  active.value = (active.value + 1) % list.value.length;
}

function onSegClick(e: MouseEvent, i: number): void {
  if (i !== active.value) {
    userTouched.value = true;
    active.value = i;
    return;
  }
  const img = (e.currentTarget as HTMLElement).querySelector('img');
  if (!img) return;
  const r = img.getBoundingClientRect();
  viewerRect.value = { left: r.left, top: r.top, width: r.width, height: r.height };
  viewerOpen.value = true;
}

/** 绝对布局：active 权重 4、其余 1，left/width 同帧过渡（无 flex reflow 分离感） */
const GAP = 3;

function segStyle(i: number) {
  const len = list.value.length;
  const total = len - 1 + 4;
  const weight = (idx: number) => (idx === active.value ? 4 : 1);
  let prefix = 0;
  for (let k = 0; k < i; k++) prefix += weight(k);
  return {
    left: `calc(${(prefix / total) * 100}% + ${i * GAP}px)`,
    width: `calc(${(weight(i) / total) * 100}% - ${((len - 1) * GAP * weight(i)) / total}px)`,
  };
}
</script>

<script lang="ts">
/** 模块级错峰调度：实例创建领号（0s/1s/2s…），卸载归还 */
let staggerCount = 0;

function useStagger() {
  const slot = staggerCount++;
  onBeforeUnmount(() => { staggerCount = Math.max(0, staggerCount - 1); });
  return { delayMs: slot * 1000 };
}
export default {};
</script>

<template>
  <div
    class="acc"
    @mouseenter="hovering = true"
    @mouseleave="hovering = false"
  >
    <button
      v-for="(src, i) in list"
      :key="src"
      type="button"
      class="seg"
      :class="{ on: i === active }"
      :aria-label="t('a11y.goSlide', { n: i + 1 })"
      :aria-current="i === active ? 'true' : undefined"
      :style="segStyle(i)"
      @click.stop.prevent="onSegClick($event, i)"
    >
      <img :src="src" alt="" loading="lazy" draggable="false" />
    </button>

    <!-- 倒计时条：贴底、随轮播状态滑入滑出；条走完切下一张 -->
    <transition name="cd">
      <span
        v-if="rolling"
        :key="`cd-${active}-${userTouched}`"
        class="countdown"
        :style="[
          segStyle(active),
          {
            animationDuration: expandMs + 'ms',
            animationDelay: initialDelayMs + 'ms',
          },
        ]"
        @animationend="onCountdownEnd"
      />
    </transition>

    <ImageViewer
      v-if="viewerOpen"
      :images="list"
      :start-index="active"
      :origin-rect="viewerRect"
      @close="viewerOpen = false"
    />
  </div>
</template>

<style scoped lang="scss">
.acc {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
  background: #040914;
}

.seg {
  position: absolute;
  top: 0;
  bottom: 0;
  border: none;
  padding: 0;
  background: #040914;
  overflow: hidden;
  cursor: pointer;
  /* 位移(left)与宽度同曲线同帧过渡 */
  transition:
    left 0.6s var(--ease-out),
    width 0.6s var(--ease-out),
    filter var(--dur-fast);

  img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    user-select: none;
  }

  &:not(.on) {
    filter: brightness(0.72);

    &:hover { filter: brightness(0.9); }
  }

  &.on { cursor: zoom-in; }
  &:focus-visible { outline: none; box-shadow: inset 0 0 0 2px var(--ink); }
}

/* 倒计时条：贴在展开段底边，与该段左右边缘完全对齐，右锚定向右收缩；中性白，不发光 */
.countdown {
  position: absolute;
  bottom: 0;
  height: 4px;
  background: rgb(255 255 255 / 0.82);
  box-shadow: 0 -0.5px 0 rgb(0 0 0 / 0.12);
  transform-origin: right center;
  animation: countdown-shrink linear both;
  pointer-events: none;
  /* left/width 跟随段换位 */
  transition: left 0.6s var(--ease-out), width 0.6s var(--ease-out);
}

@keyframes countdown-shrink {
  from { transform: scaleX(1); }
  to { transform: scaleX(0); }
}

/* 条入场自底浮入 / 退场沉底滑出 */
.cd-enter-active { transition: opacity 0.3s ease, translate 0.3s var(--ease-out); }
.cd-leave-active { transition: opacity 0.25s ease, translate 0.25s ease; }
.cd-enter-from, .cd-leave-to { opacity: 0; translate: 0 4px; }
</style>
