<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { PrinciplesData } from '../../../about/types';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 信条：一句信条 + 一行解释（建议 3–4 条，编号自动） */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<PrinciplesData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <EdList v-slot="{ item, index }" :items="d.items" :make="() => ({ title: '', text: '' })" :max="8">
      <div class="line">
        <span class="no">{{ String(index + 1).padStart(2, '0') }}</span>
        <input v-model="item.title" class="a-input flex-in" type="text" :placeholder="t('aboutKit.ed.principle')" />
      </div>
      <input v-model="item.text" class="a-input" type="text" :placeholder="t('aboutKit.ed.note')" />
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.no { width: 28px; font: 700 18px var(--font-serif); color: var(--text-3); }
</style>
