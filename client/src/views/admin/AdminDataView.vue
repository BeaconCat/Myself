<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type BackupInfo } from '../../api';
import { useDialogStore } from '../../stores/dialog';

const { t } = useI18n();

const backups = ref<BackupInfo[]>([]);
const busy = ref(false);
const autoHours = ref(0);
const savedMsg = ref('');

async function load(): Promise<void> {
  backups.value = await adminApi.backups();
  const cfg = await adminApi.settings() as { backup?: { autoHours?: number } };
  autoHours.value = Number(cfg.backup?.autoHours) || 0;
}

async function create(): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  try {
    await adminApi.createBackup();
    await load();
  } finally {
    busy.value = false;
  }
}

async function download(backup: BackupInfo): Promise<void> {
  const blob = await adminApi.downloadBackup(backup.name);
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = backup.name;
  a.click();
  URL.revokeObjectURL(url);
}

async function remove(backup: BackupInfo): Promise<void> {
  const ok = await useDialogStore().confirm({
    title: t('admin.delete'),
    message: t('admin.confirmDeleteBackup', { name: backup.name }),
    danger: true,
  });
  if (!ok) return;
  await adminApi.deleteBackup(backup.name);
  await load();
}

async function saveAuto(): Promise<void> {
  await adminApi.saveSettings({ backup: { autoHours: Math.max(0, autoHours.value) } });
  savedMsg.value = t('admin.saved');
  window.setTimeout(() => { savedMsg.value = ''; }, 2000);
}

function formatSize(bytes: number): string {
  if (bytes > 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
  return `${Math.round(bytes / 1024)} KB`;
}

onMounted(load);
</script>

<template>
  <div>
    <header class="head">
      <div>
        <h1 class="page-h">{{ t('admin.menuData') }}</h1>
        <p class="hint">{{ t('admin.dataHint') }}</p>
      </div>
      <button class="btn primary" :disabled="busy" @click="create">
        {{ busy ? t('admin.backingUp') : t('admin.backupNow') }}
      </button>
    </header>

    <!-- 自动备份 -->
    <div class="auto-card">
      <label>
        <span>{{ t('admin.autoBackup') }}</span>
        <input v-model.number="autoHours" type="number" min="0" step="1" />
      </label>
      <button class="btn ghost" @click="saveAuto">{{ t('admin.save') }}</button>
      <span v-if="savedMsg" class="msg">{{ savedMsg }}</span>
    </div>

    <!-- 备份列表 -->
    <table class="table">
      <thead>
        <tr>
          <th>{{ t('admin.backupFile') }}</th>
          <th>{{ t('admin.size') }}</th>
          <th>{{ t('admin.colDate') }}</th>
          <th />
        </tr>
      </thead>
      <tbody>
        <tr v-for="backup in backups" :key="backup.name">
          <td><code>{{ backup.name }}</code></td>
          <td>{{ formatSize(backup.size) }}</td>
          <td>{{ backup.createdAt.slice(0, 19).replace('T', ' ') }}</td>
          <td class="ops">
            <button class="op" @click="download(backup)">{{ t('admin.download') }}</button>
            <button class="op danger" @click="remove(backup)">{{ t('admin.delete') }}</button>
          </td>
        </tr>
      </tbody>
    </table>

    <p v-if="!backups.length" class="empty">{{ t('admin.noBackups') }}</p>
  </div>
</template>

<style scoped lang="scss">
.head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.page-h { font-size: 26px; margin-bottom: 6px; }
.hint { font-size: 13px; color: var(--text-2); }

.btn {
  padding: 10px 24px;
  border: 1px solid transparent;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 700;
  transition: all var(--dur-fast) var(--ease-out);

  &.primary {
    color: #fff;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.4);

    &:hover:not(:disabled) { filter: brightness(1.08); }
    &:disabled { opacity: 0.55; }
  }

  &.ghost {
    background: var(--surface);
    border-color: var(--border);
    color: var(--text);

    &:hover { border-color: var(--primary); color: var(--primary); }
  }
}

.auto-card {
  display: flex;
  align-items: flex-end;
  gap: 14px;
  padding: 16px 18px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  margin-bottom: 20px;

  label {
    display: flex;
    flex-direction: column;
    gap: 6px;

    span { font-size: 12px; font-weight: 600; color: var(--text-2); }
  }

  input {
    width: 120px;
    padding: 9px 12px;
    border-radius: 10px;
    border: 1px solid var(--border);
    background: var(--bg);
    color: var(--text);
    font-size: 14px;
    outline: none;

    &:focus { border-color: var(--primary); }
  }

  .msg { font-size: 13px; color: var(--primary); }
}

.table {
  width: 100%;
  border-collapse: collapse;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;

  th, td {
    padding: 12px 16px;
    text-align: left;
    font-size: 14px;
    border-bottom: 1px solid var(--border);
  }

  th { background: var(--surface-2); font-size: 12px; color: var(--text-2); }
  code { font-family: Consolas, monospace; font-size: 13px; }
}

.ops { text-align: right; white-space: nowrap; }

.op {
  border: none;
  background: none;
  font-size: 13px;
  font-weight: 600;
  color: var(--primary);
  margin-left: 12px;

  &.danger { color: var(--accent-red); }
  &:hover { opacity: 0.75; }
}

.empty {
  color: var(--text-2);
  text-align: center;
  padding: 48px 0;
}
</style>
