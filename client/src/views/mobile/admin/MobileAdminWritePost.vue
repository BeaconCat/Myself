<script setup lang="ts">
/**
 * 写文章 · 全屏编辑器：封面、标题、标签、Markdown 正文富文本 / Markdown 源码；
 * 与桌面共用完整 Markdown 编辑器及正文样式；
 * 元信息 sheet（slug / 标签 / 摘要 / 封面 / 状态 / 置顶）；?id= 编辑，新文章本机自动保存草稿。
 */
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, thumbOf, type PostDraft, type PublicationStatus, type PublicationResult } from '../../../api';
import { useDialogStore } from '../../../stores/dialog';
import { useAuthStore } from '../../../stores/auth';
import { useConfigStore } from '../../../stores/config';
import { scheduleValid, scheduleLabel } from '../../../utils/publication';
import MarkdownEditor from '../../../components/admin/MarkdownEditor.vue';
import MaIcon from '../../../components/mobile-admin/MaIcon.vue';
import MaSkeleton from '../../../components/mobile-admin/MaSkeleton.vue';
import PostMetaSheet from '../../../components/mobile-admin/PostMetaSheet.vue';
import { shell, toast } from '../../../components/mobile-admin/state';
import { SLUG_RE, toSlug } from '../../../components/mobile-admin/format';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const dialog = useDialogStore();
const config = useConfigStore();
const auth = useAuthStore();
const review = computed(() => auth.role === 'author' && !config.cfg.users?.authors.directPublish);

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
  publishAt: null,
  pinned: false,
});
const draft = ref<PostDraft>(empty());
const loading = ref(id.value !== null);
const busy = ref<'' | 'publish' | 'draft'>('');
const metaOpen = ref(false);
const savedAt = ref('');
const scrollY = ref(0);
let snapshot = JSON.stringify(draft.value);
let leaving = false;
const origStatus = ref<PublicationStatus | null>(null);
const origTime = ref<string | null>(null);
const unchangedSchedule = computed(() => origStatus.value === 'scheduled' && draft.value.status === 'scheduled' && draft.value.publishAt === origTime.value);

const dirty = computed(() => JSON.stringify(draft.value) !== snapshot);
const words = computed(() => draft.value.contentMd.replace(/\s/g, '').length);
const primaryLabel = computed(() => {
  if (busy.value === 'publish') return t('mobileAdmin.common.saving');
  if (draft.value.status === 'scheduled') return t(origStatus.value === 'scheduled' ? 'schedule.save' : 'schedule.later');
  if (review.value && draft.value.status !== 'draft') return t('studio.write.submitReview');
  if (draft.value.status === 'draft') return t('mobileAdmin.common.save');
  if (id.value !== null && origStatus.value === 'published') return t('mobileAdmin.post.update');
  return t('mobileAdmin.post.publish');
});

const titleEl = ref<HTMLTextAreaElement | null>(null);
const bodyEl = ref<InstanceType<typeof MarkdownEditor> | null>(null);

function autosize(el: HTMLTextAreaElement | null, min: number): void {
  if (!el) return;
  el.style.height = 'auto';
  el.style.height = `${Math.max(min, el.scrollHeight)}px`;
}
const fit = (): void => {
  autosize(titleEl.value, 40);
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
        publishAt: p.status === 'scheduled' ? p.publishAt ?? '' : null,
        pinned: p.pinned,
      };
      origStatus.value = p.status;
      origTime.value = draft.value.publishAt ?? null;
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
  if (mode !== 'draft' && d.status === 'scheduled' && !unchangedSchedule.value && !scheduleValid(d.publishAt, config.cfg.timezone)) { toast(t('schedule.invalid'), '', 'error'); metaOpen.value = true; return; }
  busy.value = mode;
  /* 主按钮按元信息里的状态保存；「存草稿」强制草稿 */
  const body: PostDraft = { ...d, status: mode === 'draft' ? 'draft' : d.status, publishAt: mode !== 'draft' && d.status === 'scheduled' ? d.publishAt : null };
  const requestedStatus = body.status;
  try {
    let result: PublicationResult;
    if (id.value === null) {
      result = await adminApi.createPost(body);
      localStorage.removeItem(LOCAL_KEY);
    } else {
      if (unchangedSchedule.value && mode !== 'draft') {
        const { status: _status, publishAt: _at, ...content } = body;
        result = await adminApi.updatePost(id.value, content);
      } else result = await adminApi.updatePost(id.value, body);
    }
    body.status = result.status ?? body.status;
    draft.value.status = body.status;
    draft.value.publishAt = result.publishAt || null;
    snapshot = JSON.stringify(draft.value);
    shell.bump.posts += 1;
    const msg =
      review.value && requestedStatus !== 'draft' ? t('studio.write.submitted') : body.status === 'scheduled' ? t('schedule.arranged') : body.status === 'draft'
        ? t(origStatus.value === 'scheduled' ? 'schedule.cancelled' : 'mobileAdmin.post.savedDraft')
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
      toast(t(code.includes('publish_time') ? 'schedule.invalid' : 'mobileAdmin.common.saveFailed'), '', 'error');
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

onBeforeUnmount(() => window.clearTimeout(autosaveTimer));
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
      <button v-if="draft.status === 'scheduled'" type="button" class="txtbtn sub tap" :disabled="!!busy" :title="t('schedule.cancel')" @click="save('draft')">{{ t('schedule.cancelShort') }}</button>
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
          <span class="st" :class="draft.status === 'published' ? 'pub' : 'dr'">{{ t(`schedule.${draft.status}`) }}</span>
          <span v-if="draft.status === 'scheduled'" class="schedule-label">{{ scheduleLabel(draft.publishAt || '', config.cfg.timezone) }}</span>
          <button v-for="tg in draft.tags" :key="tg" class="tagp tap" @click="metaOpen = true"><span class="hs">#</span>{{ tg }}</button>
          <button class="tagp add tap" @click="metaOpen = true">+ {{ t('mobileAdmin.post.tag') }}</button>
        </div>

        <MarkdownEditor ref="bodyEl" v-model="draft.contentMd" :placeholder="t('mobileAdmin.post.bodyPh')" />
        <p class="ed-stat">
          {{ t('mobileAdmin.post.words', { n: words }) }}
          <template v-if="savedAt"> · {{ t('mobileAdmin.post.autosaved', { time: savedAt }) }}</template>
        </p>
      </template>
    </div>

    <PostMetaSheet v-model:open="metaOpen" v-model:draft="draft" :allow-schedule="!review && origStatus !== 'published'" />
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

.schedule-label { font: 12px var(--font-mono); color: var(--text-2); }

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

.ed-body :deep(.markdown-editor) { margin-top: 16px; }

.ed-stat {
  margin-top: 14px;
  font-size: 12px;
  color: var(--text-3);
  font-family: var(--font-mono);
}

</style>
