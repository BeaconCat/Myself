<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';

/** 社交链接：名称 + URL + 图标类型 */
defineProps<{ mod: AboutModule }>();
const { t } = useI18n();
</script>

<template>
  <div v-for="(item, ii) in mod.data.items" :key="ii" class="line">
    <input v-model="item.name" class="a-input w-narrow" type="text" :placeholder="t('admin.socialName')" />
    <input v-model="item.url" class="a-input flex-in" type="text" placeholder="https://…" />
    <select v-model="item.icon" class="a-input w-narrow">
      <option value="github">GitHub</option>
      <option value="mail">Mail</option>
      <option value="rss">RSS</option>
      <option value="link">Link</option>
    </select>
    <button class="op danger" @click="mod.data.items.splice(ii, 1)">{{ t('admin.delete') }}</button>
  </div>
  <button class="a-btn ghost sm" @click="mod.data.items.push({ name: '', url: '', icon: 'link' })">{{ t('admin.addItem') }}</button>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
