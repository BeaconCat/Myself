<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  ArrowLeft, ArrowRight, GalleryHorizontal, ImagePlus, LayoutDashboard, LayoutGrid, Columns3, Replace, Trash2, X,
} from 'lucide';
import { thumbOf, type MediaItem } from '../../../api';
import Icon from '../../ui/Icon.vue';
import StModal from '../../../views/admin/studio/StModal.vue';
import StSeg from '../../../views/admin/studio/StSeg.vue';
import MediaLibraryModal from '../../../views/admin/studio/MediaLibraryModal.vue';
import '../../../views/admin/studio/i18n';
import {
  GALLERY_GAPS, GALLERY_RATIOS, GAP_PX, defaultGallery,
  type GalleryData, type GalleryGap, type GalleryLayout, type GalleryRatio,
} from '../../../utils/embeds';

/**
 * 拼图编辑器：左侧画布与前台同一套 .md-gallery 样式所见即所得（拖动排序、点选编辑、× 移除），
 * 右侧面板调布局 / 列数 / 比例 / 间距 / 图注，以及当前图片的说明、替换、移位。图片从素材库多选添加（可就地上传）。
 */
const props = defineProps<{ open: boolean; value: GalleryData | null }>();
const emit = defineEmits<{ save: [data: GalleryData]; close: [] }>();
const { t } = useI18n();

const MAX = 30;
interface Tile { id: number; src: string; alt: string }
let uid = 0;

const tiles = ref<Tile[]>([]);
const layout = ref<GalleryLayout>('grid');
const cols = ref(3);
const ratio = ref<GalleryRatio>('4:3');
const gap = ref<GalleryGap>('m');
const caption = ref('');
const sel = ref(-1);
const lib = ref<'' | 'add' | 'replace'>('');

watch(
  () => props.open,
  (open) => {
    if (!open) return;
    const d = { ...defaultGallery(), ...(props.value ?? {}) };
    tiles.value = d.images.map((im) => ({ id: ++uid, ...im }));
    layout.value = d.layout;
    cols.value = d.cols;
    ratio.value = d.ratio;
    gap.value = d.gap;
    caption.value = d.caption;
    sel.value = -1;
    lib.value = '';
  },
  { immediate: true },
);

const LAYOUTS: { v: GalleryLayout; icon: typeof LayoutGrid }[] = [
  { v: 'grid', icon: LayoutGrid },
  { v: 'mosaic', icon: LayoutDashboard },
  { v: 'row', icon: GalleryHorizontal },
  { v: 'masonry', icon: Columns3 },
];

const style = computed(() => ({
  '--cols': String(Math.min(cols.value, Math.max(1, tiles.value.length))),
  '--gap': `${GAP_PX[gap.value]}px`,
  '--ratio': ratio.value === 'auto' ? 'auto' : ratio.value.replace(':', ' / '),
}));
const current = computed(() => tiles.value[sel.value] ?? null);

const src = (s: string): string => (s.startsWith('/uploads/') ? thumbOf(s) : s);

function onPicked(items: MediaItem[]): void {
  if (lib.value === 'replace' && current.value && items[0]) {
    current.value.src = items[0].url;
  } else {
    const room = MAX - tiles.value.length;
    tiles.value.push(...items.slice(0, room).map((m) => ({ id: ++uid, src: m.url, alt: '' })));
    // 首次添加：列数按张数给个合适的默认（2 张两列，3 张以上三列）
    if (!props.value?.images.length && tiles.value.length > 1) cols.value = Math.min(3, tiles.value.length);
  }
  lib.value = '';
}

function remove(i: number): void {
  tiles.value.splice(i, 1);
  if (sel.value >= tiles.value.length) sel.value = tiles.value.length - 1;
}

function move(i: number, d: number): void {
  const j = i + d;
  if (j < 0 || j >= tiles.value.length) return;
  const [x] = tiles.value.splice(i, 1);
  tiles.value.splice(j, 0, x);
  sel.value = j;
}

/* ---------- 拖动排序：经过哪张就实时换位，TransitionGroup 负责位移动画 ---------- */
const dragging = ref(-1);
function onDragStart(i: number, e: DragEvent): void {
  dragging.value = i;
  e.dataTransfer?.setData('text/plain', String(i));
  if (e.dataTransfer) e.dataTransfer.effectAllowed = 'move';
}
function onDragOver(i: number): void {
  if (dragging.value < 0 || dragging.value === i) return;
  const [x] = tiles.value.splice(dragging.value, 1);
  tiles.value.splice(i, 0, x);
  if (sel.value === dragging.value) sel.value = i;
  dragging.value = i;
}

function save(): void {
  emit('save', {
    images: tiles.value.map(({ src: s, alt }) => ({ src: s, alt: alt.trim() })),
    layout: layout.value,
    cols: cols.value,
    ratio: ratio.value,
    gap: gap.value,
    caption: caption.value.trim(),
  });
  emit('close');
}

/** 素材库开着时 Esc 只关素材库 */
function onClose(): void {
  if (!lib.value) emit('close');
}
</script>

<template>
  <StModal :open="open" wide panel-class="gal-ed" @close="onClose">
    <div class="ge">
      <section class="stage">
        <header>
          <h3>{{ t('studio.gallery.title') }}</h3>
          <span class="n mono">{{ t('studio.gallery.count', { n: tiles.length, max: MAX }) }}</span>
        </header>
        <div class="canvas" @click.self="sel = -1">
          <figure v-if="tiles.length" class="md-gallery" :data-layout="layout" :data-n="Math.min(tiles.length, 9)" :style="style">
            <TransitionGroup tag="div" class="gg" name="ge-move">
              <div
                v-for="(tl, i) in tiles"
                :key="tl.id"
                class="gi"
                :class="{ on: sel === i, drag: dragging === i }"
                draggable="true"
                @dragstart="onDragStart(i, $event)"
                @dragenter.prevent="onDragOver(i)"
                @dragover.prevent
                @dragend="dragging = -1"
                @click.stop="sel = i"
              >
                <img :src="src(tl.src)" :alt="tl.alt" draggable="false" />
                <span class="no mono">{{ i + 1 }}</span>
                <button type="button" class="rm" :title="t('studio.delete')" @click.stop="remove(i)"><Icon :icon="X" :size="14" /></button>
              </div>
            </TransitionGroup>
            <figcaption v-if="caption.trim()">{{ caption }}</figcaption>
          </figure>
          <button v-else type="button" class="add-big" @click="lib = 'add'">
            <Icon :icon="ImagePlus" :size="28" />
            <b>{{ t('studio.gallery.addFirst') }}</b>
            <small>{{ t('studio.gallery.addFirstSub') }}</small>
          </button>
        </div>
        <footer>
          <span>{{ t('studio.gallery.hint') }}</span>
          <button type="button" class="st-btn sm" :disabled="tiles.length >= MAX" @click="lib = 'add'">
            <Icon :icon="ImagePlus" :size="15" />{{ t('studio.gallery.add') }}
          </button>
        </footer>
      </section>

      <aside class="panel">
        <div class="grp">
          <div class="st-flabel">{{ t('studio.gallery.layoutLabel') }}</div>
          <div class="layouts">
            <button v-for="l in LAYOUTS" :key="l.v" type="button" :class="{ on: layout === l.v }" @click="layout = l.v">
              <Icon :icon="l.icon" :size="20" />
              <b>{{ t(`studio.gallery.layout.${l.v}`) }}</b>
              <small>{{ t(`studio.gallery.layoutSub.${l.v}`) }}</small>
            </button>
          </div>
        </div>

        <div v-if="layout !== 'row'" class="grp">
          <div class="st-flabel">{{ t('studio.gallery.cols') }}</div>
          <StSeg v-model="cols" :options="[2, 3, 4, 5].map((n) => ({ value: n, label: String(n) }))" />
        </div>

        <div v-if="layout !== 'masonry'" class="grp">
          <div class="st-flabel">{{ t('studio.gallery.ratio') }}</div>
          <div class="chips">
            <button v-for="r in GALLERY_RATIOS" :key="r" type="button" class="st-chip" :class="{ on: ratio === r }" @click="ratio = r">
              {{ r === 'auto' ? t('studio.gallery.ratioAuto') : r }}
            </button>
          </div>
        </div>

        <div class="grp">
          <div class="st-flabel">{{ t('studio.gallery.gap') }}</div>
          <StSeg v-model="gap" :options="GALLERY_GAPS.map((g) => ({ value: g, label: t(`studio.gallery.gapOpt.${g}`) }))" />
        </div>

        <div class="grp">
          <div class="st-flabel">{{ t('studio.gallery.caption') }}</div>
          <label class="st-field"><input v-model="caption" maxlength="200" :placeholder="t('studio.gallery.captionPh')" /></label>
        </div>

        <Transition name="ge-cur">
          <div v-if="current" class="grp cur">
            <div class="st-flabel">{{ t('studio.gallery.current', { n: sel + 1 }) }}</div>
            <div class="cur-row">
              <img :src="src(current.src)" alt="" />
              <label class="st-field"><input v-model="current.alt" maxlength="120" :placeholder="t('studio.gallery.altPh')" /></label>
            </div>
            <div class="cur-acts">
              <button type="button" class="st-ibtn sm" :disabled="sel <= 0" :title="t('studio.gallery.left')" @click="move(sel, -1)"><Icon :icon="ArrowLeft" :size="15" /></button>
              <button type="button" class="st-ibtn sm" :disabled="sel >= tiles.length - 1" :title="t('studio.gallery.right')" @click="move(sel, 1)"><Icon :icon="ArrowRight" :size="15" /></button>
              <span class="sp" />
              <button type="button" class="st-btn sm g" @click="lib = 'replace'"><Icon :icon="Replace" :size="15" />{{ t('studio.embed.replace') }}</button>
              <button type="button" class="st-btn sm g" @click="remove(sel)"><Icon :icon="Trash2" :size="15" />{{ t('studio.delete') }}</button>
            </div>
          </div>
        </Transition>

        <div class="acts">
          <button type="button" class="st-btn g" @click="emit('close')">{{ t('studio.cancel') }}</button>
          <button type="button" class="st-btn p" :disabled="!tiles.length" @click="save">{{ t('studio.gallery.done') }}</button>
        </div>
      </aside>
    </div>

    <MediaLibraryModal
      :open="!!lib"
      :title="lib === 'replace' ? t('studio.gallery.replaceTitle') : t('studio.gallery.addTitle')"
      :accept="['image']"
      :multiple="lib === 'add'"
      :max="MAX - tiles.length"
      @pick="onPicked"
      @close="lib = ''"
    />
  </StModal>
</template>

<style scoped lang="scss">
:global(.st-modal.gal-ed) { width: min(1240px, calc(100vw - 40px)); max-height: none; padding: 0; overflow: hidden; }

.ge {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  height: min(84vh, 860px);
}

.stage {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  background: var(--well);

  header, footer {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 18px 24px;
  }

  header h3 { margin: 0; font: 700 19px var(--font-serif); }
  header .n { font-size: 12.5px; color: var(--st-ink-3); }
  footer { justify-content: space-between; padding-top: 10px; font-size: 12.5px; color: var(--st-ink-3); }
}

.canvas {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 8px 32px 16px;

  /* 画布按正文栏宽预览（与前台文章同宽比例） */
  .md-gallery { margin: 0 auto; max-width: 720px; }
}

/* 画布里的格子：可拖动、可选中 */
.canvas .gi {
  position: relative;
  cursor: grab;
  user-select: none;
  transition: box-shadow var(--dur-fast), opacity var(--dur-fast);

  &.on { box-shadow: 0 0 0 2.5px var(--ink); }
  &.drag { opacity: 0.45; }

  .no, .rm {
    position: absolute;
    top: 6px;
    height: 22px;
    display: grid;
    place-items: center;
    border-radius: var(--r-pill);
    font-size: 11.5px;
    color: #fff;
    background: rgba(0, 0, 0, 0.5);
  }

  .no { left: 6px; min-width: 22px; padding: 0 6px; }

  .rm {
    right: 6px;
    width: 22px;
    border: 0;
    padding: 0;
    cursor: pointer;
    opacity: 0;
    transition: opacity var(--dur-fast), background var(--dur-fast);

    &:hover { background: color-mix(in oklab, var(--red) 85%, black); }
  }

  &:hover .rm, &.on .rm { opacity: 1; }
}

.ge-move-move { transition: transform var(--dur) var(--ease-out); }

.add-big {
  width: 100%;
  height: 100%;
  min-height: 320px;
  display: grid;
  place-content: center;
  justify-items: center;
  gap: 8px;
  border: 0;
  border-radius: var(--r-lg);
  cursor: pointer;
  color: var(--st-ink-3);
  background: none;
  box-shadow: 0 0 0 1.5px var(--line-2) inset;

  b { font: 600 16px var(--font-serif); color: var(--st-ink); }
  small { font-size: 12.5px; }
  &:hover { color: var(--ink); box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--ink) 55%, transparent) inset; }
}

.panel {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 22px 22px 18px;
  min-height: 0;
  overflow: auto;
  background: var(--paper);
  box-shadow: -1px 0 0 var(--line);

  .grp { display: flex; flex-direction: column; gap: 8px; }
  .chips { display: flex; flex-wrap: wrap; gap: 6px; }
}

.layouts {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;

  button {
    display: grid;
    grid-template-columns: auto 1fr;
    grid-template-rows: auto auto;
    column-gap: 10px;
    align-items: center;
    padding: 10px 12px;
    border: 0;
    border-radius: var(--r-md);
    text-align: left;
    cursor: pointer;
    color: var(--st-ink-2);
    background: var(--well);
    transition: background var(--dur-fast), box-shadow var(--dur-fast), color var(--dur-fast);

    :deep(svg) { grid-row: span 2; }
    b { font-size: 13.5px; font-weight: 600; color: var(--st-ink); }
    small { font-size: 11.5px; color: var(--st-ink-3); }

    &:hover { background: var(--hover); }
    &.on { color: var(--ink); background: var(--tint); box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--ink) 55%, transparent) inset; }
  }
}

.cur {
  padding: 12px;
  border-radius: var(--r-md);
  background: var(--well);

  .cur-row { display: flex; gap: 10px; align-items: center; }
  .cur-row img { width: 56px; height: 56px; object-fit: cover; border-radius: var(--r-sm); flex: none; }
  .cur-row .st-field { flex: 1; min-width: 0; }
  .cur-acts { display: flex; align-items: center; gap: 6px; }
  .sp { flex: 1; }
}

.ge-cur-enter-active, .ge-cur-leave-active { transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-out); }
.ge-cur-enter-from, .ge-cur-leave-to { opacity: 0; transform: translateY(6px); }

.acts {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: auto;
  padding-top: 14px;
  border-top: 1px solid var(--line);
}

@media (max-width: 900px) {
  .ge { grid-template-columns: 1fr; height: auto; max-height: 88vh; overflow: auto; }
}
</style>
