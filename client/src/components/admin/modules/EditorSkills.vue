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

  /* 描边胶囊；主技能 = 抬升 + 轻染 + 前置 4px 主色圆点（与前台 chip 选中同构） */
  button {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 26px;
    padding: 0 11px;
    border: 0;
    border-radius: var(--r-pill);
    font-size: 12px;
    color: var(--text-2);
    background: none;
    box-shadow: inset 0 0 0 1px var(--line);
    transition: color var(--dur-fast), background-color var(--dur-fast), box-shadow var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

    &:hover { color: var(--text); background: var(--fill); box-shadow: inset 0 0 0 1px var(--line-2); }
    &:active { transform: scale(0.96); }

    &.on { color: var(--lift-fg); font-weight: 500; background: var(--lift); box-shadow: var(--lift-shadow); }
    &.on::before { content: ''; width: 4px; height: 4px; border-radius: 50%; background: var(--ink); }
  }

  .hint { margin-left: 4px; font-size: 11.5px; }
}
</style>
