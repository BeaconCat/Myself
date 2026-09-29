<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { SocialsData } from '../../../about/types';
import { SOCIAL_ICONS } from '../../../about/icons';
import { BRANDS } from '../../../about/brands';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 社交：名称 / handle / URL / 图标 / 主入口 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<SocialsData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ name: '', handle: '', url: '', icon: 'link' })">
      <div class="grid4">
        <input v-model="item.name" class="a-input" type="text" :placeholder="t('aboutKit.ed.linkName')" />
        <input v-model="item.handle" class="a-input" type="text" :placeholder="t('aboutKit.ed.handle')" />
        <input v-model="item.url" class="a-input" type="text" placeholder="https://…" />
        <div class="line">
          <select v-model="item.icon" class="a-input flex-in">
            <option v-for="ic in SOCIAL_ICONS" :key="ic" :value="ic">{{ BRANDS[ic]?.label ?? t(`studio.identity.ic_${ic}`) }}</option>
          </select>
          <label class="check"><input v-model="item.primary" type="checkbox" />{{ t('aboutKit.ed.primary') }}</label>
        </div>
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
