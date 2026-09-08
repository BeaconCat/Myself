<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { SiteConfig } from '../../../stores/config';

/** Loading 文案：启动 / 路由切换 */
defineProps<{ cfg: SiteConfig }>();
const { t } = useI18n();
</script>

<template>
  <section id="sec-loading" class="card">
    <h2>{{ t('admin.secLoading') }}</h2>
    <div class="row2">
      <label><span>{{ t('admin.bootText') }}</span><input v-model="cfg.loading.bootText" type="text" /></label>
      <label><span>{{ t('admin.routeText') }}</span><input v-model="cfg.loading.routeText" type="text" /></label>
    </div>
  </section>
</template>

<style scoped lang="scss">
@use './settings-shared';
</style>
