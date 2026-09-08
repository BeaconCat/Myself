<script lang="ts">
/** 对外类型沿用原入口，消费方（HomeView）无需改动导入 */
export type { HeroItem } from './hero/types';
</script>

<script setup lang="ts">
import { ref, toRef } from 'vue';
import type { HeroItem } from './hero/types';
import { useHeroRotation } from './hero/useHeroRotation';
import HeroText from './hero/HeroText.vue';
import HeroDeck from './hero/HeroDeck.vue';

const props = withDefaults(
  defineProps<{ items: HeroItem[]; photoMs?: number }>(),
  { photoMs: 3000 },
);

const sectionEl = ref<HTMLElement | null>(null);

const {
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
} = useHeroRotation(toRef(props, 'items'));

/** lightbox 关闭后按真实指针位置回填悬停态（遮罩期间 mouseleave 不可靠） */
function onLightbox(on: boolean): void {
  lightboxOn.value = on;
  if (!on) hovering.value = sectionEl.value?.matches(':hover') ?? false;
}
</script>

<template>
  <section
    ref="sectionEl"
    class="hero"
    :class="[phase, { paused }]"
    @mouseenter="hovering = true"
    @mouseleave="hovering = false"
  >
    <!-- 左：大标题 + 简介 -->
    <HeroText :item="item" :item-index="itemIndex" :phase="phase" />

    <!-- 右：3D 立体相册 + 进度胶囊 -->
    <div class="hero-stage">
      <HeroDeck
        :covers="covers"
        :photo-index="photoIndex"
        :item-index="itemIndex"
        :phase="phase"
        :title="item?.title ?? ''"
        @update:photo-index="photoIndex = $event"
        @update:lightbox="onLightbox"
        @step="stepPhoto"
      />

      <!-- 进度胶囊 -->
      <div class="pills">
        <button
          v-for="(_, i) in items"
          :key="i"
          class="pill"
          :class="{ on: i === itemIndex }"
          :aria-label="`第 ${i + 1} 条`"
          @click="swapToItem(i)"
        >
          <!-- 每段 = 一张照片：从 n/总数 填到 (n+1)/总数，填满触发步进 -->
          <span
            v-if="i === itemIndex"
            :key="`fill-${itemIndex}-${photoIndex}`"
            class="pill-fill"
            :style="{
              '--from': (photoIndex / covers.length) * 100 + '%',
              '--to': ((photoIndex + 1) / covers.length) * 100 + '%',
              animationDuration: props.photoMs + 'ms',
              animationPlayState: paused ? 'paused' : 'running',
            }"
            @animationend="onFillEnd"
          />
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped lang="scss">
.hero {
  min-height: 62vh;
  display: grid;
  grid-template-columns: 1.05fr 1fr;
  align-items: center;
  gap: 48px;
  perspective: 1300px;
  /* 展示区禁选中：双击（退出 lightbox 等）不再拉出文字选区 */
  user-select: none;
  /* 背景占位壳由 HomeView .hero-shell 常驻提供 */
}

/* ===== 右侧立体相册容器 ===== */
.hero-stage {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0;
}

/* ===== 进度胶囊 ===== */
.pills {
  display: flex;
  align-items: center;
  gap: 10px;
  /* 让开左下后排卡的探出范围 */
  margin-top: 64px;
}

.pill {
  position: relative;
  width: 9px;
  height: 9px;
  border-radius: 999px;
  border: none;
  padding: 0;
  background: var(--border);
  overflow: hidden;
  transition: width var(--dur) var(--ease-out), background var(--dur-fast), transform var(--dur-fast) var(--ease-out);

  &:hover { transform: scale(1.25); }

  &.on {
    width: 52px;
    background: var(--surface-2);

    &:hover { transform: none; }
  }
}

.pill-fill {
  position: absolute;
  inset: 0;
  width: 0;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--primary), var(--primary-deep));
  box-shadow: 0 0 8px rgba(var(--primary-rgb), 0.5);
  animation: pill-progress linear both;
}

@keyframes pill-progress {
  from { width: var(--from, 0%); }
  to { width: var(--to, 100%); }
}

/* 移动端：上下布局 */
@media (max-width: 900px) {
  .hero {
    grid-template-columns: 1fr;
    gap: 26px;
    min-height: auto;
  }

  .pills { margin-top: 40px; }
}
</style>
