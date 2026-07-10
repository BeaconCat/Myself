<script setup lang="ts">
import { ref, watch } from 'vue';
import { useLoadingStore } from '../../stores/loading';
import { useConfigStore } from '../../stores/config';

/** 首屏加载：文本 + 进度条，完成后 clip-path 圆形收缩退场（Nebula Inspector 同款） */
const loading = useLoadingStore();
const config = useConfigStore();

const fadeOut = ref(false);
const removed = ref(false);

watch(
  () => loading.bootDone,
  (done) => {
    if (!done) return;
    window.setTimeout(() => {
      fadeOut.value = true;
      window.setTimeout(() => { removed.value = true; }, 1400);
    }, 250);
  },
  // 若加载在组件挂载前已完成（缓存命中极快），立即触发退场
  { immediate: true },
);
</script>

<template>
  <div v-if="!removed" class="boot" :class="{ 'fade-out': fadeOut }" aria-live="polite">
    <div class="boot-content">
      <div class="boot-text">{{ config.cfg.loading.bootText }}</div>
      <div class="boot-bar">
        <div class="boot-bar-fill" :style="{ width: loading.bootProgress + '%' }" />
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.boot {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: var(--bg);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  clip-path: circle(200% at 50% 50%);

  &.fade-out {
    animation: boot-clip-shrink 1.3s cubic-bezier(0.65, 0, 0.35, 1) forwards;
  }
}

@keyframes boot-clip-shrink {
  0% { clip-path: circle(200% at 50% 50%); }
  100% { clip-path: circle(0% at 50% 50%); }
}

.boot-content {
  width: min(480px, 80%);
  text-align: center;
}

.boot-text {
  font-family: var(--font-serif);
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 18px;
  opacity: 0;
  transform: translateY(10px);
  animation: boot-fade-in 0.6s cubic-bezier(0.34, 1.56, 0.64, 1) 0.15s forwards;
}

.boot-bar {
  height: 4px;
  width: 33%;
  margin: 0 auto;
  border-radius: 999px;
  background: var(--surface-2);
  overflow: hidden;
  opacity: 0;
  transform: scale(0.85);
  animation: boot-fade-in 0.45s cubic-bezier(0.34, 1.56, 0.64, 1) 0.2s forwards;
}

.boot-bar-fill {
  height: 100%;
  width: 0%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--primary), var(--primary-deep));
  box-shadow: 0 0 8px rgba(var(--primary-rgb), 0.6);
  transition: width 0.25s ease-out;
}

@keyframes boot-fade-in {
  from { opacity: 0; transform: scale(0.85) translateY(10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}
</style>
