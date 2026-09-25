<script setup lang="ts">
/**
 * 素材大图查看器：从网格缩略图 FLIP 放大进入；左右滑切换（两端橡皮筋）；
 * 下拉跟手缩小、背景渐隐，过阈值或快速下甩则飞回原位关闭；底部信息 + 复制链接 / 删除。
 */
import { computed, nextTick, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import type { MediaItem } from '../../api';
import MaIcon from './MaIcon.vue';
import { bindDrag, clamp, prefersReducedMotion, rubber } from './gesture';
import { formatSize, relTime } from './format';

const props = defineProps<{
  items: MediaItem[];
  index: number;
  open: boolean;
  /** 取第 i 张在网格中的缩略图元素（FLIP 起止位） */
  source: (i: number) => HTMLElement | null;
}>();
const emit = defineEmits<{ close: []; 'update:index': [i: number]; copy: [item: MediaItem]; remove: [item: MediaItem] }>();
const { t } = useI18n();

const shown = ref(false);
const i = ref(0);
const tx = ref(0);
const trackAnim = ref(false);
const bg = ref(1);
const uiOn = ref(true);
const imgStyle = ref<Record<string, string>>({});
const rootEl = ref<HTMLElement | null>(null);
const imgs = ref<HTMLImageElement[]>([]);

const cur = computed(() => props.items[i.value]);
const W = (): number => window.innerWidth;

function flipFrom(src: HTMLElement | null, img: HTMLElement | undefined): string | null {
  if (!src || !img) return null;
  const r = src.getBoundingClientRect();
  const tr = img.getBoundingClientRect();
  if (!r.width || !tr.width) return null;
  const s = Math.max(r.width / tr.width, r.height / tr.height);
  return `translate(${r.left + r.width / 2 - (tr.left + tr.width / 2)}px, ${r.top + r.height / 2 - (tr.top + tr.height / 2)}px) scale(${s})`;
}

watch(
  () => props.open,
  async (v) => {
    if (!v) return;
    i.value = props.index;
    tx.value = -i.value * W();
    trackAnim.value = false;
    bg.value = 1;
    uiOn.value = true;
    shown.value = true;
    await nextTick();
    const img = imgs.value[i.value];
    const run = (): void => {
      const f = prefersReducedMotion() ? null : flipFrom(props.source(i.value), img);
      if (!f) {
        imgStyle.value = {};
        return;
      }
      imgStyle.value = { transform: f, transition: 'none', borderRadius: 'var(--r-md)' };
      requestAnimationFrame(() =>
        requestAnimationFrame(() => {
          imgStyle.value = { transform: 'none', transition: 'transform .45s var(--ease-sheet), border-radius .45s' };
        }),
      );
    };
    if (img?.complete) run();
    else img?.addEventListener('load', run, { once: true });
    bindGesture();
  },
);

function go(k: number): void {
  i.value = clamp(k, 0, props.items.length - 1);
  trackAnim.value = true;
  tx.value = -i.value * W();
  emit('update:index', i.value);
}

function close(): void {
  const img = imgs.value[i.value];
  const src = props.source(i.value);
  const sr = src?.getBoundingClientRect();
  const visible = sr && sr.width && sr.top > 40 && sr.bottom < window.innerHeight - 60;
  bg.value = 0;
  uiOn.value = false;
  if (visible && !prefersReducedMotion()) {
    const prev = imgStyle.value.transform;
    imgStyle.value = { transform: 'none', transition: 'none' };
    void img?.offsetWidth;
    const f = flipFrom(src ?? null, img);
    imgStyle.value = { transform: prev ?? 'none', transition: 'none' };
    requestAnimationFrame(() => {
      imgStyle.value = { transform: f ?? 'scale(.85)', transition: 'transform .42s var(--ease-sheet), border-radius .42s', borderRadius: 'var(--r-md)' };
    });
  } else {
    imgStyle.value = { transform: 'scale(.85)', opacity: '0', transition: 'transform .35s, opacity .3s' };
  }
  window.setTimeout(() => {
    shown.value = false;
    imgStyle.value = {};
    emit('close');
  }, 420);
}

let unbind: (() => void) | null = null;
function bindGesture(): void {
  unbind?.();
  if (!rootEl.value) return;
  unbind = bindDrag(rootEl.value, {
    down: (e) => ((e.target as HTMLElement).closest('.vw-top, .vw-bar') ? null : { ax: '' as '' | 'x' | 'y' }),
    begin: () => {
      trackAnim.value = false;
    },
    move: (c, dx, dy) => {
      if (!c.ax) c.ax = Math.abs(dx) > Math.abs(dy) ? 'x' : 'y';
      const n = props.items.length;
      if (c.ax === 'x') {
        const edge = (i.value === 0 && dx > 0) || (i.value === n - 1 && dx < 0);
        tx.value = -i.value * W() + (edge ? Math.sign(dx) * rubber(Math.abs(dx), 90) : dx);
      } else {
        const yy = dy > 0 ? dy : -rubber(-dy, 40);
        const sc = 1 - clamp(yy, 0, 400) / 1400;
        imgStyle.value = { transform: `translate(${dx * 0.6}px, ${yy}px) scale(${sc})`, transition: 'none' };
        bg.value = clamp(1 - yy / 420, 0.15, 1);
        uiOn.value = yy < 20;
      }
    },
    end: (c, vx, vy, dx, dy) => {
      if (c.ax === 'x') {
        let k = i.value;
        if (vx < -0.3 || dx < -W() / 3) k += 1;
        else if (vx > 0.3 || dx > W() / 3) k -= 1;
        go(k);
      } else if (dy > 110 || vy > 0.6) {
        close();
      } else {
        imgStyle.value = { transform: 'none', transition: 'transform .45s var(--ease-spring)' };
        bg.value = 1;
        uiOn.value = true;
      }
    },
  });
}

function toggleUi(): void {
  uiOn.value = !uiOn.value;
}

defineExpose({ close });
</script>

<template>
  <Teleport to="#ma-overlay" defer>
    <div v-if="shown" ref="rootEl" class="vw" :class="{ ui: uiOn }" @click.self="toggleUi">
      <div class="vw-bg" :style="{ opacity: bg }" />
      <div class="vw-track" :class="{ anim: trackAnim }" :style="{ transform: `translateX(${tx}px)` }">
        <div v-for="(m, k) in items" :key="m.name" class="vw-slide" @click="toggleUi">
          <img
            :ref="(el) => { if (el) imgs[k] = el as HTMLImageElement; }"
            :src="Math.abs(k - i) <= 1 ? m.url : m.thumb"
            :style="k === i ? imgStyle : undefined"
            alt=""
            draggable="false"
          />
        </div>
      </div>
      <header class="vw-top">
        <button class="vbtn tap" :aria-label="t('mobileAdmin.common.close')" @click="close"><MaIcon name="close" :size="20" /></button>
        <span class="cnt">{{ i + 1 }} / {{ items.length }}</span>
        <button class="vbtn tap" :aria-label="t('mobileAdmin.media.copy')" @click="cur && emit('copy', cur)"><MaIcon name="copy" :size="19" /></button>
      </header>
      <footer v-if="cur" class="vw-bar">
        <div class="info">
          <b>{{ cur.name }}</b>
          <small>{{ formatSize(cur.size) }} · {{ relTime(cur.createdAt, t) }}<template v-if="cur.hasOriginal"> · {{ t('mobileAdmin.media.hasOriginal') }}</template></small>
        </div>
        <button class="vbtn danger tap" :aria-label="t('mobileAdmin.common.delete')" @click="emit('remove', cur)"><MaIcon name="trash" :size="19" /></button>
      </footer>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.vw {
  position: absolute;
  inset: 0;
  z-index: 80;
  pointer-events: auto;
  touch-action: none;
  color: #fff;
}

.vw-bg {
  position: absolute;
  inset: 0;
  background: #000;
  transition: opacity 0.35s var(--ease-out);
}

.vw-track {
  position: absolute;
  inset: 0;
  display: flex;

  &.anim { transition: transform 0.45s var(--ease-sheet); }
}

.vw-slide {
  flex: none;
  width: 100vw;
  height: 100%;
  display: grid;
  place-items: center;

  img {
    max-width: 100%;
    max-height: 100%;
    object-fit: contain;
    will-change: transform;
    user-select: none;
  }
}

.vw-top,
.vw-bar {
  position: absolute;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 16px;
  opacity: 0;
  transition: opacity 0.3s, transform 0.4s var(--ease-sheet);

  .ui & { opacity: 1; transform: none; }
}

.vw-top {
  top: 0;
  padding-top: calc(var(--safe-t) + 6px);
  justify-content: space-between;
  transform: translateY(-10px);

  .cnt {
    font-family: var(--font-mono);
    font-size: 13px;
    opacity: 0.8;
  }
}

.vw-bar {
  bottom: 0;
  padding-bottom: calc(var(--safe-b) + 12px);
  padding-top: 40px;
  background: linear-gradient(transparent, rgba(0, 0, 0, 0.6));
  transform: translateY(10px);

  .info {
    flex: 1;
    min-width: 0;
  }

  b {
    display: block;
    font-size: 13px;
    font-family: var(--font-mono);
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  small {
    font-size: 12px;
    opacity: 0.65;
  }
}

.vbtn {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: #fff;
  background: rgba(255, 255, 255, 0.14);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);

  &.danger { color: #ff6b6b; }
}
</style>
