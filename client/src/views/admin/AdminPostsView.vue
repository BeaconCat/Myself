<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type AdminPost } from '../../api';

const { t } = useI18n();
const posts = ref<AdminPost[]>([]);

async function load(): Promise<void> {
  posts.value = await adminApi.posts();
}

async function remove(post: AdminPost): Promise<void> {
  if (!window.confirm(t('admin.confirmDeletePost', { title: post.title }))) return;
  await adminApi.deletePost(post.id);
  await load();
}

onMounted(load);
</script>

<template>
  <div>
    <header class="head">
      <h1>{{ t('admin.menuPosts') }}</h1>
      <router-link to="/admin/posts/new" class="btn-primary">{{ t('admin.newPost') }}</router-link>
    </header>

    <table class="table">
      <thead>
        <tr>
          <th>{{ t('admin.colTitle') }}</th>
          <th>{{ t('admin.colStatus') }}</th>
          <th>{{ t('admin.colTags') }}</th>
          <th>{{ t('admin.colDate') }}</th>
          <th />
        </tr>
      </thead>
      <tbody>
        <tr v-for="post in posts" :key="post.id">
          <td class="title-cell">{{ post.title }}</td>
          <td>
            <span class="status" :class="post.status">
              {{ post.status === 'draft' ? t('admin.draft') : t('admin.published') }}
            </span>
          </td>
          <td class="tags-cell">{{ post.tags.join(' / ') }}</td>
          <td class="date-cell">{{ post.createdAt.slice(0, 10) }}</td>
          <td class="ops">
            <router-link :to="`/admin/posts/${post.id}`" class="op">{{ t('admin.edit') }}</router-link>
            <button class="op danger" @click="remove(post)">{{ t('admin.delete') }}</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped lang="scss">
.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;

  h1 { font-size: 26px; }
}

.btn-primary {
  padding: 10px 22px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 700;
  color: #fff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
  background: linear-gradient(180deg, var(--primary), var(--primary-deep));
  box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.4);
  transition: transform var(--dur-fast) var(--ease-out), filter var(--dur-fast);

  &:hover { filter: brightness(1.08); transform: scale(1.04); }
}

.table {
  width: 100%;
  border-collapse: collapse;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;

  th, td {
    padding: 13px 16px;
    text-align: left;
    font-size: 14px;
    border-bottom: 1px solid var(--border);
  }

  th {
    background: var(--surface-2);
    font-size: 12px;
    color: var(--text-2);
  }

  tbody tr {
    transition: background var(--dur-fast);

    &:hover { background: rgba(var(--primary-rgb), 0.04); }
  }
}

.title-cell { font-weight: 600; }
.tags-cell, .date-cell { color: var(--text-2); font-size: 13px; }

.status {
  font-size: 11px;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 999px;

  &.published { background: rgba(var(--primary-rgb), 0.12); color: var(--primary); }
  &.draft { background: var(--surface-2); color: var(--text-2); }
}

.ops {
  white-space: nowrap;
  text-align: right;
}

.op {
  border: none;
  background: none;
  font-size: 13px;
  font-weight: 600;
  color: var(--primary);
  margin-left: 14px;
  transition: opacity var(--dur-fast);

  &:hover { opacity: 0.75; }
  &.danger { color: var(--accent-red); }
}
</style>
