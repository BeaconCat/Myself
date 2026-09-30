<script setup lang="ts">
import { computed } from 'vue';
import '@fontsource/noto-serif-sc/500.css';
import { useConfigStore } from '../stores/config';
import AboutModules from '../about/AboutModules.vue';

/**
 * 关于页（about-kit v2）：版面由后台「关于」的模块列表驱动、身份内容来自「身份」，
 * 12 栏 bento + chapter 章节节奏 + reveal 进场；缺少 profile 模块时自动补一个身份区。
 * 页尾品牌信息由全站 SiteFooter 承担，这里不再重复。
 */
const config = useConfigStore();
const about = computed(() => config.cfg.about);
</script>

<template>
  <main class="about-page">
    <div class="wrap">
      <AboutModules :modules="about.modules ?? []" :about="about" />
    </div>
  </main>
</template>

<style scoped lang="scss">
.about-page {
  position: relative;
  /* 桌面：让出悬浮胶囊导航的高度，身份区（尤其右侧形象图）不与导航重叠 */
  padding: 56px 0 32px;
  overflow-x: clip;
}

.wrap {
  position: relative;
  max-width: 1240px;
  margin: 0 auto;
  padding: 0 32px;
}

@media (max-width: 640px) {
  .about-page { padding: 4px 0 72px; }
  .wrap { padding: 0 16px; }
}
</style>
