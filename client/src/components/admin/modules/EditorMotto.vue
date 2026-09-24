<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { MottoData } from '../../../about/types';
import { useModuleData } from './useModuleData';

/** 格言：一句话 + 署名小字 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<MottoData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <div class="grid3">
      <label class="span2"><span>{{ t('aboutKit.ed.motto') }}</span><input v-model="d.text" class="a-input" type="text" /></label>
      <label><span>{{ t('aboutKit.ed.sign') }}</span><input v-model="d.sign" class="a-input" type="text" placeholder="NAME · SINCE 2026" /></label>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
