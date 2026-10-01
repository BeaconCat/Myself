<script setup lang="ts">
import { useNoteTime } from '../../composables/useNoteTime';
import { useThoughtIdentity } from '../../components/thoughts/useThoughtIdentity';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { thumbOf, type Note } from '../../api';
import { useConfigStore } from '../../stores/config';
import { useNotesFeed, type NotesTab } from '../../composables/useNotesFeed';
import { render } from '../../utils/markdown';
import LargeTitlePage from '../../components/mobile/LargeTitlePage.vue';
import MediaViewer, { type ViewerItem } from '../../components/mobile/MediaViewer.vue';
import MIcon from '../../components/mobile/MIcon.vue';
import EngageBar from '../../components/engage/EngageBar.vue';
import { useRouter } from 'vue-router';
import { openSearch } from '../../components/mobile/shell';
import { formatDateTime } from '../../utils/date';
import IdentityName from '../../components/common/IdentityName.vue';

/**
 * 移动端随想：信息流 / 媒体 segmented（横排，滑块回弹，视图按方向滑入）；
 * 九宫格配图点开全屏查看器（从缩略图展开、左右滑、下拉关闭）；触底续载；下拉刷新。
 */
const { t } = useI18n();
const router = useRouter();
const noteTime = useNoteTime();

/** 点卡片空白处进入随想详情（配图、按钮、链接除外） */
function openDetail(e: MouseEvent, n: Note): void {
  if ((e.target as HTMLElement).closest('a, button, .grid')) return;
  void router.push(`/thoughts/${n.id}`);
}
const config = useConfigStore();

const sentinel = ref<HTMLElement | null>(null);
const feed = useNotesFeed(sentinel);
const { notes, tab, loading, loadingMore, hasMore } = feed;

/* 首屏：reload 发起前 loading 仍为 false，用「是否完成过一次加载」兜底显示骨架 */
const loadedOnce = ref(false);
watch(loading, (on, was) => { if (was && !on) loadedOnce.value = true; });
const showSkeleton = computed(() => loading.value || !loadedOnce.value);

const { avatar, name, alias, fullName, handle } = useThoughtIdentity();

const viewDir = ref<'in-r' | 'in-l'>('in-r');
function switchTab(next: NotesTab): void {
  if (next === tab.value) return;
  viewDir.value = next === 'media' ? 'in-r' : 'in-l';
  feed.switchTab(next);
}


function gridClass(n: number): string {
  if (n === 1) return 'g1';
  if (n === 2) return 'g2';
  if (n === 4) return 'g4';
  return 'g3';
}

function plain(src: string): string {
  return src.replace(/!\[[^\]]*]\([^)]*\)/g, '').replace(/\[([^\]]*)]\([^)]*\)/g, '$1').replace(/[#>*_`~]/g, '').trim();
}

/* 媒体墙：把媒体随想的配图摊平 */
const media = computed(() =>
  notes.value.flatMap((n) => n.images.map((src, k) => ({ src, note: n, k, key: `${n.id}-${k}` }))),
);

/* ---------- 查看器 ---------- */
const viewerOpen = ref(false);
const viewerIndex = ref(0);
const viewerItems = ref<ViewerItem[]>([]);
const viewerKeys = ref<string[]>([]);
const thumbs = new Map<string, HTMLElement>();

function setThumb(key: string, el: unknown): void {
  if (el instanceof HTMLElement) thumbs.set(key, el);
  else thumbs.delete(key);
}

function itemOf(n: Note, src: string): ViewerItem {
  return {
    src,
    caption: plain(n.contentMd),
    meta: `${fullName.value} · ${formatDateTime(n.createdAt, true)}${n.mood ? ` · #${n.mood}` : ''}`,
  };
}

function openNote(n: Note, k: number): void {
  viewerItems.value = n.images.map((src) => itemOf(n, src));
  viewerKeys.value = n.images.map((_, i) => `${n.id}-${i}`);
  viewerIndex.value = k;
  viewerOpen.value = true;
}

function openMedia(i: number): void {
  viewerItems.value = media.value.map((m) => itemOf(m.note, m.src));
  viewerKeys.value = media.value.map((m) => `m-${m.key}`);
  viewerIndex.value = i;
  viewerOpen.value = true;
}

const origin = (i: number): HTMLElement | null => thumbs.get(viewerKeys.value[i]) ?? null;

const mediaCount = computed(() => media.value.length);
</script>

<template>
  <LargeTitlePage :title="t('mobile.thoughts.title')" :sub="config.cfg.thoughts.subtitle" :refresh="feed.reload">
    <template #right>
      <button class="m-icbtn m-tap" :aria-label="t('mobile.tabs.search')" @click="openSearch()"><MIcon name="search" /></button>
    </template>

    <template #extra>
      <div class="segc" :style="{ '--i': tab === 'media' ? 1 : 0 }">
        <span class="th" aria-hidden="true" />
        <button :class="{ on: tab === 'posts' }" @click="switchTab('posts')">
          <MIcon name="bubble" class="xs" />{{ t('mobile.thoughts.feed') }}
        </button>
        <button :class="{ on: tab === 'media' }" @click="switchTab('media')">
          <MIcon name="image" class="xs" />{{ t('mobile.thoughts.media') }}<em v-if="tab === 'media' && mediaCount">{{ mediaCount }}</em>
        </button>
      </div>
    </template>

    <Transition name="m-swap" mode="out-in">
      <!-- 骨架 -->
      <div v-if="showSkeleton" :key="`sk-${tab}`" class="sk">
        <template v-if="tab === 'posts'">
          <div v-for="n in 3" :key="n" class="post">
            <span class="m-sk sk-av" />
            <div class="post-main">
              <span class="m-sk m-sk-line" style="width: 46%; height: 12px" />
              <span class="m-sk m-sk-line" style="width: 96%; margin-top: 12px" />
              <span class="m-sk m-sk-line" style="width: 72%; margin-top: 8px" />
              <div v-if="n !== 2" class="m-sk sk-grid" :class="{ tall: n === 1 }" />
            </div>
          </div>
        </template>
        <div v-else class="mg sk-mg">
          <span v-for="n in 12" :key="n" class="m-sk" />
        </div>
      </div>

      <div v-else :key="`v-${tab}`" class="th-view" :class="viewDir">
        <template v-if="tab === 'posts'">
          <div v-if="!notes.length" class="m-empty">{{ t('mobile.thoughts.empty') }}</div>
          <article
            v-for="(n, i) in notes"
            :id="`note-${n.id}`"
            :key="n.id"
            class="post m-in"
            :style="{ '--i': Math.min(i, 8) }"
            @click="openDetail($event, n)"
          >
            <span class="av"><img class="m-avatar" :src="avatar" alt="" draggable="false" /></span>
            <div class="post-main">
              <header>
                <div class="author-line">
                  <b :title="alias ? `${name} ${alias}` : name"><IdentityName :name="name" :alias="alias" /></b>
                  <span v-if="handle" class="handle" :title="`@${handle}`">@{{ handle }}</span>
                  <time :datetime="`${n.createdAt.replace(' ', 'T')}Z`" :title="formatDateTime(n.createdAt, true)">{{ noteTime(n.createdAt) }}</time>
                </div>
                <i v-if="n.pinned" class="m-pin">{{ t('noteDetail.pinned') }}</i>
                <em v-if="n.mood" class="m-mood">{{ n.mood }}</em>
              </header>
              <!-- eslint-disable-next-line vue/no-v-html -->
              <div class="body markdown-content" v-html="render(n.contentMd, `note-${n.id}`)" />
              <div v-if="n.images.length" class="grid" :class="gridClass(n.images.length)">
                <button
                  v-for="(src, k) in n.images"
                  :key="k"
                  class="cell"
                  @click="openNote(n, k)"
                >
                  <img
                    :ref="(el) => setThumb(`${n.id}-${k}`, el)"
                    :src="thumbOf(src)"
                    alt=""
                    loading="lazy"
                    draggable="false"
                  />
                </button>
              </div>
              <EngageBar class="act" target="note" :id="n.id" :link="`/thoughts/${n.id}`" :text="n.contentMd" @comment="router.push(`/thoughts/${n.id}#comments`)" />
            </div>
          </article>
        </template>

        <template v-else>
          <div class="mg-h">
            <span>{{ t('mobile.thoughts.allMedia') }}</span>
            <span>{{ t('mobile.thoughts.mediaMeta', { n: mediaCount, m: notes.length }) }}</span>
          </div>
          <div v-if="!media.length" class="m-empty">{{ t('mobile.thoughts.empty') }}</div>
          <div class="mg">
            <button
              v-for="(m, i) in media"
              :key="m.key"
              class="cell m-in"
              :style="{ '--i': Math.min(i, 14) }"
              @click="openMedia(i)"
            >
              <img
                :ref="(el) => setThumb(`m-${m.key}`, el)"
                :src="thumbOf(m.src)"
                alt=""
                loading="lazy"
                draggable="false"
              />
            </button>
          </div>
        </template>

        <div v-if="loadingMore" class="more-sk"><span class="m-sk m-sk-line" style="width: 40%" /></div>
        <div v-else-if="!hasMore && notes.length" class="m-end">{{ config.cfg.site.listEndText }}</div>
      </div>
    </Transition>
    <div ref="sentinel" class="sentinel" />

    <MediaViewer v-model:open="viewerOpen" v-model:index="viewerIndex" :items="viewerItems" :origin="origin" />
  </LargeTitlePage>
</template>

<style scoped lang="scss">
.segc {
  position: relative;
  display: grid;
  grid-template-columns: 1fr 1fr;
  height: 38px;
  margin: 0 16px;
  padding: 3px;
  border-radius: var(--r-md);
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line);

  button {
    position: relative;
    z-index: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    font-size: 14px;
    color: var(--text-2);
    border-radius: calc(var(--r-md) - 3px);
    transition: color var(--dur);
    white-space: nowrap;

    .m-ic { transition: color var(--dur); }

    &.on {
      color: var(--lift-fg);
      font-weight: 500;

      .m-ic { color: var(--ink); }
    }

    em {
      font-style: normal;
      font-family: var(--m-font-mono);
      font-size: 12px;
      color: var(--text-3);
    }
  }

  .th {
    position: absolute;
    top: 3px;
    bottom: 3px;
    left: 3px;
    width: calc(50% - 3px);
    border-radius: calc(var(--r-md) - 3px);
    background: var(--lift);
    box-shadow: var(--lift-shadow);
    transform: translateX(calc(var(--i, 0) * 100%));
    transition: transform 0.45s var(--ease-spring), background-color var(--dur), box-shadow var(--dur);
  }
}

.th-view.in-r { animation: in-r 0.42s var(--ease-out); }
.th-view.in-l { animation: in-l 0.42s var(--ease-out); }

@keyframes in-r { from { opacity: 0; transform: translateX(24px); } }
@keyframes in-l { from { opacity: 0; transform: translateX(-24px); } }

.post {
  position: relative;
  display: flex;
  gap: 12px;
  padding: 14px 16px 6px;

  & + &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 68px;
    right: 0;
    height: 0.5px;
    background: var(--line);
  }

  .av {
    width: 40px;
    height: 40px;
    flex: none;
  }
}

.post-main {
  flex: 1;
  min-width: 0;

  header {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 6px;
    font-size: 13px;
    color: var(--text-3);
    white-space: nowrap;

    b {
      font-size: 15px;
      color: var(--text);
      font-weight: 600;
    }

    span {
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .author-line { display: flex; flex-basis: 100%; min-width: 0; align-items: baseline; gap: 6px; }
    .author-line b, .handle { min-width: 0; overflow: hidden; text-overflow: ellipsis; }
    time { flex: none; font-size: 12px; line-height: 1.6; white-space: nowrap; }
    .m-mood { margin-left: auto; }

    .m-pin {
      flex: none;
      padding: 0 7px;
      border-radius: var(--r-pill);
      background: color-mix(in oklab, var(--ink) 12%, transparent);
      color: var(--ink);
      font-size: 11.5px;
      font-style: normal;
      font-weight: 500;
      line-height: 18px;
    }
  }
}

.body { margin-top: 8px; user-select: text; -webkit-user-select: text; }

.grid {
  display: grid;
  gap: 3px;
  margin-top: 10px;
  border-radius: var(--r-lg);
  overflow: hidden;

  &.g1 { grid-template-columns: 1fr; }
  /* 单图收一档：16:10（原 4:3），不再占满大半屏 */
  &.g1 .cell { aspect-ratio: 16 / 10; }
  &.g2 { grid-template-columns: 1fr 1fr; }
  &.g3 { grid-template-columns: repeat(3, 1fr); }

  &.g4 {
    grid-template-columns: 1fr 1fr;
    width: 78%;
  }
}

.cell {
  position: relative;
  aspect-ratio: 1;
  display: block;
  overflow: hidden;
  background: var(--fill-3) !important;

  img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: filter var(--dur);
  }

  &:active img { filter: brightness(0.8); }
}

.act {
  display: flex;
  gap: 22px;
  margin-top: 2px;
  margin-left: -6px;

  button {
    display: flex;
    align-items: center;
    height: 30px;
    padding: 0 6px;
    border-radius: var(--r-sm);
    color: var(--text-3);

    .m-ic { width: 19px; height: 19px; }
  }
}

.mg-h {
  padding: 10px 16px;
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: var(--text-3);
}

.mg {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2px;
}

.sentinel { height: 1px; }

.more-sk {
  display: flex;
  justify-content: center;
  padding: 18px 0;
}

/* 骨架 */
.sk-av {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  flex: none;
}

.sk .m-sk-line { display: block; }

.sk-grid {
  margin-top: 12px;
  height: 120px;
  border-radius: var(--r-lg);

  &.tall { height: 210px; }
}

.sk-mg {
  padding-top: 38px;

  span {
    aspect-ratio: 1;
    border-radius: 0;
  }
}
</style>
