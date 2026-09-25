<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import MIcon from './MIcon.vue';
import { attachDrag, clamp, reducedMotion, rubber } from './gesture';
import { copyText, toast } from './shell';

/**
 * 全屏图片查看器：从缩略图位置放大展开（裁切形变，object-fit cover → 完整比例），
 * 左右滑切换（两端橡皮筋），下拉缩小变淡、松手关闭并飞回原缩略图。
 */
export interface ViewerItem {
  src: string;
  caption?: string;
  meta?: string;
}

const props = defineProps<{
  open: boolean;
  items: ViewerItem[];
  index: number;
  /** 第 i 张对应的缩略图元素（用于飞入 / 飞回） */
  origin?: (i: number) => HTMLElement | null;
}>();
const emit = defineEmits<{ 'update:open': [v: boolean]; 'update:index': [i: number] }>();
const { t } = useI18n();

type Phase = 'closed' | 'opening' | 'open' | 'closing';
const phase = ref<Phase>('closed');
const cur = ref(0);
const bg = ref(0);
const root = ref<HTMLElement | null>(null);

interface Rect { left: number; top: number; width: number; height: number; radius: number; opacity: number }
const flyer = ref<Rect | null>(null);
const flyerAnim = ref(false);
const trackX = ref(0);
const trackAnim = ref(false);
const drag = ref({ x: 0, y: 0, s: 1 });
const dragAnim = ref(false);
const aspects = ref<Record<number, number>>({});

const DUR = 450;
const vw = () => window.innerWidth;
const vh = () => window.innerHeight;

function aspectOf(i: number): number {
  if (aspects.value[i]) return aspects.value[i];
  const el = props.origin?.(i);
  const img = el instanceof HTMLImageElement ? el : el?.querySelector('img');
  if (img && img.naturalWidth) return img.naturalWidth / img.naturalHeight;
  return 1;
}

/** 完整比例居中的最终矩形 */
function fitRect(i: number): Rect {
  const a = aspectOf(i);
  const w = Math.min(vw(), vh() * a);
  const h = w / a;
  return { left: (vw() - w) / 2, top: (vh() - h) / 2, width: w, height: h, radius: 0, opacity: 1 };
}

/** 飞行动画的圆角跟随源缩略图（圆角由 --r-* token 决定，不写死） */
function radiusOf(el: Element): number {
  return parseFloat(getComputedStyle(el).borderTopLeftRadius) || 0;
}

function originRect(i: number): Rect | null {
  const el = props.origin?.(i);
  if (!el) return null;
  const r = el.getBoundingClientRect();
  if (!r.width || r.bottom < 40 || r.top > vh() - 40) return null;
  return { left: r.left, top: r.top, width: r.width, height: r.height, radius: radiusOf(el), opacity: 1 };
}

const next2 = (fn: () => void) => requestAnimationFrame(() => requestAnimationFrame(fn));

async function doOpen(): Promise<void> {
  cur.value = clamp(props.index, 0, props.items.length - 1);
  trackAnim.value = false;
  trackX.value = 0;
  drag.value = { x: 0, y: 0, s: 1 };
  const from = originRect(cur.value);
  const to = fitRect(cur.value);
  phase.value = 'opening';
  flyerAnim.value = false;
  flyer.value = from ?? { ...to, opacity: 0 };
  bg.value = 0;
  await nextTick();
  next2(() => {
    flyerAnim.value = true;
    flyer.value = to;
    bg.value = 1;
    window.setTimeout(() => {
      if (phase.value === 'opening') {
        phase.value = 'open';
        flyer.value = null;
      }
    }, reducedMotion() ? 0 : DUR);
  });
}

function doClose(): void {
  if (phase.value === 'closed' || phase.value === 'closing') return;
  const base = fitRect(cur.value);
  const d = drag.value;
  const start: Rect = {
    left: base.left + d.x + (base.width * (1 - d.s)) / 2,
    top: base.top + d.y + (base.height * (1 - d.s)) / 2,
    width: base.width * d.s,
    height: base.height * d.s,
    radius: 0,
    opacity: 1,
  };
  const target = originRect(cur.value);
  phase.value = 'closing';
  flyerAnim.value = false;
  flyer.value = start;
  next2(() => {
    flyerAnim.value = true;
    flyer.value = target ?? {
      left: start.left + start.width * 0.075,
      top: start.top + start.height * 0.075,
      width: start.width * 0.85,
      height: start.height * 0.85,
      /* 与 --r-lg 同比例：--r-base × 1.4 */
      radius: (parseFloat(getComputedStyle(document.documentElement).getPropertyValue('--r-base')) || 10) * 1.4,
      opacity: 0,
    };
    bg.value = 0;
    window.setTimeout(() => {
      phase.value = 'closed';
      flyer.value = null;
      drag.value = { x: 0, y: 0, s: 1 };
    }, reducedMotion() ? 0 : DUR);
  });
}

watch(
  () => props.open,
  (on) => (on ? void doOpen() : doClose()),
);

function requestClose(): void {
  emit('update:open', false);
}

function go(k: number): void {
  cur.value = clamp(k, 0, props.items.length - 1);
  trackAnim.value = true;
  trackX.value = 0;
  emit('update:index', cur.value);
}

function onLoad(i: number, e: Event): void {
  const img = e.target as HTMLImageElement;
  if (img.naturalWidth) aspects.value = { ...aspects.value, [i]: img.naturalWidth / img.naturalHeight };
}

/* 手势：横向切换 / 纵向下拉关闭 */
let off: (() => void) | null = null;
watch(root, (el) => {
  off?.();
  off = null;
  if (!el) return;
  off = attachDrag(el, {
    down: (e) => (phase.value === 'open' && !(e.target as HTMLElement).closest('[data-ui]') ? { ax: '' as '' | 'x' | 'y' } : null),
    begin: (c, dx, dy) => {
      c.ax = Math.abs(dx) > Math.abs(dy) ? 'x' : 'y';
      trackAnim.value = false;
      dragAnim.value = false;
    },
    move: (c, dx, dy) => {
      const n = props.items.length;
      if (c.ax === 'x') {
        const edge = (cur.value === 0 && dx > 0) || (cur.value === n - 1 && dx < 0);
        trackX.value = edge ? Math.sign(dx) * rubber(Math.abs(dx), 90) : dx;
      } else {
        const yy = dy >= 0 ? dy : -rubber(-dy, 40);
        drag.value = { x: dx * 0.6, y: yy, s: 1 - clamp(yy, 0, 400) / 1400 };
        bg.value = clamp(1 - yy / 420, 0.2, 1);
      }
    },
    end: (c, vx, vy, dx, dy) => {
      if (c.ax === 'x') {
        let k = cur.value;
        if (vx < -0.3 || dx < -vw() / 3) k += 1;
        else if (vx > 0.3 || dx > vw() / 3) k -= 1;
        go(k);
      } else if (dy > 110 || vy > 0.6) {
        requestClose();
      } else {
        dragAnim.value = true;
        drag.value = { x: 0, y: 0, s: 1 };
        bg.value = 1;
      }
    },
  });
});
onBeforeUnmount(() => off?.());

const item = computed(() => props.items[cur.value]);

async function share(): Promise<void> {
  const ok = await copyText(new URL(item.value?.src ?? '', window.location.origin).href);
  toast(ok ? t('mobile.linkCopied') : t('mobile.copyFailed'));
}

function flyerStyle(r: Rect) {
  return {
    left: `${r.left}px`,
    top: `${r.top}px`,
    width: `${r.width}px`,
    height: `${r.height}px`,
    borderRadius: `${r.radius}px`,
    opacity: r.opacity,
  };
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="phase !== 'closed'"
      ref="root"
      class="m-layer viewer"
      :class="phase"
      :style="{ '--vbg': bg }"
      @click.self="requestClose"
    >
      <div class="bg" :class="{ anim: !dragAnim && phase !== 'open' }" />
      <div
        class="track"
        :class="{ anim: trackAnim }"
        :style="{ transform: `translateX(${-cur * 100}vw) translateX(${trackX}px)` }"
      >
        <div v-for="(it, i) in items" :key="i" class="slide">
          <img
            v-if="Math.abs(i - cur) <= 1"
            :src="it.src"
            alt=""
            draggable="false"
            :class="{ hide: i === cur && phase !== 'open', anim: dragAnim }"
            :style="i === cur ? { transform: `translate(${drag.x}px, ${drag.y}px) scale(${drag.s})` } : undefined"
            @load="onLoad(i, $event)"
          />
        </div>
      </div>

      <div
        v-if="flyer && item"
        class="flyer"
        :class="{ anim: flyerAnim }"
        :style="flyerStyle(flyer)"
      >
        <img :src="item.src" alt="" draggable="false" />
      </div>

      <div class="ui">
        <div class="top" data-ui>
          <button class="btn m-tap" :aria-label="t('mobile.viewer.close')" @click="requestClose"><MIcon name="close" /></button>
          <span class="count">{{ cur + 1 }} / {{ items.length }}</span>
          <button class="btn m-tap" :aria-label="t('mobile.share')" @click="share"><MIcon name="share" /></button>
        </div>
        <div v-if="item?.caption || items.length > 1" class="cap" data-ui>
          <p v-if="item?.caption">{{ item.caption }}</p>
          <small v-if="item?.meta">{{ item.meta }}</small>
          <div v-if="items.length > 1" class="dots">
            <i v-for="(_, i) in items" :key="i" :class="{ on: i === cur }" />
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.viewer {
  position: fixed;
  inset: 0;
  z-index: 130;
  touch-action: none;
  overflow: hidden;
}

.bg {
  position: absolute;
  inset: 0;
  background: #000;
  opacity: var(--vbg);
  pointer-events: none;

  &.anim { transition: opacity 0.4s var(--ease-out); }
}

.opening .bg,
.closing .bg { transition: opacity 0.4s var(--ease-out); }

.track {
  position: absolute;
  inset: 0;
  display: flex;
  pointer-events: none;

  &.anim { transition: transform 0.45s var(--ease-out); }
}

.slide {
  flex: none;
  width: 100vw;
  height: 100%;
  display: grid;
  place-items: center;

  img {
    max-width: 100%;
    max-height: 100%;
    object-fit: contain;
    display: block;

    &.hide { opacity: 0; }
    &.anim { transition: transform 0.4s var(--m-ease-sheet); }
  }
}

.flyer {
  position: absolute;
  overflow: hidden;
  pointer-events: none;

  &.anim {
    transition:
      left 0.45s var(--m-ease-sheet),
      top 0.45s var(--m-ease-sheet),
      width 0.45s var(--m-ease-sheet),
      height 0.45s var(--m-ease-sheet),
      border-radius 0.45s var(--m-ease-sheet),
      opacity 0.35s var(--ease-out);
  }

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
}

.ui {
  position: absolute;
  inset: 0;
  pointer-events: none;
  opacity: var(--vbg);
  color: #fff;
  transition: opacity 0.3s var(--ease-out);

  > * { pointer-events: auto; }
}

.top {
  position: absolute;
  top: calc(var(--m-safe-t) + 12px);
  left: 16px;
  right: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: rgba(255, 255, 255, 0.12) !important;
  color: #fff !important;
  backdrop-filter: blur(12px);

  .m-ic { width: 20px; height: 20px; }
}

.count {
  font-family: var(--m-font-mono);
  font-size: 13px;
  padding: 6px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(12px);
}

.cap {
  position: absolute;
  left: 20px;
  right: 20px;
  bottom: calc(var(--m-safe-b) + 28px);

  p {
    font-size: 14.5px;
    line-height: 1.65;
    color: rgba(255, 255, 255, 0.88);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  small {
    display: block;
    margin-top: 6px;
    font-size: 12px;
    color: rgba(255, 255, 255, 0.5);
  }
}

.dots {
  display: flex;
  gap: 5px;
  justify-content: center;
  margin-top: 14px;

  i {
    width: 5px;
    height: 5px;
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.3);
    transition: all var(--dur) var(--ease-spring);

    &.on {
      width: 16px;
      background: #fff;
    }
  }
}
</style>
