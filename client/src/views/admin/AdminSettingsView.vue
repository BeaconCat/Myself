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
import StSwitch from './studio/StSwitch.vue';
import { toast } from './studio/toast';
import SectionGithub from './settings/SectionGithub.vue';
import SectionAccount from './settings/SectionAccount.vue';
import SectionMail from './settings/SectionMail.vue';
import SectionLogin from './settings/SectionLogin.vue';
import SectionUpdates from './settings/SectionUpdates.vue';

/** 设置：站点 / 加载文案 / 随想与封面 / 时区 / GitHub / 邮件 / 登录方式 / 账号。左侧锚点导航随滚动高亮 */
const { t } = useI18n();
const config = useConfigStore();
const dialog = useDialogStore();

const cfg = reactive<SiteConfig>(JSON.parse(JSON.stringify(config.cfg)));
const snapshot = ref('');
const loaded = ref(false);
const busy = ref(false);

/** 只比较本页负责的字段 */
const mine = () => stableJson([cfg.site, cfg.loading, cfg.thoughts, cfg.covers, cfg.timezone, cfg.github, cfg.mail, cfg.oauth, cfg.users?.login, cfg.session]);
const dirty = computed(() => loaded.value && mine() !== snapshot.value);

const SECTIONS = ['site', 'loading', 'content', 'timezone', 'github', 'mail', 'login', 'account', 'system'] as const;
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
    cfg.thoughts = { ...FALLBACK_CONFIG.thoughts, ...(remote.thoughts ?? {}) };
    cfg.site.url ??= '';
    cfg.mail = { enabled: false, host: '', port: 587, username: '', password: '', from: '', security: 'starttls', ...(remote.mail ?? {}) };
    cfg.oauth = { github: { clientId: '', clientSecret: '', ...(remote.oauth?.github ?? {}) } };
    cfg.session = { duration: '7d', ...(remote.session ?? {}) };
    cfg.users = {
      ...(remote.users ?? FALLBACK_CONFIG.users!),
      login: { github: false, ...(remote.users?.login ?? {}) },
    };
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
      mail: cfg.mail,
      oauth: cfg.oauth,
      session: cfg.session,
      users: { login: { github: !!cfg.users?.login.github } },
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
let anchorUntil = 0;

function jump(id: string): void {
  current.value = id;
  anchorUntil = performance.now() + 1200;
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
      if (performance.now() < anchorUntil) return;
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
          <SIcon name="check" :size="18" />{{ dirty ? t('studio.save') : t('studio.saved') }}
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
        <section id="set-site" class="sec st-card wide st-rise" style="--i: 0">
          <h2>{{ t('studio.settings.nav.site') }}</h2>
          <div class="st-opt">
            <div>{{ t('studio.settings.siteTitle') }}<small>{{ t('studio.settings.siteTitleSub') }}</small></div>
            <label class="st-field w320"><input v-model="cfg.site.title" :aria-label="t('studio.settings.siteTitle')" /></label>
          </div>
          <div class="st-opt">
            <div>{{ t('studio.settings.siteSubtitle') }}<small>{{ t('studio.settings.siteSubtitleSub') }}</small></div>
            <label class="st-field w320"><input v-model="cfg.site.subtitle" :aria-label="t('studio.settings.siteSubtitle')" /></label>
          </div>
          <div class="st-opt">
            <div>{{ t('studio.settings.listEnd') }}<small>{{ t('studio.settings.listEndSub') }}</small></div>
            <label class="st-field w320"><input v-model="cfg.site.listEndText" :aria-label="t('studio.settings.listEnd')" /></label>
          </div>
          <div class="st-opt">
            <div>{{ t('studio.settings.siteUrl') }}<small>{{ t('studio.settings.siteUrlSub') }}</small></div>
            <label class="st-field w320"><input v-model.trim="cfg.site.url" :aria-label="t('studio.settings.siteUrl')" placeholder="https://example.com" spellcheck="false" /></label>
          </div>
        </section>

        <section id="set-loading" class="sec st-card half st-rise" style="--i: 1">
          <h2>{{ t('studio.settings.nav.loading') }}</h2>
          <div class="st-opt">
            <div>{{ t('studio.settings.bootText') }}<small>{{ t('studio.settings.bootTextSub') }}</small></div>
            <label class="st-field w320"><input v-model="cfg.loading.bootText" :aria-label="t('studio.settings.bootText')" /></label>
          </div>
          <div class="st-opt">
            <div>{{ t('studio.settings.routeText') }}<small>{{ t('studio.settings.routeTextSub') }}</small></div>
            <label class="st-field w320"><input v-model="cfg.loading.routeText" :aria-label="t('studio.settings.routeText')" /></label>
          </div>
        </section>

        <section id="set-content" class="sec st-card half st-rise" style="--i: 2">
          <h2>{{ t('studio.settings.nav.content') }}</h2>
          <div class="st-opt">
            <div>{{ t('studio.settings.thoughtsSub') }}<small>{{ t('studio.settings.thoughtsSubSub') }}</small></div>
            <label class="st-field w420"><input v-model="cfg.thoughts.subtitle" :aria-label="t('studio.settings.thoughtsSub')" /></label>
          </div>
          <div class="st-opt toggle-opt">
            <div>{{ t('studio.settings.thoughtsShowAlias') }}<small>{{ t('studio.settings.thoughtsShowAliasSub') }}</small></div>
            <StSwitch v-model="cfg.thoughts.showAlias" :label="t('studio.settings.thoughtsShowAlias')" />
          </div>
          <div class="st-opt toggle-opt">
            <div>{{ t('studio.settings.thoughtsShowUsername') }}<small>{{ t('studio.settings.thoughtsShowUsernameSub') }}</small></div>
            <StSwitch v-model="cfg.thoughts.showUsername" :label="t('studio.settings.thoughtsShowUsername')" />
          </div>
          <div class="st-opt">
            <div>{{ t('studio.settings.coverMs') }}<small>{{ t('studio.settings.coverMsSub') }}</small></div>
            <label class="st-field w140"><input v-model="expandSec" :aria-label="t('studio.settings.coverMs')" type="number" min="1.5" step="0.5" /><span class="suffix">{{ t('studio.settings.seconds') }}</span></label>
          </div>
        </section>

        <section id="set-timezone" class="sec st-card wide st-rise" style="--i: 3">
          <h2>{{ t('studio.settings.nav.timezone') }}</h2>
          <div class="st-opt">
            <div>{{ t('studio.settings.tz') }}<small>{{ t('studio.settings.tzSub') }}</small></div>
            <label class="st-field w260 sel">
              <SIcon name="globe" :size="18" />
              <select v-model="cfg.timezone" :aria-label="t('studio.settings.tz')"><option v-for="tz in TIMEZONES" :key="tz" :value="tz">{{ tz }}</option></select>
              <SIcon name="chevronD" :size="16" />
            </label>
          </div>
        </section>

        <section id="set-github" class="sec st-card wide st-rise" style="--i: 4">
          <h2>{{ t('studio.settings.nav.github') }}</h2>
          <SectionGithub :cfg="cfg" />
        </section>

        <section id="set-mail" class="sec st-card wide st-rise" style="--i: 5">
          <h2>{{ t('studio.settings.nav.mail') }}</h2>
          <p class="sec-desc">{{ t('studio.settings.mailDesc') }}</p>
          <SectionMail v-if="cfg.mail" :mail="cfg.mail" :dirty="dirty" />
        </section>

        <section id="set-login" class="sec st-card wide st-rise" style="--i: 6">
          <h2>{{ t('studio.settings.nav.login') }}</h2>
          <p class="sec-desc">{{ t('studio.settings.loginDesc') }}</p>
          <SectionLogin v-if="cfg.oauth && cfg.users" :cfg="cfg" />
        </section>

        <section id="set-account" class="sec st-card wide st-rise" style="--i: 7">
          <h2>{{ t('studio.settings.nav.account') }}</h2>
          <p class="sec-desc">{{ t('studio.settings.accountDesc') }}</p>
          <SectionAccount />
        </section>
        <section id="set-system" class="sec st-card wide st-rise" style="--i: 8">
          <h2>{{ t('studio.settings.nav.system') }}</h2>
          <SectionUpdates :dirty="dirty" />
        </section>
      </div>
    </div>
  </section>
</template>

<style scoped lang="scss">
.view {
  max-width: 1280px;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

.layout {
  display: grid;
  grid-template-columns: 148px minmax(0, 1fr);
  gap: 28px;
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
    height: 38px;
    padding: 0 14px;
    border-radius: var(--r-sm);
    font-size: 14.5px;
    color: var(--st-ink-3);
    transition: all var(--dur-fast);

    &:hover:not(.on) { color: var(--st-ink); background: var(--hover); }
    /* 当前分区：抬升 + 轻染，不挂信号条 */
    &.on { color: var(--lift-fg); font-weight: 500; background: var(--lift); box-shadow: var(--lift-shadow); }
  }
}

/* 分区 = 卡片；短分区两两并排（6/6）等高，长分区整行 */
.secs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-items: stretch;
  gap: 18px;
}

.sec {
  scroll-margin-top: 24px;
  gap: 6px;

  &.wide { grid-column: 1 / -1; }

  h2 { font: 700 22px/1.3 var(--font-serif); margin: 0 0 4px; }
  .sec-desc { font-size: 14px; color: var(--st-ink-3); margin: 0 0 10px; }
}

/* 半宽卡：说明在上、输入框撑满 */
.half {
  :deep(.st-opt), .st-opt {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }

  .st-field { width: auto; }
}

.half .toggle-opt { flex-direction: row; align-items: center; }

.w140 { width: 140px; }
.w260 { width: 260px; }
.w320 { width: 320px; }
.w420 { width: 420px; }

.sel {
  position: relative;

  select { padding-right: 4px; }
}

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
  .layout { grid-template-columns: 1fr; }
  .sub { display: none; }
  .secs { grid-template-columns: 1fr; }
}

@media (max-width: 767px) {
  .secs { min-width: 0; }
  .half .st-opt { flex-direction: row; align-items: center; }
  .sec { scroll-margin-top: calc(var(--safe-t, 0px) + var(--nav-row, 48px) + 12px); }
  .sec :deep(.st-opt) { flex-wrap: wrap; gap: 10px; }
  .sec :deep(.st-opt > div:first-child) { flex: 1 1 160px; }
  .sec :deep(.st-opt > .st-field) { flex: 1 1 100%; width: 100%; }
  .sec :deep(.st-opt > .seg) { flex-basis: 100%; }
  .sec :deep(.st-flabel) { flex-wrap: wrap; gap: 4px 8px; }
}
</style>
