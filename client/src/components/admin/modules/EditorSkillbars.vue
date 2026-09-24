<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { SkillbarsData } from '../../../about/types';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 技能刻度：名称 + 0–100（数字框与滑杆联动）+ 段位（留空按数值自动） */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<SkillbarsData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ name: '', level: 50, tier: '' })" compact>
      <div class="line">
        <input v-model="item.name" class="a-input w-narrow" type="text" :placeholder="t('aboutKit.ed.itemName')" />
        <input v-model.number="item.level" class="range" type="range" min="0" max="100" />
        <input v-model.number="item.level" class="a-input w-num" type="number" min="0" max="100" />
        <input v-model="item.tier" class="a-input w-num" type="text" :placeholder="t('aboutKit.ed.auto')" />
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
