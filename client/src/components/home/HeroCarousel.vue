<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import HeroLightbox, { type OriginRect } from './HeroLightbox.vue';

export interface HeroItem {
  title: string;
  excerpt: string;
  /** 头图 1–3 张：立体相册逐张轮转，放完切下一条 */
  covers: string[];
  tag: string;
}

const props = withDefaults(
  defineProps<{ items: HeroItem[]; photoMs?: number }>(),
  { photoMs: 3000 },
);

const { t } = useI18n();

const itemIndex = ref(0);
const photoIndex = ref(0);
/** enter：卡片入场动画播放中；idle：静止（slot 换位过渡生效）；out：退场 */
const phase = ref<'enter' | 'idle' | 'out'>('enter');

const hovering = ref(false);
const viewerOpen = ref(false);

const OUT_MS = 330;
const ENTER_MS = 550;

const item = computed(() => props.items[itemIndex.value] ?? props.items[0]);
const covers = computed(() => (item.value?.covers ?? []).slice(0, 3));
const paused = computed(() => hovering.value || viewerOpen.value);

/** 标题拆字符，顺序模糊切入 */
const chars = computed(() =>
  [...(item.value?.title ?? '')].map((ch, i) => ({ ch, delay: i * 0.035 })),
);

/** 当前条展示总时长（进度胶囊填充用） */
const itemDurationMs = computed(() => Math.max(covers.value.length, 1) * props.photoMs);

let enterTimer = 0;

function settle(): void {
  window.clearTimeout(enterTimer);
  enterTimer = window.setTimeout(() => {
    if (phase.value === 'enter') phase.value = 'idle';
  }, ENTER_MS);
}

function swapToItem(next: number): void {
  if (next === itemIndex.value || phase.value === 'out') return;
  phase.value = 'out';
  window.setTimeout(() => {
    itemIndex.value = (next + props.items.length) % props.items.length;
    photoIndex.value = 0;
    phase.value = 'enter';
    settle();
  }, OUT_MS);
}

/** 自动步进：先轮照片，照片放完切下一条 */
function tick(): void {
  if (paused.value || phase.value === 'out') return;
  if (photoIndex.value < covers.value.length - 1) {
    photoIndex.value += 1;
  } else {
    swapToItem(itemIndex.value + 1);
  }
}

/** 手动切照片（悬停箭头） */
function stepPhoto(delta: number): void {
  const len = covers.value.length;
  if (len < 2) return;
  photoIndex.value = (photoIndex.value + delta + len) % len;
}

/** 相册卡位置：0 前排，1 右上后排，2 左下后排 */
function slotOf(i: number): number {
  const len = covers.value.length;
  return (i - photoIndex.value + len) % len;
}

/** Lightbox：卡片本体飞出放大；按图片索引查当前卡片矩形 */
const sectionEl = ref<HTMLElement | null>(null);

function cardRectOf(i: number): OriginRect | null {
  const src = covers.value[i];
  if (!src || !sectionEl.value) return null;
  const el = sectionEl.value.querySelector<HTMLElement>(
    `.album-card[data-src="${CSS.escape(src)}"]`,
  );
  if (!el) return null;
  const r = el.getBoundingClientRect();
  return { left: r.left, top: r.top, width: r.width, height: r.height };
}

function onViewerClose(finalIndex: number): void {
  viewerOpen.value = false;
  photoIndex.value = finalIndex;
  // 鼠标可能已不在轮播区（mouseleave 被遮罩吃掉），按真实悬停态重算
  hovering.value = sectionEl.value?.matches(':hover') ?? false;
}

let timer = 0;

onMounted(() => {
  settle();
  timer = window.setInterval(tick, props.photoMs);
});

onBeforeUnmount(() => {
  window.clearInterval(timer);
  window.clearTimeout(enterTimer);
});
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
    <div class="hero-text">
      <span class="hero-tag">{{ item?.tag }}</span>
      <h1 class="hero-title" aria-live="polite">
        <span
          v-for="(c, i) in chars"
          :key="`${itemIndex}-${i}`"
          class="char"
          :style="{ '--d': c.delay + 's' }"
        >{{ c.ch }}</span>
      </h1>
      <p class="hero-excerpt">{{ item?.excerpt }}</p>
      <button class="hero-btn">{{ t('hero.readMore') }}</button>
    </div>

    <!-- 右：3D 立体相册 -->
    <div class="hero-stage">
      <div :key="itemIndex" class="album">
        <div
          v-for="(cover, i) in covers"
          :key="cover"
          class="album-card"
          :class="`slot-${slotOf(i)}`"
          :data-src="cover"
          :style="{ '--stagger': slotOf(i) * 0.1 + 's' }"
          @click="slotOf(i) === 0 ? (viewerOpen = true) : (photoIndex = i)"
        >
          <img :src="cover" :alt="item?.title" draggable="false" />
          <div class="card-glow" />
        </div>

        <!-- 悬停左右切换 -->
        <template v-if="covers.length > 1">
          <button class="step prev" aria-label="上一张" @click.stop="stepPhoto(-1)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M15 5l-7 7 7 7" /></svg>
          </button>
          <button class="step next" aria-label="下一张" @click.stop="stepPhoto(1)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M9 5l7 7-7 7" /></svg>
          </button>
        </template>
      </div>

      <!-- 进度：胶囊条填满 → 跳下一点；点击切换 -->
      <div class="pills">
        <button
          v-for="(_, i) in items"
          :key="i"
          class="pill"
          :class="{ on: i === itemIndex }"
          :aria-label="`第 ${i + 1} 条`"
          @click="swapToItem(i)"
        >
          <!-- key 含 photoIndex：切到第 n 张即从 n/总数 处起算（负延迟跳进度） -->
          <span
            v-if="i === itemIndex"
            :key="`fill-${itemIndex}-${photoIndex}`"
            class="pill-fill"
            :style="{
              animationDuration: itemDurationMs + 'ms',
              animationDelay: -(photoIndex * props.photoMs) + 'ms',
              animationPlayState: paused ? 'paused' : 'running',
            }"
          />
        </button>
      </div>
    </div>

    <HeroLightbox
      v-if="viewerOpen"
      :images="covers"
      :start-index="photoIndex"
      :get-rect="cardRectOf"
      @change="photoIndex = $event"
      @close="onViewerClose"
    />
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
}

/* ===== 左侧文字 ===== */
.hero-text {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 18px;
}

.hero-tag {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  padding: 4px 12px;
  border-radius: 999px;
  color: var(--primary);
  background: rgba(var(--primary-rgb), 0.1);
  border: 1px solid rgba(var(--primary-rgb), 0.25);
}

.hero-title {
  font-size: clamp(30px, 4.4vw, 54px);
  line-height: 1.2;
  min-height: 2.4em;
}

.char {
  display: inline-block;
  white-space: pre;
}

.hero.enter .char {
  animation: char-in var(--dur-slow) var(--ease-out) both;
  animation-delay: var(--d);
}

@keyframes char-in {
  from { opacity: 0; filter: blur(12px); transform: translateX(36px); }
  to { opacity: 1; filter: blur(0); transform: none; }
}

.hero.out .hero-title,
.hero.out .hero-excerpt,
.hero.out .hero-tag,
.hero.out .hero-btn {
  animation: text-out 0.45s var(--ease-out) both;
}

@keyframes text-out {
  to { opacity: 0; filter: blur(10px); transform: translateX(-48px); }
}

.hero-excerpt {
  font-size: 16px;
  line-height: 1.8;
  color: var(--text-2);
  max-width: 46ch;
}

.hero.enter .hero-excerpt,
.hero.enter .hero-tag {
  animation: fade-up var(--dur-slow) var(--ease-out) both;
  animation-delay: 0.25s;
}

.hero.enter .hero-btn {
  animation: blur-up var(--dur-slow) var(--ease-out) both;
  animation-delay: 0.4s;
}

@keyframes fade-up {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: none; }
}

@keyframes blur-up {
  from { opacity: 0; filter: blur(10px); transform: translateY(16px); }
  to { opacity: 1; filter: blur(0); transform: none; }
}

.hero-btn {
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.04em;
  padding: 12px 32px;
  border-radius: 12px;
  border: none;
  color: #fff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.4);
  background: linear-gradient(180deg, var(--primary), var(--primary-deep));
  box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.45), inset 0 1px 0 rgba(255, 255, 255, 0.25);
  transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast), filter var(--dur-fast);

  &:hover {
    filter: brightness(1.08);
    transform: scale(1.05);
    box-shadow: 0 8px 24px rgba(var(--primary-rgb), 0.55), inset 0 1px 0 rgba(255, 255, 255, 0.25);
  }
}

/* ===== 右侧立体相册 ===== */
.hero-stage {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 22px;
  perspective: 1300px;
}

.album {
  position: relative;
  width: min(100%, 430px);
  aspect-ratio: 4 / 3;
  transform-style: preserve-3d;
  transform: rotateY(-15deg);
}

/* 前排居中，后排右上/左下错位 */
.album-card {
  position: absolute;
  inset: 0;
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--border);
  background: var(--surface);
  box-shadow: var(--shadow), 0 30px 70px -20px rgba(var(--primary-rgb), 0.35);
  transform-style: preserve-3d;
  /* 换位过渡（无弹性） */
  transition:
    transform 0.7s var(--ease-out),
    opacity 0.7s var(--ease-out),
    filter 0.7s ease;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    user-select: none;
  }
}

.slot-0 {
  transform: translate3d(0, 0, 0) scale(0.86);
  opacity: 1;
  z-index: 3;
  filter: none;
  cursor: zoom-in;
}

.slot-1 {
  transform: translate3d(76px, -46px, -90px) scale(0.8);
  opacity: 0.85;
  z-index: 2;
  filter: brightness(0.8);
  cursor: pointer;
}
.slot-2 {
  transform: translate3d(-104px, 72px, -90px) scale(0.74);
  opacity: 0.85;
  z-index: 1;
  filter: brightness(0.72);
  cursor: pointer;
}

/* 入场（仅 enter 阶段播放一次；无弹性）：前卡自上，右后卡自右，左后卡自左 */
.hero.enter .album-card {
  animation-duration: 0.5s;
  animation-timing-function: var(--ease-out);
  animation-fill-mode: both;
  animation-delay: var(--stagger);
}
.hero.enter .slot-0 { animation-name: enter-top; }
.hero.enter .slot-1 { animation-name: enter-right; }
.hero.enter .slot-2 { animation-name: enter-left; }

/* 不透明度前 25% 就拉满：凭空出现感；位移距离随时长同步收短 */
@keyframes enter-top {
  from { opacity: 0; transform: translate3d(0, -70%, 40px) scale(0.8); }
  25% { opacity: 1; }
}
@keyframes enter-right {
  from { opacity: 0; transform: translate3d(85%, -46px, -120px) scale(0.72); }
  25% { opacity: 0.85; }
}
@keyframes enter-left {
  from { opacity: 0; transform: translate3d(-85%, 72px, -120px) scale(0.7); }
  25% { opacity: 0.85; }
}

/* 出场：前卡向下，后卡向上，透明度前半段归零 */
.hero.out .album-card {
  animation-duration: 0.3s;
  animation-timing-function: var(--ease-out);
  animation-fill-mode: both;
  animation-delay: calc(var(--stagger) * 0.4);
}
.hero.out .slot-0 { animation-name: leave-down; }
.hero.out .slot-1,
.hero.out .slot-2 { animation-name: leave-up; }

/* 出场 30% 时间就完全透明：瞬隐；位移收短 */
@keyframes leave-down {
  30% { opacity: 0; }
  to { opacity: 0; transform: translate3d(0, 65%, 0) scale(0.82); }
}
@keyframes leave-up {
  30% { opacity: 0; }
  to { opacity: 0; transform: translate3d(0, -65%, -90px) scale(0.7); }
}

.card-glow {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(120deg, transparent 30%, rgba(255, 255, 255, 0.14) 48%, transparent 62%),
    linear-gradient(180deg, transparent 60%, rgba(var(--primary-rgb), 0.18));
}

/* 悬停左右切换按钮 */
.step {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  z-index: 5;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  border: 1px solid rgba(var(--primary-rgb), 0.3);
  background: var(--glass);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  color: var(--text);
  display: grid;
  place-items: center;
  opacity: 0;
  pointer-events: none;
  transition: opacity var(--dur-fast) ease, transform var(--dur-fast) var(--ease-out), background var(--dur-fast);

  svg { width: 18px; height: 18px; }

  &.prev { left: -58px; }
  &.next { right: -58px; }

  &:hover {
    background: rgba(var(--primary-rgb), 0.25);
    transform: translateY(-50%) scale(1.1);
  }
}

.album:hover .step {
  opacity: 1;
  pointer-events: auto;
}

/* ===== 进度胶囊 ===== */
.pills {
  display: flex;
  align-items: center;
  gap: 10px;
  /* 让开左下后排卡的探出范围 */
  margin-top: 46px;
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

  /* 激活：拉长成小圆柱条 */
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
  from { width: 0; }
  to { width: 100%; }
}

@media (max-width: 900px) {
  .hero {
    grid-template-columns: 1fr;
    gap: 28px;
    min-height: auto;
    padding-top: 8px;
  }

  .hero-text { align-items: center; text-align: center; }
  .hero-excerpt { font-size: 14px; }
  .album { transform: rotateY(-10deg); }

  /* 移动端常显切换按钮 */
  .step { opacity: 1; pointer-events: auto; }
}
</style>
