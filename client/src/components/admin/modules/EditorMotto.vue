<script setup lang="ts">
import type { AboutModule } from '../../../stores/config';

/** 格言块：单行文本 */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <input v-model="mod.data.text" class="a-input" type="text" />
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
