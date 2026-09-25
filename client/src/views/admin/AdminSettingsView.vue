<script setup lang="ts">
import { settle, stableJson } from './studio/state';
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';
import { FALLBACK_CONFIG, useConfigStore, type SiteConfig } from '../../stores/config';
import { useDialogStore } from '../../stores/dialog';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import { toast } from './studio/toast';
import SectionGithub from './settings/SectionGithub.vue';
import SectionAccount from './settings/SectionAccount.vue';

/** 设置：站点 / 加载文案 / 随想与封面 / 时区 / GitHub / 账号。左侧锚点导航随滚动高亮 */
const { t } = useI18n();
const config = useConfigStore();
const dialog = useDialogStore();

const cfg = reactive<SiteConfig>(JSON.parse(JSON.stringify(config.cfg)));
const snapshot = ref('');
const loaded = ref(false);
const busy = ref(false);

/** 只比较本页负责的字段 */
const mine = () => stableJson([cfg.site, cfg.loading, cfg.thoughts, cfg.covers, cfg.timezone, cfg.github]);
const dirty = computed(() => loaded.value && mine() !== snapshot.value);

const SECTIONS = ['site', 'loading', 'content', 'timezone', 'github', 'account'] as const;
const current = ref<string>('site');

const TIMEZONES = [
  'Asia/Shanghai', 'Asia/Hong_Kong', 'Asia/Taipei', 'Asia/Tokyo', 'Asia/Singapore',
  'UTC', 'Europe/London', 'Europe/Berlin', 'America/New_York', 'America/Los_Angeles',
];

async function load(): Promise<void> {
  try {
    const remote = (await adminApi.settings()) as unknown as SiteConfig;
    Object.assign(cfg, JSON.parse(JSON.stringify(remote)));
    cfg.github = { ...FALLBACK_CONFIG.github, ...(remote.github ?? {}) };
    cfg.github.stats = { ...FALLBACK_CONFIG.github.stats, ...(remote.github?.stats ?? {}) };
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
  await settle();
  snapshot.value = mine();
  loaded.value = true;
}

async function save(): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  try {
    await adminApi.saveSettings(JSON.parse(JSON.stringify({
      site: cfg.site,
      loading: cfg.loading,
      thoughts: cfg.thoughts,
      covers: cfg.covers,
      timezone: cfg.timezone,
      github: cfg.github,
    })));
    snapshot.value = mine();
    await config.load();
    toast(t('studio.settings.saved'));
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  } finally {
    busy.value = false;
  }
}

const expandSec = computed({
  get: () => (cfg.covers.expandMs / 1000).toFixed(1).replace(/\.0$/, ''),
  set: (v: string) => (cfg.covers.expandMs = Math.max(1500, Math.round(Number(v) * 1000) || 10000)),
});

/* ===== 锚点导航 ===== */
const root = ref<HTMLElement | null>(null);
let io: IntersectionObserver | null = null;

function jump(id: string): void {
  current.value = id;
  root.value?.querySelector(`#set-${id}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

onBeforeRouteLeave(async () => {
  if (!dirty.value) return true;
  return dialog.confirm({
    title: t('studio.settings.leaveTitle'),
    message: t('studio.appearance.leaveBody'),
    confirmText: t('studio.write.leave'),
    danger: true,
  });
});

function onKey(e: KeyboardEvent): void {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
    e.preventDefault();
    void save();
  }
}

onMounted(() => {
  void load();
  window.addEventListener('keydown', onKey);
  io = new IntersectionObserver(
    (entries) => {
      const hit = entries.filter((e) => e.isIntersecting).sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)[0];
      if (hit) current.value = hit.target.id.replace('set-', '');
    },
    { rootMargin: '-20% 0px -60% 0px' },
  );
  void nextTick(() => root.value?.querySelectorAll('.sec').forEach((el) => io?.observe(el)));
});
onBeforeUnmount(() => {
  io?.disconnect();
  window.removeEventListener('keydown', onKey);
});
</script>

<template>
  <section ref="root" class="studio view">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.settings.title') }}</h1>
        <p>{{ t('studio.settings.desc') }}</p>
      </div>
      <div class="act">
        <button type="button" class="st-btn" :class="dirty ? 'p' : 'g'" :disabled="busy || !dirty" @click="save">
          <SIcon name="check" :size="16" />{{ dirty ? t('studio.save') : t('studio.saved') }}
        </button>
      </div>
    </div>

    <div class="layout">
      <nav class="sub">
        <button v-for="s in SECTIONS" :key="s" type="button" :class="{ on: current === s }" @click="jump(s)">
          {{ t(`studio.settings.nav.${s}`) }}
        </button>
      </nav>

      <div class="secs">
        <section id="set-site" class="sec st-rise" style="--i: 0">
          <h2>{{ t('studio.settings.nav.site') }}</h2>
          <div class="st-opt">
            <div>{{ t('studio.settings.siteTitle') }}<small>{{ t('studio.settings.siteTitleSub') }}</small></div>
            <label class="st-field w320"><input v-model="cfg.site.title" /></label>
          </div>
          <div class="st-opt">
            <div>{{ t('studio.settings.siteSubtitle') }}<small>{{ t('studio.settings.siteSubtitleSub') }}</small></div>
            <label class="st-field w320"><input v-model="cfg.site.subtitle" /></label>
          </div>
          <div class="st-opt">
            <div>{{ t('studio.settings.listEnd') }}<small>{{ t('studio.settings.listEndSub') }}</small></div>
            <label class="st-field w320"><input v-model="cfg.site.listEndText" /></label>
          </div>
        </section>

        <section id="set-loading" class="sec st-rise" style="--i: 1">
          <h2>{{ t('studio.settings.nav.loading') }}</h2>
          <div class="st-opt">
            <div>{{ t('studio.settings.bootText') }}<small>{{ t('studio.settings.bootTextSub') }}</small></div>
            <label class="st-field w320"><input v-model="cfg.loading.bootText" /></label>
          </div>
          <div class="st-opt">
            <div>{{ t('studio.settings.routeText') }}<small>{{ t('studio.settings.routeTextSub') }}</small></div>
            <label class="st-field w320"><input v-model="cfg.loading.routeText" /></label>
          </div>
        </section>

        <section id="set-content" class="sec st-rise" style="--i: 2">
          <h2>{{ t('studio.settings.nav.content') }}</h2>
          <div class="st-opt">
            <div>{{ t('studio.settings.thoughtsSub') }}<small>{{ t('studio.settings.thoughtsSubSub') }}</small></div>
            <label class="st-field w420"><input v-model="cfg.thoughts.subtitle" /></label>
          </div>
          <div class="st-opt">
            <div>{{ t('studio.settings.coverMs') }}<small>{{ t('studio.settings.coverMsSub') }}</small></div>
            <label class="st-field w140"><input v-model="expandSec" type="number" min="1.5" step="0.5" /><span class="suffix">{{ t('studio.settings.seconds') }}</span></label>
          </div>
        </section>

        <section id="set-timezone" class="sec st-rise" style="--i: 3">
          <h2>{{ t('studio.settings.nav.timezone') }}</h2>
          <div class="st-opt">
            <div>{{ t('studio.settings.tz') }}<small>{{ t('studio.settings.tzSub') }}</small></div>
            <label class="st-field w260 sel">
              <SIcon name="globe" :size="16" />
              <select v-model="cfg.timezone"><option v-for="tz in TIMEZONES" :key="tz" :value="tz">{{ tz }}</option></select>
              <SIcon name="chevronD" :size="14" />
            </label>
          </div>
        </section>

        <section id="set-github" class="sec st-rise" style="--i: 4">
          <h2>{{ t('studio.settings.nav.github') }}</h2>
          <SectionGithub :cfg="cfg" />
        </section>

        <section id="set-account" class="sec st-rise" style="--i: 5">
          <h2>{{ t('studio.settings.nav.account') }}</h2>
          <p class="sec-desc">{{ t('studio.settings.accountDesc') }}</p>
          <SectionAccount />
        </section>
      </div>
    </div>
  </section>
</template>

<style scoped lang="scss">
.view {
  max-width: 1120px;
  margin: 0 auto;
  padding: 52px 64px 96px;
}

.layout {
  display: grid;
  grid-template-columns: 168px minmax(0, 1fr);
  gap: 48px;
  align-items: start;
}

.sub {
  position: sticky;
  top: 24px;
  display: flex;
  flex-direction: column;
  gap: 2px;

  button {
    position: relative;
    text-align: left;
    height: 34px;
    padding: 0 14px;
    border-radius: var(--r-sm);
    font-size: 13.5px;
    color: var(--st-ink-3);
    transition: all var(--dur-fast);

    &:hover:not(.on) { color: var(--st-ink); background: var(--hover); }
    /* 当前分区：抬升 + 轻染，不挂信号条 */
    &.on { color: var(--lift-fg); font-weight: 500; background: var(--lift); box-shadow: var(--lift-shadow); }
  }
}

.sec {
  scroll-margin-top: 24px;
  padding-bottom: 36px;
  margin-bottom: 36px;
  border-bottom: 1px solid var(--line);

  &:last-child { border-bottom: 0; }

  h2 { font: 600 18px/1.3 var(--font-serif); margin: 0 0 6px; }
  .sec-desc { font-size: 13px; color: var(--st-ink-3); margin: 0 0 14px; }
}

.w140 { width: 140px; }
.w260 { width: 260px; }
.w320 { width: 320px; }
.w420 { width: 420px; }

.sel {
  position: relative;

  select { padding-right: 4px; }
}

@media (max-width: 1180px) {
  .view { padding: 40px 36px 80px; }
  .layout { grid-template-columns: 1fr; }
  .sub { display: none; }
}
</style>
