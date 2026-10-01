<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { api, type Note } from '../../api';
import '../admin/studio/i18n';
import SIcon from '../admin/studio/SIcon.vue';
import NoteComposer from '../admin/studio/NoteComposer.vue';
import { refreshCounts } from '../admin/studio/state';
import { toast } from '../admin/studio/toast';
import { dateTimeText } from '../admin/studio/format';

/** 写随想：复用「今天」页的输入框，独立成页（?id= 为编辑） */
const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const note = ref<Note | null>(null);
const missing = ref(false);
const composer = ref<InstanceType<typeof NoteComposer> | null>(null);
const id = computed(() => {
  const n = Number(route.query.id);
  return route.query.id && Number.isFinite(n) ? n : null;
});

/** 公开流分页查找目标随想（后台与前台同源） */
async function find(target: number): Promise<Note | null> {
  for (let page = 1; page <= 20; page += 1) {
    const res = await api.notes({ page, pageSize: 50 });
    const hit = res.items.find((n) => n.id === target);
    if (hit) return hit;
    if (page * 50 >= res.total) break;
  }
  return null;
}

async function load(): Promise<void> {
  missing.value = false;
  note.value = null;
  if (id.value === null) {
    composer.value?.focus();
    return;
  }
  try {
    note.value = await find(id.value);
    missing.value = !note.value;
  } catch {
    missing.value = true;
  }
}
watch(id, () => void load());

function onPublished(): void {
  toast(t('studio.composer.published'), { action: t('studio.view'), fn: () => void router.push({ name: 'admin-notes' }) });
  void refreshCounts();
}

function onSaved(): void {
  void router.push({ name: 'admin-notes' });
}

onMounted(() => void load());
</script>

<template>
  <section class="studio view">
    <router-link class="st-link back" :to="{ name: 'admin-notes' }"><SIcon name="arrowL" :size="18" />{{ t('studio.writeNote.back') }}</router-link>
    <div class="st-vh">
      <div>
        <h1>{{ id === null ? t('studio.writeNote.title') : t('studio.writeNote.editTitle') }}</h1>
        <p v-if="note">{{ t('studio.writeNote.editSub', { when: dateTimeText(note.createdAt, true) }) }}</p>
        <p v-else>{{ t('studio.writeNote.desc') }}</p>
      </div>
    </div>

    <div class="grid">
      <div class="st-rise">
        <p v-if="missing" class="missing">{{ t('studio.writeNote.missing') }}</p>
        <NoteComposer
          v-else
          ref="composer"
          always-open
          :note="note"
          :placeholder="t('studio.composer.placeholderAlt')"
          @published="onPublished"
          @saved="onSaved"
        />
      </div>
      <aside class="tips st-rise" style="--i: 2">
        <h3>{{ t('studio.writeNote.tipsTitle') }}</h3>
        <ul>
          <li><kbd class="st-kbd">Ctrl</kbd><kbd class="st-kbd">Enter</kbd><span>{{ t('studio.writeNote.tipPublish') }}</span></li>
          <li><kbd class="st-kbd">Esc</kbd><span>{{ t('studio.writeNote.tipEsc') }}</span></li>
          <li><kbd class="st-kbd">N</kbd><span>{{ t('studio.writeNote.tipN') }}</span></li>
        </ul>
        <p>{{ t('studio.writeNote.tipMd') }}</p>
        <p>{{ t('studio.writeNote.tipImg') }}</p>
      </aside>
    </div>
  </section>
</template>

<style scoped lang="scss">
.view {
  max-width: 1280px;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

.back { margin-bottom: 20px; }

.grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 48px;
  align-items: start;
}

.missing {
  padding: 40px;
  border-radius: var(--r-lg);
  background: var(--well);
  color: var(--st-ink-3);
  text-align: center;
}

.tips {
  padding: 22px;
  border-radius: var(--r-lg);
  background: var(--well);
  font-size: 13px;
  color: var(--st-ink-3);
  line-height: 1.7;

  h3 { font: 600 16px var(--font-serif); color: var(--st-ink); margin: 0 0 14px; }

  ul {
    list-style: none;
    margin: 0 0 14px;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  li {
    display: flex;
    align-items: center;
    gap: 4px;

    span { margin-left: 8px; }
  }

  p { margin: 0 0 8px; }
}

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
  .grid { grid-template-columns: 1fr; }
}
</style>
