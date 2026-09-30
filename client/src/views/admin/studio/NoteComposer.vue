<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, thumbOf, type Note } from '../../../api';
import { useConfigStore } from '../../../stores/config';
import SIcon from './SIcon.vue';
import { toast } from './toast';

/**
 * 随想输入框：聚焦展开（心情 / 九图拖放上传 / 置顶），Ctrl+Enter 发布。
 * 传入 note 为编辑模式；alwaysOpen 用于独立写作页。
 */
const props = withDefaults(
  defineProps<{ placeholder?: string; note?: Note | null; alwaysOpen?: boolean; hint?: boolean }>(),
  { placeholder: '', note: null, alwaysOpen: false, hint: false },
);
const emit = defineEmits<{ published: [id: number]; saved: [] }>();

const { t } = useI18n();
const config = useConfigStore();

const MOODS = [
  { key: 'calm', c: 'var(--blue)' },
  { key: 'joy', c: 'var(--yellow)' },
  { key: 'muse', c: '#8a8378' },
  { key: 'tired', c: '#b9b3a8' },
  { key: 'grateful', c: 'var(--red)' },
];
const moodNames = computed(() => MOODS.map((m) => ({ ...m, name: t(`studio.mood.${m.key}`) })));

interface Img {
  id: number;
  url: string;
  preview: string;
  busy: boolean;
  bye?: boolean;
}

const MAX = 9;
const el = ref<HTMLElement | null>(null);
const ta = ref<HTMLTextAreaElement | null>(null);
const fileInput = ref<HTMLInputElement | null>(null);
const text = ref('');
const mood = ref('');
const pinned = ref(false);
const imgs = ref<Img[]>([]);
const focused = ref(false);
const dragover = ref(false);
const shaking = ref(false);
const sending = ref(false);
const sent = ref(false);
let seq = 0;

const open = computed(
  () => props.alwaysOpen || focused.value || !!text.value || imgs.value.length > 0 || dragover.value,
);
const uploading = computed(() => imgs.value.some((i) => i.busy));
const count = computed(() => text.value.trim().length);
const gridN = computed(() => Math.min(MAX, imgs.value.length + (imgs.value.length && imgs.value.length < MAX ? 1 : 0)));
const customMood = computed(() => (mood.value && !moodNames.value.some((m) => m.name === mood.value) ? mood.value : ''));

function fit(): void {
  const el2 = ta.value;
  if (!el2) return;
  el2.style.height = 'auto';
  el2.style.height = `${Math.max(open.value ? 84 : 32, el2.scrollHeight)}px`;
}
watch([text, open], () => void nextTick(fit));

function reset(): void {
  text.value = '';
  mood.value = '';
  pinned.value = false;
  imgs.value.forEach((i) => i.preview.startsWith('blob:') && URL.revokeObjectURL(i.preview));
  imgs.value = [];
}

function load(n: Note | null): void {
  reset();
  if (!n) return;
  text.value = n.contentMd;
  mood.value = n.mood;
  pinned.value = n.pinned;
  imgs.value = n.images.map((u) => ({ id: ++seq, url: u, preview: thumbOf(u), busy: false }));
}
watch(() => props.note, load, { immediate: true });

function pickMood(name: string): void {
  mood.value = mood.value === name ? '' : name;
}

async function addFiles(files: File[]): Promise<void> {
  const room = MAX - imgs.value.length;
  const list = files.filter((f) => f.type.startsWith('image/')).slice(0, room);
  if (!list.length) {
    if (room <= 0) toast(t('studio.composer.full'), { icon: 'image' });
    return;
  }
  const added = list.map((f) => ({ id: ++seq, url: '', preview: URL.createObjectURL(f), busy: true }));
  imgs.value.push(...added);
  try {
    const uploaded = await adminApi.uploadMedia(list);
    added.forEach((a, i) => {
      const target = imgs.value.find((x) => x.id === a.id);
      if (target && uploaded[i]) {
        target.url = uploaded[i].url;
        target.busy = false;
      }
    });
  } catch {
    imgs.value = imgs.value.filter((x) => !added.some((a) => a.id === x.id));
    toast(t('studio.composer.uploadFailed'), { icon: 'x' });
  }
}

function onPick(e: Event): void {
  const input = e.target as HTMLInputElement;
  if (input.files?.length) void addFiles([...input.files]);
  input.value = '';
}

function removeImg(img: Img): void {
  img.bye = true;
  window.setTimeout(() => {
    imgs.value = imgs.value.filter((x) => x.id !== img.id);
    if (img.preview.startsWith('blob:')) URL.revokeObjectURL(img.preview);
  }, 200);
}

/* 拖拽排序 */
let dragFrom = -1;
function onTileDrop(i: number): void {
  if (dragFrom < 0 || dragFrom === i) return;
  const list = [...imgs.value];
  const [m] = list.splice(dragFrom, 1);
  list.splice(i, 0, m);
  imgs.value = list;
  dragFrom = -1;
}

function onDragOver(e: DragEvent): void {
  if (!e.dataTransfer?.types.includes('Files')) return;
  e.preventDefault();
  dragover.value = true;
}
function onDragLeave(e: DragEvent): void {
  if (!el.value?.contains(e.relatedTarget as Node)) dragover.value = false;
}
function onDrop(e: DragEvent): void {
  if (!e.dataTransfer?.files.length) return;
  e.preventDefault();
  dragover.value = false;
  void addFiles([...e.dataTransfer.files]);
}

function shake(): void {
  shaking.value = false;
  void nextTick(() => {
    shaking.value = true;
    window.setTimeout(() => (shaking.value = false), 420);
  });
  ta.value?.focus();
}

async function submit(): Promise<void> {
  if (sending.value) return;
  if (!text.value.trim()) return shake();
  if (uploading.value) {
    toast(t('studio.composer.waitUpload'), { icon: 'upload' });
    return;
  }
  sending.value = true;
  const body = {
    contentMd: text.value.trim(),
    mood: mood.value,
    images: imgs.value.map((i) => i.url).filter(Boolean),
    pinned: pinned.value,
  };
  try {
    if (props.note) {
      await adminApi.updateNote(props.note.id, body);
      toast(t('studio.composer.saved'));
      emit('saved');
    } else {
      const { id } = await adminApi.createNote(body);
      sent.value = true;
      window.setTimeout(() => {
        sent.value = false;
        reset();
        ta.value?.blur();
        focused.value = false;
      }, 260);
      emit('published', id);
    }
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  } finally {
    sending.value = false;
  }
}

function onKey(e: KeyboardEvent): void {
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
    e.preventDefault();
    void submit();
  }
  if (e.key === 'Escape') {
    ta.value?.blur();
    focused.value = false;
  }
}

/* 点外部收起（空内容时） */
function outside(e: MouseEvent): void {
  if (!el.value?.contains(e.target as Node) && !(e.target as HTMLElement).closest?.('.st-menu, .modal-mask')) {
    focused.value = false;
  }
}

function focus(): void {
  focused.value = true;
  void nextTick(() => ta.value?.focus());
}

onMounted(() => {
  document.addEventListener('mousedown', outside);
  fit();
});
onBeforeUnmount(() => document.removeEventListener('mousedown', outside));
defineExpose({ focus });

const avatar = computed(() => config.cfg.about?.avatar || config.cfg.site.logo || '/favicon-64.png');
</script>

<template>
  <div
    ref="el"
    class="cmp"
    :class="{ open, dragover, shake: shaking, sent, 'has-img': imgs.length > 0 }"
    @dragover="onDragOver"
    @dragleave="onDragLeave"
    @drop="onDrop"
  >
    <div class="cmp-top">
      <span class="avatar"><img :src="avatar" alt="" /></span>
      <textarea
        ref="ta"
        v-model="text"
        rows="1"
        :placeholder="placeholder || t('studio.composer.placeholder')"
        @focus="focused = true"
        @keydown="onKey"
      />
      <span v-if="hint" class="hint"><kbd class="st-kbd">N</kbd></span>
    </div>

    <div class="cmp-more">
      <div>
        <div class="cmp-inner">
          <div class="moods">
            <button
              v-for="m in moodNames"
              :key="m.key"
              type="button"
              class="mood"
              :class="{ on: mood === m.name }"
              :style="{ '--c': m.c }"
              @click="pickMood(m.name)"
            ><i class="st-dot" />{{ m.name }}</button>
            <button v-if="customMood" type="button" class="mood on" style="--c: var(--ink)" @click="mood = ''">
              <i class="st-dot" />{{ customMood }}
            </button>
          </div>

          <div v-if="imgs.length" class="imgs" :data-n="gridN">
            <div
              v-for="(img, i) in imgs"
              :key="img.id"
              class="t"
              :class="{ bye: img.bye, busy: img.busy }"
              draggable="true"
              @dragstart="dragFrom = i"
              @dragover.prevent
              @drop.stop="onTileDrop(i)"
            >
              <img :src="img.preview" alt="" draggable="false" />
              <span v-if="img.busy" class="spin" />
              <button type="button" class="rm" :aria-label="t('studio.remove')" @click="removeImg(img)">
                <SIcon name="x" :size="14" />
              </button>
            </div>
            <button v-if="imgs.length < MAX" type="button" class="add" @click="fileInput?.click()">
              <SIcon name="plus" />
            </button>
          </div>
          <button v-else type="button" class="drop" @click="fileInput?.click()">
            <span class="di"><SIcon name="image" /></span>
            <span>
              <b>{{ t('studio.composer.dropTitle') }}</b>{{ t('studio.composer.dropOr') }}<br />
              <small>{{ t('studio.composer.dropHint') }}</small>
            </span>
          </button>
        </div>

        <div class="cmp-bar">
          <button type="button" class="st-ibtn" :title="t('studio.composer.addImage')" @click="fileInput?.click()">
            <SIcon name="image" />
          </button>
          <button type="button" class="st-ibtn" :class="{ on: pinned }" :title="t('studio.pin')" @click="pinned = !pinned">
            <SIcon name="pin" />
          </button>
          <span class="sp" />
          <span class="cnt mono">{{ t('studio.composer.count', { n: count }) }}</span>
          <span class="keys"><kbd class="st-kbd">Ctrl</kbd><kbd class="st-kbd">Enter</kbd></span>
          <button type="button" class="st-btn p sm" :disabled="sending" @click="submit">
            {{ note ? t('studio.composer.save') : t('studio.composer.publish') }}
          </button>
        </div>
      </div>
    </div>
    <div v-if="dragover" class="drop-veil">{{ t('studio.composer.dropVeil') }}</div>
    <input ref="fileInput" type="file" accept="image/*" multiple hidden @change="onPick" />
  </div>
</template>

<style scoped lang="scss">
.cmp {
  position: relative;
  border-radius: var(--r-lg);
  background: var(--paper);
  box-shadow: 0 0 0 1px var(--line-2), 0 1px 2px var(--line);
  transition: box-shadow var(--dur) var(--ease-out), transform var(--dur) var(--ease-spring);

  &:hover:not(.open) {
    box-shadow: 0 0 0 1px var(--line-3), 0 10px 30px -18px color-mix(in oklab, var(--st-shade) 44%, transparent);
  }

  &.open {
    box-shadow:
      0 0 0 1px color-mix(in oklab, var(--ink) 55%, transparent),
      0 0 0 3px color-mix(in oklab, var(--ink) 16%, transparent),
      0 30px 60px -30px color-mix(in oklab, var(--st-shade) 48%, transparent);
  }

  &.shake { animation: shake 0.4s var(--ease-out); }
  &.sent .cmp-top, &.sent .cmp-more { animation: sent 0.5s var(--ease-out); }
}

:root[data-mode='dark'] .cmp { background: var(--well); }

@keyframes shake { 20%, 60% { transform: translateX(-5px); } 40%, 80% { transform: translateX(5px); } }
@keyframes sent { 40% { opacity: 0; transform: translateY(-14px) scale(0.98); } 41% { transform: translateY(8px); } }

.cmp-top {
  display: flex;
  gap: 14px;
  padding: 16px 18px 14px 16px;
  align-items: flex-start;

  textarea {
    flex: 1;
    border: 0;
    outline: 0;
    background: none;
    resize: none;
    font: 400 18px/1.75 var(--font-serif);
    color: var(--st-ink);
    min-height: 32px;
    height: 32px;
    padding: 1px 0;
    transition: height var(--dur) var(--ease-out);

    &::placeholder { color: var(--st-ink-4); }
  }
}

.avatar {
  width: 32px;
  height: 32px;
  margin-top: 3px;
  border-radius: 50%;
  flex: none;
  overflow: hidden;
  box-shadow: 0 0 0 2px var(--paper), 0 0 0 3px var(--line-2);

  img { width: 100%; height: 100%; object-fit: cover; display: block; }
}

.hint {
  margin-top: 5px;
  transition: opacity var(--dur-fast);

  .open & { opacity: 0; pointer-events: none; }
}

.cmp-more {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows var(--dur) var(--ease-out);

  > div { overflow: hidden; min-height: 0; }

  .open & { grid-template-rows: 1fr; }
}

.cmp-inner {
  padding: 0 18px 14px 62px;
  opacity: 0;
  transform: translateY(-6px);
  transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-out);

  .open & { opacity: 1; transform: none; transition-delay: 0.08s; }
}

.moods {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin: 4px 0 14px;
}

.mood {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  height: 30px;
  padding: 0 12px 0 10px;
  border-radius: var(--r-pill);
  font-size: 13px;
  color: var(--st-ink-2);
  box-shadow: 0 0 0 1px var(--line-2) inset;
  transition: all var(--dur) var(--ease-spring);

  &:hover { background: var(--hover); }
  &:active { transform: scale(0.95); }

  .st-dot { transition: transform var(--dur) var(--ease-spring), box-shadow var(--dur); }

  /* 选中：抬升 + 轻染；心情圆点保留自身颜色作为唯一信号 */
  &.on {
    background: var(--lift);
    box-shadow: var(--lift-shadow);
    color: var(--lift-fg);
    font-weight: 500;

    .st-dot { transform: scale(1.3); }
  }
}

.imgs {
  display: grid;
  gap: 6px;
  margin-bottom: 14px;
  max-width: 420px;
  grid-template-columns: repeat(3, 1fr);

  &[data-n='1'] { grid-template-columns: 1fr; max-width: 300px; }
  &[data-n='2'], &[data-n='4'] { grid-template-columns: repeat(2, 1fr); max-width: 300px; }

  .t {
    position: relative;
    aspect-ratio: 1;
    border-radius: var(--r-sm);
    overflow: hidden;
    cursor: grab;
    background: var(--well-2);
    animation: tile-in var(--dur) var(--ease-spring) both;

    img { width: 100%; height: 100%; object-fit: cover; display: block; transition: filter var(--dur); }

    &.busy img { filter: saturate(0.4) brightness(0.8); }
    &.bye { animation: tile-out 0.22s ease-in forwards; }
    &:hover .rm { opacity: 1; transform: none; }
  }

  .rm {
    position: absolute;
    top: 5px;
    right: 5px;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: rgba(10, 10, 14, 0.55);
    color: #fff;
    backdrop-filter: blur(6px);
    opacity: 0;
    transform: scale(0.7);
    transition: all var(--dur-fast) var(--ease-spring);
  }

  .spin {
    position: absolute;
    left: 50%;
    top: 50%;
    width: 22px;
    height: 22px;
    margin: -11px 0 0 -11px;
    border-radius: 50%;
    border: 2px solid rgba(255, 255, 255, 0.35);
    border-top-color: #fff;
    animation: spin 0.8s linear infinite;
  }

  .add {
    aspect-ratio: 1;
    border-radius: var(--r-sm);
    display: grid;
    place-items: center;
    color: var(--st-ink-3);
    box-shadow: 0 0 0 1.5px var(--line-2) inset;
    background: repeating-linear-gradient(45deg, transparent 0 6px, var(--hover) 6px 7px);
    transition: all var(--dur-fast);

    &:hover { color: var(--ink); box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--ink) 55%, transparent) inset; }
  }
}

@keyframes tile-in { from { opacity: 0; transform: scale(0.6); } }
@keyframes tile-out { to { opacity: 0; transform: scale(0.6); } }
@keyframes spin { to { transform: rotate(360deg); } }

.drop {
  display: flex;
  align-items: center;
  gap: 14px;
  width: 100%;
  padding: 14px 16px;
  border-radius: var(--r-md);
  margin-bottom: 14px;
  color: var(--st-ink-3);
  font-size: 13px;
  line-height: 1.6;
  text-align: left;
  box-shadow: 0 0 0 1.5px var(--line-2) inset;
  transition: all var(--dur-fast);

  &:hover { background: var(--hover); color: var(--st-ink-2); }

  .di {
    width: 36px;
    height: 36px;
    border-radius: var(--r-sm);
    display: grid;
    place-items: center;
    background: var(--well);
    color: var(--st-ink-2);
    flex: none;
  }

  b { color: var(--st-ink); font-weight: 500; }
  small { font-size: 12px; color: var(--st-ink-4); }
}

.cmp-bar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 10px 12px 10px 16px;
  border-top: 1px solid var(--line);

  .sp { flex: 1; }
  .cnt { font-size: 12px; color: var(--st-ink-4); }

  .keys {
    display: flex;
    gap: 3px;
    align-items: center;
    margin: 0 8px 0 10px;
  }
}

.drop-veil {
  position: absolute;
  inset: 6px;
  border-radius: var(--r-md);
  display: grid;
  place-items: center;
  font: 500 15px var(--font-serif);
  color: var(--ink);
  background: color-mix(in oklab, var(--paper) 80%, transparent);
  border: 1.5px dashed color-mix(in oklab, var(--ink) 60%, transparent);
  backdrop-filter: blur(3px);
  pointer-events: none;
  animation: fade-in var(--dur-fast) both;
}

@keyframes fade-in { from { opacity: 0; } }
</style>
