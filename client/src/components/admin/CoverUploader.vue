<script setup lang="ts">
import { ref } from 'vue';
import { adminApi, thumbOf } from '../../api';
import { useI18n } from 'vue-i18n';
import '../../views/admin/studio/i18n';
import SIcon from '../../views/admin/studio/SIcon.vue';

/**
 * 图片上传宫格：点击 / 拖拽入框，最多 max 张；缩略图可移除、可拖拽排序。
 * square = 方格预览（随想配图），否则 4:3（文章封面，第一张为主封面）。
 */
const props = withDefaults(
  defineProps<{ modelValue: string[]; max?: number; square?: boolean }>(),
  { max: 3, square: false },
);
const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>();
const { t } = useI18n();

const fileInput = ref<HTMLInputElement | null>(null);
const dragging = ref(false);
const busy = ref(false);

async function addFiles(files: File[]): Promise<void> {
  const room = props.max - props.modelValue.length;
  const list = files.filter((f) => f.type.startsWith('image/')).slice(0, room);
  if (!list.length || busy.value) return;
  busy.value = true;
  try {
    const uploaded = await adminApi.uploadMedia(list);
    emit('update:modelValue', [...props.modelValue, ...uploaded.map((u) => u.url)].slice(0, props.max));
  } finally {
    busy.value = false;
  }
}

function onPick(e: Event): void {
  const input = e.target as HTMLInputElement;
  if (input.files?.length) void addFiles([...input.files]);
  input.value = '';
}

function onDrop(e: DragEvent): void {
  dragging.value = false;
  if (e.dataTransfer?.files.length) void addFiles([...e.dataTransfer.files]);
}

function remove(url: string): void {
  emit('update:modelValue', props.modelValue.filter((u) => u !== url));
}

/* 拖拽排序 */
let dragIndex = -1;
function onDropTo(i: number): void {
  if (dragIndex < 0 || dragIndex === i) return;
  const list = [...props.modelValue];
  const [moved] = list.splice(dragIndex, 1);
  list.splice(i, 0, moved);
  dragIndex = -1;
  emit('update:modelValue', list);
}
</script>

<template>
  <div class="covers" :class="{ square }">
    <div
      v-for="(url, i) in modelValue"
      :key="url"
      class="thumb"
      :class="{ main: !square && i === 0 && max > 1 }"
      draggable="true"
      @dragstart="dragIndex = i"
      @dragover.prevent
      @drop.stop="onDropTo(i)"
    >
      <img :src="thumbOf(url)" alt="" draggable="false" />
      <span v-if="!square && i === 0 && max > 1" class="tag">{{ t('studio.editor.cover') }}</span>
      <button type="button" class="remove" :aria-label="t('studio.remove')" @click="remove(url)">
        <SIcon name="x" :size="14" />
      </button>
    </div>
    <button
      v-if="modelValue.length < max"
      type="button"
      class="drop"
      :class="{ dragging, busy }"
      @click="fileInput?.click()"
      @dragover.prevent="dragging = true"
      @dragleave="dragging = false"
      @drop.prevent="onDrop"
    >
      <span v-if="busy" class="spin" />
      <SIcon v-else name="upload" />
      <i>{{ modelValue.length }} / {{ max }}</i>
    </button>
    <input ref="fileInput" type="file" accept="image/*" multiple hidden @change="onPick" />
  </div>
</template>

<style scoped lang="scss">
.covers {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.thumb,
.drop {
  position: relative;
  aspect-ratio: 4 / 3;
  border-radius: var(--r-sm);
  overflow: hidden;

  .square & { aspect-ratio: 1; }
}

.thumb {
  cursor: grab;
  background: var(--surface-2);
  box-shadow: 0 0 0 1px var(--border);
  animation: pop-in var(--dur) var(--ease-spring) both;

  img { width: 100%; height: 100%; object-fit: cover; display: block; }

  &.main { box-shadow: 0 0 0 2px var(--surface), 0 0 0 3.5px var(--text); }

  .tag {
    position: absolute;
    left: 6px;
    bottom: 6px;
    padding: 1px 7px;
    border-radius: var(--r-xs);
    font-size: 11px;
    color: #fff;
    background: rgba(10, 10, 14, 0.5);
    backdrop-filter: blur(6px);
  }

  &:hover .remove { opacity: 1; transform: none; }
}

.remove {
  position: absolute;
  top: 5px;
  right: 5px;
  width: 22px;
  height: 22px;
  border: 0;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: rgba(10, 10, 14, 0.55);
  color: #fff;
  backdrop-filter: blur(6px);
  opacity: 0;
  transform: scale(0.7);
  cursor: pointer;
  transition: all var(--dur-fast) var(--ease-spring);
}

.drop {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  border: 0;
  color: var(--text-2);
  background: var(--surface-2);
  box-shadow: 0 0 0 1.5px var(--border) inset;
  cursor: pointer;
  transition: all var(--dur-fast);

  i { font: 500 11px ui-monospace, Consolas, monospace; font-style: normal; opacity: 0.7; }

  &:hover, &.dragging { color: var(--ink); box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--ink) 55%, transparent) inset; }
  &.dragging { transform: scale(1.03); }
}

.spin {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 2px solid var(--border);
  border-top-color: var(--ink);
  animation: spin 0.8s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pop-in { from { opacity: 0; transform: scale(0.7); } }
</style>
