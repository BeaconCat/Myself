<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

/** 品牌三色即站点身份 */
const accents = [
  { color: '#ff0032', key: 'about.red' },
  { color: '#ffb300', key: 'about.yellow' },
  { color: '#0078ff', key: 'about.blue' },
];
</script>

<template>
  <main class="page">
    <h1 v-reveal class="page-title">{{ t('nav.about') }}</h1>

    <section v-reveal class="card intro">
      <p>{{ t('about.intro') }}</p>
    </section>

    <div class="accent-grid">
      <div
        v-for="a in accents"
        :key="a.color"
        v-reveal
        class="card accent hover-lift"
        :style="{ '--a': a.color }"
      >
        <span class="swatch" />
        <code>{{ a.color }}</code>
        <p>{{ t(a.key) }}</p>
      </div>
    </div>
  </main>
</template>

<style scoped lang="scss">
.page {
  max-width: 760px;
  margin: 0 auto;
  padding: 110px 24px 80px;
}

.page-title {
  font-size: clamp(30px, 4vw, 42px);
  margin-bottom: 34px;
}

.card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 26px;
}

.intro p {
  line-height: 1.9;
  color: var(--text-2);
}

.accent-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-top: 16px;
}

.accent {
  text-align: center;

  .swatch {
    display: block;
    width: 44px;
    height: 44px;
    border-radius: 50%;
    background: var(--a);
    margin: 0 auto 14px;
    box-shadow: 0 0 18px color-mix(in srgb, var(--a) 55%, transparent);
  }

  code {
    font-family: Consolas, monospace;
    font-size: 13px;
    color: var(--text);
  }

  p {
    margin-top: 8px;
    font-size: 13px;
    color: var(--text-2);
  }
}

@media (max-width: 768px) {
  .page { padding-top: 88px; }
  .accent-grid { grid-template-columns: 1fr; }
}
</style>
