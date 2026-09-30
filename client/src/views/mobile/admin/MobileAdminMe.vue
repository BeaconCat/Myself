<script setup lang="ts">
/**
 * 后台 · 我的：身份卡、外观（深浅 + 色盘，圆形扩散切换，仅本机）、站点管理入口（回落桌面组件）、
 * API Key 只读列表、备份立即执行、查看站点、退出登录。挂在 admin-settings 路由。
 */
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, type ApiKeyInfo, type BackupInfo } from '../../../api';
import { useAuthStore } from '../../../stores/auth';
import { useConfigStore } from '../../../stores/config';
import { useDialogStore } from '../../../stores/dialog';
import { useThemeStore } from '../../../stores/theme';
import { circularReveal } from '../../../utils/circularReveal';
import MaPage from '../../../components/mobile-admin/MaPage.vue';
import MaIcon from '../../../components/mobile-admin/MaIcon.vue';
import MaRing from '../../../components/mobile-admin/MaRing.vue';
import MaSkeleton from '../../../components/mobile-admin/MaSkeleton.vue';
import type { IconName } from '../../../components/mobile-admin/icons';
import { toast } from '../../../components/mobile-admin/state';
import { formatSize, relTime } from '../../../components/mobile-admin/format';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const config = useConfigStore();
const dialog = useDialogStore();
const theme = useThemeStore();

const keys = ref<ApiKeyInfo[] | null>(null);
const backups = ref<BackupInfo[] | null>(null);
const backingUp = ref(false);
const backupEl = ref<HTMLElement | null>(null);

function load(): Promise<unknown> {
  return Promise.allSettled([
    adminApi.apiKeys().then((k) => (keys.value = k)),
    adminApi.backups().then((b) => (backups.value = b)),
  ]);
}
onMounted(async () => {
  await load();
  if (route.query.focus === 'backup') {
    backupEl.value?.scrollIntoView({ behavior: 'smooth', block: 'center' });
    backupEl.value?.classList.add('flash');
  }
});
function refresh(done: () => void): void {
  void load().finally(done);
}

const avatar = computed(() => config.cfg.about.avatar || config.cfg.site.logo || '/favicon-256.png');
const name = computed(() => config.cfg.about.name || 'Myself');

/* ---------- 外观：仅本机生效 ---------- */
function origin(e: MouseEvent): { x: number; y: number } {
  const r = (e.currentTarget as HTMLElement).getBoundingClientRect();
  return { x: r.left + r.width / 2, y: r.top + r.height / 2 };
}
function setMode(e: MouseEvent, mode: 'light' | 'dark'): void {
  if (theme.mode === mode) return;
  circularReveal(origin(e), () => theme.setMode(mode), mode === 'light' ? 'expand' : 'contract');
}
function setPalette(e: MouseEvent, id: string, label: string): void {
  if (theme.paletteId === id) return;
  circularReveal(origin(e), () => theme.setPalette(id, true), 'expand', { edge: true });
  window.setTimeout(() => toast(label, theme.mode === 'dark' ? t('mobileAdmin.me.dark') : t('mobileAdmin.me.light')), 380);
}

/* ---------- 站点管理入口（桌面组件回落） ---------- */
interface Entry {
  name: string;
  icon: IconName;
  color: string;
  query?: Record<string, string>;
}
const siteEntries: Entry[] = [
  { name: 'admin-settings', icon: 'gear', color: '#0078ff', query: { full: '1' } },
  { name: 'admin-appearance', icon: 'palette', color: '#ffb300' },
  { name: 'admin-identity', icon: 'user', color: '#7a5af8' },
  { name: 'admin-about', icon: 'grid', color: '#12b76a' },
  { name: 'admin-comments', icon: 'comment', color: '#ff7a1a' },
  { name: 'admin-users', icon: 'users', color: '#5b6b86' },
  { name: 'admin-data', icon: 'database', color: '#0b8aa8' },
];
function openEntry(e: Entry): void {
  void router.push({ name: e.name, query: e.query });
}

/* ---------- 备份 ---------- */
async function backupNow(): Promise<void> {
  if (backingUp.value) return;
  backingUp.value = true;
  try {
    const res = await adminApi.createBackup();
    backups.value = await adminApi.backups();
    const b = backups.value.find((x) => x.name === res.name);
    toast(t('mobileAdmin.me.backupDone'), b ? formatSize(b.size) : res.name);
  } catch {
    toast(t('mobileAdmin.me.backupFailed'), '', 'error');
  } finally {
    backingUp.value = false;
  }
}

/* ---------- 其它 ---------- */
function viewSite(): void {
  window.open('/', '_blank', 'noopener');
}
async function logout(): Promise<void> {
  const ok = await dialog.confirm({
    title: t('mobileAdmin.me.logoutTitle'),
    message: t('mobileAdmin.me.logoutMsg'),
    confirmText: t('mobileAdmin.me.logout'),
    cancelText: t('mobileAdmin.common.cancel'),
    danger: true,
  });
  if (!ok) return;
  await auth.logout();
  void router.replace('/admin/login');
}
</script>

<template>
  <MaPage :title="t('mobileAdmin.me.title')" @refresh="refresh">
    <!-- 高密度：身份与外观合并为一张卡（左身份 / 右操作，下方外观设置） -->
    <div class="me-card">
      <div class="me-head">
        <img class="av" :src="avatar" alt="" draggable="false" />
        <div class="who">
          <b>{{ name }}</b>
          <small>{{ t('mobileAdmin.me.role') }}</small>
        </div>
        <button class="icbtn tap" :aria-label="t('mobileAdmin.me.viewSite')" @click="viewSite"><MaIcon name="globe" :size="19" /></button>
      </div>
      <div class="appear">
        <div class="ap-h"><span>{{ t('mobileAdmin.me.appearance') }}</span><small>{{ t('mobileAdmin.me.localOnly') }}</small></div>
        <div class="seg2" :style="{ '--i': theme.mode === 'dark' ? 1 : 0 }">
          <span class="th" />
          <button class="tap" :class="{ on: theme.mode === 'light' }" @click="setMode($event, 'light')"><MaIcon name="sun" :size="18" />{{ t('mobileAdmin.me.light') }}</button>
          <button class="tap" :class="{ on: theme.mode === 'dark' }" @click="setMode($event, 'dark')"><MaIcon name="moon" :size="18" />{{ t('mobileAdmin.me.dark') }}</button>
        </div>
        <div class="pals">
          <button
            v-for="p in theme.allPalettes"
            :key="p.id"
            class="pal tap"
            :class="{ on: theme.paletteId === p.id }"
            :style="{ '--c': p[theme.mode].primary }"
            @click="setPalette($event, p.id, p.nameKey)"
          >
            <i />
            <span>{{ p.nameKey.split('·')[0].trim() }}</span>
          </button>
        </div>
      </div>
    </div>

    <div class="sec-h"><h2>{{ t('mobileAdmin.me.site') }}</h2></div>
    <div class="ma-list grp">
      <button v-for="e in siteEntries" :key="e.name" class="ma-li tap" @click="openEntry(e)">
        <span class="lic" :style="{ '--c': e.color }"><MaIcon :name="e.icon" :size="17" /></span>
        <span class="lt">{{ t(`mobileAdmin.route.${e.name}`) }}</span>
        <small>{{ t(`mobileAdmin.me.hint.${e.name}`) }}</small>
        <MaIcon name="chev" :size="16" class="chev" />
      </button>
    </div>

    <div class="sec-h">
      <h2>{{ t('mobileAdmin.me.apiKeys') }}</h2>
      <button class="more tap" @click="router.push({ name: 'admin-apikeys' })">{{ t('mobileAdmin.me.manage') }}<MaIcon name="chev" :size="14" /></button>
    </div>
    <MaSkeleton v-if="!keys" variant="list" :count="2" />
    <div v-else class="ma-list grp">
      <div v-for="k in keys" :key="k.id" class="ma-li key">
        <span class="lic" style="--c: #12b76a"><MaIcon name="key" :size="17" /></span>
        <span class="lt">
          <span class="kn">{{ k.name }}</span>
          <code>{{ k.prefix }}••••</code>
        </span>
        <small>{{ k.lastUsedAt ? relTime(k.lastUsedAt, t) : t('mobileAdmin.me.neverUsed') }}</small>
      </div>
      <div v-if="!keys.length" class="ma-li muted">
        <span class="lic" style="--c: var(--text-3)"><MaIcon name="key" :size="17" /></span>
        <span class="lt">{{ t('mobileAdmin.me.noKeys') }}</span>
      </div>
    </div>

    <div class="sec-h"><h2>{{ t('mobileAdmin.me.backup') }}</h2></div>
    <div ref="backupEl" class="ma-list grp backup">
      <div class="ma-li">
        <span class="lic" style="--c: #5b6b86"><MaIcon name="archive" :size="17" /></span>
        <span class="lt">
          <span>{{ t('mobileAdmin.me.lastBackup') }}</span>
          <small v-if="backups">{{ backups[0] ? `${relTime(backups[0].createdAt, t)} · ${formatSize(backups[0].size)}` : t('mobileAdmin.me.noBackup') }}</small>
        </span>
        <button class="bk-btn tap" :disabled="backingUp" @click="backupNow">
          <MaRing v-if="backingUp" indeterminate :size="16" :stroke="2" class="bk-ring" />
          {{ backingUp ? t('mobileAdmin.me.backingUp') : t('mobileAdmin.me.backupNow') }}
        </button>
      </div>
      <div v-if="backups && backups.length" class="ma-li muted small">
        <span class="lt indent"><small>{{ t('mobileAdmin.me.backupCount', { n: backups.length, size: formatSize(backups.reduce((a, b) => a + b.size, 0)) }) }}</small></span>
      </div>
    </div>

    <div class="ma-list grp gap">
      <button class="ma-li danger tap" @click="logout">{{ t('mobileAdmin.me.logout') }}</button>
    </div>
    <p class="foot">{{ config.cfg.site.title }} · {{ config.cfg.site.subtitle }}</p>
  </MaPage>
</template>

<style scoped lang="scss">
.me-card {
  margin: 2px 16px 0;
  padding: 18px;
  border-radius: var(--r-lg);
  background: var(--elev);
  box-shadow: var(--shadow-card);
}

.me-head {
  display: flex;
  align-items: center;
  gap: 14px;

  .av {
    width: 60px;
    height: 60px;
    border-radius: var(--r-lg);
    object-fit: cover;
    box-shadow: 0 0 0 0.5px var(--line-2);
  }

  .who { flex: 1; min-width: 0; }

  b {
    display: block;
    font-family: var(--font-serif);
    font-size: 24px;
    line-height: 1.25;
  }

  small {
    font-size: 13px;
    color: var(--text-3);
  }
}

.appear {
  margin-top: 18px;
  padding-top: 16px;
  box-shadow: 0 -1px 0 var(--line);
}

.ap-h {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 15px;
  font-weight: 600;

  small {
    font-size: 13px;
    font-weight: 400;
    color: var(--text-3);
  }
}

.seg2 {
  position: relative;
  display: grid;
  grid-template-columns: 1fr 1fr;
  padding: 3px;
  border-radius: var(--r-md);
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line);

  button {
    position: relative;
    z-index: 1;
    height: 38px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 7px;
    font-size: 14px;
    color: var(--text-2);
    border-radius: calc(var(--r-md) - 3px);
    transition: color var(--dur) var(--ease-out);

    :deep(.ma-ic) { transition: color var(--dur); }

    &.on {
      color: var(--lift-fg);
      font-weight: 500;

      :deep(.ma-ic) { color: var(--ink); }
    }
  }

  .th {
    position: absolute;
    top: 3px;
    bottom: 3px;
    left: 3px;
    width: calc(50% - 3px);
    border-radius: calc(var(--r-md) - 3px);
    background: var(--lift);
    box-shadow: var(--lift-shadow);
    transform: translateX(calc(var(--i, 0) * 100%));
    transition: transform var(--dur) var(--ease-spring), background-color var(--dur), box-shadow var(--dur);
  }
}

.pals {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(58px, 1fr));
  gap: 10px 6px;
  margin-top: 14px;
}

.pal {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-3);

  i {
    position: relative;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: var(--c);
    box-shadow: inset 0 0 0 0.5px rgb(0 0 0 / 0.12);

    /* 选中：外侧 1.5px 细环（--text） */
    &::after {
      content: '';
      position: absolute;
      inset: -4.5px;
      border-radius: 50%;
      box-shadow: 0 0 0 1.5px var(--text);
      opacity: 0;
      transform: scale(0.8);
      transition: all var(--dur) var(--ease-spring);
    }
  }

  &.on {
    color: var(--text);
    font-weight: 500;

    i::after {
      opacity: 1;
      transform: none;
    }
  }
}

.grp {
  margin: 0 16px;
}

.gap {
  margin-top: 18px;
}

.key {
  .lt {
    display: flex;
    flex-direction: column;
    padding: 9px 0;
  }

  .kn { font-size: 15px; }

  code {
    font-family: var(--font-mono);
    font-size: 11.5px;
    color: var(--text-3);
    margin-top: 1px;
  }
}

.muted {
  color: var(--text-3);
}

.small { min-height: 40px; }

.indent { padding-left: 42px; }

.backup {
  transition: box-shadow 0.6s;

  &.flash { box-shadow: inset 0 0 0 1.5px var(--ink); }

  .lt {
    display: flex;
    flex-direction: column;
    padding: 10px 0;

    small { font-size: 12px; margin-top: 1px; }
  }
}

.bk-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 14px;
  border-radius: 999px;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--on-solid);
  background: var(--solid);
  box-shadow: var(--btn-shadow);

  &:disabled { opacity: 0.7; }

  .bk-ring {
    --ring-bg: color-mix(in oklab, var(--on-solid) 30%, transparent);
    --ring-fg: var(--on-solid);
  }
}

.ma-li.danger {
  justify-content: center;
  color: var(--accent-red);
  font-weight: 500;
}

.foot {
  margin-top: 22px;
  text-align: center;
  font-size: 12px;
  color: var(--text-3);
}
</style>
