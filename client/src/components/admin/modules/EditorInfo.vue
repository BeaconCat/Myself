<script setup lang="ts">
import type { AboutModule } from '../../../stores/config';
import { metaOf } from '../../../about/registry';

/** 无需配置的模块（stats / github）：仅展示摘要提示 */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <p class="hint">{{ metaOf(mod.type)?.summary(mod.data) }}</p>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
