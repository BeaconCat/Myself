<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { Star } from 'lucide';
import type { AboutModule } from '../../../stores/config';
import type { Project, ProjectsData } from '../../../about/types';
import Icon from '../../ui/Icon.vue';
import ColorSwatch from '../../ui/ColorSwatch.vue';
import Switch from '../../ui/Switch.vue';
import EdCover from './EdCover.vue';
import EdSceneSelect from './EdSceneSelect.vue';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/**
 * 作品：左侧封面位（上传 / 素材库 / 无图时显示所选光影），右侧 名称 · 链接 / 描述 /
 * 光影 · 语言与颜色 · Stars · 精选（全组唯一，打开即取消其他项）。
 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<ProjectsData>(() => props.mod);
const { t } = useI18n();

function setFeatured(p: Project, on: boolean): void {
  for (const x of d.value.items) x.featured = on && x === p;
}
</script>

<template>
  <div class="ed">
    <EdList
      v-slot="{ item }"
      :items="d.items"
      :make="() => ({ name: '', desc: '', url: '', cover: '', scene: '05', lang: '', color: '#0078ff', stars: 0 })"
      :max="6"
    >
      <div class="pj">
        <EdCover v-model:src="item.cover" class="pj-cov" :scene="item.scene" :pick-title="t('aboutKit.ed.pickCover')" />
        <div class="pj-main">
          <div class="pj-r1">
            <input v-model="item.name" class="a-input nm" type="text" :placeholder="t('aboutKit.ed.itemName')" :aria-label="t('aboutKit.ed.itemName')" />
            <input v-model="item.url" class="a-input" type="text" placeholder="https://…" :aria-label="t('aboutKit.ed.link')" />
          </div>
          <textarea v-model="item.desc" class="a-input" rows="2" :placeholder="t('aboutKit.ed.desc')" :aria-label="t('aboutKit.ed.desc')" />
          <div class="pj-r3">
            <EdSceneSelect v-model="item.scene" :disabled="!!item.cover" />
            <div class="lang">
              <input v-model="item.lang" class="a-input" type="text" :placeholder="t('aboutKit.ed.langName')" :aria-label="t('aboutKit.ed.langName')" />
              <ColorSwatch v-model="item.color" :label="t('aboutKit.ed.color')" />
            </div>
            <label class="affix" :title="t('aboutKit.ed.stars')">
              <Icon class="aff" :icon="Star" :size="15" />
              <input v-model.number="item.stars" type="number" min="0" :aria-label="t('aboutKit.ed.stars')" />
            </label>
            <Switch :model-value="!!item.featured" @update:model-value="setFeatured(item, $event)">{{ t('aboutKit.ed.featured') }}</Switch>
          </div>
        </div>
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.pj { display: flex; gap: 14px; align-items: stretch; }

.pj-cov { flex: none; width: 200px; aspect-ratio: 4 / 3; align-self: flex-start; }

.pj-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 8px; }

.pj-r1 { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 2fr); gap: 8px; }
.nm { font-weight: 600; }

.pj-r3 {
  display: grid;
  grid-template-columns: minmax(150px, 1.1fr) minmax(150px, 1.2fr) minmax(90px, 0.6fr) auto;
  align-items: center;
  gap: 8px 10px;
}

.lang { display: flex; gap: 6px; min-width: 0; }

@container ed (max-width: 820px) {
  .pj-r3 { grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); }
}

@container ed (max-width: 600px) {
  .pj { flex-direction: column; }
  .pj-cov { width: 100%; max-width: 320px; }
  .pj-r1, .pj-r3 { grid-template-columns: minmax(0, 1fr); }
}
</style>
