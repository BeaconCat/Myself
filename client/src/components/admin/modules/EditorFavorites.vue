<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { FavoritesData } from '../../../about/types';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 喜好：分组（Tab）→ 条目 名称 / 作者 / 年份 / 一句评语 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<FavoritesData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <EdList v-slot="{ item: group }" :items="d.groups" :make="() => ({ title: '', items: [] })" :add-label="t('aboutKit.ed.addGroup')">
      <input v-model="group.title" class="a-input grp" type="text" :placeholder="t('aboutKit.ed.groupTitle')" />
      <EdList v-slot="{ item }" :items="group.items" :make="() => ({ name: '', by: '', year: '', note: '' })" compact>
        <div class="grid4">
          <input v-model="item.name" class="a-input" type="text" :placeholder="t('aboutKit.ed.itemName')" />
          <input v-model="item.by" class="a-input" type="text" :placeholder="t('aboutKit.ed.author')" />
          <input v-model="item.year" class="a-input" type="text" :placeholder="t('aboutKit.ed.year')" />
          <input v-model="item.note" class="a-input" type="text" :placeholder="t('aboutKit.ed.note')" />
        </div>
      </EdList>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.grp { font-weight: 600; }
</style>
