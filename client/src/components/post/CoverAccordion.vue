<script setup lang="ts">
import { computed, ref } from 'vue';
import { useConfigStore } from '../../stores/config';
import ImageViewer, { type OriginRect } from '../media/ImageViewer.vue';

/**
 * 封面挤压手风琴（≤3 张）：|AAAA|B|C| → |A|BBBB|C| → |A|B|CCCC|
 * 自动轮换间隔走站点配置；点未展开段=快速展开，点展开段=FLIP 飞出 Lightbox。
 */
const props = withDefaults(
  defineProps<{
    images: string[];
    /** always：常驻轮换（文章页）；hover：悬停才轮换（列表缩略图） */
    autoplay?: 'always' | 'hover';
  }>(),
  { autoplay: 'always' },
);

const config = useConfigStore();

const active = ref(0);
const hovering = ref(false);
const viewerOpen = ref(false);
const viewerRect = ref<OriginRect | undefined>();

const list = computed(() => props.images.slice(0, 3));
const expandMs = computed(() => Math.max(1500, Number(config.cfg.covers?.expandMs) || 5000));
/** 倒计时/轮换暂停条件：hover 模式未悬停、或 Lightbox 打开 */
const paused = computed(
  () => viewerOpen.value || (props.autoplay === 'hover' && !hovering.value),
);

/** 倒计时条走完即切换：视觉与逻辑同一时钟 */
function onCountdownEnd(): void {
  if (paused.value || list.value.length < 2) return;
  active.value = (active.value + 1) % list.value.length;
}

function onSegClick(e: MouseEvent, i: number): void {
  if (i !== active.value) {
    active.value = i;
    return;
  }
  // 已展开：FLIP 飞出 Lightbox
  const img = (e.currentTarget as HTMLElement).querySelector('img');
  if (!img) return;
  const r = img.getBoundingClientRect();
  viewerRect.value = { left: r.left, top: r.top, width: r.width, height: r.height };
  viewerOpen.value = true;
}

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
      @click.stop.prevent="onSegClick($event, i)"
    >
      <img :src="src" alt="" loading="lazy" draggable="false" />
      <!-- 倒计时：右侧锚定，从左向右收缩 -->
      <span
        v-if="i === active && list.length > 1"
        :key="`p-${active}`"
        class="countdown"
        :style="{
          animationDuration: expandMs + 'ms',
          animationPlayState: paused ? 'paused' : 'running',
        }"
        @animationend="onCountdownEnd"
      />
    </button>

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
  display: flex;
  gap: 4px;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.seg {
  position: relative;
  flex: 1 1 0%;
  min-width: 0;
  border: none;
  padding: 0;
  background: var(--surface-2);
  overflow: hidden;
  cursor: pointer;
  /* 位移与缩放同帧：flex 简写整体过渡 */
  transition: flex 0.6s var(--ease-out), filter var(--dur-fast);

  /* 图片绝对铺满：容器变宽时裁切随动，无二段跳感 */
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

  &.on {
    flex: 4 1 0%;
    cursor: zoom-in;
  }
}

/* 倒计时条：右端固定，宽度从满向右收缩归零 */
.countdown {
  position: absolute;
  right: 8px;
  bottom: 8px;
  left: 8px;
  height: 3px;
  border-radius: 999px;
  background: linear-gradient(90deg, var(--primary), var(--primary-deep));
  box-shadow: 0 0 8px rgba(var(--primary-rgb), 0.6);
  transform-origin: right center;
  animation: countdown-shrink linear both;
  pointer-events: none;
}

@keyframes countdown-shrink {
  from { transform: scaleX(1); }
  to { transform: scaleX(0); }
}
</style>
