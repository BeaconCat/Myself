<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';

/** 卡片清单（devices / stack）：名称 + 描述（devices 存 desc，stack 存 role） */
defineProps<{ mod: AboutModule }>();
const { t } = useI18n();
</script>

<template>
  <div v-for="(item, ii) in mod.data.items" :key="ii" class="line">
    <input v-model="item.name" class="a-input w-narrow" type="text" :placeholder="t('admin.fieldName')" />
    <input
      v-if="mod.type === 'devices'"
      v-model="item.desc"
      class="a-input flex-in"
      type="text"
      :placeholder="t('admin.fieldDesc')"
    />
    <input
      v-else
      v-model="item.role"
      class="a-input flex-in"
      type="text"
      :placeholder="t('admin.fieldDesc')"
    />
    <button class="op danger" @click="mod.data.items.splice(ii, 1)">{{ t('admin.delete') }}</button>
  </div>
  <button
    class="a-btn ghost sm"
    @click="mod.data.items.push(mod.type === 'devices' ? { name: '', desc: '' } : { name: '', role: '' })"
  >{{ t('admin.addItem') }}</button>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
