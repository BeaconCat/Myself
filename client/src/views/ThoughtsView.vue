<script setup lang="ts">
import { ref, useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';
import ImageViewer from '../components/media/ImageViewer.vue';
import NoteCard from '../components/thoughts/NoteCard.vue';
import MediaWall from '../components/thoughts/MediaWall.vue';
import DateRangePicker from '../components/thoughts/DateRangePicker.vue';
import { useNotesFeed } from '../composables/useNotesFeed';
import { useConfigStore } from '../stores/config';

const { t } = useI18n();
const config = useConfigStore();

/* 触底哨兵（交给 useNotesFeed 观测） */
const sentinel = useTemplateRef<HTMLElement>('sentinel');

/* 帖文 / 媒体 双视图 + 搜索 + 时间筛选 + 分段加载（逻辑见 useNotesFeed） */
const {
  PAGE_SIZE,
  notes,
  tab,
  keyword,
  loading,
  loadingMore,
  hasMore,
  dateFrom,
  dateTo,
  reload,
  switchTab,
  onSearch,
  applyQuickRange,
  clearDate,
} = useNotesFeed(sentinel);

/* Lightbox 状态 */
const viewerImages = ref<string[]>([]);
const viewerIndex = ref(0);
const viewerOpen = ref(false);

function openViewer(images: string[], index: number): void {
  viewerImages.value = images;
  viewerIndex.value = index;
  viewerOpen.value = true;
}
</script>

<template>
  <main class="page">
    <h1 v-reveal class="page-title">{{ t('nav.thoughts') }}</h1>
    <p v-reveal class="page-sub">{{ config.cfg.thoughts.subtitle }}</p>

    <!-- 工具栏：帖文/媒体胶囊 + 搜索 -->
    <div v-reveal class="toolbar">
      <div class="tabs">
        <button class="tab" :class="{ on: tab === 'posts' }" @click="switchTab('posts')">
          {{ t('thoughts.tabPosts') }}
        </button>
        <button class="tab" :class="{ on: tab === 'media' }" @click="switchTab('media')">
          {{ t('thoughts.tabMedia') }}
        </button>
      </div>
      <div class="search">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round">
          <circle cx="11" cy="11" r="7" />
          <path d="M20 20l-3.8-3.8" />
        </svg>
        <input
          v-model="keyword"
          type="search"
          :placeholder="t('thoughts.searchPlaceholder')"
          @input="onSearch"
        />
      </div>

      <!-- 时间选择器 -->
      <DateRangePicker
        v-model:from="dateFrom"
        v-model:to="dateTo"
        @quick="applyQuickRange"
        @clear="clearDate"
        @change="reload"
      />
    </div>

    <!-- 媒体墙（X 风） -->
    <MediaWall v-if="tab === 'media'" :notes="notes" :loading="loading" @open="openViewer" />

    <!-- X 风信息流 -->
    <div v-else class="feed" :class="{ loading }">
      <NoteCard
        v-for="(note, i) in notes"
        :key="note.id"
        :note="note"
        :index="i % PAGE_SIZE"
        @open="openViewer"
      />
    </div>

    <!-- 触底续载 + 到底标注 -->
    <div ref="sentinel" class="sentinel" aria-hidden="true" />
    <p v-if="loadingMore" class="more-hint">{{ t('thoughts.loadingMore') }}</p>
    <p v-else-if="!hasMore && notes.length" class="end-text">{{ config.cfg.site.listEndText }}</p>

    <ImageViewer
      v-if="viewerOpen"
      :images="viewerImages"
      :start-index="viewerIndex"
      @close="viewerOpen = false"
    />
  </main>
</template>

<style scoped lang="scss">
.page {
  max-width: 640px;
  margin: 0 auto;
  padding: 110px 24px 80px;
}

.page-title {
  font-size: clamp(30px, 4vw, 42px);
}

.page-sub {
  color: var(--text-2);
  font-size: 14px;
  margin: 8px 0 30px;
}

/* 工具栏 */
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 18px;
}

.tabs {
  display: flex;
  gap: 2px;
  padding: 4px;
  border-radius: 999px;
  background: var(--surface);
  border: 1px solid var(--border);
}

.tab {
  padding: 7px 20px;
  border: none;
  border-radius: 999px;
  background: none;
  color: var(--text-2);
  font-size: 13px;
  font-weight: 700;
  transition: all var(--dur-fast) var(--ease-out);

  &.on {
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    color: #fff;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.35);
    box-shadow: 0 2px 10px rgba(var(--primary-rgb), 0.45);
  }
}

.search {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  padding: 9px 14px;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--surface);
  transition: border-color var(--dur-fast), box-shadow var(--dur-fast);

  svg { width: 16px; height: 16px; color: var(--text-2); flex-shrink: 0; }

  input {
    flex: 1;
    min-width: 0;
    border: none;
    outline: none;
    background: none;
    color: var(--text);
    font-size: 14px;
    font-family: inherit;

    &::placeholder { color: var(--text-2); }
  }

  &:focus-within {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(var(--primary-rgb), 0.15);
  }
}

.feed {
  border-top: 1px solid var(--border);
  transition: opacity var(--dur-fast);

  &.loading { opacity: 0.55; }
}

/* 触底与到底 */
.sentinel { height: 1px; }

.more-hint, .end-text {
  text-align: center;
  padding: 22px 0;
  font-size: 13px;
  color: var(--text-2);
}

.end-text { font-family: var(--font-serif); letter-spacing: 0.1em; }

@media (max-width: 768px) {
  .page { padding-top: 88px; }
}
</style>
