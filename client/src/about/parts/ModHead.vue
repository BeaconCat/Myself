<script setup lang="ts">
/** 统一模块头：左宋体标题，右 mono 元信息（slot） */
defineProps<{ title: string }>();
</script>

<template>
  <header class="ak-mh">
    <h3 class="ak-mt">{{ title }}</h3>
    <span v-if="$slots.default" class="ak-mm"><slot /></span>
  </header>
</template>
