<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
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

let timer = 0;

function restart(): void {
  window.clearInterval(timer);
  if (list.value.length < 2) return;
  timer = window.setInterval(() => {
    if (viewerOpen.value) return;
    if (props.autoplay === 'hover' && !hovering.value) return;
    active.value = (active.value + 1) % list.value.length;
  }, expandMs.value);
}

function onSegClick(e: MouseEvent, i: number): void {
  if (i !== active.value) {
    active.value = i;
    restart();
    return;
  }
  // 已展开：FLIP 飞出 Lightbox
  const img = (e.currentTarget as HTMLElement).querySelector('img');
  if (!img) return;
  const r = img.getBoundingClientRect();
  viewerRect.value = { left: r.left, top: r.top, width: r.width, height: r.height };
  viewerOpen.value = true;
}

watch(expandMs, restart);
onMounted(restart);
onBeforeUnmount(() => window.clearInterval(timer));
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
  flex: 1 1 0;
  min-width: 0;
  border: none;
  padding: 0;
  background: var(--surface-2);
  overflow: hidden;
  cursor: pointer;
  /* 挤压展开动画 */
  transition: flex-grow 0.6s var(--ease-out), filter var(--dur-fast);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    user-select: none;
  }

  &:not(.on) {
    filter: brightness(0.72);

    &:hover { filter: brightness(0.9); }
  }

  &.on {
    flex-grow: 4;
    cursor: zoom-in;
  }
}
</style>
