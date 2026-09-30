<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { Place, PlacesData } from '../../../about/types';
import { CITY_COORDS } from '../../../about/geo';
import Combobox from '../../ui/Combobox.vue';
import Switch from '../../ui/Switch.vue';
import type { UiOption } from '../../ui/listbox';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 足迹：城市（组合框模糊搜索常用城市，选中 / 输入命中即填经纬度，也可自由输入）/ 经度 / 纬度 / 年份 / 常住地（唯一） */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<PlacesData>(() => props.mod);
const { t } = useI18n();

const CITY_OPTIONS: UiOption[] = Object.entries(CITY_COORDS).map(([name, [lon, lat]]) => ({
  value: name,
  label: name,
  hint: `${lon.toFixed(1)}, ${lat.toFixed(1)}`,
}));

function onName(p: Place): void {
  const c = CITY_COORDS[p.name.trim()];
  if (c) [p.lon, p.lat] = c;
}

function setHome(p: Place, on: boolean): void {
  for (const x of d.value.items) x.home = on && x === p;
}
</script>

<template>
  <div class="ed">
    <p class="hint">{{ t('aboutKit.ed.placesHint') }}</p>
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ name: '', lon: 116.4, lat: 39.9, year: '' })" compact>
      <div class="p-row">
        <Combobox
          v-model="item.name"
          :options="CITY_OPTIONS"
          :placeholder="t('aboutKit.ed.city')"
          :aria-label="t('aboutKit.ed.city')"
          :empty="t('aboutKit.ed.cityFree')"
          :min-width="240"
          @change="onName(item)"
        />
        <label class="affix">
          <span>{{ t('aboutKit.ed.lon') }}</span>
          <input v-model.number="item.lon" type="number" step="0.1" min="-180" max="180" />
        </label>
        <label class="affix">
          <span>{{ t('aboutKit.ed.lat') }}</span>
          <input v-model.number="item.lat" type="number" step="0.1" min="-90" max="90" />
        </label>
        <input v-model="item.year" class="a-input" type="text" :placeholder="t('aboutKit.ed.year')" :aria-label="t('aboutKit.ed.year')" />
        <Switch :model-value="!!item.home" @update:model-value="setHome(item, $event)">{{ t('aboutKit.ed.home') }}</Switch>
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.p-row {
  display: grid;
  grid-template-columns: minmax(160px, 1fr) 128px 128px minmax(90px, 0.7fr) auto;
  align-items: center;
  gap: 8px;
}

@container ed (max-width: 640px) {
  .p-row { grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); }
  .p-row :deep(.ui-cb) { grid-column: 1 / -1; }
}
</style>
