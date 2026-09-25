<script setup lang="ts">
/**
 * 后台 · 内容：文章 / 随想分段；列表左滑置顶 / 删除（删除走全局确认模态），点击进入编辑。
 * 分段与路由同步（admin-posts / admin-notes），切换时列表按方向滑入；随想分页无限加载。
 */
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, thumbOf, type AdminPost, type Note } from '../../../api';
import { useDialogStore } from '../../../stores/dialog';
import MaPage from '../../../components/mobile-admin/MaPage.vue';
import MaIcon from '../../../components/mobile-admin/MaIcon.vue';
import MaSegmented from '../../../components/mobile-admin/MaSegmented.vue';
import MaSkeleton from '../../../components/mobile-admin/MaSkeleton.vue';
import MaSwipeRow, { type SwipeAction } from '../../../components/mobile-admin/MaSwipeRow.vue';
import {
  cache,
  loadMoreNotes,
  loadNotes,
  loadPosts,
  openComposer,
  shell,
  toast,
} from '../../../components/mobile-admin/state';
import { mdPlain, relTime } from '../../../components/mobile-admin/format';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const dialog = useDialogStore();

const seg = computed(() => (route.name === 'admin-posts' ? 0 : 1));
const dir = ref<'l' | 'r' | ''>('');
function setSeg(v: number): void {
  if (v === seg.value) return;
  dir.value = v > seg.value ? 'r' : 'l';
  void router.replace({ name: v ? 'admin-notes' : 'admin-posts' });
}

const onlyDrafts = computed(() => route.query.filter === 'draft');
const searching = ref(false);
const q = ref('');
const searchEl = ref<HTMLInputElement | null>(null);
async function toggleSearch(): Promise<void> {
  searching.value = !searching.value;
  if (!searching.value) q.value = '';
  await nextTick();
  if (searching.value) searchEl.value?.focus();
}

const postsReady = computed(() => cache.posts !== null);
const notesReady = computed(() => cache.notes !== null);

function load(): Promise<unknown> {
  return Promise.allSettled([loadPosts(), loadNotes()]);
}
onMounted(() => void load());
watch(() => shell.bump.posts, () => void loadPosts().catch(() => undefined));

/* 新发布的随想：高亮一次 */
const freshId = ref<number | null>(null);
let knownNoteIds = new Set<number>();
watch(
  () => cache.notes,
  (list, prev) => {
    if (prev && list?.length && !knownNoteIds.has(list[0].id)) freshId.value = list[0].id;
    knownNoteIds = new Set((list ?? []).map((n) => n.id));
  },
  { immediate: true },
);

function refresh(done: () => void): void {
  void load().finally(done);
}

const needle = computed(() => q.value.trim().toLowerCase());
const posts = computed<AdminPost[]>(() => {
  let list = [...(cache.posts ?? [])];
  if (onlyDrafts.value) list = list.filter((p) => p.status === 'draft');
  if (needle.value) list = list.filter((p) => `${p.title} ${p.tags.join(' ')} ${p.slug}`.toLowerCase().includes(needle.value));
  return list.sort((a, b) => Number(b.pinned) - Number(a.pinned));
});
const notes = computed<Note[]>(() => {
  const list = cache.notes ?? [];
  if (!needle.value) return list;
  return list.filter((n) => `${n.contentMd} ${n.mood}`.toLowerCase().includes(needle.value));
});
const draftCount = computed(() => (cache.posts ?? []).filter((p) => p.status === 'draft').length);
const sub = computed(() =>
  t('mobileAdmin.content.sub', {
    posts: cache.posts?.length ?? 0,
    notes: cache.notesTotal,
    drafts: draftCount.value,
  }),
);

/* ---------- 左滑操作 ---------- */
function actionsFor(pinned: boolean): SwipeAction[] {
  return [
    { id: 'pin', label: pinned ? t('mobileAdmin.content.unpin') : t('mobileAdmin.content.pin'), icon: 'pin', color: '#ff9500' },
    { id: 'delete', label: t('mobileAdmin.common.delete'), icon: 'trash', color: '#ff3b30' },
  ];
}

async function onPostAction(p: AdminPost, id: string): Promise<void> {
  if (id === 'pin') {
    try {
      const full = await adminApi.post(p.id);
      await adminApi.updatePost(p.id, {
        slug: full.slug,
        title: full.title,
        excerpt: full.excerpt,
        contentMd: full.contentMd ?? '',
        covers: full.covers,
        tags: full.tags,
        status: full.status,
        pinned: !full.pinned,
      });
      p.pinned = !full.pinned;
      toast(p.pinned ? t('mobileAdmin.content.pinned') : t('mobileAdmin.content.unpinned'), p.title);
    } catch {
      toast(t('mobileAdmin.common.saveFailed'), '', 'error');
    }
    return;
  }
  const ok = await dialog.confirm({
    title: t('mobileAdmin.content.deletePostTitle'),
    message: t('mobileAdmin.content.deletePostMsg', { title: p.title }),
    confirmText: t('mobileAdmin.common.delete'),
    cancelText: t('mobileAdmin.common.cancel'),
    danger: true,
  });
  if (!ok) return;
  try {
    await adminApi.deletePost(p.id);
    cache.posts = (cache.posts ?? []).filter((x) => x.id !== p.id);
    toast(t('mobileAdmin.content.deleted'), p.title);
  } catch {
    toast(t('mobileAdmin.common.deleteFailed'), '', 'error');
  }
}

async function onNoteAction(n: Note, id: string): Promise<void> {
  const preview = mdPlain(n.contentMd).slice(0, 12);
  if (id === 'pin') {
    try {
      await adminApi.updateNote(n.id, { contentMd: n.contentMd, mood: n.mood, images: n.images, pinned: !n.pinned });
      n.pinned = !n.pinned;
      toast(n.pinned ? t('mobileAdmin.content.pinned') : t('mobileAdmin.content.unpinned'), preview);
    } catch {
      toast(t('mobileAdmin.common.saveFailed'), '', 'error');
    }
    return;
  }
  const ok = await dialog.confirm({
    title: t('mobileAdmin.content.deleteNoteTitle'),
    message: t('mobileAdmin.content.deleteNoteMsg'),
    confirmText: t('mobileAdmin.common.delete'),
    cancelText: t('mobileAdmin.common.cancel'),
    danger: true,
  });
  if (!ok) return;
  try {
    await adminApi.deleteNote(n.id);
    cache.notes = (cache.notes ?? []).filter((x) => x.id !== n.id);
    cache.notesTotal = Math.max(0, cache.notesTotal - 1);
    toast(t('mobileAdmin.content.deleted'), preview);
  } catch {
    toast(t('mobileAdmin.common.deleteFailed'), '', 'error');
  }
}

function editPost(p: AdminPost): void {
  void router.push({ name: 'admin-write-post', query: { id: String(p.id) } });
}

/* ---------- 随想无限加载 ---------- */
const sentinel = ref<HTMLElement | null>(null);
const loadingMore = ref(false);
let io: IntersectionObserver | null = null;
watch(sentinel, (el) => {
  io?.disconnect();
  if (!el) return;
  io = new IntersectionObserver(async (entries) => {
    if (!entries[0]?.isIntersecting || loadingMore.value) return;
    loadingMore.value = true;
    try {
      await loadMoreNotes();
    } finally {
      loadingMore.value = false;
    }
  }, { rootMargin: '200px' });
  io.observe(el);
});
onBeforeUnmount(() => io?.disconnect());

const hasMoreNotes = computed(() => (cache.notes?.length ?? 0) < cache.notesTotal);
</script>

<template>
  <MaPage :title="t('mobileAdmin.content.title')" :sub="sub" @refresh="refresh">
    <template #right>
      <button class="icbtn tap" :class="{ on: searching }" :aria-label="t('mobileAdmin.common.search')" @click="toggleSearch">
        <MaIcon :name="searching ? 'close' : 'search'" :size="20" />
      </button>
    </template>
    <template #extra>
      <MaSegmented
        :model-value="seg"
        :items="[
          { label: t('mobileAdmin.content.posts'), count: cache.posts?.length ?? '' },
          { label: t('mobileAdmin.content.notes'), count: cache.notes ? cache.notesTotal : '' },
        ]"
        @update:model-value="setSeg"
      />
      <div class="search" :class="{ open: searching }">
        <label class="sf">
          <MaIcon name="search" :size="17" />
          <input
            ref="searchEl"
            v-model="q"
            type="search"
            enterkeyhint="search"
            :placeholder="seg === 0 ? t('mobileAdmin.content.searchPosts') : t('mobileAdmin.content.searchNotes')"
          />
        </label>
      </div>
      <div v-if="onlyDrafts && seg === 0" class="filter">
        <button class="chip on tap" @click="router.replace({ name: 'admin-posts' })">
          {{ t('mobileAdmin.content.onlyDrafts') }}<MaIcon name="close" :size="13" />
        </button>
      </div>
    </template>

    <div :key="seg" class="lists" :class="dir && `in-${dir}`">
      <!-- 文章 -->
      <template v-if="seg === 0">
        <MaSkeleton v-if="!postsReady" variant="rows" :count="6" />
        <TransitionGroup v-else tag="div" name="row" class="rows">
          <MaSwipeRow
            v-for="(p, k) in posts"
            :key="p.id"
            :actions="actionsFor(p.pinned)"
            :style="{ '--k': k }"
            class="stagger"
            @action="onPostAction(p, $event)"
            @tap="editPost(p)"
          >
            <div class="crow">
              <div class="thumb">
                <img v-if="p.covers[0]" :src="thumbOf(p.covers[0])" alt="" loading="lazy" draggable="false" />
                <span v-else class="tx"><MaIcon :name="p.status === 'draft' ? 'pen' : 'docs'" :size="20" /></span>
              </div>
              <div class="ct">
                <b>{{ p.title }}</b>
                <small>
                  <span class="st" :class="p.status === 'draft' ? 'dr' : 'pub'">{{ p.status === 'draft' ? t('mobileAdmin.content.draft') : t('mobileAdmin.content.published') }}</span>
                  <span v-if="p.pinned" class="st pin">{{ t('mobileAdmin.content.pinnedTag') }}</span>
                  <span class="meta">{{ relTime(p.updatedAt, t) }}<template v-if="p.tags.length"> · {{ p.tags.slice(0, 2).join(' / ') }}</template></span>
                </small>
              </div>
              <MaIcon name="chev" :size="16" class="chev" />
            </div>
          </MaSwipeRow>
        </TransitionGroup>
        <div v-if="postsReady && !posts.length" class="empty">
          <span class="ring"><MaIcon :name="needle ? 'search' : 'docs'" /></span>
          <b>{{ needle ? t('mobileAdmin.content.noMatch') : t('mobileAdmin.content.noPosts') }}</b>
          <button v-if="!needle" class="pill-btn tap" @click="router.push({ name: 'admin-write-post' })">{{ t('mobileAdmin.create.post') }}</button>
        </div>
      </template>

      <!-- 随想 -->
      <template v-else>
        <MaSkeleton v-if="!notesReady" variant="rows" :count="6" />
        <TransitionGroup v-else tag="div" name="row" class="rows">
          <MaSwipeRow
            v-for="(n, k) in notes"
            :key="n.id"
            :actions="actionsFor(n.pinned)"
            :style="{ '--k': Math.min(k, 12) }"
            class="stagger"
            @action="onNoteAction(n, $event)"
            @tap="openComposer(n.id)"
          >
            <div class="crow" :class="{ fresh: freshId === n.id }">
              <div class="thumb">
                <img v-if="n.images[0]" :src="thumbOf(n.images[0])" alt="" loading="lazy" draggable="false" />
                <span v-else class="tx"><MaIcon name="bubble" :size="20" /></span>
              </div>
              <div class="ct">
                <b>{{ mdPlain(n.contentMd) }}</b>
                <small>
                  <span v-if="n.pinned" class="st pin">{{ t('mobileAdmin.content.pinnedTag') }}</span>
                  <span v-if="n.mood" class="mood">{{ n.mood }}</span>
                  <span class="meta">{{ relTime(n.createdAt, t) }}<template v-if="n.images.length"> · {{ t('mobileAdmin.content.images', { n: n.images.length }) }}</template></span>
                </small>
              </div>
            </div>
          </MaSwipeRow>
        </TransitionGroup>
        <div v-if="notesReady && hasMoreNotes && !needle" ref="sentinel" class="more">
          <MaSkeleton variant="rows" :count="2" />
        </div>
        <div v-if="notesReady && !notes.length" class="empty">
          <span class="ring"><MaIcon :name="needle ? 'search' : 'bubble'" /></span>
          <b>{{ needle ? t('mobileAdmin.content.noMatch') : t('mobileAdmin.content.noNotes') }}</b>
          <button v-if="!needle" class="pill-btn tap" @click="openComposer(null)">{{ t('mobileAdmin.create.note') }}</button>
        </div>
      </template>
    </div>

    <p v-if="(seg === 0 ? posts.length : notes.length) > 0" class="hint">{{ t('mobileAdmin.content.swipeHint') }}</p>
  </MaPage>
</template>

<style scoped lang="scss">
.icbtn.on {
  background: var(--lift);
  box-shadow: var(--lift-shadow);
  color: var(--ink);
}

.search {
  max-height: 0;
  overflow: hidden;
  opacity: 0;
  margin: 0 20px;
  transition: max-height 0.4s var(--ease-sheet), opacity 0.3s;

  &.open {
    max-height: 60px;
    opacity: 1;
  }
}

.sf {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 40px;
  padding: 0 12px;
  margin-top: 10px;
  border-radius: var(--r-md);
  background: var(--fill-2);
  color: var(--text-3);

  input {
    flex: 1;
    min-width: 0;
    font-size: 16px;
    color: var(--text);
    caret-color: var(--ink);

    &::placeholder { color: var(--text-3); }
  }

}

.filter {
  padding: 10px 20px 0;

  .chip { height: 30px; font-size: 13px; }
}

.lists {
  padding-top: 4px;

  &.in-r { animation: ma-seg-r 0.45s var(--ease-sheet); }
  &.in-l { animation: ma-seg-l 0.45s var(--ease-sheet); }
}

@keyframes ma-seg-r {
  from { opacity: 0; transform: translateX(40px); }
}

@keyframes ma-seg-l {
  from { opacity: 0; transform: translateX(-40px); }
}

.stagger {
  animation: ma-row-in 0.5s var(--ease-out) backwards;
  animation-delay: calc(var(--k) * 35ms);
}

@keyframes ma-row-in {
  from { opacity: 0; transform: translateY(14px); }
}

.crow {
  position: relative;
  width: 100%;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  text-align: left;
  cursor: pointer;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 84px;
    right: 0;
    height: 0.5px;
    background: var(--line);
  }

  &:active { background: var(--fill); }

  &.fresh { animation: ma-fresh 1.8s var(--ease-out); }

  .chev { color: var(--text-3); }
}

.rows > :first-child .crow::before { display: none; }

@keyframes ma-fresh {
  0% { background: var(--tint); }
  100% { background: transparent; }
}

.thumb {
  position: relative;
  width: 52px;
  height: 52px;
  border-radius: var(--r-md);
  overflow: hidden;
  flex: none;
  box-shadow: 0 0 0 0.5px var(--line);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .tx {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    background: var(--fill);
    color: var(--text-3);
  }
}

.ct {
  flex: 1;
  min-width: 0;

  b {
    display: block;
    font-size: 15px;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  small {
    display: flex;
    gap: 6px;
    align-items: center;
    margin-top: 4px;
    font-size: 12px;
    color: var(--text-3);
    white-space: nowrap;
    overflow: hidden;
  }

  .meta {
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .mood {
    color: var(--text-2);

    &::before {
      content: '#';
      color: var(--text-3);
      margin-right: 2px;
    }
  }
}

/* 删除：行向左收起并折叠高度 */
.row-leave-active {
  transition: opacity 0.3s var(--ease-out), transform 0.35s var(--ease-out), max-height 0.4s var(--ease-sheet) 0.1s;
  max-height: 90px;
}

.row-leave-to {
  opacity: 0;
  transform: translateX(-40%);
  max-height: 0;
}

.row-move {
  transition: transform 0.45s var(--ease-spring);
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 60px 30px;
  text-align: center;
  color: var(--text-3);

  .ring {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: var(--fill);
  }

  b {
    font-size: 15px;
    font-weight: 500;
    color: var(--text-2);
  }

  .pill-btn { margin-top: 6px; }
}

.hint {
  padding: 18px 20px 0;
  text-align: center;
  font-size: 12px;
  color: var(--text-3);
}
</style>
