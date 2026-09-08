<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { SiteConfig } from '../../../stores/config';

/** 首页轮播规则 + 随想副标题 + 封面手风琴间隔 */
defineProps<{ cfg: SiteConfig }>();
const { t } = useI18n();
</script>

<template>
  <section id="sec-hero" class="card">
    <h2>{{ t('admin.secHero') }}</h2>
    <div class="row3">
      <label>
        <span>{{ t('admin.heroInterval') }}</span>
        <input v-model.number="cfg.hero.intervalMs" type="number" min="1000" step="500" />
      </label>
      <label>
        <span>{{ t('admin.heroCount') }}</span>
        <input v-model.number="cfg.hero.count" type="number" min="1" max="10" />
      </label>
      <label>
        <span>{{ t('admin.heroPinnedRule') }}</span>
        <select v-model="cfg.hero.pinnedRule">
          <option value="pinned-first">{{ t('admin.pinnedFirst') }}</option>
          <option value="ignore">{{ t('admin.pinnedIgnore') }}</option>
        </select>
      </label>
    </div>
    <label>
      <span>{{ t('admin.thoughtsSubtitle') }}</span>
      <input v-model="cfg.thoughts.subtitle" type="text" />
    </label>
    <label class="narrow" style="margin-top: 14px">
      <span>{{ t('admin.coverExpandMs') }}</span>
      <input v-model.number="cfg.covers.expandMs" type="number" min="1500" step="500" />
    </label>
  </section>
</template>

<style scoped lang="scss">
@use './settings-shared';
</style>
