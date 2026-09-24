<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { LanguagesData } from '../../../about/types';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 语言占比：口径说明 + 名称 / 百分比 / 颜色；合计不足 100 前台自动归一 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<LanguagesData>(() => props.mod);
const { t } = useI18n();
const sum = computed(() => d.value.items.reduce((a, b) => a + (Number(b.percent) || 0), 0));
</script>

<template>
  <div class="ed">
    <label><span>{{ t('aboutKit.ed.unit') }}</span><input v-model="d.unit" class="a-input" type="text" /></label>
    <div class="bar" aria-hidden="true">
      <i v-for="(it, i) in d.items" :key="i" :style="{ flex: Math.max(0, it.percent), background: it.color }" />
    </div>
    <p class="hint">{{ t('aboutKit.ed.sum', { n: sum }) }}</p>
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ name: '', percent: 10, color: '#8a96ab' })" compact>
      <div class="line">
        <input v-model="item.name" class="a-input flex-in" type="text" :placeholder="t('aboutKit.ed.itemName')" />
        <input v-model.number="item.percent" class="a-input w-num" type="number" min="0" max="100" />
        <input v-model="item.color" class="color-in" type="color" />
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.bar { display: flex; gap: 3px; height: 10px; border-radius: 5px; overflow: hidden; }
.bar i { display: block; min-width: 2px; }
</style>
