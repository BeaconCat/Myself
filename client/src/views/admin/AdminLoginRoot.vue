<script setup lang="ts">
/** 登录页根：按设备选择桌面登录或移动端登录。 */
import { defineAsyncComponent } from 'vue';
import { useDevice } from '../../composables/useDevice';
import { mobileAdminLogin } from '../../router/mobile-admin';

const AdminLoginView = defineAsyncComponent(() => import('./AdminLoginView.vue'));
const MobileLogin = defineAsyncComponent(mobileAdminLogin);
const { isMobile } = useDevice();
</script>

<template>
  <MobileLogin v-if="isMobile" />
  <AdminLoginView v-else />
</template>
