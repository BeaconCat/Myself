<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

export interface HeroItem {
  title: string;
  excerpt: string;
  /** 头图 1–3 张：立体相册每 photoMs 轮转一张，全部展示完才切下一条文本 */
  covers: string[];
  tag: string;
}

const props = withDefaults(
  defineProps<{ groups: HeroItem[][]; photoMs?: number }>(),
  { photoMs: 3000 },
);

const { t } = useI18n();

const groupIndex = ref(0);
const itemIndex = ref(0);
const photoIndex = ref(0);
const phase = ref<'in' | 'out'>('in');

const OUT_MS = 450;
const MAX_COVERS = 3;

const group = computed(() => props.groups[groupIndex.value] ?? []);
const item = computed(() => group.value[itemIndex.value] ?? group.value[0]);
const covers = computed(() => (item.value?.covers ?? []).slice(0, MAX_COVERS));

/** 标题拆字符，随机延迟模糊切入 */
const chars = computed(() =>
  [...(item.value?.title ?? '')].map((ch) => ({ ch, delay: Math.random() * 0.35 })),
);

/** 当前 item 内进度（含正在展示的这张） */
const progress = computed(() => {
  const total = Math.max(covers.value.length, 1);
  return ((photoIndex.value + 1) / total) * 100;
});

function swapItem(next: () => void): void {
  phase.value = 'out';
  window.setTimeout(() => {
    next();
    photoIndex.value = 0;
    phase.value = 'in';
  }, OUT_MS);
}

/** 相册每 tick 前进一张；照片放完 → 下一条文本；组内放完 → 下一组 */
function tick(): void {
  if (photoIndex.value < covers.value.length - 1) {
    photoIndex.value += 1;
    return;
  }
  swapItem(() => {
    if (itemIndex.value < group.value.length - 1) {
      itemIndex.value += 1;
    } else {
      itemIndex.value = 0;
      groupIndex.value = (groupIndex.value + 1) % props.groups.length;
    }
  });
}

/** 相册卡位置：0 = 前排，1/2 = 后排堆叠 */
function slotOf(i: number): number {
  const len = covers.value.length;
  return (i - photoIndex.value + len) % len;
}

let timer = 0;

onMounted(() => {
  timer = window.setInterval(tick, props.photoMs);
});

onBeforeUnmount(() => window.clearInterval(timer));
</script>

<template>
  <section class="hero" :class="phase">
    <!-- 左：大标题 + 简介 -->
    <div class="hero-text">
      <span class="hero-tag">{{ item?.tag }}</span>
      <h1 class="hero-title" aria-live="polite">
        <span
          v-for="(c, i) in chars"
          :key="`${groupIndex}-${itemIndex}-${i}`"
          class="char"
          :style="{ '--d': c.delay + 's' }"
        >{{ c.ch }}</span>
      </h1>
      <p class="hero-excerpt">{{ item?.excerpt }}</p>
      <button class="hero-btn">{{ t('hero.readMore') }}</button>
    </div>

    <!-- 右：3D 立体相册（常驻左倾 15°），卡片上下飞入/离场 -->
    <div class="hero-stage">
      <div :key="`${groupIndex}-${itemIndex}`" class="album">
        <div
          v-for="(cover, i) in covers"
          :key="cover"
          class="album-card"
          :class="`slot-${slotOf(i)}`"
          :style="{ '--stagger': slotOf(i) * 0.1 + 's' }"
        >
          <img :src="cover" :alt="item?.title" draggable="false" />
          <div class="card-glow" />
        </div>
      </div>

      <!-- 进度：分段点 + 连续进度条 -->
      <div class="progress">
        <div class="progress-track">
          <!-- key 随 item 重建，避免换条时进度条倒退回滚 -->
          <div
            :key="`${groupIndex}-${itemIndex}`"
            class="progress-fill"
            :style="{ width: phase === 'out' ? '100%' : progress + '%' }"
          />
        </div>
        <div class="progress-dots">
          <span v-for="(_, i) in group" :key="i" :class="{ on: i === itemIndex }" />
        </div>
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

.hero.in .char {
  animation: char-in var(--dur-slow) var(--ease-out) both;
  animation-delay: var(--d);
}

@keyframes char-in {
  from { opacity: 0; filter: blur(12px); transform: translateX(36px); }
  to { opacity: 1; filter: blur(0); transform: none; }
}

.hero.out .hero-title,
.hero.out .hero-excerpt,
.hero.out .hero-tag {
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

.hero.in .hero-excerpt,
.hero.in .hero-tag {
  animation: fade-up var(--dur-slow) var(--ease-out) both;
  animation-delay: 0.25s;
}

@keyframes fade-up {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: none; }
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
    transform: translateY(-2px);
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

/* 相册整体常驻左倾 15° */
.album {
  position: relative;
  width: min(100%, 430px);
  aspect-ratio: 4 / 3;
  transform-style: preserve-3d;
  transform: rotateY(-15deg);
}

/* 相册单卡：参考错位堆叠——前排居中，后排右上/左下探出 */
.album-card {
  position: absolute;
  inset: 0;
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--border);
  background: var(--surface);
  box-shadow: var(--shadow), 0 30px 70px -20px rgba(var(--primary-rgb), 0.35);
  transform-style: preserve-3d;
  /* 切换：水平换位（transform）+ 层级跳变（z-index 不可动画，瞬切） */
  transition:
    transform 0.8s var(--ease-spring),
    opacity 0.8s var(--ease-out),
    filter 0.8s ease;

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
}
.slot-1 {
  transform: translate3d(76px, -46px, -90px) scale(0.8);
  opacity: 0.85;
  z-index: 2;
  filter: brightness(0.8);
}
.slot-2 {
  transform: translate3d(-72px, 48px, -90px) scale(0.8);
  opacity: 0.85;
  z-index: 1;
  filter: brightness(0.72);
}

/* 入场：前卡自上坠入，右后卡自右滑入，左后卡自左滑入；级联延迟。
   单侧 keyframe：to 省略 → 落到各自 slot 的 transform */
.hero.in .album-card {
  animation-duration: 0.8s;
  animation-timing-function: var(--ease-spring);
  animation-fill-mode: both;
  animation-delay: var(--stagger);
}
.hero.in .slot-0 { animation-name: enter-top; }
.hero.in .slot-1 { animation-name: enter-right; }
.hero.in .slot-2 { animation-name: enter-left; }

@keyframes enter-top {
  from { opacity: 0; transform: translate3d(0, -130%, 40px) scale(0.8); }
}
@keyframes enter-right {
  from { opacity: 0; transform: translate3d(150%, -46px, -120px) scale(0.72); }
}
@keyframes enter-left {
  from { opacity: 0; transform: translate3d(-150%, 48px, -120px) scale(0.72); }
}

/* 出场：前卡向下坠离，两张后卡向上飞离；透明度前半段即降为 0 */
.hero.out .album-card {
  animation-duration: 0.4s;
  animation-timing-function: var(--ease-out);
  animation-fill-mode: both;
  animation-delay: calc(var(--stagger) * 0.4);
}
.hero.out .slot-0 { animation-name: leave-down; }
.hero.out .slot-1,
.hero.out .slot-2 { animation-name: leave-up; }

@keyframes leave-down {
  55% { opacity: 0; }
  to { opacity: 0; transform: translate3d(0, 120%, 0) scale(0.82); }
}
@keyframes leave-up {
  55% { opacity: 0; }
  to { opacity: 0; transform: translate3d(0, -120%, -90px) scale(0.74); }
}

.card-glow {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(120deg, transparent 30%, rgba(255, 255, 255, 0.14) 48%, transparent 62%),
    linear-gradient(180deg, transparent 60%, rgba(var(--primary-rgb), 0.18));
}

/* ===== 进度 ===== */
.progress {
  width: min(100%, 430px);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.progress-track {
  width: 100%;
  height: 4px;
  border-radius: 999px;
  background: var(--surface-2);
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--primary), var(--primary-deep));
  box-shadow: 0 0 8px rgba(var(--primary-rgb), 0.6);
  /* 每 3s 前进一格，线性推进读秒感 */
  transition: width 2.9s linear;
}

.hero.out .progress-fill { transition: width 0.4s var(--ease-out); }

.progress-dots {
  display: flex;
  gap: 8px;

  span {
    width: 8px;
    height: 8px;
    border-radius: 999px;
    background: var(--border);
    transition: all var(--dur) var(--ease-out);

    &.on {
      width: 26px;
      background: linear-gradient(90deg, var(--primary), var(--primary-deep));
      box-shadow: 0 0 8px rgba(var(--primary-rgb), 0.5);
    }
  }
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
}
</style>
