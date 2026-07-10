<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import NavBar from './components/layout/NavBar.vue';
import AppLoading from './components/loading/AppLoading.vue';
import RouteLoading from './components/loading/RouteLoading.vue';
import { useLoadingStore } from './stores/loading';

const loading = useLoadingStore();
/** 路由加载时页面整体缩小 10%（模糊变暗由 RouteLoading 的遮罩层承担） */
const shrunk = computed(() => loading.routeLoading);
/** 幕布落下（首屏/路由 loading 覆盖中）：页面动画整体暂停，揭幕才播 */
const covered = computed(() => !loading.bootDone || loading.routeLoading);

/* 缩放原点锁定当前视口中心（元素比视口高时 50% 会落到视口外，表现为向下缩） */
const originY = ref('50vh');
watch(shrunk, (on) => {
  if (on) originY.value = `${window.scrollY + window.innerHeight / 2}px`;
});
</script>

<template>
  <div class="app-shell" :class="{ shrunk, covered }" :style="{ transformOrigin: `50% ${originY}` }">
    <NavBar />
    <router-view v-slot="{ Component }">
      <transition name="page" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </div>
  <RouteLoading />
  <AppLoading />
</template>

<style scoped lang="scss">
.app-shell {
  min-height: 100vh;
  transition: transform 0.55s var(--ease-out);

  &.shrunk {
    transform: scale(0.9);
  }
}
</style>
