<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { ContactData } from '../../../about/types';
import { useModuleData } from './useModuleData';

/** 联系：标题 / 正文 / 邮箱（可复制）/ 按钮文案 / 链接 / 回复时效 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<ContactData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <div class="grid2">
      <label><span>{{ t('aboutKit.ed.title') }}</span><input v-model="d.title" class="a-input" type="text" /></label>
      <label><span>{{ t('aboutKit.ed.email') }}</span><input v-model="d.email" class="a-input" type="email" /></label>
    </div>
    <label><span>{{ t('aboutKit.ed.text') }}</span><input v-model="d.text" class="a-input" type="text" /></label>
    <div class="grid3">
      <label><span>{{ t('aboutKit.ed.buttonText') }}</span><input v-model="d.buttonText" class="a-input" type="text" /></label>
      <label><span>{{ t('aboutKit.ed.url') }}</span><input v-model="d.url" class="a-input" type="text" placeholder="mailto: / https://" /></label>
      <label><span>{{ t('aboutKit.ed.sla') }}</span><input v-model="d.sla" class="a-input" type="text" /></label>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
