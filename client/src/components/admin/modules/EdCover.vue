<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ImagePlus, LoaderCircle, Upload, X } from 'lucide';
import Icon from '../../ui/Icon.vue';
import Scene from '../../../about/parts/Scene.vue';
import { adminApi, thumbOf } from '../../../api';
import MediaPicker from '../../../views/admin/studio/MediaPicker.vue';
import { toast } from '../../../views/admin/studio/toast';

/**
 * 单张配图位（作品封面 / 画廊照片）：有图显示缩略图，无图显示所选光影（scene）作底并提示上传。
 * 悬停出现 上传 / 素材库 / 移除；也可以直接把图片拖进来。尺寸由外层决定（宽度 + aspect-ratio）。
 * compact = 行内小缩略图：操作按钮悬停时居中浮现，不显示「光影」角标。
 */
const props = defineProps<{ src?: string; scene?: string; pickTitle?: string; compact?: boolean }>();
const emit = defineEmits<{ 'update:src': [value: string] }>();
const { t } = useI18n();

const fileInput = ref<HTMLInputElement | null>(null);
const busy = ref(false);
const over = ref(false);
const picking = ref(false);

async function upload(file: File | undefined): Promise<void> {
  if (!file || !file.type.startsWith('image/') || busy.value) return;
  busy.value = true;
  try {
    const [up] = await adminApi.uploadMedia([file]);
    if (up) emit('update:src', up.url);
  } catch {
    toast(t('aboutKit.ed.uploadFailed'), { icon: 'x' });
  } finally {
    busy.value = false;
  }
}

function onPick(e: Event): void {
  const input = e.target as HTMLInputElement;
  void upload(input.files?.[0]);
  input.value = '';
}

function onDrop(e: DragEvent): void {
  over.value = false;
  void upload(e.dataTransfer?.files[0]);
}
</script>

<template>
  <div
    class="cov"
    :class="{ has: !!src, over, busy, sm: compact }"
    @dragover.prevent="over = true"
    @dragleave="over = false"
    @drop.prevent="onDrop"
  >
    <img v-if="src" class="cov-img" :src="thumbOf(src)" alt="" draggable="false" />
    <Scene v-else class="cov-img dim" :scene="scene" small />
    <span v-if="!src && !busy && !compact" class="cov-tip"><Icon :icon="ImagePlus" :size="16" />{{ t('aboutKit.ed.sceneShort') }}</span>
    <span v-if="busy" class="cov-busy"><Icon class="spin" :icon="LoaderCircle" :size="18" /></span>

    <div class="cov-ops">
      <button type="button" :title="t('aboutKit.ed.upload')" :aria-label="t('aboutKit.ed.upload')" :disabled="busy" @click="fileInput?.click()">
        <Icon :icon="Upload" :size="15" />
      </button>
      <button type="button" :title="t('aboutKit.ed.library')" :aria-label="t('aboutKit.ed.library')" :disabled="busy" @click="picking = true">
        <Icon :icon="ImagePlus" :size="15" />
      </button>
      <button v-if="src" type="button" class="del" :title="t('aboutKit.ed.removeCover')" :aria-label="t('aboutKit.ed.removeCover')" @click="emit('update:src', '')">
        <Icon :icon="X" :size="15" />
      </button>
    </div>
    <input ref="fileInput" type="file" accept="image/*" hidden @change="onPick" />
    <MediaPicker
      :open="picking"
      :title="props.pickTitle || t('aboutKit.ed.pickCover')"
      :current="src"
      @pick="emit('update:src', $event)"
      @close="picking = false"
    />
  </div>
</template>

<style scoped lang="scss">
.cov {
  position: relative;
  overflow: hidden;
  border-radius: var(--r-sm);
  background: var(--well);
  box-shadow: 0 0 0 1px var(--line) inset;
  isolation: isolate;
  transition: box-shadow var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  &.over { box-shadow: 0 0 0 2px var(--ink) inset; transform: scale(1.02); }
}

.cov-img {
  position: absolute;
  inset: 0;
  z-index: -1;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: opacity var(--dur), filter var(--dur);

  &.dim { opacity: 0.55; filter: saturate(0.7); }
}

.cov-tip {
  position: absolute;
  left: 6px;
  bottom: 6px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 7px 2px 5px;
  border-radius: var(--r-xs);
  background: color-mix(in oklab, var(--paper) 82%, transparent);
  font-size: 11.5px;
  color: var(--text-2);
  backdrop-filter: blur(6px);
}

.cov-busy {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  background: color-mix(in oklab, var(--paper) 60%, transparent);
  color: var(--ink);
}

.spin { animation: cov-spin 0.9s linear infinite; }
@keyframes cov-spin { to { transform: rotate(360deg); } }

.cov-ops {
  position: absolute;
  top: 6px;
  right: 6px;
  display: flex;
  gap: 4px;
  opacity: 0;
  transform: translateY(-3px);
  transition: opacity var(--dur-fast), transform var(--dur) var(--ease-out);

  .cov:hover &, .cov:focus-within &, .cov:not(.has) & { opacity: 1; transform: none; }

  button {
    display: grid;
    place-items: center;
    width: 26px;
    height: 26px;
    padding: 0;
    border: 0;
    border-radius: var(--r-xs);
    background: color-mix(in oklab, var(--paper) 86%, transparent);
    color: var(--text);
    box-shadow: 0 0 0 1px var(--line);
    backdrop-filter: blur(6px);
    cursor: pointer;
    transition: background-color var(--dur-fast), color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

    &:hover:not(:disabled) { background: var(--paper); }
    &:active:not(:disabled) { transform: scale(0.92); }
    &.del:hover { color: var(--accent-red); }
    &:disabled { opacity: 0.5; cursor: default; }
  }
}

/* 行内小图：按钮居中叠在图上，仅悬停 / 聚焦时出现 */
.cov.sm .cov-ops {
  inset: 0;
  align-items: center;
  justify-content: center;
  gap: 3px;
  background: color-mix(in oklab, var(--paper) 35%, transparent);
  transform: none;

  button { width: 22px; height: 22px; }
}
.cov.sm:not(:hover):not(:focus-within) .cov-ops { opacity: 0; }

@media (prefers-reduced-motion: reduce) {
  .spin { animation-duration: 2.4s; }
  .cov, .cov-ops { transition: none; }
}
</style>
