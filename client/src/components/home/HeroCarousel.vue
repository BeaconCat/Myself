<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { watch } from 'vue';
import { useLoadingStore } from '../../stores/loading';

export interface HeroItem {
  title: string;
  excerpt: string;
  /** 头图 1–3 张：立体相册逐张轮转，放完切下一条 */
  covers: string[];
  tag: string;
  /** 阅读全文跳转目标 */
  slug?: string;
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
  /** 家位的 3D 姿态（相册角 + 槽位偏移缩放），起降两端 1:1 */
  homeTransform: string;
  /**
   * 已离地：teleport 移动节点会重置该帧的 CSS 过渡（样式瞬变），
   * 光效等类驱动的缓动要等下一帧此标记置位后才触发
   */
  lifted: boolean;
}

const props = withDefaults(
  defineProps<{ items: HeroItem[]; photoMs?: number }>(),
  { photoMs: 3000 },
);

const { t } = useI18n();
const router = useRouter();

const itemIndex = ref(0);
const photoIndex = ref(0);
/** enter：卡片入场动画；idle：静止（slot 换位过渡）；out：退场 */
const phase = ref<'enter' | 'idle' | 'out'>('enter');

const hovering = ref(false);
const sectionEl = ref<HTMLElement | null>(null);

/* ===== Lightbox（原卡节点 teleport 飞行） ===== */
const flights = reactive<Record<number, Flight>>({});
/** 刚归位的卡：一帧内禁 transition，防槽位变换回弹 */
const noTrans = reactive<Record<number, boolean>>({});
const expandedIndex = ref<number | null>(null);
const lightboxOn = ref(false);
const lbClosing = ref(false);
const dragging = ref(false);

/** enter 保持时长必须 ≥ 卡片动画 0.75s + 末卡级联 0.2s，过早摘类会尾段跳变闪烁 */
const ENTER_MS = 1000;
/** 出场启动后多久切入下一条（重叠期：出场透明度已归零但位移未播完） */
const ITEM_SWAP_MS = 380;
const FLY_MS = 550;
const FLY_EASE = 'cubic-bezier(0.2, 0.8, 0.3, 1)';

const isTouch = 'ontouchstart' in window;

const item = computed(() => props.items[itemIndex.value] ?? props.items[0]);
const covers = computed(() => (item.value?.covers ?? []).slice(0, 3));
const paused = computed(() => hovering.value || lightboxOn.value);

/**
 * 标题分词：只在空格/标点/`|` 标记处允许换行，词段内绝不从中间断开。
 * `|` 为编辑期换行标记（仅作可断点，不渲染）：一行放得下就完整一行。
 */
interface TitleWord {
  chars: { ch: string; delay: number }[];
  gap: boolean;
}

const titleWords = computed<TitleWord[]>(() => {
  const raw = (item.value?.title ?? '').split('|').join('​');
  const tokens = raw.match(/[^\s​，。：；！？、,.:;!?]+[，。：；！？、,.:;!?]?|\s+|​/g) ?? [];
  const words: TitleWord[] = [];
  let index = 0;
  for (const token of tokens) {
    if (token === '​') {
      words.push({ chars: [], gap: true });
      continue;
    }
    if (/^\s+$/.test(token)) {
      words.push({ chars: [{ ch: ' ', delay: index * 0.035 }], gap: true });
      index += 1;
      continue;
    }
    words.push({
      chars: [...token].map((ch) => ({ ch, delay: (index += 1) * 0.035 })),
      gap: false,
    });
  }
  return words;
});

const guide = computed(() => [
  isTouch ? t('viewer.pinch') : t('viewer.wheel'),
  t('viewer.drag'),
  t('viewer.exit'),
]);

/* ===== 轮播状态机 ===== */
let enterTimer = 0;

/**
 * enter → idle 的计时要等「幕布」揭开才起跑：
 * loading 覆盖期间 CSS 动画被全局暂停，但 JS 定时器照走，
 * 若不等待，揭幕时入场动画类已被摘掉，卡片动效直接跳终态。
 */
const loadingStore = useLoadingStore();

function curtainDown(): boolean {
  return loadingStore.bootOverlayVisible || loadingStore.routeOverlayVisible;
}

function settle(): void {
  window.clearTimeout(enterTimer);
  if (curtainDown()) {
    const stop = watch(
      () => curtainDown(),
      (down) => {
        if (down) return;
        stop();
        settle();
      },
    );
    return;
  }
  enterTimer = window.setTimeout(() => {
    if (phase.value === 'enter') phase.value = 'idle';
  }, ENTER_MS);
}

function swapToItem(next: number): void {
  if (next === itemIndex.value || phase.value === 'out' || lightboxOn.value) return;
  phase.value = 'out';
  // 出/入场重叠：出场透明度前 30% 已归零，中途即切数据开始入场，衔接更流畅
  window.setTimeout(() => {
    itemIndex.value = (next + props.items.length) % props.items.length;
    photoIndex.value = 0;
    phase.value = 'enter';
    settle();
  }, ITEM_SWAP_MS);
}

/** 自动步进由进度条驱动：当前段填满（animationend）才前进，与视觉天然同步 */
function onFillEnd(): void {
  if (phase.value === 'out' || lightboxOn.value) return;
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

/** 相册的 3D 姿态：飞行起降时保持同角度，避免瞬间拍平 */
function deckPose(): string {
  const angle = window.matchMedia('(max-width: 900px)').matches ? -10 : -15;
  return `perspective(1300px) rotateY(${angle}deg)`;
}

/** 相册态阴影：起飞帧与落地帧使用，和槽位卡完全一致 */
const DECK_SHADOW = 'var(--shadow), 0 30px 70px -20px rgba(var(--primary-rgb), 0.35)';
const CENTER_SHADOW = '0 30px 80px -20px rgba(0, 0, 0, 0.55)';

function rectStyle(r: Rect): Record<string, string> {
  return {
    left: `${r.left}px`,
    top: `${r.top}px`,
    width: `${r.width}px`,
    height: `${r.height}px`,
    borderRadius: 'var(--radius-lg)',
    boxShadow: DECK_SHADOW,
  };
}

/** 槽位复合变换：等价于卡在相册中的 3D 姿态 */
function slotTransform(slot: number): string {
  const pose = deckPose();
  if (slot === 1) return `${pose} translate3d(76px, -46px, -90px) scale(0.8)`;
  if (slot === 2) return `${pose} translate3d(-104px, 72px, -90px) scale(0.74)`;
  return `${pose} scale(0.86)`;
}

/**
 * 家位 = 未旋转的 album-zone 布局盒（真实 4:3，不用投影包围盒——
 * 那个比例被透视压过，落地时裁切比例对不上）。姿态由 homeTransform 补。
 */
function captureHome(i: number): { rect: Rect; transform: string } | null {
  // album-zone 自身无变换：其布局盒 = album 的真实布局盒。
  // 不能用 album 的 getBoundingClientRect——rotateY 透视投影会把中心推向右侧，
  // 起飞/落地按偏移中心定位就会向右错位并在归位交接帧闪跳。
  const zone = sectionEl.value?.querySelector<HTMLElement>('.album-zone');
  if (!zone) return null;
  const r = zone.getBoundingClientRect();
  return {
    rect: { left: r.left, top: r.top, width: r.width, height: r.width * 0.75 },
    transform: slotTransform(slotOf(i)),
  };
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
  const home = captureHome(i);
  if (!home || flights[i]) return;
  flights[i] = {
    mode: 'toCenter',
    homeRect: home.rect,
    homeTransform: home.transform,
    lifted: false,
    // 起飞帧保持家位 3D 姿态，飞行中转正
    style: { ...rectStyle(home.rect), transform: home.transform, transition: 'none' },
  };
  const target = await centerStyle(covers.value[i]);
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      const flight = flights[i];
      if (!flight || flight.mode !== 'toCenter') return;
      flight.lifted = true;
      flight.style = {
        ...target,
        transform: 'perspective(1300px) rotateY(0deg)',
        boxShadow: CENTER_SHADOW,
        transition: `all ${FLY_MS}ms ${FLY_EASE}`,
      };
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
    // 归途中转回家位姿态（含槽位偏移缩放），裁切比例与角度同步复原
    transform: flight.homeTransform,
    transition: `all ${FLY_MS}ms ${FLY_EASE}`,
  };
  window.setTimeout(() => {
    delete flights[i];
    // 归位首帧禁过渡：否则节点回相册后从无变换滑向槽位变换，会肉眼卡一下
    noTrans[i] = true;
    window.setTimeout(() => { delete noTrans[i]; }, 80);
  }, FLY_MS + 20);
}

function openLightbox(i: number): void {
  if (lightboxOn.value) return;
  resetZoom();
  expandedIndex.value = i;
  lightboxOn.value = true;
  lbClosing.value = false;
  void launch(i);
}

/**
 * 切换：中央原地标准淡切（旧卡淡出静默归位，新卡在中央淡入），
 * 背景牌组同步轮转占位——展开卡的家位始终是前排，关闭时落回中心位。
 */
const SWAP_MS = 240;

async function lbGo(delta: number): Promise<void> {
  const from = expandedIndex.value;
  if (from === null || flights[from]?.mode !== 'center') return;
  const len = covers.value.length;
  const to = (from + delta + len) % len;
  if (to === from || flights[to]) return;
  resetZoom();

  // 旧卡：中央淡出，随后静默归位（无飞行）
  const fromFlight = flights[from];
  fromFlight.mode = 'toHome';
  fromFlight.style = {
    ...fromFlight.style,
    opacity: '0',
    transition: `opacity ${SWAP_MS}ms ease`,
  };
  window.setTimeout(() => {
    delete flights[from];
    noTrans[from] = true;
    window.setTimeout(() => { delete noTrans[from]; }, 80);
  }, SWAP_MS + 10);

  // 背景牌组同步轮转：新卡的家位变为前排
  photoIndex.value = to;
  expandedIndex.value = to;

  // 新卡：直接在中央淡入（不从槽位飞入）
  const home = captureHome(to);
  if (!home) return;
  const target = await centerStyle(covers.value[to]);
  flights[to] = {
    mode: 'toCenter',
    homeRect: home.rect,
    homeTransform: home.transform,
    lifted: true,
    style: { ...target, boxShadow: CENTER_SHADOW, opacity: '0', transition: 'none' },
  };
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      const flight = flights[to];
      if (!flight || flight.mode !== 'toCenter') return;
      flight.style = {
        ...flight.style,
        opacity: '1',
        transition: `opacity ${SWAP_MS}ms ease`,
      };
      window.setTimeout(() => {
        const f = flights[to];
        if (f && f.mode === 'toCenter') f.mode = 'center';
      }, SWAP_MS);
    });
  });
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
    expandedIndex.value = null;
    hovering.value = sectionEl.value?.matches(':hover') ?? false;
  }, FLY_MS);
  // 牌组轮转等卡完全落地归位后再做，避免落地途中被瞬移
  window.setTimeout(() => {
    if (photoIndex.value !== i) photoIndex.value = i;
  }, FLY_MS + 140);
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
let panReady = false;
let lastPt = { x: 0, y: 0 };
let pinchDist = 0;

function onPointerDown(e: PointerEvent): void {
  if (expandedIndex.value === null || flights[expandedIndex.value]?.mode !== 'center') return;
  (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
  if (pointers.size === 1) {
    lastPt = { x: e.clientX, y: e.clientY };
    // 即时拖动：按下即可平移（指南文案仍写长按，直接拖同样响应）
    panReady = true;
    dragging.value = true;
  } else if (pointers.size === 2) {
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

/**
 * Lightbox 期间不动 overflow（滚动条消失会引发布局横移，干扰飞行定位），
 * 改为事件级锁滚动：滚轮已 preventDefault，这里拦截滚动类按键。
 */
const SCROLL_KEYS = new Set([
  ' ', 'PageUp', 'PageDown', 'Home', 'End', 'ArrowUp', 'ArrowDown',
]);

function onKey(e: KeyboardEvent): void {
  if (!lightboxOn.value) return;
  if (SCROLL_KEYS.has(e.key)) e.preventDefault();
  if (e.key === 'Escape') closeLightbox();
  if (e.key === 'ArrowLeft') lbGo(-1);
  if (e.key === 'ArrowRight') lbGo(1);
}

function activateCard(i: number): void {
  if (flights[i]) return;
  if (slotOf(i) === 0) openLightbox(i);
  else photoIndex.value = i;
}

/**
 * 强化点击穿透：相册态卡片全部 pointer-events: none，
 * 手势统一落在未变换的 album-zone 上，再按槽位前后顺序用
 * 投影矩形手动解析命中目标——彻底绕开 3D 命中测试的不稳定。
 */
let tap: { x: number; y: number; t: number } | null = null;

function resolveCardAt(x: number, y: number): number | null {
  const len = covers.value.length;
  for (let slot = 0; slot < len; slot++) {
    const i = (photoIndex.value + slot) % len;
    const src = covers.value[i];
    const el = sectionEl.value?.querySelector<HTMLElement>(
      `.album-card[data-src="${CSS.escape(src)}"]`,
    );
    if (!el) continue;
    const r = el.getBoundingClientRect();
    if (x >= r.left && x <= r.right && y >= r.top && y <= r.bottom) return i;
  }
  return null;
}

function onZoneDown(e: PointerEvent): void {
  // 切换按钮有自己的 click：zone 手势必须无视，否则一次点击双触发
  // （tap 解析命中外扩的后排卡 + 按钮 click 各走一步，3 张时 +2 ≡ -1 表现为倒退）
  if ((e.target as HTMLElement).closest('.step')) {
    tap = null;
    return;
  }
  tap = { x: e.clientX, y: e.clientY, t: performance.now() };
}

function onZoneUp(e: PointerEvent): void {
  if (
    tap
    && Math.hypot(e.clientX - tap.x, e.clientY - tap.y) < 8
    && performance.now() - tap.t < 600
  ) {
    const i = resolveCardAt(e.clientX, e.clientY);
    if (i !== null) activateCard(i);
  }
  tap = null;
}

onMounted(() => {
  settle();
  window.addEventListener('keydown', onKey);
  window.addEventListener('wheel', onWheel, { passive: false });
});

onBeforeUnmount(() => {
  window.clearTimeout(enterTimer);
  window.removeEventListener('keydown', onKey);
  window.removeEventListener('wheel', onWheel);
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
      <span :key="`tag-${itemIndex}`" class="hero-tag">{{ item?.tag }}</span>
      <h1 class="hero-title" aria-live="polite">
        <span
          v-for="(word, wi) in titleWords"
          :key="`${itemIndex}-${wi}`"
          class="word"
          :class="{ gap: word.gap }"
        ><span
          v-for="(c, ci) in word.chars"
          :key="ci"
          class="char"
          :style="{ '--d': c.delay + 's' }"
        >{{ c.ch }}</span></span>
      </h1>
      <p :key="`ex-${itemIndex}`" class="hero-excerpt">{{ item?.excerpt }}</p>
      <button
        :key="`btn-${itemIndex}`"
        class="hero-btn"
        @click="item?.slug && router.push(`/articles/${item.slug}`)"
      >{{ t('hero.readMore') }}</button>
    </div>

    <!-- 右：3D 立体相册（按钮在旋转容器外，不受透视挤压） -->
    <div class="hero-stage">
      <div class="album-zone" @pointerdown="onZoneDown" @pointerup="onZoneUp">
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
                ? ['fly', {
                  airborne: flights[i].lifted,
                  settled: flights[i].mode === 'center',
                  homing: flights[i].mode === 'toHome',
                  dragging,
                }]
                : [`slot-${slotOf(i)}`, { 'no-trans': noTrans[i] }]"
              :data-src="cover"
              :style="flights[i]
                ? [flights[i].style, flights[i].mode === 'center' ? zoomStyle : {}]
                : { '--stagger': slotOf(i) * 0.1 + 's' }"
              @pointerdown="flights[i]?.mode === 'center' && onPointerDown($event)"
              @pointermove="onPointerMove"
              @pointerup="onPointerUp"
              @pointercancel="onPointerUp"
            >
              <img :src="cover" :alt="item?.title" draggable="false" />
              <!-- 光照层跟卡走：飞行/中央态同样保留 -->
              <div class="card-glow" />
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
  /* 展示区禁选中：双击（退出 lightbox 等）不再拉出文字选区 */
  user-select: none;
  /* 背景占位壳由 HomeView .hero-shell 常驻提供 */
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

/* 词段整体不拆行；gap（空格/换行标记）处才允许换行 */
.word {
  display: inline-block;
  white-space: nowrap;

  &.gap { display: inline; white-space: normal; }
}

.char {
  display: inline-block;
  white-space: pre;
}

/* 文字入场绑元素挂载（key 随条目重建），不受状态机提前切 idle 影响，动画完整播完 */
.char {
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

/* 55% 时间即完全透明：与卡片同步瞬隐，支撑出/入场重叠切换 */
@keyframes text-out {
  55% { opacity: 0; filter: blur(10px); }
  to { opacity: 0; filter: blur(10px); transform: translateX(-48px); }
}

.hero-excerpt {
  font-size: 16px;
  line-height: 1.8;
  color: var(--text-2);
  max-width: 46ch;
}

.hero-excerpt,
.hero-tag {
  animation: fade-up var(--dur-slow) var(--ease-out) both;
  animation-delay: 0.25s;
}

.hero-btn {
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
  /* 光标定在容器级：3D 命中测试在层叠卡片间跳动时指针形态不变 */
  cursor: pointer;

  /* 透明外扩热区：垫底接收 hover，不挡卡片点击 */
  &::after {
    content: '';
    position: absolute;
    inset: -24px -80px;
    z-index: -1;
  }
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
  /* 相册态不接事件：手势统一由 album-zone 解析（3D 命中不稳定） */
  pointer-events: none;
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
  pointer-events: auto;
  /* 边框/阴影沿用相册卡并由 inline style 过渡，飞行两端 1:1 对齐 */
  filter: none;
  opacity: 1;
  will-change: left, top, width, height, transform;
  touch-action: none;

  &.settled { cursor: grab; }
  &.dragging { cursor: grabbing; }
}

/* 归位首帧禁过渡 */
.album-card.no-trans {
  transition: none;
}

.slot-0 {
  transform: translate3d(0, 0, 0) scale(0.86);
  opacity: 1;
  z-index: 3;
  filter: none;
  cursor: pointer;
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
  animation-duration: 0.75s;
  animation-timing-function: var(--ease-out);
  animation-fill-mode: both;
  animation-delay: var(--stagger);
}
.hero.enter .slot-0 { animation-name: enter-top; }
.hero.enter .slot-1 { animation-name: enter-right; }
.hero.enter .slot-2 { animation-name: enter-left; }

@keyframes enter-top {
  from { opacity: 0; transform: translate3d(0, -70%, 40px) rotateX(48deg) scale(0.8); }
  25% { opacity: 1; }
}
/* 后排卡绕 Y 轴翻正 */
@keyframes enter-right {
  from { opacity: 0; transform: translate3d(85%, -46px, -120px) rotateY(-62deg) scale(0.72); }
  25% { opacity: 0.85; }
}
@keyframes enter-left {
  from { opacity: 0; transform: translate3d(-85%, 72px, -120px) rotateY(62deg) scale(0.7); }
  25% { opacity: 0.85; }
}

/* 出场：30% 时间透明，瞬隐 */
.hero.out .album-card:not(.fly) {
  animation-duration: 0.85s;
  animation-timing-function: var(--ease-out);
  animation-fill-mode: both;
  animation-delay: calc(var(--stagger) * 0.4);
}
.hero.out .slot-0 { animation-name: leave-down; }
.hero.out .slot-1,
.hero.out .slot-2 { animation-name: leave-up; }

@keyframes leave-down {
  40% { opacity: 0; }
  to { opacity: 0; transform: translate3d(0, 65%, 0) rotateX(-42deg) scale(0.82); }
}
@keyframes leave-up {
  40% { opacity: 0; }
  to { opacity: 0; transform: translate3d(0, -65%, -90px) rotateX(38deg) scale(0.7); }
}

.card-glow {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(120deg, transparent 30%, rgba(255, 255, 255, 0.14) 48%, transparent 62%),
    linear-gradient(180deg, transparent 60%, rgba(var(--primary-rgb), 0.18));
  transition: opacity 0.55s ease;
}

/* 升空后光照渐隐（起飞帧不减——teleport 移动节点那帧过渡会被重置成瞬变）；归途渐显 */
.album-card.fly.airborne .card-glow { opacity: 0; }
.album-card.fly.homing .card-glow { opacity: 1; }

/* 左右切换按钮（在 album-zone 上，无 3D 变形） */
.step {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  z-index: 10;
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
  from { width: var(--from, 0%); }
  to { width: var(--to, 100%); }
}

@media (max-width: 900px) {
  .hero {
    grid-template-columns: 1fr;
    gap: 28px;
    min-height: auto;
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
  user-select: none;
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
