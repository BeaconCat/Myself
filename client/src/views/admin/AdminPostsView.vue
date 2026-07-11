<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type AdminPost } from '../../api';
import { useDialogStore } from '../../stores/dialog';

const { t } = useI18n();
const posts = ref<AdminPost[]>([]);

async function load(): Promise<void> {
  posts.value = await adminApi.posts();
}

async function remove(post: AdminPost): Promise<void> {
  const ok = await useDialogStore().confirm({
    title: t('admin.delete'),
    message: t('admin.confirmDeletePost', { title: post.title }),
    danger: true,
  });
  if (!ok) return;
  await adminApi.deletePost(post.id);
  await load();
}

onMounted(load);
</script>

<template>
  <div>
    <header class="a-head">
      <div>
        <h1>{{ t('admin.menuPosts') }}</h1>
        <p>{{ t('admin.postsHint') }}</p>
      </div>
      <router-link to="/write/post" class="a-btn primary">{{ t('admin.newPost') }}</router-link>
    </header>

    <!-- 卡片网格（同随想管理风格），有封面则封面置顶 -->
    <div class="post-grid">
      <article v-for="post in posts" :key="post.id" class="post-card a-card">
        <router-link :to="{ path: '/write/post', query: { id: String(post.id) } }" class="cover-wrap">
          <img v-if="post.covers[0]" class="cover" :src="post.covers[0]" loading="lazy" alt="" />
          <div v-else class="cover placeholder">
            <span>{{ post.title.slice(0, 1) }}</span>
          </div>
          <span class="status" :class="post.status">
            {{ post.status === 'draft' ? t('admin.draft') : t('admin.published') }}
          </span>
          <span v-if="post.pinned" class="pin-badge">{{ t('admin.pinned') }}</span>
        </router-link>

        <div class="body">
          <h3>{{ post.title }}</h3>
          <p class="excerpt">{{ post.excerpt }}</p>
          <div class="tags">
            <span v-for="tag in post.tags" :key="tag">{{ tag }}</span>
          </div>
          <footer class="foot">
            <time>{{ post.createdAt.slice(0, 10) }}</time>
            <span class="spacer" />
            <router-link class="op" :to="{ path: '/write/post', query: { id: String(post.id) } }">{{ t('admin.edit') }}</router-link>
            <button class="op danger" @click="remove(post)">{{ t('admin.delete') }}</button>
          </footer>
        </div>
      </article>
    </div>

    <p v-if="!posts.length" class="empty">{{ t('admin.postsEmpty') }}</p>
  </div>
</template>

<style scoped lang="scss">
.post-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 14px;
}

.post-card {
  padding: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur), border-color var(--dur-fast);

  &:hover {
    transform: scale(1.015);
    box-shadow: 0 10px 30px -12px rgba(var(--primary-rgb), 0.25);
    border-color: rgba(var(--primary-rgb), 0.35);
  }
}

.cover-wrap {
  position: relative;
  display: block;
}

.cover {
  width: 100%;
  aspect-ratio: 21 / 9;
  object-fit: cover;
  display: block;
  background: var(--surface-2);

  &.placeholder {
    display: grid;
    place-items: center;
    background:
      radial-gradient(300px 120px at 30% 0%, rgba(var(--primary-rgb), 0.16), transparent 70%),
      var(--surface-2);

    span {
      font-family: var(--font-serif);
      font-size: 34px;
      font-weight: 700;
      background: var(--grad-title);
      background-clip: text;
      -webkit-background-clip: text;
      color: transparent;
    }
  }
}

.pin-badge {
  position: absolute;
  top: 10px;
  left: 10px;
  font-size: 11px;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 999px;
  backdrop-filter: blur(8px);
  background: rgba(255, 179, 0, 0.28);
  color: #fff;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.4);
}

.status {
  position: absolute;
  top: 10px;
  right: 10px;
  font-size: 11px;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 999px;
  backdrop-filter: blur(8px);

  &.published { background: rgba(var(--primary-rgb), 0.2); color: #fff; text-shadow: 0 1px 2px rgba(0,0,0,.4); }
  &.draft { background: rgba(0, 0, 0, 0.4); color: #fff; }
}

.body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 14px 16px 12px;
  flex: 1;

  h3 {
    font-size: 16.5px;
    line-height: 1.5;
  }
}

.excerpt {
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-2);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;

  span {
    font-size: 11px;
    font-weight: 600;
    padding: 2px 9px;
    border-radius: 999px;
    background: rgba(var(--primary-rgb), 0.1);
    color: var(--primary);
  }
}

.foot {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: auto;
  padding-top: 8px;
  font-size: 12px;
  color: var(--text-2);

  .spacer { flex: 1; }
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

.empty {
  color: var(--text-2);
  text-align: center;
  padding: 60px 0;
}
</style>
