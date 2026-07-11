<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, type PostDraft } from '../../api';
import RichEditor from '../../components/admin/RichEditor.vue';
import CoverUploader from '../../components/admin/CoverUploader.vue';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const id = computed(() => {
  const raw = route.params.id;
  return raw === 'new' ? null : Number(raw);
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

/** 富文本为主，可切 Markdown 源码微调 */
const sourceMode = ref(false);

function collect(): PostDraft {
  return {
    ...draft.value,
    tags: tagsText.value.split(/[,，\s]+/).filter(Boolean),
  };
}

async function save(status: 'published' | 'draft'): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  message.value = '';
  try {
    const body = { ...collect(), status };
    if (id.value === null) {
      const { id: newId } = await adminApi.createPost(body);
      message.value = t('admin.saved');
      void router.replace(`/admin/posts/${newId}`);
    } else {
      await adminApi.updatePost(id.value, body);
      draft.value.status = status;
      message.value = t('admin.saved');
    }
  } catch (err) {
    message.value = err instanceof Error && err.message === 'slug_exists'
      ? t('admin.slugExists')
      : t('admin.saveFailed');
  } finally {
    busy.value = false;
  }
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
});
</script>

<template>
  <div class="editor-page">
    <header class="head">
      <h1>{{ id === null ? t('admin.newPost') : t('admin.editPost') }}</h1>
      <div class="actions">
        <span v-if="message" class="msg">{{ message }}</span>
        <label class="pin-toggle">
          <input v-model="draft.pinned" type="checkbox" />
          <span>{{ t('admin.pinned') }}</span>
        </label>
        <button class="btn ghost" :disabled="busy" @click="save('draft')">{{ t('admin.saveDraft') }}</button>
        <button class="btn primary" :disabled="busy" @click="save('published')">{{ t('admin.publish') }}</button>
      </div>
    </header>

    <!-- 元信息 -->
    <div class="meta-grid">
      <label>
        <span>{{ t('admin.fieldTitle') }}</span>
        <input v-model="draft.title" type="text" />
      </label>
      <label>
        <span>Slug</span>
        <input v-model="draft.slug" type="text" placeholder="my-post-slug" />
      </label>
      <label>
        <span>{{ t('admin.fieldTags') }}</span>
        <input v-model="tagsText" type="text" :placeholder="t('admin.tagsPlaceholder')" />
      </label>
      <label>
        <span>{{ t('admin.fieldExcerpt') }}</span>
        <input v-model="draft.excerpt" type="text" />
      </label>
      <div class="full field">
        <span class="field-label">{{ t('admin.fieldCovers') }}</span>
        <CoverUploader v-model="draft.covers" :max="3" />
      </div>
    </div>

    <!-- 正文：所见即所得（底层仍以 Markdown 存储） -->
    <div class="editor-wrap">
      <div class="mode-bar">
        <button
          type="button"
          class="mode-btn"
          :class="{ on: !sourceMode }"
          @click="sourceMode = false"
        >{{ t('admin.wysiwyg') }}</button>
        <button
          type="button"
          class="mode-btn"
          :class="{ on: sourceMode }"
          @click="sourceMode = true"
        >Markdown</button>
      </div>

      <RichEditor
        v-if="!sourceMode"
        v-model="draft.contentMd"
        :placeholder="t('admin.mdPlaceholder')"
      />
      <textarea
        v-else
        v-model="draft.contentMd"
        class="md-input"
        :placeholder="t('admin.mdPlaceholder')"
        spellcheck="false"
      />
    </div>
  </div>
</template>

<style scoped lang="scss">
.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;

  h1 { font-size: 24px; }
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.msg {
  font-size: 13px;
  color: var(--primary);
}

.pin-toggle {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-2);
  cursor: pointer;

  input { accent-color: var(--primary); width: 15px; height: 15px; }
}

.btn {
  padding: 9px 20px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 700;
  border: 1px solid transparent;
  transition: all var(--dur-fast) var(--ease-out);

  &.primary {
    color: #fff;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.4);

    &:hover:not(:disabled) { filter: brightness(1.08); transform: scale(1.04); }
  }

  &.ghost {
    background: var(--surface);
    border-color: var(--border);
    color: var(--text);

    &:hover:not(:disabled) { border-color: var(--primary); color: var(--primary); }
  }

  &:disabled { opacity: 0.6; }
}

.meta-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
  margin-bottom: 18px;

  .full { grid-column: 1 / -1; }
}

.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-2);
}

label {
  display: flex;
  flex-direction: column;
  gap: 6px;

  span {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-2);
  }
}

input, textarea {
  padding: 10px 13px;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  font-size: 14px;
  font-family: inherit;
  outline: none;
  resize: vertical;
  transition: border-color var(--dur-fast), box-shadow var(--dur-fast);

  &:focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(var(--primary-rgb), 0.12);
  }
}

.editor-wrap {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.mode-bar {
  display: flex;
  gap: 6px;
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

.md-input {
  font-family: Consolas, 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.8;
  padding: 18px;
  resize: none;
  min-height: 56vh;
}

.md-preview {
  padding: 18px 22px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 10px;
  overflow-y: auto;
  line-height: 1.9;
  font-size: 15px;

  :deep(h1) { font-size: 26px; margin-bottom: 16px; }
  :deep(h2) {
    font-size: 20px;
    margin: 24px 0 10px;
    padding-left: 12px;
    border-left: 4px solid var(--primary);
  }
  :deep(h3) { font-size: 17px; margin: 18px 0 8px; }
  :deep(p) { margin: 10px 0; }
  :deep(ul), :deep(ol) { padding-left: 24px; margin: 10px 0; }
  :deep(code) {
    font-family: Consolas, monospace;
    font-size: 0.88em;
    background: var(--surface-2);
    padding: 2px 6px;
    border-radius: 6px;
  }
  :deep(pre) {
    background: var(--surface-2);
    border-radius: 8px;
    padding: 14px;
    overflow-x: auto;

    code { background: none; padding: 0; }
  }
  :deep(blockquote) {
    border-left: 4px solid rgba(var(--primary-rgb), 0.5);
    padding: 8px 14px;
    margin: 12px 0;
    color: var(--text-2);
  }
  :deep(table) {
    border-collapse: collapse;

    th, td { border: 1px solid var(--border); padding: 7px 12px; }
  }
  :deep(img) { max-width: 100%; border-radius: 8px; }
}

@media (max-width: 1000px) {
  .split { grid-template-columns: 1fr; }
  .meta-grid { grid-template-columns: 1fr; }
}
</style>
