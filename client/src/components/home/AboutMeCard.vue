<script setup lang="ts">
import { useConfigStore } from '../../stores/config';

/** 关于我卡片：文案与技能标签走站点配置 */
const config = useConfigStore();
</script>

<template>
  <section v-reveal class="me-card">
    <!-- 三色信标顶缘 -->
    <div class="me-beacon" aria-hidden="true">
      <span /><span /><span />
    </div>

    <div class="me-body">
      <div class="me-avatar" aria-hidden="true">
        <img src="/favicon-256.png" alt="" draggable="false" />
      </div>

      <div class="me-info">
        <h2>关于我</h2>
        <p>{{ config.cfg.about.bio }}</p>
        <div class="me-skills">
          <span v-for="s in config.cfg.about.skills" :key="s">{{ s }}</span>
        </div>
      </div>

      <router-link to="/about" class="me-more">
        了解更多
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round">
          <path d="M5 12h14M13 6l6 6-6 6" />
        </svg>
      </router-link>
    </div>
  </section>
</template>

<style scoped lang="scss">
.me-card {
  position: relative;
  background: var(--surface);
  border: 1px solid var(--border);
  overflow: hidden;
}

/* 顶缘三色光带 */
.me-beacon {
  display: flex;
  height: 3px;

  span { flex: 1; }
  span:nth-child(1) { background: var(--accent-red); }
  span:nth-child(2) { background: var(--accent-yellow); }
  span:nth-child(3) { background: var(--accent-blue); }
}

.me-body {
  display: flex;
  align-items: center;
  gap: 26px;
  padding: 30px 28px;
}

.me-avatar {
  width: 84px;
  height: 84px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  border: 2px solid var(--border);
  background: var(--surface-2);
  transition: transform var(--dur) var(--ease-spring), box-shadow var(--dur);

  img { width: 100%; height: 100%; object-fit: cover; }
}

.me-card:hover .me-avatar {
  transform: scale(1.06) rotate(-4deg);
  box-shadow: 0 0 24px rgba(var(--primary-rgb), 0.4);
}

.me-info {
  flex: 1;
  min-width: 0;

  h2 { font-size: 21px; margin-bottom: 10px; }

  p {
    font-size: 14px;
    line-height: 1.8;
    color: var(--text-2);
  }
}

.me-skills {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 14px;

  span {
    font-size: 12px;
    font-weight: 600;
    padding: 3px 12px;
    background: rgba(var(--primary-rgb), 0.1);
    color: var(--primary);
    transition: transform var(--dur-fast) var(--ease-spring);

    &:hover { transform: scale(1.08); }
  }
}

.me-more {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  font-size: 14px;
  font-weight: 700;
  color: #fff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
  padding: 11px 22px;
  background: linear-gradient(180deg, var(--primary), var(--primary-deep));
  box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.45);
  transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast), filter var(--dur-fast);

  svg { width: 17px; height: 17px; transition: transform var(--dur-fast) var(--ease-out); }

  &:hover {
    filter: brightness(1.08);
    transform: scale(1.05);
    box-shadow: 0 8px 24px rgba(var(--primary-rgb), 0.55);

    svg { transform: translateX(3px); }
  }
}

@media (max-width: 768px) {
  .me-body {
    flex-direction: column;
    text-align: center;
  }

  .me-skills { justify-content: center; }
}
</style>
