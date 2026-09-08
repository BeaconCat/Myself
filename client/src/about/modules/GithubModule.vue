<script setup lang="ts">
import type { AboutModule } from '../../stores/config';
import GithubStatusCard from '../../components/home/GithubStatusCard.vue';

/** GitHub 状态卡 */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <h2 class="block-title">GitHub</h2>
  <GithubStatusCard />
</template>

<style scoped lang="scss">
@use './shared';
</style>
