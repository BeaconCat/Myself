<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { api, type Post, type Tag } from '../api';

const { t } = useI18n();

const posts = ref<Post[]>([]);
const tags = ref<Tag[]>([]);
const activeTag = ref('');
const loading = ref(false);

async function load(tag = ''): Promise<void> {
  loading.value = true;
  try {
    activeTag.value = tag;
    const list = await api.posts({ tag: tag || undefined, pageSize: 20 });
    posts.value = list.items;
  } finally {
    loading.value = false;
  }
}

function formatDate(s: string): string {
  return s.slice(0, 10);
}

onMounted(async () => {
  await load();
  tags.value = await api.tags();
});
</script>

<template>
  <main class="page">
    <h1 v-reveal class="page-title">{{ t('nav.articles') }}</h1>

    <!-- 标签过滤 -->
    <div v-reveal class="tag-bar">
      <button
        class="tag-chip"
        :class="{ on: !activeTag }"
        @click="load()"
      >{{ t('articles.all') }}</button>
      <button
        v-for="tag in tags"
        :key="tag.name"
        class="tag-chip"
        :class="{ on: activeTag === tag.name }"
        @click="load(tag.name)"
      >{{ tag.name }}<i>{{ tag.count }}</i></button>
    </div>

    <!-- 列表 -->
    <transition-group name="list" tag="div" class="post-grid" :class="{ loading }">
      <router-link
        v-for="post in posts"
        :key="post.slug"
        :to="`/articles/${post.slug}`"
        class="post-card hover-lift"
      >
        <div class="post-body">
          <h2>{{ post.title }}</h2>
          <p>{{ post.excerpt }}</p>
          <div class="post-meta">
            <span class="date">{{ formatDate(post.createdAt) }}</span>
            <span v-for="tag in post.tags" :key="tag" class="mini-tag">{{ tag }}</span>
          </div>
        </div>
        <span class="post-arrow" aria-hidden="true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round">
            <path d="M5 12h14M13 6l6 6-6 6" />
          </svg>
        </span>
      </router-link>
    </transition-group>

    <p v-if="!loading && !posts.length" class="empty">{{ t('articles.empty') }}</p>
  </main>
</template>

<style scoped lang="scss">
.page {
  max-width: 900px;
  margin: 0 auto;
  padding: 110px 24px 80px;
}

.page-title {
  font-size: clamp(30px, 4vw, 42px);
  margin-bottom: 26px;
}

/* 标签栏 */
.tag-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 34px;
}

.tag-chip {
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

/* 列表 */
.post-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
  transition: opacity var(--dur-fast);

  &.loading { opacity: 0.55; }
}

.post-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 22px 24px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
  background: var(--surface);
  position: relative;
  overflow: hidden;

  /* 左缘主色细条 */
  &::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 3px;
    background: linear-gradient(180deg, var(--primary), transparent);
    opacity: 0;
    transition: opacity var(--dur-fast);
  }

  &:hover::before { opacity: 1; }
  &:hover .post-arrow { transform: translateX(4px); color: var(--primary); }
}

.post-body {
  flex: 1;
  min-width: 0;

  h2 {
    font-size: 20px;
    margin-bottom: 8px;
    transition: color var(--dur-fast);
  }

  p {
    font-size: 14px;
    line-height: 1.7;
    color: var(--text-2);
  }
}

.post-card:hover h2 { color: var(--primary); }

.post-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;

  .date { font-size: 12px; color: var(--text-2); }
}

.mini-tag {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 9px;
  border-radius: 999px;
  background: rgba(var(--primary-rgb), 0.1);
  color: var(--primary);
}

.post-arrow {
  width: 22px;
  height: 22px;
  color: var(--text-2);
  transition: transform var(--dur-fast) var(--ease-out), color var(--dur-fast);

  svg { width: 100%; height: 100%; }
}

.empty {
  color: var(--text-2);
  text-align: center;
  padding: 48px 0;
}

/* 过滤切换动画 */
.list-enter-active { transition: all var(--dur) var(--ease-out); }
.list-leave-active { transition: all var(--dur-fast) ease; position: absolute; opacity: 0; }
.list-enter-from { opacity: 0; transform: translateY(18px); }
.list-move { transition: transform var(--dur) var(--ease-out); }

@media (max-width: 768px) {
  .page { padding-top: 88px; }
}
</style>
