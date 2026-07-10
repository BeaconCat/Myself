<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { api, type Post, type Tag } from '../api';
import PostZigzagList from '../components/post/PostZigzagList.vue';

const { t } = useI18n();

const posts = ref<Post[]>([]);
const tags = ref<Tag[]>([]);
const activeTag = ref('');
const keyword = ref('');
const loading = ref(false);

async function load(): Promise<void> {
  loading.value = true;
  try {
    const list = await api.posts({
      tag: activeTag.value || undefined,
      q: keyword.value || undefined,
      pageSize: 20,
    });
    posts.value = list.items;
  } finally {
    loading.value = false;
  }
}

function pickTag(tag: string): void {
  activeTag.value = tag;
  void load();
}

/* 搜索防抖 */
let debounce = 0;
function onSearch(): void {
  window.clearTimeout(debounce);
  debounce = window.setTimeout(() => void load(), 300);
}

onBeforeUnmount(() => window.clearTimeout(debounce));

onMounted(async () => {
  await load();
  tags.value = await api.tags();
});
</script>

<template>
  <main class="page">
    <h1 v-reveal class="page-title">{{ t('nav.articles') }}</h1>

    <!-- 工具栏：搜索 + 标签筛选一行 -->
    <div v-reveal class="toolbar">
      <div class="search">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round">
          <circle cx="11" cy="11" r="7" />
          <path d="M20 20l-3.8-3.8" />
        </svg>
        <input
          v-model="keyword"
          type="search"
          :placeholder="t('articles.searchPlaceholder')"
          @input="onSearch"
        />
      </div>

      <div class="tag-scroll">
        <button class="tag-chip" :class="{ on: !activeTag }" @click="pickTag('')">
          {{ t('articles.all') }}
        </button>
        <button
          v-for="tag in tags"
          :key="tag.name"
          class="tag-chip"
          :class="{ on: activeTag === tag.name }"
          @click="pickTag(tag.name)"
        >{{ tag.name }}<i>{{ tag.count }}</i></button>
      </div>
    </div>

    <!-- 统一文左图右的窄行列表 -->
    <div class="list-wrap" :class="{ loading }">
      <PostZigzagList :posts="posts" :alternate="false" compact />
    </div>

    <p v-if="!loading && !posts.length" class="empty">{{ t('articles.empty') }}</p>
  </main>
</template>

<style scoped lang="scss">
.page {
  max-width: 1080px;
  margin: 0 auto;
  padding: 110px 24px 80px;
}

.page-title {
  font-size: clamp(30px, 4vw, 42px);
  margin-bottom: 26px;
}

/* 工具栏 */
.toolbar {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 20px;
}

.search {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  width: min(300px, 42vw);
  padding: 9px 14px;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--surface);
  transition: border-color var(--dur-fast), box-shadow var(--dur-fast);

  svg {
    width: 17px;
    height: 17px;
    color: var(--text-2);
    flex-shrink: 0;
  }

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

.tag-scroll {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  scrollbar-width: none;
  padding: 2px;

  &::-webkit-scrollbar { display: none; }
}

.tag-chip {
  flex-shrink: 0;
  font-size: 13px;
  font-weight: 600;
  padding: 7px 16px;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text-2);
  transition: all var(--dur-fast) var(--ease-out);

  i {
    font-style: normal;
    margin-left: 6px;
    font-size: 11px;
    opacity: 0.7;
  }

  &:hover { border-color: var(--primary); color: var(--primary); transform: translateY(-2px); }

  &.on {
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    border-color: transparent;
    color: #fff;
    box-shadow: 0 2px 10px rgba(var(--primary-rgb), 0.45);
  }
}

.list-wrap {
  transition: opacity var(--dur-fast);

  &.loading { opacity: 0.55; }
}

.empty {
  color: var(--text-2);
  text-align: center;
  padding: 48px 0;
}

@media (max-width: 768px) {
  .page { padding-top: 88px; }

  .toolbar { flex-direction: column; align-items: stretch; }
  .search { width: 100%; }
}
</style>
