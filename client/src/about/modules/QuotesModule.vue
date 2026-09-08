<script setup lang="ts">
import type { AboutModule } from '../../stores/config';

/** 语录集 */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <h2 class="block-title">语录</h2>
  <div class="quotes">
    <figure v-for="(q, i) in mod.data.items" :key="i" class="quote card-box">
      <blockquote>{{ q.text }}</blockquote>
      <figcaption v-if="q.from">—— {{ q.from }}</figcaption>
    </figure>
  </div>
</template>

<style scoped lang="scss">
@use './shared';

.quotes {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 14px;
}

.quote {
  blockquote {
    font-family: var(--font-serif);
    font-size: 15.5px;
    line-height: 1.9;
  }

  figcaption {
    margin-top: 10px;
    font-size: 12.5px;
    color: var(--text-2);
    text-align: right;
  }
}
</style>
