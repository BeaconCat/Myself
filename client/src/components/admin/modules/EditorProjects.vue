<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { Project, ProjectsData } from '../../../about/types';
import { SCENES, SCENE_LABELS } from '../../../about/icons';
import CoverUploader from '../CoverUploader.vue';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 作品：名称 / 链接 / 描述 / 封面（图片或光影构图）/ 语言与颜色 / Stars / 精选 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<ProjectsData>(() => props.mod);
const { t } = useI18n();

const coverOf = (p: Project) => (p.cover ? [p.cover] : []);
function setCover(p: Project, v: string[]): void { p.cover = v[0] ?? ''; }
function feature(p: Project): void {
  for (const x of d.value.items) x.featured = x === p;
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
      <div class="grid3">
        <input v-model="item.name" class="a-input" type="text" :placeholder="t('aboutKit.ed.itemName')" />
        <input v-model="item.url" class="a-input span2" type="text" placeholder="https://…" />
      </div>
      <textarea v-model="item.desc" class="a-input" rows="2" :placeholder="t('aboutKit.ed.desc')" />
      <div class="proj">
        <CoverUploader :model-value="coverOf(item)" :max="1" @update:model-value="(v: string[]) => setCover(item, v)" />
        <div class="grid2 flex-in">
          <label>
            <span>{{ t('aboutKit.ed.scene') }}</span>
            <select v-model="item.scene" class="a-input" :disabled="!!item.cover">
              <option v-for="s in SCENES" :key="s" :value="s">{{ SCENE_LABELS[s] }}</option>
            </select>
          </label>
          <label><span>Stars</span><input v-model.number="item.stars" class="a-input" type="number" min="0" /></label>
          <label>
            <span>{{ t('aboutKit.ed.lang') }}</span>
            <div class="line"><input v-model="item.lang" class="a-input flex-in" type="text" /><input v-model="item.color" class="color-in" type="color" /></div>
          </label>
          <label class="check feat"><input type="radio" :checked="!!item.featured" @change="feature(item)" />{{ t('aboutKit.ed.featured') }}</label>
        </div>
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.proj { display: flex; gap: 14px; align-items: flex-start; }
.feat { align-self: end; padding-bottom: 8px; }

@media (max-width: 720px) { .proj { flex-direction: column; } }
</style>
