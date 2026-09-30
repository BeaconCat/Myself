<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  adminApi, api, thumbOf, type AdminPost, type BackupInfo, type CompressJob, type MediaItem, type Note, type QualityItem,
} from '../../api';
import { useConfigStore } from '../../stores/config';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import LightCover from './studio/LightCover.vue';
import NoteComposer from './studio/NoteComposer.vue';
import { refreshCounts } from './studio/state';
import { toast } from './studio/toast';
import {
  WEEKDAYS, WEEK_SHORT, formatSize, greeting, parseTime, relTime, solarTerm, startOfDay, ymd,
} from './studio/format';

/** 「今天」：紧凑问候 + 关键数字统计条 + 随想输入 + 本周柱状图 / 待办 + 继续写作 */
const { t } = useI18n();
const router = useRouter();
const config = useConfigStore();

const now = new Date();
const term = solarTerm(now);
const posts = ref<AdminPost[]>([]);
const notes = ref<Note[]>([]);
const quality = ref<QualityItem[] | null>(null);
const loaded = ref(false);
const notesTotal = ref(0);
const media = ref<MediaItem[] | null>(null);
const backups = ref<BackupInfo[] | null>(null);
const composer = ref<InstanceType<typeof NoteComposer> | null>(null);
/** 待审评论数（null = 读取中） */
const pendingComments = ref<number | null>(null);

async function load(): Promise<void> {
  const from = new Date(now);
  from.setDate(from.getDate() - 15);
  const [p, n, nt] = await Promise.all([
    adminApi.posts().catch(() => [] as AdminPost[]),
    api.notes({ pageSize: 50, from: ymd(from), all: true }).then((r) => r.items).catch(() => [] as Note[]),
    api.notes({ pageSize: 1 }).then((r) => r.total).catch(() => 0),
  ]);
  posts.value = p;
  notes.value = n;
  notesTotal.value = nt;
  loaded.value = true;
  adminApi.media().then((m) => (media.value = m)).catch(() => (media.value = []));
  adminApi.backups().then((b) => (backups.value = b)).catch(() => (backups.value = []));
  adminApi.comments('pending').then((c) => (pendingComments.value = c.length)).catch(() => (pendingComments.value = 0));
  adminApi.qualityScan()
    .then((q) => (quality.value = q.filter((i) => i.compressible)))
    .catch(() => (quality.value = []));
}

/* ===== 问候 ===== */
const name = computed(() => config.cfg.about?.name || 'Myself');
const dateLine = computed(() => {
  const y = now.getFullYear();
  const m = String(now.getMonth() + 1).padStart(2, '0');
  const d = String(now.getDate()).padStart(2, '0');
  return `${y} · ${m} · ${d}`;
});
const termText = computed(() =>
  term.days === 0 ? t('studio.today.termToday', { term: term.name }) : t('studio.today.termDays', { term: term.name, n: term.days + 1 }),
);
const siteDays = computed(() => {
  const f = config.cfg.about?.foundedAt;
  const d = f ? new Date(`${f}T00:00:00`) : null;
  if (!d || Number.isNaN(d.getTime())) return 0;
  return Math.max(1, Math.floor((startOfDay(now).getTime() - d.getTime()) / 86400000) + 1);
});
const drafts = computed(() =>
  posts.value
    .filter((p) => p.status === 'draft')
    .sort((a, b) => (b.updatedAt || b.createdAt).localeCompare(a.updatedAt || a.createdAt)),
);
const lede = computed(() => {
  if (drafts.value.length) return t('studio.today.ledeDrafts', { n: drafts.value.length });
  if (term.days === 0) return t('studio.today.ledeTerm', { term: term.name });
  return t('studio.today.lede');
});

/* ===== 关键数字 ===== */
const published = computed(() => posts.value.filter((p) => p.status === 'published').length);
const mediaSize = computed(() => (media.value ?? []).reduce((s, m) => s + m.size, 0));
const lastBackup = computed(() => (backups.value ?? []).reduce<BackupInfo | null>(
  (a, b) => (!a || b.createdAt > a.createdAt ? b : a), null,
));

/* ===== 继续写作 ===== */
const shelf = computed(() => {
  if (drafts.value.length) return { title: t('studio.today.continue'), items: drafts.value.slice(0, 3), draft: true };
  return { title: t('studio.today.recent'), items: posts.value.slice(0, 3), draft: false };
});

function edit(p: AdminPost): void {
  void router.push({ name: 'admin-write-post', query: { id: String(p.id) } });
}

/* ===== 待你处理 ===== */
const job = ref<CompressJob | null>(null);
const compressDone = ref<{ n: number; saved: number } | null>(null);
let pollTimer = 0;

const compressible = computed(() => quality.value ?? []);
const compressSize = computed(() => compressible.value.reduce((s, i) => s + i.size, 0));
const pending = computed(() => (drafts.value.length ? 1 : 0) + (compressible.value.length && !compressDone.value ? 1 : 0));

async function runCompress(): Promise<void> {
  if (job.value || !compressible.value.length) return;
  try {
    const { id } = await adminApi.qualityCompress(compressible.value.map((i) => i.name), 80);
    job.value = await adminApi.qualityJob(id);
    const tick = async (): Promise<void> => {
      try {
        job.value = await adminApi.qualityJob(id);
      } catch {
        job.value = null;
        return;
      }
      if (job.value.running) {
        pollTimer = window.setTimeout(() => void tick(), 600);
        return;
      }
      const res = job.value.results.filter((r) => !r.error);
      const saved = res.reduce((s, r) => s + Math.max(0, (r.before ?? 0) - (r.after ?? 0)), 0);
      compressDone.value = { n: res.length, saved };
      job.value = null;
      toast(t('studio.today.compressToast', { size: formatSize(saved) }));
      quality.value = (await adminApi.qualityScan()).filter((i) => i.compressible);
    };
    pollTimer = window.setTimeout(() => void tick(), 600);
  } catch (err) {
    job.value = null;
    toast((err as Error).message === 'job_running' ? t('studio.media.jobRunning') : t('studio.saveFailed'), { icon: 'x' });
  }
}
const jobPercent = computed(() => (job.value && job.value.total ? Math.round((job.value.done / job.value.total) * 100) : 0));

/* ===== 本周小结 ===== */
const week = computed(() => {
  const today = startOfDay(now);
  const monday = new Date(today);
  monday.setDate(today.getDate() - ((today.getDay() + 6) % 7));
  const lastMonday = new Date(monday);
  lastMonday.setDate(monday.getDate() - 7);
  const days = Array.from({ length: 7 }, (_, i) => {
    const d = new Date(monday);
    d.setDate(monday.getDate() + i);
    return { key: ymd(d), label: WEEK_SHORT[d.getDay()], posts: 0, notes: 0, future: d > today, today: d.getTime() === today.getTime() };
  });
  let lastWeek = 0;
  const bump = (iso: string, kind: 'posts' | 'notes') => {
    const d = parseTime(iso);
    if (!d) return;
    const day = days.find((x) => x.key === ymd(d));
    if (day) day[kind] += 1;
    else if (d >= lastMonday && d < monday) lastWeek += 1;
  };
  posts.value.forEach((p) => bump(p.createdAt, 'posts'));
  notes.value.forEach((n) => bump(n.createdAt, 'notes'));
  const total = days.reduce((s, d) => s + d.posts + d.notes, 0);
  const max = Math.max(1, ...days.map((d) => d.posts + d.notes));
  const best = days.reduce((a, b) => (b.posts + b.notes > a.posts + a.notes ? b : a), days[0]);
  return {
    days,
    max,
    posts: days.reduce((s, d) => s + d.posts, 0),
    notes: days.reduce((s, d) => s + d.notes, 0),
    total,
    diff: total - lastWeek,
    best: best.posts + best.notes > 0 ? WEEKDAYS[(days.indexOf(best) + 1) % 7] : '',
  };
});

const weekSub = computed(() => {
  const w = week.value;
  if (!w.total) return t('studio.today.weekEmpty');
  const cmp = w.diff > 0 ? t('studio.today.weekMore', { n: w.diff }) : w.diff < 0 ? t('studio.today.weekLess', { n: -w.diff }) : t('studio.today.weekSame');
  return w.best ? `${cmp}${t('studio.today.weekBest', { day: w.best })}` : cmp;
});

function onPublished(): void {
  toast(t('studio.composer.published'), { action: t('studio.view'), fn: () => void router.push({ name: 'admin-notes' }) });
  void load();
  void refreshCounts();
}

function onCompose(): void {
  composer.value?.focus();
}

onMounted(() => {
  void load();
  window.addEventListener('studio:compose', onCompose);
});
onBeforeUnmount(() => {
  window.clearTimeout(pollTimer);
  window.removeEventListener('studio:compose', onCompose);
});
</script>

<template>
  <section class="studio view today">
    <header class="hello st-rise" style="--i: 0">
      <div class="hl">
        <h1>{{ greeting(now.getHours()) }}，{{ name }}。</h1>
        <p>{{ lede }}</p>
      </div>
      <div class="date">
        <span class="mono">{{ dateLine }}</span>
        <span>{{ WEEKDAYS[now.getDay()] }}</span>
        <span class="term">{{ termText }}</span>
      </div>
    </header>

    <div class="st-stats kpis st-rise" style="--i: 1; --n: 6">
      <router-link class="st-stat" :to="{ name: 'admin-posts', query: { status: 'published' } }">
        <span v-if="!loaded" class="sk num-sk" /><b v-else>{{ published }}</b>
        <small>{{ t('studio.today.kPublished') }}</small>
      </router-link>
      <router-link class="st-stat" :to="{ name: 'admin-posts', query: { status: 'draft' } }">
        <span v-if="!loaded" class="sk num-sk" /><b v-else>{{ drafts.length }}</b>
        <small>{{ t('studio.today.kDrafts') }}</small>
      </router-link>
      <router-link class="st-stat" :to="{ name: 'admin-notes' }">
        <span v-if="!loaded" class="sk num-sk" /><b v-else>{{ notesTotal }}</b>
        <small>{{ t('studio.today.kNotes') }}</small>
      </router-link>
      <router-link class="st-stat" :to="{ name: 'admin-media' }">
        <span v-if="media === null" class="sk num-sk" /><b v-else>{{ media.length }}</b>
        <small>{{ t('studio.today.kMedia', { size: formatSize(mediaSize) }) }}</small>
      </router-link>
      <router-link class="st-stat" :to="{ name: 'admin-data' }">
        <span v-if="backups === null" class="sk num-sk" /><b v-else>{{ backups.length }}</b>
        <small>{{ lastBackup ? t('studio.today.kBackups', { when: relTime(lastBackup.createdAt) }) : t('studio.today.kBackupsNone') }}</small>
      </router-link>
      <div class="st-stat">
        <b>{{ siteDays }}</b>
        <small>{{ t('studio.today.kDays') }}</small>
      </div>
    </div>

    <div class="st-rise cmp-row" style="--i: 2">
      <NoteComposer ref="composer" hint @published="onPublished" />
    </div>

    <div class="row">
      <section class="st-card week st-rise" style="--i: 3">
        <div class="st-sec-t">
          <h2>{{ t('studio.today.weekTitle') }}</h2>
          <span>{{ weekSub }}</span>
        </div>
        <div class="st-stats" style="--n: 4">
          <div class="st-stat"><b>{{ week.posts }}</b><small>{{ t('studio.today.wPosts') }}</small></div>
          <div class="st-stat"><b>{{ week.notes }}</b><small>{{ t('studio.today.wNotes') }}</small></div>
          <div class="st-stat">
            <b :class="week.diff > 0 ? 'up' : week.diff < 0 ? 'down' : ''">{{ week.diff > 0 ? '+' : '' }}{{ week.diff }}</b>
            <small>{{ t('studio.today.wDiff') }}</small>
          </div>
          <div class="st-stat"><b class="best">{{ week.best || '—' }}</b><small>{{ t('studio.today.wBest') }}</small></div>
        </div>
        <div class="chart">
          <div class="st-bars" style="--n: 7; --bh: 140px">
            <div
              v-for="(d, i) in week.days"
              :key="d.key"
              class="st-bar"
              :class="{ on: d.today, fut: d.future }"
              :style="{ '--i': i, '--v': (d.posts + d.notes) / week.max }"
              :title="d.future ? undefined : t('studio.today.barTip', { p: d.posts, n: d.notes })"
            >
              <div class="col">
                <em v-if="!d.future && d.posts + d.notes" class="v">{{ d.posts + d.notes }}</em>
                <div class="stack" :class="{ zero: !(d.posts + d.notes) }" :style="{ '--a': d.posts, '--b': d.notes }"><i class="a" /><i class="b" /></div>
              </div>
              <span>{{ d.today ? t('studio.today.todayShort') : d.label }}</span>
            </div>
          </div>
          <div class="st-legend">
            <span><i class="sw" style="--c: var(--ink)" />{{ t('studio.today.lgPosts') }}</span>
            <span><i class="sw" style="--c: color-mix(in oklab, var(--ink) 38%, var(--well-2))" />{{ t('studio.today.lgNotes') }}</span>
          </div>
        </div>
      </section>

      <section class="st-card todo-card st-rise" style="--i: 4">
        <div class="st-sec-t">
          <h2>{{ t('studio.today.todo') }}</h2>
          <span v-if="loaded">{{ pending ? t('studio.today.todoLeft', { n: pending }) : t('studio.today.todoClear') }}</span>
        </div>
        <div class="todo">
          <div class="todo-i" :class="{ done: loaded && !drafts.length }" style="--c: var(--blue)">
            <span class="ti"><SIcon :name="loaded && !drafts.length ? 'check' : 'doc'" :size="20" /></span>
            <div class="tx">
              <b>{{ drafts.length ? t('studio.today.draftTodo', { n: drafts.length }) : t('studio.today.draftClear') }}</b>
              <small v-if="drafts.length">「{{ drafts[0].title || t('studio.untitled') }}」· {{ relTime(drafts[0].updatedAt || drafts[0].createdAt) }}</small>
              <small v-else>{{ t('studio.today.draftClearSub') }}</small>
            </div>
            <router-link v-if="drafts.length" class="st-btn g sm" :to="{ name: 'admin-posts', query: { status: 'draft' } }">
              {{ t('studio.today.review') }}
            </router-link>
          </div>

          <div class="todo-i" :class="{ done: compressDone || (quality && !compressible.length) }" style="--c: var(--yellow)">
            <span class="ti"><SIcon :name="compressDone || (quality && !compressible.length) ? 'check' : 'compress'" :size="20" /></span>
            <div class="tx">
              <template v-if="job">
                <b>{{ t('studio.today.compressing', { done: job.done, total: job.total }) }}</b>
                <div class="pbar"><i :style="{ width: `${jobPercent}%` }" /></div>
              </template>
              <template v-else-if="compressDone">
                <b>{{ t('studio.today.compressedN', { n: compressDone.n }) }}</b>
                <small>{{ t('studio.today.compressedSub', { size: formatSize(compressDone.saved) }) }}</small>
              </template>
              <template v-else-if="quality === null">
                <b>{{ t('studio.today.scanning') }}</b>
                <small>{{ t('studio.today.scanningSub') }}</small>
              </template>
              <template v-else-if="compressible.length">
                <b>{{ t('studio.today.compressTodo', { n: compressible.length }) }}</b>
                <small>{{ t('studio.today.compressSub', { size: formatSize(compressSize) }) }}</small>
              </template>
              <template v-else>
                <b>{{ t('studio.today.compressClear') }}</b>
                <small>{{ t('studio.today.compressClearSub') }}</small>
              </template>
            </div>
            <button
              v-if="quality && compressible.length && !compressDone"
              type="button"
              class="st-btn g sm"
              :disabled="!!job"
              @click="runCompress"
            >{{ job ? t('studio.today.compressBusy') : t('studio.today.compress') }}</button>
          </div>

          <div class="todo-i" :class="{ done: pendingComments === 0 }" style="--c: var(--ink)">
            <span class="ti"><SIcon :name="pendingComments === 0 ? 'check' : 'message'" :size="20" /></span>
            <div class="tx">
              <template v-if="pendingComments">
                <b>{{ t('studio.today.commentTodoN', { n: pendingComments }) }}</b>
                <small>{{ t('studio.today.commentSubN') }}</small>
              </template>
              <template v-else>
                <b>{{ t('studio.today.commentClear') }}</b>
                <small>{{ t('studio.today.commentClearSub') }}</small>
              </template>
            </div>
            <router-link v-if="pendingComments" class="st-btn g sm" :to="{ name: 'admin-comments' }">{{ t('studio.today.review') }}</router-link>
          </div>
        </div>
      </section>
    </div>

    <div class="st-sec-t shelf-t st-rise" style="--i: 5">
      <h2>{{ shelf.title }}</h2>
      <router-link class="st-link" :to="{ name: 'admin-posts', query: shelf.draft ? { status: 'draft' } : {} }">
        {{ shelf.draft ? t('studio.today.allDrafts') : t('studio.today.allPosts') }}<SIcon name="arrowR" :size="16" />
      </router-link>
    </div>
    <div class="drafts">
      <button
        v-for="(p, i) in shelf.items"
        :key="p.id"
        type="button"
        class="draft st-rise"
        :style="{ '--i': 6 + i }"
        @click="edit(p)"
      >
        <LightCover class="dcv" :src="p.covers[0] ? thumbOf(p.covers[0]) : ''" :seed="p.slug" />
        <span class="go"><SIcon name="arrowR" :size="16" /></span>
        <div class="bd">
          <h3>{{ p.title || t('studio.untitled') }}</h3>
          <div class="meta">
            <span class="st-badge" :class="`st-${p.status}`"><i class="st-dot" />{{ t(`studio.status.${p.status}`) }}</span>
            <span>{{ t('studio.today.edited', { when: relTime(p.updatedAt || p.createdAt) }) }}</span>
          </div>
          <div v-if="p.tags.length" class="tags">
            <span v-for="tag in p.tags.slice(0, 3)" :key="tag" class="tag">{{ tag }}</span>
          </div>
        </div>
      </button>
      <router-link :to="{ name: 'admin-write-post' }" class="draft new st-rise" :style="{ '--i': 6 + shelf.items.length }">
        <span class="pl"><SIcon name="plus" :size="20" /></span>{{ t('studio.today.newOne') }}
      </router-link>
    </div>
  </section>
</template>

<style scoped lang="scss">
.view {
  position: relative;
  max-width: 1280px;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

/* 问候：标题 + 一句话（左）｜ 日期与节气（右），压缩成一行高度 */
.hello {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 22px;

  .hl { min-width: 0; }

  h1 {
    font: 700 30px/1.25 var(--font-serif);
    margin: 0 0 6px;
    letter-spacing: 0.01em;
  }

  p {
    margin: 0;
    color: var(--st-ink-2);
    font: 400 15px/1.6 var(--font-serif);
  }

  .date {
    flex: none;
    display: flex;
    align-items: center;
    gap: 10px;
    padding-bottom: 2px;
    font-size: 13.5px;
    color: var(--st-ink-3);
    letter-spacing: 0.04em;
  }

  .term { color: var(--ink); font-weight: 500; }
}

.kpis {
  margin-bottom: 20px;

  .st-stat { display: block; }
  .st-stat:hover b { color: var(--ink); }
}

.cmp-row { margin-bottom: 20px; }

/* 本周（7）｜ 待办（5）并排等高 */
.row {
  display: grid;
  grid-template-columns: minmax(0, 7fr) minmax(0, 5fr);
  align-items: stretch;
  gap: 20px;
}

.week {
  .best { font-family: var(--font-serif); font-size: 26px; }
  b.up { color: color-mix(in oklab, var(--green) 70%, var(--st-ink)); }
  b.down { color: color-mix(in oklab, var(--red) 70%, var(--st-ink)); }
  .chart { display: flex; flex-direction: column; gap: 12px; margin-top: auto; }
  .st-legend { justify-content: flex-end; }
  .v { font-style: normal; }
}

.todo-card .todo { display: flex; flex-direction: column; flex: 1; }

.todo-i {
  display: flex;
  flex: 1;
  gap: 14px;
  align-items: center;
  padding: 14px 0;
  border-bottom: 1px solid var(--line);

  &:first-child { padding-top: 0; }
  &:last-child { border-bottom: 0; padding-bottom: 0; }

  .ti {
    width: 44px;
    height: 44px;
    border-radius: var(--r-md);
    display: grid;
    place-items: center;
    flex: none;
    background: color-mix(in oklab, var(--c) 12%, var(--paper));
    color: color-mix(in oklab, var(--c) 75%, var(--st-ink));
    transition: background var(--dur), color var(--dur);
  }

  .tx { min-width: 0; flex: 1; }

  b { display: block; font-size: 15px; font-weight: 600; line-height: 1.45; }

  small {
    display: block;
    margin-top: 2px;
    font-size: 13px;
    color: var(--st-ink-3);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .st-btn { flex: none; }

  .pbar {
    height: 4px;
    max-width: 180px;
    margin-top: 8px;
    overflow: hidden;
    border-radius: calc(var(--r-xs) / 2);
    background: var(--well-2);

    i { display: block; height: 100%; background: var(--ink); border-radius: calc(var(--r-xs) / 2); transition: width 0.3s linear; }
  }

  &.done .ti {
    background: color-mix(in oklab, var(--green) 14%, var(--paper));
    color: color-mix(in oklab, var(--green) 70%, var(--st-ink));
  }

  &.muted b { color: var(--st-ink-2); }
}

.shelf-t { margin-top: 36px; }

/* 继续写作：封面约 2:1，标题 / 状态 / 时间 / 标签在封面下同一信息块 */
.drafts {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.draft {
  position: relative;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: var(--r-lg);
  text-align: left;
  background: var(--paper);
  box-shadow: 0 0 0 1px var(--line-2), 0 1px 2px var(--line);
  transition: transform var(--dur) var(--ease-spring), box-shadow var(--dur) var(--ease-out);

  &:hover {
    transform: translateY(-3px);
    box-shadow: 0 0 0 1px var(--line-2), var(--sh-card-hover);
  }

  &:active { transform: translateY(-1px) scale(0.985); }

  .dcv {
    aspect-ratio: 2 / 1;

    &::after {
      content: '';
      position: absolute;
      inset: 0;
      z-index: 2;
      /* 封面掠光：属于封面光影（品牌），非控件发光 */
      background: linear-gradient(105deg, transparent 30%, rgba(255, 255, 255, 0.35) 48%, transparent 60%);
      transform: translateX(-100%);
      transition: transform 0.9s var(--ease-out);
    }
  }

  &:hover .dcv::after { transform: translateX(100%); }

  .bd {
    display: flex;
    flex-direction: column;
    flex: 1;
    gap: 10px;
    padding: 14px 16px 16px;
  }

  h3 {
    font: 700 17px/1.45 var(--font-serif);
    margin: 0;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .meta {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: auto;
    font-size: 13px;
    color: var(--st-ink-3);
    white-space: nowrap;
  }

  .tags {
    display: flex;
    gap: 10px;
    overflow: hidden;
    font-size: 13px;
  }

  .go {
    position: absolute;
    right: 12px;
    top: 12px;
    z-index: 3;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: rgba(255, 255, 255, 0.92);
    color: #1e1c19;
    opacity: 0;
    transform: translateX(-6px) scale(0.8);
    transition: all var(--dur) var(--ease-spring);
  }

  &:hover .go { opacity: 1; transform: none; }

  &.new {
    align-items: center;
    justify-content: center;
    gap: 12px;
    min-height: 200px;
    color: var(--st-ink-3);
    background: none;
    box-shadow: 0 0 0 1.5px var(--line-2) inset;
    font-size: 15px;

    &:hover { color: var(--st-ink); box-shadow: 0 0 0 1.5px var(--line-3) inset, var(--sh-card-hover); }

    .pl {
      width: 48px;
      height: 48px;
      border-radius: 50%;
      display: grid;
      place-items: center;
      background: var(--well);
      transition: all var(--dur) var(--ease-spring);
    }

    &:hover .pl { background: var(--solid); color: var(--on-solid); box-shadow: var(--btn-shadow); transform: rotate(90deg); }
  }
}

:root[data-mode='dark'] .draft:not(.new) { background: color-mix(in oklab, var(--paper) 55%, var(--well)); }

.tag {
  color: var(--st-ink-2);
  white-space: nowrap;

  &::before { content: '#'; color: var(--st-ink-4); margin-right: 2px; }
}

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
  .hello { flex-direction: column; align-items: flex-start; gap: 8px; }
  .kpis { --n: 3 !important; }
  .kpis .st-stat:nth-child(4) { box-shadow: 0 -1px 0 var(--line-2); }
  .kpis .st-stat:nth-child(n + 5) { box-shadow: -1px 0 0 var(--line-2), 0 -1px 0 var(--line-2); }
  .row { grid-template-columns: 1fr; }
  .drafts { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
</style>
