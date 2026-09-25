<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { useLoadingStore } from '../../stores/loading';
import { useConfigStore } from '../../stores/config';
import { shell } from './shell';

/**
 * 移动端品牌 loading：门形 logo 居中，一道光从门缝扫过，下方细进度线；
 * 首屏资源就绪后 logo 轻微放大淡出、幕布溶解，舞台从 1.04 回落入场。最短展示 650ms。
 */
const loading = useLoadingStore();
const config = useConfigStore();
const leaving = ref(false);
const gone = ref(false);
let shownAt = 0;

function finish(): void {
  if (leaving.value) return;
  const wait = Math.max(0, 650 - (performance.now() - shownAt));
  window.setTimeout(() => {
    leaving.value = true;
    shell.booted = true;
    window.setTimeout(() => { gone.value = true; }, 700);
  }, wait);
}

onMounted(() => {
  shownAt = performance.now();
  if (loading.bootDone) finish();
});
watch(() => loading.bootDone, (done) => { if (done) finish(); });
</script>

<template>
  <div v-if="!gone" class="splash" :class="{ leaving }" aria-hidden="true">
    <div class="mark">
      <img src="/favicon-256.png" alt="" draggable="false" />
      <i class="sweep" />
    </div>
    <div class="name">{{ config.cfg.loading.bootText || config.cfg.site.title }}</div>
    <div class="bar"><i :style="{ transform: `scaleX(${loading.bootProgress / 100})` }" /></div>
  </div>
</template>

<style scoped lang="scss">
.splash {
  position: fixed;
  z-index: 300;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 18px;
  background: var(--bg);
  transition: opacity 0.55s var(--ease-out) 0.1s;

  &.leaving {
    opacity: 0;
    pointer-events: none;
  }
}

.mark {
  position: relative;
  width: 84px;
  height: 84px;
  border-radius: var(--r-xl);
  overflow: hidden;
  box-shadow: var(--shadow-pop);
  animation: mark-in 0.7s var(--ease-spring) both;
  transition: transform 0.6s var(--ease-out), opacity 0.45s var(--ease-out);

  img {
    width: 100%;
    height: 100%;
    display: block;
  }

  .leaving & {
    transform: scale(1.35);
    opacity: 0;
  }
}

/* 门缝扫光（品牌光影，允许发光） */
.sweep {
  position: absolute;
  inset: -20%;
  background: linear-gradient(100deg, transparent 38%, rgba(255, 255, 255, 0.55) 50%, transparent 62%);
  transform: translateX(-80%);
  animation: sweep 1.6s var(--ease-out) 0.35s infinite;
}

.name {
  font-family: var(--font-serif);
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--text);
  animation: fade-up 0.6s var(--ease-out) 0.15s both;
  transition: opacity 0.3s;

  .leaving & { opacity: 0; }
}

.bar {
  width: 72px;
  height: 3px;
  border-radius: 999px;
  background: var(--fill-3);
  overflow: hidden;
  animation: fade-up 0.5s var(--ease-out) 0.25s both;
  transition: opacity 0.3s;

  i {
    display: block;
    height: 100%;
    border-radius: inherit;
    /* 品牌三色常量（非主色渐变），仅用于启动进度 */
    background: linear-gradient(90deg, var(--accent-red), var(--accent-yellow), var(--accent-blue));
    transform-origin: left center;
    transition: transform 0.35s var(--ease-out);
  }

  .leaving & { opacity: 0; }
}

@keyframes mark-in {
  from {
    opacity: 0;
    transform: scale(0.7);
  }
}

@keyframes fade-up {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
}

@keyframes sweep {
  to { transform: translateX(80%); }
}
</style>
