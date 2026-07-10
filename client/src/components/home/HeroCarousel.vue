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

/** 每条 item 的随机 3D 飞入角度 */
const flyAngles = ref({ rx: 8, ry: -40, rz: 6 });

function rollFlyAngles(): void {
  const rand = (min: number, max: number) => min + Math.random() * (max - min);
  const sign = () => (Math.random() < 0.5 ? -1 : 1);
  flyAngles.value = {
    rx: rand(10, 28) * sign(),
    ry: rand(30, 55) * sign(),
    rz: rand(4, 14) * sign(),
  };
}

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
    rollFlyAngles();
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
  rollFlyAngles();
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

    <!-- 右：3D 立体相册（常驻左倾 15°） -->
    <div class="hero-stage">
      <div
        :key="`${groupIndex}-${itemIndex}`"
        class="album"
        :style="{
          '--fly-rx': flyAngles.rx + 'deg',
          '--fly-ry': flyAngles.ry + 'deg',
          '--fly-rz': flyAngles.rz + 'deg',
        }"
      >
        <div
          v-for="(cover, i) in covers"
          :key="cover"
          class="album-card"
          :class="`slot-${slotOf(i)}`"
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

.hero.in .album {
  animation: album-in 0.85s var(--ease-spring) both;
}

@keyframes album-in {
  from {
    opacity: 0;
    transform:
      translate3d(140px, -30px, -220px)
      rotateX(var(--fly-rx)) rotateY(var(--fly-ry)) rotateZ(var(--fly-rz))
      scale(0.8);
  }
  to {
    opacity: 1;
    transform: rotateY(-15deg);
  }
}

.hero.out .album {
  animation: album-out 0.45s var(--ease-out) both;
}

@keyframes album-out {
  to {
    opacity: 0;
    transform:
      translate3d(-120px, 40px, -180px)
      rotateX(calc(var(--fly-rx) * -1)) rotateY(calc(var(--fly-ry) * -1)) rotateZ(calc(var(--fly-rz) * -1))
      scale(0.8);
  }
}

/* 相册单卡：slot-0 前排，slot-1/2 依次退后堆叠 */
.album-card {
  position: absolute;
  inset: 0;
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--border);
  background: var(--surface);
  box-shadow: var(--shadow), 0 30px 70px -20px rgba(var(--primary-rgb), 0.35);
  transform-style: preserve-3d;
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

.slot-0 { transform: translateZ(0); opacity: 1; z-index: 3; filter: none; }
.slot-1 {
  transform: translate3d(34px, 18px, -70px) rotateZ(3deg);
  opacity: 0.75;
  z-index: 2;
  filter: brightness(0.85) blur(0.5px);
}
.slot-2 {
  transform: translate3d(68px, 36px, -140px) rotateZ(6deg);
  opacity: 0.5;
  z-index: 1;
  filter: brightness(0.7) blur(1px);
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
