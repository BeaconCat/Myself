<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { StackData } from '../../../about/types';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 技术栈：名称 / 角色 / 2 字母徽标 / 品牌色 / 链接 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<StackData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ name: '', role: '', glyph: '', color: '#0078ff', url: '' })" compact>
      <div class="line">
        <span class="g" :style="{ '--c': item.color || 'var(--primary)' }">{{ item.glyph || item.name.slice(0, 2) }}</span>
        <input v-model="item.name" class="a-input w-narrow" type="text" :placeholder="t('aboutKit.ed.itemName')" />
        <input v-model="item.role" class="a-input flex-in" type="text" :placeholder="t('aboutKit.ed.role')" />
        <input v-model="item.glyph" class="a-input w-num" type="text" maxlength="3" :placeholder="t('aboutKit.ed.glyph')" />
        <input v-model="item.color" class="color-in" type="color" />
        <input v-model="item.url" class="a-input w-narrow" type="text" placeholder="https://…" />
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.g {
  display: grid;
  place-items: center;
  flex-shrink: 0;
  width: 34px;
  height: 34px;
  border-radius: 10px;
  font: 600 12px ui-monospace, Consolas, monospace;
  color: var(--c);
  background: color-mix(in oklab, var(--c) 13%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--c) 30%, transparent);
}
</style>
