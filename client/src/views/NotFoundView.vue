<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import { Compass } from 'lucide';
import Icon from '../components/ui/Icon.vue';
import { usePageTitle } from '../composables/usePageTitle';

/** 站内未知地址：说明页面不存在，给出回首页与看文章两个出口 */
const { t } = useI18n();
const route = useRoute();

usePageTitle(computed(() => t('notFound.title')));
</script>

<template>
  <main class="nf">
    <div class="in rise-stagger">
      <span class="ic"><Icon :icon="Compass" :size="28" /></span>
      <h1>{{ t('notFound.title') }}</h1>
      <p>{{ t('notFound.desc') }}</p>
      <code class="path">{{ route.fullPath }}</code>
      <div class="acts">
        <router-link to="/" class="b solid">{{ t('notFound.home') }}</router-link>
        <router-link to="/articles" class="b">{{ t('notFound.articles') }}</router-link>
      </div>
    </div>
  </main>
</template>

<style scoped lang="scss">
.nf {
  min-height: 70vh;
  display: grid;
  place-items: center;
  padding: 140px 16px 60px;
}

.in {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  max-width: 460px;
  text-align: center;
}

.ic {
  width: 64px;
  height: 64px;
  display: grid;
  place-items: center;
  border-radius: var(--r-lg);
  color: var(--primary);
  background: color-mix(in oklab, var(--primary) 10%, transparent);
}

h1 { margin: 6px 0 0; font: 600 28px var(--font-serif); color: var(--text); }
p { margin: 0; color: var(--text-2); line-height: 1.7; }

.path {
  max-width: 100%;
  padding: 4px 10px;
  border-radius: var(--r-sm);
  font: 12.5px var(--font-mono);
  color: var(--text-3);
  background: var(--fill);
  overflow-wrap: anywhere;
}

.acts { display: flex; gap: 10px; margin-top: 10px; flex-wrap: wrap; justify-content: center; }

.b {
  display: inline-flex;
  align-items: center;
  height: 40px;
  padding: 0 18px;
  border-radius: var(--r-pill);
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
  background: var(--elev);
  box-shadow: var(--shadow-card);
  text-decoration: none;
  transition: transform var(--dur-fast) var(--ease-out), background var(--dur-fast);

  &:hover { transform: translateY(-1px); }
  &.solid { color: var(--on-solid); background: var(--solid); box-shadow: var(--btn-shadow); }
  &.solid:hover { background: var(--solid-hover); }
}

@media (prefers-reduced-motion: reduce) { .b:hover { transform: none; } }
</style>
