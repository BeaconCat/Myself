<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import NavBar from './components/layout/NavBar.vue';
import AppLoading from './components/loading/AppLoading.vue';
import RouteLoading from './components/loading/RouteLoading.vue';
import AppModal from './components/ui/AppModal.vue';
import SearchPalette from './components/search/SearchPalette.vue';
import SiteFooter from './components/common/SiteFooter.vue';
import { useLoadingStore } from './stores/loading';
import { useDevice } from './composables/useDevice';
import MobileShell from './components/mobile/MobileShell.vue';

const loading = useLoadingStore();
/** 路由加载时页面整体缩小 10%（模糊变暗由 RouteLoading 的遮罩层承担） */
const shrunk = computed(() => loading.routeLoading);
/**
 * 幕布盖住（首屏/路由遮罩完全遮挡）：页面入场动画按住在首帧；
 * 遮罩开始退场的同一帧即放开，入场与退场重叠交接（见 stores/loading.ts）。
 */
const covered = computed(() => loading.curtain);

/** 后台等 bare 页面不渲染前台导航 / 页脚 / 搜索 */
const route = useRoute();
const bare = computed(() => !!route.meta.bare);
/** 移动端前台走独立外壳（底栏 + 抽屉 + SearchOverlay），后台由 AdminRoot 自行切换 */
const { isMobile } = useDevice();
const mobileShell = computed(() => isMobile.value && !bare.value);

/* 缩放原点锁定当前视口中心（元素比视口高时 50% 会落到视口外，表现为向下缩） */
const originY = ref('50vh');
watch(shrunk, (on) => {
  if (on) originY.value = `${window.scrollY + window.innerHeight / 2}px`;
});
</script>

<template>
  <MobileShell v-if="mobileShell" />
  <div
    v-else
    class="app-shell"
    :class="{ shrunk, covered, bare }"
    :style="{ transformOrigin: `50% ${originY}` }"
  >
    <NavBar v-if="!bare" />
    <div class="route-view">
      <!-- 换页发生在路由遮罩背后，不再套 out-in 过渡：旧页根节点自带的 transition 会被 Vue 当成离场时长，
           白白推迟新页挂载；入场由遮罩揭幕 + 页面自身 rise / reveal 承担 -->
      <router-view />
    </div>
    <SiteFooter v-if="!bare" />
    <SearchPalette v-if="!bare" />
  </div>
  <RouteLoading />
  <AppLoading />
  <AppModal />
</template>

<style scoped lang="scss">
.app-shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  transition: transform 0.55s var(--ease-out);

  &.shrunk {
    transform: scale(0.9);
  }
}

.route-view {
  flex: 1 0 auto;
  min-width: 0;
}
</style>
