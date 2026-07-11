<script setup lang="ts">
import { ref } from 'vue';
import { adminApi } from '../../api';

/** 图片上传：点击 / 拖拽入框，最多 max 张，缩略图可移除、可拖拽排序；square = 方格拼图预览 */
const props = withDefaults(
  defineProps<{ modelValue: string[]; max?: number; square?: boolean }>(),
  { max: 3, square: false },
);
const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>();

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

function onDragStart(i: number): void {
  dragIndex = i;
}

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
      draggable="true"
      @dragstart="onDragStart(i)"
      @dragover.prevent
      @drop="onDropTo(i)"
    >
      <img :src="url" alt="" />
      <button type="button" class="remove" aria-label="移除" @click="remove(url)">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round">
          <path d="M6 6l12 12M18 6L6 18" />
        </svg>
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
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
        <path d="M12 5v14M5 12h14" />
      </svg>
      <span>{{ busy ? '上传中…' : '点击或拖拽上传' }}</span>
      <i>{{ modelValue.length }}/{{ max }}</i>
    </button>

    <input ref="fileInput" type="file" accept="image/*" multiple hidden @change="onPick" />
  </div>
</template>

<style scoped lang="scss">
.covers {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.thumb {
  position: relative;
  width: 168px;
  aspect-ratio: 16 / 9;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--border);
  cursor: grab;
  transition: transform var(--dur-fast) var(--ease-out), border-color var(--dur-fast);

  &:active { cursor: grabbing; }
  &:hover { border-color: rgba(var(--primary-rgb), 0.5); }

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
}

.remove {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  display: grid;
  place-items: center;
  opacity: 0;
  transition: opacity var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  svg { width: 12px; height: 12px; }

  &:hover { transform: scale(1.15); background: rgba(255, 0, 50, 0.8); }
}

.thumb:hover .remove { opacity: 1; }

.drop {
  width: 168px;
  aspect-ratio: 16 / 9;
  border: 1.5px dashed var(--border);
  border-radius: 10px;
  background: var(--bg);
  color: var(--text-2);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  font-size: 12px;
  transition: all var(--dur-fast) var(--ease-out);

  svg { width: 20px; height: 20px; }
  i { font-style: normal; font-size: 10.5px; opacity: 0.7; }

  &:hover, &.dragging {
    border-color: var(--primary);
    color: var(--primary);
    background: rgba(var(--primary-rgb), 0.05);
    transform: scale(1.02);
  }

  &.busy { opacity: 0.6; pointer-events: none; }
}

/* 方格拼图模式：随想配图，三列宫格预览可拖拽换位 */
.covers.square {
  display: grid;
  grid-template-columns: repeat(3, 1fr);

  .thumb, .drop {
    width: 100%;
    aspect-ratio: 1;
  }
}
</style>
