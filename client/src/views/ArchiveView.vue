<script setup lang="ts">
// P1 阶段实现归档
</script>

<template>
  <main class="page">
    <h1 v-reveal>归档</h1>
    <p v-reveal class="hint">P1 阶段接入按时间归档。</p>
  </main>
</template>

<style scoped lang="scss">
.page {
  max-width: 1180px;
  margin: 0 auto;
  padding: 110px 24px 80px;
}

.hint { color: var(--text-2); margin-top: 12px; }

@media (max-width: 768px) {
  .page { padding-top: 88px; }
}
</style>
