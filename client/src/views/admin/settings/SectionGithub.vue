<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../../api';
import type { SiteConfig } from '../../../stores/config';
import { formatDateTime } from '../../../utils/date';

/** GitHub：账号 / 模式 / 手动数据 + 同步面板（立即同步、数据预览、同步日志） */
defineProps<{ cfg: SiteConfig }>();
const { t } = useI18n();

/** ISO 时间按站点时区显示 */
function logTime(iso: string): string {
  try {
    return formatDateTime(iso.replace('T', ' ').replace('Z', '').slice(0, 19)).slice(5);
  } catch {
    return iso.slice(5, 16);
  }
}

interface SyncLogEntry { at: string; ok: boolean; message: string }
const ghLog = ref<SyncLogEntry[]>([]);
const ghPreview = ref<{ stats?: Record<string, number>; fetchedAt?: string } | null>(null);
const ghSyncing = ref(false);
const ghMsg = ref('');

async function loadGhLog(): Promise<void> {
  try {
    const data = await adminApi.githubLog();
    ghLog.value = data.log;
    ghPreview.value = data.preview as typeof ghPreview.value;
  } catch { /* 忽略 */ }
}

async function syncNow(): Promise<void> {
  if (ghSyncing.value) return;
  ghSyncing.value = true;
  ghMsg.value = '';
  try {
    await adminApi.githubSync();
    ghMsg.value = t('admin.ghSyncOk');
  } catch {
    ghMsg.value = t('admin.ghSyncFail');
  } finally {
    ghSyncing.value = false;
    await loadGhLog();
  }
}

onMounted(loadGhLog);
</script>

<template>
  <section id="sec-github" class="card">
    <h2>{{ t('admin.secGithub') }}</h2>
    <div class="row3">
      <label><span>{{ t('admin.ghUser') }}</span><input v-model="cfg.github.username" type="text" /></label>
      <label>
        <span>{{ t('admin.ghMode') }}</span>
        <select v-model="cfg.github.mode">
          <option value="manual">{{ t('admin.ghManual') }}</option>
          <option value="api">{{ t('admin.ghApi') }}</option>
        </select>
      </label>
      <label><span>{{ t('admin.ghRefresh') }}</span><input v-model.number="cfg.github.refreshMinutes" type="number" min="1" /></label>
    </div>
    <label v-if="cfg.github.mode === 'api'">
      <span>{{ t('admin.ghToken') }}</span>
      <input v-model="cfg.github.token" type="password" autocomplete="off" placeholder="ghp_…（可留空走匿名公开接口）" />
    </label>
    <div v-if="cfg.github.mode === 'api'" class="row2" style="margin-top: 14px">
      <label>
        <span>{{ t('admin.ghProxy') }}</span>
        <input v-model="cfg.github.proxy" type="text" placeholder="http://127.0.0.1:7890（留空读 HTTPS_PROXY）" />
      </label>
      <label class="switch" style="align-self: end">
        <input v-model="cfg.github.insecureTls" type="checkbox" />
        <i class="track" aria-hidden="true" />
        <span>{{ t('admin.ghInsecure') }}</span>
      </label>
    </div>
    <div v-if="cfg.github.mode === 'manual'" class="row3" style="margin-top: 14px">
      <label><span>{{ t('admin.ghRepos') }}</span><input v-model.number="cfg.github.stats.repos" type="number" /></label>
      <label><span>Stars</span><input v-model.number="cfg.github.stats.stars" type="number" /></label>
      <label><span>{{ t('admin.ghFollowers') }}</span><input v-model.number="cfg.github.stats.followers" type="number" /></label>
    </div>
    <div v-if="cfg.github.mode === 'manual'" class="row3">
      <label><span>{{ t('admin.ghCommits') }}</span><input v-model.number="cfg.github.stats.commits" type="number" /></label>
    </div>

    <!-- 同步面板：立即同步 / 数据预览 / 同步日志 -->
    <div v-if="cfg.github.mode === 'api'" class="gh-panel">
      <div class="gh-actions">
        <button class="btn primary" :disabled="ghSyncing" @click="syncNow">
          {{ ghSyncing ? t('admin.ghSyncing') : t('admin.ghSyncNow') }}
        </button>
        <span v-if="ghMsg" class="msg">{{ ghMsg }}</span>
      </div>

      <div v-if="ghPreview?.stats" class="gh-preview">
        <div class="gp"><strong>{{ ghPreview.stats.repos }}</strong><span>{{ t('admin.ghRepos') }}</span></div>
        <div class="gp"><strong>{{ ghPreview.stats.stars }}</strong><span>Stars</span></div>
        <div class="gp"><strong>{{ ghPreview.stats.followers }}</strong><span>{{ t('admin.ghFollowers') }}</span></div>
        <div class="gp"><strong>{{ ghPreview.stats.commits }}</strong><span>{{ t('admin.ghCommits') }}</span></div>
      </div>

      <ul v-if="ghLog.length" class="gh-log">
        <li v-for="(entry, i) in ghLog" :key="i" :class="{ err: !entry.ok }">
          <time>{{ logTime(entry.at) }}</time>
          <span>{{ entry.message }}</span>
        </li>
      </ul>
      <p v-else class="hint">{{ t('admin.ghNoLog') }}</p>
    </div>
  </section>
</template>

<style scoped lang="scss">
@use './settings-shared';

/* GitHub 同步面板 */
.gh-panel {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px dashed var(--border);
}

.gh-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}

.gh-preview {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
  margin-bottom: 14px;
}

.gp {
  text-align: center;
  padding: 12px 6px;
  border-radius: 10px;
  background: var(--bg);
  border: 1px solid var(--border);

  strong {
    display: block;
    font-family: var(--font-serif);
    font-size: 20px;
    background: var(--grad-title);
    background-clip: text;
    -webkit-background-clip: text;
    color: transparent;
  }

  span { font-size: 11px; color: var(--text-2); }
}

.gh-log {
  list-style: none;
  max-height: 180px;
  overflow-y: auto;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--bg);

  li {
    display: flex;
    gap: 12px;
    padding: 8px 12px;
    font-size: 12.5px;
    border-bottom: 1px solid var(--border);

    time { color: var(--text-2); flex-shrink: 0; font-variant-numeric: tabular-nums; }
    span { color: var(--text); }

    &.err span { color: var(--accent-red); }
    &:last-child { border-bottom: none; }
  }
}
</style>
