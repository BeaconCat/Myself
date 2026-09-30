<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, thumbOf, type MediaItem, type PostDraft } from '../../api';
import { useDialogStore } from '../../stores/dialog';
import { useAuthStore } from '../../stores/auth';
import { useConfigStore } from '../../stores/config';
import RichEditor from '../../components/admin/RichEditor.vue';
import CoverUploader from '../../components/admin/CoverUploader.vue';
import '../admin/studio/i18n';
import SIcon from '../admin/studio/SIcon.vue';
import StSwitch from '../admin/studio/StSwitch.vue';
import StModal from '../admin/studio/StModal.vue';
import { refreshCounts } from '../admin/studio/state';
import { toast } from '../admin/studio/toast';
import { dateText, wordCount } from '../admin/studio/format';
import PublishPanel from './PublishPanel.vue';
import DoorStage from './DoorStage.vue';

/**
 * 沉浸式写作：正文居中 680px 宋体，浮动极简工具条，右侧元信息抽屉。
 * 内容以 Markdown 存储（RichEditor 负责 Markdown ⇄ 富文本）。Ctrl+S 保存，Ctrl+Enter 发布。
 */
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const dialog = useDialogStore();
const auth = useAuthStore();
const config = useConfigStore();
/** 协作作者；未开「直接发布」时发布即提交审阅（服务端存为草稿） */
const isAuthor = computed(() => auth.role === 'author');
const review = computed(() => isAuthor.value && !config.cfg.users?.authors.directPublish);

const id = ref<number | null>(null);
const draft = reactive<PostDraft>({
  slug: '',
  title: '',
  excerpt: '',
  contentMd: '',
  covers: [],
  tags: [],
  status: 'draft',
  pinned: false,
});
const createdAt = ref('');

type SaveState = 'new' | 'clean' | 'dirty' | 'saving' | 'error';
const saveState = ref<SaveState>('new');
const busy = ref(false);
const loading = ref(false);
const drawer = ref(false);
const typing = ref(false);
const source = ref(false);
const slugTouched = ref(false);
/** 定时发布：预留界面，保存时忽略 */
const schedule = reactive({ on: false, date: '', time: '09:00' });

const rich = ref<InstanceType<typeof RichEditor> | null>(null);
const titleEl = ref<HTMLTextAreaElement | null>(null);

/* ===== 载入 ===== */
function parseId(): number | null {
  const raw = route.query.id;
  const n = typeof raw === 'string' && raw ? Number(raw) : NaN;
  return Number.isFinite(n) ? n : null;
}

let hydrating = false;
async function load(): Promise<void> {
  id.value = parseId();
  if (id.value === null) {
    saveState.value = 'new';
    void nextTick(() => titleEl.value?.focus());
    return;
  }
  loading.value = true;
  try {
    const post = await adminApi.post(id.value);
    hydrating = true;
    Object.assign(draft, {
      slug: post.slug,
      title: post.title,
      excerpt: post.excerpt,
      contentMd: post.contentMd ?? '',
      covers: [...post.covers],
      tags: [...post.tags],
      status: post.status,
      pinned: post.pinned,
    });
    createdAt.value = post.createdAt;
    slugTouched.value = true;
    saveState.value = 'clean';
    await nextTick();
    hydrating = false;
    fitTitle();
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  } finally {
    loading.value = false;
  }
}

watch(
  () => route.query.id,
  () => {
    if (String(parseId()) !== String(id.value)) void load();
  },
);

/* ===== 标题 / slug ===== */
function fitTitle(): void {
  const el = titleEl.value;
  if (!el) return;
  el.style.height = 'auto';
  el.style.height = `${el.scrollHeight}px`;
}

function makeSlug(title: string): string {
  return title
    .toLowerCase()
    .replace(/[^\w一-龥]+/g, '-')
    .replace(/[一-龥]/g, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 60);
}

function onTitle(): void {
  fitTitle();
  if (!slugTouched.value) draft.slug = makeSlug(draft.title);
}

function ensureSlug(): void {
  if (!draft.slug) draft.slug = makeSlug(draft.title) || `post-${Date.now().toString(36)}`;
}

/* ===== 脏标记 + 自动保存（仅已存在的草稿） ===== */
let autoTimer = 0;
watch(
  () => [draft.title, draft.contentMd, draft.excerpt, draft.slug, draft.covers.join(), draft.tags.join(), draft.pinned],
  () => {
    if (hydrating || loading.value) return;
    if (saveState.value !== 'saving') saveState.value = 'dirty';
    window.clearTimeout(autoTimer);
    if (id.value !== null && draft.status === 'draft') {
      autoTimer = window.setTimeout(() => void save(), 2400);
    }
  },
);

const stateText = computed(() => t(`studio.write.state.${saveState.value}`));

/* ===== 保存 / 发布 ===== */
async function save(status: 'draft' | 'published' = draft.status): Promise<boolean> {
  if (busy.value) return false;
  if (!draft.title.trim()) {
    toast(t('studio.write.needTitle'), { icon: 'pen' });
    titleEl.value?.focus();
    return false;
  }
  window.clearTimeout(autoTimer);
  ensureSlug();
  busy.value = true;
  saveState.value = 'saving';
  const body: PostDraft = { ...draft, covers: [...draft.covers], tags: [...draft.tags], status };
  try {
    if (id.value === null) {
      const res = await adminApi.createPost(body);
      id.value = res.id;
      createdAt.value = new Date().toISOString();
      void router.replace({ name: 'admin-write-post', query: { id: String(res.id) } });
    } else {
      await adminApi.updatePost(id.value, body);
    }
    draft.status = status;
    saveState.value = 'clean';
    void refreshCounts();
    return true;
  } catch (err) {
    saveState.value = 'error';
    toast((err as Error).message === 'slug_exists' ? t('studio.write.slugExists') : t('studio.saveFailed'), { icon: 'x' });
    if ((err as Error).message === 'slug_exists') drawer.value = true;
    return false;
  } finally {
    busy.value = false;
  }
}

async function saveNow(): Promise<void> {
  if (await save()) toast(t('studio.saved'));
}

async function unpublish(): Promise<void> {
  const ok = await dialog.confirm({
    title: t('studio.write.unpublishTitle'),
    message: t('studio.write.unpublishBody'),
    confirmText: t('studio.write.unpublish'),
  });
  if (ok && (await save('draft'))) toast(t('studio.write.unpublished'), { icon: 'lock' });
}

/* 发布面板 */
const pubOpen = ref(false);
const announce = ref(false);
const republish = ref(false);

function openPublish(): void {
  if (!draft.title.trim()) {
    toast(t('studio.write.needTitle'), { icon: 'pen' });
    titleEl.value?.focus();
    return;
  }
  ensureSlug();
  republish.value = draft.status === 'published';
  pubOpen.value = true;
}

/* 推门发布 */
const stage = reactive({ open: false, title: '', cover: '', url: '', meta: '' });

async function confirmPublish(): Promise<void> {
  if (review.value) {
    if (!(await save('published'))) return;
    draft.status = 'draft';
    pubOpen.value = false;
    toast(t('studio.write.submitted'), { icon: 'send' });
    return;
  }
  const first = draft.status !== 'published';
  if (!(await save('published'))) return;
  pubOpen.value = false;
  if (first && announce.value && !isAuthor.value) {
    const lines = [t('studio.write.announceText', { title: draft.title }), draft.excerpt, `[${t('studio.write.readMore')}](/articles/${draft.slug})`];
    adminApi.createNote({ contentMd: lines.filter(Boolean).join('\n\n'), mood: '', images: [], pinned: false })
      .then(() => refreshCounts())
      .catch(() => toast(t('studio.write.announceFailed'), { icon: 'x' }));
  }
  if (!first) {
    toast(t('studio.write.updated'), { icon: 'check', action: t('studio.view'), fn: viewPost });
    return;
  }
  Object.assign(stage, {
    title: draft.title,
    cover: draft.covers[0] ? thumbOf(draft.covers[0]) : '',
    url: `${window.location.host}/articles/${draft.slug}`,
    meta: `${dateText(createdAt.value) || new Date().toISOString().slice(0, 10)} · ${t('studio.write.minutes', { n: minutes.value })}`,
    open: true,
  });
}

function viewPost(): void {
  window.open(`/articles/${encodeURIComponent(draft.slug)}`, '_blank');
}

function stageView(): void {
  viewPost();
  stage.open = false;
}

function stageClose(): void {
  stage.open = false;
  void router.push({ name: 'admin-today' });
}

/* ===== 从素材库选封面 ===== */
const picker = ref(false);
const library = ref<MediaItem[]>([]);
async function openPicker(): Promise<void> {
  picker.value = true;
  try {
    library.value = await adminApi.media();
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
}
function toggleCover(url: string): void {
  const i = draft.covers.indexOf(url);
  if (i >= 0) draft.covers.splice(i, 1);
  else if (draft.covers.length < 3) draft.covers.push(url);
}

/* ===== 字数 ===== */
const words = computed(() => wordCount(draft.contentMd) + draft.title.trim().length);
const minutes = computed(() => Math.max(1, Math.round(words.value / 400)));

/* ===== 浮动工具条 ===== */
interface Tool { icon?: string; text?: string; title: string; run: () => unknown; active?: () => boolean }
const ed = () => rich.value?.editor;
const chain = () => ed()!.chain().focus();
const TOOLS: (Tool | 'sep')[] = [
  { text: 'H2', title: t('studio.write.tool.h2'), run: () => chain().toggleHeading({ level: 2 }).run(), active: () => !!ed()?.isActive('heading', { level: 2 }) },
  { text: 'H3', title: t('studio.write.tool.h3'), run: () => chain().toggleHeading({ level: 3 }).run(), active: () => !!ed()?.isActive('heading', { level: 3 }) },
  'sep',
  { icon: 'bold', title: t('studio.write.tool.bold'), run: () => chain().toggleBold().run(), active: () => !!ed()?.isActive('bold') },
  { icon: 'italic', title: t('studio.write.tool.italic'), run: () => chain().toggleItalic().run(), active: () => !!ed()?.isActive('italic') },
  { icon: 'strike', title: t('studio.write.tool.strike'), run: () => chain().toggleStrike().run(), active: () => !!ed()?.isActive('strike') },
  'sep',
  { icon: 'quote', title: t('studio.write.tool.quote'), run: () => chain().toggleBlockquote().run(), active: () => !!ed()?.isActive('blockquote') },
  { icon: 'code', title: t('studio.write.tool.code'), run: () => chain().toggleCode().run(), active: () => !!ed()?.isActive('code') },
  { icon: 'codeBlock', title: t('studio.write.tool.codeBlock'), run: () => chain().toggleCodeBlock().run(), active: () => !!ed()?.isActive('codeBlock') },
  { icon: 'ul', title: t('studio.write.tool.ul'), run: () => chain().toggleBulletList().run(), active: () => !!ed()?.isActive('bulletList') },
  { icon: 'task', title: t('studio.write.tool.task'), run: () => chain().toggleTaskList().run(), active: () => !!ed()?.isActive('taskList') },
  'sep',
  { icon: 'link', title: t('studio.write.tool.link'), run: () => rich.value?.setLink(), active: () => !!ed()?.isActive('link') },
  { icon: 'image', title: t('studio.write.tool.image'), run: () => rich.value?.pickImage() },
  { icon: 'collage', title: t('studio.write.tool.collage'), run: () => rich.value?.pickCollage() },
  { icon: 'media', title: t('studio.write.tool.media'), run: () => rich.value?.pickMedia() },
  { icon: 'table', title: t('studio.write.tool.table'), run: () => rich.value?.insertTable(), active: () => !!ed()?.isActive('table') },
];
/** 依赖编辑器事务版本号，保证激活态实时刷新 */
const activeMap = computed(() => {
  void rich.value?.version;
  return TOOLS.map((tool) => (tool !== 'sep' && tool.active ? tool.active() : false));
});

/* 打字时收起顶栏，鼠标移动再浮现 */
function onTyping(): void {
  typing.value = true;
}
function onMove(e: MouseEvent): void {
  if (Math.abs(e.movementX) + Math.abs(e.movementY) > 4) typing.value = false;
}

/* ===== 标签 ===== */
const tagInput = ref('');
function addTag(): void {
  const parts = tagInput.value.split(/[,，\s]+/).map((s) => s.trim()).filter(Boolean);
  for (const p of parts) if (!draft.tags.includes(p)) draft.tags.push(p);
  tagInput.value = '';
}
function onTagKey(e: KeyboardEvent): void {
  if (e.key === 'Enter' || e.key === ',' || e.key === '，') {
    e.preventDefault();
    addTag();
  } else if (e.key === 'Backspace' && !tagInput.value && draft.tags.length) {
    draft.tags.pop();
  }
}

/* ===== 快捷键 ===== */
function onKey(e: KeyboardEvent): void {
  const mod = e.ctrlKey || e.metaKey;
  if (mod && e.key.toLowerCase() === 's') {
    e.preventDefault();
    void saveNow();
  } else if (mod && e.key === 'Enter' && !pubOpen.value && !stage.open) {
    e.preventDefault();
    openPublish();
  } else if (e.key === 'Escape' && drawer.value && !pubOpen.value) {
    drawer.value = false;
  }
}

function back(): void {
  void router.push({ name: 'admin-posts' });
}

onBeforeRouteLeave(async () => {
  if (saveState.value !== 'dirty' && saveState.value !== 'error') return true;
  if (!draft.title.trim() && !draft.contentMd.trim()) return true;
  return dialog.confirm({
    title: t('studio.write.leaveTitle'),
    message: t('studio.write.leaveBody'),
    confirmText: t('studio.write.leave'),
    danger: true,
  });
});

onMounted(() => {
  void load();
  window.addEventListener('keydown', onKey);
});
onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey);
  window.clearTimeout(autoTimer);
});
</script>

<template>
  <div class="studio editor" :class="{ drawer, typing }" @mousemove="onMove">
    <div class="ed-top">
      <div class="capsule">
        <button type="button" class="st-ibtn" :title="t('studio.write.back')" @click="back"><SIcon name="arrowL" /></button>
        <div class="ed-status" :class="saveState"><span class="st-dot" />{{ stateText }}</div>
      </div>

      <div class="capsule fmt" :class="{ off: source }">
        <template v-for="(tool, i) in TOOLS" :key="i">
          <span v-if="tool === 'sep'" class="sep" />
          <button
            v-else
            type="button"
            class="st-ibtn"
            :class="{ on: activeMap[i], tx: !!tool.text }"
            :title="tool.title"
            :disabled="source"
            @mousedown.prevent
            @click="tool.run()"
          >
            <SIcon v-if="tool.icon" :name="tool.icon" />
            <template v-else>{{ tool.text }}</template>
          </button>
        </template>
        <span class="sep" />
        <button type="button" class="st-ibtn" :class="{ on: source }" :title="t('studio.write.tool.source')" @click="source = !source">
          <SIcon name="markdown" />
        </button>
      </div>

      <div class="capsule">
        <button type="button" class="st-ibtn" :class="{ on: drawer }" :title="t('studio.write.settings')" @click="drawer = !drawer">
          <SIcon name="panel" />
        </button>
        <button type="button" class="st-btn q sm" :disabled="busy" @click="saveNow">{{ t('studio.save') }}</button>
        <button type="button" class="st-btn p sm" :disabled="busy" @click="openPublish">
          {{ draft.status === 'published' ? t('studio.write.update') : t('studio.write.publish') }}
        </button>
      </div>
    </div>

    <div class="ed-scroll">
      <div class="ed-page">
        <div class="ed-cover" :class="{ on: draft.covers.length > 0 }">
          <img v-if="draft.covers[0]" :src="draft.covers[0]" alt="" />
        </div>
        <div class="ed-kicker">
          <span class="st-badge" :class="`st-${draft.status}`"><i class="st-dot" />{{ id === null ? t('studio.write.newDraft') : t(`studio.status.${draft.status}`) }}</span>
          <span v-for="tag in draft.tags.slice(0, 3)" :key="tag" class="tag">{{ tag }}</span>
          <button v-if="!draft.tags.length" type="button" class="add-tag" @click="drawer = true">{{ t('studio.write.addTag') }}</button>
        </div>
        <textarea
          ref="titleEl"
          v-model="draft.title"
          class="ed-title"
          rows="1"
          :placeholder="t('studio.write.titlePh')"
          @input="onTitle"
          @keydown.enter.prevent="rich?.focus()"
        />
        <RichEditor
          v-if="!source"
          ref="rich"
          v-model="draft.contentMd"
          bare
          class="prose"
          :placeholder="t('studio.write.bodyPh')"
          @typing="onTyping"
        />
        <textarea
          v-else
          v-model="draft.contentMd"
          class="md-src"
          spellcheck="false"
          :placeholder="t('studio.write.bodyPh')"
          @keydown="onTyping"
        />
      </div>
    </div>

    <div class="ed-foot">
      <span class="mono">{{ words.toLocaleString('en-US') }}</span>{{ t('studio.write.words') }}
      <span>·</span>{{ t('studio.write.readTime', { n: minutes }) }}
    </div>

    <aside class="drawer-p" :aria-hidden="!drawer">
      <div class="dr-h">
        <h3>{{ t('studio.write.settings') }}</h3>
        <button type="button" class="st-ibtn" @click="drawer = false"><SIcon name="x" /></button>
      </div>
      <div class="dr-b">
        <div>
          <div class="st-flabel"><span>{{ t('studio.write.covers') }}</span><span>{{ t('studio.write.coversHint') }}</span></div>
          <CoverUploader v-model="draft.covers" :max="3" />
          <button v-if="!isAuthor" type="button" class="st-link pick-btn" @click="openPicker"><SIcon name="image" :size="14" />{{ t('studio.write.pickFromLib') }}</button>
        </div>
        <div>
          <div class="st-flabel">{{ t('studio.write.tags') }}</div>
          <div class="tags" @click="($event.currentTarget as HTMLElement).querySelector('input')?.focus()">
            <span v-for="(tag, i) in draft.tags" :key="tag">{{ tag }}<button type="button" @click.stop="draft.tags.splice(i, 1)"><SIcon name="x" :size="14" /></button></span>
            <input v-model="tagInput" :placeholder="t('studio.write.tagPh')" @keydown="onTagKey" @blur="addTag" />
          </div>
        </div>
        <div>
          <div class="st-flabel"><span>{{ t('studio.write.excerpt') }}</span><span class="mono" :class="{ warn: draft.excerpt.length > 120 }">{{ draft.excerpt.length }} / 120</span></div>
          <label class="st-field ta"><textarea v-model="draft.excerpt" rows="3" :placeholder="t('studio.write.excerptPh')" /></label>
        </div>
        <div>
          <div class="st-flabel">{{ t('studio.write.slug') }}</div>
          <label class="slug">/articles/<input v-model="draft.slug" spellcheck="false" @input="slugTouched = true" /></label>
        </div>
        <div class="opts">
          <div class="row-opt">
            <div>{{ t('studio.write.pinned') }}<small>{{ t('studio.write.pinnedSub') }}</small></div>
            <StSwitch v-model="draft.pinned" :label="t('studio.write.pinned')" />
          </div>
          <div class="row-opt">
            <div>{{ t('studio.write.status') }}<small>{{ draft.status === 'published' ? t('studio.write.statusPub') : t('studio.write.statusDraft') }}</small></div>
            <button v-if="draft.status === 'published'" type="button" class="st-btn g sm" :disabled="busy" @click="unpublish">{{ t('studio.write.unpublish') }}</button>
            <span v-else class="st-badge st-draft"><i class="st-dot" />{{ t('studio.status.draft') }}</span>
          </div>
          <div>
            <div class="row-opt">
              <div>{{ t('studio.write.schedule') }}<small>{{ t('studio.write.scheduleSub') }}</small></div>
              <StSwitch v-model="schedule.on" :label="t('studio.write.schedule')" />
            </div>
            <div class="reveal" :class="{ on: schedule.on }">
              <div>
                <div class="dt">
                  <label class="st-field"><SIcon name="calendar" :size="16" /><input v-model="schedule.date" type="date" /></label>
                  <label class="st-field"><SIcon name="clock" :size="16" /><input v-model="schedule.time" type="time" /></label>
                </div>
                <p class="sched-note"><SIcon name="info" :size="14" />{{ t('studio.write.scheduleNote') }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="dr-f">
        <button type="button" class="st-btn g" :disabled="busy" @click="saveNow">{{ t('studio.save') }}</button>
        <button type="button" class="st-btn p" :disabled="busy" @click="openPublish">
          <SIcon name="send" :size="16" />{{ draft.status === 'published' ? t('studio.write.update') : t('studio.write.publish') }}
        </button>
      </div>
    </aside>

    <StModal :open="picker" wide panel-class="pick" @close="picker = false">
      <h3>{{ t('studio.write.pickTitle') }}</h3>
      <p>{{ t('studio.write.pickDesc') }}</p>
      <div class="pick-grid">
        <button
          v-for="m in library"
          :key="m.name"
          type="button"
          :class="{ on: draft.covers.includes(m.url), full: draft.covers.length >= 3 && !draft.covers.includes(m.url) }"
          @click="toggleCover(m.url)"
        >
          <img :src="m.thumb" alt="" loading="lazy" />
          <span class="n">{{ draft.covers.indexOf(m.url) + 1 || '' }}</span>
        </button>
      </div>
      <div class="ft">
        <span class="cnt mono">{{ draft.covers.length }} / 3</span>
        <button type="button" class="st-btn p" @click="picker = false">{{ t('studio.write.pickDone') }}</button>
      </div>
    </StModal>

    <PublishPanel
      v-model:pinned="draft.pinned"
      v-model:announce="announce"
      :open="pubOpen"
      :title="draft.title"
      :excerpt="draft.excerpt"
      :cover="draft.covers[0] ? thumbOf(draft.covers[0]) : ''"
      :seed="draft.slug"
      :tags="draft.tags"
      :slug="draft.slug"
      :minutes="minutes"
      :republish="republish"
      :busy="busy"
      :author="isAuthor"
      :review="review"
      @close="pubOpen = false"
      @confirm="confirmPublish"
    />
    <DoorStage
      :open="stage.open"
      :title="stage.title"
      :cover="stage.cover"
      :seed="draft.slug"
      :url="stage.url"
      :meta="stage.meta"
      @view="stageView"
      @close="stageClose"
    />
  </div>
</template>

<style scoped lang="scss">
.editor {
  position: relative;
  height: 100%;
  /* 收起的设置抽屉平移在屏幕外：裁掉，不能被横向滚动露出来（clip 不建滚动容器） */
  overflow: clip;
  display: flex;
  flex-direction: column;
  background: var(--paper);
  animation: ed-in var(--dur-slow) var(--ease-spring) both;
}

@keyframes ed-in { from { opacity: 0; transform: scale(0.985); } }

.ed-top {
  position: absolute;
  left: 0;
  right: 0;
  top: 16px;
  z-index: 3;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 0 20px;
  pointer-events: none;
  transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-out);

  > * { pointer-events: auto; }

  .typing & { opacity: 0; transform: translateY(-8px); }
  .typing &:hover { opacity: 1; transform: none; }
}

.capsule {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 4px;
  border-radius: var(--r-md);
  background: color-mix(in oklab, var(--paper) 82%, transparent);
  backdrop-filter: blur(16px) saturate(1.3);
  box-shadow: var(--sh-pop);

  .st-ibtn { width: 32px; height: 32px; border-radius: var(--r-sm); }
  .st-ibtn.tx { font: 600 13px var(--font-serif); width: auto; padding: 0 8px; }
  .sep { width: 1px; height: 18px; background: var(--line-2); margin: 0 4px; }
  .st-btn.sm { margin-left: 2px; }

  &.fmt.off .st-ibtn:disabled { opacity: 0.3; }

  &.fmt { transition: transform var(--dur-slow) var(--ease-spring); }
  .drawer &.fmt { transform: translateX(-180px); }
}

.ed-status {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12.5px;
  color: var(--st-ink-3);
  padding: 0 10px 0 4px;
  white-space: nowrap;

  .st-dot { --c: var(--green); }

  &.dirty .st-dot, &.new .st-dot { --c: var(--yellow); }
  &.saving .st-dot { --c: var(--blue); animation: breath 1s ease-in-out infinite; }
  &.error .st-dot { --c: var(--red); }
  &.clean .st-dot { animation: breath 2.4s ease-in-out infinite; }
}

@keyframes breath { 50% { opacity: 0.35; } }

.ed-scroll { flex: 1; overflow: auto; }

.ed-page {
  max-width: 680px;
  margin: 0 auto;
  padding: 120px 0 200px;
  transition: transform var(--dur-slow) var(--ease-spring);

  .drawer & { transform: translateX(-180px); }
}

.ed-cover {
  height: 0;
  margin-bottom: 0;
  border-radius: var(--r-md);
  overflow: hidden;
  background: var(--well);
  transition: all var(--dur-slow) var(--ease-spring);

  &.on { height: 220px; margin-bottom: 40px; }

  img { width: 100%; height: 100%; object-fit: cover; display: block; }
}

.ed-kicker {
  display: flex;
  gap: 10px;
  align-items: center;
  font-size: 13px;
  color: var(--st-ink-3);
  margin-bottom: 18px;

  .tag::before { content: '#'; color: var(--st-ink-4); margin-right: 2px; }

  .add-tag {
    color: var(--st-ink-4);
    transition: color var(--dur-fast);

    &:hover { color: var(--ink); }
  }
}

.ed-title {
  font: 700 40px/1.35 var(--font-serif);
  border: 0;
  outline: 0;
  width: 100%;
  background: none;
  resize: none;
  padding: 0;
  margin: 0 0 28px;
  color: var(--st-ink);
  letter-spacing: 0.01em;
  overflow: hidden;

  &::placeholder { color: var(--st-ink-4); }
}

.prose {
  caret-color: var(--ink);

  :deep(.ProseMirror) {
    min-height: 40vh;
    font: 400 18px/2 var(--font-serif);
    color: var(--st-ink);

    > * + * { margin-top: 0; }
    p { margin: 0 0 1.1em; }
    h2 { font: 700 24px/1.5 var(--font-serif); margin: 1.8em 0 0.6em; }
    h3 { font: 700 20px/1.5 var(--font-serif); margin: 1.6em 0 0.5em; }

    blockquote {
      margin: 1.8em 0;
      padding: 0.2em 0 0.2em 1.5em;
      position: relative;
      color: var(--st-ink);
      font: 600 22px/1.7 var(--font-serif);
      letter-spacing: 0.04em;
      border: 0;
      background: none;

      &::before {
        content: '\201C';
        position: absolute;
        left: -0.05em;
        top: 0.02em;
        font: 700 2.6em/1 Georgia, 'Times New Roman', serif;
        color: var(--ink);
      }

      /* 右引号：跟在最后一段文字末尾，与左引号成对 */
      > :last-child::after {
        content: '\201D';
        display: inline-block;
        margin-left: 0.12em;
        font: 700 1.6em/0 Georgia, 'Times New Roman', serif;
        vertical-align: -0.42em;
        color: var(--ink);
      }

      p { margin: 0; }
    }

    code { font: 14.5px var(--font-mono); background: var(--well); }
    ul, ol { padding-left: 1.2em; margin: 0 0 1.1em; }
    li::marker { color: var(--ink); }
    li p { margin: 0 0 0.3em; }
    img { margin: 0.6em 0; }
  }
}

.md-src {
  width: 100%;
  min-height: 60vh;
  border: 0;
  outline: 0;
  resize: vertical;
  padding: 18px 20px;
  border-radius: var(--r-md);
  background: var(--well);
  color: var(--st-ink);
  font: 14px/1.9 var(--font-mono);
}

.ed-foot {
  position: absolute;
  left: 50%;
  bottom: 20px;
  transform: translateX(-50%);
  display: flex;
  gap: 8px;
  align-items: center;
  font-size: 12.5px;
  color: var(--st-ink-3);
  padding: 8px 16px;
  border-radius: var(--r-sm);
  background: color-mix(in oklab, var(--paper) 85%, transparent);
  backdrop-filter: blur(10px);
  transition: transform var(--dur-slow) var(--ease-spring);

  .mono { font-size: 12px; color: var(--st-ink-2); }
  .drawer & { transform: translateX(calc(-50% - 180px)); }
}

/* ---------- 抽屉 ---------- */
.drawer-p {
  position: absolute;
  right: 14px;
  top: 14px;
  bottom: 14px;
  width: 360px;
  z-index: 4;
  border-radius: var(--r-lg);
  background: var(--paper);
  box-shadow: var(--sh-pop);
  transform: translateX(calc(100% + 30px));
  /* 收起：滑出后再隐藏（visibility 延迟切换），读屏与 Tab 也不会进到里面 */
  visibility: hidden;
  transition: transform var(--dur-slow) var(--ease-spring), visibility 0s linear var(--dur-slow);
  display: flex;
  flex-direction: column;

  .drawer & { transform: none; visibility: visible; transition: transform var(--dur-slow) var(--ease-spring), visibility 0s; }
}

:root[data-mode='dark'] .drawer-p { background: var(--well); }

.dr-h {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 18px 10px 24px;

  h3 { font: 600 17px var(--font-serif); margin: 0; }
}

.dr-f {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 14px 24px 18px;
  border-top: 1px solid var(--line);
}

.pick-btn { margin-top: 10px; }

:global(.st-modal.pick) { width: min(720px, calc(100vw - 32px)); }

.pick-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 10px;
  max-height: 50vh;
  overflow: auto;
  padding: 4px;

  button {
    position: relative;
    aspect-ratio: 4 / 3;
    border-radius: var(--r-sm);
    overflow: hidden;
    background: var(--well-2);
    box-shadow: 0 0 0 1px var(--line);
    transition: all var(--dur-fast) var(--ease-out);

    img { width: 100%; height: 100%; object-fit: cover; display: block; }

    &:hover { transform: translateY(-2px); }
    &.on { box-shadow: 0 0 0 2px var(--paper), 0 0 0 3.5px var(--st-ink); }
    &.full { opacity: 0.45; }
  }

  .n {
    position: absolute;
    right: 6px;
    top: 6px;
    min-width: 22px;
    height: 22px;
    border-radius: var(--r-sm);
    display: grid;
    place-items: center;
    font: 600 12px var(--font-mono);
    color: var(--on-solid);
    background: var(--solid);
    box-shadow: var(--btn-shadow);
    transform: scale(0);
    transition: transform var(--dur) var(--ease-bounce);
  }

  .on .n { transform: scale(1); }
}

.cnt { margin-right: auto; font-size: 12.5px; color: var(--st-ink-3); }

.dr-b {
  flex: 1;
  overflow: auto;
  padding: 8px 24px 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;

  .warn { color: var(--red); }
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 7px;
  border-radius: var(--r-sm);
  background: var(--well);
  box-shadow: 0 0 0 1px var(--line) inset;
  min-height: 44px;
  align-items: center;
  cursor: text;

  &:focus-within { background: var(--paper); box-shadow: 0 0 0 1px color-mix(in oklab, var(--ink) 70%, transparent) inset, 0 0 0 3px color-mix(in oklab, var(--ink) 18%, transparent); }

  span {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    height: 28px;
    padding: 0 6px 0 10px;
    border-radius: var(--r-xs);
    background: var(--paper);
    box-shadow: 0 0 0 1px var(--line-2);
    font-size: 13px;
    animation: tag-in var(--dur) var(--ease-spring);
  }

  span button {
    width: 18px;
    height: 18px;
    display: grid;
    place-items: center;
    border-radius: var(--r-xs);
    color: var(--st-ink-3);

    &:hover { background: var(--hover); color: var(--st-ink); }
  }

  input {
    flex: 1;
    min-width: 80px;
    border: 0;
    outline: 0;
    background: none;
    font-size: 13px;
    padding: 0 4px;
    height: 28px;
  }
}

@keyframes tag-in { from { opacity: 0; transform: scale(0.6); } }

.slug {
  display: flex;
  align-items: center;
  height: 38px;
  border-radius: var(--r-sm);
  background: var(--well);
  box-shadow: 0 0 0 1px var(--line) inset;
  padding: 0 12px;
  font: 12.5px var(--font-mono);
  color: var(--st-ink-3);

  input {
    flex: 1;
    min-width: 0;
    border: 0;
    outline: 0;
    background: none;
    font: 12.5px var(--font-mono);
    color: var(--st-ink);
  }

  &:focus-within { box-shadow: 0 0 0 1px color-mix(in oklab, var(--ink) 70%, transparent) inset, 0 0 0 3px color-mix(in oklab, var(--ink) 18%, transparent); background: var(--paper); }
}

.opts { display: flex; flex-direction: column; gap: 18px; }

.row-opt {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 14px;

  small { display: block; font-size: 12px; color: var(--st-ink-3); }
}

.reveal {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows var(--dur) var(--ease-out);

  > div { overflow: hidden; }
  &.on { grid-template-rows: 1fr; }
}

.dt {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  padding-top: 12px;

  input { font-size: 13px; }
}

.sched-note {
  display: flex;
  gap: 6px;
  align-items: flex-start;
  margin: 10px 0 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--st-ink-3);

  .st-ic { margin-top: 2px; }
}
</style>
