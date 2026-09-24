<script setup lang="ts">
/**
 * 写随想 composer：半屏 / 全屏 detent 的底部 sheet。
 * 正文 Markdown、九图上传（即时上传 + 进度环）、心情、置顶；新建或按 id 编辑。
 * 未发布内容关闭前确认；发布后灵动岛提示并通知内容页刷新。
 */
import { computed, nextTick, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, api, thumbOf, type Note } from '../../api';
import { useConfigStore } from '../../stores/config';
import { useDialogStore } from '../../stores/dialog';
import MaSheet from './MaSheet.vue';
import MaIcon from './MaIcon.vue';
import MaRing from './MaRing.vue';
import MaSwitch from './MaSwitch.vue';
import { cache, loadNotes, shell, toast } from './state';

const props = defineProps<{ open: boolean; noteId: number | null; seq: number }>();
const emit = defineEmits<{ 'update:open': [v: boolean]; stack: [v: number, animated: boolean]; closed: [] }>();
const { t } = useI18n();
const config = useConfigStore();
const dialog = useDialogStore();

const MAX_IMAGES = 9;
const SOFT_LIMIT = 500;
const MOODS = ['record', 'idea', 'joy', 'calm', 'tired'] as const;

const sheet = ref<InstanceType<typeof MaSheet> | null>(null);
const textEl = ref<HTMLTextAreaElement | null>(null);

const content = ref('');
const mood = ref('');
const images = ref<string[]>([]);
const pinned = ref(false);
const uploads = ref<{ key: number; preview: string }[]>([]);
const busy = ref(false);
const loading = ref(false);
let snapshot = '';
let uploadKey = 0;

const editing = computed(() => props.noteId !== null);
const moodLabels = computed(() => MOODS.map((m) => t(`mobileAdmin.note.moods.${m}`)));
const customMood = computed(() => (mood.value && !moodLabels.value.includes(mood.value) ? mood.value : ''));
const count = computed(() => content.value.length);
const canPost = computed(() => !!content.value.trim() && !busy.value && !uploads.value.length);
const avatar = computed(() => config.cfg.about.avatar || '/favicon-256.png');
const name = computed(() => config.cfg.about.name || 'Myself');

const state = (): string => JSON.stringify([content.value, mood.value, images.value, pinned.value]);
const dirty = computed(() => state() !== snapshot);

function reset(): void {
  content.value = '';
  mood.value = '';
  images.value = [];
  pinned.value = false;
  uploads.value = [];
  snapshot = state();
}

async function findNote(id: number): Promise<Note | undefined> {
  const hit = cache.notes?.find((n) => n.id === id);
  if (hit) return hit;
  for (let page = 1; page <= 10; page += 1) {
    const res = await api.notes({ page, pageSize: 50 });
    const n = res.items.find((x) => x.id === id);
    if (n || res.items.length < 50) return n;
  }
  return undefined;
}

watch(
  () => [props.open, props.seq] as const,
  async ([open]) => {
    if (!open) return;
    reset();
    if (props.noteId === null) return;
    loading.value = true;
    try {
      const n = await findNote(props.noteId);
      if (n) {
        content.value = n.contentMd;
        mood.value = n.mood;
        images.value = [...n.images];
        pinned.value = n.pinned;
        snapshot = state();
        await nextTick();
        autosize();
      }
    } finally {
      loading.value = false;
    }
  },
  { immediate: true },
);

async function beforeClose(): Promise<boolean> {
  if (!dirty.value || busy.value) return !busy.value;
  return dialog.confirm({
    title: t('mobileAdmin.note.discardTitle'),
    message: t('mobileAdmin.note.discardMsg'),
    confirmText: t('mobileAdmin.note.discard'),
    cancelText: t('mobileAdmin.note.keep'),
    danger: true,
  });
}

function autosize(): void {
  const el = textEl.value;
  if (!el) return;
  el.style.height = 'auto';
  el.style.height = `${Math.max(104, el.scrollHeight)}px`;
}

function onFocus(): void {
  if (window.matchMedia('(pointer: coarse)').matches) sheet.value?.expand();
}

function insertAtCursor(before: string, after = ''): void {
  const el = textEl.value;
  if (!el) return;
  const s = el.selectionStart ?? content.value.length;
  const e = el.selectionEnd ?? s;
  const sel = content.value.slice(s, e);
  content.value = content.value.slice(0, s) + before + sel + after + content.value.slice(e);
  void nextTick(() => {
    el.focus();
    const pos = s + before.length + sel.length;
    el.setSelectionRange(pos, pos);
    autosize();
  });
}

async function onFiles(e: Event): Promise<void> {
  const input = e.target as HTMLInputElement;
  const room = MAX_IMAGES - images.value.length - uploads.value.length;
  const files = Array.from(input.files ?? []).slice(0, Math.max(0, room));
  input.value = '';
  if (!files.length) {
    if (room <= 0) toast(t('mobileAdmin.note.maxImages'), '', 'info');
    return;
  }
  const tiles = files.map((f) => ({ key: ++uploadKey, preview: URL.createObjectURL(f) }));
  uploads.value.push(...tiles);
  try {
    const items = await adminApi.uploadMedia(files);
    images.value.push(...items.map((i) => i.url));
    shell.bump.media += 1;
  } catch {
    toast(t('mobileAdmin.common.uploadFailed'), '', 'error');
  } finally {
    const keys = new Set(tiles.map((x) => x.key));
    uploads.value = uploads.value.filter((u) => !keys.has(u.key));
    tiles.forEach((x) => URL.revokeObjectURL(x.preview));
  }
}

function removeImage(i: number): void {
  images.value.splice(i, 1);
}

function pickMood(m: string): void {
  mood.value = mood.value === m ? '' : m;
}

async function customMoodPrompt(): Promise<void> {
  const v = await dialog.prompt({
    title: t('mobileAdmin.note.customMood'),
    placeholder: t('mobileAdmin.note.customMoodHint'),
    inputValue: customMood.value,
    confirmText: t('mobileAdmin.common.ok'),
    cancelText: t('mobileAdmin.common.cancel'),
  });
  if (v !== null) mood.value = v.trim().slice(0, 12);
}

async function publish(): Promise<void> {
  if (!canPost.value) return;
  busy.value = true;
  const body = { contentMd: content.value.trim(), mood: mood.value, images: images.value, pinned: pinned.value };
  try {
    if (props.noteId === null) await adminApi.createNote(body);
    else await adminApi.updateNote(props.noteId, body);
    snapshot = state();
    emit('update:open', false);
    shell.bump.notes += 1;
    void loadNotes().catch(() => undefined);
    const preview = body.contentMd.replace(/\s+/g, ' ').slice(0, 12);
    window.setTimeout(
      () => toast(editing.value ? t('mobileAdmin.note.saved') : t('mobileAdmin.note.published'), preview),
      320,
    );
  } catch {
    toast(t('mobileAdmin.common.saveFailed'), '', 'error');
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <MaSheet
    ref="sheet"
    :open="open"
    :detents="['half', 'full']"
    :before-close="beforeClose"
    :label="editing ? t('mobileAdmin.note.edit') : t('mobileAdmin.note.new')"
    stackable
    @update:open="emit('update:open', $event)"
    @stack="(v, a) => emit('stack', v, a)"
    @closed="emit('closed')"
  >
    <template #head="{ close }">
      <div class="cp-head">
        <button class="txtbtn tap" @click="close">{{ t('mobileAdmin.common.cancel') }}</button>
        <b>{{ editing ? t('mobileAdmin.note.edit') : t('mobileAdmin.note.new') }}</b>
        <button class="pill-btn tap" :disabled="!canPost" @click="publish">
          {{ busy ? t('mobileAdmin.common.saving') : editing ? t('mobileAdmin.common.save') : t('mobileAdmin.note.post') }}
        </button>
      </div>
    </template>

    <div class="cp-body" :class="{ loading }">
      <div class="cp-who">
        <img class="av" :src="avatar" alt="" draggable="false" />
        <b>{{ name }}</b>
        <span class="cp-vis"><MaIcon name="globe" :size="12" />{{ t('mobileAdmin.note.public') }}</span>
      </div>

      <textarea
        ref="textEl"
        v-model="content"
        :placeholder="t('mobileAdmin.note.placeholder')"
        rows="4"
        @input="autosize"
        @focus="onFocus"
      />

      <div class="cp-media">
        <div v-for="(img, i) in images" :key="img" class="t">
          <img :src="thumbOf(img)" alt="" draggable="false" />
          <button class="x" :aria-label="t('mobileAdmin.common.remove')" @click="removeImage(i)">
            <MaIcon name="close" :size="11" />
          </button>
        </div>
        <div v-for="u in uploads" :key="u.key" class="t up">
          <img :src="u.preview" alt="" />
          <span class="prog"><MaRing indeterminate :size="30" /></span>
        </div>
        <label v-if="images.length + uploads.length < MAX_IMAGES" class="add tap" :aria-label="t('mobileAdmin.note.addImage')">
          <MaIcon name="plus" />
          <input type="file" accept="image/*" multiple hidden @change="onFiles" />
        </label>
      </div>

      <div class="cp-tools">
        <label class="tool tap" :aria-label="t('mobileAdmin.note.addImage')">
          <MaIcon name="image" :size="19" />
          <input type="file" accept="image/*" multiple hidden @change="onFiles" />
        </label>
        <label class="tool tap" :aria-label="t('mobileAdmin.note.camera')">
          <MaIcon name="camera" :size="19" />
          <input type="file" accept="image/*" capture="environment" hidden @change="onFiles" />
        </label>
        <button class="tool tap" :aria-label="t('mobileAdmin.md.tag')" @click="insertAtCursor(' #')"><MaIcon name="hash" :size="19" /></button>
        <button class="tool tap" :aria-label="t('mobileAdmin.md.bold')" @click="insertAtCursor('**', '**')"><MaIcon name="bold" :size="19" /></button>
        <span class="cp-count" :class="{ over: count > SOFT_LIMIT }">
          <span>{{ count }}/{{ SOFT_LIMIT }}</span>
          <svg viewBox="0 0 22 22"><circle class="bg" cx="11" cy="11" r="8" /><circle class="fg" cx="11" cy="11" r="8" :style="{ strokeDashoffset: 50.27 * (1 - Math.min(1, count / SOFT_LIMIT)) }" /></svg>
        </span>
      </div>

      <div class="cp-more">
        <h5>{{ t('mobileAdmin.note.mood') }}</h5>
        <div class="cp-moods">
          <button v-for="m in moodLabels" :key="m" class="chip tap" :class="{ on: mood === m }" @click="pickMood(m)">{{ m }}</button>
          <button v-if="customMood" class="chip tap on" @click="customMoodPrompt">{{ customMood }}</button>
          <button v-else class="chip tap ghost" @click="customMoodPrompt"><MaIcon name="plus" :size="14" />{{ t('mobileAdmin.note.customMood') }}</button>
        </div>
        <div class="ma-list">
          <div class="ma-li">
            <span class="lic" style="--c: #ff9500"><MaIcon name="pin" :size="17" /></span>
            <span class="lt">{{ t('mobileAdmin.note.pin') }}</span>
            <MaSwitch v-model="pinned" :label="t('mobileAdmin.note.pin')" />
          </div>
          <div class="ma-li">
            <span class="lic" style="--c: #5b6b86"><MaIcon name="type" :size="17" /></span>
            <span class="lt">{{ t('mobileAdmin.note.markdown') }}</span>
            <small>{{ t('mobileAdmin.note.markdownHint') }}</small>
          </div>
        </div>
      </div>
    </div>
  </MaSheet>
</template>

<style scoped lang="scss">
.cp-head {
  display: flex;
  align-items: center;
  padding: 0 16px 10px;

  b {
    flex: 1;
    text-align: center;
    font-size: 16.5px;
    font-weight: 600;
  }
}

.cp-body {
  padding: 6px 18px 0;
  transition: opacity var(--dur);

  &.loading { opacity: 0.4; }
}

.cp-who {
  display: flex;
  align-items: center;
  gap: 10px;

  .av {
    width: 38px;
    height: 38px;
    border-radius: 50%;
    object-fit: cover;
    box-shadow: 0 0 0 0.5px var(--line-2);
  }

  b { font-size: 15px; }
}

.cp-vis {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: 4px;
  padding: 3px 9px;
  border-radius: 999px;
  background: var(--soft);
  color: var(--ink);
  font-size: 12px;
}

textarea {
  display: block;
  width: 100%;
  min-height: 104px;
  margin-top: 12px;
  resize: none;
  font-size: 17px;
  line-height: 1.7;
  color: var(--text);
  caret-color: var(--primary);
  overflow: hidden;

  &::placeholder { color: var(--text-3); }
}

.cp-media {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-top: 6px;

  .t,
  .add {
    position: relative;
    aspect-ratio: 1;
    border-radius: 14px;
    overflow: hidden;
    animation: ma-tile-in 0.45s var(--ease-spring) backwards;
  }

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .up img { filter: blur(3px) brightness(0.6); }

  .prog {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
  }

  .x {
    position: absolute;
    top: 5px;
    right: 5px;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: rgba(0, 0, 0, 0.55);
    color: #fff;

    :deep(.ma-ic) { stroke-width: 2.4; }
  }

  .add {
    display: grid;
    place-items: center;
    color: var(--text-3);
    box-shadow: inset 0 0 0 1px var(--line-2);
    background: var(--fill);
    cursor: pointer;
  }
}

@keyframes ma-tile-in {
  from { opacity: 0; transform: scale(0.7); }
}

.cp-tools {
  display: flex;
  align-items: center;
  gap: 4px;
  margin: 16px -6px 0;
  padding-top: 10px;
  border-top: 0.5px solid var(--line);

  .tool {
    width: 40px;
    height: 40px;
    border-radius: 12px;
    display: grid;
    place-items: center;
    color: var(--text-2);
    cursor: pointer;

    &:active { background: var(--fill-2); }
  }
}

.cp-count {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 8px;
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-3);

  svg {
    width: 22px;
    height: 22px;
    transform: rotate(-90deg);
  }

  circle {
    fill: none;
    stroke-width: 2.6;
  }

  .bg { stroke: var(--fill-2); }

  .fg {
    stroke: var(--primary);
    stroke-linecap: round;
    stroke-dasharray: 50.27;
    transition: stroke-dashoffset var(--dur) var(--ease-out), stroke var(--dur);
  }

  &.over {
    color: var(--accent-yellow);
    .fg { stroke: var(--accent-yellow); }
  }
}

.cp-more {
  margin-top: 22px;

  h5 {
    font-size: 12.5px;
    font-weight: 500;
    color: var(--text-3);
    letter-spacing: 0.06em;
    margin-bottom: 10px;
  }

  .ma-list {
    margin-top: 18px;
    background: var(--fill);
  }
}

.cp-moods {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;

  .ghost {
    color: var(--text-3);
    background: transparent;
    box-shadow: inset 0 0 0 1px var(--line-2);
  }
}
</style>
