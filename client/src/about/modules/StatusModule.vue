<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useConfigStore } from '../../stores/config';
import type { StatusData } from '../types';
import type { ModProps } from './props';
import ModHead from '../parts/ModHead.vue';
import { ago, pad2, useClock, zoned } from '../useClock';

/** 在线状态（status）：头像 + 状态环、当前活动（均衡器动效）、最后活跃与本地时间 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as StatusData);
const { t } = useI18n();
const now = useClock();
const config = useConfigStore();

const COLORS: Record<string, string> = { online: '#00c853', focus: '#ffb300', away: '#8390a6' };
const color = computed(() => COLORS[d.value.state] ?? COLORS.online);

/** 头像：取 profile 形象图或旧 about.avatar，最后回落 logo */
const avatar = computed(() => {
  const profile = config.cfg.about.modules?.find((m) => m.type === 'profile');
  return profile?.data?.portrait?.src || config.cfg.about.avatar || '/favicon-256.png';
});

const tz = computed(() => {
  const profile = config.cfg.about.modules?.find((m) => m.type === 'profile');
  return Number(profile?.data?.status?.tz ?? 8);
});
const clock = computed(() => zoned(now.value, tz.value));
const last = computed(() => ago(d.value.lastActive, t, now.value) || t('aboutKit.status.justNow'));
</script>

<template>
  <ModHead :title="title">
    <span class="ak-dot live" :style="{ background: color }" />{{ t(`aboutKit.status.${d.state}`) }}
  </ModHead>
  <div class="ss">
    <div class="ss-top">
      <div class="ss-wrap">
        <img class="ss-av" :src="avatar" alt="" draggable="false" />
        <span class="ss-ring" :style="{ '--sc': color }" />
      </div>
      <div>
        <div class="ss-state">{{ t(`aboutKit.status.${d.state}`) }}</div>
        <div class="ss-sub">{{ d.note || t(`aboutKit.status.${d.state}Sub`) }}</div>
      </div>
    </div>
    <div v-if="d.activity" class="ss-act">
      <small><span class="eq"><i /><i /><i /></span>{{ t('aboutKit.status.now') }}</small>
      {{ d.activity }}
      <span v-if="d.app" class="ss-app">{{ d.app }}</span>
    </div>
    <dl class="ss-meta">
      <div><dt>{{ t('aboutKit.status.lastActive') }}</dt><dd>{{ last }}</dd></div>
      <div><dt>{{ d.device ? t('aboutKit.status.device') : t('aboutKit.localTime') }}</dt><dd :class="{ 'ak-mono': !d.device }">{{ d.device || `${pad2(clock.h)}:${pad2(clock.m)}` }}</dd></div>
    </dl>
  </div>
</template>

<style scoped lang="scss">
/* 身份头像 + 状态 → 当前活动（下沉面）→ 统计条式元信息压到底部 */
.ss { display: flex; flex-direction: column; gap: 18px; flex: 1; }
.ss-top { display: flex; align-items: center; gap: 16px; }
.ss-wrap { position: relative; flex: none; }

.ss-av {
  display: block;
  width: 64px;
  height: 64px;
  border-radius: var(--r-lg);
  object-fit: cover;
  background: #050b17;
  box-shadow: inset 0 0 0 1px rgb(255 255 255 / 0.08);
}

.ss-ring {
  position: absolute;
  right: -5px;
  bottom: -5px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--sc);
  border: 3px solid var(--ak-surface-hi);
}

.ss-state { font: 700 22px/1.25 var(--font-serif); }
.ss-sub { margin-top: 4px; font-size: 14px; color: var(--text-2); }

.ss-act {
  padding: 14px 16px;
  border-radius: var(--r-md);
  background: var(--ak-sunken);
  font-size: 15px;
  font-weight: 500;
  line-height: 1.6;

  small {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 6px;
    font: 400 12.5px var(--font-sans);
    color: var(--ak-text-3);
  }

  .ss-app { display: block; margin-top: 2px; font: 400 12.5px var(--ak-mono); color: var(--ak-text-3); }
}

.eq {
  display: inline-flex;
  align-items: flex-end;
  gap: 2px;
  height: 10px;

  i { width: 2px; border-radius: var(--r-pill); background: var(--ink); animation: ss-eq 1s ease-in-out infinite; }
  i:nth-child(2) { animation-delay: -0.3s; }
  i:nth-child(3) { animation-delay: -0.6s; }
}

@keyframes ss-eq { 0%, 100% { height: 3px; } 50% { height: 10px; } }

.ss-meta {
  display: grid;
  grid-template-columns: 1fr 1fr;
  margin-top: auto;
  border-radius: var(--r-md);
  background: var(--fill);

  div { min-width: 0; padding: 14px 16px 12px; }
  div + div { box-shadow: -1px 0 0 var(--ak-line); }
  dt { font-size: 12.5px; color: var(--ak-text-3); }
  dd { margin-top: 4px; overflow: hidden; font-size: 16px; font-weight: 600; white-space: nowrap; text-overflow: ellipsis; }
  dd.ak-mono { font-size: 20px; }
}

@media (prefers-reduced-motion: reduce) { .eq i { animation: none; height: 7px; } }
</style>
