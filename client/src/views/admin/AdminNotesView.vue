<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, api, thumbOf, type Note } from '../../api';
import { useConfigStore } from '../../stores/config';
import { useDialogStore } from '../../stores/dialog';
import { render as renderMarkdown } from '../../utils/markdown';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import DoorArt from './studio/DoorArt.vue';
import NoteComposer from './studio/NoteComposer.vue';
import { refreshCounts } from './studio/state';
import { toast } from './studio/toast';
import { WEEKDAYS, parseTime } from './studio/format';

/** 随想：顶部输入框 + 按月分组的时间线（编辑 / 置顶 / 删除） */
const { t } = useI18n();
const router = useRouter();
const config = useConfigStore();
const dialog = useDialogStore();

const PAGE = 30;
const notes = ref<Note[]>([]);
const total = ref(0);
const page = ref(1);
const loaded = ref(false);
const loadingMore = ref(false);
const fresh = ref<number | null>(null);
const leaving = ref<Set<number>>(new Set());

async function load(reset = true): Promise<void> {
  if (reset) page.value = 1;
  try {
    const res = await api.notes({ page: page.value, pageSize: PAGE });
    notes.value = reset ? res.items : [...notes.value, ...res.items];
    total.value = res.total;
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
  loaded.value = true;
}

async function more(): Promise<void> {
  if (loadingMore.value) return;
  loadingMore.value = true;
  page.value += 1;
  await load(false);
  loadingMore.value = false;
}

interface Group {
  key: string;
  label: string;
  items: { note: Note; day: string; week: string }[];
}

const groups = computed<Group[]>(() => {
  const out: Group[] = [];
  for (const note of notes.value) {
    const d = parseTime(note.createdAt) ?? new Date();
    const key = `${d.getFullYear()}-${d.getMonth()}`;
    let g = out.find((x) => x.key === key);
    if (!g) {
      g = { key, label: t('studio.notes.month', { y: d.getFullYear(), m: d.getMonth() + 1 }), items: [] };
      out.push(g);
    }
    g.items.push({ note, day: String(d.getDate()).padStart(2, '0'), week: WEEKDAYS[d.getDay()] });
  }
  return out;
});

function gridN(n: number): number {
  return Math.min(9, n);
}

async function togglePin(note: Note): Promise<void> {
  try {
    await adminApi.updateNote(note.id, { contentMd: note.contentMd, mood: note.mood, images: note.images, pinned: !note.pinned });
    note.pinned = !note.pinned;
    toast(note.pinned ? t('studio.pinned') : t('studio.unpinned'), { icon: 'pin' });
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

async function remove(note: Note): Promise<void> {
  const ok = await dialog.confirm({
    title: t('studio.notes.deleteTitle'),
    message: t('studio.notes.deleteBody'),
    confirmText: t('studio.delete'),
    danger: true,
  });
  if (!ok) return;
  try {
    await adminApi.deleteNote(note.id);
    leaving.value = new Set([...leaving.value, note.id]);
    window.setTimeout(() => {
      notes.value = notes.value.filter((n) => n.id !== note.id);
      total.value -= 1;
    }, 450);
    toast(t('studio.notes.deleted'), { icon: 'trash' });
    void refreshCounts();
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

async function onPublished(id: number): Promise<void> {
  toast(t('studio.composer.published'));
  await load();
  fresh.value = id;
  void refreshCounts();
}

function edit(note: Note): void {
  void router.push({ name: 'admin-write-note', query: { id: String(note.id) } });
}

onMounted(() => void load());
</script>

<template>
  <section class="studio view">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.notes.title') }}</h1>
        <p>{{ config.cfg.thoughts?.subtitle || t('studio.notes.desc') }}</p>
      </div>
      <div class="act">
        <a class="st-btn g" href="/thoughts" target="_blank" rel="noopener"><SIcon name="eye" :size="16" />{{ t('studio.notes.preview') }}</a>
      </div>
    </div>

    <div class="tl-cmp st-rise">
      <NoteComposer :placeholder="t('studio.composer.placeholderAlt')" @published="onPublished" />
    </div>

    <div v-if="loaded && !notes.length" class="st-empty">
      <DoorArt />
      <h4>{{ t('studio.notes.empty') }}</h4>
      <p>{{ t('studio.notes.emptySub') }}</p>
    </div>

    <div class="tl">
      <template v-for="g in groups" :key="g.key">
        <div class="tl-m">{{ g.label }}</div>
        <div
          v-for="(it, i) in g.items"
          :key="it.note.id"
          class="tl-i st-rise"
          :class="{ fresh: fresh === it.note.id, bye: leaving.has(it.note.id), pinned: it.note.pinned }"
          :style="{ '--i': Math.min(i, 8) }"
        >
          <div class="d">
            <b class="mono">{{ it.day }}</b>
            <small>{{ it.week }}</small>
          </div>
          <div class="c">
            <!-- eslint-disable-next-line vue/no-v-html -->
            <div class="md" v-html="renderMarkdown(it.note.contentMd)" />
            <div v-if="it.note.images.length" class="imgs" :data-n="gridN(it.note.images.length)">
              <div v-for="src in it.note.images" :key="src" class="t"><img :src="thumbOf(src)" alt="" loading="lazy" /></div>
            </div>
            <div class="foot">
              <span v-if="it.note.mood" class="mood"><i class="st-dot" />{{ it.note.mood }}</span>
              <span v-if="it.note.pinned" class="pin"><SIcon name="pin" :size="14" />{{ t('studio.pinned') }}</span>
              <span class="sp" />
              <span class="ops">
                <button type="button" class="st-ibtn" :title="t('studio.edit')" @click="edit(it.note)"><SIcon name="pen" :size="16" /></button>
                <button type="button" class="st-ibtn" :class="{ on: it.note.pinned }" :title="it.note.pinned ? t('studio.unpin') : t('studio.pin')" @click="togglePin(it.note)"><SIcon name="pin" :size="16" /></button>
                <button type="button" class="st-ibtn" :title="t('studio.delete')" @click="remove(it.note)"><SIcon name="trash" :size="16" /></button>
              </span>
            </div>
          </div>
        </div>
      </template>
    </div>

    <div v-if="notes.length < total" class="more">
      <button type="button" class="st-btn g" :disabled="loadingMore" @click="more">{{ loadingMore ? t('studio.loading') : t('studio.loadMore') }}</button>
    </div>
  </section>
</template>

<style scoped lang="scss">
.view {
  max-width: 1120px;
  margin: 0 auto;
  padding: 52px 64px 96px;
}

.tl-cmp { max-width: 780px; margin-bottom: 8px; }

.tl { position: relative; max-width: 780px; }

.tl-m {
  font: 600 13px/1 var(--font-serif);
  color: var(--ink-3);
  letter-spacing: 0.14em;
  margin: 44px 0 6px 128px;
}

.tl-i {
  position: relative;
  display: grid;
  grid-template-columns: 80px 1fr;
  gap: 24px;
  padding: 4px 0;

  &::before {
    content: '';
    position: absolute;
    left: 92px;
    top: 0;
    bottom: 0;
    width: 1px;
    background: var(--line-2);
  }

  &::after {
    content: '';
    position: absolute;
    left: 88px;
    top: 34px;
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: var(--paper);
    box-shadow: 0 0 0 1.5px var(--ink-4);
    transition: all var(--dur) var(--ease-spring);
  }

  &:hover::after, &.pinned::after {
    box-shadow: 0 0 0 1.5px var(--primary), 0 0 0 5px var(--primary-ring);
    background: var(--primary);
  }

  .d {
    text-align: right;
    padding-top: 22px;

    b { display: block; font-size: 20px; line-height: 1; font-weight: 500; letter-spacing: -0.02em; }
    small { font-size: 12px; color: var(--ink-3); }
  }

  .c {
    min-width: 0;
    padding: 18px 20px 12px 24px;
    border-radius: 16px;
    transition: background var(--dur-fast);
  }

  &:hover .c { background: var(--well); }

  .md {
    font: 400 16px/1.9 var(--font-serif);
    word-break: break-word;

    :deep(p) { margin: 0 0 0.6em; }
    :deep(p:last-child) { margin-bottom: 0; }
    :deep(a) { color: var(--primary-ink); text-decoration: underline; text-underline-offset: 3px; }
    :deep(code) { font: 14px var(--font-mono); background: var(--well-2); padding: 1px 5px; border-radius: 5px; }
    :deep(ul), :deep(ol) { padding-left: 1.3em; margin: 0 0 0.6em; }
    :deep(blockquote) { margin: 0.6em 0; padding-left: 1em; border-left: 2px solid var(--primary-ring); color: var(--ink-2); }
  }

  .imgs {
    display: grid;
    gap: 6px;
    margin: 14px 0 0;
    max-width: 400px;
    grid-template-columns: repeat(3, 1fr);

    &[data-n='1'] { grid-template-columns: 1fr; max-width: 300px; }
    &[data-n='2'], &[data-n='4'] { grid-template-columns: repeat(2, 1fr); max-width: 300px; }

    .t { aspect-ratio: 1; border-radius: 10px; overflow: hidden; background: var(--well-2); }
    &[data-n='1'] .t { aspect-ratio: 16 / 10; }
    img { width: 100%; height: 100%; object-fit: cover; display: block; }
  }

  .foot {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-top: 12px;
    font-size: 12.5px;
    color: var(--ink-3);
    min-height: 34px;

    .mood, .pin { display: inline-flex; align-items: center; gap: 6px; }
    .mood .st-dot { --c: var(--primary); }
    .pin { color: var(--primary-ink); }
    .sp { flex: 1; }
  }

  .ops {
    display: flex;
    gap: 2px;
    opacity: 0;
    transform: translateX(6px);
    transition: all var(--dur) var(--ease-out);
  }

  &:hover .ops, .ops:focus-within { opacity: 1; transform: none; }

  &.fresh .c { animation: fresh 1.6s var(--ease-out); }
  &.bye { animation: collapse 0.45s var(--ease-out) forwards; overflow: hidden; }
}

@keyframes fresh { 0% { background: var(--primary-soft-2); } 100% { background: transparent; } }
@keyframes collapse { 40% { opacity: 0; transform: translateX(20px); } 100% { opacity: 0; max-height: 0; padding: 0; } }

.more { max-width: 780px; display: flex; justify-content: center; margin-top: 32px; }

@media (max-width: 1180px) {
  .view { padding: 40px 36px 80px; }
}
</style>
