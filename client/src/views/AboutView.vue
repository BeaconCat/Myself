<script setup lang="ts">
import { computed } from 'vue';
import { useConfigStore } from '../stores/config';
import AboutModules from '../about/AboutModules.vue';

/** 关于页：身份 Hero 固定，其余全部由后台「关于管理」的模块列表驱动 */
const config = useConfigStore();
const about = computed(() => config.cfg.about);
</script>

<template>
  <main class="page">
    <!-- 身份 Hero（默认模块，不可移除） -->
    <section v-reveal class="hero">
      <div class="hero-avatar">
        <img :src="about.avatar || '/favicon-256.png'" :alt="about.name" draggable="false" />
        <span class="ring" aria-hidden="true" />
      </div>
      <h1 class="hero-name">{{ about.name }}</h1>
      <p class="hero-line">{{ about.tagline }}</p>
      <p class="hero-bio">{{ about.bio }}</p>
      <blockquote v-if="about.motto" class="hero-motto">{{ about.motto }}</blockquote>
      <div class="hero-dots" aria-hidden="true">
        <span style="--c: #ff0032" /><span style="--c: #ffb300" /><span style="--c: #0078ff" />
      </div>
    </section>

    <!-- 可配置模块流 -->
    <AboutModules :modules="about.modules" />

    <!-- Powered by -->
    <a
      v-reveal
      class="powered"
      href="https://github.com/BeaconCat/Myself"
      target="_blank"
      rel="noopener"
    >
      <span class="pw-glow g-red" aria-hidden="true" />
      <span class="pw-glow g-yellow" aria-hidden="true" />
      <span class="pw-glow g-blue" aria-hidden="true" />
      <span class="pw-sheen" aria-hidden="true" />

      <img class="pw-logo" src="/favicon-256.png" alt="" draggable="false" />
      <span class="pw-text">
        <span class="pw-label">Powered by</span>
        <span class="pw-name">Myself</span>
        <span class="pw-sub">GitHub {{ '@BeaconCat' }} · MIT License</span>
      </span>
      <span class="pw-arrow" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M7 17L17 7M9 7h8v8" />
        </svg>
      </span>
    </a>
  </main>
</template>

<style scoped lang="scss">
.page {
  max-width: 860px;
  margin: 0 auto;
  padding: 120px 24px 90px;
}

/* ===== 身份 Hero ===== */
.hero {
  text-align: center;
  margin-bottom: 64px;
}

.hero-avatar {
  position: relative;
  width: 112px;
  height: 112px;
  margin: 0 auto 22px;

  img {
    width: 100%;
    height: 100%;
    border-radius: 50%;
    border: 2px solid var(--border);
    background: var(--surface-2);
    object-fit: cover;
  }

  .ring {
    position: absolute;
    inset: -8px;
    border-radius: 50%;
    border: 2px solid rgba(var(--primary-rgb), 0.4);
    animation: ring-pulse 2.6s ease-in-out infinite;
  }
}

@keyframes ring-pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.08); opacity: 0.4; }
}

.hero-name {
  font-size: clamp(36px, 5vw, 52px);
  background: var(--grad-title);
  background-clip: text;
  -webkit-background-clip: text;
  color: transparent;
}

.hero-line {
  margin-top: 8px;
  font-size: 15px;
  font-weight: 600;
  color: var(--primary);
  letter-spacing: 0.12em;
}

.hero-bio {
  max-width: 52ch;
  margin: 18px auto 0;
  font-size: 15px;
  line-height: 2;
  color: var(--text-2);
}

.hero-motto {
  margin: 20px auto 0;
  padding: 10px 26px;
  width: fit-content;
  max-width: 100%;
  font-family: var(--font-serif);
  font-size: 15px;
  color: var(--text-2);
  border-left: 3px solid rgba(var(--primary-rgb), 0.6);
  border-right: 3px solid rgba(var(--primary-rgb), 0.6);
  background: rgba(var(--primary-rgb), 0.05);
  border-radius: 8px;
}

.hero-dots {
  display: flex;
  justify-content: center;
  gap: 14px;
  margin-top: 26px;

  span {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--c);
    box-shadow: 0 0 12px var(--c);
    animation: dot-breathe 2.2s ease-in-out infinite;

    &:nth-child(2) { animation-delay: 0.35s; }
    &:nth-child(3) { animation-delay: 0.7s; }
  }
}

@keyframes dot-breathe {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.35); }
}

/* ===== Powered by ===== */
.powered {
  position: relative;
  display: flex;
  align-items: center;
  gap: 22px;
  margin-top: 72px;
  padding: 34px 38px;
  overflow: hidden;
  background:
    radial-gradient(600px 220px at 18% 0%, rgba(var(--primary-rgb), 0.1), transparent 60%),
    var(--surface);
  border: 1px solid var(--border);
  color: var(--text);
  transition: transform var(--dur) var(--ease-out), box-shadow var(--dur), border-color var(--dur);

  &:hover {
    transform: scale(1.015);
    border-color: rgba(var(--primary-rgb), 0.5);
    box-shadow: 0 24px 70px -20px rgba(var(--primary-rgb), 0.4);

    .pw-glow { opacity: 0.4; }
    .pw-sheen { animation: sheen-sweep 1.1s var(--ease-out); }
    .pw-arrow { transform: translate(4px, -4px); color: var(--primary); }
    .pw-logo { transform: rotate(-5deg) scale(1.08); }
  }
}

.pw-glow {
  position: absolute;
  width: 220px;
  height: 220px;
  border-radius: 50%;
  filter: blur(70px);
  opacity: 0.18;
  transition: opacity var(--dur-slow) ease;
  pointer-events: none;

  &.g-red { background: var(--primary); left: -60px; top: -110px; }
  &.g-yellow { background: var(--primary-deep); left: 40%; bottom: -160px; }
  &.g-blue { background: var(--primary); right: -60px; top: -100px; }
}

.pw-sheen {
  position: absolute;
  inset: 0;
  background: linear-gradient(115deg, transparent 30%, rgba(var(--primary-rgb), 0.1) 48%, transparent 62%);
  transform: translateX(-120%);
  pointer-events: none;
}

@keyframes sheen-sweep {
  from { transform: translateX(-120%); }
  to { transform: translateX(120%); }
}

.pw-logo {
  width: 64px;
  height: 64px;
  flex-shrink: 0;
  transition: transform var(--dur) var(--ease-spring);
  filter: drop-shadow(0 6px 18px rgba(var(--primary-rgb), 0.4));
}

.pw-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.pw-label {
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.08em;
  color: var(--text-2);
}

.pw-name {
  font-family: var(--font-serif);
  font-size: 30px;
  font-weight: 700;
  line-height: 1.25;
  padding-bottom: 0.12em;
  margin-top: -3px;
  background: var(--grad-title);
  background-clip: text;
  -webkit-background-clip: text;
  color: transparent;
}

.pw-sub {
  font-size: 12.5px;
  color: var(--text-2);
}

.pw-arrow {
  width: 26px;
  height: 26px;
  flex-shrink: 0;
  color: var(--text-2);
  transition: transform var(--dur-fast) var(--ease-out), color var(--dur-fast);

  svg { width: 100%; height: 100%; }
}

@media (max-width: 768px) {
  .page { padding-top: 96px; }
  .powered { padding: 26px 22px; gap: 16px; }
  .pw-name { font-size: 24px; }
}
</style>
