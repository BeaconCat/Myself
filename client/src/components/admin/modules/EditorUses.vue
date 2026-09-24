<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { UsesData } from '../../../about/types';
import { ICONS, USES_ICONS } from '../../../about/icons';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 工作台：分组（硬件 / 软件 / 服务…）→ 条目 图标 / 名称 / 规格 / 标签 / 链接（旧 devices 自动迁移为「硬件」组） */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<UsesData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <EdList v-slot="{ item: g }" :items="d.groups" :make="() => ({ title: '', items: [] })" :add-label="t('aboutKit.ed.addGroup')">
      <input v-model="g.title" class="a-input grp" type="text" :placeholder="t('aboutKit.ed.groupTitle')" />
      <EdList v-slot="{ item }" :items="g.items" :make="() => ({ icon: 'laptop', name: '', desc: '', tag: '' })" compact>
        <div class="line">
          <!-- eslint-disable-next-line vue/no-v-html -->
          <svg class="ic" viewBox="0 0 24 24" aria-hidden="true" v-html="ICONS[item.icon || 'link'] ?? ICONS.link" />
          <select v-model="item.icon" class="a-input w-num">
            <option v-for="ic in USES_ICONS" :key="ic" :value="ic">{{ ic }}</option>
          </select>
          <input v-model="item.name" class="a-input flex-in" type="text" :placeholder="t('aboutKit.ed.itemName')" />
          <input v-model="item.desc" class="a-input flex-in" type="text" :placeholder="t('aboutKit.ed.spec')" />
          <input v-model="item.tag" class="a-input w-num" type="text" :placeholder="t('aboutKit.ed.tag')" />
        </div>
      </EdList>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.grp { font-weight: 600; }

.ic {
  width: 30px;
  height: 30px;
  flex-shrink: 0;
  padding: 6px;
  border-radius: 9px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.5;
  stroke-linecap: round;
  stroke-linejoin: round;
  color: var(--text-2);
  background: color-mix(in oklab, var(--text) 5%, transparent);
}
</style>
