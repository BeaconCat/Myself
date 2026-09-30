<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, thumbOf, type AdminPost, type BatchAction } from '../../api';
import { useDialogStore } from '../../stores/dialog';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import StSeg from './studio/StSeg.vue';
import PopMenu from './studio/PopMenu.vue';
import LightCover from './studio/LightCover.vue';
import EmptyArt from './studio/EmptyArt.vue';
import { Check, FilePen } from 'lucide';
import Icon from '../../components/ui/Icon.vue';
import BatchBar from './studio/BatchBar.vue';
import { refreshCounts } from './studio/state';
import { toast } from './studio/toast';
import { dateText, relTime } from './studio/format';
import type { MenuItem } from './studio/types';

/** 文章：网格 / 列表，状态筛选，搜索，置顶，隐藏，删除；勾选后批量隐藏 / 取消隐藏 / 删除 */
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const dialog = useDialogStore();

type Filter = 'all' | 'published' | 'draft' | 'pinned' | 'hidden';
type Layout = 'grid' | 'list';

const LAYOUT_KEY = 'myself.studio.postsLayout';
const posts = ref<AdminPost[]>([]);
const loaded = ref(false);
const filter = ref<Filter>(['published', 'draft', 'pinned', 'hidden'].includes(String(route.query.status)) ? (route.query.status as Filter) : 'all');
const query = ref('');
const layout = ref<Layout>((() => {
  try {
    return localStorage.getItem(LAYOUT_KEY) === 'list' ? 'list' : 'grid';
  } catch {
    return 'grid';
  }
})());
watch(layout, (v) => {
  try {
    localStorage.setItem(LAYOUT_KEY, v);
  } catch { /* 忽略 */ }
});
watch(filter, (v) => void router.replace({ query: v === 'all' ? {} : { status: v } }));

async function load(): Promise<void> {
  try {
    posts.value = await adminApi.posts();
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
  loaded.value = true;
}

const counts = computed(() => ({
  all: posts.value.length,
  published: posts.value.filter((p) => p.status === 'published').length,
  draft: posts.value.filter((p) => p.status === 'draft').length,
  pinned: posts.value.filter((p) => p.pinned).length,
  hidden: posts.value.filter((p) => p.hidden).length,
}));

const list = computed(() => {
  const q = query.value.trim().toLowerCase();
  return posts.value.filter((p) => {
    if (filter.value === 'pinned') {
      if (!p.pinned) return false;
    } else if (filter.value === 'hidden') {
      if (!p.hidden) return false;
    } else if (filter.value !== 'all' && p.status !== filter.value) return false;
    if (!q) return true;
    return `${p.title} ${p.excerpt} ${p.tags.join(' ')} ${p.slug}`.toLowerCase().includes(q);
  });
});

const FILTERS: Filter[] = ['all', 'published', 'draft', 'pinned', 'hidden'];

/* ---------- 勾选与批量：勾选角标进入选择态，此后点卡片即勾选 ---------- */
const picked = ref<Set<number>>(new Set());
const picking = computed(() => picked.value.size > 0);
watch(filter, () => { picked.value = new Set(); });

function togglePick(p: AdminPost): void {
  const next = new Set(picked.value);
  if (next.has(p.id)) next.delete(p.id);
  else next.add(p.id);
  picked.value = next;
}

function pickAll(): void {
  picked.value = picked.value.size >= list.value.length ? new Set() : new Set(list.value.map((p) => p.id));
}

function open(p: AdminPost): void {
  if (picking.value) togglePick(p);
  else edit(p);
}

/** 批量（或单篇）隐藏 / 取消隐藏 / 删除；删除先确认 */
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
    const { affected } = await adminApi.batchPosts(ids, action);
    const set = new Set(ids);
    if (action === 'delete') posts.value = posts.value.filter((p) => !set.has(p.id));
    else posts.value.forEach((p) => { if (set.has(p.id)) p.hidden = action === 'hide'; });
    picked.value = new Set();
    toast(t(`studio.batch.${action === 'hide' ? 'hiddenDone' : action === 'show' ? 'shownDone' : 'deletedDone'}`, { n: affected }), {
      icon: action === 'delete' ? 'trash' : action === 'hide' ? 'eyeOff' : 'eye',
    });
    void refreshCounts();
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

function edit(p: AdminPost): void {
  void router.push({ name: 'admin-write-post', query: { id: String(p.id) } });
}

async function togglePin(p: AdminPost): Promise<void> {
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
    toast(p.pinned ? t('studio.pinned') : t('studio.unpinned'), { icon: 'pin' });
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

async function remove(p: AdminPost): Promise<void> {
  const ok = await dialog.confirm({
    title: t('studio.posts.deleteTitle', { title: p.title || t('studio.untitled') }),
    message: t('studio.posts.deleteBody'),
    confirmText: t('studio.delete'),
    danger: true,
  });
  if (!ok) return;
  try {
    await adminApi.deletePost(p.id);
    posts.value = posts.value.filter((x) => x.id !== p.id);
    toast(t('studio.deleted'), { icon: 'trash' });
    void refreshCounts();
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

function menu(p: AdminPost): MenuItem[] {
  const items: MenuItem[] = [
    { icon: 'pen', label: t('studio.edit'), run: () => edit(p) },
    { icon: 'pin', label: p.pinned ? t('studio.unpin') : t('studio.pin'), run: () => void togglePin(p) },
  ];
  items.push({ icon: p.hidden ? 'eye' : 'eyeOff', label: p.hidden ? t('studio.batch.show') : t('studio.batch.hide'), run: () => void batch(p.hidden ? 'show' : 'hide', [p.id]) });
  if (p.status === 'published' && !p.hidden) {
    items.push({ icon: 'external', label: t('studio.posts.openSite'), run: () => window.open(`/articles/${encodeURIComponent(p.slug)}`, '_blank', 'noopener') });
  }
  items.push({ icon: 'trash', label: t('studio.delete'), danger: true, divider: true, run: () => void remove(p) });
  return items;
}

/* 键盘：/ 聚焦搜索 */
const searchEl = ref<HTMLInputElement | null>(null);
function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape' && picking.value) picked.value = new Set();
  const el = document.activeElement as HTMLElement | null;
  if (el && (/INPUT|TEXTAREA|SELECT/.test(el.tagName) || el.isContentEditable)) return;
  if (e.key === '/') {
    e.preventDefault();
    searchEl.value?.focus();
  }
}

onMounted(() => {
  void load();
  window.addEventListener('keydown', onKey);
});
onBeforeUnmount(() => window.removeEventListener('keydown', onKey));
</script>

<template>
  <section class="studio view">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.posts.title') }}</h1>
        <p>{{ t('studio.posts.desc') }}</p>
      </div>
      <div class="act">
        <router-link class="st-btn p" :to="{ name: 'admin-write-post' }"><SIcon name="plus" :size="18" />{{ t('studio.posts.new') }}</router-link>
      </div>
    </div>

    <div class="toolbar">
      <div class="chips">
        <button v-for="f in FILTERS" :key="f" type="button" class="st-chip" :class="{ on: filter === f }" @click="filter = f">
          {{ t(`studio.posts.f.${f}`) }}<span class="n">{{ counts[f] }}</span>
        </button>
      </div>
      <span class="sp" />
      <label class="st-field search">
        <SIcon name="search" :size="18" />
        <input ref="searchEl" v-model="query" :placeholder="t('studio.posts.search')" />
        <kbd class="st-kbd">/</kbd>
      </label>
      <StSeg
        v-model="layout"
        icon-only
        :options="[
          { value: 'grid', icon: 'grid', title: t('studio.posts.grid') },
          { value: 'list', icon: 'list', title: t('studio.posts.list') },
        ]"
      />
    </div>

    <div v-if="loaded && !list.length" class="st-empty">
      <EmptyArt :icon="FilePen" />
      <h4>{{ query ? t('studio.posts.emptyQuery', { q: query }) : t('studio.posts.empty') }}</h4>
      <p>{{ t('studio.posts.emptySub') }}</p>
      <router-link class="st-btn p" :to="{ name: 'admin-write-post' }"><SIcon name="pen" :size="18" />{{ t('studio.posts.writeOne') }}</router-link>
    </div>

    <div v-else-if="layout === 'grid'" :key="`g-${filter}`" class="pgrid">
      <article
        v-for="(p, i) in list"
        :key="p.id"
        class="pcard st-rise"
        :class="{ on: picked.has(p.id), picking, hid: p.hidden }"
        :style="{ '--i': Math.min(i, 12) }"
        @click="open(p)"
      >
        <LightCover class="pcv" :src="p.covers[0] ? thumbOf(p.covers[0]) : ''" :seed="p.slug">
          <span class="pk" :title="t('studio.batch.pick')" @click.stop="togglePick(p)"><span class="st-ck" :class="{ on: picked.has(p.id) }"><Icon :icon="Check" /></span></span>
          <span v-if="p.pinned" class="pin" :title="t('studio.pinned')"><SIcon name="pin" :size="16" /></span>
        </LightCover>
        <div class="bd">
          <h3>{{ p.title || t('studio.untitled') }}</h3>
          <p>{{ p.excerpt || t('studio.posts.noExcerpt') }}</p>
          <div class="meta">
            <span class="st-badge" :class="`st-${p.status}`"><i class="st-dot" />{{ t(`studio.status.${p.status}`) }}</span>
            <span v-if="p.hidden" class="hid-tag" :title="t('studio.batch.hiddenHint')"><SIcon name="eyeOff" :size="14" />{{ t('studio.batch.hidden') }}</span>
            <span class="mono">{{ dateText(p.createdAt) }}</span>
            <span class="tags"><span v-for="tag in p.tags.slice(0, 2)" :key="tag" class="tag">{{ tag }}</span></span>
            <span class="more" @click.stop><PopMenu :items="menu(p)" /></span>
          </div>
        </div>
      </article>
    </div>

    <div v-else :key="`l-${filter}`" class="plist">
      <div class="plist-h">
        <span />
        <span>{{ t('studio.posts.colTitle') }}</span>
        <span>{{ t('studio.posts.colTags') }}</span>
        <span>{{ t('studio.posts.colStatus') }}</span>
        <span>{{ t('studio.posts.colCreated') }}</span>
        <span>{{ t('studio.posts.colUpdated') }}</span>
        <span />
      </div>
      <div
        v-for="(p, i) in list"
        :key="p.id"
        class="prow st-rise"
        :class="{ on: picked.has(p.id), picking, hid: p.hidden }"
        :style="{ '--i': Math.min(i, 12) }"
        @click="open(p)"
      >
        <LightCover class="rcv" :src="p.covers[0] ? thumbOf(p.covers[0]) : ''" :seed="p.slug">
          <span class="pk" :title="t('studio.batch.pick')" @click.stop="togglePick(p)"><span class="st-ck" :class="{ on: picked.has(p.id) }"><Icon :icon="Check" /></span></span>
        </LightCover>
        <div class="tt">
          <h3><SIcon v-if="p.pinned" name="pin" :size="16" class="pin-i" />{{ p.title || t('studio.untitled') }}</h3>
          <small>/{{ p.slug }}</small>
        </div>
        <span class="tags"><span v-for="tag in p.tags.slice(0, 2)" :key="tag" class="tag">{{ tag }}</span></span>
        <span class="stc">
          <span class="st-badge" :class="`st-${p.status}`"><i class="st-dot" />{{ t(`studio.status.${p.status}`) }}</span>
          <span v-if="p.hidden" class="hid-tag" :title="t('studio.batch.hiddenHint')"><SIcon name="eyeOff" :size="14" />{{ t('studio.batch.hidden') }}</span>
        </span>
        <span class="num">{{ dateText(p.createdAt) }}</span>
        <span class="num">{{ relTime(p.updatedAt || p.createdAt) }}</span>
        <span @click.stop><PopMenu :items="menu(p)" /></span>
      </div>
    </div>

    <BatchBar :show="picking" :count="picked.size" :total="list.length" @all="pickAll" @clear="picked = new Set()">
      <button type="button" class="st-btn sm" @click="batch('hide')"><SIcon name="eyeOff" :size="16" />{{ t('studio.batch.hide') }}</button>
      <button type="button" class="st-btn sm" @click="batch('show')"><SIcon name="eye" :size="16" />{{ t('studio.batch.show') }}</button>
      <button type="button" class="st-btn sm danger" @click="batch('delete')"><SIcon name="trash" :size="16" />{{ t('studio.delete') }}</button>
    </BatchBar>
  </section>
</template>

<style scoped lang="scss">
.view {
  max-width: 1280px;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
  flex-wrap: wrap;

  .chips { display: flex; gap: 6px; flex-wrap: wrap; }
  .sp { flex: 1; }

  .search {
    width: 260px;
  }
}

.pgrid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  align-items: stretch;
  gap: 18px;
}

/* 卡片：封面约 2:1 满幅，标题 / 摘要 / 状态·日期·标签·菜单 在封面下同一信息块 */
.pcard {
  position: relative;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: var(--r-lg);
  cursor: pointer;
  background: var(--paper);
  box-shadow: 0 0 0 1px var(--line-2), 0 1px 2px var(--line);
  transition: transform var(--dur) var(--ease-spring), box-shadow var(--dur) var(--ease-out);

  &:hover {
    transform: translateY(-3px);
    box-shadow: 0 0 0 1px var(--line-2), var(--sh-card-hover);
  }

  .pcv {
    aspect-ratio: 2 / 1;

    &::after {
      content: '';
      position: absolute;
      inset: 0;
      z-index: 2;
      /* 封面掠光：属于封面光影（品牌），非控件发光 */
      background: linear-gradient(105deg, transparent 30%, rgba(255, 255, 255, 0.28) 48%, transparent 62%);
      transform: translateX(-110%);
      transition: transform 1s var(--ease-out);
    }

    :deep(img) { transition: transform 1.2s var(--ease-out); }
  }

  &:hover .pcv::after { transform: translateX(110%); }
  &:hover .pcv :deep(img) { transform: scale(1.04); }

  .pin {
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
  }

  .bd {
    display: flex;
    flex-direction: column;
    flex: 1;
    padding: 16px 18px 12px;
  }

  h3 {
    font: 700 19px/1.45 var(--font-serif);
    margin: 0 0 6px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  p {
    margin: 0 0 12px;
    font-size: 14.5px;
    line-height: 1.65;
    color: var(--st-ink-2);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .meta {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-top: auto;
    padding-top: 10px;
    border-top: 1px solid var(--line);
    font-size: 13px;
    color: var(--st-ink-3);
    white-space: nowrap;

    .mono { font-size: 12.5px; }
    .tags { display: flex; gap: 8px; min-width: 0; overflow: hidden; flex: 1; }
  }

  .more { flex: none; margin-right: -8px; }
}

:root[data-mode='dark'] .pcard { background: color-mix(in oklab, var(--paper) 55%, var(--well)); }

.tag {
  color: var(--st-ink-2);
  white-space: nowrap;

  &::before { content: '#'; color: var(--st-ink-4); margin-right: 2px; }
}

/* 勾选角标：悬停浮现，进入选择态后常显；已选卡片描边 */
.pk {
  position: absolute;
  left: 10px;
  top: 10px;
  z-index: 3;
  padding: 3px;
  opacity: 0;
  transform: scale(0.85);
  transition: opacity var(--dur-fast), transform var(--dur-fast) var(--ease-out);

  .st-ck { width: 20px; height: 20px; background: rgba(0, 0, 0, 0.18); box-shadow: 0 0 0 1.5px rgba(255, 255, 255, 0.92) inset, 0 1px 4px rgba(0, 0, 0, 0.25); }
  .st-ck.on { background: var(--solid); box-shadow: 0 0 0 1.5px var(--solid) inset; }
}

.pcard:hover .pk, .prow:hover .pk, .picking .pk { opacity: 1; transform: none; }
.pcard.on { box-shadow: 0 0 0 2px var(--ink), var(--sh-card-hover); }
.prow.on { background: var(--tint); }
.prow .rcv .pk { left: 4px; top: 4px; padding: 0; }
.prow .rcv .pk .st-ck { width: 18px; height: 18px; }

/* 已隐藏：封面褪色，标题降一级 */
.pcard.hid .pcv, .prow.hid .rcv { filter: grayscale(0.7) opacity(0.6); }
.pcard.hid h3, .prow.hid h3 { color: var(--st-ink-3); }

.hid-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12.5px;
  color: var(--st-ink-3);
  white-space: nowrap;
}

.stc { display: flex; flex-direction: column; align-items: flex-start; gap: 4px; }


$cols: 96px minmax(0, 1fr) 150px 96px 100px 92px 38px;

.plist { display: flex; flex-direction: column; }

.plist-h {
  display: grid;
  grid-template-columns: $cols;
  gap: 18px;
  padding: 0 12px 10px;
  font-size: 12.5px;
  color: var(--st-ink-4);
  letter-spacing: 0.06em;
  border-bottom: 1px solid var(--line);
  margin-bottom: 6px;

  span:nth-child(5), span:nth-child(6) { text-align: right; }
}

.prow {
  display: grid;
  grid-template-columns: $cols;
  align-items: center;
  gap: 18px;
  padding: 12px;
  border-radius: var(--r-md);
  cursor: pointer;
  transition: background var(--dur-fast);

  &:hover { background: var(--well); }

  .rcv { aspect-ratio: 2 / 1; border-radius: var(--r-sm); }
  .tt { min-width: 0; }

  h3 {
    font: 700 16.5px/1.4 var(--font-serif);
    margin: 0 0 3px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .pin-i { color: var(--ink); }

  small {
    font: 12.5px var(--font-mono);
    color: var(--st-ink-3);
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tags { display: flex; gap: 8px; font-size: 13px; overflow: hidden; }
  .num { font: 500 12.5px var(--font-mono); color: var(--st-ink-3); text-align: right; }
}

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
  .pgrid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
</style>
