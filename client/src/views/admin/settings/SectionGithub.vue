<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../../api';
import type { SiteConfig } from '../../../stores/config';
import '../studio/i18n';
import SIcon from '../studio/SIcon.vue';
import StSeg from '../studio/StSeg.vue';
import StSwitch from '../studio/StSwitch.vue';
import { dateTimeText } from '../studio/format';

/** GitHub：账号 / 数据模式 / 手动数据 + 同步面板（立即同步、数据预览、同步日志） */
defineProps<{ cfg: SiteConfig }>();
const { t } = useI18n();

interface SyncLogEntry { at: string; ok: boolean; message: string }
const log = ref<SyncLogEntry[]>([]);
const preview = ref<{ stats?: Record<string, number>; fetchedAt?: string } | null>(null);
const syncing = ref(false);
const msg = ref<{ ok: boolean; text: string } | null>(null);

async function loadLog(): Promise<void> {
  try {
    const data = await adminApi.githubLog();
    log.value = data.log;
    preview.value = data.preview as typeof preview.value;
  } catch { /* 忽略 */ }
}

async function sync(): Promise<void> {
  if (syncing.value) return;
  syncing.value = true;
  msg.value = null;
  try {
    await adminApi.githubSync();
    msg.value = { ok: true, text: t('studio.settings.ghSyncOk') };
  } catch {
    msg.value = { ok: false, text: t('studio.settings.ghSyncFail') };
  } finally {
    syncing.value = false;
    await loadLog();
  }
}

onMounted(loadLog);
</script>

<template>
  <div class="gh">
    <div class="st-opt">
      <div>{{ t('studio.settings.ghUser') }}<small>{{ t('studio.settings.ghUserSub') }}</small></div>
      <label class="st-field w260"><span class="suffix">github.com/</span><input v-model="cfg.github.username" :aria-label="t('studio.settings.ghUser')" spellcheck="false" /></label>
    </div>
    <div class="st-opt">
      <div>{{ t('studio.settings.ghMode') }}<small>{{ cfg.github.mode === 'api' ? t('studio.settings.ghApiSub') : t('studio.settings.ghManualSub') }}</small></div>
      <StSeg
        v-model="cfg.github.mode"
        :options="[
          { value: 'api', label: t('studio.settings.ghApi') },
          { value: 'manual', label: t('studio.settings.ghManual') },
        ]"
      />
    </div>

    <template v-if="cfg.github.mode === 'api'">
      <div class="st-opt">
        <div>{{ t('studio.settings.ghToken') }}<small>{{ t(cfg.github.clearToken ? 'studio.settings.ghTokenClearing' : cfg.github.tokenConfigured ? 'studio.settings.ghTokenKeep' : 'studio.settings.ghTokenSub') }}</small></div>
        <div class="gh-token-control">
          <label class="st-field mono-in"><input v-model="cfg.github.token" :aria-label="t('studio.settings.ghToken')" type="text" class="token-input" name="github-stats-token" autocomplete="off" autocapitalize="off" spellcheck="false" :disabled="cfg.github.clearToken" :placeholder="t(cfg.github.tokenConfigured ? 'studio.settings.ghTokenKeep' : 'studio.settings.ghTokenSub')" /></label>
          <button v-if="cfg.github.tokenConfigured" type="button" class="st-btn g" @click="cfg.github.clearToken = !cfg.github.clearToken; cfg.github.token = ''">{{ t(cfg.github.clearToken ? 'studio.settings.ghTokenUndo' : 'studio.settings.ghTokenClear') }}</button>
        </div>
      </div>
      <div class="st-opt">
        <div>{{ t('studio.settings.ghRefresh') }}</div>
        <label class="st-field w140"><input v-model.number="cfg.github.refreshMinutes" :aria-label="t('studio.settings.ghRefresh')" type="number" min="1" /><span class="suffix">{{ t('studio.settings.minutes') }}</span></label>
      </div>
      <div class="st-opt">
        <div>{{ t('studio.settings.ghProxy') }}<small>{{ t('studio.settings.ghProxySub') }}</small></div>
        <label class="st-field w260 mono-in"><input v-model="cfg.github.proxy" :aria-label="t('studio.settings.ghProxy')" placeholder="http://127.0.0.1:7890" /></label>
      </div>
      <div class="st-opt">
        <div>{{ t('studio.settings.ghInsecure') }}<small>{{ t('studio.settings.ghInsecureSub') }}</small></div>
        <StSwitch v-model="cfg.github.insecureTls" :label="t('studio.settings.ghInsecure')" />
      </div>

      <div class="panel">
        <div class="ph">
          <div class="stats">
            <div v-for="k in ['repos', 'stars', 'followers', 'commits']" :key="k" class="s">
              <b class="mono">{{ preview?.stats?.[k] ?? '—' }}</b>
              <small>{{ t(`studio.settings.gh_${k}`) }}</small>
            </div>
          </div>
          <button type="button" class="st-btn g" :disabled="syncing" @click="sync">
            <SIcon name="refresh" :size="18" :class="{ spin: syncing }" />{{ syncing ? t('studio.settings.ghSyncing') : t('studio.settings.ghSync') }}
          </button>
        </div>
        <p v-if="msg" class="msg" :class="{ err: !msg.ok }">{{ msg.text }}</p>
        <ul v-if="log.length" class="log">
          <li v-for="(e, i) in log.slice(0, 8)" :key="i" :class="{ err: !e.ok }">
            <i class="st-dot" /><time class="mono">{{ dateTimeText(e.at).slice(5) }}</time><span>{{ e.message }}</span>
          </li>
        </ul>
        <p v-else class="empty">{{ t('studio.settings.ghNoLog') }}</p>
      </div>
    </template>

    <template v-else>
      <div class="manual">
        <label v-for="k in (['repos', 'stars', 'followers', 'commits'] as const)" :key="k">
          <span class="st-flabel">{{ t(`studio.settings.gh_${k}`) }}</span>
          <span class="st-field"><input v-model.number="cfg.github.stats[k]" type="number" min="0" /></span>
        </label>
      </div>
    </template>
  </div>
</template>

<style scoped lang="scss">
.w260 { width: 260px; }
.w140 { width: 140px; }
.gh-token-control { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; width: 440px; max-width: 100%; min-width: 0; }.gh-token-control > label { flex: 1 1 200px; min-width: 0; }.gh-token-control > button { margin-inline-start: auto; flex-shrink: 0; }.token-input { -webkit-text-security: disc; }.token-input::placeholder { -webkit-text-security: none; }

.panel {
  margin-top: 14px;
  padding: 18px;
  border-radius: var(--r-md);
  box-shadow: 0 0 0 1px var(--line-2) inset;
}

.ph { display: flex; align-items: center; justify-content: space-between; gap: 20px; }

/* GitHub 数字：统计条（与首页 GitHub 卡同构） */
.stats {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  border-radius: var(--r-md);
  background: var(--well);

  .s { padding: 14px 16px 12px; min-width: 0; }
  .s + .s { box-shadow: -1px 0 0 var(--line-2); }
  b { display: block; font-size: 28px; font-weight: 600; line-height: 1.1; letter-spacing: -0.02em; font-variant-numeric: tabular-nums; }
  small { display: block; margin-top: 4px; font-size: 12.5px; color: var(--st-ink-3); }
}

.spin { animation: spin 0.9s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.msg { margin: 12px 0 0; font-size: 13px; color: color-mix(in oklab, var(--green) 70%, var(--st-ink)); &.err { color: var(--red); } }

.log {
  list-style: none;
  margin: 14px 0 0;
  padding: 12px 0 0;
  border-top: 1px solid var(--line);
  display: flex;
  flex-direction: column;
  gap: 6px;

  li {
    display: grid;
    grid-template-columns: 8px 92px 1fr;
    gap: 10px;
    align-items: center;
    font-size: 13px;
    color: var(--st-ink-2);

    .st-dot { --c: var(--green); }
    &.err .st-dot { --c: var(--red); }
    &.err span { color: var(--red); }
    time { font-size: 11.5px; color: var(--st-ink-3); }
    span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  }
}

.empty { margin: 12px 0 0; font-size: 12.5px; color: var(--st-ink-3); }

.manual {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  padding-top: 14px;
}

@media (max-width: 767px) {
  .gh-token-control { width: 100%; }
  .panel { padding: 12px; }
  .ph { flex-wrap: wrap; gap: 12px; }
  .stats { flex-basis: 100%; grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .manual { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .manual > label { min-width: 0; }
  .log li { grid-template-columns: 8px minmax(0, 1fr); gap: 6px 8px; }
  .log li span { grid-column: 2; white-space: normal; overflow-wrap: anywhere; }
}
</style>
