<script setup lang="ts">
import { useConfigStore } from '../stores/config';

/** 关于页：身份 Hero + 技能 + 历程 + 本站技术栈 + 联系方式（身份文案走站点配置） */
const config = useConfigStore();

const skillGroups = [
  { title: '前端', items: ['Vue', 'TypeScript', 'Vite', 'SCSS', '动效设计'] },
  { title: '后端', items: ['Node.js', 'Express', 'SQLite', 'RESTful'] },
  { title: '兴趣', items: ['摄影', 'UI 设计', '写作'] },
];

const milestones = [
  { year: '2026', text: '启动 Myself 博客计划，代号定名，灯塔点亮' },
  { year: '2025', text: '开始系统性沉淀技术笔记与四季随想' },
  { year: '2024', text: '第一次把玩具项目部署上线，从此掉进前端动效的坑' },
];

const stack = [
  { name: 'Vue 3', role: '前端框架' },
  { name: 'Vite 7', role: '构建工具' },
  { name: 'Express', role: 'API 服务' },
  { name: 'SQLite', role: '数据存储' },
  { name: 'Markdown', role: '内容规范' },
];

const links = [
  { name: 'GitHub', url: 'https://github.com/BeaconCat', icon: 'github' },
  { name: 'Email', url: 'mailto:hi@example.com', icon: 'mail' },
  { name: 'RSS', url: '/feed', icon: 'rss' },
];
</script>

<template>
  <main class="page">
    <!-- 身份 Hero -->
    <section v-reveal class="hero">
      <div class="hero-avatar">
        <img src="/favicon-256.png" alt="BeaconCat" draggable="false" />
        <span class="ring" aria-hidden="true" />
      </div>
      <h1 class="hero-name">{{ config.cfg.about.name }}</h1>
      <p class="hero-line">{{ config.cfg.about.tagline }}</p>
      <p class="hero-bio">{{ config.cfg.about.bio }}</p>
      <div class="hero-dots" aria-hidden="true">
        <span style="--c: #ff0032" /><span style="--c: #ffb300" /><span style="--c: #0078ff" />
      </div>
    </section>

    <!-- 技能 -->
    <section v-reveal class="block">
      <h2 class="block-title">技能</h2>
      <div class="skills">
        <div v-for="group in skillGroups" :key="group.title" class="skill-group">
          <h3>{{ group.title }}</h3>
          <div class="chips">
            <span v-for="item in group.items" :key="item">{{ item }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- 历程 -->
    <section v-reveal class="block">
      <h2 class="block-title">历程</h2>
      <div class="milestones">
        <div v-for="m in milestones" :key="m.year" class="milestone">
          <span class="year">{{ m.year }}</span>
          <span class="dot" aria-hidden="true" />
          <p>{{ m.text }}</p>
        </div>
      </div>
    </section>

    <!-- 本站 -->
    <section v-reveal class="block">
      <h2 class="block-title">本站</h2>
      <div class="stack">
        <div v-for="s in stack" :key="s.name" class="stack-item">
          <strong>{{ s.name }}</strong>
          <span>{{ s.role }}</span>
        </div>
      </div>
    </section>

    <!-- 联系 -->
    <section v-reveal class="block">
      <h2 class="block-title">找到我</h2>
      <div class="links">
        <a v-for="l in links" :key="l.name" :href="l.url" class="link" target="_blank" rel="noopener">
          <svg v-if="l.icon === 'github'" viewBox="0 0 16 16" fill="currentColor">
            <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8Z" />
          </svg>
          <svg v-else-if="l.icon === 'mail'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="5" width="18" height="14" rx="2" />
            <path d="m3 7 9 6 9-6" />
          </svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round">
            <path d="M4 11a9 9 0 0 1 9 9M4 4a16 16 0 0 1 16 16" />
            <circle cx="5" cy="19" r="1.6" fill="currentColor" stroke="none" />
          </svg>
          <span>{{ l.name }}</span>
        </a>
      </div>
    </section>

    <!-- Powered by -->
    <a
      v-reveal
      class="powered"
      href="https://github.com/BeaconCat/Myself"
      target="_blank"
      rel="noopener"
    >
      <!-- 氛围光 -->
      <span class="pw-glow g-red" aria-hidden="true" />
      <span class="pw-glow g-yellow" aria-hidden="true" />
      <span class="pw-glow g-blue" aria-hidden="true" />
      <!-- 流光扫过 -->
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
  max-width: 820px;
  margin: 0 auto;
  padding: 120px 24px 90px;
}

/* ===== 身份 Hero ===== */
.hero {
  text-align: center;
  margin-bottom: 72px;
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
  }

  /* 呼吸光环 */
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

/* ===== 区块 ===== */
.block {
  margin-bottom: 56px;
}

.block-title {
  font-size: 24px;
  margin-bottom: 22px;
  padding-left: 14px;
  border-left: 4px solid var(--primary);
}

/* 技能 */
.skills {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.skill-group {
  padding: 20px;
  background: var(--surface);
  border: 1px solid var(--border);

  h3 {
    font-size: 15px;
    margin-bottom: 12px;
    color: var(--primary);
  }
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;

  span {
    font-size: 12px;
    font-weight: 600;
    padding: 4px 12px;
    background: var(--surface-2);
    transition: all var(--dur-fast) var(--ease-spring);

    &:hover {
      transform: scale(1.08);
      background: rgba(var(--primary-rgb), 0.12);
      color: var(--primary);
    }
  }
}

/* 历程 */
.milestones {
  display: flex;
  flex-direction: column;
}

.milestone {
  display: flex;
  align-items: baseline;
  gap: 18px;
  padding: 16px 4px;
  border-bottom: 1px solid var(--border);
  transition: background var(--dur-fast);

  &:hover {
    background: rgba(var(--primary-rgb), 0.04);

    .dot { transform: scale(1.4); box-shadow: 0 0 10px rgba(var(--primary-rgb), 0.6); }
  }

  .year {
    font-family: var(--font-serif);
    font-size: 20px;
    font-weight: 700;
    color: var(--primary);
    flex-shrink: 0;
    width: 64px;
  }

  .dot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: var(--primary);
    flex-shrink: 0;
    align-self: center;
    transition: all var(--dur-fast) var(--ease-spring);
  }

  p {
    font-size: 14.5px;
    line-height: 1.8;
    color: var(--text-2);
  }
}

/* 本站技术栈 */
.stack {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
}

.stack-item {
  text-align: center;
  padding: 18px 10px;
  background: var(--surface);
  border: 1px solid var(--border);
  transition: transform var(--dur-fast) var(--ease-out), border-color var(--dur-fast);

  strong {
    display: block;
    font-family: var(--font-serif);
    font-size: 16px;
  }

  span {
    display: block;
    margin-top: 6px;
    font-size: 12px;
    color: var(--text-2);
  }

  &:hover {
    transform: scale(1.05);
    border-color: var(--primary);
  }
}

/* 联系 */
.links {
  display: flex;
  gap: 14px;
  flex-wrap: wrap;
}

.link {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 12px 24px;
  background: var(--surface);
  border: 1px solid var(--border);
  font-size: 14px;
  font-weight: 600;
  transition: all var(--dur-fast) var(--ease-out);

  svg { width: 18px; height: 18px; }

  &:hover {
    transform: scale(1.05);
    border-color: var(--primary);
    color: var(--primary);
    box-shadow: 0 6px 20px -6px rgba(var(--primary-rgb), 0.4);
  }
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
  /* 主题自适应：表面底色 + 主色氛围 */
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

/* 主色系氛围光斑（深浅两侧渐次） */
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

/* 悬停流光扫过 */
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
  /* 底部留白防 y 降部裁切；负上边距贴近 Powered by */
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

  .skills { grid-template-columns: 1fr; }
  .stack { grid-template-columns: repeat(2, 1fr); }

  .powered { padding: 26px 22px; gap: 16px; }
  .pw-name { font-size: 24px; }
}
</style>
