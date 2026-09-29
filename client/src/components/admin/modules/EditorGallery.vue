<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { GalleryData, GalleryImage } from '../../../about/types';
import { SCENES, SCENE_LABELS } from '../../../about/icons';
import CoverUploader from '../CoverUploader.vue';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/**
 * 画廊：上传器管理图片顺序（最多 12 张），下方逐张填写 标题 / 地点 / 日期；
 * 也可添加「光影构图」占位（无图片时显示 CSS 场景）。
 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<GalleryData>(() => props.mod);
const { t } = useI18n();

/** 上传器只管 src：按 src 保留已有的元信息 */
const srcs = computed<string[]>({
  get: () => d.value.images.filter((im) => im.src).map((im) => im.src),
  set: (list) => {
    const bySrc = new Map(d.value.images.filter((im) => im.src).map((im) => [im.src, im]));
    const scenes = d.value.images.filter((im) => !im.src);
    // eslint-disable-next-line vue/no-mutating-props
    d.value.images = [...list.map((src) => bySrc.get(src) ?? { src, title: '', place: '', date: '' }), ...scenes];
  },
});

const make = (): GalleryImage => ({ src: '', scene: '05', title: '', place: '', date: '' });
</script>

<template>
  <div class="ed">
    <CoverUploader v-model="srcs" :max="12" />
    <p class="hint">{{ t('aboutKit.ed.galleryHint') }}</p>
    <EdList v-slot="{ item }" :items="d.images" :make="make" :add-label="t('aboutKit.ed.addScene')" :max="16" compact>
      <div class="line">
        <img v-if="item.src" class="thumb" :src="item.src" alt="" />
        <select v-else v-model="item.scene" class="a-input w-narrow">
          <option v-for="s in SCENES" :key="s" :value="s">{{ SCENE_LABELS[s] }}</option>
        </select>
        <input v-model="item.title" class="a-input flex-in" type="text" :placeholder="t('aboutKit.ed.title')" />
        <input v-model="item.place" class="a-input w-narrow" type="text" :placeholder="t('aboutKit.ed.place')" />
        <input v-model="item.date" class="a-input w-narrow" type="text" placeholder="2026.07.02" />
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.thumb { width: 120px; height: 38px; flex-shrink: 0; object-fit: cover; border-radius: var(--r-sm); }
</style>
