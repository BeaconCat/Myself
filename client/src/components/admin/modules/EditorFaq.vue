<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { FaqData } from '../../../about/types';
import Switch from '../../ui/Switch.vue';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 问答：问题 + 回答（支持 `行内代码`）；是否同时只展开一条 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<FaqData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <Switch v-model="d.single">{{ t('aboutKit.ed.single') }}</Switch>
    <EdList v-slot="{ item }" :items="d.items" :make="() => ({ q: '', a: '' })">
      <input v-model="item.q" class="a-input q" type="text" :placeholder="t('aboutKit.ed.question')" />
      <textarea v-model="item.a" class="a-input" rows="2" :placeholder="t('aboutKit.ed.answer')" />
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.q { font-weight: 600; }
</style>
