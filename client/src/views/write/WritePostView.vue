<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, type PostDraft } from '../../api';
import RichEditor from '../../components/admin/RichEditor.vue';
import CoverUploader from '../../components/admin/CoverUploader.vue';

/** 沉浸式写文章：大标题 + 全宽富文本，元信息折叠在下方 */
const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const id = computed(() => {
  const raw = route.query.id;
  const n = typeof raw === 'string' && raw ? Number(raw) : NaN;
  return Number.isFinite(n) ? n : null;
});

const draft = ref<PostDraft>({
  slug: '',
  title: '',
  excerpt: '',
  contentMd: '',
  covers: [],
  tags: [],
  status: 'draft',
  pinned: false,
});

const tagsText = ref('');
const message = ref('');
const busy = ref(false);
const metaOpen = ref(false);
const sourceMode = ref(false);

/** 标题自动生成 slug（仅新建且未手改时） */
const slugTouched = ref(false);

function autoSlug(): void {
  if (id.value !== null || slugTouched.value || !draft.value.title) return;
  draft.value.slug = draft.value.title
    .toLowerCase()
    .replace(/[^\w一-龥]+/g, '-')
    .replace(/[一-龥]/g, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 60) || `post-${Date.now().toString(36)}`;
}

async function save(status: 'published' | 'draft'): Promise<void> {
  if (busy.value) return;
  if (!draft.value.title.trim()) {
    message.value = t('write.needTitle');
    return;
  }
  autoSlug();
  if (!draft.value.slug) draft.value.slug = `post-${Date.now().toString(36)}`;
  busy.value = true;
  message.value = '';
  try {
    const body = {
      ...draft.value,
      tags: tagsText.value.split(/[,，\s]+/).filter(Boolean),
      status,
    };
    if (id.value === null) {
      const { id: newId } = await adminApi.createPost(body);
      message.value = status === 'published' ? t('write.published') : t('admin.saved');
      void router.replace({ path: '/write/post', query: { id: String(newId) } });
    } else {
      await adminApi.updatePost(id.value, body);
      draft.value.status = status;
      message.value = status === 'published' ? t('write.published') : t('admin.saved');
    }
  } catch (err) {
    message.value = err instanceof Error && err.message === 'slug_exists'
      ? t('admin.slugExists')
      : t('admin.saveFailed');
  } finally {
    busy.value = false;
  }
}

function goBack(): void {
  void router.push('/admin/posts');
}

onMounted(async () => {
  if (id.value === null) return;
  const post = await adminApi.post(id.value);
  draft.value = {
    slug: post.slug,
    title: post.title,
    excerpt: post.excerpt,
    contentMd: post.contentMd ?? '',
    covers: post.covers,
    tags: post.tags,
    status: post.status,
    pinned: post.pinned,
  };
  tagsText.value = post.tags.join(', ');
  slugTouched.value = true;
});
</script>

<template>
  <div class="write-page">
    <!-- 顶部薄工具条 -->
    <header class="bar">
      <button class="back" :title="t('write.back')" @click="goBack">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round">
          <path d="M19 12H5M11 6l-6 6 6 6" />
        </svg>
      </button>
      <span class="bar-title">{{ id === null ? t('write.newPost') : t('write.editPost') }}</span>
      <span v-if="message" class="msg">{{ message }}</span>
      <span class="spacer" />
      <label class="pin">
        <input v-model="draft.pinned" type="checkbox" />
        <span>{{ t('admin.pinned') }}</span>
      </label>
      <button class="a-btn ghost" :disabled="busy" @click="save('draft')">{{ t('admin.saveDraft') }}</button>
      <button class="a-btn primary" :disabled="busy" @click="save('published')">{{ t('admin.publish') }}</button>
    </header>

    <main class="paper">
      <!-- 大标题 -->
      <input
        v-model="draft.title"
        class="title-input"
        type="text"
        :placeholder="t('write.titlePlaceholder')"
        @input="autoSlug"
      />

      <!-- 正文 -->
      <div class="mode-row">
        <button type="button" class="mode-btn" :class="{ on: !sourceMode }" @click="sourceMode = false">
          {{ t('admin.wysiwyg') }}
        </button>
        <button type="button" class="mode-btn" :class="{ on: sourceMode }" @click="sourceMode = true">
          Markdown
        </button>
      </div>

      <RichEditor
        v-if="!sourceMode"
        v-model="draft.contentMd"
        :placeholder="t('admin.mdPlaceholder')"
      />
      <textarea
        v-else
        v-model="draft.contentMd"
        class="md-src a-input"
        spellcheck="false"
        :placeholder="t('admin.mdPlaceholder')"
      />

      <!-- 折叠元信息 -->
      <section class="meta" :class="{ open: metaOpen }">
        <button class="meta-toggle" @click="metaOpen = !metaOpen">
          <span>{{ t('write.metaTitle') }}</span>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M6 9l6 6 6-6" />
          </svg>
        </button>
        <div v-if="metaOpen" class="meta-body">
          <div class="row2">
            <label>
              <span>Slug</span>
              <input v-model="draft.slug" class="a-input" type="text" @input="slugTouched = true" />
            </label>
            <label>
              <span>{{ t('admin.fieldTags') }}</span>
              <input v-model="tagsText" class="a-input" type="text" :placeholder="t('admin.tagsPlaceholder')" />
            </label>
          </div>
          <label>
            <span>{{ t('admin.fieldExcerpt') }}</span>
            <input v-model="draft.excerpt" class="a-input" type="text" />
          </label>
          <div class="field">
            <span class="f-label">{{ t('admin.fieldCovers') }}</span>
            <CoverUploader v-model="draft.covers" :max="3" />
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped lang="scss">
.write-page {
  min-height: 100vh;
  background:
    radial-gradient(700px 300px at 85% -5%, rgba(var(--primary-rgb), 0.06), transparent 65%),
    var(--bg);
}

/* 顶条 */
.bar {
  position: sticky;
  top: 0;
  z-index: 40;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 18px;
  background: var(--glass);
  backdrop-filter: blur(14px) saturate(1.4);
  -webkit-backdrop-filter: blur(14px) saturate(1.4);
  border-bottom: 1px solid var(--border);
}

.back {
  width: 36px;
  height: 36px;
  border: 1px solid var(--border);
  border-radius: 50%;
  background: var(--surface);
  color: var(--text);
  display: grid;
  place-items: center;
  transition: all var(--dur-fast) var(--ease-out);

  svg { width: 17px; height: 17px; }
  &:hover { border-color: var(--primary); color: var(--primary); transform: scale(1.08); }
}

.bar-title { font-size: 14px; font-weight: 700; }
.msg { font-size: 12.5px; color: var(--primary); }
.spacer { flex: 1; }

.pin {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-2);
  cursor: pointer;

  input { accent-color: var(--primary); width: 15px; height: 15px; }
}

/* 写作纸面 */
.paper {
  max-width: 860px;
  margin: 0 auto;
  padding: 34px 24px 90px;
}

.title-input {
  width: 100%;
  border: none;
  outline: none;
  background: none;
  color: var(--text);
  font-family: var(--font-serif);
  font-size: clamp(26px, 4vw, 38px);
  font-weight: 700;
  line-height: 1.4;
  padding: 8px 2px 16px;

  &::placeholder { color: var(--text-2); opacity: 0.5; }
}

.mode-row {
  display: flex;
  gap: 6px;
  margin-bottom: 10px;
}

.mode-btn {
  padding: 6px 16px;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text-2);
  font-size: 12.5px;
  font-weight: 700;
  transition: all var(--dur-fast);

  &.on {
    background: rgba(var(--primary-rgb), 0.12);
    border-color: rgba(var(--primary-rgb), 0.4);
    color: var(--primary);
  }
}

.md-src {
  width: 100%;
  min-height: 56vh;
  font-family: Consolas, 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.8;
  padding: 18px;
  resize: vertical;
}

/* 元信息折叠 */
.meta {
  margin-top: 20px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--surface);
  overflow: hidden;

  &.open { border-color: rgba(var(--primary-rgb), 0.35); }
}

.meta-toggle {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 14px 18px;
  border: none;
  background: none;
  color: var(--text);
  font-size: 14px;
  font-weight: 700;

  svg {
    width: 16px;
    height: 16px;
    transition: transform var(--dur-fast) var(--ease-out);
  }
}

.meta.open .meta-toggle svg { transform: rotate(180deg); }

.meta-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 4px 18px 18px;
}

.row2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

label, .field {
  display: flex;
  flex-direction: column;
  gap: 6px;

  span, .f-label { font-size: 12px; font-weight: 600; color: var(--text-2); }
}

@media (max-width: 700px) {
  .bar { flex-wrap: wrap; padding: 10px 12px; }
  .bar-title { display: none; }
  .paper { padding: 20px 14px 70px; }
  .row2 { grid-template-columns: 1fr; }
}
</style>
