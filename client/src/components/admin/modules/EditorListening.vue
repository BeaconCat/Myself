<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import type { ListeningData } from '../../../about/types';
import EdSceneSelect from './EdSceneSelect.vue';
import Switch from '../../ui/Switch.vue';
import EdList from './EdList.vue';
import { useModuleData } from './useModuleData';

/** 最近在听：当前曲目（曲名 / 艺术家 / 专辑 / 时长与位置秒数 / 是否播放）+ 最近曲目 */
const props = defineProps<{ mod: AboutModule }>();
const d = useModuleData<ListeningData>(() => props.mod);
const { t } = useI18n();
</script>

<template>
  <div class="ed">
    <div class="grid3">
      <label><span>{{ t('aboutKit.ed.track') }}</span><input v-model="d.now.title" class="a-input" type="text" /></label>
      <label><span>{{ t('aboutKit.ed.artist') }}</span><input v-model="d.now.artist" class="a-input" type="text" /></label>
      <label><span>{{ t('aboutKit.ed.album') }}</span><input v-model="d.now.album" class="a-input" type="text" /></label>
      <label><span>{{ t('aboutKit.ed.duration') }}</span><input v-model.number="d.now.duration" class="a-input" type="number" min="1" /></label>
      <label><span>{{ t('aboutKit.ed.position') }}</span><input v-model.number="d.now.position" class="a-input" type="number" min="0" /></label>
      <Switch v-model="d.playing" class="pl">{{ t('aboutKit.ed.playing') }}</Switch>
    </div>
    <div class="sub-title">{{ t('aboutKit.ed.recent') }}</div>
    <EdList v-slot="{ item }" :items="d.recent" :make="() => ({ title: '', artist: '', at: '', scene: '01' })" :max="6" compact>
      <div class="grid4">
        <input v-model="item.title" class="a-input" type="text" :placeholder="t('aboutKit.ed.track')" />
        <input v-model="item.artist" class="a-input" type="text" :placeholder="t('aboutKit.ed.artist')" />
        <input v-model="item.at" class="a-input" type="text" :placeholder="t('aboutKit.ed.when')" />
        <EdSceneSelect v-model="item.scene" />
      </div>
    </EdList>
  </div>
</template>

<style scoped lang="scss">
@use './module-editor';

.pl { align-self: end; }
</style>
