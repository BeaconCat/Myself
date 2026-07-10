<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { api, type ArchiveGroup } from '../api';

const { t } = useI18n();
const groups = ref<ArchiveGroup[]>([]);

function formatMonth(m: string): string {
  const [year, month] = m.split('-');
  return t('archive.month', { year, month: Number(month) });
}

onMounted(async () => {
  groups.value = await api.archive();
});
</script>

<template>
  <main class="page">
    <h1 v-reveal class="page-title">{{ t('nav.archive') }}</h1>

    <!-- 时间线 -->
    <div class="timeline">
      <section v-for="group in groups" :key="group.month" v-reveal class="month">
        <h2>{{ formatMonth(group.month) }}<i>{{ group.items.length }}</i></h2>
        <router-link
          v-for="post in group.items"
          :key="post.slug"
          :to="`/articles/${post.slug}`"
          class="entry"
        >
          <span class="dot" />
          <span class="date">{{ post.createdAt.slice(5, 10) }}</span>
          <span class="title">{{ post.title }}</span>
        </router-link>
      </section>
    </div>
  </main>
</template>

<style scoped lang="scss">
.page {
  max-width: 760px;
  margin: 0 auto;
  padding: 110px 24px 80px;
}

.page-title {
  font-size: clamp(30px, 4vw, 42px);
  margin-bottom: 34px;
}

.timeline {
  position: relative;
  padding-left: 22px;

  /* 竖线 */
  &::before {
    content: '';
    position: absolute;
    left: 5px;
    top: 6px;
    bottom: 6px;
    width: 2px;
    background: linear-gradient(180deg, var(--primary), var(--border));
    border-radius: 2px;
  }
}

.month {
  margin-bottom: 36px;

  h2 {
    font-size: 21px;
    margin-bottom: 14px;

    i {
      font-style: normal;
      font-size: 12px;
      font-weight: 600;
      margin-left: 10px;
      padding: 2px 10px;
      border-radius: 999px;
      background: rgba(var(--primary-rgb), 0.1);
      color: var(--primary);
      vertical-align: 3px;
    }
  }
}

.entry {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 10px 14px;
  margin-left: -22px;
  border-radius: var(--radius);
  transition: background var(--dur-fast), transform var(--dur-fast);

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--border);
    flex-shrink: 0;
    margin-left: 2px;
    transition: all var(--dur-fast);
  }

  .date {
    font-size: 13px;
    color: var(--text-2);
    font-variant-numeric: tabular-nums;
    flex-shrink: 0;
  }

  .title {
    font-size: 15px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    transition: color var(--dur-fast);
  }

  &:hover {
    background: var(--surface);
    transform: translateX(4px);

    .dot {
      background: var(--primary);
      box-shadow: 0 0 8px rgba(var(--primary-rgb), 0.6);
    }

    .title { color: var(--primary); }
  }
}

@media (max-width: 768px) {
  .page { padding-top: 88px; }
}
</style>
