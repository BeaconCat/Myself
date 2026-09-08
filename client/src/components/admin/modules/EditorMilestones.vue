<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';

/** 历程时间线：年份 + 描述 */
defineProps<{ mod: AboutModule }>();
const { t } = useI18n();
</script>

<template>
  <div v-for="(item, ii) in mod.data.items" :key="ii" class="line">
    <input v-model="item.year" class="a-input w-narrow" type="text" :placeholder="t('admin.msYear')" />
    <input v-model="item.text" class="a-input flex-in" type="text" :placeholder="t('admin.msText')" />
    <button class="op danger" @click="mod.data.items.splice(ii, 1)">{{ t('admin.delete') }}</button>
  </div>
  <button class="a-btn ghost sm" @click="mod.data.items.push({ year: '', text: '' })">{{ t('admin.addItem') }}</button>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
