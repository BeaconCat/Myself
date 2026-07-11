<script setup lang="ts">
import MarkdownIt from 'markdown-it';
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, api, type Note } from '../../api';
import { useDialogStore } from '../../stores/dialog';

const { t } = useI18n();

const md = new MarkdownIt({ linkify: true });
const notes = ref<Note[]>([]);

async function load(): Promise<void> {
  notes.value = (await api.notes({ pageSize: 50 })).items;
}

/** 卡片置顶开关：即时保存 */
async function togglePin(note: Note): Promise<void> {
  note.pinned = !note.pinned;
  try {
    await adminApi.updateNote(note.id, {
      contentMd: note.contentMd,
      mood: note.mood,
      images: note.images,
      pinned: note.pinned,
    });
    await load();
  } catch {
    note.pinned = !note.pinned;
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
      <router-link to="/write/note" class="a-btn primary">{{ t('write.newNote') }}</router-link>
    </header>

    <!-- 卡片流 -->
    <div class="note-grid">
      <article
        v-for="note in notes"
        :key="note.id"
        class="note-card a-card"
        :class="{ pinned: note.pinned }"
      >
        <div class="note-text" v-html="md.render(note.contentMd)" />
        <div v-if="note.images.length" class="thumbs">
          <img v-for="src in note.images.slice(0, 4)" :key="src" :src="src" loading="lazy" alt="" />
          <span v-if="note.images.length > 4" class="more">+{{ note.images.length - 4 }}</span>
        </div>
        <footer class="note-foot">
          <span v-if="note.mood" class="mood">{{ note.mood }}</span>
          <time>{{ note.createdAt.slice(0, 16) }}</time>
          <span class="spacer" />
          <!-- 置顶开关 -->
          <label class="pin" :title="t('admin.pinned')">
            <input type="checkbox" :checked="note.pinned" @change="togglePin(note)" />
            <i class="track" aria-hidden="true" />
          </label>
          <router-link class="op" :to="{ path: '/write/note', query: { id: String(note.id) } }">
            {{ t('admin.edit') }}
          </router-link>
          <button class="op danger" @click="remove(note)">{{ t('admin.delete') }}</button>
        </footer>
      </article>
    </div>
  </div>
</template>

<style scoped lang="scss">
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

  &.pinned { border-color: rgba(var(--primary-rgb), 0.45); }
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

  .more { font-size: 12px; font-weight: 700; color: var(--text-2); }
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

/* 置顶滑轨（迷你） */
.pin {
  cursor: pointer;

  input { display: none; }

  .track {
    display: block;
    width: 32px;
    height: 18px;
    border-radius: 999px;
    background: var(--surface-2);
    border: 1px solid var(--border);
    position: relative;
    transition: all var(--dur-fast);

    &::after {
      content: '';
      position: absolute;
      top: 2px;
      left: 2px;
      width: 12px;
      height: 12px;
      border-radius: 50%;
      background: var(--text-2);
      transition: transform var(--dur-fast) var(--ease-spring), background var(--dur-fast);
    }
  }

  input:checked + .track {
    background: rgba(var(--primary-rgb), 0.25);
    border-color: rgba(var(--primary-rgb), 0.5);

    &::after { transform: translateX(14px); background: var(--primary); }
  }
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
