<script setup lang="ts">
import { computed } from 'vue';
import NavBar from './components/layout/NavBar.vue';
import AppLoading from './components/loading/AppLoading.vue';
import RouteLoading from './components/loading/RouteLoading.vue';
import { useLoadingStore } from './stores/loading';

const loading = useLoadingStore();
/** 路由加载时页面整体缩小 10%（模糊变暗由 RouteLoading 的遮罩层承担） */
const shrunk = computed(() => loading.routeLoading);
</script>

<template>
  <div class="app-shell" :class="{ shrunk }">
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
  transform-origin: 50% 40%;

  &.shrunk {
    transform: scale(0.9);
  }
}
</style>
