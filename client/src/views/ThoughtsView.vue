<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Note } from '../api';
import ImageViewer, { type OriginRect } from '../components/media/ImageViewer.vue';
import NoteCard from '../components/thoughts/NoteCard.vue';
import MediaWall from '../components/thoughts/MediaWall.vue';
import DateRangePicker from '../components/thoughts/DateRangePicker.vue';
import MonthCalendar from '../components/thoughts/MonthCalendar.vue';
import ContentIcon from '../components/post/ContentIcon.vue';
import ContentToast, { showToast } from '../components/post/ContentToast.vue';
import { copyText, isoDay, isStuck } from '../components/post/content';
import { useNotesFeed, useNotesIndex, type NotesTab } from '../composables/useNotesFeed';
import { useConfigStore } from '../stores/config';

/**
 * 桌面随想：信息流 / 媒体 segmented（抬升 + 轻染，选中块在两项间 morph）+ 搜索 + 日期范围；
 * 控制条吸顶毛玻璃；右栏日历（有随想的日子一个小点，点击筛选该天）+ 心情列表（点击按心情搜索）。
 */
const { t } = useI18n();
const config = useConfigStore();

const sentinel = useTemplateRef<HTMLElement>('sentinel');
const feed = useNotesFeed(sentinel);
const { notes, tab, keyword, loading, loadingMore, hasMore, dateFrom, dateTo, loadedOnce } = feed;
const index = useNotesIndex();

const avatar = computed(() => config.cfg.about.avatar || '/favicon-256.png');
const name = computed(() => config.cfg.about.name);
const handle = computed(() => config.cfg.github.username || 'myself');

const booting = computed(() => !loadedOnce.value);
const filtered = computed(() => !!(keyword.value.trim() || dateFrom.value || dateTo.value));

/* ---------- segmented：选中块 morph（前缘 .30s ease-out，后缘 .46s spring 延迟 .05s） ---------- */
const seg = ref<HTMLElement | null>(null);
const lift = ref({ l: 0, w: 0, toL: false, instant: true });
/** 视图切换方向：媒体在右 → 从右滑入 */
const dir = ref<'from-r' | 'from-l'>('from-r');

function placeLift(instant = false): void {
  const on = seg.value?.querySelector<HTMLElement>('button.on');
  if (!on) return;
  const l = on.offsetLeft;
  lift.value = { l, w: on.offsetWidth, toL: l < lift.value.l, instant };
  if (instant) requestAnimationFrame(() => requestAnimationFrame(() => { lift.value.instant = false; }));
}

function pickTab(next: NotesTab): void {
  if (tab.value === next) return;
  dir.value = next === 'media' ? 'from-r' : 'from-l';
  feed.switchTab(next);
}

watch(tab, () => void nextTick(() => placeLift()));

/* ---------- 日期 / 日历 / 心情 ---------- */
const railMonth = ref(isoDay(new Date()).slice(0, 7));
const singleDay = computed(() => (dateFrom.value && dateFrom.value === dateTo.value ? dateFrom.value : ''));

watch(railMonth, (ym) => void index.loadMonth(ym));

/* 首屏：日历停在最近一条随想所在的月份（当月可能还没写） */
const stopInit = watch(loadedOnce, (ok) => {
  if (!ok) return;
  const latest = notes.value.reduce((m, n) => (n.createdAt > m ? n.createdAt : m), '');
  if (latest) railMonth.value = latest.slice(0, 7);
  void index.loadMonth(railMonth.value);
  stopInit();
});

function pickDay(day: string): void {
  if (singleDay.value === day) feed.clearDate();
  else feed.setRange(day, day);
}

function pickMood(mood: string): void {
  keyword.value = keyword.value.trim() === mood ? '' : mood;
  void feed.reload();
}

/* ---------- 查看器 / 复制 ---------- */
const viewer = ref<{ images: string[]; index: number; rect?: OriginRect } | null>(null);

function openViewer(images: string[], i: number, rect: DOMRect): void {
  viewer.value = { images, index: i, rect: { left: rect.left, top: rect.top, width: rect.width, height: rect.height } };
}

async function copyNote(n: Note): Promise<void> {
  const ok = await copyText(n.contentMd);
  showToast(ok ? t('content.thoughts.textCopied') : t('content.article.copyFailed'));
}

/* ---------- 吸顶 ---------- */
const NAV_H = 64;
const ctl = ref<HTMLElement | null>(null);
const stuck = ref(false);
function onScroll(): void {
  stuck.value = isStuck(ctl.value, NAV_H);
}

function onResize(): void {
  placeLift(true);
}

onMounted(() => {
  window.addEventListener('scroll', onScroll, { passive: true });
  window.addEventListener('resize', onResize);
  void index.loadMoods();
  void nextTick(() => placeLift(true));
  /* 字体到位后按钮宽度会变，再校准一次 */
  void document.fonts?.ready.then(() => placeLift(true));
});

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll);
  window.removeEventListener('resize', onResize);
});
</script>

<template>
  <main class="thoughts">
    <div class="wrap">
      <header class="ph rise-stagger">
        <h1>{{ t('nav.thoughts') }}</h1>
        <p>{{ config.cfg.thoughts.subtitle }}</p>
      </header>

      <div class="th-grid">
        <div class="feed">
          <!-- 控制条：segmented + 搜索 + 日期（吸顶毛玻璃） -->
          <div ref="ctl" class="th-ctl rise" style="--i: 2" :class="{ stuck }">
            <div
              ref="seg"
              class="seg"
              :class="{ 'to-l': lift.toL, 'no-anim': lift.instant }"
              :style="{ '--l': `${lift.l}px`, '--w': `${lift.w}px` }"
              role="tablist"
            >
              <span class="lift" aria-hidden="true" />
              <button type="button" role="tab" :aria-selected="tab === 'posts'" :class="{ on: tab === 'posts' }" @click="pickTab('posts')">
                <ContentIcon name="bubble" size="s" />{{ t('content.thoughts.feed') }}
              </button>
              <button type="button" role="tab" :aria-selected="tab === 'media'" :class="{ on: tab === 'media' }" @click="pickTab('media')">
                <ContentIcon name="image" size="s" />{{ t('content.thoughts.media') }}
              </button>
            </div>
            <label class="field">
              <ContentIcon name="search" size="s" />
              <input v-model="keyword" type="search" :placeholder="t('thoughts.searchPlaceholder')" @input="feed.onSearch" />
              <button v-if="keyword" type="button" class="clr" :aria-label="t('content.thoughts.clearDate')" @click="keyword = ''; feed.reload()">
                <ContentIcon name="x" size="xs" />
              </button>
            </label>
            <DateRangePicker
              :from="dateFrom"
              :to="dateTo"
              :marks="index.days.value"
              :default-month="railMonth"
              @apply="feed.setRange"
              @quick="feed.applyQuickRange"
              @clear="feed.clearDate"
              @month="index.loadMonth"
            />
          </div>

          <!-- 首屏骨架：与信息流同形 -->
          <div v-if="booting" class="skel" aria-hidden="true">
            <div v-for="n in 3" :key="n" class="sk-post">
              <span class="sk sk-av" />
              <div>
                <span class="sk sk-line" style="width: 42%" />
                <span class="sk sk-line" style="width: 96%; margin-top: 12px" />
                <span class="sk sk-line" style="width: 70%; margin-top: 8px" />
                <span v-if="n !== 3" class="sk sk-img" :class="{ wide: n === 1 }" />
              </div>
            </div>
          </div>

          <div v-else :key="`${tab}-${dir}`" class="th-view" :class="[dir, { dim: loading }]">
            <MediaWall v-if="tab === 'media'" :notes="notes" @open="openViewer" />
            <template v-else>
              <div class="posts rise-stagger">
                <NoteCard
                  v-for="n in notes"
                  :key="n.id"
                  :note="n"
                  :avatar="avatar"
                  :name="name"
                  :handle="handle"
                  @open="openViewer"
                  @mood="pickMood"
                  @copy="copyNote"
                />
              </div>
              <p v-if="!loading && !notes.length" class="nores">
                {{ filtered ? t('content.thoughts.emptyFiltered') : t('content.thoughts.empty') }}
              </p>
            </template>

            <div v-if="loadingMore" class="skel more" aria-hidden="true">
              <div class="sk-post">
                <span class="sk sk-av" />
                <div>
                  <span class="sk sk-line" style="width: 42%" />
                  <span class="sk sk-line" style="width: 90%; margin-top: 12px" />
                </div>
              </div>
            </div>
            <p v-else-if="!hasMore && notes.length" class="list-end">{{ config.cfg.site.listEndText }}</p>
          </div>

          <div ref="sentinel" class="sentinel" aria-hidden="true" />
        </div>

        <!-- 右栏：日历 + 心情 -->
        <aside class="rail rise-stagger">
          <div class="card">
            <MonthCalendar
              v-model:month="railMonth"
              :marks="index.days.value"
              :from="singleDay"
              :to="singleDay"
              only-marked
              @pick="pickDay"
            />
          </div>
          <div v-if="index.moods.value.length" class="card">
            <h6>{{ t('content.thoughts.moods') }}<small>{{ t('content.thoughts.moodKinds', { n: index.moods.value.length }) }}</small></h6>
            <div class="moods">
              <button
                v-for="m in index.moods.value"
                :key="m.name"
                type="button"
                :class="{ on: keyword.trim() === m.name }"
                @click="pickMood(m.name)"
              >
                <span class="tag">{{ m.name }}</span><em>{{ m.count }}</em>
              </button>
            </div>
          </div>
        </aside>
      </div>
    </div>

    <ImageViewer
      v-if="viewer"
      :images="viewer.images"
      :start-index="viewer.index"
      :origin-rect="viewer.rect"
      @close="viewer = null"
    />
    <ContentToast />
  </main>
</template>

<style scoped lang="scss">
.thoughts {
  --nav-h: 64px;
  padding: 64px 0 96px;
  min-height: 100vh;
  /* 控制条毛玻璃底横向铺满视口，超出部分裁掉（clip 不建滚动容器，不影响 sticky） */
  overflow-x: clip;
}

.wrap {
  width: min(1040px, calc(100% - 80px));
  margin-inline: auto;
}

.ph {
  padding: 64px 0 28px;

  h1 {
    font-family: var(--font-serif);
    font-size: 52px;
    font-weight: 900;
    line-height: 1.12;
    letter-spacing: 0.01em;
    color: var(--text);
  }

  p {
    margin-top: 12px;
    font-size: 15px;
    color: var(--text-2);
  }
}

.th-grid {
  display: grid;
  grid-template-columns: minmax(0, 640px) 300px;
  justify-content: space-between;
  gap: 64px;
}

.feed { min-width: 0; }

/* ---------- 控制条 ---------- */
.th-ctl {
  position: sticky;
  top: var(--nav-h);
  z-index: 30;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;

  &::before {
    content: '';
    position: absolute;
    inset: calc(var(--nav-h) * -1) -9999px 0;
    background: color-mix(in oklab, var(--bg) 88%, transparent);
    backdrop-filter: blur(20px) saturate(170%);
    -webkit-backdrop-filter: blur(20px) saturate(170%);
    box-shadow: inset 0 -0.5px 0 var(--line);
    opacity: 0;
    transition: opacity var(--dur) var(--ease-out);
    pointer-events: none;
  }

  &.stuck::before { opacity: 1; }
  > * { position: relative; }
}

/* segmented：抬升 + 轻染，选中项图标着色 --ink */
.seg {
  --pad: 3px;
  position: relative;
  display: inline-flex;
  flex: none;
  padding: var(--pad);
  border-radius: var(--r-md);
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line);

  .lift {
    position: absolute;
    top: var(--pad);
    bottom: var(--pad);
    left: var(--l, 0);
    width: var(--w, 0);
    border-radius: max(2px, calc(var(--r-md) - var(--pad)));
    background: var(--lift);
    box-shadow: var(--lift-shadow);
    pointer-events: none;
    transition: left 0.3s var(--ease-out), width 0.46s var(--ease-spring) 0.05s,
      background-color var(--dur), box-shadow var(--dur);
  }

  &.to-l .lift {
    transition: left 0.46s var(--ease-spring) 0.05s, width 0.3s var(--ease-out),
      background-color var(--dur), box-shadow var(--dur);
  }

  &.no-anim .lift { transition: none; }

  button {
    position: relative;
    z-index: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 7px;
    height: 32px;
    padding: 0 16px;
    border: 0;
    border-radius: max(2px, calc(var(--r-md) - var(--pad)));
    background: none;
    font-size: 13.5px;
    color: var(--text-2);
    transition: color var(--dur) var(--ease-out), transform var(--dur-fast) var(--ease-spring);

    :deep(.ci) { transition: color var(--dur); }

    &:hover { color: var(--text); }
    &:active { transform: scale(0.96); }
    &:focus-visible { outline: none; box-shadow: var(--focus); }

    &.on {
      font-weight: 500;
      color: var(--lift-fg);

      :deep(.ci) { color: var(--ink); }
    }
  }
}

.field {
  display: flex;
  flex: 1;
  min-width: 0;
  align-items: center;
  gap: 9px;
  height: 40px;
  padding: 0 10px 0 13px;
  border-radius: var(--r-md);
  background: var(--fill);
  box-shadow: inset 0 0 0 1px var(--line);
  color: var(--text-3);
  cursor: text;
  transition: background-color var(--dur-fast), box-shadow var(--dur-fast), color var(--dur-fast);

  input {
    flex: 1;
    width: 0;
    min-width: 0;
    border: 0;
    outline: none;
    background: none;
    font: inherit;
    font-size: 14px;
    color: var(--text);
    caret-color: var(--ink);

    &::placeholder { color: var(--text-3); }
    &::-webkit-search-cancel-button { display: none; }
  }

  .clr {
    display: grid;
    place-items: center;
    width: 22px;
    height: 22px;
    border: 0;
    border-radius: 50%;
    background: var(--fill-2);
    color: var(--text-2);

    &:hover { color: var(--text); }
  }

  &:hover { box-shadow: inset 0 0 0 1px var(--line-2); }

  &:focus-within {
    background: var(--elev);
    color: var(--text-2);
    box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--ink) 70%, transparent),
      0 0 0 3px color-mix(in oklab, var(--ink) 18%, transparent);
  }
}

:root[data-mode='light'] .field:not(:focus-within) { background: color-mix(in oklab, var(--bg) 40%, white); }

/* ---------- 视图 ---------- */
.th-view {
  animation: rise 0.45s var(--ease-out) backwards;
  transition: opacity 0.26s ease;

  &.from-r { animation-name: slide-l; }
  &.from-l { animation-name: slide-r; }
  &.dim { opacity: 0.5; pointer-events: none; }
}

@keyframes slide-l { from { opacity: 0; transform: translateX(24px); } }
@keyframes slide-r { from { opacity: 0; transform: translateX(-24px); } }

.nores {
  padding: 72px 0;
  text-align: center;
  color: var(--text-3);
}

.list-end {
  padding: 36px 0 0;
  text-align: center;
  font-size: 12px;
  letter-spacing: 0.12em;
  color: var(--text-3);
}

.sentinel { height: 1px; }

/* ---------- 右栏 ---------- */
.rail {
  position: sticky;
  top: calc(var(--nav-h) + 24px);
  align-self: start;
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding-top: 12px;
}

.card {
  padding: 20px;
  border-radius: var(--r-lg);
  background: var(--elev);
  box-shadow: var(--shadow-card);
}

:root[data-mode='dark'] .card {
  background: linear-gradient(180deg, color-mix(in oklab, var(--surface) 90%, white), var(--surface) 70%);
}

h6 {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text);

  small { font-size: 12px; font-weight: 400; color: var(--text-3); }
}

.moods {
  display: flex;
  flex-direction: column;

  button {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 36px;
    margin: 0 -10px;
    padding: 0 10px;
    border: 0;
    border-radius: var(--r-sm);
    background: none;
    font-size: 13.5px;
    color: var(--text-2);
    transition: background-color var(--dur-fast), color var(--dur-fast);

    &:hover { background: var(--fill); color: var(--text); }
    &:focus-visible { outline: none; box-shadow: var(--focus); }
    &.on { background: var(--lift); color: var(--lift-fg); box-shadow: var(--lift-shadow); }
  }

  em {
    font-style: normal;
    font-family: var(--font-mono);
    font-size: 11.5px;
    color: var(--text-3);
  }
}

.tag::before {
  content: '#';
  margin-right: 3px;
  font-family: var(--font-mono);
  font-size: 0.92em;
  color: var(--text-3);
}

/* ---------- 骨架 ---------- */
.sk-post {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr);
  gap: 16px;
  padding: 26px 0;

  & + & { box-shadow: inset 0 0.5px 0 var(--line); }
}

.sk-av { width: 44px; height: 44px; border-radius: 50%; }

.sk-img {
  display: block;
  width: 62%;
  aspect-ratio: 1;
  max-width: 300px;
  margin-top: 14px;
  border-radius: var(--r-lg);

  &.wide { width: 100%; max-width: 560px; aspect-ratio: 16 / 10; }
}

/* ---------- 响应式 ---------- */
@media (max-width: 1300px) {
  .th-grid { grid-template-columns: minmax(0, 1fr) 280px; gap: 48px; }
}

@media (max-width: 1100px) {
  .wrap { width: min(680px, calc(100% - 64px)); }
  .th-grid { grid-template-columns: minmax(0, 1fr); }
  .rail { display: none; }
}
</style>
