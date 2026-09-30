<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { SocialsData } from '../../../about/types';
import { SOCIAL_ICONS } from '../../../about/icons';
import { BRANDS, brandColor } from '../../../about/brands';
import KitIcon from '../../../about/parts/KitIcon.vue';
import Select from '../../ui/Select.vue';
import Switch from '../../ui/Switch.vue';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 社交：名称 / handle / URL / 图标 / 主入口 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<SocialsData>(() => props.mod);
const { t } = useI18n();
const iconOptions = computed(() => SOCIAL_ICONS.map((ic) => ({ value: ic, label: BRANDS[ic]?.label ?? t(`studio.identity.ic_${ic}`), keywords: ic })));
</script>

<template>
  <div class="ed">
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ name: '', handle: '', url: '', icon: 'link' })">
      <div class="grid4">
        <input v-model="item.name" class="a-input" type="text" :placeholder="t('aboutKit.ed.linkName')" />
        <input v-model="item.handle" class="a-input" type="text" :placeholder="t('aboutKit.ed.handle')" />
        <input v-model="item.url" class="a-input" type="text" placeholder="https://…" />
        <div class="line">
          <Select
            v-model="item.icon"
            class="flex-in"
            :options="iconOptions"
            :min-width="200"
            searchable
            :search-placeholder="t('aboutKit.ed.searchIcon')"
            :empty="t('aboutKit.ed.noMatch')"
            :aria-label="t('aboutKit.ed.techIcon')"
          >
            <template #icon="{ option }"><KitIcon :name="option.value" :size="16" :style="{ color: brandColor(option.value) }" /></template>
          </Select>
          <Switch v-model="item.primary">{{ t('aboutKit.ed.primary') }}</Switch>
        </div>
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';
</style>
