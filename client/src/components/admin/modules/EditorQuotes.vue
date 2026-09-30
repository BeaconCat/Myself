<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { QuotesData } from '../../../about/types';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 语录：轮播间隔（秒）+ 语句 / 出处 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<QuotesData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <label class="line"><span>{{ t('aboutKit.ed.interval') }}</span><input v-model.number="d.interval" class="a-input w-num" type="number" min="2" max="30" /></label>
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ text: '', from: '' })" compact>
      <div class="q-row">
        <input v-model="item.text" class="a-input" type="text" :placeholder="t('aboutKit.ed.quote')" :aria-label="t('aboutKit.ed.quote')" />
        <input v-model="item.from" class="a-input" type="text" :placeholder="t('aboutKit.ed.from')" :aria-label="t('aboutKit.ed.from')" />
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

label.line { flex-direction: row; align-items: center; }

/* 语句 : 出处 = 2 : 1；编辑区很窄时上下排 */
.q-row { display: grid; grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); gap: 8px; }

@container ed (max-width: 520px) { .q-row { grid-template-columns: minmax(0, 1fr); } }
</style>
