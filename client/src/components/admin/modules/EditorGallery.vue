<script setup lang="ts">
import type { AboutModule } from '../../../stores/config';
import CoverUploader from '../CoverUploader.vue';

/** 照片墙：复用封面上传器，最多 12 张 */
defineProps<{ mod: AboutModule }>();
</script>

<template>
  <CoverUploader v-model="mod.data.images" :max="12" />
</template>
