<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

export interface HeroItem {
  title: string;
  excerpt: string;
  cover: string;
  tag: string;
}

/** groups：每组多张卡；组内每 itemMs 轮转一张，每 groupMs 切换一组 */
const props = withDefaults(
  defineProps<{ groups: HeroItem[][]; itemMs?: number; groupMs?: number }>(),
  { itemMs: 3200, groupMs: 10000 },
);

const { t } = useI18n();

const groupIndex = ref(0);
const itemIndex = ref(0);
/** in：入场播放中/静止；out：出场播放中 */
const phase = ref<'in' | 'out'>('in');

const OUT_MS = 450;

const group = computed(() => props.groups[groupIndex.value] ?? []);
const item = computed(() => group.value[itemIndex.value] ?? group.value[0]);

/** 标题拆字符，配随机延迟实现「随机模糊切入」 */
const chars = computed(() =>
  [...(item.value?.title ?? '')].map((ch) => ({ ch, delay: Math.random() * 0.35 })),
);

function swap(next: () => void): void {
  phase.value = 'out';
  window.setTimeout(() => {
    next();
    phase.value = 'in';
  }, OUT_MS);
}

let itemTimer = 0;
let groupTimer = 0;

function nextItem(): void {
  if (group.value.length < 2) return;
  swap(() => {
    itemIndex.value = (itemIndex.value + 1) % group.value.length;
  });
}

function nextGroup(): void {
  swap(() => {
    groupIndex.value = (groupIndex.value + 1) % props.groups.length;
    itemIndex.value = 0;
  });
  restartItemTimer();
}

function restartItemTimer(): void {
  window.clearInterval(itemTimer);
  itemTimer = window.setInterval(nextItem, props.itemMs);
}

onMounted(() => {
  restartItemTimer();
  groupTimer = window.setInterval(nextGroup, props.groupMs);
});

onBeforeUnmount(() => {
  window.clearInterval(itemTimer);
  window.clearInterval(groupTimer);
});
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

    <!-- 右：3D 封面卡 -->
    <div class="hero-stage">
      <div :key="`${groupIndex}-${itemIndex}`" class="hero-card">
        <img :src="item?.cover" :alt="item?.title" draggable="false" />
        <div class="card-glow" />
      </div>
      <!-- 组内进度点 -->
      <div class="dots">
        <span
          v-for="(_, i) in group"
          :key="i"
          :class="{ on: i === itemIndex }"
        />
      </div>
    </div>
  </section>
</template>

<style scoped lang="scss">
.hero {
  min-height: 62vh;
  display: grid;
  grid-template-columns: 1.1fr 1fr;
  align-items: center;
  gap: 48px;
  perspective: 1200px;
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

/* 字符：随机延迟模糊切入（从右），出场整体飞左 */
.char {
  display: inline-block;
  white-space: pre;
}

.hero.in .char {
  animation: char-in var(--dur-slow) var(--ease-out) both;
  animation-delay: var(--d);
}

@keyframes char-in {
  from {
    opacity: 0;
    filter: blur(12px);
    transform: translateX(36px);
  }
  to {
    opacity: 1;
    filter: blur(0);
    transform: none;
  }
}

.hero.out .hero-title,
.hero.out .hero-excerpt,
.hero.out .hero-tag {
  animation: text-out 0.45s var(--ease-out) both;
}

@keyframes text-out {
  to {
    opacity: 0;
    filter: blur(10px);
    transform: translateX(-48px);
  }
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

/* ===== 右侧 3D 卡 ===== */
.hero-stage {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  transform-style: preserve-3d;
}

.hero-card {
  position: relative;
  width: min(100%, 440px);
  aspect-ratio: 4 / 3;
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--border);
  background: var(--surface);
  box-shadow: var(--shadow), 0 30px 70px -20px rgba(var(--primary-rgb), 0.35);
  transform-style: preserve-3d;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    user-select: none;
  }
}

.card-glow {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(120deg, transparent 30%, rgba(255, 255, 255, 0.14) 48%, transparent 62%),
    linear-gradient(180deg, transparent 60%, rgba(var(--primary-rgb), 0.18));
}

.hero.in .hero-card {
  animation: card-in 0.7s var(--ease-spring) both;
}

@keyframes card-in {
  from {
    opacity: 0;
    transform: translateX(120px) rotateY(-38deg) rotateX(6deg) scale(0.85);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

.hero.out .hero-card {
  animation: card-out 0.45s var(--ease-out) both;
}

@keyframes card-out {
  to {
    opacity: 0;
    transform: translateX(-90px) rotateY(32deg) rotateX(-4deg) scale(0.85);
  }
}

.dots {
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
}
</style>
