<script setup lang="ts">
import { computed, ref, type Component } from 'vue';
import type { AboutModule } from '../stores/config';
import { metaOf } from './registry';
import StatsModule from './modules/StatsModule.vue';
import MottoModule from './modules/MottoModule.vue';
import SkillsModule from './modules/SkillsModule.vue';
import SkillbarsModule from './modules/SkillbarsModule.vue';
import LanguagesModule from './modules/LanguagesModule.vue';
import MilestonesModule from './modules/MilestonesModule.vue';
import GalleryModule from './modules/GalleryModule.vue';
import QuotesModule from './modules/QuotesModule.vue';
import DevicesModule from './modules/DevicesModule.vue';
import FavoritesModule from './modules/FavoritesModule.vue';
import FaqModule from './modules/FaqModule.vue';
import NowModule from './modules/NowModule.vue';
import GithubModule from './modules/GithubModule.vue';
import SocialsModule from './modules/SocialsModule.vue';
import StackModule from './modules/StackModule.vue';
import ContactModule from './modules/ContactModule.vue';

/** 关于页模块分发器：按配置顺序渲染全部模块，type → 渲染组件 */
const props = defineProps<{ modules: AboutModule[] }>();

const COMPONENTS: Record<string, Component> = {
  stats: StatsModule,
  motto: MottoModule,
  skills: SkillsModule,
  skillbars: SkillbarsModule,
  languages: LanguagesModule,
  milestones: MilestonesModule,
  gallery: GalleryModule,
  quotes: QuotesModule,
  devices: DevicesModule,
  favorites: FavoritesModule,
  faq: FaqModule,
  now: NowModule,
  github: GithubModule,
  socials: SocialsModule,
  stack: StackModule,
  contact: ContactModule,
};

const known = computed(() => props.modules.filter((m) => metaOf(m.type) && COMPONENTS[m.type]));

/* faq 展开（全页共享：同一时刻仅一条展开） */
const openFaq = ref<string>('');
</script>

<template>
  <div class="mods">
    <section
      v-for="mod in known"
      :key="mod.id"
      v-reveal
      class="mod"
      :class="`mod-${mod.type}`"
    >
      <FaqModule v-if="mod.type === 'faq'" v-model:open="openFaq" :mod="mod" />
      <component :is="COMPONENTS[mod.type]" v-else :mod="mod" />
    </section>
  </div>
</template>

<style scoped lang="scss">
.mod { margin-bottom: 56px; }
</style>
