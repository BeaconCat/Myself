<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { Post } from '../../api';
import { ymdOf } from './content';

/** 文章元信息行：# 首个标签 · M月D日（· 可选附加项） */
defineProps<{ post: Post; extra?: string }>();
const { t } = useI18n();

function md(s: string): string {
  const { m, d } = ymdOf(s);
  return `${m}月${d}日`;
}
</script>

<template>
  <div class="meta">
    <span v-if="post.tags[0]" class="tag">{{ post.tags[0] }}</span>
    <span v-if="post.tags[0]" class="dotsep" />
    <time>{{ md(post.createdAt) }}</time>
    <template v-if="post.pinned">
      <span class="dotsep" />
      <span>{{ t('content.articles.pinned') }}</span>
    </template>
    <template v-if="extra">
      <span class="dotsep" />
      <span>{{ extra }}</span>
    </template>
  </div>
</template>

<style scoped lang="scss">
.meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0 10px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-3);
}

.dotsep {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: currentColor;
  opacity: 0.7;
}

/* 列表标签：「# 名称」纯文字，# 用三级灰 */
.tag {
  color: var(--text-2);

  &::before {
    content: '#';
    margin-right: 3px;
    font-family: var(--font-mono);
    font-size: 0.92em;
    color: var(--text-3);
  }
}
</style>
