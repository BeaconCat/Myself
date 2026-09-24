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
      <div class="line">
        <input v-model="item.text" class="a-input flex-in" type="text" :placeholder="t('aboutKit.ed.quote')" />
        <input v-model="item.from" class="a-input w-narrow" type="text" :placeholder="t('aboutKit.ed.from')" />
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

label.line { flex-direction: row; align-items: center; }
</style>
