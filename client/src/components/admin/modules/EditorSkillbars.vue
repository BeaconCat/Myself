<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';

/** 技能图表：名称 + 0–100 熟练度（数字框与滑杆联动） */
defineProps<{ mod: AboutModule }>();
const { t } = useI18n();
</script>

<template>
  <div v-for="(item, ii) in mod.data.items" :key="ii" class="line">
    <input v-model="item.name" class="a-input w-narrow" type="text" :placeholder="t('admin.fieldName')" />
    <input v-model.number="item.level" class="a-input w-num" type="number" min="0" max="100" />
    <input v-model.number="item.level" class="range" type="range" min="0" max="100" />
    <button class="op danger" @click="mod.data.items.splice(ii, 1)">{{ t('admin.delete') }}</button>
  </div>
  <button class="a-btn ghost sm" @click="mod.data.items.push({ name: '', level: 50 })">{{ t('admin.addItem') }}</button>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
