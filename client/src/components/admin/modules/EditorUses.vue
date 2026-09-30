<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { UsesData } from '../../../about/types';
import { USES_ICONS } from '../../../about/icons';
import KitIcon from '../../../about/parts/KitIcon.vue';
import Select from '../../ui/Select.vue';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 工作台：分组（硬件 / 软件 / 服务…）→ 条目 图标 / 名称 / 规格 / 标签（旧 devices 自动迁移为「硬件」组） */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<UsesData>(() => props.mod);
const { t } = useI18n();

/** 图标下拉：图标 + 中文名；数据里若是列表外的旧图标名，也补一项以免显示为空 */
const baseOptions = computed(() => USES_ICONS.map((ic) => ({ value: ic, label: t(`aboutKit.ed.usesIcon.${ic}`), keywords: ic })));
function optionsFor(icon: string | undefined) {
  if (!icon || (USES_ICONS as readonly string[]).includes(icon)) return baseOptions.value;
  return [...baseOptions.value, { value: icon, label: icon }];
}
</script>

<template>
  <div class="ed">
    <EdList v-slot="{ item: g }" :items="d.groups" :make="() => ({ title: '', items: [] })" :add-label="t('aboutKit.ed.addGroup')">
      <input v-model="g.title" class="a-input grp" type="text" :placeholder="t('aboutKit.ed.groupTitle')" :aria-label="t('aboutKit.ed.groupTitle')" />
      <EdList v-slot="{ item }" :items="g.items" :make="() => ({ icon: 'laptop', name: '', desc: '', tag: '' })" compact>
        <div class="u-row">
          <Select
            v-model="item.icon"
            :options="optionsFor(item.icon)"
            :min-width="200"
            searchable
            :search-placeholder="t('aboutKit.ed.searchIcon')"
            :empty="t('aboutKit.ed.noMatch')"
            :aria-label="t('aboutKit.ed.techIcon')"
          >
            <template #icon="{ option }"><KitIcon :name="option.value" :size="17" /></template>
          </Select>
          <input v-model="item.name" class="a-input nm" type="text" :placeholder="t('aboutKit.ed.itemName')" :aria-label="t('aboutKit.ed.itemName')" />
          <input v-model="item.desc" class="a-input" type="text" :placeholder="t('aboutKit.ed.spec')" :aria-label="t('aboutKit.ed.spec')" />
          <input v-model="item.tag" class="a-input" type="text" :placeholder="t('aboutKit.ed.tag')" :aria-label="t('aboutKit.ed.tag')" />
        </div>
      </EdList>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.grp { font-weight: 600; }
.nm { font-weight: 500; }

.u-row {
  display: grid;
  grid-template-columns: minmax(150px, 170px) minmax(0, 1fr) minmax(0, 1.2fr) minmax(80px, 110px);
  align-items: center;
  gap: 8px;
}

@container ed (max-width: 640px) {
  .u-row { grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); }
}
</style>
