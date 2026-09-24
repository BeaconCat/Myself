<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { Place, PlacesData } from '../../../about/types';
import { CITY_COORDS } from '../../../about/geo';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 足迹：城市名（常用城市自动填经纬度）/ 经度 / 纬度 / 年份 / 常住地（唯一） */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<PlacesData>(() => props.mod);
const { t } = useI18n();
const CITIES = Object.keys(CITY_COORDS);

function onName(p: Place): void {
  const c = CITY_COORDS[p.name.trim()];
  if (c) [p.lon, p.lat] = c;
}

function setHome(p: Place): void {
  for (const x of d.value.items) x.home = x === p;
}
</script>

<template>
  <div class="ed">
    <p class="hint">{{ t('aboutKit.ed.placesHint') }}</p>
    <datalist id="ak-cities"><option v-for="c in CITIES" :key="c" :value="c" /></datalist>
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ name: '', lon: 116.4, lat: 39.9, year: '' })" compact>
      <div class="line">
        <input v-model="item.name" class="a-input w-narrow" type="text" list="ak-cities" :placeholder="t('aboutKit.ed.city')" @change="onName(item)" />
        <input v-model.number="item.lon" class="a-input w-num" type="number" step="0.1" :title="t('aboutKit.ed.lon')" />
        <input v-model.number="item.lat" class="a-input w-num" type="number" step="0.1" :title="t('aboutKit.ed.lat')" />
        <input v-model="item.year" class="a-input flex-in" type="text" :placeholder="t('aboutKit.ed.year')" />
        <label class="check"><input type="radio" :checked="!!item.home" @change="setHome(item)" />{{ t('aboutKit.ed.home') }}</label>
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
