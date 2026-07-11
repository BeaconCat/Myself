<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type MediaItem } from '../../api';
import { useDialogStore } from '../../stores/dialog';

const { t } = useI18n();

const items = ref<MediaItem[]>([]);
const uploading = ref(false);
const fileInput = ref<HTMLInputElement | null>(null);

async function load(): Promise<void> {
  items.value = await adminApi.media();
}

async function onFiles(e: Event): Promise<void> {
  const input = e.target as HTMLInputElement;
  if (!input.files?.length) return;
  uploading.value = true;
  try {
    await adminApi.uploadMedia([...input.files]);
    input.value = '';
    await load();
  } finally {
    uploading.value = false;
  }
}

async function remove(item: MediaItem): Promise<void> {
  const ok = await useDialogStore().confirm({
    title: t('admin.delete'),
    message: t('admin.confirmDeleteMedia', { name: item.name }),
    danger: true,
  });
  if (!ok) return;
  await adminApi.deleteMedia(item.name);
  await load();
}

function formatSize(bytes: number): string {
  if (bytes > 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
  return `${Math.round(bytes / 1024)} KB`;
}

/* ===== 裁切弹窗 ===== */
const cropping = ref<MediaItem | null>(null);
const cropImg = ref<HTMLImageElement | null>(null);
/** 显示坐标系下的框（渲染），提交时换算回原图像素 */
const box = reactive({ x: 0, y: 0, w: 100, h: 100 });
const natural = reactive({ w: 1, h: 1 });
const display = reactive({ w: 1, h: 1 });
const busyCrop = ref(false);

const boxStyle = computed(() => ({
  left: `${box.x}px`,
  top: `${box.y}px`,
  width: `${box.w}px`,
  height: `${box.h}px`,
}));

const originalUrl = ref('');

async function openCrop(item: MediaItem): Promise<void> {
  // 原图接口需要鉴权头，img 标签带不了，取 blob 转对象 URL
  const res = await fetch(`/api/v1/admin/media/${encodeURIComponent(item.name)}/original`, {
    headers: { Authorization: `Bearer ${localStorage.getItem('myself.token') ?? ''}` },
  });
  if (!res.ok) return;
  URL.revokeObjectURL(originalUrl.value);
  originalUrl.value = URL.createObjectURL(await res.blob());
  cropping.value = item;
}

function onCropImgLoad(): void {
  const el = cropImg.value;
  const item = cropping.value;
  if (!el || !item) return;
  natural.w = el.naturalWidth;
  natural.h = el.naturalHeight;
  display.w = el.clientWidth;
  display.h = el.clientHeight;
  const scale = display.w / natural.w;
  if (item.crop) {
    // 调出上次裁切框
    box.x = item.crop.left * scale;
    box.y = item.crop.top * scale;
    box.w = item.crop.width * scale;
    box.h = item.crop.height * scale;
  } else {
    box.x = display.w * 0.1;
    box.y = display.h * 0.1;
    box.w = display.w * 0.8;
    box.h = display.h * 0.8;
  }
}

/* 拖动/缩放框 */
type DragKind = 'move' | 'nw' | 'ne' | 'sw' | 'se';
let drag: { kind: DragKind; startX: number; startY: number; snapshot: typeof box } | null = null;

function startDrag(e: PointerEvent, kind: DragKind): void {
  (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
  drag = { kind, startX: e.clientX, startY: e.clientY, snapshot: { ...box } };
}

function onDrag(e: PointerEvent): void {
  if (!drag) return;
  const dx = e.clientX - drag.startX;
  const dy = e.clientY - drag.startY;
  const s = drag.snapshot;
  const clamp = (v: number, min: number, max: number) => Math.min(max, Math.max(min, v));
  const MIN = 24;

  if (drag.kind === 'move') {
    box.x = clamp(s.x + dx, 0, display.w - s.w);
    box.y = clamp(s.y + dy, 0, display.h - s.h);
    return;
  }
  if (drag.kind.includes('w')) {
    const right = s.x + s.w;
    box.x = clamp(s.x + dx, 0, right - MIN);
    box.w = right - box.x;
  }
  if (drag.kind.includes('e')) {
    box.w = clamp(s.w + dx, MIN, display.w - s.x);
  }
  if (drag.kind.includes('n')) {
    const bottom = s.y + s.h;
    box.y = clamp(s.y + dy, 0, bottom - MIN);
    box.h = bottom - box.y;
  }
  if (drag.kind.includes('s')) {
    box.h = clamp(s.h + dy, MIN, display.h - s.y);
  }
}

function endDrag(): void {
  drag = null;
}

async function confirmCrop(): Promise<void> {
  const item = cropping.value;
  if (!item || busyCrop.value) return;
  busyCrop.value = true;
  try {
    const scale = natural.w / display.w;
    await adminApi.cropMedia(item.name, {
      left: Math.round(box.x * scale),
      top: Math.round(box.y * scale),
      width: Math.round(box.w * scale),
      height: Math.round(box.h * scale),
    });
    cropping.value = null;
    await load();
  } finally {
    busyCrop.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div>
    <header class="head">
      <div>
        <h1 class="page-h">{{ t('admin.menuMedia') }}</h1>
        <p class="hint">{{ t('admin.mediaHint') }}</p>
      </div>
      <button class="btn primary" :disabled="uploading" @click="fileInput?.click()">
        {{ uploading ? t('admin.uploading') : t('admin.upload') }}
      </button>
      <input
        ref="fileInput"
        type="file"
        accept=".png,.jpg,.jpeg,.webp,.gif"
        multiple
        hidden
        @change="onFiles"
      />
    </header>

    <!-- 素材网格 -->
    <div class="grid">
      <div v-for="item in items" :key="item.name" class="cell">
        <!-- size 随裁切变化，作缓存戳保证裁完即时生效 -->
        <img :src="`${item.url}?v=${item.size}`" loading="lazy" alt="" />
        <div class="cell-bar">
          <span class="size">{{ formatSize(item.size) }}</span>
          <span v-if="item.crop" class="badge">{{ t('admin.cropped') }}</span>
        </div>
        <div class="cell-ops">
          <button @click="openCrop(item)">{{ item.crop ? t('admin.recrop') : t('admin.crop') }}</button>
          <button class="danger" @click="remove(item)">{{ t('admin.delete') }}</button>
        </div>
      </div>
    </div>

    <p v-if="!items.length" class="empty">{{ t('admin.mediaEmpty') }}</p>

    <!-- 裁切弹窗：始终基于原图 + 上次裁切框 -->
    <Teleport to="body">
      <div v-if="cropping" class="crop-mask" @click.self="cropping = null">
        <div class="crop-modal">
          <h2>{{ t('admin.cropTitle') }}</h2>
          <div class="crop-stage">
            <img
              ref="cropImg"
              :src="originalUrl"
              alt=""
              draggable="false"
              @load="onCropImgLoad"
            />
            <div
              class="crop-box"
              :style="boxStyle"
              @pointerdown.self="startDrag($event, 'move')"
              @pointermove="onDrag"
              @pointerup="endDrag"
              @pointercancel="endDrag"
            >
              <span
                v-for="corner in (['nw', 'ne', 'sw', 'se'] as const)"
                :key="corner"
                class="handle"
                :class="corner"
                @pointerdown.stop="startDrag($event, corner)"
                @pointermove="onDrag"
                @pointerup="endDrag"
              />
            </div>
          </div>
          <div class="crop-actions">
            <button class="btn ghost" @click="cropping = null">{{ t('admin.cancel') }}</button>
            <button class="btn primary" :disabled="busyCrop" @click="confirmCrop">
              {{ t('admin.applyCrop') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped lang="scss">
.head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 22px;
}

.page-h { font-size: 26px; margin-bottom: 6px; }
.hint { font-size: 13px; color: var(--text-2); }

.btn {
  padding: 10px 24px;
  border: 1px solid transparent;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 700;
  transition: all var(--dur-fast) var(--ease-out);

  &.primary {
    color: #fff;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.4);

    &:hover:not(:disabled) { filter: brightness(1.08); }
    &:disabled { opacity: 0.55; }
  }

  &.ghost {
    background: var(--surface);
    border-color: var(--border);
    color: var(--text);

    &:hover { border-color: var(--primary); color: var(--primary); }
  }
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 14px;
}

.cell {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;

  img {
    width: 100%;
    aspect-ratio: 4 / 3;
    object-fit: cover;
    display: block;
    background: var(--surface-2);
  }
}

.cell-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 10px 0;
  font-size: 11px;
  color: var(--text-2);

  .badge {
    padding: 1px 8px;
    border-radius: 999px;
    background: rgba(var(--primary-rgb), 0.1);
    color: var(--primary);
    font-weight: 600;
  }
}

.cell-ops {
  display: flex;
  gap: 4px;
  padding: 6px;

  button {
    flex: 1;
    border: none;
    background: none;
    font-size: 12px;
    font-weight: 600;
    color: var(--primary);
    padding: 6px 0;
    border-radius: 8px;
    transition: background var(--dur-fast);

    &:hover { background: var(--surface-2); }
    &.danger { color: var(--accent-red); }
  }
}

.empty {
  color: var(--text-2);
  text-align: center;
  padding: 60px 0;
}

/* ===== 裁切弹窗 ===== */
/* 遮罩只轻微衬底：真正的"变暗"只发生在裁切框外的图片区域（box-shadow 蒙版） */
.crop-mask {
  position: fixed;
  inset: 0;
  z-index: 9000;
  background: rgba(8, 8, 12, 0.18);
  backdrop-filter: blur(6px);
  display: grid;
  place-items: center;
  padding: 24px;
}

.crop-modal {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 22px;
  max-width: min(880px, 94vw);

  h2 { font-size: 18px; margin-bottom: 14px; }
}

.crop-stage {
  position: relative;
  user-select: none;
  overflow: hidden; /* 裁掉 box-shadow 蒙版溢出弹窗的部分 */
  width: fit-content;
  margin: 0 auto;

  img {
    display: block;
    max-width: 100%;
    max-height: 62vh;
  }
}

.crop-box {
  position: absolute;
  border: 2px solid var(--primary);
  /* 未选区域变暗（只作用于图片范围，stage overflow 裁掉溢出） */
  box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.55);
  cursor: move;
  touch-action: none;
}

.handle {
  position: absolute;
  width: 14px;
  height: 14px;
  background: var(--primary);
  border: 2px solid #fff;
  border-radius: 50%;
  touch-action: none;

  &.nw { left: -8px; top: -8px; cursor: nwse-resize; }
  &.ne { right: -8px; top: -8px; cursor: nesw-resize; }
  &.sw { left: -8px; bottom: -8px; cursor: nesw-resize; }
  &.se { right: -8px; bottom: -8px; cursor: nwse-resize; }
}

.crop-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}
</style>
