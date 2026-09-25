<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { api, type Note, type Post, type Tag } from '../../api';
import CoverArt from '../common/CoverArt.vue';
import UiIcon from '../ui/UiIcon.vue';
import { closeSearch, openSearch, searchPalette } from './palette';

/**
 * ⌘K 全局搜索浮层（仅桌面外壳；移动端用 SearchOverlay）。
 * 空输入：最近搜索（localStorage）/ 标签 / 跳转；输入后分组：标签 / 文章 / 随想，关键词高亮。
 * 键盘：Ctrl/⌘+K 开关，↑↓ 选择，Enter 打开，Esc 关闭。
 */
const { t } = useI18n();
const router = useRouter();

type Item =
  | { kind: 'recent'; key: string; term: string }
  | { kind: 'tag'; key: string; tag: Tag }
  | { kind: 'jump'; key: string; to: string; label: string; icon: string }
  | { kind: 'post'; key: string; post: Post }
  | { kind: 'note'; key: string; note: Note };

interface Section {
  key: string;
  title: string;
  layout: 'rows' | 'chips';
  clearable?: boolean;
  items: Item[];
}

const RECENT_KEY = 'myself.search.recent';
const RECENT_MAX = 6;
const TAG_MAX = 12;

const input = ref<HTMLInputElement | null>(null);
const body = ref<HTMLElement | null>(null);
const q = ref('');
const recent = ref<string[]>(loadRecent());
const allTags = ref<Tag[]>([]);
const posts = ref<Post[]>([]);
const notes = ref<Note[]>([]);
const searching = ref(false);
const searched = ref('');
const sel = ref(0);
let restoreFocus: HTMLElement | null = null;

/* ---------- 最近搜索 ---------- */
function loadRecent(): string[] {
  try {
    const raw = localStorage.getItem(RECENT_KEY);
    return raw ? (JSON.parse(raw) as string[]).slice(0, RECENT_MAX) : [];
  } catch {
    return [];
  }
}

function remember(term: string): void {
  const v = term.trim();
  if (!v) return;
  recent.value = [v, ...recent.value.filter((x) => x !== v)].slice(0, RECENT_MAX);
  try {
    localStorage.setItem(RECENT_KEY, JSON.stringify(recent.value));
  } catch { /* 隐私模式忽略 */ }
}

function clearRecent(): void {
  recent.value = [];
  try {
    localStorage.removeItem(RECENT_KEY);
  } catch { /* 忽略 */ }
  input.value?.focus();
}

/* ---------- 检索 ---------- */
let timer = 0;
let seq = 0;
watch(q, (v) => {
  window.clearTimeout(timer);
  sel.value = 0;
  const term = v.trim();
  if (!term) {
    seq += 1;
    posts.value = [];
    notes.value = [];
    searched.value = '';
    searching.value = false;
    return;
  }
  searching.value = true;
  timer = window.setTimeout(() => void run(term), 200);
});

async function run(term: string): Promise<void> {
  const my = ++seq;
  try {
    const [p, n] = await Promise.all([
      api.posts({ q: term, pageSize: 6 }),
      api.notes({ q: term, pageSize: 5 }),
    ]);
    if (my !== seq) return;
    posts.value = p.items;
    notes.value = n.items;
  } catch {
    if (my !== seq) return;
    posts.value = [];
    notes.value = [];
  } finally {
    if (my === seq) {
      searched.value = term;
      searching.value = false;
      sel.value = 0;
    }
  }
}

const term = computed(() => q.value.trim());

const matchedTags = computed(() => {
  const k = term.value.toLowerCase();
  return k ? allTags.value.filter((tag) => tag.name.toLowerCase().includes(k)) : [];
});

const sections = computed<Section[]>(() => {
  if (!term.value) {
    const out: Section[] = [];
    if (recent.value.length) {
      out.push({
        key: 'recent',
        title: t('search.recent'),
        layout: 'rows',
        clearable: true,
        items: recent.value.map((r) => ({ kind: 'recent', key: `r:${r}`, term: r })),
      });
    }
    if (allTags.value.length) {
      out.push({
        key: 'tags',
        title: t('search.tags'),
        layout: 'chips',
        items: allTags.value.slice(0, TAG_MAX).map((tag) => ({ kind: 'tag', key: `t:${tag.name}`, tag })),
      });
    }
    out.push({
      key: 'jump',
      title: t('search.jump'),
      layout: 'rows',
      items: [
        { kind: 'jump', key: 'j:articles', to: '/articles', label: t('nav.articles'), icon: 'book' },
        { kind: 'jump', key: 'j:thoughts', to: '/thoughts', label: t('nav.thoughts'), icon: 'bubble' },
        { kind: 'jump', key: 'j:about', to: '/about', label: t('nav.about'), icon: 'user' },
      ],
    });
    return out;
  }
  const out: Section[] = [];
  if (matchedTags.value.length) {
    out.push({
      key: 'tags',
      title: t('search.tags'),
      layout: 'chips',
      items: matchedTags.value.map((tag) => ({ kind: 'tag', key: `t:${tag.name}`, tag })),
    });
  }
  if (posts.value.length) {
    out.push({
      key: 'posts',
      title: `${t('search.posts')} · ${posts.value.length}`,
      layout: 'rows',
      items: posts.value.map((post) => ({ kind: 'post', key: `p:${post.id}`, post })),
    });
  }
  if (notes.value.length) {
    out.push({
      key: 'notes',
      title: `${t('search.notes')} · ${notes.value.length}`,
      layout: 'rows',
      items: notes.value.map((note) => ({ kind: 'note', key: `n:${note.id}`, note })),
    });
  }
  return out;
});

/** 各分组在扁平列表中的起始下标（键盘导航用） */
const offsets = computed(() => {
  let n = 0;
  return sections.value.map((s) => {
    const o = n;
    n += s.items.length;
    return o;
  });
});
const flat = computed(() => sections.value.flatMap((s) => s.items));

/** 首次检索未返回前显示骨架；再次输入时保留旧结果并调暗 */
const showSkeleton = computed(() => searching.value && !flat.value.length);
const showEmpty = computed(() => !!term.value && !searching.value && !!searched.value && !flat.value.length);

/* ---------- 文本处理 ---------- */
function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

/** 纯文本高亮：先转义，再把关键词包进 <mark> */
function hl(text: string): string {
  const safe = escapeHtml(text);
  if (!term.value) return safe;
  const pattern = escapeHtml(term.value).replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return safe.replace(new RegExp(pattern, 'gi'), (m) => `<mark>${m}</mark>`);
}

function plain(md: string): string {
  return md
    .replace(/!\[[^\]]*]\([^)]*\)/g, '')
    .replace(/\[([^\]]*)]\([^)]*\)/g, '$1')
    .replace(/[#>*_`~]/g, '')
    .replace(/\s+/g, ' ')
    .trim();
}

/** 截取关键词附近的片段 */
function snippet(md: string): string {
  const text = plain(md);
  const i = text.toLowerCase().indexOf(term.value.toLowerCase());
  if (i < 30) return text.slice(0, 80);
  return `…${text.slice(i - 20, i + 60)}`;
}

function monthDay(iso: string): string {
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? '' : `${d.getMonth() + 1}月${d.getDate()}日`;
}

/* ---------- 动作 ---------- */
function go(to: string | { path: string; query?: Record<string, string> }): void {
  closeSearch();
  void router.push(to);
}

function activate(item: Item | undefined): void {
  if (!item) {
    remember(q.value);
    return;
  }
  switch (item.kind) {
    case 'recent':
      q.value = item.term;
      input.value?.focus();
      break;
    case 'tag':
      remember(q.value || item.tag.name);
      go({ path: '/articles', query: { tag: item.tag.name } });
      break;
    case 'jump':
      go(item.to);
      break;
    case 'post':
      remember(q.value);
      go(`/articles/${item.post.slug}`);
      break;
    case 'note':
      remember(q.value);
      go('/thoughts');
      break;
  }
}

function move(delta: number): void {
  const n = flat.value.length;
  if (!n) return;
  sel.value = (sel.value + delta + n) % n;
  void nextTick(() => {
    body.value?.querySelector<HTMLElement>('[data-sel="true"]')?.scrollIntoView({ block: 'nearest' });
  });
}

function onInputKey(e: KeyboardEvent): void {
  if (e.isComposing) return;
  if (e.key === 'ArrowDown') {
    e.preventDefault();
    move(1);
  } else if (e.key === 'ArrowUp') {
    e.preventDefault();
    move(-1);
  } else if (e.key === 'Enter') {
    e.preventDefault();
    activate(flat.value[sel.value]);
  }
}

/* 滚轮只滚结果列表，不穿透到页面（不锁 html 滚动，避免滚动条槽位露出未模糊的底色） */
function onWheel(e: WheelEvent): void {
  const list = body.value;
  const inList = !!list && e.target instanceof Node && list.contains(e.target);
  if (!inList || list.scrollHeight <= list.clientHeight) e.preventDefault();
}

/* ---------- 开关 ---------- */
function onGlobalKey(e: KeyboardEvent): void {
  if ((e.ctrlKey || e.metaKey) && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'k') {
    e.preventDefault();
    if (searchPalette.open) closeSearch();
    else openSearch();
    return;
  }
  if (searchPalette.open && e.key === 'Escape') {
    e.preventDefault();
    closeSearch();
  }
}

watch(
  () => searchPalette.open,
  async (on) => {
    if (!on) {
      window.clearTimeout(timer);
      restoreFocus?.focus({ preventScroll: true });
      restoreFocus = null;
      return;
    }
    restoreFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    recent.value = loadRecent();
    q.value = searchPalette.seed;
    sel.value = 0;
    if (!allTags.value.length) {
      api.tags()
        .then((list) => { allTags.value = [...list].sort((a, b) => b.count - a.count); })
        .catch(() => undefined);
    }
    await nextTick();
    input.value?.focus({ preventScroll: true });
    input.value?.select();
  },
);

onMounted(() => window.addEventListener('keydown', onGlobalKey));
onBeforeUnmount(() => {
  window.removeEventListener('keydown', onGlobalKey);
  window.clearTimeout(timer);
  searchPalette.open = false;
});
</script>

<template>
  <Teleport to="body">
    <Transition name="kk" :duration="{ enter: 460, leave: 220 }">
      <div v-if="searchPalette.open" class="kk" role="dialog" @wheel="onWheel" aria-modal="true" :aria-label="t('search.label')">
        <div class="kk-bg" @click="closeSearch" />
        <div class="kk-panel">
          <label class="kk-head">
            <UiIcon name="search" class="kk-head-ic" />
            <input
              ref="input"
              v-model="q"
              type="text"
              :placeholder="t('search.placeholder')"
              autocomplete="off"
              spellcheck="false"
              role="combobox"
              aria-expanded="true"
              aria-controls="kk-list"
              @keydown="onInputKey"
            />
            <button class="kk-esc" type="button" :aria-label="t('search.close')" @click="closeSearch">
              <kbd>Esc</kbd>
            </button>
          </label>

          <div id="kk-list" ref="body" class="kk-body" role="listbox">
            <!-- 首次检索中：同形骨架 -->
            <div v-if="showSkeleton" class="kk-sk">
              <div class="kk-h"><span class="sk sk-line" style="width: 64px" /></div>
              <div v-for="i in 3" :key="i" class="kk-row is-sk">
                <span class="kk-th sk" />
                <span class="kk-t">
                  <span class="sk sk-line" :style="{ width: `${72 - i * 12}%`, '--lh': '1.4em' }" />
                  <span class="sk sk-line" :style="{ width: `${88 - i * 8}%`, fontSize: '12px' }" />
                </span>
              </div>
            </div>

            <!-- 无结果 -->
            <div v-else-if="showEmpty" class="kk-empty rise">
              <div class="kk-ring"><UiIcon name="search" /></div>
              <b>{{ t('search.emptyTitle', { q: searched }) }}</b>
              <p>{{ t('search.emptyTip') }}</p>
            </div>

            <!-- 分组 -->
            <div v-else class="kk-groups" :class="{ dim: searching }">
              <section v-for="(s, si) in sections" :key="s.key" class="kk-sec">
                <div class="kk-h">
                  <span>{{ s.title }}</span>
                  <button v-if="s.clearable" type="button" @click="clearRecent">{{ t('search.clear') }}</button>
                </div>

                <div v-if="s.layout === 'chips'" class="kk-chips">
                  <template v-for="(it, ii) in s.items" :key="it.key">
                    <button
                      v-if="it.kind === 'tag'"
                      type="button"
                      class="kk-chip"
                      :class="{ sel: sel === offsets[si] + ii, hit: !!term }"
                      role="option"
                      :aria-selected="sel === offsets[si] + ii"
                      :data-sel="sel === offsets[si] + ii"
                      @pointermove="sel = offsets[si] + ii"
                      @click="activate(it)"
                    >
                      <span class="hash">#</span>
                      <!-- eslint-disable-next-line vue/no-v-html -->
                      <span v-html="hl(it.tag.name)" />
                      <em>{{ it.tag.count }}</em>
                    </button>
                  </template>
                </div>

                <template v-else>
                  <button
                    v-for="(it, ii) in s.items"
                    :key="it.key"
                    type="button"
                    class="kk-row"
                    :class="{ sel: sel === offsets[si] + ii }"
                    role="option"
                    :aria-selected="sel === offsets[si] + ii"
                    :data-sel="sel === offsets[si] + ii"
                    @pointermove="sel = offsets[si] + ii"
                    @click="activate(it)"
                  >
                    <template v-if="it.kind === 'recent'">
                      <span class="kk-ic"><UiIcon name="history" class="s" /></span>
                      <span class="kk-t kk-plain">{{ it.term }}</span>
                    </template>
                    <template v-else-if="it.kind === 'jump'">
                      <span class="kk-ic"><UiIcon :name="it.icon" class="s" /></span>
                      <span class="kk-t kk-plain">{{ it.label }}</span>
                    </template>
                    <template v-else-if="it.kind === 'post'">
                      <span class="kk-th"><CoverArt :src="it.post.covers[0]" :seed="it.post.slug" thumb /></span>
                      <span class="kk-t">
                        <!-- eslint-disable-next-line vue/no-v-html -->
                        <b v-html="hl(it.post.title)" />
                        <small>
                          <template v-if="it.post.tags[0]">{{ it.post.tags[0] }} · </template>{{ monthDay(it.post.createdAt) }}
                          <!-- eslint-disable-next-line vue/no-v-html -->
                          · <span v-html="hl(it.post.excerpt)" />
                        </small>
                      </span>
                    </template>
                    <template v-else-if="it.kind === 'note'">
                      <span class="kk-ic"><UiIcon name="bubble" class="s" /></span>
                      <span class="kk-t">
                        <!-- eslint-disable-next-line vue/no-v-html -->
                        <span class="kk-note" v-html="hl(snippet(it.note.contentMd))" />
                        <small>{{ monthDay(it.note.createdAt) }}<template v-if="it.note.mood"> · # {{ it.note.mood }}</template></small>
                      </span>
                    </template>
                    <UiIcon name="enter" class="s kk-go" />
                  </button>
                </template>
              </section>
            </div>
          </div>

          <div class="kk-foot">
            <span><kbd>↑</kbd><kbd>↓</kbd>{{ t('search.select') }}</span>
            <span><kbd>Enter</kbd>{{ t('search.open') }}</span>
            <span><kbd>Esc</kbd>{{ t('search.close') }}</span>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped lang="scss">
.kk {
  position: fixed;
  inset: 0;
  z-index: 9500;
}

.kk-bg {
  position: absolute;
  inset: 0;
  background: var(--scrim);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  transition: opacity 0.3s var(--ease-out);
}

.kk-panel {
  position: absolute;
  left: 50%;
  top: 12vh;
  width: min(640px, calc(100% - 64px));
  max-height: min(72vh, 640px);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: var(--r-xl);
  background: color-mix(in oklab, var(--surface) 94%, transparent);
  backdrop-filter: blur(30px) saturate(170%);
  -webkit-backdrop-filter: blur(30px) saturate(170%);
  box-shadow: var(--shadow-pop);
  transform: translateX(-50%);
  transform-origin: 50% 0;
  transition: transform 0.45s var(--ease-spring), opacity 0.2s var(--ease-out);
}

:root[data-mode='light'] .kk-panel { background: rgb(255 255 255 / 0.94); }

/* 进出场：背景淡入 + 面板 scale .98→1（离场更快、无回弹） */
.kk-enter-from,
.kk-leave-to {
  .kk-bg { opacity: 0; }
  .kk-panel { opacity: 0; transform: translateX(-50%) translateY(-8px) scale(0.98); }
}

.kk-leave-active {
  .kk-bg { transition-duration: 0.2s; }
  .kk-panel { transition: transform 0.2s var(--ease-out), opacity 0.16s ease; }
}

/* ---------- 输入 ---------- */
.kk-head {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 14px 14px 20px;
  box-shadow: inset 0 -0.5px 0 var(--line);
  cursor: text;

  input {
    flex: 1;
    min-width: 0;
    border: 0;
    outline: 0;
    background: none;
    font: inherit;
    font-size: 17px;
    color: var(--text);
    caret-color: var(--ink);

    &::placeholder { color: var(--text-3); }
    &:focus-visible { outline: none; }
  }
}

.kk-head-ic { color: var(--text-3); }

.kk-esc {
  display: inline-grid;
  place-items: center;
  height: 30px;
  padding: 0 6px;
  border: 0;
  border-radius: var(--r-sm);
  background: none;
  transition: background-color var(--dur-fast);

  &:hover { background: var(--fill-2); }
}

kbd {
  display: inline-grid;
  place-items: center;
  min-width: 20px;
  height: 20px;
  padding: 0 5px;
  border-radius: var(--r-xs);
  font-family: var(--font-mono);
  font-size: 11px;
  line-height: 1;
  color: var(--text-2);
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line-2), inset 0 -1px 0 var(--line);
}

/* ---------- 列表 ---------- */
.kk-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 6px 8px 10px;
  scrollbar-width: thin;
}

.kk-groups { transition: opacity var(--dur-fast); }
.kk-groups.dim { opacity: 0.55; }

.kk-h {
  display: flex;
  align-items: baseline;
  padding: 14px 12px 8px;
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.06em;
  color: var(--text-3);

  button {
    margin-left: auto;
    border: 0;
    background: none;
    font-size: 12px;
    letter-spacing: 0;
    color: var(--ink);
    border-radius: var(--r-xs);

    &:hover { text-decoration: underline; text-underline-offset: 3px; }
  }
}

.kk-row {
  position: relative;
  width: 100%;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 10px 12px;
  border: 0;
  border-radius: var(--r-md);
  background: none;
  color: var(--text);
  text-align: left;
  transition: background-color var(--dur-fast), box-shadow var(--dur-fast);

  &.sel {
    background: var(--lift);
    box-shadow: var(--lift-shadow);
  }

  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

.kk-ic {
  flex: none;
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border-radius: var(--r-sm);
  background: var(--fill);
  color: var(--text-3);
  transition: color var(--dur-fast);

  .sel & { color: var(--text-2); }
}

.kk-th {
  position: relative;
  flex: none;
  width: 60px;
  height: 44px;
  overflow: hidden;
  border-radius: var(--r-sm);
  isolation: isolate;

  :deep(.cv) { position: absolute; inset: 0; }
}

.kk-t {
  flex: 1;
  min-width: 0;

  b {
    display: block;
    font-family: var(--font-serif);
    font-size: 16px;
    font-weight: 700;
    line-height: 1.4;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  small {
    display: block;
    margin-top: 2px;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    font-size: 13px;
    color: var(--text-3);
  }
}

.kk-plain { font-size: 15px; font-weight: 500; }

.kk-note {
  display: block;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-size: 15px;
  font-weight: 400;
  line-height: 1.5;
}

.kk-go {
  flex: none;
  color: var(--text-3);
  opacity: 0;
  transform: translateX(-3px);
  transition: opacity var(--dur-fast), transform var(--dur-fast) var(--ease-out);

  .sel & { opacity: 1; transform: none; }
}

/* 关键词：--ink 字 + 底部轻染（唯一使用 --tint 的地方之一） */
:deep(mark) {
  background: none;
  color: var(--ink);
  box-shadow: inset 0 -0.38em 0 var(--tint);
  border-radius: calc(var(--r-base) * 0.2px);
}

/* ---------- 标签胶囊 ---------- */
.kk-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 2px 12px 6px;
}

.kk-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 30px;
  padding: 0 12px;
  border: 0;
  border-radius: var(--r-pill);
  background: none;
  box-shadow: inset 0 0 0 0.5px var(--line-2);
  font-size: 13px;
  color: var(--text-2);
  transition: background-color var(--dur-fast), color var(--dur-fast), box-shadow var(--dur-fast);

  .hash { color: var(--text-3); }

  em {
    margin-left: 4px;
    font-style: normal;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-3);
  }

  &.hit { color: var(--text); }

  &.sel {
    color: var(--lift-fg);
    background: var(--lift);
    box-shadow: var(--lift-shadow);
  }

  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

/* ---------- 骨架 / 空态 ---------- */
.kk-row.is-sk { pointer-events: none; }
.kk-row.is-sk .kk-t { display: flex; flex-direction: column; gap: 6px; }

.kk-empty {
  padding: 48px 20px 56px;
  text-align: center;
  font-size: 13.5px;
  color: var(--text-3);

  b {
    display: block;
    margin-bottom: 4px;
    font-size: 15px;
    font-weight: 500;
    color: var(--text);
  }
}

.kk-ring {
  width: 56px;
  height: 56px;
  margin: 0 auto 14px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background: var(--fill);
  color: var(--text-3);
}

/* ---------- 底栏提示 ---------- */
.kk-foot {
  display: flex;
  gap: 16px;
  padding: 10px 20px;
  font-size: 12px;
  color: var(--text-3);
  box-shadow: inset 0 0.5px 0 var(--line);

  span {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  kbd + kbd { margin-left: -2px; }
}
</style>
