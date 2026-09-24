<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { ChapterData } from '../../../about/types';
import { useModuleData } from './useModuleData';

/** 章节：编号（留空自动）/ 标题 / 副题 / 右侧小字 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<ChapterData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <div class="grid4">
      <label><span>{{ t('aboutKit.ed.chapterNo') }}</span><input v-model="d.no" class="a-input" type="text" :placeholder="t('aboutKit.ed.auto')" /></label>
      <label class="span2"><span>{{ t('aboutKit.ed.title') }}</span><input v-model="d.title" class="a-input" type="text" /></label>
      <label><span>{{ t('aboutKit.ed.meta') }}</span><input v-model="d.meta" class="a-input" type="text" /></label>
    </div>
    <label><span>{{ t('aboutKit.ed.subtitle') }}</span><input v-model="d.subtitle" class="a-input" type="text" /></label>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
