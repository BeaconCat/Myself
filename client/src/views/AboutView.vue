<script setup lang="ts">
import { computed } from 'vue';
import '@fontsource/noto-serif-sc/500.css';
import { useConfigStore } from '../stores/config';
import AboutModules from '../about/AboutModules.vue';

/**
 * 关于页（about-kit v2）：全部内容由后台「关于管理」的模块列表驱动，
 * 12 栏 bento + chapter 章节节奏 + reveal 进场；旧的顶层身份字段在缺少 profile 模块时自动合成身份区。
 */
const config = useConfigStore();
const about = computed(() => config.cfg.about);
</script>

<template>
  <main class="about-page">
    <div class="ambient" aria-hidden="true" />
    <div class="wrap">
      <AboutModules :modules="about.modules ?? []" :about="about" />

      <a class="powered" href="https://github.com/BeaconCat/Myself" target="_blank" rel="noopener">
        <img src="/favicon-256.png" alt="" draggable="false" />
        <span class="pw-text">
          <span class="pw-label">Powered by</span>
          <span class="pw-name">Myself</span>
        </span>
        <span class="pw-sub">MIT License · GitHub @BeaconCat</span>
        <svg class="pw-arrow" viewBox="0 0 24 24" aria-hidden="true"><path d="M7 17 17 7M9 7h8v8" /></svg>
      </a>
    </div>
  </main>
</template>

<style scoped lang="scss">
.about-page {
  position: relative;
  padding: 24px 0 110px;
  overflow-x: clip;
}

/* 页顶一团极淡的主色环境光，落在身份区右侧形象图背后 */
.ambient {
  position: absolute;
  top: -120px;
  right: -10%;
  width: min(900px, 80vw);
  height: 760px;
  background: radial-gradient(closest-side, rgba(var(--primary-rgb), 0.12), transparent 70%);
  pointer-events: none;
  filter: blur(20px);
}

:root[data-mode='light'] .ambient { opacity: 0.7; }

.wrap {
  position: relative;
  max-width: 1240px;
  margin: 0 auto;
  padding: 0 32px;
}

.powered {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-top: 40px;
  padding: 18px 22px;
  border-radius: 18px;
  border: 1px solid color-mix(in oklab, var(--text) 9%, transparent);
  color: var(--text-2);
  transition: border-color var(--dur), transform var(--dur) var(--ease-out), box-shadow var(--dur);

  img { width: 38px; height: 38px; transition: transform var(--dur) var(--ease-spring); }

  .pw-text { display: flex; flex-direction: column; line-height: 1.25; }
  .pw-label { font-size: 11.5px; letter-spacing: 0.08em; color: color-mix(in oklab, var(--text-2) 70%, var(--bg)); }
  .pw-name { font: 700 18px var(--font-serif); color: var(--text); }
  .pw-sub { margin-left: auto; font: 400 12px ui-monospace, Consolas, monospace; }

  .pw-arrow {
    width: 18px;
    height: 18px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.8;
    stroke-linecap: round;
    stroke-linejoin: round;
    transition: transform var(--dur) var(--ease-spring), color var(--dur);
  }

  &:hover {
    border-color: rgba(var(--primary-rgb), 0.45);
    transform: translateY(-2px);
    box-shadow: 0 18px 40px -26px rgba(var(--primary-rgb), 0.7);

    img { transform: rotate(-6deg) scale(1.06); }
    .pw-arrow { transform: translate(3px, -3px); color: var(--text); }
  }
}

@media (max-width: 640px) {
  .about-page { padding: 8px 0 90px; }
  .wrap { padding: 0 16px; }
  .powered .pw-sub { display: none; }
  .powered .pw-arrow { margin-left: auto; }
}
</style>
