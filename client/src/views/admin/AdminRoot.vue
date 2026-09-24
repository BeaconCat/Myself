<script setup lang="ts">
/** 后台根：按设备形态选择桌面工作室布局或移动端后台布局。 */
import { defineAsyncComponent } from 'vue';
import { useDevice } from '../../composables/useDevice';

const AdminLayout = defineAsyncComponent(() => import('./AdminLayout.vue'));
const MobileAdminLayout = defineAsyncComponent(() => import('../mobile/admin/MobileAdminLayout.vue'));
const { isMobile } = useDevice();
</script>

<template>
  <MobileAdminLayout v-if="isMobile" />
  <AdminLayout v-else />
</template>
