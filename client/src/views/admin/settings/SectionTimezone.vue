<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { SiteConfig } from '../../../stores/config';

/** 站点时区 */
defineProps<{ cfg: SiteConfig }>();
const { t } = useI18n();

const TIMEZONES = [
  'Asia/Shanghai', 'Asia/Tokyo', 'Asia/Singapore', 'Asia/Hong_Kong',
  'UTC', 'Europe/London', 'Europe/Berlin', 'America/New_York', 'America/Los_Angeles',
];
</script>

<template>
  <section id="sec-timezone" class="card">
    <h2>{{ t('admin.secTimezone') }}</h2>
    <label class="narrow">
      <span>{{ t('admin.timezone') }}</span>
      <select v-model="cfg.timezone">
        <option v-for="tz in TIMEZONES" :key="tz" :value="tz">{{ tz }}</option>
      </select>
    </label>
  </section>
</template>

<style scoped lang="scss">
@use './settings-shared';
</style>
