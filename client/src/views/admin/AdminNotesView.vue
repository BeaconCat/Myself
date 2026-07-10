<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, api, type Note } from '../../api';

const { t } = useI18n();

const notes = ref<Note[]>([]);
const contentMd = ref('');
const mood = ref('');
const imagesText = ref('');
const busy = ref(false);

async function load(): Promise<void> {
  notes.value = (await api.notes({ pageSize: 50 })).items;
}

async function publish(): Promise<void> {
  if (busy.value || !contentMd.value.trim()) return;
  busy.value = true;
  try {
    await adminApi.createNote({
      contentMd: contentMd.value,
      mood: mood.value,
      images: imagesText.value.split(/\n+/).map((s) => s.trim()).filter(Boolean),
    });
    contentMd.value = '';
    mood.value = '';
    imagesText.value = '';
    await load();
  } finally {
    busy.value = false;
  }
}

async function remove(note: Note): Promise<void> {
  if (!window.confirm(t('admin.confirmDeleteNote'))) return;
  await adminApi.deleteNote(note.id);
  await load();
}

onMounted(load);
</script>

<template>
  <div>
    <h1 class="page-h">{{ t('admin.menuNotes') }}</h1>

    <!-- 发布框 -->
    <div class="composer">
      <textarea
        v-model="contentMd"
        rows="4"
        :placeholder="t('admin.notePlaceholder')"
      />
      <textarea
        v-model="imagesText"
        rows="2"
        :placeholder="t('admin.noteImagesPlaceholder')"
      />
      <div class="composer-bar">
        <input v-model="mood" type="text" :placeholder="t('admin.moodPlaceholder')" />
        <button class="btn primary" :disabled="busy || !contentMd.trim()" @click="publish">
          {{ t('admin.publishNote') }}
        </button>
      </div>
    </div>

    <!-- 列表 -->
    <ul class="note-list">
      <li v-for="note in notes" :key="note.id">
        <div class="note-main">
          <p class="note-text">{{ note.contentMd }}</p>
          <div class="note-meta">
            <span v-if="note.mood" class="mood">{{ note.mood }}</span>
            <span v-if="note.images.length" class="imgs">{{ t('admin.imageCount', { n: note.images.length }) }}</span>
            <time>{{ note.createdAt }}</time>
          </div>
        </div>
        <button class="op danger" @click="remove(note)">{{ t('admin.delete') }}</button>
      </li>
    </ul>
  </div>
</template>

<style scoped lang="scss">
.page-h {
  font-size: 26px;
  margin-bottom: 22px;
}

.composer {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 18px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  margin-bottom: 26px;
}

textarea, input {
  padding: 11px 14px;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: var(--bg);
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

.composer-bar {
  display: flex;
  gap: 12px;

  input { flex: 1; }
}

.btn.primary {
  padding: 10px 26px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 700;
  color: #fff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
  background: linear-gradient(180deg, var(--primary), var(--primary-deep));
  box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.4);
  transition: transform var(--dur-fast) var(--ease-out), filter var(--dur-fast);

  &:hover:not(:disabled) { filter: brightness(1.08); transform: scale(1.04); }
  &:disabled { opacity: 0.55; }
}

.note-list {
  list-style: none;
  display: flex;
  flex-direction: column;

  li {
    display: flex;
    align-items: flex-start;
    gap: 14px;
    padding: 16px 6px;
    border-bottom: 1px solid var(--border);
  }
}

.note-main { flex: 1; min-width: 0; }

.note-text {
  font-size: 14px;
  line-height: 1.8;
  white-space: pre-wrap;
  word-break: break-word;
}

.note-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-2);

  .mood {
    padding: 2px 10px;
    border-radius: 999px;
    background: rgba(var(--primary-rgb), 0.1);
    color: var(--primary);
    font-weight: 600;
  }
}

.op {
  border: none;
  background: none;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;

  &.danger { color: var(--accent-red); }
  &:hover { opacity: 0.75; }
}
</style>
