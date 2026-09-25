<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import CoverArt, { coverKindOf } from './CoverArt.vue';
import { EASE } from './choreo/engine';
import { F, makeGeom, tf } from './choreo/geom';
import type { CardEl, CardEls } from './choreo/types';
import { prefersReducedMotion } from './useHeroChoreo';

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
  /**
   * 已离地：teleport 移动节点会重置该帧的 CSS 过渡（样式瞬变），
   * 光效等类驱动的缓动要等下一帧此标记置位后才触发
   */
  lifted: boolean;
}

const props = withDefaults(defineProps<{
  /** 头图（≤3 张）；`css:<kind>` 或无图占位渲染为 CSS 光影封面 */
  covers: string[];
  /** 前排卡序号（由父级轮播状态机持有） */
  photoIndex: number;
  /** 条目序号（编舞采集用） */
  idx: number;
  /** 图片 alt */
  title: string;
  /** 离场层：压在新卡组之上（同槽位 +5），不接事件 */
  leaving?: boolean;
  /** 切换进行中：暂不接手势 */
  busy?: boolean;
  /** 组内轮转动画倍速 */
  speed?: number;
}>(), { leaving: false, busy: false, speed: 1 });

const emit = defineEmits<{
  /** 相册轮转：点击后排卡 / lightbox 切换 / 关闭落地后归位 */
  'update:photoIndex': [index: number];
  /** lightbox 开合（父级据此暂停轮播） */
  'update:lightbox': [on: boolean];
  /** 悬停左右切换按钮 */
  step: [delta: number];
}>();

const { t } = useI18n();

const zoneEl = ref<HTMLElement | null>(null);
const albumEl = ref<HTMLElement | null>(null);
const bgEl = ref<HTMLElement | null>(null);
const spillEl = ref<HTMLElement | null>(null);
const cardEls: (HTMLElement | null)[] = [];

const kinds = computed(() => props.covers.map(coverKindOf));

/* ===== 几何：随相册实际宽度 / 断点等比缩放 ===== */
const albumW = ref(540);
const mobile = ref(false);
const geo = computed(() => makeGeom(albumW.value, mobile.value));

let ro: ResizeObserver | null = null;
const mq = window.matchMedia('(max-width: 900px)');

function measure(): void {
  mobile.value = mq.matches;
  if (zoneEl.value) albumW.value = zoneEl.value.offsetWidth || albumW.value;
}

/** 相册卡位置：0 前排，1 右上后排，2 左下后排 */
function slotOf(i: number, photo = props.photoIndex): number {
  const len = props.covers.length;
  return (i - photo + len) % len;
}

function restStyle(i: number): Record<string, string> {
  const s = slotOf(i);
  return {
    transform: geo.value.T(s),
    filter: F(geo.value.bright(s)),
    zIndex: String(geo.value.restZ(s) + (props.leaving ? 5 : 0)),
  };
}

/*
 * ===== 组内轮转：前卡「后退 → 让位 → 归队」，其余卡错峰晋升一位 =====
 * 旧实现的生涩来自两点：① photoIndex 一变，restStyle 立即把前卡 z-index 降到最底，
 * 它此刻仍在最前位置，于是第一帧就被后卡「啪」地盖住；② 退场位移只有 40px 且用回弹曲线，
 * 读起来像闪了一下。现在：
 *   - 离场前卡在关键帧里自带 zIndex：前 34% 仍压在最上层，同时原地后退、缩小、变暗；
 *     在最暗最小的那一刻交换层级（此时晋升卡已接近到位，交换被「从它身后穿过」掩盖），
 *     再沿弧线滑入后排槽位；
 *   - 晋升卡按目标槽位错峰 70ms 出发，用无回弹的品牌缓动，整组读作一次连贯的洗牌。
 */
const rot: Record<number, Animation | undefined> = {};
const ROT_MS = 900;
const ROT_STAGGER = 70;

watch(() => props.photoIndex, (now, before) => {
  const len = props.covers.length;
  if (len < 2 || props.leaving || prefersReducedMotion()) return;
  const g = geo.value;
  const topZ = g.restZ(0) + 5;
  for (let i = 0; i < len; i++) {
    const el = cardEls[i];
    if (!el || flights[i]) continue;
    const from = slotOf(i, before);
    const to = slotOf(i, now);
    if (from === to) continue;
    rot[i]?.cancel();
    const leavingFront = from === 0;
    const kf: Keyframe[] = leavingFront
      ? [
        { transform: g.T(0), filter: F(1), zIndex: topZ, easing: EASE.brand },
        // 原地后退：仍在最上层，缩小变暗，给晋升卡让出视线
        { transform: g.T(0, { dx: -26 * g.k, dy: 10 * g.k, dz: -60 * g.k, ds: 0.9 }), filter: F(0.72), zIndex: topZ, offset: 0.34 },
        // 最暗最小处交换层级，随后沿弧线归队
        { transform: g.T(0, { dx: -30 * g.k, dy: 12 * g.k, dz: -70 * g.k, ds: 0.88 }), filter: F(0.7), zIndex: g.restZ(to), offset: 0.36, easing: EASE.expoOut },
        { transform: g.T(to), filter: F(g.bright(to)), zIndex: g.restZ(to) },
      ]
      : [
        { transform: g.T(from), filter: F(g.bright(from)) },
        { transform: g.T(to), filter: F(g.bright(to)) },
      ];
    const a = el.animate(kf, {
      duration: ROT_MS,
      delay: leavingFront ? 0 : to * ROT_STAGGER,
      easing: leavingFront ? 'linear' : EASE.brand,
      fill: 'backwards',
    });
    a.playbackRate = props.speed;
    rot[i] = a;
  }
});

function stopRotation(i: number): void {
  rot[i]?.cancel();
  delete rot[i];
}

/* ===== 编舞采集 ===== */
function els(): CardEls {
  const cards: CardEl[] = props.covers
    .map((_, i) => {
      const el = cardEls[i]!;
      const sheet = el.querySelector<HTMLElement>(':scope > .sheet')!;
      const shade = el.querySelector<HTMLElement>(':scope > .shade')!;
      const slot = slotOf(i);
      return { el, shade, sheet, cv: sheet.firstElementChild as HTMLElement, slot, b: geo.value.bright(slot) };
    })
    .sort((a, b) => a.slot - b.slot);
  // 编舞接管前先停掉组内轮转，避免两套动画叠加
  Object.keys(rot).forEach((k) => stopRotation(Number(k)));
  return {
    ...geo.value,
    root: zoneEl.value!,
    album: albumEl.value!,
    bg: bgEl.value!,
    spill: spillEl.value!,
    radius: getComputedStyle(cards[0].sheet).borderTopLeftRadius || '0px',
    cards,
    front: cards[0],
    backs: cards.slice(1),
    idx: props.idx,
  };
}

/* ===== Lightbox（原卡节点 teleport 飞行） ===== */
const flights = reactive<Record<number, Flight>>({});
const expandedIndex = ref<number | null>(null);
const lightboxOn = ref(false);
const lbClosing = ref(false);
const dragging = ref(false);

const FLY_MS = 550;
const FLY_EASE = 'cubic-bezier(0.2, 0.8, 0.3, 1)';

const isTouch = 'ontouchstart' in window;

const guide = computed(() => [
  isTouch ? t('viewer.pinch') : t('viewer.wheel'),
  t('viewer.drag'),
  t('viewer.exit'),
]);

function setPhotoIndex(i: number): void {
  emit('update:photoIndex', i);
}

function setLightbox(on: boolean): void {
  lightboxOn.value = on;
  emit('update:lightbox', on);
}

/** 中央态：规范函数表的恒等形式，与槽位变换逐函数插值 */
const CENTER_TF = tf({});

function rectStyle(r: Rect, slot: number): Record<string, string> {
  return {
    left: `${r.left}px`,
    top: `${r.top}px`,
    width: `${r.width}px`,
    height: `${r.height}px`,
    borderRadius: 'var(--r-lg)',
    filter: F(geo.value.bright(slot)),
  };
}

/** 家位 = 未变换的相册布局盒（真实 4:3，不用投影包围盒），姿态由槽位变换补 */
function homeRect(): Rect | null {
  const zone = zoneEl.value;
  if (!zone) return null;
  const r = zone.getBoundingClientRect();
  return { left: r.left, top: r.top, width: r.width, height: r.width * 0.75 };
}

/** 按图片自然比例计算中央目标框（CSS 光影封面按 4:3） */
function centerStyle(i: number): Promise<Record<string, string>> {
  const fit = (w0: number, h0: number): Record<string, string> => {
    const maxW = document.documentElement.clientWidth * 0.86;
    const maxH = window.innerHeight * 0.8;
    const ratio = Math.min(maxW / w0, maxH / h0);
    const w = w0 * ratio;
    const h = h0 * ratio;
    return {
      left: `${(document.documentElement.clientWidth - w) / 2}px`,
      top: `${(window.innerHeight - h) / 2}px`,
      width: `${w}px`,
      height: `${h}px`,
      borderRadius: '0px',
      filter: F(1),
    };
  };
  if (kinds.value[i]) return Promise.resolve(fit(4, 3));
  return new Promise((resolve) => {
    const probe = new Image();
    probe.onload = () => resolve(fit(probe.naturalWidth, probe.naturalHeight));
    probe.onerror = () => resolve(fit(4, 3));
    probe.src = props.covers[i];
  });
}

/** 原卡起飞：记录家位 → teleport（inline 定格原位）→ 下一帧飞向中央 */
async function launch(i: number): Promise<void> {
  const home = homeRect();
  if (!home || flights[i]) return;
  stopRotation(i);
  const slot = slotOf(i);
  flights[i] = {
    mode: 'toCenter',
    homeRect: home,
    lifted: false,
    // 起飞帧保持家位 3D 姿态，飞行中转正
    style: { ...rectStyle(home, slot), transform: geo.value.T(slot), transition: 'none' },
  };
  const target = await centerStyle(i);
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      const flight = flights[i];
      if (!flight || flight.mode !== 'toCenter') return;
      flight.lifted = true;
      flight.style = {
        ...target,
        transform: CENTER_TF,
        transition: `all ${FLY_MS}ms ${FLY_EASE}`,
      };
      window.setTimeout(() => {
        const f = flights[i];
        if (f && f.mode === 'toCenter') f.mode = 'center';
      }, FLY_MS);
    });
  });
}

/** 原卡回家：按当前槽位飞回（组内顺序可能已在 lightbox 中变化）→ 落地删除 flight */
function sendHome(i: number): void {
  const flight = flights[i];
  if (!flight || flight.mode === 'toHome') return;
  flight.mode = 'toHome';
  const slot = slotOf(i);
  flight.style = {
    ...rectStyle(flight.homeRect, slot),
    transform: geo.value.T(slot),
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
  setLightbox(true);
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
  const len = props.covers.length;
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
  }, SWAP_MS + 10);

  // 新卡：先占位（阻止背景轮转动画作用到它），再在中央淡入
  const home = homeRect();
  if (!home) return;
  stopRotation(to);
  flights[to] = {
    mode: 'toCenter',
    homeRect: home,
    lifted: true,
    style: { ...rectStyle(home, 0), opacity: '0', transition: 'none' },
  };
  // 背景牌组同步轮转：新卡的家位变为前排
  setPhotoIndex(to);
  expandedIndex.value = to;
  const target = await centerStyle(to);
  const flight = flights[to];
  if (!flight) return;
  flight.style = { ...target, transform: CENTER_TF, opacity: '0', transition: 'none' };
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      const f = flights[to];
      if (!f || f.mode !== 'toCenter') return;
      f.style = { ...f.style, opacity: '1', transition: `opacity ${SWAP_MS}ms ease` };
      window.setTimeout(() => {
        const g = flights[to];
        if (g && g.mode === 'toCenter') g.mode = 'center';
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
    setLightbox(false);
    lbClosing.value = false;
    expandedIndex.value = null;
  }, FLY_MS);
  // 牌组轮转等卡完全落地归位后再做，避免落地途中被瞬移
  window.setTimeout(() => {
    if (props.photoIndex !== i) setPhotoIndex(i);
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
  const cx = e.clientX - document.documentElement.clientWidth / 2;
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
    // 即时拖动：按下即可平移
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
      const cx = (a.x + b.x) / 2 - document.documentElement.clientWidth / 2;
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
const SCROLL_KEYS = new Set([' ', 'PageUp', 'PageDown', 'Home', 'End', 'ArrowUp', 'ArrowDown']);

function onKey(e: KeyboardEvent): void {
  if (!lightboxOn.value) return;
  if (SCROLL_KEYS.has(e.key)) e.preventDefault();
  if (e.key === 'Escape') closeLightbox();
  if (e.key === 'ArrowLeft') void lbGo(-1);
  if (e.key === 'ArrowRight') void lbGo(1);
}

function activateCard(i: number): void {
  if (flights[i]) return;
  if (slotOf(i) === 0) openLightbox(i);
  else setPhotoIndex(i);
}

/**
 * 点击穿透：相册态卡片全部 pointer-events: none，手势统一落在未变换的 album-zone 上，
 * 再按槽位前后顺序用投影矩形手动解析命中目标——绕开 3D 命中测试的不稳定。
 */
let tap: { x: number; y: number; t: number } | null = null;

function resolveCardAt(x: number, y: number): number | null {
  const len = props.covers.length;
  for (let slot = 0; slot < len; slot++) {
    const i = (props.photoIndex + slot) % len;
    const el = cardEls[i];
    if (!el) continue;
    const r = el.getBoundingClientRect();
    if (x >= r.left && x <= r.right && y >= r.top && y <= r.bottom) return i;
  }
  return null;
}

function onZoneDown(e: PointerEvent): void {
  // 切换按钮有自己的 click：zone 手势必须无视，否则一次点击双触发
  if ((e.target as HTMLElement).closest('.step') || props.busy || props.leaving) {
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
  measure();
  ro = new ResizeObserver(measure);
  if (zoneEl.value) ro.observe(zoneEl.value);
  mq.addEventListener('change', measure);
  window.addEventListener('keydown', onKey);
  window.addEventListener('wheel', onWheel, { passive: false });
});

onBeforeUnmount(() => {
  ro?.disconnect();
  mq.removeEventListener('change', measure);
  window.removeEventListener('keydown', onKey);
  window.removeEventListener('wheel', onWheel);
  Object.values(rot).forEach((a) => a?.cancel());
});

defineExpose({ els });
</script>

<template>
  <!-- 未变换的定位层：手势解析、切换按钮都在这里，不受 3D 透视挤压 -->
  <div
    ref="zoneEl"
    class="album-zone"
    :class="{ leaving, busy, lifting: Object.keys(flights).length > 0 }"
    :aria-hidden="leaving ? 'true' : undefined"
    @pointerdown="onZoneDown"
    @pointerup="onZoneUp"
  >
    <div ref="bgEl" class="deck-ambient" />
    <div ref="albumEl" class="album">
      <!-- 前卡门缝光外溢到地面：全站卡片唯一的品牌光（编舞只动画 opacity） -->
      <div ref="spillEl" class="deck-spill" aria-hidden="true" />
      <!-- teleport 开启时原节点整体搬到 body 飞行 -->
      <Teleport
        v-for="(cover, i) in covers"
        :key="`${i}-${cover}`"
        to="body"
        :disabled="!flights[i]"
      >
        <!-- 变换层（槽位 3D 姿态 / 层级）> 外观层（圆角裁切、开合）> 画面层 -->
        <div
          :ref="(el) => { cardEls[i] = el as HTMLElement | null; }"
          class="album-card"
          :class="flights[i]
            ? ['fly', {
              airborne: flights[i].lifted,
              settled: flights[i].mode === 'center',
              homing: flights[i].mode === 'toHome',
              dragging,
            }]
            : null"
          :style="flights[i]
            ? [flights[i].style, flights[i].mode === 'center' ? zoomStyle : {}]
            : restStyle(i)"
          @pointerdown="flights[i]?.mode === 'center' && onPointerDown($event)"
          @pointermove="onPointerMove"
          @pointerup="onPointerUp"
          @pointercancel="onPointerUp"
        >
          <!-- 中性阴影层：与外观层并列，不被开合 / 裁切编舞裁掉 -->
          <div class="shade" />
          <div class="sheet">
            <div class="cv">
              <CoverArt v-if="kinds[i]" :kind="kinds[i]!" />
              <img v-else :src="cover" :alt="title" draggable="false" />
            </div>
            <!-- 光照层跟卡走：飞行/中央态同样保留 -->
            <div class="card-glow" />
          </div>
        </div>
      </Teleport>
    </div>

    <!-- 悬停左右切换 -->
    <template v-if="covers.length > 1 && !leaving">
      <button class="step prev" aria-label="上一张" @click.stop="emit('step', -1)">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M15 5l-7 7 7 7" /></svg>
      </button>
      <button class="step next" aria-label="下一张" @click.stop="emit('step', 1)">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M9 5l7 7-7 7" /></svg>
      </button>
    </template>

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
  </div>
</template>

<style scoped lang="scss">
/*
 * 未变换的定位层。注意：这里及祖先都不能建立层叠上下文（不设 perspective / transform / z-index），
 * 卡片 z-index 需要与文字层、光效层、缩略导航在 Hero 内交错。
 */
.album-zone {
  position: relative;
  width: 100%;
  /* 光标定在容器级：3D 命中测试在层叠卡片间跳动时指针形态不变 */
  cursor: pointer;

  /* 透明外扩热区：垫底接收 hover，不挡卡片点击 */
  &::after {
    content: '';
    position: absolute;
    inset: -24px -80px;
    z-index: -1;
    transition: opacity var(--dur) ease;
  }

  &.leaving,
  &.busy { pointer-events: none; }

  /* 离场层热区：透明度过渡，不再按类名瞬切 display */
  &.leaving::after { opacity: 0; }
}

/* 卡组地面环境光：中性冷光（浅色为淡影），编舞驱动其明灭 */
.deck-ambient {
  position: absolute;
  left: -10%;
  right: -14%;
  top: 58%;
  bottom: -34%;
  pointer-events: none;
  background: radial-gradient(50% 42% at 52% 30%, rgb(170 195 255 / 0.14), transparent 72%);
  filter: blur(10px);
}

:root[data-mode='light'] .deck-ambient {
  background: radial-gradient(50% 42% at 52% 30%, rgb(20 50 120 / 0.12), transparent 72%);
}

/*
 * 前卡门缝光外溢到地面（品牌光，强调系统里唯一允许的卡片发光）。
 * 位于卡片之下、文字之下（z 4），相册态常亮；lightbox 起飞时随前卡离开而熄灭。
 */
.deck-spill {
  position: absolute;
  z-index: 4;
  left: 4%;
  right: -12%;
  top: 74%;
  height: 52%;
  pointer-events: none;
  background: radial-gradient(
    50% 50% at 42% 30%,
    color-mix(in srgb, var(--primary) 18%, rgb(220 230 255 / 0.5)),
    color-mix(in srgb, var(--primary) 16%, transparent) 50%,
    transparent 72%
  );
  filter: blur(18px);
  transition: opacity var(--dur-slow) var(--ease-out);
}

:root[data-mode='light'] .deck-spill {
  background: radial-gradient(50% 50% at 42% 30%, color-mix(in srgb, var(--primary) 22%, transparent), transparent 72%);
}

.album-zone.lifting .deck-spill { opacity: 0; }

.album {
  position: relative;
  width: 100%;
  aspect-ratio: 4 / 3;
}

/* 变换层：槽位姿态由 inline style（规范变换函数表）给出 */
.album-card {
  position: absolute;
  inset: 0;
  border-radius: var(--r-lg);
  transform-origin: 50% 50%;
  /* 相册态不接事件：手势统一由 album-zone 解析（3D 命中不稳定） */
  pointer-events: none;
}

/* 中性阴影层：外观层的兄弟节点（不被 clip-path / 开合裁掉），编舞只动画 opacity */
.shade {
  position: absolute;
  inset: 0;
  border-radius: var(--r-lg);
  pointer-events: none;
  box-shadow: 0 50px 90px -40px rgb(0 0 0 / 0.9), 0 0 0 0.5px rgb(255 255 255 / 0.08);
}

:root[data-mode='light'] .shade {
  box-shadow: 0 50px 90px -40px rgb(16 24 60 / 0.5), 0 0 0 0.5px rgb(16 24 40 / 0.08);
}

/* 外观层：圆角裁切；开合 / 裁切类编舞作用在这一层（不挂阴影） */
.sheet {
  position: absolute;
  inset: 0;
  border-radius: var(--r-lg);
  overflow: hidden;
  background: #050a14;
  outline: 1px solid rgba(255, 255, 255, 0.07);
  outline-offset: -1px;
}

/* 画面层：镜头回收 / 反向视差作用在这一层 */
.cv {
  position: absolute;
  inset: 0;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    user-select: none;
    display: block;
  }
}

/* 飞行态：原节点 teleport 到 body，fixed 定位 + inline 样式驱动（阴影 / 圆角挪到变换层） */
.album-card.fly {
  position: fixed;
  inset: auto;
  z-index: 9600;
  pointer-events: auto;
  opacity: 1;
  will-change: left, top, width, height, transform;
  touch-action: none;

  .shade { border-radius: inherit; }

  .sheet {
    border-radius: inherit;
    outline-color: transparent;
    transition: outline-color 0.3s ease;
  }

  &.settled { cursor: grab; }
  &.dragging { cursor: grabbing; }
}

/* 封面光照（品牌封面光影的一部分：斜向高光 + 底部一抹主色，只在画面内部） */
.card-glow {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(120deg, transparent 30%, rgba(255, 255, 255, 0.12) 48%, transparent 62%),
    linear-gradient(180deg, transparent 60%, rgba(var(--primary-rgb), 0.16));
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
  z-index: 40;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 0;
  box-shadow: inset 0 0 0 0.5px var(--line-2), 0 4px 12px -6px rgb(0 0 0 / 0.4);
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
    background: var(--fill-3);
    transform: translateY(-50%) scale(1.06);
  }

  &:active { transform: translateY(-50%) scale(0.97); }
  &:focus-visible { box-shadow: var(--focus); }
}

.album-zone:hover .step {
  opacity: 1;
  pointer-events: auto;
}

@media (max-width: 900px) {
  .step {
    opacity: 1;
    pointer-events: auto;
    width: 32px;
    height: 32px;

    svg { width: 15px; height: 15px; }

    &.prev { left: -40px; }
    &.next { right: -40px; }
  }
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

  &:hover { transform: scale(1.08) rotate(90deg); background: rgba(255, 255, 255, 0.18); }
}

.lb-nav {
  top: 50%;
  transform: translateY(-50%);
  width: 46px;
  height: 46px;

  svg { width: 22px; height: 22px; }

  &.prev { left: 20px; }
  &.next { right: 20px; }

  &:hover { transform: translateY(-50%) scale(1.08); background: rgba(255, 255, 255, 0.18); }
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
    border-radius: var(--r-pill);
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
