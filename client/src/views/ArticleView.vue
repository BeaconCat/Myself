<script setup lang="ts">
import MarkdownIt from 'markdown-it';
// @ts-expect-error 无类型声明的离线插件
import taskLists from 'markdown-it-task-lists';
import { computed, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { api, type Post } from '../api';
import CoverAccordion from '../components/post/CoverAccordion.vue';

const { t } = useI18n();
const route = useRoute();

const post = ref<Post | null>(null);
const notFound = ref(false);

const md = new MarkdownIt({ linkify: true }).use(taskLists);

const html = computed(() =>
  post.value?.contentMd ? md.render(post.value.contentMd) : '',
);

watch(
  () => route.params.slug,
  async (slug) => {
    if (typeof slug !== 'string') return;
    notFound.value = false;
    post.value = null;
    try {
      post.value = await api.post(slug);
    } catch {
      notFound.value = true;
    }
  },
  { immediate: true },
);
</script>

<template>
  <main class="page">
    <template v-if="post">
      <!-- 封面手风琴：挤压展开，点展开段飞出 Lightbox -->
      <div v-if="post.covers.length" v-reveal class="cover-band">
        <CoverAccordion :images="post.covers" autoplay="always" />
      </div>

      <header v-reveal class="article-head">
        <div class="meta">
          <span class="date">{{ post.createdAt.slice(0, 10) }}</span>
          <span v-for="tag in post.tags" :key="tag" class="mini-tag">{{ tag }}</span>
        </div>
      </header>
      <!-- Markdown 渲染区（标题由正文一级标题承担） -->
      <article v-reveal class="markdown" v-html="html" />
      <router-link v-reveal to="/articles" class="back">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round">
          <path d="M19 12H5M11 6l-6 6 6 6" />
        </svg>
        {{ t('article.back') }}
      </router-link>
    </template>

    <p v-else-if="notFound" class="empty">{{ t('article.notFound') }}</p>
  </main>
</template>

<style scoped lang="scss">
.page {
  max-width: 760px;
  margin: 0 auto;
  padding: 110px 24px 80px;
}

.cover-band {
  height: clamp(240px, 40vw, 440px);
  border-radius: var(--radius-lg);
  overflow: hidden;
  margin-bottom: 26px;
}

.article-head .meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 18px;

  .date { font-size: 13px; color: var(--text-2); }
}

.mini-tag {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 9px;
  border-radius: 999px;
  background: rgba(var(--primary-rgb), 0.1);
  color: var(--primary);
}

/* 文章排版：宋体标题 + 黑体正文 */
.markdown {
  line-height: 1.9;
  font-size: 16px;

  :deep(h1) {
    font-size: clamp(28px, 3.6vw, 38px);
    line-height: 1.35;
    margin-bottom: 28px;
  }

  :deep(h2) {
    font-size: 24px;
    margin: 40px 0 16px;
    padding-left: 14px;
    border-left: 4px solid var(--primary);
  }

  :deep(h3) { font-size: 19px; margin: 28px 0 12px; }
  :deep(p) { margin: 14px 0; }
  :deep(ul), :deep(ol) { padding-left: 26px; margin: 14px 0; }
  :deep(li) { margin: 6px 0; }

  :deep(blockquote) {
    margin: 18px 0;
    padding: 12px 18px;
    border-left: 4px solid rgba(var(--primary-rgb), 0.5);
    background: var(--surface);
    border-radius: 0 var(--radius) var(--radius) 0;
    color: var(--text-2);
  }

  :deep(code) {
    font-family: Consolas, 'Courier New', monospace;
    font-size: 0.9em;
    background: var(--surface-2);
    padding: 2px 6px;
    border-radius: 6px;
  }

  :deep(pre) {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 18px;
    overflow-x: auto;
    margin: 18px 0;

    code { background: none; padding: 0; }
  }

  :deep(table) {
    width: 100%;
    border-collapse: collapse;
    margin: 18px 0;

    th, td {
      border: 1px solid var(--border);
      padding: 9px 14px;
      text-align: left;
    }

    th { background: var(--surface-2); }
  }

  :deep(a) {
    color: var(--primary);
    border-bottom: 1px solid rgba(var(--primary-rgb), 0.35);
    transition: border-color var(--dur-fast);

    &:hover { border-bottom-color: var(--primary); }
  }

  :deep(img) {
    max-width: 100%;
    border-radius: var(--radius);
  }

  /* 拼图：同段落多张图并排成宫格 */
  :deep(p:has(img + img)) {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 6px;

    img {
      width: 100%;
      height: 100%;
      aspect-ratio: 4 / 3;
      object-fit: cover;
    }
  }

  /* 待办清单 */
  :deep(.task-list-item) {
    list-style: none;

    input { accent-color: var(--primary); margin-right: 8px; }
  }

  :deep(ul:has(.task-list-item)) { padding-left: 8px; }
}

.back {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin-top: 44px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-2);
  transition: color var(--dur-fast), transform var(--dur-fast);

  svg { width: 18px; height: 18px; }

  &:hover { color: var(--primary); transform: translateX(-4px); }
}

.empty {
  color: var(--text-2);
  text-align: center;
  padding: 48px 0;
}

@media (max-width: 768px) {
  .page { padding-top: 88px; }
}
</style>
