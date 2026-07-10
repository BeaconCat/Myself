<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';

export interface HeroItem {
  title: string;
  excerpt: string;
  /** 头图 1–3 张：立体相册逐张轮转，放完切下一条 */
  covers: string[];
  tag: string;
}

interface Rect {
  left: number;
  top: number;
  width: number;
  height: number;
}

/** 飞行中的卡：teleport 到 body 的原节点状态 */
interface Flight {
  mode: 'toCenter' | 'center' | 'toHome';
  style: Record<string, string>;
  homeRect: Rect;
}

const props = withDefaults(
  defineProps<{ items: HeroItem[]; photoMs?: number }>(),
  { photoMs: 3000 },
);

const { t } = useI18n();

const itemIndex = ref(0);
const photoIndex = ref(0);
/** enter：卡片入场动画；idle：静止（slot 换位过渡）；out：退场 */
const phase = ref<'enter' | 'idle' | 'out'>('enter');

const hovering = ref(false);
const sectionEl = ref<HTMLElement | null>(null);

/* ===== Lightbox（原卡节点 teleport 飞行） ===== */
const flights = reactive<Record<number, Flight>>({});
const expandedIndex = ref<number | null>(null);
const lightboxOn = ref(false);
const lbClosing = ref(false);
const dragging = ref(false);

const OUT_MS = 330;
const ENTER_MS = 550;
const FLY_MS = 550;
const FLY_EASE = 'cubic-bezier(0.2, 0.8, 0.3, 1)';

const isTouch = 'ontouchstart' in window;

const item = computed(() => props.items[itemIndex.value] ?? props.items[0]);
const covers = computed(() => (item.value?.covers ?? []).slice(0, 3));
const paused = computed(() => hovering.value || lightboxOn.value);

const chars = computed(() =>
  [...(item.value?.title ?? '')].map((ch, i) => ({ ch, delay: i * 0.035 })),
);

const itemDurationMs = computed(() => Math.max(covers.value.length, 1) * props.photoMs);

const guide = computed(() => [
  isTouch ? t('viewer.pinch') : t('viewer.wheel'),
  t('viewer.drag'),
  t('viewer.exit'),
]);

/* ===== 轮播状态机 ===== */
let enterTimer = 0;
let timer = 0;

function settle(): void {
  window.clearTimeout(enterTimer);
  enterTimer = window.setTimeout(() => {
    if (phase.value === 'enter') phase.value = 'idle';
  }, ENTER_MS);
}

function swapToItem(next: number): void {
  if (next === itemIndex.value || phase.value === 'out' || lightboxOn.value) return;
  phase.value = 'out';
  window.setTimeout(() => {
    itemIndex.value = (next + props.items.length) % props.items.length;
    photoIndex.value = 0;
    phase.value = 'enter';
    settle();
  }, OUT_MS);
}

function tick(): void {
  if (paused.value || phase.value === 'out') return;
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

/** 相册卡位置：0 前排，1 右上后排，2 左下后排 */
function slotOf(i: number): number {
  const len = covers.value.length;
  return (i - photoIndex.value + len) % len;
}

/* ===== Lightbox 飞行 ===== */
function rectStyle(r: Rect): Record<string, string> {
  return {
    left: `${r.left}px`,
    top: `${r.top}px`,
    width: `${r.width}px`,
    height: `${r.height}px`,
    borderRadius: 'var(--radius-lg)',
  };
}

function captureRect(i: number): Rect | null {
  const src = covers.value[i];
  if (!src || !sectionEl.value) return null;
  const el = sectionEl.value.querySelector<HTMLElement>(
    `.album-card[data-src="${CSS.escape(src)}"]`,
  );
  if (!el) return null;
  const r = el.getBoundingClientRect();
  return { left: r.left, top: r.top, width: r.width, height: r.height };
}

/** 按图片自然比例计算中央目标框 */
function centerStyle(src: string): Promise<Record<string, string>> {
  return new Promise((resolve) => {
    const probe = new Image();
    probe.onload = () => {
      const maxW = window.innerWidth * 0.86;
      const maxH = window.innerHeight * 0.8;
      const ratio = Math.min(maxW / probe.naturalWidth, maxH / probe.naturalHeight);
      const w = probe.naturalWidth * ratio;
      const h = probe.naturalHeight * ratio;
      resolve({
        left: `${(window.innerWidth - w) / 2}px`,
        top: `${(window.innerHeight - h) / 2}px`,
        width: `${w}px`,
        height: `${h}px`,
        borderRadius: '0px',
      });
    };
    probe.onerror = () => resolve({
      left: `${window.innerWidth * 0.15}px`,
      top: `${window.innerHeight * 0.15}px`,
      width: `${window.innerWidth * 0.7}px`,
      height: `${window.innerHeight * 0.7}px`,
      borderRadius: '0px',
    });
    probe.src = src;
  });
}

/** 原卡起飞：记录家位 → teleport（inline 定格原位）→ 下一帧飞向中央 */
async function launch(i: number): Promise<void> {
  const home = captureRect(i);
  if (!home || flights[i]) return;
  flights[i] = {
    mode: 'toCenter',
    homeRect: home,
    style: { ...rectStyle(home), transition: 'none' },
  };
  const target = await centerStyle(covers.value[i]);
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      const flight = flights[i];
      if (!flight || flight.mode !== 'toCenter') return;
      flight.style = { ...target, transition: `all ${FLY_MS}ms ${FLY_EASE}` };
      window.setTimeout(() => {
        const f = flights[i];
        if (f && f.mode === 'toCenter') f.mode = 'center';
      }, FLY_MS);
    });
  });
}

/** 原卡回家：飞回家位 → 落地删除 flight（teleport 关闭，节点归位相册） */
function sendHome(i: number): void {
  const flight = flights[i];
  if (!flight || flight.mode === 'toHome') return;
  flight.mode = 'toHome';
  flight.style = {
    ...rectStyle(flight.homeRect),
    transition: `all ${FLY_MS}ms ${FLY_EASE}`,
  };
  window.setTimeout(() => {
    delete flights[i];
  }, FLY_MS + 20);
}

function openLightbox(i: number): void {
  if (lightboxOn.value) return;
  resetZoom();
  expandedIndex.value = i;
  lightboxOn.value = true;
  lbClosing.value = false;
  document.documentElement.style.overflow = 'hidden';
  void launch(i);
}

/** 切换：当前卡飞回原位，目标卡从其槽位飞入中央（双卡交叉飞行） */
function lbGo(delta: number): void {
  const from = expandedIndex.value;
  if (from === null || flights[from]?.mode !== 'center') return;
  const len = covers.value.length;
  const to = (from + delta + len) % len;
  if (to === from || flights[to]) return;
  resetZoom();
  sendHome(from);
  expandedIndex.value = to;
  void launch(to);
}

function closeLightbox(): void {
  const i = expandedIndex.value;
  if (i === null || lbClosing.value) return;
  lbClosing.value = true;
  resetZoom();
  sendHome(i);
  window.setTimeout(() => {
    lightboxOn.value = false;
    lbClosing.value = false;
    document.documentElement.style.overflow = '';
    // 看到哪张，落地后转到前排
    if (photoIndex.value !== i) photoIndex.value = i;
    expandedIndex.value = null;
    hovering.value = sectionEl.value?.matches(':hover') ?? false;
  }, FLY_MS);
}

/* ===== 中央态缩放/拖动 ===== */
const scale = ref(1);
const tx = ref(0);
const ty = ref(0);

const zoomStyle = computed(() => ({
  transform: `translate(${tx.value}px, ${ty.value}px) scale(${scale.value})`,
  transition: dragging.value ? 'none' : 'transform 0.08s linear',
}));

function resetZoom(): void {
  scale.value = 1;
  tx.value = 0;
  ty.value = 0;
}

function clampScale(v: number): number {
  return Math.min(8, Math.max(0.5, v));
}

function onWheel(e: WheelEvent): void {
  if (!lightboxOn.value) return;
  e.preventDefault();
  const factor = e.deltaY < 0 ? 1.12 : 1 / 1.12;
  const next = clampScale(scale.value * factor);
  const ratio = next / scale.value;
  const cx = e.clientX - window.innerWidth / 2;
  const cy = e.clientY - window.innerHeight / 2;
  tx.value = cx - (cx - tx.value) * ratio;
  ty.value = cy - (cy - ty.value) * ratio;
  scale.value = next;
}

const pointers = new Map<number, { x: number; y: number }>();
let holdTimer = 0;
let panReady = false;
let lastPt = { x: 0, y: 0 };
let pinchDist = 0;
const HOLD_MS = 180;

function onPointerDown(e: PointerEvent): void {
  if (expandedIndex.value === null || flights[expandedIndex.value]?.mode !== 'center') return;
  (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
  if (pointers.size === 1) {
    lastPt = { x: e.clientX, y: e.clientY };
    panReady = false;
    holdTimer = window.setTimeout(() => {
      panReady = true;
      dragging.value = true;
    }, HOLD_MS);
  } else if (pointers.size === 2) {
    window.clearTimeout(holdTimer);
    panReady = false;
    dragging.value = false;
    const [a, b] = [...pointers.values()];
    pinchDist = Math.hypot(a.x - b.x, a.y - b.y);
  }
}

function onPointerMove(e: PointerEvent): void {
  if (!pointers.has(e.pointerId)) return;
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
  if (pointers.size === 2) {
    const [a, b] = [...pointers.values()];
    const dist = Math.hypot(a.x - b.x, a.y - b.y);
    if (pinchDist > 0) {
      const next = clampScale(scale.value * (dist / pinchDist));
      const ratio = next / scale.value;
      const cx = (a.x + b.x) / 2 - window.innerWidth / 2;
      const cy = (a.y + b.y) / 2 - window.innerHeight / 2;
      tx.value = cx - (cx - tx.value) * ratio;
      ty.value = cy - (cy - ty.value) * ratio;
      scale.value = next;
    }
    pinchDist = dist;
    return;
  }
  if (panReady) {
    tx.value += e.clientX - lastPt.x;
    ty.value += e.clientY - lastPt.y;
  }
  lastPt = { x: e.clientX, y: e.clientY };
}

function onPointerUp(e: PointerEvent): void {
  pointers.delete(e.pointerId);
  window.clearTimeout(holdTimer);
  if (pointers.size < 2) pinchDist = 0;
  if (pointers.size === 0) {
    panReady = false;
    dragging.value = false;
  }
}

/* 双击空白退出 */
let lastTap = 0;
function onBackdropTap(e: MouseEvent): void {
  if (e.target !== e.currentTarget) return;
  const now = performance.now();
  if (now - lastTap < 320) closeLightbox();
  lastTap = now;
}

function onKey(e: KeyboardEvent): void {
  if (!lightboxOn.value) return;
  if (e.key === 'Escape') closeLightbox();
  if (e.key === 'ArrowLeft') lbGo(-1);
  if (e.key === 'ArrowRight') lbGo(1);
}

function onCardClick(i: number): void {
  if (flights[i]) return;
  if (slotOf(i) === 0) openLightbox(i);
  else photoIndex.value = i;
}

onMounted(() => {
  settle();
  timer = window.setInterval(tick, props.photoMs);
  window.addEventListener('keydown', onKey);
  window.addEventListener('wheel', onWheel, { passive: false });
});

onBeforeUnmount(() => {
  window.clearInterval(timer);
  window.clearTimeout(enterTimer);
  window.removeEventListener('keydown', onKey);
  window.removeEventListener('wheel', onWheel);
  document.documentElement.style.overflow = '';
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

    <!-- 右：3D 立体相册（按钮在旋转容器外，不受透视挤压） -->
    <div class="hero-stage">
      <div class="album-zone">
        <div :key="itemIndex" class="album">
          <!-- teleport 开启时原节点整体搬到 body 飞行 -->
          <Teleport
            v-for="(cover, i) in covers"
            :key="cover"
            to="body"
            :disabled="!flights[i]"
          >
            <div
              class="album-card"
              :class="flights[i]
                ? ['fly', { settled: flights[i].mode === 'center', dragging }]
                : [`slot-${slotOf(i)}`]"
              :data-src="cover"
              :style="flights[i]
                ? [flights[i].style, flights[i].mode === 'center' ? zoomStyle : {}]
                : { '--stagger': slotOf(i) * 0.1 + 's' }"
              @click="onCardClick(i)"
              @pointerdown="flights[i]?.mode === 'center' && onPointerDown($event)"
              @pointermove="onPointerMove"
              @pointerup="onPointerUp"
              @pointercancel="onPointerUp"
            >
              <img :src="cover" :alt="item?.title" draggable="false" />
              <div v-if="!flights[i]" class="card-glow" />
            </div>
          </Teleport>
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

    <!-- Lightbox 遮罩与控件（卡片是上面 teleport 出去的原节点） -->
    <Teleport to="body">
      <div v-if="lightboxOn" class="lb" :class="{ closing: lbClosing }" @click="onBackdropTap">
        <button class="lb-close" :aria-label="t('viewer.close')" @click="closeLightbox">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round">
            <path d="M6 6l12 12M18 6L6 18" />
          </svg>
        </button>

        <template v-if="covers.length > 1">
          <button class="lb-nav prev" aria-label="prev" @click.stop="lbGo(-1)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M15 5l-7 7 7 7" /></svg>
          </button>
          <button class="lb-nav next" aria-label="next" @click.stop="lbGo(1)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M9 5l7 7-7 7" /></svg>
          </button>
          <span class="lb-counter">{{ (expandedIndex ?? 0) + 1 }} / {{ covers.length }}</span>
        </template>

        <ul class="lb-guide">
          <li v-for="(line, i) in guide" :key="line" :style="{ '--i': i }">{{ line }}</li>
        </ul>
      </div>
    </Teleport>
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
  gap: 0;
}

/* 未旋转的定位层：切换按钮挂这里，不受 3D 透视影响 */
.album-zone {
  position: relative;
  width: min(100%, 430px);
  perspective: 1300px;
}

.album {
  position: relative;
  width: 100%;
  aspect-ratio: 4 / 3;
  transform-style: preserve-3d;
  transform: rotateY(-15deg);
}

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
    transform 0.7s var(--ease-out),
    opacity 0.7s var(--ease-out),
    filter 0.7s ease;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    user-select: none;
    display: block;
  }
}

/* 飞行态：原节点 teleport 到 body，fixed 定位 + inline 样式驱动 */
.album-card.fly {
  position: fixed;
  inset: auto;
  z-index: 9600;
  border: none;
  transform: none;
  filter: none;
  opacity: 1;
  box-shadow: 0 30px 80px -20px rgba(0, 0, 0, 0.55);
  will-change: left, top, width, height, transform;
  touch-action: none;

  &.settled { cursor: grab; }
  &.dragging { cursor: grabbing; }
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

/* 入场（无弹性；透明度前 25% 拉满，凭空出现） */
.hero.enter .album-card:not(.fly) {
  animation-duration: 0.5s;
  animation-timing-function: var(--ease-out);
  animation-fill-mode: both;
  animation-delay: var(--stagger);
}
.hero.enter .slot-0 { animation-name: enter-top; }
.hero.enter .slot-1 { animation-name: enter-right; }
.hero.enter .slot-2 { animation-name: enter-left; }

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

/* 出场：30% 时间透明，瞬隐 */
.hero.out .album-card:not(.fly) {
  animation-duration: 0.3s;
  animation-timing-function: var(--ease-out);
  animation-fill-mode: both;
  animation-delay: calc(var(--stagger) * 0.4);
}
.hero.out .slot-0 { animation-name: leave-down; }
.hero.out .slot-1,
.hero.out .slot-2 { animation-name: leave-up; }

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

/* 左右切换按钮（在 album-zone 上，无 3D 变形） */
.step {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  z-index: 5;
  width: 40px;
  height: 40px;
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

  &.prev { left: -56px; }
  &.next { right: -56px; }

  &:hover {
    background: rgba(var(--primary-rgb), 0.25);
    transform: translateY(-50%) scale(1.1);
  }
}

.album-zone:hover .step {
  opacity: 1;
  pointer-events: auto;
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
  .step { opacity: 1; pointer-events: auto; }
  .step.prev { left: -12px; }
  .step.next { right: -12px; }
  .pills { margin-top: 48px; }
}
</style>

<!-- Lightbox 遮罩与控件：teleport 到 body，脱离 scoped 作用域用全局样式 -->
<style lang="scss">
.lb {
  position: fixed;
  inset: 0;
  z-index: 9500;
  background: rgba(8, 8, 12, 0.82);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  animation: lb-in 0.35s ease both;

  &.closing { animation: lb-out 0.45s ease both; }
}

@keyframes lb-in { from { opacity: 0; } }
@keyframes lb-out { to { opacity: 0; } }

.lb-close,
.lb-nav {
  position: absolute;
  z-index: 9700;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.25);
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
  display: grid;
  place-items: center;
  cursor: pointer;
  transition: transform 0.2s ease, background 0.2s ease;

  svg { width: 20px; height: 20px; }
}

.lb-close {
  top: 22px;
  right: 22px;
  width: 44px;
  height: 44px;

  &:hover { transform: scale(1.12) rotate(90deg); background: rgba(var(--primary-rgb), 0.35); }
}

.lb-nav {
  top: 50%;
  transform: translateY(-50%);
  width: 46px;
  height: 46px;

  svg { width: 22px; height: 22px; }

  &.prev { left: 20px; }
  &.next { right: 20px; }

  &:hover { transform: translateY(-50%) scale(1.12); background: rgba(var(--primary-rgb), 0.35); }
}

.lb-counter {
  position: absolute;
  top: 30px;
  left: 50%;
  transform: translateX(-50%);
  color: rgba(255, 255, 255, 0.85);
  font-size: 14px;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.1em;
}

.lb-guide {
  position: absolute;
  left: 24px;
  bottom: 24px;
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none;

  li {
    color: rgba(255, 255, 255, 0.78);
    font-size: 13px;
    padding: 7px 14px;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.14);
    backdrop-filter: blur(8px);
    width: fit-content;
    animation: lb-guide-in 0.5s ease both;
    animation-delay: calc(0.3s + var(--i) * 0.12s);
  }
}

@keyframes lb-guide-in {
  from { opacity: 0; transform: translateX(-28px); filter: blur(6px); }
  to { opacity: 1; transform: none; filter: blur(0); }
}

.lb.closing .lb-guide li {
  animation: lb-guide-out 0.3s ease both;
  animation-delay: calc(var(--i) * 0.06s);
}

@keyframes lb-guide-out {
  to { opacity: 0; transform: translateX(-28px); filter: blur(6px); }
}

@media (max-width: 768px) {
  .lb-nav { display: none; }
  .lb-guide { left: 14px; bottom: 18px; }
}
</style>
