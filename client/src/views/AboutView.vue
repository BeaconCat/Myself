<script setup lang="ts">
// 关于页，后续填充
</script>

<template>
  <main class="page">
    <h1 v-reveal>关于</h1>
    <p v-reveal class="hint">个人介绍占位。</p>
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
