<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { SiteConfig } from '../../../stores/config';

/** 平台：站点标题 / 副标题 / 列表尾文案（直接编辑父级 cfg 的嵌套字段） */
defineProps<{ cfg: SiteConfig }>();
const { t } = useI18n();
</script>

<template>
  <section id="sec-platform" class="card">
    <h2>{{ t('admin.secPlatform') }}</h2>
    <div class="row2">
      <label><span>{{ t('admin.siteTitle') }}</span><input v-model="cfg.site.title" type="text" /></label>
      <label><span>{{ t('admin.siteSubtitle') }}</span><input v-model="cfg.site.subtitle" type="text" /></label>
    </div>
    <label><span>{{ t('admin.listEndText') }}</span><input v-model="cfg.site.listEndText" type="text" /></label>
  </section>
</template>

<style scoped lang="scss">
@use './settings-shared';
</style>
