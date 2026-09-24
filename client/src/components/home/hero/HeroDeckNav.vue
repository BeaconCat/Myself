<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import CoverArt, { coverKindOf } from './CoverArt.vue';
import { thumbOf } from '../../../api';
import type { HeroItem } from './types';
import type { NavEls } from './choreo/types';

/**
 * 卡组下方的常驻导航：进度胶囊（默认）或缩略导航（共享元素变形需要）。
 * 组内进度由 WAAPI 驱动（不受全局 reduced-motion / 幕布对 CSS 动画的干预），填满即步进。
 */
const props = defineProps<{
  items: HeroItem[];
  itemIndex: number;
  photoIndex: number;
  photoCount: number;
  /** 每张照片的停留时长（ms，1x） */
  photoMs: number;
  speed: number;
  /** 暂停计时（悬停 / lightbox / 幕布 / 混搭器暂停） */
  paused: boolean;
  /** 切换进行中：进度不走 */
  busy: boolean;
  mode: 'pills' | 'rail';
}>();

const emit = defineEmits<{
  go: [index: number];
  fillEnd: [];
}>();

const THUMB_W = 60;
const THUMB_GAP = 10;
const indX = (i: number): number => i * (THUMB_W + THUMB_GAP);

const fillEl = ref<HTMLElement | null>(null);
const indEl = ref<HTMLElement | null>(null);
const thumbEls: (HTMLElement | null)[] = [];

let anim: Animation | null = null;

/** v-for 内的静态 ref 会变成数组：用函数 ref，只收非空节点 */
function setFill(el: unknown): void {
  if (el) fillEl.value = el as HTMLElement;
}

function pct(k: number): string {
  return `${(k / Math.max(1, props.photoCount)) * 100}%`;
}

function restart(): void {
  anim?.cancel();
  anim = null;
  const el = fillEl.value;
  if (!el || !el.isConnected || props.busy) return;
  anim = el.animate(
    [{ width: pct(props.photoIndex) }, { width: pct(props.photoIndex + 1) }],
    { duration: props.photoMs, fill: 'forwards' },
  );
  anim.playbackRate = props.speed;
  const self = anim;
  anim.onfinish = () => {
    if (anim === self) emit('fillEnd');
  };
  if (props.paused) anim.pause();
}

watch(
  () => [props.itemIndex, props.photoIndex, props.photoCount, props.busy, props.photoMs, props.mode],
  () => void nextTick(restart),
);

watch(() => props.paused, (p) => {
  if (!anim) return;
  if (p) anim.pause();
  else anim.play();
});

watch(() => props.speed, (v) => {
  if (!anim) return;
  if (anim.playState === 'running') anim.updatePlaybackRate(v);
  else anim.playbackRate = v;
});

onMounted(restart);
onBeforeUnmount(() => anim?.cancel());

function nav(prevIdx: number): NavEls | null {
  if (props.mode !== 'rail') return null;
  return { thumb: (i) => thumbEls[i] ?? null, ind: indEl.value, indX, prevIdx };
}

defineExpose({ nav });
</script>

<template>
  <div v-if="mode === 'pills'" class="pills">
    <button
      v-for="(it, i) in items"
      :key="i"
      class="pill"
      :class="{ on: i === itemIndex }"
      :aria-label="`第 ${i + 1} 条：${it.title.split('|').join('')}`"
      @click="emit('go', i)"
    >
      <span v-if="i === itemIndex" :ref="setFill" class="pill-fill" :style="{ width: pct(photoIndex) }" />
    </button>
  </div>

  <div v-else class="rail">
    <button
      v-for="(it, i) in items"
      :key="i"
      :ref="(el) => { thumbEls[i] = el as HTMLElement | null; }"
      class="thumb"
      :class="{ on: i === itemIndex }"
      :aria-label="it.title.split('|').join('')"
      @click="emit('go', i)"
    >
      <CoverArt v-if="coverKindOf(it.covers[0] ?? '')" :kind="coverKindOf(it.covers[0] ?? '')!" />
      <img v-else :src="thumbOf(it.covers[0])" alt="" draggable="false" />
    </button>
    <span ref="indEl" class="rail-ind" :style="{ transform: `translateX(${indX(itemIndex)}px)` }">
      <span :ref="setFill" class="rail-fill" :style="{ width: pct(photoIndex) }" />
    </span>
  </div>
</template>

<style scoped lang="scss">
/* ===== 进度胶囊 ===== */
.pills {
  position: relative;
  z-index: 46;
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
  inset: 0 auto 0 0;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--primary), var(--primary-deep));
  box-shadow: 0 0 8px rgba(var(--primary-rgb), 0.5);
}

/* ===== 缩略导航 ===== */
.rail {
  position: relative;
  z-index: 46;
  display: flex;
  gap: 10px;
  padding: 6px;
  margin-top: 40px;
  border-radius: 14px;
  background: var(--glass);
  border: 1px solid var(--border);
  backdrop-filter: blur(14px) saturate(1.3);
  -webkit-backdrop-filter: blur(14px) saturate(1.3);
}

.thumb {
  position: relative;
  width: 60px;
  height: 45px;
  border-radius: 8px;
  overflow: hidden;
  border: 0;
  padding: 0;
  background: #050a14;
  opacity: 0.72;
  transition: opacity var(--dur-fast) ease, transform var(--dur-fast) var(--ease-out);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  &::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: inherit;
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.08);
  }

  &:hover { opacity: 1; transform: translateY(-2px); }
  &.on { opacity: 1; }
}

.rail-ind {
  position: absolute;
  left: 6px;
  top: 6px;
  width: 60px;
  height: 45px;
  border-radius: 8px;
  overflow: hidden;
  pointer-events: none;
  z-index: 2;
  box-shadow: inset 0 0 0 2px var(--primary), 0 0 16px rgba(var(--primary-rgb), 0.55);
  transition: transform var(--dur-slow) var(--ease-out);
}

.rail-fill {
  position: absolute;
  left: 0;
  bottom: 0;
  height: 3px;
  background: linear-gradient(90deg, var(--primary), var(--primary-deep));
}

@media (max-width: 900px) {
  .pills { margin-top: 40px; }
  .rail { margin-top: 28px; }
}
</style>
