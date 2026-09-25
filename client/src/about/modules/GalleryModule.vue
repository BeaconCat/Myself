<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { GalleryData } from '../types';
import type { ModProps } from './props';
import KitIcon from '../parts/KitIcon.vue';
import ModHead from '../parts/ModHead.vue';
import Scene from '../parts/Scene.vue';

/**
 * 画廊（gallery）：mosaic = 不等宫格（首图 2×2），strip = 横滑。
 * 悬停放大并浮出标题/地点/日期；点击从缩略图位置 FLIP 飞入 lightbox，支持键盘左右与 Esc。
 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as GalleryData);
const { t } = useI18n();

const cells = ref<HTMLElement[]>([]);
const open = ref(false);
const index = ref(0);
const box = ref<Record<string, string>>({});
const stepping = ref(0);
const cur = computed(() => d.value.images[index.value]);

const reduce = () => window.matchMedia('(prefers-reduced-motion: reduce)').matches;

function target(): Record<string, string> {
  const vw = window.innerWidth;
  const vh = window.innerHeight;
  const w = Math.min(vw * 0.78, 1100);
  const h = Math.min(w * 0.62, vh * 0.72);
  return { left: `${(vw - w) / 2}px`, top: `${(vh - h) / 2 - 24}px`, width: `${w}px`, height: `${h}px` };
}

function rectOf(i: number): Record<string, string> | null {
  const el = cells.value[i];
  if (!el || !el.offsetParent) return null;
  const r = el.getBoundingClientRect();
  return { left: `${r.left}px`, top: `${r.top}px`, width: `${r.width}px`, height: `${r.height}px` };
}

async function show(i: number): Promise<void> {
  index.value = i;
  box.value = { ...(rectOf(i) ?? target()), transition: 'none', borderRadius: 'var(--r-md)' };
  open.value = true;
  window.addEventListener('keydown', onKey);
  await nextTick();
  requestAnimationFrame(() => requestAnimationFrame(() => { box.value = { ...target(), borderRadius: 'var(--r-lg)' }; }));
}

function close(): void {
  const back = rectOf(index.value);
  box.value = back ? { ...back, borderRadius: 'var(--r-md)' } : { ...box.value, opacity: '0' };
  open.value = false;
  window.removeEventListener('keydown', onKey);
}

function step(delta: number): void {
  const n = d.value.images.length;
  if (n < 2) return;
  stepping.value = delta;
  window.setTimeout(() => {
    index.value = (index.value + delta + n) % n;
    stepping.value = 0;
  }, reduce() ? 0 : 180);
}

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape') close();
  if (e.key === 'ArrowRight') step(1);
  if (e.key === 'ArrowLeft') step(-1);
}

onBeforeUnmount(() => window.removeEventListener('keydown', onKey));

const setCell = (el: unknown, i: number) => { if (el) cells.value[i] = el as HTMLElement; };
const pad = (n: number) => String(n).padStart(2, '0');
</script>

<template>
  <ModHead :title="title">{{ t('aboutKit.gallery.meta', { n: d.images.length }) }}</ModHead>
  <div class="ga" :class="{ strip: variant === 'strip' }">
    <button
      v-for="(im, i) in d.images"
      :key="i"
      :ref="(el) => setCell(el, i)"
      :class="[`g${i}`, { 'hide-m': i >= 6, 'hide-s': i >= 4 }]"
      :aria-label="im.title || t('aboutKit.gallery.photo', { n: i + 1 })"
      @click="show(i)"
    >
      <span class="sc"><Scene :src="im.src" :scene="im.scene" :alt="im.title" /></span>
      <figcaption v-if="im.title || im.place">{{ im.title }}<small>{{ [im.place, im.date].filter(Boolean).join(' · ') }}</small></figcaption>
      <span v-if="i === 5 && d.images.length > 6" class="more more-m">+{{ d.images.length - 6 }}</span>
      <span v-if="i === 3 && d.images.length > 4" class="more more-s">+{{ d.images.length - 4 }}</span>
    </button>
  </div>

  <Teleport to="body">
    <div class="ak ga-lb" :class="{ on: open }" :aria-hidden="!open" @click.self="close">
      <div v-if="cur" class="lb-img" :class="{ out: stepping !== 0 }" :style="{ ...box, '--dx': `${-stepping * 30}px` }">
        <Scene :src="cur.src" :scene="cur.scene" :alt="cur.title" />
      </div>
      <button class="lb-x" :aria-label="t('aboutKit.gallery.close')" @click="close"><KitIcon name="close" /></button>
      <div class="lb-ui">
        <button :aria-label="t('aboutKit.gallery.prev')" @click="step(-1)"><KitIcon name="left" /></button>
        <div class="lb-cap">
          <b>{{ cur?.title }}</b>
          <small>{{ [cur?.place, cur?.date].filter(Boolean).join(' · ') }} · {{ pad(index + 1) }} / {{ pad(d.images.length) }}</small>
        </div>
        <button :aria-label="t('aboutKit.gallery.next')" @click="step(1)"><KitIcon name="right" /></button>
      </div>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.ga {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  grid-auto-rows: clamp(110px, 12.5cqi, 150px);
  gap: 10px;

  button {
    position: relative;
    overflow: hidden;
    padding: 0;
    border-radius: var(--r-md);
    cursor: zoom-in;

    &::after {
      content: '';
      position: absolute;
      inset: 0;
      border-radius: inherit;
      box-shadow: inset 0 0 0 1px rgb(255 255 255 / 0.08);
      background: linear-gradient(180deg, transparent 55%, rgb(0 0 0 / 0.55));
      opacity: 0;
      transition: opacity var(--dur);
    }

    &:hover::after { opacity: 1; }
    &:hover .sc { transform: scale(1.08); }
    &:hover figcaption { opacity: 1; transform: none; }
  }

  .sc { position: absolute; inset: 0; transition: transform 0.9s var(--ease-out); }

  figcaption {
    position: absolute;
    left: 14px;
    right: 14px;
    bottom: 11px;
    z-index: 2;
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: 8px;
    font-size: 14px;
    font-weight: 500;
    text-align: left;
    color: #fff;
    opacity: 0;
    transform: translateY(6px);
    transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-out);

    small { font: 400 12px var(--ak-mono); opacity: 0.75; white-space: nowrap; }
  }

  .g0 { grid-column: span 2; grid-row: span 2; }
  .g3, .g6 { grid-column: span 2; }

  .more {
    position: absolute;
    inset: 0;
    z-index: 3;
    display: none;
    place-items: center;
    background: rgb(5 10 20 / 0.55);
    backdrop-filter: blur(3px);
    color: #fff;
    font: 600 20px var(--ak-mono);
  }
}

@container (max-width: 760px) {
  .ga { grid-template-columns: repeat(3, 1fr); grid-auto-rows: 110px; }
  .ga .g3, .ga .g6 { grid-column: span 1; }
  .ga .hide-m { display: none; }
  .ga .more-m { display: grid; }
}

@container (max-width: 420px) {
  .ga { grid-template-columns: 1fr 1fr; grid-auto-rows: 96px; }
  .ga .g0 { grid-column: span 2; grid-row: span 2; }
  .ga .hide-s { display: none; }
  .ga .more-m { display: none; }
  .ga .more-s { display: grid; }
}

.ga.strip {
  grid-template-columns: none;
  grid-auto-flow: column;
  grid-auto-columns: minmax(180px, 1fr);
  grid-auto-rows: 200px;
  overflow-x: auto;
  scrollbar-width: none;

  button { grid-column: auto !important; grid-row: auto !important; display: block !important; }
  .more { display: none !important; }
}

/* ---------- lightbox ---------- */
.ga-lb {
  position: fixed;
  inset: 0;
  z-index: 9500;
  display: grid;
  place-items: center;
  background: rgb(3 6 14 / 0);
  pointer-events: none;
  transition: background 0.45s var(--ease-out), backdrop-filter 0.45s;

  &.on { background: rgb(3 6 14 / 0.82); backdrop-filter: blur(14px); pointer-events: auto; }
}

.lb-img {
  position: fixed;
  overflow: hidden;
  box-shadow: 0 40px 100px -20px rgb(0 0 0 / 0.8);
  transition: left 0.55s var(--ease-out), top 0.55s var(--ease-out), width 0.55s var(--ease-out), height 0.55s var(--ease-out), border-radius 0.55s var(--ease-out), opacity 0.3s, transform 0.18s ease-in;

  > * { position: absolute; inset: 0; }
  &.out { opacity: 0; transform: translateX(var(--dx)) scale(0.98); }
}

.ga-lb:not(.on) .lb-img {
  visibility: hidden;
  transition: left 0.5s var(--ease-out), top 0.5s var(--ease-out), width 0.5s var(--ease-out), height 0.5s var(--ease-out), border-radius 0.5s var(--ease-out), opacity 0.3s, visibility 0s 0.55s;
}

.lb-ui {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 34px;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 18px;
  color: #fff;
  opacity: 0;
  transition: opacity 0.3s 0.2s;
}

.ga-lb.on .lb-ui { opacity: 1; }

.lb-ui button, .lb-x {
  display: grid;
  place-items: center;
  width: 40px;
  height: 40px;
  border-radius: var(--r-pill);
  color: #fff !important;
  background: rgb(255 255 255 / 0.08) !important;
  border: 1px solid rgb(255 255 255 / 0.12) !important;

  &:hover { background: rgb(255 255 255 / 0.16) !important; }
}

.lb-x { position: fixed; top: 24px; right: 24px; opacity: 0; transition: opacity 0.3s; }
.ga-lb.on .lb-x { opacity: 1; }

.lb-cap {
  min-width: 220px;
  text-align: center;

  b { display: block; font: 700 16px var(--font-serif); }
  small { font: 400 11.5px var(--ak-mono); opacity: 0.6; }
}
</style>
