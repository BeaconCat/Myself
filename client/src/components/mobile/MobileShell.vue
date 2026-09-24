<script setup lang="ts">
/** 移动端前台外壳（占位实现，待移动端前台重写）。 */
</script>

<template>
  <router-view v-slot="{ Component, route }" name="mobile">
    <component :is="Component" :key="route.path" />
  </router-view>
</template>
