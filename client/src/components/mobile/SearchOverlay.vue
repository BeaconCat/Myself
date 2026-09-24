<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { api, type Note, type Post, type Tag } from '../../api';
import CoverArt from './CoverArt.vue';
import MIcon from './MIcon.vue';
import { monthDay, shell } from './shell';

/**
 * 全屏搜索：输入框从底栏搜索钮位置飞入，背景毛玻璃；
 * 空输入显示最近搜索（localStorage）与标签，输入后分组展示 标签 / 文章 / 随想，关键词高亮。
 */
const { t } = useI18n();
const router = useRouter();

const RECENT_KEY = 'myself.m.recent';
const input = ref<HTMLInputElement | null>(null);
const q = ref('');
const recent = ref<string[]>(loadRecent());
const allTags = ref<Tag[]>([]);
const posts = ref<Post[]>([]);
const notes = ref<Note[]>([]);
const searching = ref(false);
const searched = ref('');

function loadRecent(): string[] {
  try {
    const raw = localStorage.getItem(RECENT_KEY);
    return raw ? (JSON.parse(raw) as string[]).slice(0, 8) : [];
  } catch {
    return [];
  }
}

function remember(term: string): void {
  const v = term.trim();
  if (!v) return;
  recent.value = [v, ...recent.value.filter((x) => x !== v)].slice(0, 8);
  try {
    localStorage.setItem(RECENT_KEY, JSON.stringify(recent.value));
  } catch { /* 隐私模式忽略 */ }
}

function clearRecent(): void {
  recent.value = [];
  try {
    localStorage.removeItem(RECENT_KEY);
  } catch { /* 忽略 */ }
}

watch(
  () => shell.searchOpen,
  async (on) => {
    if (!on) {
      input.value?.blur();
      return;
    }
    q.value = shell.searchSeed;
    if (!allTags.value.length) {
      api.tags().then((list) => { allTags.value = list; }).catch(() => undefined);
    }
    await nextTick();
    window.setTimeout(() => input.value?.focus({ preventScroll: true }), 320);
  },
);

let timer = 0;
let seq = 0;
watch(q, (v) => {
  window.clearTimeout(timer);
  const term = v.trim();
  if (!term) {
    posts.value = [];
    notes.value = [];
    searched.value = '';
    searching.value = false;
    return;
  }
  searching.value = true;
  timer = window.setTimeout(() => void run(term), 240);
});

async function run(term: string): Promise<void> {
  const my = ++seq;
  try {
    const [p, n] = await Promise.all([
      api.posts({ q: term, pageSize: 8 }),
      api.notes({ q: term, pageSize: 8 }),
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
    }
  }
}

const matchedTags = computed(() => {
  const term = q.value.trim().toLowerCase();
  if (!term) return [];
  return allTags.value.filter((tag) => tag.name.toLowerCase().includes(term));
});

const hasResults = computed(() => matchedTags.value.length + posts.value.length + notes.value.length > 0);

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

/** 纯文本高亮：先转义，再把关键词包进 <mark> */
function hl(text: string): string {
  const safe = escapeHtml(text);
  const term = q.value.trim();
  if (!term) return safe;
  const pattern = escapeHtml(term).replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
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
  const i = text.toLowerCase().indexOf(q.value.trim().toLowerCase());
  if (i < 40) return text.slice(0, 90);
  return `…${text.slice(i - 24, i + 66)}`;
}

function close(): void {
  shell.searchOpen = false;
}

function submit(): void {
  remember(q.value);
  input.value?.blur();
}

function openPost(p: Post): void {
  remember(q.value);
  close();
  void router.push(`/articles/${p.slug}`);
}

function openTag(name: string): void {
  remember(q.value || name);
  close();
  void router.push({ path: '/articles', query: { tag: name } });
}

function openNote(): void {
  remember(q.value);
  close();
  void router.push('/thoughts');
}
</script>

<template>
  <div class="search" :class="{ on: shell.searchOpen, 'has-q': !!q }" :aria-hidden="!shell.searchOpen">
    <div class="bg" @click="close" />
    <header class="head">
      <label class="field">
        <MIcon name="search" class="s" />
        <input
          ref="input"
          v-model="q"
          type="search"
          :placeholder="t('mobile.search.placeholder')"
          enterkeyhint="search"
          autocomplete="off"
          @keydown.enter="submit"
        />
        <button class="clear" type="button" aria-label="clear" @click="q = ''; input?.focus()">
          <MIcon name="close" />
        </button>
      </label>
      <button class="cancel m-tap" @click="close">{{ t('mobile.search.cancel') }}</button>
    </header>

    <div class="body">
      <!-- 空输入：最近搜索 + 标签 -->
      <template v-if="!q.trim()">
        <template v-if="recent.length">
          <div class="h">{{ t('mobile.search.recent') }}<button @click="clearRecent">{{ t('mobile.search.clear') }}</button></div>
          <button v-for="r in recent" :key="r" class="recent" @click="q = r">
            <MIcon name="history" /><span>{{ r }}</span><MIcon name="chev" class="xs" />
          </button>
        </template>
        <template v-if="allTags.length">
          <div class="h">{{ t('mobile.search.tags') }}</div>
          <div class="chips">
            <button v-for="tag in allTags" :key="tag.name" class="chip m-tap" @click="openTag(tag.name)">
              # {{ tag.name }}<em>{{ tag.count }}</em>
            </button>
          </div>
        </template>
      </template>

      <!-- 结果 -->
      <template v-else-if="hasResults || searching">
        <div :class="{ dim: searching }" class="results">
          <template v-if="matchedTags.length">
            <div class="h">{{ t('mobile.search.tags') }}</div>
            <div class="chips">
              <button v-for="tag in matchedTags" :key="tag.name" class="chip on m-tap" @click="openTag(tag.name)">
                <!-- eslint-disable-next-line vue/no-v-html -->
                # <span v-html="hl(tag.name)" /><em>{{ tag.count }}</em>
              </button>
            </div>
          </template>
          <template v-if="posts.length">
            <div class="h">{{ t('mobile.search.posts') }} · {{ posts.length }}</div>
            <button v-for="(p, i) in posts" :key="p.id" class="m-arow m-in" :style="{ '--i': i }" @click="openPost(p)">
              <div class="rt">
                <div class="meta"><em>{{ p.tags[0] }}</em> · {{ monthDay(p.createdAt).m }}月{{ monthDay(p.createdAt).d }}日</div>
                <!-- eslint-disable-next-line vue/no-v-html -->
                <b v-html="hl(p.title)" />
                <!-- eslint-disable-next-line vue/no-v-html -->
                <span class="ex" v-html="hl(p.excerpt)" />
              </div>
              <div class="thumb"><CoverArt :src="p.covers[0]" :seed="p.slug" thumb /></div>
            </button>
          </template>
          <template v-if="notes.length">
            <div class="h">{{ t('mobile.search.notes') }} · {{ notes.length }}</div>
            <button v-for="(n, i) in notes" :key="n.id" class="note m-in" :style="{ '--i': i + posts.length }" @click="openNote">
              <!-- eslint-disable-next-line vue/no-v-html -->
              <span v-html="hl(snippet(n.contentMd))" />
              <small>{{ monthDay(n.createdAt).m }}月{{ monthDay(n.createdAt).d }}日<template v-if="n.mood"> · #{{ n.mood }}</template></small>
            </button>
          </template>
        </div>
      </template>

      <div v-else-if="searched" class="empty">
        <div class="ring"><MIcon name="search" /></div>
        <b>{{ t('mobile.search.emptyTitle') }}</b>
        <p>{{ t('mobile.search.emptyTip') }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.search {
  position: fixed;
  z-index: 120;
  inset: 0;
  visibility: hidden;
  pointer-events: none;
  transition: visibility 0s linear 0.5s;

  &.on {
    visibility: visible;
    pointer-events: auto;
    transition-delay: 0s;
  }
}

.bg {
  position: absolute;
  inset: 0;
  background: color-mix(in oklab, var(--bg) 84%, transparent);
  backdrop-filter: blur(30px) saturate(160%);
  -webkit-backdrop-filter: blur(30px) saturate(160%);
  opacity: 0;
  transition: opacity 0.4s var(--ease-out);

  .on & { opacity: 1; }
}

.head {
  position: absolute;
  top: calc(var(--m-safe-t) + 12px);
  left: 16px;
  right: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  opacity: 0;
  transform: translateY(70vh) scale(0.6);
  transform-origin: right center;
  transition: transform 0.55s var(--ease-spring), opacity 0.25s;

  .on & {
    opacity: 1;
    transform: none;
  }
}

.field {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  height: 44px;
  padding: 0 12px;
  border-radius: 14px;
  background: var(--m-fill-2);
  color: var(--m-text-3);

  input {
    flex: 1;
    min-width: 0;
    font: inherit;
    font-size: 16px;
    color: var(--text);
    background: none;
    border: 0;
    outline: none;
    caret-color: var(--primary);
    user-select: text;
    -webkit-user-select: text;

    &::placeholder { color: var(--m-text-3); }
    &::-webkit-search-cancel-button { display: none; }
  }
}

.clear {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: var(--m-text-3) !important;
  color: var(--bg) !important;
  opacity: 0;
  pointer-events: none;
  transition: opacity var(--dur-fast);

  .m-ic { width: 12px; height: 12px; stroke-width: 2.4; }

  .has-q & {
    opacity: 1;
    pointer-events: auto;
  }
}

.cancel {
  font-size: 16px;
  color: var(--m-ink) !important;
}

.body {
  position: absolute;
  top: calc(var(--m-safe-t) + 66px);
  left: 0;
  right: 0;
  bottom: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-width: none;
  padding: 6px 0 calc(var(--m-safe-b) + 60px);
  opacity: 0;
  transform: translateY(12px);
  transition: opacity 0.35s var(--ease-out) 0.1s, transform 0.45s var(--ease-out) 0.1s;

  .on & {
    opacity: 1;
    transform: none;
  }
}

.h {
  display: flex;
  align-items: baseline;
  padding: 14px 20px 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--m-text-3);
  letter-spacing: 0.04em;

  button {
    margin-left: auto;
    font-size: 13px;
    font-weight: 400;
    color: var(--m-ink) !important;
  }
}

.recent {
  position: relative;
  width: 100%;
  display: flex;
  align-items: center;
  gap: 12px;
  height: 48px;
  padding: 0 20px;
  font-size: 15.5px;
  text-align: left;

  .m-ic { color: var(--m-text-3); width: 18px; height: 18px; }
  span { flex: 1; }

  & + &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 50px;
    right: 0;
    height: 0.5px;
    background: var(--m-line);
  }

  &:active { background: var(--m-fill); }
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 4px 20px;
}

.chip {
  height: 34px;
  padding: 0 14px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: var(--text-2);
  background: var(--m-fill);
  box-shadow: inset 0 0 0 0.5px var(--m-line);

  em {
    font-style: normal;
    font-family: var(--m-font-mono);
    font-size: 11px;
    opacity: 0.6;
    margin-left: 2px;
  }

  &.on {
    background: var(--m-soft);
    color: var(--m-ink);
    box-shadow: none;
    font-weight: 500;
  }
}

.results {
  transition: opacity var(--dur) var(--ease-out);

  &.dim { opacity: 0.5; }

  .m-arow { width: calc(100% - 8px); margin: 0 4px; }
}

.note {
  position: relative;
  width: 100%;
  display: block;
  padding: 12px 20px;
  text-align: left;
  font-size: 14.5px;
  line-height: 1.65;

  small {
    display: block;
    margin-top: 4px;
    font-size: 12px;
    color: var(--m-text-3);
  }

  & + &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 20px;
    right: 0;
    height: 0.5px;
    background: var(--m-line);
  }

  &:active { background: var(--m-fill); }
}

.search :deep(mark) {
  background: none;
  color: var(--m-ink);
  font-weight: 700;
  box-shadow: inset 0 -0.35em 0 var(--m-soft);
}

.empty {
  padding: 70px 40px;
  text-align: center;

  .ring {
    width: 64px;
    height: 64px;
    margin: 0 auto 18px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: var(--m-fill);
    color: var(--m-text-3);
  }

  b {
    display: block;
    font-size: 16px;
  }

  p {
    margin-top: 6px;
    font-size: 13.5px;
    color: var(--m-text-3);
    line-height: 1.7;
  }
}
</style>
