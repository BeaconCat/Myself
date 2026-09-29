<script setup lang="ts">
import Icon from '../components/ui/Icon.vue';
import { ArrowLeft, ChevronLeft, ChevronRight } from 'lucide';
import { computed, nextTick, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { api, type Note } from '../api';
import { useIdentity } from '../about/useIdentity';
import { useConfigStore } from '../stores/config';
import { useLoadingStore } from '../stores/loading';
import NoteCard from '../components/thoughts/NoteCard.vue';
import CommentSection from '../components/engage/CommentSection.vue';
import ImageViewer, { type OriginRect } from '../components/media/ImageViewer.vue';

/**
 * 随想详情（/thoughts/:id）：正文放大的单条随想 + 互动栏 + 评论区 + 更早 / 更新一条。
 * 返回：从随想列表点进来时 history 回退（列表保活，停在原来的时间轴位置）；直接打开则回到随想列表。
 * 带 #comments 进入时滚到评论区。
 */
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const config = useConfigStore();
const { avatar, fullName } = useIdentity();
const handle = computed(() => config.cfg.github.username || 'myself');

const note = ref<Note | null>(null);
const older = ref(0);
const newer = ref(0);
const missing = ref(false);
const id = computed(() => Number(route.params.id));

async function load(): Promise<void> {
  if (!Number.isFinite(id.value)) {
    missing.value = true;
    return;
  }
  const release = useLoadingStore().holdRoute();
  try {
    const res = await api.note(id.value);
    note.value = res.note;
    older.value = res.older;
    newer.value = res.newer;
    missing.value = false;
  } catch {
    note.value = null;
    missing.value = true;
  } finally {
    release();
  }
  if (route.hash === '#comments') {
    await nextTick();
    window.setTimeout(() => document.getElementById('comments')?.scrollIntoView({ behavior: 'smooth', block: 'start' }), 350);
  }
}

watch(id, load, { immediate: true });

function back(): void {
  const prev = (window.history.state as { back?: string } | null)?.back ?? '';
  if (prev.startsWith('/thoughts') || prev === '/') router.back();
  else void router.push('/thoughts');
}

const viewer = ref<{ images: string[]; index: number; rect?: OriginRect } | null>(null);
function openViewer(images: string[], i: number, rect: DOMRect): void {
  viewer.value = { images, index: i, rect: { left: rect.left, top: rect.top, width: rect.width, height: rect.height } };
}
</script>

<template>
  <main class="nd">
    <div class="bar">
      <button type="button" class="back" @click="back">
        <Icon :icon="ArrowLeft" :size="18" />
        {{ t('noteDetail.back') }}
      </button>
    </div>

    <template v-if="note">
      <NoteCard class="rise" :note="note" :avatar="avatar" :name="fullName" :handle="handle" detail @open="openViewer" @mood="(m) => router.push({ path: '/thoughts', query: { q: m } })" />

      <nav class="nb rise" style="--i: 1">
        <router-link v-if="older" :to="`/thoughts/${older}`" class="o">
          <Icon :icon="ChevronLeft" :size="16" />{{ t('noteDetail.older') }}
        </router-link>
        <span class="sp" />
        <router-link v-if="newer" :to="`/thoughts/${newer}`" class="n">
          {{ t('noteDetail.newer') }}<Icon :icon="ChevronRight" :size="16" />
        </router-link>
      </nav>

      <CommentSection target="note" :comment-key="String(note.id)" :engage-id="note.id" />
    </template>
    <p v-else-if="missing" class="missing">{{ t('noteDetail.notFound') }}</p>

    <ImageViewer v-if="viewer" :images="viewer.images" :start-index="viewer.index" :origin-rect="viewer.rect" @close="viewer = null" />
  </main>
</template>

<style scoped lang="scss">
.nd {
  max-width: 760px;
  margin: 0 auto;
  padding: 92px 24px 72px;
}

.bar { margin-bottom: 6px; }

.back {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 14px 0 10px;
  margin-left: -10px;
  border: 0;
  border-radius: var(--r-pill);
  background: none;
  color: var(--text-2);
  font-size: 14px;
  cursor: pointer;
  transition: background var(--dur-fast), color var(--dur-fast);

  &:hover { background: var(--fill); color: var(--text); }
  svg { transition: transform var(--dur-fast) var(--ease-spring); }
  &:hover svg { transform: translateX(-3px); }
}

.nb {
  display: flex;
  align-items: center;
  margin-top: 18px;
  padding-top: 14px;
  border-top: 0.5px solid var(--line-2);

  .sp { flex: 1; }

  a {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 14px;
    color: var(--text-2);
    transition: color var(--dur-fast);

    &:hover { color: var(--ink); }
  }
}

.missing { padding: 80px 0; text-align: center; color: var(--text-3); }

@media (max-width: 767px) {
  .nd { padding: 24px 16px 96px; }
}
</style>
