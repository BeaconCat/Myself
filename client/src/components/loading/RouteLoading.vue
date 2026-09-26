<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useLoadingStore } from '../../stores/loading';
import { useConfigStore } from '../../stores/config';

/**
 * 路由切换全屏 Loading：
 * 进场：模糊+变暗层浮现（页面同时被 App 缩小 10%），纯色层从上往下切下覆盖全屏，
 *       中央文本 + 旋转指示器。
 * 退场：整个 Loading（含模糊层）被容器 clip-path 裁切，随纯色层向下丝滑消失。
 * 交接：routeLoading 变 false 即「揭幕开始」—— 本组件同帧开始退场，新页入场动画同帧开播（stores/loading.ts）。
 */
const loading = useLoadingStore();
const config = useConfigStore();

type Stage = 'idle' | 'enter' | 'leave';
const stage = ref<Stage>('idle');
const LEAVE_MS = 700;
const reduced = (): boolean => window.matchMedia('(prefers-reduced-motion: reduce)').matches;

watch(
  () => loading.routeLoading,
  (on) => {
    if (on) {
      stage.value = 'enter';
    } else if (stage.value === 'enter') {
      stage.value = 'leave';
      window.setTimeout(() => {
        if (stage.value === 'leave') stage.value = 'idle';
        // 遮罩彻底离屏（页面入场早在揭幕开始时已开播）
        loading.routeOverlayVisible = false;
      }, reduced() ? 0 : LEAVE_MS);
    }
  },
);

const visible = computed(() => stage.value !== 'idle');
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="route-loading" :class="stage">
      <!-- 模糊 + 变暗（被容器 clip 一并裁切） -->
      <div class="dim" />
      <!-- 纯色覆盖层：从上往下切入 -->
      <div class="panel" />
      <!-- 中央指示器 -->
      <div class="indicator">
        <span class="spinner" />
        <span class="label">{{ config.cfg.loading.routeText }}</span>
      </div>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.route-loading {
  position: fixed;
  inset: 0;
  z-index: 9000;
  overflow: hidden;

  /* 退场：容器 clip 上缘下移，整体被裁切向下消失 */
  &.leave {
    animation: wipe-away 0.7s cubic-bezier(0.65, 0, 0.35, 1) forwards;
  }
}

@keyframes wipe-away {
  from { clip-path: inset(0 0 0 0); }
  to { clip-path: inset(100% 0 0 0); }
}

.dim {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
}

.route-loading.enter .dim {
  animation: dim-in 0.3s ease both;
}

@keyframes dim-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* 纯色层：主色深调，从上往下切下 */
.panel {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, var(--primary-deep), var(--primary));
}

.route-loading.enter .panel {
  animation: panel-down 0.55s cubic-bezier(0.65, 0, 0.35, 1) 0.12s both;
}

@keyframes panel-down {
  from { transform: translateY(-100%); }
  to { transform: translateY(0); }
}

.indicator {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 18px;
  color: #fff;
}

.route-loading.enter .indicator {
  animation: dim-in 0.35s ease 0.4s both;
}

.spinner {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  border: 3px solid rgba(255, 255, 255, 0.3);
  border-top-color: #fff;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.label {
  font-family: var(--font-serif);
  font-size: 17px;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
  animation: label-pulse 1.6s ease-in-out infinite;
}

@keyframes label-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}
</style>
