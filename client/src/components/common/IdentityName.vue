<script setup lang="ts">
/**
 * 名字 + 别名：别名不加括号，跟在名字后面，字号小一截、颜色比名字浅（最小 12px，小字号位置也看得清）。
 * 名字的字号 / 字重由外层决定。
 */
defineProps<{ name: string; alias?: string }>();
</script>

<template>
  <span class="idn">{{ name }}<small v-if="alias">{{ alias }}</small></span>
</template>

<style scoped>
.idn small {
  margin-left: 0.4em;
  font-family: var(--font-sans);
  font-size: max(12px, 0.62em);
  font-weight: 500;
  letter-spacing: 0;
  color: var(--text-3);
}
</style>
