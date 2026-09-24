<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, api, thumbOf, type AdminPost, type CompressJob, type Note, type QualityItem } from '../../api';
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

/** 「今天」：问候是主角，随想输入其次，其余是配角 */
const { t } = useI18n();
const router = useRouter();
const config = useConfigStore();

const now = new Date();
const term = solarTerm(now);
const posts = ref<AdminPost[]>([]);
const notes = ref<Note[]>([]);
const quality = ref<QualityItem[] | null>(null);
const loaded = ref(false);
const composer = ref<InstanceType<typeof NoteComposer> | null>(null);

async function load(): Promise<void> {
  const from = new Date(now);
  from.setDate(from.getDate() - 15);
  const [p, n] = await Promise.all([
    adminApi.posts().catch(() => [] as AdminPost[]),
    api.notes({ pageSize: 50, from: ymd(from) }).then((r) => r.items).catch(() => [] as Note[]),
  ]);
  posts.value = p;
  notes.value = n;
  loaded.value = true;
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

/* ===== 继续写作 ===== */
const shelf = computed(() => {
  if (drafts.value.length) return { title: t('studio.today.continue'), items: drafts.value.slice(0, 2), draft: true };
  return { title: t('studio.today.recent'), items: posts.value.slice(0, 2), draft: false };
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
    <div class="hello">
      <div class="date st-rise" style="--i: 0">
        <span class="mono">{{ dateLine }}</span>
        <span>{{ WEEKDAYS[now.getDay()] }}</span>
        <span class="term">{{ termText }}</span>
      </div>
      <h1 class="st-rise" style="--i: 1">{{ greeting(now.getHours()) }}，{{ name }}。</h1>
      <p class="st-rise" style="--i: 2">
        <template v-if="siteDays">{{ t('studio.today.dayPre') }}<em>{{ siteDays }}</em>{{ t('studio.today.dayPost') }}</template>{{ lede }}
      </p>
    </div>

    <div class="t-grid">
      <div class="main">
        <div class="st-rise" style="--i: 3">
          <NoteComposer ref="composer" hint @published="onPublished" />
        </div>

        <div class="st-sec-t st-rise shelf-t" style="--i: 4">
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
            :style="{ '--i': 5 + i }"
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
          <router-link :to="{ name: 'admin-write-post' }" class="draft new st-rise" style="--i: 7">
            <span class="pl"><SIcon name="plus" /></span>{{ t('studio.today.newOne') }}
          </router-link>
        </div>
      </div>

      <aside class="st-rise" style="--i: 6">
        <div class="st-sec-t">
          <h2>{{ t('studio.today.todo') }}</h2>
          <span v-if="loaded">{{ pending ? t('studio.today.todoLeft', { n: pending }) : t('studio.today.todoClear') }}</span>
        </div>
        <div class="todo">
          <div class="todo-i" :class="{ done: loaded && !drafts.length }" style="--c: var(--blue)">
            <span class="ti"><SIcon :name="loaded && !drafts.length ? 'check' : 'doc'" /></span>
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
            <span class="ti"><SIcon :name="compressDone || (quality && !compressible.length) ? 'check' : 'compress'" /></span>
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

          <div class="todo-i muted" style="--c: var(--primary)">
            <span class="ti"><SIcon name="message" /></span>
            <div class="tx">
              <b>{{ t('studio.today.commentTodo') }}</b>
              <small>{{ t('studio.today.commentSub') }}</small>
            </div>
            <router-link class="st-btn g sm" :to="{ name: 'admin-comments' }">{{ t('studio.view') }}</router-link>
          </div>
        </div>

        <div class="week">
          <p class="say">
            <template v-if="week.posts && week.notes">
              {{ t('studio.today.weekPre') }}<b>{{ week.posts }}</b>{{ t('studio.today.weekMid') }}<b>{{ week.notes }}</b>{{ t('studio.today.weekPost') }}
            </template>
            <template v-else-if="week.posts">{{ t('studio.today.weekPre') }}<b>{{ week.posts }}</b>{{ t('studio.today.weekPostsOnly') }}</template>
            <template v-else-if="week.notes">{{ t('studio.today.weekPre') }}<b>{{ week.notes }}</b>{{ t('studio.today.weekNotesOnly') }}</template>
            <template v-else>{{ t('studio.today.weekNone') }}</template>
          </p>
          <p class="sub">{{ weekSub }}</p>
          <div class="bars">
            <div
              v-for="(d, i) in week.days"
              :key="d.key"
              class="b"
              :class="{ today: d.today, fut: d.future }"
              :style="{ '--i': i, '--h': `${d.posts + d.notes ? Math.max(8, Math.round(((d.posts + d.notes) / week.max) * 56)) : 4}px` }"
            >
              <em v-if="!d.future">{{ t('studio.today.barTip', { p: d.posts, n: d.notes }) }}</em>
              <i />
              <span>{{ d.today ? t('studio.today.todayShort') : d.label }}</span>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </section>
</template>

<style scoped lang="scss">
.view {
  position: relative;
  max-width: 1120px;
  margin: 0 auto;
  padding: 52px 64px 96px;
}

.hello {
  max-width: 760px;

  .date {
    font-size: 13px;
    color: var(--ink-3);
    letter-spacing: 0.08em;
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .term { color: var(--primary-ink); font-weight: 500; }

  h1 {
    font: 600 38px/1.3 var(--font-serif);
    margin: 14px 0 10px;
    letter-spacing: 0.01em;
  }

  p {
    margin: 0;
    color: var(--ink-2);
    font: 400 16px/1.8 var(--font-serif);
  }

  em {
    font-style: normal;
    color: var(--ink);
    border-bottom: 1.5px solid var(--primary-ring);
    margin: 0 2px;
  }
}

.t-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 336px;
  gap: 56px;
  margin-top: 40px;
}

.shelf-t { margin-top: 52px; }

.drafts {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.draft {
  position: relative;
  display: flex;
  flex-direction: column;
  border-radius: 16px;
  padding: 8px;
  text-align: left;
  transition: transform var(--dur) var(--ease-spring), box-shadow var(--dur) var(--ease-out), background var(--dur-fast);

  &:hover {
    transform: translateY(-4px);
    box-shadow: var(--sh-card-hover);
    background: var(--paper);
  }

  &:active { transform: translateY(-2px) scale(0.98); }

  .dcv {
    height: 108px;
    border-radius: 11px;

    &::after {
      content: '';
      position: absolute;
      inset: 0;
      z-index: 2;
      background: linear-gradient(105deg, transparent 30%, rgba(255, 255, 255, 0.35) 48%, transparent 60%);
      transform: translateX(-100%);
      transition: transform 0.9s var(--ease-out);
    }
  }

  &:hover .dcv::after { transform: translateX(100%); }

  .bd { padding: 14px 6px 6px; }

  h3 {
    font: 600 16px/1.5 var(--font-serif);
    margin: 0 0 8px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    min-height: 48px;
  }

  .meta {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12.5px;
    color: var(--ink-3);
  }

  .tags {
    display: flex;
    gap: 10px;
    margin-top: 10px;
    font-size: 12.5px;
  }

  .go {
    position: absolute;
    right: 16px;
    top: 16px;
    z-index: 3;
    width: 30px;
    height: 30px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: rgba(255, 255, 255, 0.9);
    color: #1e1c19;
    opacity: 0;
    transform: translateX(-6px) scale(0.8);
    transition: all var(--dur) var(--ease-spring);
  }

  &:hover .go { opacity: 1; transform: none; }

  &.new {
    align-items: center;
    justify-content: center;
    gap: 10px;
    color: var(--ink-3);
    box-shadow: 0 0 0 1.5px var(--line-2) inset;
    min-height: 230px;
    font-size: 14px;

    &:hover { color: var(--primary-ink); box-shadow: 0 0 0 1.5px var(--primary) inset, var(--sh-card-hover); }

    .pl {
      width: 44px;
      height: 44px;
      border-radius: 50%;
      display: grid;
      place-items: center;
      background: var(--well);
      transition: all var(--dur) var(--ease-spring);
    }

    &:hover .pl { background: var(--primary); color: var(--on-primary); transform: rotate(90deg); }
  }
}

.tag {
  color: var(--ink-2);

  &::before { content: '#'; color: var(--ink-4); margin-right: 2px; }
}

.todo { display: flex; flex-direction: column; }

.todo-i {
  display: flex;
  gap: 14px;
  align-items: center;
  padding: 14px 0;
  border-bottom: 1px solid var(--line);

  &:last-child { border-bottom: 0; }

  .ti {
    width: 38px;
    height: 38px;
    border-radius: 12px;
    display: grid;
    place-items: center;
    flex: none;
    background: color-mix(in oklab, var(--c) 12%, var(--paper));
    color: color-mix(in oklab, var(--c) 75%, var(--ink));
    transition: background var(--dur), color var(--dur);
  }

  .tx { min-width: 0; flex: 1; }

  b { display: block; font-size: 14.5px; font-weight: 500; line-height: 1.4; }

  small {
    display: block;
    font-size: 12.5px;
    color: var(--ink-3);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .st-btn { flex: none; }

  .pbar {
    height: 4px;
    border-radius: 2px;
    background: var(--well-2);
    overflow: hidden;
    margin-top: 8px;
    max-width: 160px;

    i { display: block; height: 100%; background: var(--primary); border-radius: 2px; transition: width 0.3s linear; }
  }

  &.done .ti {
    background: color-mix(in oklab, var(--green) 14%, var(--paper));
    color: color-mix(in oklab, var(--green) 70%, var(--ink));
  }

  &.muted b { color: var(--ink-2); }
}

.week {
  margin-top: 28px;
  padding: 22px;
  border-radius: 18px;
  background: var(--well);

  .say {
    font: 500 17px/1.7 var(--font-serif);
    margin: 0;

    b { font-weight: 700; color: var(--primary-ink); margin: 0 3px; }
  }

  .sub { font-size: 13px; color: var(--ink-3); margin: 6px 0 18px; }
}

.bars {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 10px;
  align-items: end;
  height: 84px;

  .b {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    height: 100%;
    justify-content: flex-end;

    i {
      display: block;
      width: 100%;
      max-width: 26px;
      border-radius: 6px;
      background: var(--ink-4);
      height: var(--h);
      transform-origin: bottom;
      animation: bar-up 0.9s var(--ease-spring) both;
      animation-delay: calc(var(--i) * 60ms + 0.3s);
      opacity: 0.55;
      transition: opacity var(--dur-fast);
    }

    span { font-size: 11.5px; color: var(--ink-3); }

    em {
      position: absolute;
      bottom: calc(var(--h) + 30px);
      font: 500 11px var(--font-mono);
      font-style: normal;
      color: var(--ink);
      background: var(--paper);
      padding: 3px 7px;
      border-radius: 6px;
      box-shadow: var(--sh-pop);
      opacity: 0;
      transform: translateY(4px);
      transition: all var(--dur-fast) var(--ease-out);
      white-space: nowrap;
      pointer-events: none;
    }

    &:hover i { opacity: 1; }
    &:hover em { opacity: 1; transform: none; }

    &.today {
      i { background: var(--primary); opacity: 1; }
      span { color: var(--primary-ink); font-weight: 500; }
    }

    &.fut i { height: 4px; background: none; box-shadow: 0 0 0 1px var(--line-3) inset; opacity: 1; }
  }
}

@keyframes bar-up { from { transform: scaleY(0); } }

@media (max-width: 1180px) {
  .view { padding: 40px 36px 80px; }
  .t-grid { grid-template-columns: 1fr; }
  .drafts { grid-template-columns: repeat(2, 1fr); }
}
</style>
