<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';

/** 分组标签（skills / favorites）：分组标题 + 逗号分隔的条目 */
defineProps<{ mod: AboutModule }>();
const { t } = useI18n();

function parseList(text: string): string[] {
  return text.split(/[,，]+/).map((s) => s.trim()).filter(Boolean);
}
</script>

<template>
  <div v-for="(group, gi) in mod.data.groups" :key="gi" class="line">
    <input v-model="group.title" class="a-input w-narrow" type="text" :placeholder="t('admin.groupTitle')" />
    <input
      class="a-input flex-in"
      type="text"
      :value="group.items.join(', ')"
      :placeholder="t('admin.tagsPlaceholder')"
      @change="group.items = parseList(($event.target as HTMLInputElement).value)"
    />
    <button class="op danger" @click="mod.data.groups.splice(gi, 1)">{{ t('admin.delete') }}</button>
  </div>
  <button class="a-btn ghost sm" @click="mod.data.groups.push({ title: '', items: [] })">{{ t('admin.addItem') }}</button>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
