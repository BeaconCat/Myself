<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { MilestonesData } from '../../../about/types';
import Switch from '../../ui/Switch.vue';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 历程：日期（2026.07）/ 标题 / 描述 / 点亮 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<MilestonesData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <p class="hint">{{ t('aboutKit.ed.milestonesHint') }}</p>
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ date: '', title: '', text: '' })">
      <div class="line">
        <input v-model="item.date" class="a-input w-narrow" type="text" placeholder="2026.07" />
        <input v-model="item.title" class="a-input flex-in" type="text" :placeholder="t('aboutKit.ed.title')" />
        <Switch v-model="item.now">{{ t('aboutKit.ed.lit') }}</Switch>
      </div>
      <input v-model="item.text" class="a-input" type="text" :placeholder="t('aboutKit.ed.note')" />
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
