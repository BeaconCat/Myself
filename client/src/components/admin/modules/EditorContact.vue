<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';

/** 联系我 CTA：标题 / 按钮文案 / 正文 / URL */
defineProps<{ mod: AboutModule }>();
const { t } = useI18n();
</script>

<template>
  <div class="row2">
    <label><span>{{ t('admin.contactTitle') }}</span><input v-model="mod.data.title" class="a-input" type="text" /></label>
    <label><span>{{ t('admin.contactBtn') }}</span><input v-model="mod.data.buttonText" class="a-input" type="text" /></label>
  </div>
  <label><span>{{ t('admin.contactText') }}</span><input v-model="mod.data.text" class="a-input" type="text" /></label>
  <label><span>URL</span><input v-model="mod.data.url" class="a-input" type="text" /></label>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
