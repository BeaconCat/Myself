<script setup lang="ts">
import type { AboutModule } from '../../stores/config';

/** 社交链接：GitHub / 邮箱 / RSS / 任意链接 */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <h2 class="block-title">找到我</h2>
  <div class="links">
    <a
      v-for="l in mod.data.items"
      :key="l.name"
      :href="l.url"
      class="link"
      target="_blank"
      rel="noopener"
    >
      <svg v-if="l.icon === 'github'" viewBox="0 0 16 16" fill="currentColor">
        <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8Z" />
      </svg>
      <svg v-else-if="l.icon === 'mail'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="3" y="5" width="18" height="14" rx="2" />
        <path d="m3 7 9 6 9-6" />
      </svg>
      <svg v-else-if="l.icon === 'rss'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round">
        <path d="M4 11a9 9 0 0 1 9 9M4 4a16 16 0 0 1 16 16" />
        <circle cx="5" cy="19" r="1.6" fill="currentColor" stroke="none" />
      </svg>
      <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M10 14a5 5 0 0 0 7.07 0l2.83-2.83a5 5 0 0 0-7.07-7.07L11.4 5.5" />
        <path d="M14 10a5 5 0 0 0-7.07 0L4.1 12.83a5 5 0 0 0 7.07 7.07l1.42-1.4" />
      </svg>
      <span>{{ l.name }}</span>
    </a>
  </div>
</template>

<style scoped lang="scss">
@use './shared';

.links { display: flex; gap: 14px; flex-wrap: wrap; }

.link {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 12px 24px;
  background: var(--surface);
  border: 1px solid var(--border);
  font-size: 14px;
  font-weight: 600;
  transition: all var(--dur-fast) var(--ease-out);

  svg { width: 18px; height: 18px; }

  &:hover {
    transform: scale(1.05);
    border-color: var(--primary);
    color: var(--primary);
    box-shadow: 0 6px 20px -6px rgba(var(--primary-rgb), 0.4);
  }
}
</style>
