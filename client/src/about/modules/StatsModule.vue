<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { api } from '../../api';
import type { AboutModule } from '../../stores/config';
import { useConfigStore } from '../../stores/config';

/** 站点数字：运行天数 + 文章 / 随想 / 标签实时统计 */
defineProps<{ mod: AboutModule }>();

const config = useConfigStore();

const postCount = ref(0);
const noteCount = ref(0);
const tagCount = ref(0);

const daysRunning = computed(() => {
  const from = new Date(config.cfg.about.foundedAt || '2026-01-01').getTime();
  return Math.max(1, Math.floor((Date.now() - from) / 864e5));
});

const stats = computed(() => [
  { value: daysRunning.value, label: '运行天数' },
  { value: postCount.value, label: '文章' },
  { value: noteCount.value, label: '随想' },
  { value: tagCount.value, label: '标签' },
]);

onMounted(async () => {
  try {
    const [posts, notes, tags] = await Promise.all([
      api.posts({ pageSize: 1 }),
      api.notes({ pageSize: 1 }),
      api.tags(),
    ]);
    postCount.value = posts.total;
    noteCount.value = notes.total;
    tagCount.value = tags.length;
  } catch { /* 后端未启动统计留零 */ }
});
</script>

<template>
  <div class="stats">
    <div v-for="s in stats" :key="s.label" class="stat">
      <strong>{{ s.value }}</strong>
      <span>{{ s.label }}</span>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use './shared';

.stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.stat {
  text-align: center;
  padding: 20px 8px;
  background: var(--surface);
  border: 1px solid var(--border);
  transition: transform var(--dur-fast) var(--ease-out), border-color var(--dur-fast);

  strong {
    display: block;
    font-family: var(--font-serif);
    font-size: 26px;
    background: var(--grad-title);
    background-clip: text;
    -webkit-background-clip: text;
    color: transparent;
  }

  span { display: block; margin-top: 6px; font-size: 12px; color: var(--text-2); }
  &:hover { transform: scale(1.04); border-color: var(--primary); }
}

@media (max-width: 720px) {
  .stats { grid-template-columns: repeat(2, 1fr); }
}
</style>
