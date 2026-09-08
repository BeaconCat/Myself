<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';

/** 问答 FAQ：问题 + 多行回答 */
defineProps<{ mod: AboutModule }>();
const { t } = useI18n();
</script>

<template>
  <div v-for="(item, ii) in mod.data.items" :key="ii" class="faq-line">
    <input v-model="item.q" class="a-input" type="text" :placeholder="t('admin.faqQ')" />
    <textarea v-model="item.a" class="a-input" rows="2" :placeholder="t('admin.faqA')" />
    <button class="op danger" @click="mod.data.items.splice(ii, 1)">{{ t('admin.delete') }}</button>
  </div>
  <button class="a-btn ghost sm" @click="mod.data.items.push({ q: '', a: '' })">{{ t('admin.addItem') }}</button>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
