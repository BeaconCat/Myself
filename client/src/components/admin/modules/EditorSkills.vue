<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { SkillsData } from '../../../about/types';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 技能：分组标题 + 逗号分隔的标签；点击标签设为该组主技能（高亮） */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<SkillsData>(() => props.mod);
const { t } = useI18n();

const parse = (text: string) => text.split(/[,，、]+/).map((s) => s.trim()).filter(Boolean);
</script>

<template>
  <div class="ed">
    <EdList v-slot="{ item: g }" :items="d.groups" :make="() => ({ title: '', items: [], star: '' })" :add-label="t('aboutKit.ed.addGroup')">
      <div class="line">
        <input v-model="g.title" class="a-input w-narrow" type="text" :placeholder="t('aboutKit.ed.groupTitle')" />
        <input
          class="a-input flex-in"
          type="text"
          :value="g.items.join(', ')"
          :placeholder="t('aboutKit.ed.tagsHint')"
          @change="g.items = parse(($event.target as HTMLInputElement).value)"
        />
      </div>
      <div v-if="g.items.length" class="keys">
        <button v-for="k in g.items" :key="k" type="button" :class="{ on: g.star === k }" @click="g.star = g.star === k ? '' : k">{{ k }}</button>
        <span class="hint">{{ t('aboutKit.ed.starHint') }}</span>
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.keys {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;

  button {
    height: 26px;
    padding: 0 10px;
    border-radius: 8px;
    font-size: 12px;
    color: var(--text);
    background: var(--surface);
    border: 1px solid color-mix(in oklab, var(--text) 14%, transparent);
    box-shadow: 0 2px 0 color-mix(in oklab, var(--text) 14%, transparent);

    &.on { color: var(--primary); border-color: rgba(var(--primary-rgb), 0.5); background: rgba(var(--primary-rgb), 0.08); }
  }

  .hint { margin-left: 4px; font-size: 11.5px; }
}
</style>
