<script setup lang="ts">
import type { AboutModule } from '../../stores/config';

/** 喜好清单：分组标签卡（与技能分组同款） */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <h2 class="block-title">喜好</h2>
  <div class="skills">
    <div v-for="group in mod.data.groups" :key="group.title" class="skill-group">
      <h3>{{ group.title }}</h3>
      <div class="chips">
        <span v-for="item in group.items" :key="item">{{ item }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use './shared';
@use './groups';
</style>
