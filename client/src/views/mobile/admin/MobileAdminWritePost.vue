<script setup lang="ts">
/**
 * 写文章 · 全屏编辑器：封面、标题、标签、Markdown 正文 textarea；
 * 键盘上方悬浮 Markdown 工具条（跟随 visualViewport，点按不收起键盘）；
 * 元信息 sheet（slug / 标签 / 摘要 / 封面 / 状态 / 置顶）；?id= 编辑，新文章本机自动保存草稿。
 */
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, thumbOf, type PostDraft } from '../../../api';
import { useDialogStore } from '../../../stores/dialog';
import MaIcon from '../../../components/mobile-admin/MaIcon.vue';
import MaRing from '../../../components/mobile-admin/MaRing.vue';
import MaSkeleton from '../../../components/mobile-admin/MaSkeleton.vue';
import PostMetaSheet from '../../../components/mobile-admin/PostMetaSheet.vue';
import type { IconName } from '../../../components/mobile-admin/icons';
import { shell, toast } from '../../../components/mobile-admin/state';
import { SLUG_RE, toSlug } from '../../../components/mobile-admin/format';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const dialog = useDialogStore();

const id = computed(() => {
  const raw = route.query.id;
  const n = typeof raw === 'string' && raw ? Number(raw) : NaN;
  return Number.isFinite(n) ? n : null;
});
const LOCAL_KEY = 'myself.m.draft.new';

const empty = (): PostDraft => ({
  slug: '',
  title: '',
  excerpt: '',
  contentMd: '',
  covers: [],
  tags: [],
  status: 'published',
  pinned: false,
});
const draft = ref<PostDraft>(empty());
const loading = ref(id.value !== null);
const busy = ref<'' | 'publish' | 'draft'>('');
const metaOpen = ref(false);
const savedAt = ref('');
let snapshot = JSON.stringify(draft.value);
let leaving = false;
const origStatus = ref<'published' | 'draft' | null>(null);

const dirty = computed(() => JSON.stringify(draft.value) !== snapshot);
const words = computed(() => draft.value.contentMd.replace(/\s/g, '').length);
const primaryLabel = computed(() => {
  if (busy.value === 'publish') return t('mobileAdmin.common.saving');
  if (draft.value.status === 'draft') return t('mobileAdmin.common.save');
  if (id.value !== null && origStatus.value === 'published') return t('mobileAdmin.post.update');
  return t('mobileAdmin.post.publish');
});

const titleEl = ref<HTMLTextAreaElement | null>(null);
const bodyEl = ref<HTMLTextAreaElement | null>(null);

function autosize(el: HTMLTextAreaElement | null, min: number): void {
  if (!el) return;
  el.style.height = 'auto';
  el.style.height = `${Math.max(min, el.scrollHeight)}px`;
}
const fit = (): void => {
  autosize(titleEl.value, 40);
  autosize(bodyEl.value, 260);
};

onMounted(async () => {
  if (id.value !== null) {
    try {
      const p = await adminApi.post(id.value);
      draft.value = {
        slug: p.slug,
        title: p.title,
        excerpt: p.excerpt,
        contentMd: p.contentMd ?? '',
        covers: [...p.covers],
        tags: [...p.tags],
        status: p.status,
        pinned: p.pinned,
      };
      origStatus.value = p.status;
    } catch {
      toast(t('mobileAdmin.post.loadFailed'), '', 'error');
    }
    loading.value = false;
  } else {
    try {
      const raw = localStorage.getItem(LOCAL_KEY);
      if (raw) {
        const saved = JSON.parse(raw) as PostDraft;
        if (saved.title || saved.contentMd) {
          draft.value = { ...empty(), ...saved };
          window.setTimeout(() => toast(t('mobileAdmin.post.restored'), saved.title || ''), 450);
        }
      }
    } catch {
      /* 本机草稿损坏时忽略 */
    }
  }
  snapshot = JSON.stringify(draft.value);
  await nextTick();
  fit();
});

/* 新文章：本机自动保存 */
let autosaveTimer = 0;
watch(
  draft,
  () => {
    if (id.value !== null || loading.value) return;
    window.clearTimeout(autosaveTimer);
    autosaveTimer = window.setTimeout(() => {
      try {
        localStorage.setItem(LOCAL_KEY, JSON.stringify(draft.value));
        const d = new Date();
        savedAt.value = `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
      } catch {
        /* 存储不可用时静默 */
      }
    }, 700);
  },
  { deep: true },
);

/* ---------- 保存 / 发布 ---------- */
async function save(mode: 'publish' | 'draft'): Promise<void> {
  if (busy.value) return;
  const d = draft.value;
  if (!d.title.trim()) {
    toast(t('mobileAdmin.post.needTitle'), '', 'error');
    titleEl.value?.focus();
    return;
  }
  if (!d.slug) d.slug = toSlug(d.title);
  if (!SLUG_RE.test(d.slug)) {
    toast(t('mobileAdmin.post.slugInvalid'), '', 'error');
    metaOpen.value = true;
    return;
  }
  busy.value = mode;
  /* 主按钮按元信息里的状态保存；「存草稿」强制草稿 */
  const body: PostDraft = { ...d, status: mode === 'draft' ? 'draft' : d.status };
  try {
    if (id.value === null) {
      await adminApi.createPost(body);
      localStorage.removeItem(LOCAL_KEY);
    } else {
      await adminApi.updatePost(id.value, body);
    }
    snapshot = JSON.stringify(draft.value);
    shell.bump.posts += 1;
    const msg =
      body.status === 'draft'
        ? t('mobileAdmin.post.savedDraft')
        : id.value !== null && origStatus.value === 'published'
          ? t('mobileAdmin.post.updated')
          : t('mobileAdmin.post.published');
    leave();
    window.setTimeout(() => toast(msg, body.title), 380);
  } catch (err) {
    const code = err instanceof Error ? err.message : '';
    if (code === 'slug_exists') {
      toast(t('mobileAdmin.post.slugExists'), d.slug, 'error');
      metaOpen.value = true;
    } else {
      toast(t('mobileAdmin.common.saveFailed'), '', 'error');
    }
  } finally {
    busy.value = '';
  }
}

function leave(): void {
  leaving = true;
  if (window.history.state?.back) router.back();
  else void router.replace({ name: 'admin-posts' });
}

async function cancel(): Promise<void> {
  if (id.value !== null && dirty.value) {
    const ok = await dialog.confirm({
      title: t('mobileAdmin.post.discardTitle'),
      message: t('mobileAdmin.post.discardMsg'),
      confirmText: t('mobileAdmin.post.discard'),
      cancelText: t('mobileAdmin.note.keep'),
      danger: true,
    });
    if (!ok) return;
  } else if (id.value === null && dirty.value) {
    window.setTimeout(() => toast(t('mobileAdmin.post.keptLocal'), '', 'info'), 380);
  }
  leave();
}

onBeforeRouteLeave(async () => {
  if (leaving || id.value === null || !dirty.value) return true;
  return dialog.confirm({
    title: t('mobileAdmin.post.discardTitle'),
    message: t('mobileAdmin.post.discardMsg'),
    confirmText: t('mobileAdmin.post.discard'),
    cancelText: t('mobileAdmin.note.keep'),
    danger: true,
  });
});

/* ---------- Markdown 工具条 ---------- */
let sel = { s: 0, e: 0 };
function remember(): void {
  const el = bodyEl.value;
  if (el) sel = { s: el.selectionStart, e: el.selectionEnd };
}

function apply(next: string, s: number, e: number): void {
  draft.value.contentMd = next;
  void nextTick(() => {
    const el = bodyEl.value;
    if (!el) return;
    el.focus({ preventScroll: true });
    el.setSelectionRange(s, e);
    sel = { s, e };
    autosize(el, 260);
  });
}

function wrap(before: string, after: string, placeholder: string): void {
  const v = draft.value.contentMd;
  const { s, e } = sel;
  const text = v.slice(s, e) || placeholder;
  apply(v.slice(0, s) + before + text + after + v.slice(e), s + before.length, s + before.length + text.length);
}

function linePrefix(prefix: string | ((i: number) => string)): void {
  const v = draft.value.contentMd;
  const start = v.lastIndexOf('\n', Math.max(0, sel.s - 1)) + 1;
  const endIdx = v.indexOf('\n', sel.e);
  const end = endIdx < 0 ? v.length : endIdx;
  const lines = v.slice(start, end).split('\n');
  const pf = (i: number): string => (typeof prefix === 'string' ? prefix : prefix(i));
  const all = lines.every((l, i) => l.startsWith(pf(i)));
  const out = lines.map((l, i) => (all ? l.slice(pf(i).length) : pf(i) + l)).join('\n');
  apply(v.slice(0, start) + out + v.slice(end), start, start + out.length);
}

function insertBlock(text: string, selectFrom = 0, selectLen = 0): void {
  const v = draft.value.contentMd;
  const { s, e } = sel;
  const pre = s > 0 && v[s - 1] !== '\n' ? '\n\n' : '';
  const ins = pre + text;
  const pos = s + pre.length + selectFrom;
  apply(v.slice(0, s) + ins + v.slice(e), pos, pos + selectLen);
}

interface Tool {
  id: string;
  icon?: IconName;
  label?: string;
  run: () => void;
}
const tools: Tool[] = [
  { id: 'bold', icon: 'bold', run: () => wrap('**', '**', t('mobileAdmin.md.boldText')) },
  { id: 'h2', label: 'H2', run: () => linePrefix('## ') },
  { id: 'h3', label: 'H3', run: () => linePrefix('### ') },
  { id: 'italic', icon: 'italic', run: () => wrap('*', '*', t('mobileAdmin.md.italicText')) },
  { id: 'quote', icon: 'quote', run: () => linePrefix('> ') },
  {
    id: 'code',
    icon: 'code',
    run: () => {
      const text = draft.value.contentMd.slice(sel.s, sel.e);
      if (text.includes('\n')) wrap('```\n', '\n```', '');
      else if (text) wrap('`', '`', '');
      else insertBlock('```\n\n```\n', 4, 0);
    },
  },
  {
    id: 'link',
    icon: 'link',
    run: () => {
      const v = draft.value.contentMd;
      const text = v.slice(sel.s, sel.e) || t('mobileAdmin.md.linkText');
      const ins = `[${text}](https://)`;
      apply(v.slice(0, sel.s) + ins + v.slice(sel.e), sel.s + text.length + 3, sel.s + ins.length - 1);
    },
  },
  { id: 'sep1', run: () => undefined },
  { id: 'list', icon: 'list', run: () => linePrefix('- ') },
  { id: 'olist', icon: 'olist', run: () => linePrefix((i) => `${i + 1}. `) },
  { id: 'task', icon: 'task', run: () => linePrefix('- [ ] ') },
  { id: 'hr', icon: 'hr', run: () => insertBlock('---\n\n', 5, 0) },
];

/* 触屏：touchend 阻止默认，避免焦点离开 textarea 收起键盘；横滑工具条时不触发 */
let touchX = 0;
function onToolTouchStart(e: TouchEvent): void {
  touchX = e.touches[0].clientX;
}
function onToolTouchEnd(e: TouchEvent, tool: Tool): void {
  if (Math.abs(e.changedTouches[0].clientX - touchX) > 8) return;
  e.preventDefault();
  tool.run();
}

const imgUploading = ref(false);
async function onImage(e: Event): Promise<void> {
  const input = e.target as HTMLInputElement;
  const files = Array.from(input.files ?? []);
  input.value = '';
  if (!files.length) return;
  imgUploading.value = true;
  try {
    const items = await adminApi.uploadMedia(files);
    shell.bump.media += 1;
    insertBlock(`${items.map((i) => `![](${i.url})`).join('\n\n')}\n\n`, 0, 0);
    toast(t('mobileAdmin.post.imageInserted', { n: items.length }));
  } catch {
    toast(t('mobileAdmin.common.uploadFailed'), '', 'error');
  } finally {
    imgUploading.value = false;
  }
}

/* 工具条跟随软键盘：visualViewport 底边 */
const kbOffset = ref(0);
const kbOpen = ref(false);
function onViewport(): void {
  const vv = window.visualViewport;
  if (!vv) return;
  const gap = window.innerHeight - (vv.height + vv.offsetTop);
  kbOffset.value = Math.max(0, gap);
  kbOpen.value = gap > 80;
}
onMounted(() => {
  window.visualViewport?.addEventListener('resize', onViewport);
  window.visualViewport?.addEventListener('scroll', onViewport);
  onViewport();
});
onBeforeUnmount(() => {
  window.visualViewport?.removeEventListener('resize', onViewport);
  window.visualViewport?.removeEventListener('scroll', onViewport);
  window.clearTimeout(autosaveTimer);
});

function hideKeyboard(): void {
  (document.activeElement as HTMLElement | null)?.blur();
}

const scrollY = ref(0);
</script>

<template>
  <div class="ed" :class="{ scrolled: scrollY > 4 }">
    <header class="ed-top">
      <button class="txtbtn tap" @click="cancel">{{ t('mobileAdmin.common.cancel') }}</button>
      <span class="grow" />
      <button
        v-if="draft.status === 'published' && origStatus !== 'published'"
        class="txtbtn sub tap"
        :disabled="!!busy"
        @click="save('draft')"
      >
        {{ busy === 'draft' ? t('mobileAdmin.common.saving') : t('mobileAdmin.post.saveDraft') }}
      </button>
      <button class="icbtn tap" :aria-label="t('mobileAdmin.post.meta')" @click="metaOpen = true">
        <MaIcon name="sliders" :size="20" />
      </button>
      <button class="pill-btn tap" :disabled="!!busy || loading" @click="save('publish')">{{ primaryLabel }}</button>
    </header>

    <div class="ed-body" @scroll.passive="scrollY = ($event.target as HTMLElement).scrollTop">
      <MaSkeleton v-if="loading" variant="rows" :count="3" />
      <template v-else>
        <button v-if="draft.covers[0]" class="ed-cover tap" @click="metaOpen = true">
          <img :src="thumbOf(draft.covers[0])" alt="" draggable="false" />
          <span class="swap"><MaIcon name="image" :size="14" />{{ t('mobileAdmin.post.changeCover') }}</span>
          <span v-if="draft.covers.length > 1" class="more">+{{ draft.covers.length - 1 }}</span>
        </button>
        <button v-else class="ed-addcover tap" @click="metaOpen = true">
          <MaIcon name="image" :size="18" />{{ t('mobileAdmin.post.addCover') }}
        </button>

        <textarea
          ref="titleEl"
          v-model="draft.title"
          class="ed-title"
          rows="1"
          :placeholder="t('mobileAdmin.post.titlePh')"
          @input="autosize(titleEl, 40)"
          @keydown.enter.prevent="bodyEl?.focus()"
        />

        <div class="ed-tags">
          <span class="st" :class="draft.status === 'draft' ? 'dr' : 'pub'">{{ draft.status === 'draft' ? t('mobileAdmin.content.draft') : t('mobileAdmin.content.published') }}</span>
          <button v-for="tg in draft.tags" :key="tg" class="tagp tap" @click="metaOpen = true"><span class="hs">#</span>{{ tg }}</button>
          <button class="tagp add tap" @click="metaOpen = true">+ {{ t('mobileAdmin.post.tag') }}</button>
        </div>

        <textarea
          ref="bodyEl"
          v-model="draft.contentMd"
          class="ed-text"
          :placeholder="t('mobileAdmin.post.bodyPh')"
          @input="autosize(bodyEl, 260); remember()"
          @select="remember"
          @keyup="remember"
          @click="remember"
          @blur="remember"
        />
        <p class="ed-stat">
          {{ t('mobileAdmin.post.words', { n: words }) }}
          <template v-if="savedAt"> · {{ t('mobileAdmin.post.autosaved', { time: savedAt }) }}</template>
        </p>
      </template>
    </div>

    <div class="ed-kb glass" :class="{ lifted: kbOpen }" :style="{ transform: `translateY(${-kbOffset}px)` }">
      <div class="kb-scroll">
        <template v-for="tool in tools" :key="tool.id">
          <span v-if="tool.id.startsWith('sep')" class="sep" />
          <button
            v-else
            class="kb-btn"
            :aria-label="t(`mobileAdmin.md.${tool.id}`)"
            @touchstart.passive="onToolTouchStart"
            @touchend="onToolTouchEnd($event, tool)"
            @mousedown.prevent
            @click="tool.run()"
          >
            <MaIcon v-if="tool.icon" :name="tool.icon" :size="18" />
            <span v-else>{{ tool.label }}</span>
          </button>
        </template>
        <span class="sep" />
        <label class="kb-btn" :aria-label="t('mobileAdmin.md.image')" @mousedown.prevent>
          <MaRing v-if="imgUploading" indeterminate :size="18" :stroke="2" class="kb-ring" />
          <MaIcon v-else name="image" :size="18" />
          <input type="file" accept="image/*" multiple hidden @change="onImage" />
        </label>
      </div>
      <button v-if="kbOpen" class="kb-hide" :aria-label="t('mobileAdmin.md.hide')" @click="hideKeyboard">
        <MaIcon name="keyboard" :size="18" />
      </button>
    </div>

    <PostMetaSheet v-model:open="metaOpen" v-model:draft="draft" />
  </div>
</template>

<style scoped lang="scss">
.ed {
  position: absolute;
  inset: 0;
  background: var(--bg);
}

.ed-top {
  position: absolute;
  z-index: 10;
  top: 0;
  left: 0;
  right: 0;
  height: calc(var(--safe-t) + 52px);
  padding: var(--safe-t) 14px 0;
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--glass-2);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  box-shadow: 0 0.5px 0 transparent;
  transition: box-shadow var(--dur);

  .scrolled & { box-shadow: 0 0.5px 0 var(--line-2); }

  .grow { flex: 1; }

  .sub {
    font-size: 15px;
    color: var(--text-2);

    &:disabled { opacity: 0.5; }
  }
}

.ed-body {
  position: absolute;
  inset: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: calc(var(--safe-t) + 58px) 20px calc(var(--safe-b) + 110px);
  scrollbar-width: none;

  &::-webkit-scrollbar { display: none; }
}

.ed-cover {
  position: relative;
  display: block;
  width: 100%;
  height: 176px;
  border-radius: var(--r-xl);
  overflow: hidden;
  box-shadow: var(--shadow-card);
  animation: ma-ed-in 0.5s var(--ease-out);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .swap,
  .more {
    position: absolute;
    bottom: 12px;
    display: flex;
    gap: 6px;
    align-items: center;
    padding: 7px 12px;
    border-radius: 999px;
    font-size: 12.5px;
    color: #fff;
    background: rgba(0, 0, 0, 0.38);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }

  .swap { right: 12px; }
  .more { left: 12px; font-family: var(--font-mono); }
}

.ed-addcover {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 44px;
  padding: 0 14px;
  border-radius: var(--r-md);
  font-size: 14px;
  color: var(--text-3);
  box-shadow: inset 0 0 0 1px var(--line-2);
  background: var(--fill);
}

@keyframes ma-ed-in {
  from { opacity: 0; transform: translateY(10px) scale(0.98); }
}

.ed-title {
  display: block;
  width: 100%;
  margin-top: 18px;
  resize: none;
  overflow: hidden;
  font-family: var(--font-serif);
  font-size: 27px;
  font-weight: 700;
  line-height: 1.35;
  color: var(--text);
  caret-color: var(--ink);

  &::placeholder { color: var(--text-3); }
}

.ed-tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-top: 12px;

  .st {
    padding: 4px 9px;
    border-radius: 999px;
  }
}

/* 标签：极轻描边胶囊，「# 名称」# 用三级灰 */
.tagp {
  padding: 4px 10px;
  border-radius: 999px;
  box-shadow: inset 0 0 0 1px var(--line-2);
  color: var(--text-2);
  font-weight: 500;
  font-size: 12px;

  .hs { color: var(--text-3); margin-right: 2px; }

  &.add {
    background: var(--fill);
    box-shadow: none;
    color: var(--text-3);
  }
}

.ed-text {
  display: block;
  width: 100%;
  min-height: 260px;
  margin-top: 16px;
  resize: none;
  overflow: hidden;
  font-family: var(--font-serif);
  font-size: 17px;
  line-height: 1.9;
  color: var(--text);
  caret-color: var(--ink);

  &::placeholder { color: var(--text-3); }
}

.ed-stat {
  margin-top: 14px;
  font-size: 12px;
  color: var(--text-3);
  font-family: var(--font-mono);
}

/* 键盘上方工具条 */
.ed-kb {
  position: absolute;
  z-index: 12;
  left: 10px;
  right: 10px;
  bottom: calc(var(--safe-b) + 8px);
  height: 50px;
  border-radius: var(--r-lg);
  display: flex;
  align-items: center;
  padding: 0 6px;
  box-shadow: inset 0 0 0 0.5px var(--glass-line), inset 0 1px 0 var(--glass-hi), var(--shadow-pop);
  transition: transform 0.25s var(--ease-out), bottom 0.25s var(--ease-out), border-radius 0.25s;

  &.lifted {
    bottom: 6px;
  }
}

.kb-scroll {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 2px;
  overflow-x: auto;
  scrollbar-width: none;

  &::-webkit-scrollbar { display: none; }
}

.kb-btn {
  flex: none;
  min-width: 40px;
  height: 38px;
  padding: 0 8px;
  border-radius: var(--r-sm);
  display: grid;
  place-items: center;
  font-size: 14px;
  font-weight: 700;
  color: var(--text-2);
  font-family: var(--font-mono);
  cursor: pointer;
  transition: background var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  &:active {
    background: var(--fill-2);
    transform: scale(0.92);
  }

  .kb-ring {
    --ring-bg: var(--fill-2);
    --ring-fg: var(--ink);
  }
}

.sep {
  flex: none;
  width: 0.5px;
  height: 22px;
  background: var(--line-2);
  margin: 0 4px;
}

.kb-hide {
  flex: none;
  width: 40px;
  height: 38px;
  margin-left: 4px;
  border-radius: var(--r-sm);
  display: grid;
  place-items: center;
  color: var(--ink);
  box-shadow: -8px 0 12px -8px rgba(0, 0, 0, 0.3);
}
</style>
