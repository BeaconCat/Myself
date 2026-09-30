<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Minus, Plus, RotateCcw, X } from 'lucide';
import Icon from './Icon.vue';

/**
 * 头像裁切：正方形取景框 + 圆形遮罩；拖动平移，滚轮 / 双指 / 滑杆缩放（围绕指针或中心），
 * 图片始终铺满取景框。确认后在前端导出 512px 正方形（WebP，不支持时 PNG），交给调用方上传。
 */
const props = defineProps<{ file: File | null; busy?: boolean }>();
const emit = defineEmits<{ done: [blob: Blob]; cancel: [] }>();
const { t } = useI18n();

const OUT = 512;
const MAX_ZOOM = 4;

const stage = ref<HTMLElement | null>(null);
const src = ref('');
const img = ref<HTMLImageElement | null>(null);
const size = ref(300);
const nat = reactive({ w: 1, h: 1 });
const view = reactive({ zoom: 1, x: 0, y: 0 });
const failed = ref(false);

/** 铺满取景框的最小比例 */
const base = computed(() => Math.max(size.value / nat.w, size.value / nat.h));
const scale = computed(() => base.value * view.zoom);

function clamp(): void {
  const w = nat.w * scale.value;
  const h = nat.h * scale.value;
  view.x = Math.min(0, Math.max(size.value - w, view.x));
  view.y = Math.min(0, Math.max(size.value - h, view.y));
}

function center(): void {
  view.zoom = 1;
  view.x = (size.value - nat.w * base.value) / 2;
  view.y = (size.value - nat.h * base.value) / 2;
}

/** 以取景框内一点为中心缩放 */
function zoomAt(next: number, px = size.value / 2, py = size.value / 2): void {
  const z = Math.min(MAX_ZOOM, Math.max(1, next));
  const k = z / view.zoom;
  view.x = px - (px - view.x) * k;
  view.y = py - (py - view.y) * k;
  view.zoom = z;
  clamp();
}

watch(
  () => props.file,
  (f) => {
    if (src.value) URL.revokeObjectURL(src.value);
    src.value = '';
    failed.value = false;
    if (!f) return;
    src.value = URL.createObjectURL(f);
    const el = new Image();
    el.onload = () => {
      nat.w = el.naturalWidth || 1;
      nat.h = el.naturalHeight || 1;
      img.value = el;
      size.value = Math.min(300, Math.round(window.innerWidth * 0.78));
      center();
    };
    el.onerror = () => (failed.value = true);
    el.src = src.value;
  },
  { immediate: true },
);
onBeforeUnmount(() => {
  if (src.value) URL.revokeObjectURL(src.value);
});

/* ---------- 拖动 / 双指缩放 ---------- */
const pointers = new Map<number, { x: number; y: number }>();
let pinch = 0;

function local(e: PointerEvent): { x: number; y: number } {
  const r = stage.value!.getBoundingClientRect();
  return { x: e.clientX - r.left, y: e.clientY - r.top };
}

function onDown(e: PointerEvent): void {
  (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
  pointers.set(e.pointerId, local(e));
  if (pointers.size === 2) {
    const [a, b] = [...pointers.values()];
    pinch = Math.hypot(a.x - b.x, a.y - b.y);
  }
}

function onMove(e: PointerEvent): void {
  const prev = pointers.get(e.pointerId);
  if (!prev) return;
  const p = local(e);
  pointers.set(e.pointerId, p);
  if (pointers.size === 1) {
    view.x += p.x - prev.x;
    view.y += p.y - prev.y;
    clamp();
  } else if (pointers.size === 2) {
    const [a, b] = [...pointers.values()];
    const d = Math.hypot(a.x - b.x, a.y - b.y);
    if (pinch > 0) zoomAt(view.zoom * (d / pinch), (a.x + b.x) / 2, (a.y + b.y) / 2);
    pinch = d;
  }
}

function onUp(e: PointerEvent): void {
  pointers.delete(e.pointerId);
  pinch = 0;
}

function onWheel(e: WheelEvent): void {
  const p = { x: e.offsetX, y: e.offsetY };
  zoomAt(view.zoom * Math.exp(-e.deltaY * 0.0015), p.x, p.y);
}

/* ---------- 导出 ---------- */
async function confirm(): Promise<void> {
  if (!img.value || props.busy) return;
  const canvas = document.createElement('canvas');
  canvas.width = canvas.height = OUT;
  const ctx = canvas.getContext('2d')!;
  ctx.imageSmoothingQuality = 'high';
  const s = scale.value;
  ctx.drawImage(img.value, -view.x / s, -view.y / s, size.value / s, size.value / s, 0, 0, OUT, OUT);
  const blob = await new Promise<Blob | null>((r) => canvas.toBlob(r, 'image/webp', 0.92));
  const out = blob && blob.type === 'image/webp' ? blob : await new Promise<Blob | null>((r) => canvas.toBlob(r, 'image/png'));
  if (out) emit('done', out);
}

function onKey(e: KeyboardEvent): void {
  if (!props.file) return;
  if (e.key === 'Escape') emit('cancel');
  if (e.key === 'Enter') void confirm();
}
watch(
  () => !!props.file,
  (open) => (open ? window.addEventListener('keydown', onKey) : window.removeEventListener('keydown', onKey)),
  { immediate: true },
);
onBeforeUnmount(() => window.removeEventListener('keydown', onKey));

const imgStyle = computed(() => ({
  width: `${nat.w * scale.value}px`,
  height: `${nat.h * scale.value}px`,
  transform: `translate(${view.x}px, ${view.y}px)`,
}));
</script>

<template>
  <Teleport to="body">
    <Transition name="crop">
      <div v-if="file" class="crop-scrim" @mousedown.self="emit('cancel')">
        <div class="crop-card" role="dialog" aria-modal="true" :aria-label="t('account.crop.title')">
          <header>
            <b>{{ t('account.crop.title') }}</b>
            <button type="button" class="ic" :aria-label="t('account.crop.cancel')" @click="emit('cancel')"><Icon :icon="X" :size="18" /></button>
          </header>

          <div
            ref="stage"
            class="stage"
            :style="{ width: `${size}px`, height: `${size}px` }"
            @pointerdown="onDown"
            @pointermove="onMove"
            @pointerup="onUp"
            @pointercancel="onUp"
            @wheel.prevent="onWheel"
          >
            <img v-if="src && !failed" :src="src" alt="" draggable="false" :style="imgStyle" />
            <p v-if="failed" class="bad">{{ t('account.crop.bad') }}</p>
            <span class="mask" aria-hidden="true" />
          </div>

          <div class="zoom">
            <button type="button" class="ic" :aria-label="t('account.crop.zoomOut')" @click="zoomAt(view.zoom / 1.2)"><Icon :icon="Minus" :size="16" /></button>
            <input
              type="range"
              min="1"
              :max="MAX_ZOOM"
              step="0.01"
              :value="view.zoom"
              :aria-label="t('account.crop.zoom')"
              @input="zoomAt(Number(($event.target as HTMLInputElement).value))"
            />
            <button type="button" class="ic" :aria-label="t('account.crop.zoomIn')" @click="zoomAt(view.zoom * 1.2)"><Icon :icon="Plus" :size="16" /></button>
            <button type="button" class="ic" :aria-label="t('account.crop.reset')" :title="t('account.crop.reset')" @click="center"><Icon :icon="RotateCcw" :size="16" /></button>
          </div>
          <p class="hint">{{ t('account.crop.hint') }}</p>

          <footer>
            <button type="button" class="btn ghost" @click="emit('cancel')">{{ t('account.crop.cancel') }}</button>
            <button type="button" class="btn" :disabled="busy || failed || !img" @click="confirm">
              {{ busy ? t('account.busy') : t('account.crop.confirm') }}
            </button>
          </footer>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped lang="scss">
.crop-scrim {
  position: fixed;
  inset: 0;
  z-index: 300;
  display: grid;
  place-items: center;
  padding: 16px;
  background: rgb(0 0 0 / 0.42);
  backdrop-filter: blur(6px);
}

.crop-card {
  width: min(380px, 100%);
  padding: 18px 20px 20px;
  border-radius: var(--r-xl);
  background: var(--elev);
  box-shadow: inset 0 0 0 0.5px var(--line-2), 0 30px 80px -30px rgb(0 0 0 / 0.55);

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;

    b { font: 600 17px var(--font-serif); }
  }
}

.ic {
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  border: 0;
  border-radius: 50%;
  background: none;
  color: var(--text-2);
  cursor: pointer;
  transition: background var(--dur-fast), color var(--dur-fast);

  &:hover { background: var(--fill); color: var(--text); }
}

.stage {
  position: relative;
  margin: 0 auto;
  overflow: hidden;
  border-radius: var(--r-lg);
  background: var(--fill);
  cursor: grab;
  touch-action: none;
  user-select: none;

  &:active { cursor: grabbing; }

  img {
    position: absolute;
    top: 0;
    left: 0;
    max-width: none;
    pointer-events: none;
    transform-origin: 0 0;
  }

  /* 圆形取景：圆外压暗，圆周一道细白线 */
  .mask {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    pointer-events: none;
    box-shadow: 0 0 0 999px rgb(0 0 0 / 0.45), inset 0 0 0 1.5px rgb(255 255 255 / 0.85);
  }

  .bad { position: absolute; inset: 0; display: grid; place-items: center; font-size: 14px; color: var(--text-3); }
}

.zoom {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 14px;

  input { flex: 1; accent-color: var(--solid); }
}

.hint { margin-top: 6px; font-size: 12.5px; color: var(--text-3); text-align: center; }

footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 16px;
}

.btn {
  height: 38px;
  padding: 0 20px;
  border: 0;
  border-radius: var(--r-pill);
  background: var(--solid);
  color: var(--on-solid);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: transform var(--dur-fast) var(--ease-spring), opacity var(--dur-fast);

  &:active:not(:disabled) { transform: scale(0.97); }
  &:disabled { opacity: 0.6; cursor: progress; }
  &.ghost { background: var(--fill); color: var(--text); box-shadow: inset 0 0 0 1px var(--line); }
}

.crop-enter-active, .crop-leave-active { transition: opacity var(--dur) var(--ease-out); }
.crop-enter-active .crop-card { transition: transform var(--dur) var(--ease-spring), opacity var(--dur) var(--ease-out); }
.crop-leave-active .crop-card { transition: transform var(--dur-fast) ease-in, opacity var(--dur-fast) ease-in; }
.crop-enter-from, .crop-leave-to { opacity: 0; }
.crop-enter-from .crop-card, .crop-leave-to .crop-card { opacity: 0; transform: translateY(14px) scale(0.96); }

@media (prefers-reduced-motion: reduce) {
  .crop-enter-active, .crop-leave-active, .crop-enter-active .crop-card, .crop-leave-active .crop-card { transition: none; }
}
</style>
