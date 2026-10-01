<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, api, thumbOf, type BatchAction, type Note } from '../../api';
import { useConfigStore } from '../../stores/config';
import { useDialogStore } from '../../stores/dialog';
import { render as renderMarkdown } from '../../utils/markdown';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import EmptyArt from './studio/EmptyArt.vue';
import { Check, Feather } from 'lucide';
import Icon from '../../components/ui/Icon.vue';
import BatchBar from './studio/BatchBar.vue';
import NoteComposer from './studio/NoteComposer.vue';
import { refreshCounts } from './studio/state';
import { toast } from './studio/toast';
import { WEEKDAYS, parseTime, ymd, plainText } from './studio/format';

/** 随想：左栏输入框 + 按月分组的时间线（编辑 / 置顶 / 隐藏 / 删除，勾选后批量）；右栏概览统计条 + 近 6 月柱状图 */
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
    const res = await api.notes({ page: page.value, pageSize: PAGE, all: true });
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

/** 复选框名称：去 Markdown 的纯文本，超过 30 字截断 */
function shortText(md: string): string {
  const s = plainText(md);
  return s.length > 30 ? `${s.slice(0, 30)}…` : s;
}

const groups = computed<Group[]>(() => {
  const out: Group[] = [];
  for (const note of notes.value) {
    const d = parseTime(note.createdAt) ?? new Date();
    // 置顶的（接口排在最前）单独成组，其余按月；否则置顶的旧随想会把新月份挤到最后
    const key = note.pinned ? 'pinned' : `${d.getFullYear()}-${d.getMonth()}`;
    let g = out.find((x) => x.key === key);
    if (!g) {
      g = { key, label: note.pinned ? t('studio.notes.pinnedGroup') : t('studio.notes.month', { y: d.getFullYear(), m: d.getMonth() + 1 }), items: [] };
      out.push(g);
    }
    g.items.push({ note, day: String(d.getDate()).padStart(2, '0'), week: WEEKDAYS[d.getDay()] });
  }
  return out;
});

/* ===== 概览：总数 / 本月 / 带图 / 置顶 + 近 6 个月条数（按月 from/to 取 total） ===== */
interface MonthBar { key: string; label: string; n: number; now: boolean }
const months = ref<MonthBar[]>([]);
const mediaTotal = ref<number | null>(null);
const pinnedN = computed(() => notes.value.filter((n) => n.pinned).length);
const monthMax = computed(() => Math.max(1, ...months.value.map((m) => m.n)));
const moods = computed(() => {
  const m = new Map<string, number>();
  notes.value.forEach((n) => n.mood && m.set(n.mood, (m.get(n.mood) ?? 0) + 1));
  return [...m.entries()].sort((a, b) => b[1] - a[1]).slice(0, 6);
});

async function loadOverview(): Promise<void> {
  const now = new Date();
  const ranges = Array.from({ length: 6 }, (_, i) => {
    const from = new Date(now.getFullYear(), now.getMonth() - 5 + i, 1);
    const to = new Date(from.getFullYear(), from.getMonth() + 1, 0);
    return { from, to };
  });
  const [counts, media] = await Promise.all([
    Promise.all(ranges.map((r) => api.notes({ pageSize: 1, from: ymd(r.from), to: ymd(r.to) }).then((x) => x.total).catch(() => 0))),
    api.notes({ pageSize: 1, media: true }).then((x) => x.total).catch(() => 0),
  ]);
  months.value = ranges.map((r, i) => ({
    key: ymd(r.from),
    label: t('studio.notes.monthShort', { m: r.from.getMonth() + 1 }),
    n: counts[i],
    now: i === ranges.length - 1,
  }));
  mediaTotal.value = media;
}

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
      void loadOverview();
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
  void loadOverview();
}

/* ---------- 勾选与批量：日期下的勾选角标进入选择态，此后点卡片即勾选 ---------- */
const picked = ref<Set<number>>(new Set());
const picking = computed(() => picked.value.size > 0);
/** 所选里有没有可隐藏 / 可取消隐藏的：批量条只给用得上的那个动作（混选时两个都给） */
const pickedHidden = computed(() => {
  const sel = notes.value.filter((x) => picked.value.has(x.id));
  return { any: sel.some((x) => x.hidden), anyShown: sel.some((x) => !x.hidden) };
});

function togglePick(note: Note): void {
  const next = new Set(picked.value);
  if (next.has(note.id)) next.delete(note.id);
  else next.add(note.id);
  picked.value = next;
}

function pickAll(): void {
  picked.value = picked.value.size >= notes.value.length ? new Set() : new Set(notes.value.map((n) => n.id));
}

function onCard(note: Note, e: MouseEvent): void {
  if (!picking.value || (e.target as HTMLElement).closest('button, a')) return;
  togglePick(note);
}

/** 批量（或单条）隐藏 / 取消隐藏 / 删除；删除先确认 */
async function batch(action: BatchAction, ids = [...picked.value]): Promise<void> {
  if (!ids.length) return;
  if (action === 'delete') {
    const ok = await dialog.confirm({
      title: t('studio.batch.deleteTitle', { n: ids.length }),
      message: t('studio.batch.deleteBody'),
      confirmText: t('studio.delete'),
      danger: true,
    });
    if (!ok) return;
  }
  try {
    const { affected } = await adminApi.batchNotes(ids, action);
    const set = new Set(ids);
    if (action === 'delete') {
      notes.value = notes.value.filter((n) => !set.has(n.id));
      total.value -= affected;
    } else {
      notes.value.forEach((n) => { if (set.has(n.id)) n.hidden = action === 'hide'; });
    }
    picked.value = new Set();
    toast(t(`studio.batch.${action === 'hide' ? 'hiddenDone' : action === 'show' ? 'shownDone' : 'deletedDone'}`, { n: affected }), {
      icon: action === 'delete' ? 'trash' : action === 'hide' ? 'eyeOff' : 'eye',
    });
    void refreshCounts();
    void loadOverview();
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape' && picking.value) picked.value = new Set();
}

function edit(note: Note): void {
  void router.push({ name: 'admin-write-note', query: { id: String(note.id) } });
}

onMounted(() => {
  window.addEventListener('keydown', onKey);
  void load();
  void loadOverview();
});
onBeforeUnmount(() => window.removeEventListener('keydown', onKey));
</script>

<template>
  <section class="studio view">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.notes.title') }}</h1>
        <p>{{ config.cfg.thoughts?.subtitle || t('studio.notes.desc') }}</p>
      </div>
      <div class="act">
        <a class="st-btn g" href="/thoughts" target="_blank" rel="noopener"><SIcon name="eye" :size="18" />{{ t('studio.notes.preview') }}</a>
      </div>
    </div>

    <div class="nt-grid">
    <div class="main">
    <div class="tl-cmp st-rise">
      <NoteComposer :placeholder="t('studio.composer.placeholderAlt')" @published="onPublished" />
    </div>

    <div v-if="loaded && !notes.length" class="st-empty">
      <EmptyArt :icon="Feather" />
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
          :class="{ fresh: fresh === it.note.id, bye: leaving.has(it.note.id), pinned: it.note.pinned, on: picked.has(it.note.id), picking, hid: it.note.hidden }"
          :style="{ '--i': Math.min(i, 8) }"
        >
          <div class="d">
            <b class="mono">{{ it.day }}</b>
            <small>{{ it.week }}</small>
            <span class="pk" role="checkbox" tabindex="0" :aria-checked="picked.has(it.note.id)" :aria-label="t('studio.a11y.select', { name: shortText(it.note.contentMd) })" :title="t('studio.batch.pick')" @click.stop="togglePick(it.note)" @keydown.enter.space.prevent.stop="togglePick(it.note)"><span class="st-ck" :class="{ on: picked.has(it.note.id) }"><Icon :icon="Check" /></span></span>
          </div>
          <div class="c" @click="onCard(it.note, $event)">
            <!-- eslint-disable-next-line vue/no-v-html -->
            <div class="md markdown-content" v-html="renderMarkdown(it.note.contentMd, `admin-note-${it.note.id}`)" />
            <div v-if="it.note.images.length" class="imgs" :data-n="gridN(it.note.images.length)">
              <div v-for="src in it.note.images" :key="src" class="t"><img :src="thumbOf(src)" alt="" loading="lazy" /></div>
            </div>
            <div class="foot">
              <span v-if="it.note.mood" class="mood"><i class="st-dot" />{{ it.note.mood }}</span>
              <span v-if="it.note.pinned" class="pin"><SIcon name="pin" :size="16" />{{ t('studio.pinned') }}</span>
              <span v-if="it.note.hidden" class="hid-tag" :title="t('studio.batch.hiddenHint')"><SIcon name="eyeOff" :size="15" />{{ t('studio.batch.hidden') }}</span>
              <span class="sp" />
              <span class="ops">
                <button type="button" class="st-ibtn" :title="t('studio.edit')" @click="edit(it.note)"><SIcon name="pen" :size="18" /></button>
                <button type="button" class="st-ibtn" :class="{ on: it.note.pinned }" :title="it.note.pinned ? t('studio.unpin') : t('studio.pin')" @click="togglePin(it.note)"><SIcon name="pin" :size="18" /></button>
                <button type="button" class="st-ibtn" :title="it.note.hidden ? t('studio.batch.show') : t('studio.batch.hide')" @click="batch(it.note.hidden ? 'show' : 'hide', [it.note.id])">
                  <SIcon :name="it.note.hidden ? 'eye' : 'eyeOff'" :size="18" />
                </button>
                <button type="button" class="st-ibtn" :title="t('studio.delete')" @click="remove(it.note)"><SIcon name="trash" :size="18" /></button>
              </span>
            </div>
          </div>
        </div>
      </template>
    </div>

    <div v-if="notes.length < total" class="more">
      <button type="button" class="st-btn g" :disabled="loadingMore" @click="more">{{ loadingMore ? t('studio.loading') : t('studio.loadMore') }}</button>
    </div>

    <BatchBar :show="picking" :count="picked.size" :total="notes.length" @all="pickAll" @clear="picked = new Set()">
      <button v-if="pickedHidden.anyShown" type="button" class="st-btn sm" @click="batch('hide')"><SIcon name="eyeOff" :size="16" />{{ t('studio.batch.hide') }}</button>
      <button v-if="pickedHidden.any" type="button" class="st-btn sm" @click="batch('show')"><SIcon name="eye" :size="16" />{{ t('studio.batch.show') }}</button>
      <button type="button" class="st-btn sm danger" @click="batch('delete')"><SIcon name="trash" :size="16" />{{ t('studio.delete') }}</button>
    </BatchBar>
    </div>

    <aside class="side">
      <section class="st-card st-rise" style="--i: 1">
        <div class="st-sec-t"><h2>{{ t('studio.notes.overview') }}</h2></div>
        <div class="st-stats two">
          <div class="st-stat"><span v-if="!loaded" class="sk num-sk" /><b v-else>{{ total }}</b><small>{{ t('studio.notes.sTotal') }}</small></div>
          <div class="st-stat"><span v-if="!months.length" class="sk num-sk" /><b v-else>{{ months[months.length - 1].n }}</b><small>{{ t('studio.notes.sMonth') }}</small></div>
          <div class="st-stat"><span v-if="mediaTotal === null" class="sk num-sk" /><b v-else>{{ mediaTotal }}</b><small>{{ t('studio.notes.sImages') }}</small></div>
          <div class="st-stat"><b>{{ pinnedN }}</b><small>{{ t('studio.notes.sPinned') }}</small></div>
        </div>
        <div class="st-sec-t sub"><h3>{{ t('studio.notes.months') }}</h3></div>
        <div class="st-bars" style="--n: 6; --bh: 150px">
          <div
            v-for="(m, i) in months"
            :key="m.key"
            class="st-bar"
            :class="{ on: m.now }"
            :style="{ '--i': i, '--v': m.n / monthMax }"
          >
            <div class="col">
              <em v-if="m.n" class="v">{{ m.n }}</em>
              <div class="stack" :class="{ zero: !m.n }" :style="{ '--a': m.n }"><i class="a" /></div>
            </div>
            <span>{{ m.label }}</span>
          </div>
        </div>
        <template v-if="moods.length">
          <div class="st-sec-t sub"><h3>{{ t('studio.notes.moods') }}</h3></div>
          <div class="moods">
            <span v-for="[mood, n] in moods" :key="mood" class="mood-c">{{ mood }}<b class="mono">{{ n }}</b></span>
          </div>
        </template>
      </section>
    </aside>
    </div>
  </section>
</template>

<style scoped lang="scss">
.view {
  max-width: 1280px;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

/* 左：输入 + 时间线（8）｜ 右：概览卡（4），右栏吸顶 */
.nt-grid {
  display: grid;
  grid-template-columns: minmax(0, 8fr) minmax(0, 4fr);
  align-items: start;
  gap: 28px;
}

.main { min-width: 0; }

.side {
  position: sticky;
  top: 24px;

  .st-sec-t.sub { margin: 4px 0 -6px; }
  h3 { margin: 0; font: 700 16px/1.3 var(--font-serif); }
  .v { font-style: normal; }
}

.st-stats.two {
  grid-template-columns: repeat(2, minmax(0, 1fr));

  .st-stat:nth-child(3) { box-shadow: 0 -1px 0 var(--line-2); }
  .st-stat:nth-child(4) { box-shadow: -1px 0 0 var(--line-2), 0 -1px 0 var(--line-2); }
}

.moods { display: flex; flex-wrap: wrap; gap: 8px; }

.mood-c {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 5px 12px;
  border-radius: var(--r-pill);
  font-size: 13px;
  color: var(--st-ink-2);
  box-shadow: 0 0 0 1px var(--line-2) inset;

  b { font-size: 12px; font-weight: 600; color: var(--st-ink-3); }
}

.tl-cmp { margin-bottom: 4px; }

.tl { position: relative; }

.tl-m {
  font: 700 15px/1 var(--font-serif);
  color: var(--st-ink-2);
  letter-spacing: 0.1em;
  margin: 30px 0 4px 104px;
}

.tl-i {
  position: relative;
  display: grid;
  grid-template-columns: 64px 1fr;
  gap: 16px;
  padding: 2px 0;

  &::before {
    content: '';
    position: absolute;
    left: 76px;
    top: 0;
    bottom: 0;
    width: 1px;
    background: var(--line-2);
  }

  &::after {
    content: '';
    position: absolute;
    left: 72px;
    top: 26px;
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: var(--paper);
    box-shadow: 0 0 0 1.5px var(--st-ink-4);
    transition: all var(--dur) var(--ease-spring);
  }

  &:hover::after, &.pinned::after {
    box-shadow: 0 0 0 1.5px var(--ink);
    background: var(--ink);
  }

  .d {
    text-align: right;
    padding-top: 14px;

    b { display: block; font-size: 24px; line-height: 1.05; font-weight: 600; letter-spacing: -0.02em; }
    small { font-size: 12.5px; color: var(--st-ink-3); }
  }

  .c {
    min-width: 0;
    padding: 12px 16px 8px 20px;
    border-radius: var(--r-md);
    transition: background var(--dur-fast);
  }

  &:hover .c { background: var(--well); }

  .md { word-break: break-word; }

  .imgs {
    display: grid;
    gap: 6px;
    margin: 12px 0 0;
    max-width: 420px;
    grid-template-columns: repeat(3, 1fr);

    &[data-n='1'] { grid-template-columns: 1fr; max-width: 320px; }
    &[data-n='2'], &[data-n='4'] { grid-template-columns: repeat(2, 1fr); max-width: 320px; }

    .t { aspect-ratio: 1; border-radius: var(--r-sm); overflow: hidden; background: var(--well-2); }
    &[data-n='1'] .t { aspect-ratio: 2 / 1; }
    img { width: 100%; height: 100%; object-fit: cover; display: block; }
  }

  .foot {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-top: 8px;
    font-size: 13px;
    color: var(--st-ink-3);
    min-height: 38px;

    .mood, .pin { display: inline-flex; align-items: center; gap: 6px; }
    .mood .st-dot { --c: var(--ink); }
    .pin { color: var(--ink); }
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

  /* 勾选角标：在日期下方，悬停浮现，选择态常显 */
  .pk {
    display: flex;
    width: fit-content;
    margin: 8px 0 0 auto;
    padding: 2px;
    cursor: pointer;
    opacity: 0;
    transition: opacity var(--dur-fast);
  }

  &:hover .pk, &.picking .pk { opacity: 1; }
  &.picking .c { cursor: pointer; }
  &.on .c { background: var(--tint); }

  /* 已隐藏：正文与配图褪色 */
  &.hid .md, &.hid .imgs { opacity: 0.5; }

  .hid-tag { display: inline-flex; align-items: center; gap: 6px; }

  &.fresh .c { animation: fresh 1.6s var(--ease-out); }
  &.bye { animation: collapse 0.45s var(--ease-out) forwards; overflow: hidden; }
}

@keyframes fresh { 0% { background: var(--tint); } 100% { background: transparent; } }
@keyframes collapse { 40% { opacity: 0; transform: translateX(20px); } 100% { opacity: 0; max-height: 0; padding: 0; } }

.more { display: flex; justify-content: center; margin-top: 24px; }


@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
  .nt-grid { grid-template-columns: 1fr; }
  .side { position: static; order: -1; }
}
</style>
