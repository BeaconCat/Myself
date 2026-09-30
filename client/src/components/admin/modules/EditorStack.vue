<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { StackData, StackItem } from '../../../about/types';
import { TECH, techInk } from '../../../about/tech';
import TechIcon from '../../../about/parts/TechIcon.vue';
import Select from '../../ui/Select.vue';
import ColorSwatch from '../../ui/ColorSwatch.vue';
import type { UiOption } from '../../ui/listbox';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/**
 * 技术栈：品牌图标（simple-icons，可搜索）/ 名称 / 角色 / 文字徽标（无图标时）/ 品牌色 / 链接。
 * 选图标时颜色取官方品牌色、空名称顺带填上品牌名；选「文字徽标」则回到 1–3 个字母的等宽徽标。
 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<StackData>(() => props.mod);
const { t } = useI18n();

const options = computed<UiOption[]>(() => [
  { value: '', label: t('aboutKit.ed.noIcon') },
  ...Object.entries(TECH)
    .map(([value, x]) => ({ value, label: x.label, keywords: x.keywords }))
    .sort((a, b) => a.label.localeCompare(b.label, 'en', { sensitivity: 'base' })),
]);

function setIcon(it: StackItem, key: string): void {
  it.icon = key;
  const x = TECH[key];
  if (!x) return;
  it.color = x.hex;
  if (!it.name.trim()) it.name = x.label;
}

const make = (): StackItem => ({ name: '', role: '', icon: '', glyph: '', color: '#0078ff', url: '' });
</script>

<template>
  <div class="ed">
    <EdList v-slot="{ item }" :items="d.items" :make="make" compact>
      <div class="s-row">
        <Select
          :model-value="item.icon ?? ''"
          :options="options"
          :min-width="240"
          searchable
          :search-placeholder="t('aboutKit.ed.searchTech')"
          :empty="t('aboutKit.ed.noMatch')"
          :aria-label="t('aboutKit.ed.techIcon')"
          @update:model-value="setIcon(item, $event)"
        >
          <template #icon="{ option }">
            <TechIcon v-if="option.value" :name="option.value" :size="17" :style="{ color: techInk(TECH[option.value]?.hex) }" />
            <span v-else class="gl" :style="{ '--c': techInk(item.color) || 'var(--primary)' }">{{ item.glyph || item.name.slice(0, 2) || 'Aa' }}</span>
          </template>
        </Select>
        <input v-model="item.name" class="a-input nm c-nm" type="text" :placeholder="t('aboutKit.ed.itemName')" :aria-label="t('aboutKit.ed.itemName')" />
        <input v-model="item.role" class="a-input c-rl" type="text" :placeholder="t('aboutKit.ed.role')" :aria-label="t('aboutKit.ed.role')" />
        <input
          v-model="item.glyph"
          class="a-input glyph c-gl"
          type="text"
          maxlength="3"
          :placeholder="t('aboutKit.ed.glyph')"
          :disabled="!!TECH[item.icon ?? '']"
          :title="t('aboutKit.ed.glyphHint')"
          :aria-label="t('aboutKit.ed.glyph')"
        />
        <ColorSwatch v-model="item.color" class="c-co" :label="t('aboutKit.ed.color')" show-hex />
        <input v-model="item.url" class="a-input c-url" type="text" placeholder="https://…" :aria-label="t('aboutKit.ed.link')" />
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.s-row {
  display: grid;
  grid-template-columns: minmax(160px, 190px) minmax(0, 0.9fr) minmax(0, 1.2fr) 64px auto minmax(0, 1fr);
  align-items: center;
  gap: 8px;
}

.nm { font-weight: 500; }
.glyph { text-align: center; font-family: var(--font-mono); &:disabled { opacity: 0.45; } }

/* 文字徽标预览（与前台 StackModule 同一字形规则） */
.gl {
  display: grid;
  place-items: center;
  min-width: 22px;
  height: 20px;
  padding: 0 3px;
  border-radius: var(--r-xs);
  background: var(--fill);
  font: 600 10.5px var(--font-mono);
  color: color-mix(in oklab, var(--c) 60%, var(--text));
}

/* 窄时两行：图标 · 名称 · 角色 / 徽标 · 颜色 · 链接 */
@container ed (max-width: 900px) {
  .s-row {
    grid-template-columns: minmax(140px, 170px) auto minmax(0, 1fr) minmax(0, 1.3fr);
    grid-template-areas: 'ic nm nm rl' 'gl co url url';
  }
  .s-row :deep(.ui-sel) { grid-area: ic; }
  .c-nm { grid-area: nm; }
  .c-rl { grid-area: rl; }
  .c-gl { grid-area: gl; }
  .c-co { grid-area: co; }
  .c-url { grid-area: url; }
}

@container ed (max-width: 520px) {
  .s-row { grid-template-columns: minmax(0, 1fr) auto; grid-template-areas: 'ic ic' 'nm nm' 'rl rl' 'gl co' 'url url'; }
}
</style>
