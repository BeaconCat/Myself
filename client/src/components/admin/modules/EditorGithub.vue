<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { GithubData } from '../../../about/types';
import { useModuleData } from './useModuleData';

/** GitHub：数据来自「设置 · GitHub」；这里只配置是否显示最近动态与条数 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<GithubData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <p class="hint">{{ t('aboutKit.ed.githubHint') }}</p>
    <div class="line">
      <label class="check"><input v-model="d.showCommits" type="checkbox" />{{ t('aboutKit.ed.showCommits') }}</label>
      <label class="line"><span>{{ t('aboutKit.ed.commitCount') }}</span><input v-model.number="d.commitCount" class="a-input w-num" type="number" min="1" max="8" :disabled="!d.showCommits" /></label>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

label.line { flex-direction: row; align-items: center; margin-left: 12px; }
</style>
