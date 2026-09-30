<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ImagePlus, LoaderCircle, Upload } from 'lucide';
import type { AboutModule } from '../../../stores/config';
import type { GalleryData, GalleryImage } from '../../../about/types';
import Icon from '../../ui/Icon.vue';
import Scene from '../../../about/parts/Scene.vue';
import { adminApi, thumbOf } from '../../../api';
import MediaPicker from '../../../views/admin/studio/MediaPicker.vue';
import { toast } from '../../../views/admin/studio/toast';
import EdCover from './EdCover.vue';
import EdSceneSelect from './EdSceneSelect.vue';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/**
 * 画廊：顶部缩略图条（总览 + 拖拽排序 + 上传 / 素材库添加，照片最多 12 张），
 * 下方逐张一行：配图位（可换图 / 移除，无图时显示所选光影）· 光影 · 标题 · 地点 · 日期。
 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<GalleryData>(() => props.mod);
const { t } = useI18n();

const MAX_PHOTOS = 12;
const MAX_ALL = 16;

const photoCount = computed(() => d.value.images.filter((im) => im.src).length);
const room = computed(() => Math.min(MAX_PHOTOS - photoCount.value, MAX_ALL - d.value.images.length));

const make = (): GalleryImage => ({ src: '', scene: '05', title: '', place: '', date: '' });
const photo = (src: string): GalleryImage => ({ src, title: '', place: '', date: '' });

/* 上传（多选）/ 素材库（单选）追加到末尾 */
const fileInput = ref<HTMLInputElement | null>(null);
const busy = ref(false);
const picking = ref(false);

async function addFiles(files: File[]): Promise<void> {
  const list = files.filter((f) => f.type.startsWith('image/')).slice(0, room.value);
  if (!list.length || busy.value) return;
  busy.value = true;
  try {
    const up = await adminApi.uploadMedia(list);
    d.value.images.push(...up.map((u) => photo(u.url)));
  } catch {
    toast(t('aboutKit.ed.uploadFailed'), { icon: 'x' });
  } finally {
    busy.value = false;
  }
}

function onPick(e: Event): void {
  const input = e.target as HTMLInputElement;
  if (input.files?.length) void addFiles([...input.files]);
  input.value = '';
}

function onLibrary(url: string): void {
  if (room.value > 0) d.value.images.push(photo(url));
}

/** 缩略图条的稳定 key（同一张图可能被加入两次，不能用 src） */
const keys = new WeakMap<object, number>();
let seq = 0;
function keyOf(im: GalleryImage): number {
  if (!keys.has(im)) keys.set(im, ++seq);
  return keys.get(im)!;
}

/* 缩略图条拖拽排序 */
const dragFrom = ref(-1);
const dragOver = ref(-1);
const dropping = ref(false);

function onDrop(i: number): void {
  const from = dragFrom.value;
  dragFrom.value = dragOver.value = -1;
  if (from < 0 || from === i) return;
  const list = d.value.images;
  const [moved] = list.splice(from, 1);
  list.splice(i, 0, moved);
}

/** 把外部文件拖到缩略图条上也能上传 */
function onStripDrop(e: DragEvent): void {
  dropping.value = false;
  if (dragFrom.value < 0 && e.dataTransfer?.files.length) void addFiles([...e.dataTransfer.files]);
}
</script>

<template>
  <div class="ed">
    <div
      class="strip"
      :class="{ dropping }"
      @dragover.prevent="dragFrom < 0 && (dropping = true)"
      @dragleave.self="dropping = false"
      @drop.prevent="onStripDrop"
    >
      <TransitionGroup name="tile">
        <div
          v-for="(im, i) in d.images"
          :key="keyOf(im)"
          class="tile"
          :class="{ from: dragFrom === i, to: dragOver === i && dragFrom !== i }"
          draggable="true"
          :title="im.title || undefined"
          @dragstart="dragFrom = i"
          @dragend="dragFrom = dragOver = -1"
          @dragenter.prevent="dragOver = i"
          @drop.stop.prevent="onDrop(i)"
        >
          <img v-if="im.src" :src="thumbOf(im.src)" alt="" draggable="false" />
          <Scene v-else class="sc" :scene="im.scene" small />
          <i class="no">{{ i + 1 }}</i>
        </div>
      </TransitionGroup>
      <button v-if="room > 0" type="button" class="tile add" :disabled="busy" @click="fileInput?.click()">
        <Icon :class="{ spin: busy }" :icon="busy ? LoaderCircle : Upload" :size="17" />
        <span>{{ busy ? t('aboutKit.ed.uploading') : t('aboutKit.ed.uploadPhotos') }}</span>
      </button>
      <button v-if="room > 0" type="button" class="tile add" :disabled="busy" @click="picking = true">
        <Icon :icon="ImagePlus" :size="17" />
        <span>{{ t('aboutKit.ed.library') }}</span>
      </button>
      <input ref="fileInput" type="file" accept="image/*" multiple hidden @change="onPick" />
    </div>
    <p class="hint">
      <b class="count">{{ t('aboutKit.ed.photoCount', { n: photoCount, max: MAX_PHOTOS }) }}</b>
      {{ t('aboutKit.ed.galleryHint') }}
    </p>

    <EdList v-slot="{ item }" :items="d.images" :make="make" :add-label="t('aboutKit.ed.addScene')" :max="MAX_ALL" compact>
      <div class="g-row">
        <EdCover v-model:src="item.src" class="g-cov" :scene="item.scene" :pick-title="t('aboutKit.ed.pickPhoto')" compact />
        <EdSceneSelect v-model="item.scene" :disabled="!!item.src" />
        <input v-model="item.title" class="a-input nm" type="text" :placeholder="t('aboutKit.ed.title')" :aria-label="t('aboutKit.ed.title')" />
        <input v-model="item.place" class="a-input" type="text" :placeholder="t('aboutKit.ed.place')" :aria-label="t('aboutKit.ed.place')" />
        <input v-model="item.date" class="a-input mono" type="text" placeholder="2026.07.02" :aria-label="t('aboutKit.ed.when')" />
      </div>
    </EdList>

    <MediaPicker :open="picking" :title="t('aboutKit.ed.pickPhoto')" @pick="onLibrary" @close="picking = false" />
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.strip {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(92px, 1fr));
  gap: 8px;
  padding: 10px;
  border-radius: var(--r-md);
  background: color-mix(in oklab, var(--text) 3%, transparent);
  box-shadow: inset 0 0 0 1px var(--line);
  transition: box-shadow var(--dur-fast);

  &.dropping { box-shadow: inset 0 0 0 2px var(--ink); }
}

.tile {
  position: relative;
  aspect-ratio: 4 / 3;
  overflow: hidden;
  border-radius: var(--r-sm);
  background: var(--well);
  box-shadow: 0 0 0 1px var(--line);
  cursor: grab;
  transition: transform var(--dur-fast) var(--ease-spring), box-shadow var(--dur-fast), opacity var(--dur-fast);

  img, .sc { position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover; }
  .sc { opacity: 0.6; }

  &.from { opacity: 0.4; }
  &.to { box-shadow: 0 0 0 2px var(--ink); transform: scale(1.04); }

  .no {
    position: absolute;
    left: 4px;
    top: 4px;
    min-width: 18px;
    padding: 0 4px;
    border-radius: var(--r-xs);
    background: color-mix(in oklab, var(--paper) 82%, transparent);
    font: 500 10.5px/16px var(--font-mono);
    font-style: normal;
    text-align: center;
    color: var(--text-2);
    backdrop-filter: blur(6px);
  }
}

.tile.add {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 4px;
  border: 0;
  box-shadow: inset 0 0 0 1.5px var(--line-2);
  background: none;
  color: var(--text-2);
  font-size: 12px;
  cursor: pointer;

  &:hover:not(:disabled) { color: var(--ink); box-shadow: inset 0 0 0 1.5px color-mix(in oklab, var(--ink) 55%, transparent); }
  &:active:not(:disabled) { transform: scale(0.96); }
  &:disabled { cursor: progress; }
}

.spin { animation: g-spin 0.9s linear infinite; }
@keyframes g-spin { to { transform: rotate(360deg); } }

.count { margin-right: 6px; font-weight: 600; color: var(--text); font-variant-numeric: tabular-nums; }

.g-row {
  display: grid;
  grid-template-columns: 72px minmax(140px, 170px) minmax(0, 1.6fr) minmax(0, 1fr) minmax(110px, 130px);
  align-items: center;
  gap: 8px;
}

.g-cov { width: 72px; aspect-ratio: 3 / 2; }

.nm { font-weight: 600; }
.mono { font-family: var(--font-mono); font-size: 13px; }

.tile-enter-active, .tile-leave-active { transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-spring); }
.tile-enter-from, .tile-leave-to { opacity: 0; transform: scale(0.8); }
.tile-move { transition: transform var(--dur) var(--ease-out); }

@container ed (max-width: 760px) {
  .g-row { grid-template-columns: 72px minmax(0, 1fr) minmax(0, 1fr); }
  .g-row > .nm { grid-column: 2 / -1; }
}

@media (prefers-reduced-motion: reduce) {
  .tile, .tile-enter-active, .tile-leave-active, .tile-move { transition: none; }
  .spin { animation-duration: 2.4s; }
}
</style>
