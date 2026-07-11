<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, api } from '../../api';
import CoverUploader from '../../components/admin/CoverUploader.vue';
import RichEditor from '../../components/admin/RichEditor.vue';

/** 沉浸式写随想：居中卡片，自增高文本 + 拼图配图（拖拽排位）+ 心情 + 置顶 */
const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const id = computed(() => {
  const raw = route.query.id;
  const n = typeof raw === 'string' && raw ? Number(raw) : NaN;
  return Number.isFinite(n) ? n : null;
});

const contentMd = ref('');
const mood = ref('');
const images = ref<string[]>([]);
const pinned = ref(false);
const message = ref('');
const busy = ref(false);

async function publish(): Promise<void> {
  if (busy.value || !contentMd.value.trim()) return;
  busy.value = true;
  message.value = '';
  try {
    const body = {
      contentMd: contentMd.value,
      mood: mood.value,
      images: images.value,
      pinned: pinned.value,
    };
    if (id.value === null) {
      await adminApi.createNote(body);
      message.value = t('write.notePublished');
      contentMd.value = '';
      mood.value = '';
      images.value = [];
      pinned.value = false;
    } else {
      await adminApi.updateNote(id.value, body);
      message.value = t('admin.saved');
    }
  } catch {
    message.value = t('admin.saveFailed');
  } finally {
    busy.value = false;
  }
}

function goBack(): void {
  void router.push(id.value === null ? '/thoughts' : '/admin/notes');
}

onMounted(async () => {
  if (id.value !== null) {
    // 编辑：从公开流取该条（后台列表数据源一致）
    const list = await api.notes({ pageSize: 50 });
    const note = list.items.find((n) => n.id === id.value);
    if (note) {
      contentMd.value = note.contentMd;
      mood.value = note.mood;
      images.value = [...note.images];
      pinned.value = note.pinned;
    }
  }
});
</script>

<template>
  <div class="note-page">
    <header class="bar">
      <button class="back" :title="t('write.back')" @click="goBack">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round">
          <path d="M19 12H5M11 6l-6 6 6 6" />
        </svg>
      </button>
      <span class="bar-title">{{ id === null ? t('write.newNote') : t('write.editNote') }}</span>
      <span v-if="message" class="msg">{{ message }}</span>
    </header>

    <main class="stage">
      <div class="card">
        <!-- 轻量所见即所得（底层仍存 Markdown） -->
        <RichEditor v-model="contentMd" lite :placeholder="t('admin.notePlaceholder')" />

        <!-- 配图拼图：宫格预览 + 拖拽换位 -->
        <CoverUploader v-model="images" :max="9" square />

        <footer class="foot">
          <input
            v-model="mood"
            class="mood a-input"
            type="text"
            :placeholder="t('admin.moodPlaceholder')"
          />
          <label class="pin">
            <input v-model="pinned" type="checkbox" />
            <i class="track" aria-hidden="true" />
            <span>{{ t('admin.pinned') }}</span>
          </label>
          <button class="a-btn primary send" :disabled="busy || !contentMd.trim()" @click="publish">
            {{ id === null ? t('admin.publishNote') : t('admin.saveNote') }}
          </button>
        </footer>
      </div>

      <p class="tip">{{ t('write.noteTip') }}</p>
    </main>
  </div>
</template>

<style scoped lang="scss">
.note-page {
  min-height: 100vh;
  background:
    radial-gradient(600px 280px at 15% -5%, rgba(var(--primary-rgb), 0.07), transparent 65%),
    radial-gradient(500px 260px at 90% 105%, rgba(var(--primary-rgb), 0.05), transparent 65%),
    var(--bg);
}

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

.stage {
  max-width: 620px;
  margin: 0 auto;
  padding: 46px 18px 80px;
}

.card {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 24px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow);
}

.note-input {
  width: 100%;
  min-height: 140px;
  border: none;
  outline: none;
  background: none;
  color: var(--text);
  font-family: inherit;
  font-size: 16.5px;
  line-height: 1.9;
  resize: none;
  overflow: hidden;

  &::placeholder { color: var(--text-2); opacity: 0.6; }
}

.foot {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-top: 14px;
  border-top: 1px solid var(--border);
}

.mood { flex: 1; min-width: 0; }

/* 置顶滑轨 */
.pin {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  flex-shrink: 0;

  input { display: none; }

  .track {
    width: 36px;
    height: 20px;
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
      width: 14px;
      height: 14px;
      border-radius: 50%;
      background: var(--text-2);
      transition: transform var(--dur-fast) var(--ease-spring), background var(--dur-fast);
    }
  }

  input:checked + .track {
    background: rgba(var(--primary-rgb), 0.25);
    border-color: rgba(var(--primary-rgb), 0.5);

    &::after { transform: translateX(16px); background: var(--primary); }
  }

  span { font-size: 13px; font-weight: 600; color: var(--text-2); }
}

.send { flex-shrink: 0; }

.tip {
  text-align: center;
  margin-top: 18px;
  font-size: 12.5px;
  color: var(--text-2);
}

@media (max-width: 560px) {
  .stage { padding: 22px 12px 60px; }
  .card { padding: 16px; }

  .foot { flex-wrap: wrap; }
  .mood { width: 100%; flex: none; }
}
</style>
