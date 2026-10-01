<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type UpdateStatus } from '../../../api';
import { useDialogStore } from '../../../stores/dialog';
import SIcon from '../studio/SIcon.vue';

const props = defineProps<{ dirty: boolean }>();
const { t } = useI18n();
const dialog = useDialogStore();
const status = ref<UpdateStatus | null>(null);
const busy = ref(false);
const error = ref('');
const reconnecting = ref(false);
let timer = 0;
let disposed = false;
const active = computed(() => ['downloading', 'verifying', 'backing_up', 'restarting', 'rolling_back'].includes(status.value?.phase ?? ''));
const percent = computed(() => status.value?.total ? Math.min(100, Math.round(status.value.downloaded / status.value.total * 100)) : 0);

async function load(): Promise<void> {
  window.clearTimeout(timer);
  try {
    status.value = await adminApi.systemStatus();
    reconnecting.value = false;
    error.value = '';
  } catch {
    if (active.value) reconnecting.value = true;
    else error.value = t('studio.updates.loadFailed');
  } finally {
    if (!disposed && (active.value || reconnecting.value)) timer = window.setTimeout(() => void load(), 2000);
  }
}

async function check(): Promise<void> {
  busy.value = true;
  error.value = '';
  try { status.value = { ...status.value, ...await adminApi.checkUpdate() }; }
  catch { error.value = t('studio.updates.loadFailed'); }
  finally { busy.value = false; }
}

async function install(): Promise<void> {
  const version = status.value?.available?.version;
  if (!version || props.dirty || active.value) return;
  if (!await dialog.confirm({ title: t('studio.updates.confirmTitle', { version }), message: t('studio.updates.confirmBody'), confirmText: t('studio.updates.install') })) return;
  busy.value = true;
  error.value = '';
  try {
    status.value = { ...status.value, ...await adminApi.applyUpdate(version) };
    await load();
  } catch (e) { error.value = (e as Error).message; }
  finally { busy.value = false; }
}

function refresh(): void { window.location.reload(); }
onMounted(() => void load());
onBeforeUnmount(() => { disposed = true; window.clearTimeout(timer); });
</script>

<template>
  <div class="updates">
    <p class="description">{{ t('studio.updates.description') }}</p>
    <template v-if="status">
      <dl class="facts">
        <div><dt>{{ t('studio.updates.current') }}</dt><dd>{{ status.current.version }}</dd></div>
        <div><dt>{{ t('studio.updates.platform') }}</dt><dd>{{ status.current.os }} / {{ status.current.arch }}</dd></div>
        <div><dt>{{ t('studio.updates.database') }}</dt><dd>{{ status.databaseDriver === 'mysql' ? 'MySQL' : 'SQLite' }}</dd></div>
      </dl>
      <div class="status" role="status" aria-live="polite">
        <strong>{{ reconnecting ? t('studio.updates.reconnecting') : t(`studio.updates.phase.${status.phase}`) }}</strong>
        <span v-if="status.checkedAt">{{ t('studio.updates.checked') }} {{ new Date(status.checkedAt).toLocaleString() }}</span>
      </div>
      <template v-if="active">
        <progress :value="percent" max="100" :aria-label="t('studio.updates.progress')" />
        <p v-if="status.phase === 'downloading'">{{ percent }}% · {{ (status.downloaded / 1048576).toFixed(1) }} / {{ (status.total / 1048576).toFixed(1) }} MB</p>
      </template>
      <p v-if="status.reason" class="note">{{ t(`studio.updates.reason.${status.reason}`) }}</p>
      <p v-if="dirty" class="note">{{ t('studio.updates.saveFirst') }}</p>
      <p v-if="status.backup" class="backup">{{ t('studio.updates.backup') }} <code>{{ status.backup }}</code></p>
      <div v-if="status.available" class="release">
        <div><b>{{ status.available.version }}</b><span>{{ (status.available.size / 1048576).toFixed(1) }} MB</span></div>
        <a :href="status.available.url" target="_blank" rel="noopener noreferrer">{{ t('studio.updates.releaseNotes') }}</a>
        <details v-if="status.available.notes"><summary>{{ t('studio.updates.changes') }}</summary><pre>{{ status.available.notes }}</pre></details>
      </div>
      <p v-if="status.error" class="error" role="alert">{{ status.error === 'no_release' ? t('studio.updates.noRelease') : status.error }}</p>
    </template>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <div class="actions">
      <button class="st-btn" :disabled="busy || active" @click="check"><SIcon name="refresh" :size="16" />{{ busy && !active ? t('studio.updates.checking') : t('studio.updates.check') }}</button>
      <button v-if="status?.canApply" class="st-btn p" :disabled="busy || active || dirty" @click="install">{{ t('studio.updates.installVersion', { version: status.available?.version }) }}</button>
      <button v-if="status?.phase === 'installed' || status?.phase === 'rolled_back'" class="st-btn p" @click="refresh">{{ t('studio.updates.refresh') }}</button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.updates { display: grid; gap: 16px; min-width: 0; }
.description, .note, .backup { color: var(--st-ink-3); font-size: 13px; line-height: 1.7; margin: 0; }
.facts { display: flex; flex-wrap: wrap; gap: 20px 40px; margin: 0; }
.facts div { display: grid; gap: 5px; }
dt { font-size: 12px; color: var(--st-ink-3); }
dd { margin: 0; font-family: var(--font-mono); font-size: 18px; }
.status { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 20px; }
.status span, .release span { color: var(--st-ink-3); font-size: 12px; }
.release { border: 1px solid var(--border); border-radius: var(--r-md); padding: 16px; display: grid; gap: 10px; }
.release > div { display: flex; align-items: center; gap: 20px; }
.release a { color: var(--primary); font-size: 13px; }
pre { white-space: pre-wrap; overflow-wrap: anywhere; font: inherit; font-size: 12px; max-height: 260px; overflow: auto; }
summary { cursor: pointer; font-size: 13px; }
progress { width: 100%; height: 8px; accent-color: var(--primary); }
.error { color: var(--danger, #c63b46); overflow-wrap: anywhere; font-size: 13px; margin: 0; }
.actions { display: flex; flex-wrap: wrap; gap: 10px; }
.backup code { overflow-wrap: anywhere; }
</style>
