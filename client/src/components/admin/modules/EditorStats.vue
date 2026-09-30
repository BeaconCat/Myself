<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { StatKey, StatsData } from '../../../about/types';
import Select from '../../ui/Select.vue';
import Switch from '../../ui/Switch.vue';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 站点数字：数值实时统计；这里只配置显示哪些项、文案覆盖、小字注释与年度进度条 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<StatsData>(() => props.mod);
const { t } = useI18n();
const KEYS: StatKey[] = ['days', 'posts', 'notes', 'tags'];
</script>

<template>
  <div class="ed">
    <p class="hint">{{ t('aboutKit.ed.statsHint') }}</p>
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ key: 'posts' as StatKey, label: '', hint: '' })" :max="6" compact>
      <div class="grid3">
        <Select v-model="item.key" :options="KEYS.map((k) => ({ value: k, label: t(`aboutKit.stats.${k}`) }))" :aria-label="t('aboutKit.ed.itemName')" />
        <input v-model="item.label" class="a-input" type="text" :placeholder="t(`aboutKit.stats.${item.key}`)" />
        <input v-model="item.hint" class="a-input" type="text" :placeholder="t('aboutKit.ed.hint')" />
      </div>
    </EdList>
    <Switch v-model="d.showYearProgress">{{ t('aboutKit.ed.yearProgress') }}</Switch>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
