<script setup lang="ts">
import type { AboutModule } from '../../stores/config';

/** 格言块 */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <blockquote class="motto">{{ mod.data.text }}</blockquote>
</template>

<style scoped lang="scss">
.motto {
  margin: 0 auto;
  padding: 14px 30px;
  width: fit-content;
  max-width: 100%;
  font-family: var(--font-serif);
  font-size: 17px;
  color: var(--text-2);
  border-left: 3px solid rgba(var(--primary-rgb), 0.6);
  border-right: 3px solid rgba(var(--primary-rgb), 0.6);
  background: rgba(var(--primary-rgb), 0.05);
  border-radius: 8px;
}
</style>
