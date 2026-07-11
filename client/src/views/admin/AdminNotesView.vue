<script setup lang="ts">
import MarkdownIt from 'markdown-it';
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, api, type Note } from '../../api';
import { useDialogStore } from '../../stores/dialog';

const { t } = useI18n();

const md = new MarkdownIt({ linkify: true });
const notes = ref<Note[]>([]);
const editingId = ref<number | null>(null);
const contentMd = ref('');
const mood = ref('');
const imagesText = ref('');
const busy = ref(false);

async function load(): Promise<void> {
  notes.value = (await api.notes({ pageSize: 50 })).items;
}

function startEdit(note: Note): void {
  editingId.value = note.id;
  contentMd.value = note.contentMd;
  mood.value = note.mood;
  imagesText.value = note.images.join('\n');
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

function resetComposer(): void {
  editingId.value = null;
  contentMd.value = '';
  mood.value = '';
  imagesText.value = '';
}

async function submit(): Promise<void> {
  if (busy.value || !contentMd.value.trim()) return;
  busy.value = true;
  try {
    const body = {
      contentMd: contentMd.value,
      mood: mood.value,
      images: imagesText.value.split(/\n+/).map((s) => s.trim()).filter(Boolean),
    };
    if (editingId.value === null) await adminApi.createNote(body);
    else await adminApi.updateNote(editingId.value, body);
    resetComposer();
    await load();
  } finally {
    busy.value = false;
  }
}

async function remove(note: Note): Promise<void> {
  const ok = await useDialogStore().confirm({
    title: t('admin.delete'),
    message: t('admin.confirmDeleteNote'),
    danger: true,
  });
  if (!ok) return;
  await adminApi.deleteNote(note.id);
  if (editingId.value === note.id) resetComposer();
  await load();
}

onMounted(load);
</script>

<template>
  <div>
    <header class="a-head">
      <div>
        <h1>{{ t('admin.menuNotes') }}</h1>
        <p>{{ t('admin.notesHint') }}</p>
      </div>
    </header>

    <!-- 发布 / 编辑 -->
    <div class="composer a-card" :class="{ editing: editingId !== null }">
      <div v-if="editingId !== null" class="edit-flag">
        {{ t('admin.editingNote', { id: editingId }) }}
        <button class="op" @click="resetComposer">{{ t('admin.cancel') }}</button>
      </div>
      <textarea
        v-model="contentMd"
        rows="4"
        class="a-input"
        :placeholder="t('admin.notePlaceholder')"
      />
      <textarea
        v-model="imagesText"
        rows="2"
        class="a-input"
        :placeholder="t('admin.noteImagesPlaceholder')"
      />
      <div class="composer-bar">
        <input v-model="mood" class="a-input" type="text" :placeholder="t('admin.moodPlaceholder')" />
        <button class="a-btn primary" :disabled="busy || !contentMd.trim()" @click="submit">
          {{ editingId === null ? t('admin.publishNote') : t('admin.saveNote') }}
        </button>
      </div>
    </div>

    <!-- 卡片流 -->
    <div class="note-grid">
      <article
        v-for="note in notes"
        :key="note.id"
        class="note-card a-card"
        :class="{ on: editingId === note.id }"
      >
        <!-- Markdown 渲染预览 -->
        <div class="note-text" v-html="md.render(note.contentMd)" />
        <div v-if="note.images.length" class="thumbs">
          <img v-for="src in note.images.slice(0, 4)" :key="src" :src="src" loading="lazy" alt="" />
          <span v-if="note.images.length > 4" class="more">+{{ note.images.length - 4 }}</span>
        </div>
        <footer class="note-foot">
          <span v-if="note.mood" class="mood">{{ note.mood }}</span>
          <time>{{ note.createdAt.slice(0, 16) }}</time>
          <span class="spacer" />
          <button class="op" @click="startEdit(note)">{{ t('admin.edit') }}</button>
          <button class="op danger" @click="remove(note)">{{ t('admin.delete') }}</button>
        </footer>
      </article>
    </div>
  </div>
</template>

<style scoped lang="scss">
.composer {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 24px;

  &.editing { border-color: rgba(var(--primary-rgb), 0.5); }
}

.edit-flag {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  font-weight: 600;
  color: var(--primary);
}

.composer-bar {
  display: flex;
  gap: 12px;

  input { flex: 1; }
}

.note-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 14px;
}

.note-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  transition: transform var(--dur-fast) var(--ease-out), border-color var(--dur-fast), box-shadow var(--dur);

  &:hover {
    transform: scale(1.015);
    box-shadow: 0 10px 30px -12px rgba(var(--primary-rgb), 0.25);
  }

  &.on { border-color: rgba(var(--primary-rgb), 0.5); }
}

.note-text {
  font-size: 14px;
  line-height: 1.8;
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;

  :deep(p) { margin: 2px 0; }
  :deep(strong) { color: var(--primary); }

  :deep(code) {
    font-family: Consolas, 'Courier New', monospace;
    font-size: 0.88em;
    background: var(--surface-2);
    padding: 1px 6px;
    border-radius: 6px;
  }

  :deep(a) { color: var(--primary); }
}

.thumbs {
  display: flex;
  gap: 6px;
  align-items: center;

  img {
    width: 52px;
    height: 52px;
    object-fit: cover;
    border-radius: 8px;
    background: var(--surface-2);
  }

  .more {
    font-size: 12px;
    font-weight: 700;
    color: var(--text-2);
  }
}

.note-foot {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  color: var(--text-2);
  margin-top: auto;

  .mood {
    padding: 2px 10px;
    border-radius: 999px;
    background: rgba(var(--primary-rgb), 0.1);
    color: var(--primary);
    font-weight: 600;
  }

  .spacer { flex: 1; }
}

.op {
  border: none;
  background: none;
  font-size: 13px;
  font-weight: 600;
  color: var(--primary);

  &.danger { color: var(--accent-red); }
  &:hover { opacity: 0.75; }
}
</style>
