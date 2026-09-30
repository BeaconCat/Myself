<script setup lang="ts">
/**
 * 后台 · 概览：问候 + 本周发布柱状图与计数统计条（合并一卡）+ 待办 + 继续写 + 最近动静。
 * 全部来自真实数据（文章 / 随想 / 素材 / 备份 / 图片质量扫描），支持下拉刷新。
 */
import { computed, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, thumbOf, type BackupInfo, type QualityItem } from '../../../api';
import { useConfigStore } from '../../../stores/config';
import MaPage from '../../../components/mobile-admin/MaPage.vue';
import MaIcon from '../../../components/mobile-admin/MaIcon.vue';
import MaSkeleton from '../../../components/mobile-admin/MaSkeleton.vue';
import type { IconName } from '../../../components/mobile-admin/icons';
import { cache, loadMedia, loadNotes, loadPosts, openComposer, shell, toast } from '../../../components/mobile-admin/state';
import { formatSize, mdPlain, parseTime, relTime } from '../../../components/mobile-admin/format';

const { t } = useI18n();
const router = useRouter();
const config = useConfigStore();

const backups = ref<BackupInfo[] | null>(null);
const quality = ref<QualityItem[] | null>(null);
const ready = ref(!!(cache.posts && cache.notes));
const grown = ref(false);

async function load(): Promise<void> {
  await Promise.allSettled([
    loadPosts(),
    loadNotes(),
    loadMedia(),
    adminApi.backups().then((b) => (backups.value = b)),
    adminApi.qualityScan().then((q) => (quality.value = q)),
  ]);
  ready.value = true;
  requestAnimationFrame(() => requestAnimationFrame(() => (grown.value = true)));
}
onMounted(load);
watch(() => [shell.bump.posts, shell.bump.notes, shell.bump.media], () => void load());

function refresh(done: () => void): void {
  grown.value = false;
  void load().finally(done);
}

/* ---------- 问候 ---------- */
const greeting = computed(() => {
  const h = new Date().getHours();
  const key = h < 5 ? 'night' : h < 11 ? 'morning' : h < 13 ? 'noon' : h < 18 ? 'afternoon' : 'evening';
  return t(`mobileAdmin.today.greet.${key}`, { name: config.cfg.about.name || 'Myself' });
});

/* ---------- 本周柱状：近 7 天每天发布（文章 + 随想） ---------- */
const DAY = 86400000;
function dayStart(d: Date): number {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime();
}
const week = computed(() => {
  const today = dayStart(new Date());
  const days = Array.from({ length: 7 }, (_, k) => ({ start: today - (6 - k) * DAY, posts: 0, notes: 0 }));
  let last = 0;
  const bucket = (s: string, kind: 'posts' | 'notes'): void => {
    const ts = dayStart(parseTime(s));
    const idx = Math.round((ts - days[0].start) / DAY);
    if (idx >= 0 && idx < 7) days[idx][kind] += 1;
    else if (idx >= -7 && idx < 0) last += 1;
  };
  cache.posts?.forEach((p) => bucket(p.createdAt, 'posts'));
  cache.notes?.forEach((n) => bucket(n.createdAt, 'notes'));
  const total = days.reduce((a, d) => a + d.posts + d.notes, 0);
  const max = Math.max(1, ...days.map((d) => d.posts + d.notes));
  const labels = t('mobileAdmin.today.weekdays').split(',');
  return {
    total,
    delta: total - last,
    days: days.map((d, k) => ({
      ...d,
      sum: d.posts + d.notes,
      h: (d.posts + d.notes) / max,
      label: k === 6 ? t('mobileAdmin.today.todayShort') : labels[new Date(d.start).getDay()],
    })),
  };
});

/** 最近一次发布（本周为空时在图表里提示） */
const lastPublished = computed(() => {
  const times = [...(cache.posts ?? []).map((p) => p.createdAt), ...(cache.notes ?? []).map((n) => n.createdAt)];
  if (!times.length) return '';
  const latest = times.reduce((a, b) => (parseTime(a).getTime() > parseTime(b).getTime() ? a : b));
  return relTime(latest, t);
});

const counts = computed(() => ({
  posts: cache.posts?.length ?? 0,
  notes: cache.notesTotal,
  media: cache.media?.length ?? 0,
}));

/* ---------- 待办 ---------- */
interface Todo {
  id: string;
  icon: IconName;
  color: string;
  text: string;
  sub: string;
  go: () => void;
}
const drafts = computed(() => (cache.posts ?? []).filter((p) => p.status === 'draft'));
const todos = computed<Todo[]>(() => {
  const list: Todo[] = [];
  if (drafts.value.length) {
    list.push({
      id: 'drafts',
      icon: 'pen',
      color: '#ff7a1a',
      text: t('mobileAdmin.today.todoDrafts', { n: drafts.value.length }),
      sub: t('mobileAdmin.today.todoDraftsSub'),
      go: () => void router.push({ name: 'admin-posts', query: { filter: 'draft' } }),
    });
  }
  const heavy = (quality.value ?? []).filter((q) => q.compressible && q.format !== 'webp');
  if (heavy.length) {
    const size = heavy.reduce((a, q) => a + q.size, 0);
    list.push({
      id: 'compress',
      icon: 'image',
      color: '#12b76a',
      text: t('mobileAdmin.today.todoCompress', { n: heavy.length }),
      sub: t('mobileAdmin.today.todoCompressSub', { size: formatSize(size) }),
      go: () => void router.push({ name: 'admin-media' }),
    });
  }
  if (backups.value) {
    const lastB = backups.value[0];
    const age = lastB ? Math.floor((Date.now() - parseTime(lastB.createdAt).getTime()) / DAY) : Infinity;
    if (age >= 7) {
      list.push({
        id: 'backup',
        icon: 'archive',
        color: '#5b6b86',
        text: lastB ? t('mobileAdmin.today.todoBackup', { n: age }) : t('mobileAdmin.today.todoBackupNever'),
        sub: t('mobileAdmin.today.todoBackupSub'),
        go: () => void router.push({ name: 'admin-settings', query: { focus: 'backup' } }),
      });
    }
  }
  return list;
});

/* ---------- 继续写 ---------- */
const lastDraft = computed(() =>
  [...drafts.value].sort((a, b) => parseTime(b.updatedAt).getTime() - parseTime(a.updatedAt).getTime())[0],
);

/* ---------- 最近动静 ---------- */
interface Activity {
  key: string;
  icon: IconName;
  html: string;
  sub: string;
  at: number;
  thumb?: string;
}
const esc = (s: string): string => s.replace(/[&<>"]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[c] ?? c);
const activity = computed<Activity[]>(() => {
  const items: Activity[] = [];
  cache.posts?.slice(0, 6).forEach((p) => {
    const updated = p.updatedAt !== p.createdAt;
    items.push({
      key: `p${p.id}`,
      icon: p.status === 'draft' ? 'pen' : 'docs',
      html: t(
        p.status === 'draft' ? 'mobileAdmin.today.actDraft' : updated ? 'mobileAdmin.today.actPostUpdated' : 'mobileAdmin.today.actPost',
        { title: `<b>${esc(p.title)}</b>` },
      ),
      sub: relTime(p.updatedAt, t),
      at: parseTime(p.updatedAt).getTime(),
      thumb: p.covers[0] ? thumbOf(p.covers[0]) : undefined,
    });
  });
  cache.notes?.slice(0, 6).forEach((n) => {
    items.push({
      key: `n${n.id}`,
      icon: 'bubble',
      html: t('mobileAdmin.today.actNote', { text: `<b>${esc(mdPlain(n.contentMd).slice(0, 18))}</b>` }),
      sub: `${relTime(n.createdAt, t)}${n.images.length ? ` · ${t('mobileAdmin.content.images', { n: n.images.length })}` : ''}`,
      at: parseTime(n.createdAt).getTime(),
    });
  });
  const media = cache.media ?? [];
  if (media.length) {
    const newest = parseTime(media[0].createdAt).getTime();
    const batch = media.filter((m) => newest - parseTime(m.createdAt).getTime() < 10 * 60000);
    items.push({
      key: 'media',
      icon: 'image',
      html: t('mobileAdmin.today.actMedia', { n: `<b>${batch.length}</b>` }),
      sub: `${relTime(media[0].createdAt, t)} · ${formatSize(batch.reduce((a, m) => a + m.size, 0))}`,
      at: newest,
    });
  }
  const b = backups.value?.[0];
  if (b) {
    items.push({
      key: 'backup',
      icon: 'archive',
      html: t('mobileAdmin.today.actBackup'),
      sub: `${relTime(b.createdAt, t)} · ${formatSize(b.size)}`,
      at: parseTime(b.createdAt).getTime(),
    });
  }
  return items.sort((a, c) => c.at - a.at).slice(0, 6);
});

function openSite(): void {
  window.open('/', '_blank', 'noopener');
}
function syncNow(): void {
  void load().then(() => toast(t('mobileAdmin.today.synced'), t('mobileAdmin.time.now')));
}
</script>

<template>
  <MaPage :title="t('mobileAdmin.today.title')" :eyebrow="greeting" @refresh="refresh">
    <template #left>
      <button class="av-btn tap" :aria-label="t('mobileAdmin.today.viewSite')" @click="openSite">
        <img :src="config.cfg.about.avatar || config.cfg.site.logo || '/favicon-256.png'" alt="" draggable="false" />
      </button>
    </template>
    <template #right>
      <button class="icbtn tap" :aria-label="t('mobileAdmin.today.sync')" @click="syncNow"><MaIcon name="bolt" :size="20" /></button>
    </template>

    <MaSkeleton v-if="!ready" variant="hero" />
    <template v-else>
      <!-- 高密度：本周柱状图与计数统计条合并为一张信息完整的大卡 -->
      <section class="ov-hero">
        <header class="ov-head">
          <small>{{ t('mobileAdmin.today.weekTitle') }}</small>
          <div class="legend">
            <span style="--c: var(--primary)">{{ t('mobileAdmin.today.legendPosts') }}</span>
            <span style="--c: var(--bar-2)">{{ t('mobileAdmin.today.legendNotes') }}</span>
          </div>
        </header>
        <div class="ov-num">
          <b>{{ week.total }}</b>
          <em v-if="week.delta !== 0" :class="{ down: week.delta < 0 }">{{ week.delta > 0 ? '+' : '' }}{{ week.delta }}</em>
          <span class="unit">{{ t('mobileAdmin.today.weekUnit') }}</span>
        </div>
        <div class="bars-box">
          <p v-if="!week.total" class="bars-empty">
            {{ lastPublished ? t('mobileAdmin.today.weekEmptySince', { time: lastPublished }) : t('mobileAdmin.today.weekEmpty') }}
          </p>
          <div class="bars" :class="{ grown, quiet: !week.total }">
            <div v-for="(d, k) in week.days" :key="d.start" class="bar" :class="{ today: k === 6, empty: !d.sum }">
              <span class="val">{{ d.sum || '' }}</span>
              <i :style="{ '--h': Math.max(0.06, d.h), '--k': k, '--p': d.sum ? d.posts / d.sum : 0 }" />
              <small>{{ d.label }}</small>
            </div>
          </div>
        </div>
        <div class="ov-stats">
          <button class="tap" @click="router.push({ name: 'admin-posts' })"><b>{{ counts.posts }}</b><small>{{ t('mobileAdmin.today.posts') }}</small></button>
          <button class="tap" @click="router.push({ name: 'admin-notes' })"><b>{{ counts.notes }}</b><small>{{ t('mobileAdmin.today.notes') }}</small></button>
          <button class="tap" @click="router.push({ name: 'admin-media' })"><b>{{ counts.media }}</b><small>{{ t('mobileAdmin.today.media') }}</small></button>
          <button class="tap" @click="router.push({ name: 'admin-posts', query: { filter: 'draft' } })"><b>{{ drafts.length }}</b><small>{{ t('mobileAdmin.content.draft') }}</small></button>
        </div>
      </section>

      <section>
        <div class="sec-h"><h2>{{ t('mobileAdmin.today.todo') }}</h2><span v-if="todos.length" class="badge">{{ todos.length }}</span></div>
        <div v-if="todos.length" class="ma-list todo-list">
          <button v-for="td in todos" :key="td.id" class="ma-li tap" @click="td.go">
            <span class="lic" :style="{ '--c': td.color }"><MaIcon :name="td.icon" :size="17" /></span>
            <span class="lt"><span class="tt">{{ td.text }}</span><small>{{ td.sub }}</small></span>
            <MaIcon name="chev" :size="16" class="chev" />
          </button>
        </div>
        <div v-else class="all-clear">
          <span class="ring"><MaIcon name="check" :size="20" /></span>
          <div><b>{{ t('mobileAdmin.today.clear') }}</b><small>{{ t('mobileAdmin.today.clearSub') }}</small></div>
        </div>
      </section>

      <section>
        <div class="sec-h"><h2>{{ t('mobileAdmin.today.continue') }}</h2></div>
        <button
          v-if="lastDraft"
          class="draft tap"
          @click="router.push({ name: 'admin-write-post', query: { id: String(lastDraft.id) } })"
        >
          <div class="dt">
            <small>{{ t('mobileAdmin.today.draftAt', { time: relTime(lastDraft.updatedAt, t) }) }}</small>
            <b>{{ lastDraft.title }}</b>
          </div>
          <span class="go"><MaIcon name="pen" :size="18" /></span>
        </button>
        <div v-else class="draft-pair">
          <button class="draft tap" @click="router.push({ name: 'admin-write-post' })">
            <div class="dt"><small>{{ t('mobileAdmin.today.noDraft') }}</small><b>{{ t('mobileAdmin.create.post') }}</b></div>
            <span class="go"><MaIcon name="pen" :size="18" /></span>
          </button>
          <button class="draft alt tap" @click="openComposer(null)">
            <div class="dt"><small>{{ t('mobileAdmin.create.noteSub') }}</small><b>{{ t('mobileAdmin.create.note') }}</b></div>
            <span class="go"><MaIcon name="bubble" :size="18" /></span>
          </button>
        </div>
      </section>

      <section>
        <div class="sec-h"><h2>{{ t('mobileAdmin.today.activity') }}</h2></div>
        <div class="act-list">
          <div v-for="(a, k) in activity" :key="a.key" class="act" :style="{ '--k': k }">
            <span class="ai"><MaIcon :name="a.icon" :size="18" /></span>
            <div class="at">
              <!-- 文案来自 i18n 字典，插值已转义 -->
              <!-- eslint-disable-next-line vue/no-v-html -->
              <p v-html="a.html" />
              <small>{{ a.sub }}</small>
            </div>
            <img v-if="a.thumb" class="athumb" :src="a.thumb" alt="" />
          </div>
          <p v-if="!activity.length" class="empty">{{ t('mobileAdmin.today.noActivity') }}</p>
        </div>
      </section>
    </template>
  </MaPage>
</template>

<style scoped lang="scss">
.av-btn {
  position: relative;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  padding: 0;

  img {
    width: 100%;
    height: 100%;
    border-radius: 50%;
    object-fit: cover;
    box-shadow: 0 0 0 0.5px rgba(255, 255, 255, 0.14);
  }

  &::after {
    content: '';
    position: absolute;
    right: -1px;
    bottom: 1px;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: #2bd46a;
    box-shadow: 0 0 0 2px var(--bg);
  }
}

.ov-hero {
  /* 图表第二系列：主色掺底，数据色而非发光 */
  --bar-2: color-mix(in oklab, var(--primary) 38%, var(--fill-2));

  margin: 2px 16px 0;
  padding: 18px;
  border-radius: var(--r-lg);
  background: var(--elev);
  box-shadow: var(--shadow-card);
  position: relative;
  overflow: hidden;
}

.ov-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;

  > small {
    font-size: 13px;
    color: var(--text-3);
  }
}

.ov-num {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-top: 2px;

  b {
    font-family: var(--font-mono);
    font-size: 44px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    letter-spacing: -0.02em;
    line-height: 1.1;
  }

  em {
    font-style: normal;
    font-size: 13px;
    font-weight: 600;
    color: var(--good);
    padding: 2px 8px;
    border-radius: 999px;
    background: color-mix(in oklab, var(--good) 13%, transparent);

    &.down {
      color: var(--text-2);
      background: var(--fill);
    }
  }

  .unit {
    font-size: 13px;
    color: var(--text-3);
  }
}

.bars-box { position: relative; }

.bars-empty {
  margin-top: 12px;
  text-align: center;
  font-size: 13px;
  color: var(--text-3);
  animation: ma-act-in 0.6s var(--ease-out) 0.3s backwards;
}

/* 柱状图撑满卡片宽度：7 列等分，柱宽随宽度放大（不设上限） */
.bars {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 8px;
  height: 128px;
  margin-top: 10px;
  align-items: end;

  /* 本周无发布：柱区收成一条基线，不留大片空白 */
  &.quiet { height: 46px; }
}

.bar {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;

  i {
    display: block;
    width: 100%;
    height: calc(var(--h) * 86px);
    border-radius: var(--r-sm);
    /* 堆叠柱：底部实色段（文章）+ 其余为第二系列（随想），两段均为纯色 */
    background: var(--bar-2) linear-gradient(var(--primary), var(--primary)) bottom / 100% calc(var(--p) * 100%) no-repeat;
    transform-origin: bottom;
    transform: scaleY(0.02);
    transition: transform 0.7s var(--ease-spring);
    transition-delay: calc(var(--k) * 45ms);
  }

  .val {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-2);
    opacity: 0;
    transition: opacity 0.4s calc(var(--k, 0) * 45ms + 0.3s);
  }

  small {
    font-size: 12px;
    color: var(--text-3);
  }

  &.empty i {
    background: var(--fill-2);
  }

  .quiet & {
    i { height: 6px; }
  }

  .quiet & .val { display: none; }

  &.today {
    small { color: var(--ink); font-weight: 600; }
  }
}

.grown .bar {
  i { transform: none; }
  .val { opacity: 1; }
}

.legend {
  display: flex;
  gap: 12px;
  font-size: 12.5px;
  color: var(--text-2);

  span::before {
    content: '';
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin-right: 5px;
    background: var(--c);
  }
}

/* 统计条：等分格 + 大号等宽数字，窄屏 2 列 */
.ov-stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin-top: 16px;
  border-radius: var(--r-md);
  background: var(--fill);
  overflow: hidden;

  button {
    min-width: 0;
    padding: 16px 16px 14px;
    text-align: left;

    &:nth-child(2) { box-shadow: -1px 0 0 var(--line); }
    &:nth-child(3) { box-shadow: 0 -1px 0 var(--line); }
    &:nth-child(4) { box-shadow: -1px 0 0 var(--line), 0 -1px 0 var(--line); }
    &:active { background: var(--fill-2); }
  }

  b {
    display: block;
    font-family: var(--font-mono);
    font-size: 30px;
    line-height: 1.1;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }

  small {
    display: block;
    margin-top: 6px;
    font-size: 12.5px;
    color: var(--text-3);
  }
}

.badge {
  margin-left: 8px;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 999px;
  display: inline-grid;
  place-items: center;
  font-size: 12px;
  font-weight: 600;
  color: #fff;
  background: var(--accent-red);
  align-self: center;
}

.todo-list {
  margin: 0 16px;

  .lt {
    display: flex;
    flex-direction: column;
    padding: 10px 0;
  }

  .tt { font-size: 15px; }
  small { font-size: 13px; margin-top: 2px; }
}

.all-clear {
  display: flex;
  align-items: center;
  gap: 14px;
  margin: 0 16px;
  padding: 14px 16px;
  border-radius: var(--r-lg);
  background: var(--elev);
  box-shadow: inset 0 0 0 0.5px var(--line);

  .ring {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    color: var(--good);
    background: color-mix(in oklab, var(--good) 14%, transparent);
  }

  b { display: block; font-size: 15px; }
  small { font-size: 12.5px; color: var(--text-3); }
}

.draft {
  display: flex;
  align-items: center;
  gap: 14px;
  width: calc(100% - 32px);
  margin: 0 16px;
  padding: 14px 16px;
  border-radius: var(--r-lg);
  text-align: left;
  background: var(--elev);
  box-shadow: inset 0 0 0 0.5px var(--line);

  .dt { min-width: 0; }

  b {
    display: block;
    font-family: var(--font-serif);
    font-size: 16px;
    margin-top: 2px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  small { font-size: 12.5px; color: var(--text-3); }

  .go {
    margin-left: auto;
    width: 38px;
    height: 38px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: var(--solid);
    color: var(--on-solid);
    flex: none;
    box-shadow: var(--btn-shadow);
  }
}

.draft-pair {
  display: grid;
  grid-template-columns: 1fr;
  margin: 0 16px;
  border-radius: var(--r-lg);
  background: var(--elev);
  box-shadow: inset 0 0 0 0.5px var(--line);
  overflow: hidden;

  /* 高密度：两个入口合并为一组列表行（图标 + 标题 + 说明同一行） */
  .draft {
    position: relative;
    width: auto;
    margin: 0;
    flex-direction: row-reverse;
    justify-content: flex-end;
    gap: 12px;
    padding: 12px 14px;
    border-radius: 0;
    background: none;
    box-shadow: none;

    & + .draft::before {
      content: '';
      position: absolute;
      top: 0;
      left: 64px;
      right: 0;
      height: 0.5px;
      background: var(--line-2);
    }
  }

  .dt {
    display: flex;
    flex-direction: column-reverse;
  }

  .go { margin-left: 0; }

  .alt .go {
    background: var(--elev-2);
    color: var(--ink);
    box-shadow: inset 0 0 0 0.5px var(--line-2);
  }
}

.act-list {
  margin: 0 16px;
}

.act {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 11px 0;
  position: relative;
  animation: ma-act-in 0.5s var(--ease-out) backwards;
  animation-delay: calc(var(--k) * 50ms);

  & + .act::before {
    content: '';
    position: absolute;
    top: 0;
    left: 50px;
    right: 0;
    height: 0.5px;
    background: var(--line);
  }

  .ai {
    width: 38px;
    height: 38px;
    border-radius: var(--r-md);
    display: grid;
    place-items: center;
    flex: none;
    background: var(--tint);
    color: var(--ink);
  }

  .at {
    flex: 1;
    min-width: 0;
  }

  p {
    font-size: 15px;
    line-height: 1.45;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;

    :deep(b) { font-weight: 600; }
  }

  small {
    font-size: 13px;
    color: var(--text-3);
  }

  .athumb {
    width: 44px;
    height: 44px;
    border-radius: var(--r-sm);
    object-fit: cover;
    flex: none;
    box-shadow: 0 0 0 0.5px var(--line);
  }
}

@keyframes ma-act-in {
  from { opacity: 0; transform: translateY(12px); }
}

.empty {
  padding: 20px 0;
  text-align: center;
  font-size: 13.5px;
  color: var(--text-3);
}
</style>
