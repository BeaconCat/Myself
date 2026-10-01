<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type UpdateStatus, type UpdateRelease, type UpdateReleasePage, type UpdatePreferences } from '../../../api';
import { useDialogStore } from '../../../stores/dialog';
import SIcon from '../studio/SIcon.vue';
import StSwitch from '../studio/StSwitch.vue';

const props = defineProps<{ dirty: boolean }>();
const { t, te } = useI18n();
const dialog = useDialogStore();
const status = ref<UpdateStatus | null>(null);
const catalogue = ref<UpdateReleasePage | null>(null);
const preferences = reactive<UpdatePreferences>({ repository: '', autoUpdate: false, subscribe: false, email: '' });
const saved = ref('');
const changed = computed(() => !!saved.value && saved.value !== JSON.stringify(preferences));
const busy = ref(false), listBusy = ref(false), reconnecting = ref(false);
const error = ref(''), listError = ref('');
const page = ref(1);
let timer = 0, disposed = false;
const active = computed(() => ['downloading','verifying','backing_up','restarting','rolling_back'].includes(status.value?.phase ?? ''));
const disabled = computed(() => busy.value || listBusy.value || active.value || !!status.value?.busy);
const percent = computed(() => status.value?.total ? Math.min(100, Math.round(status.value.downloaded / status.value.total * 100)) : 0);
const buildName = computed(() => ['Myself', status.value?.current.codename, status.value?.current.version].filter(Boolean).join(' '));
const repoURL = computed(() => status.value?.repository ? `https://github.com/${status.value.repository}/releases` : '');

function explain(code: string): string {
  const key = `studio.updates.errors.${code}`;
  return te(key) ? t(key) : `${t('studio.updates.operationFailed')} (${code})`;
}
function accept(next: UpdateStatus, force = false): void {
  const editable = force || !saved.value || !changed.value;
  status.value = next;
  if (next.preferences && editable) {
    Object.assign(preferences, next.preferences);
    preferences.email ||= next.adminEmail ?? '';
    saved.value = JSON.stringify(preferences);
  }
}
async function load(): Promise<void> {
  window.clearTimeout(timer);
  try { accept(await adminApi.systemStatus()); reconnecting.value = false; error.value = ''; }
  catch (e) { if (active.value) reconnecting.value = true; else error.value = explain((e as Error).message); }
  finally { if (!disposed && (active.value || reconnecting.value || status.value?.busy)) timer = window.setTimeout(() => void load(), 2000); }
}
async function releases(next = 1): Promise<void> {
  if (disabled.value || !status.value) return;
  listBusy.value = true; listError.value = '';
  try { catalogue.value = await adminApi.updateReleases(next); page.value = next; }
  catch (e) { listError.value = explain((e as Error).message); }
  finally { listBusy.value = false; }
}
async function save(): Promise<void> {
  if (disabled.value) return;
  if (preferences.repository !== status.value?.repository) {
    if (!await dialog.confirm({ title: t('studio.updates.sourceConfirm'), message: t('studio.updates.sourceBody', { repository: preferences.repository }) })) return;
    preferences.autoUpdate = false;
  }
  busy.value = true; error.value = '';
  try { accept(await adminApi.saveUpdatePreferences({ ...preferences }), true); catalogue.value = null; }
  catch (e) { error.value = explain((e as Error).message); }
  finally { busy.value = false; }
  if (!error.value) await releases();
}
async function check(): Promise<void> {
  busy.value = true; error.value = '';
  try { accept(await adminApi.checkUpdate()); await load(); }
  catch (e) { error.value = explain((e as Error).message); }
  finally { busy.value = false; }
  await releases(page.value);
}
async function install(release: UpdateRelease): Promise<void> {
  if (props.dirty || changed.value || disabled.value || !release.installable || status.value?.reason) return;
  const rollback = release.relation === 'older';
  if (!await dialog.confirm({ title: t(rollback ? 'studio.updates.rollbackTitle' : 'studio.updates.confirmTitle', { version: release.version }), message: t(rollback ? 'studio.updates.rollbackBody' : 'studio.updates.confirmBody'), confirmText: t(rollback ? 'studio.updates.rollback' : 'studio.updates.install'), danger: rollback })) return;
  busy.value = true; error.value = '';
  try { accept(await adminApi.selectUpdateRelease(status.value!.repository, release.id), true); await load(); }
  catch (e) { error.value = explain((e as Error).message); }
  finally { busy.value = false; }
}
function refresh(): void { window.location.reload(); }
onMounted(async () => { await load(); if (status.value && !active.value) await releases(); });
onBeforeUnmount(() => { disposed = true; window.clearTimeout(timer); });
</script>

<template>
  <div class="updates">
    <p class="description">{{ t('studio.updates.description') }}</p>
    <template v-if="status">
      <dl class="facts">
        <div><dt>{{ t('studio.updates.current') }}</dt><dd>{{ buildName }}</dd></div>
        <div><dt>{{ t('studio.updates.platform') }}</dt><dd>{{ status.current.os }} / {{ status.current.arch }}</dd></div>
        <div><dt>{{ t('studio.updates.database') }}</dt><dd>{{ status.databaseDriver === 'mysql' ? 'MySQL' : 'SQLite' }}</dd></div>
      </dl>
      <form v-if="status.preferences" class="preferences" @submit.prevent="save">
        <label><span class="st-flabel">{{ t('studio.updates.repository') }}</span><span class="st-field"><input v-model.trim="preferences.repository" placeholder="owner/repository" required :disabled="disabled" /></span></label>
        <p class="description">{{ t('studio.updates.repositoryHint') }}</p>
        <div class="option"><div><b>{{ t('studio.updates.automatic') }}</b><p>{{ t('studio.updates.automaticHint') }}</p></div><StSwitch v-model="preferences.autoUpdate" :label="t('studio.updates.automatic')" :disabled="disabled || (!!status.reason && !preferences.autoUpdate)" /></div>
        <div class="option"><div><b>{{ t('studio.updates.subscribe') }}</b><p>{{ t('studio.updates.subscribeHint') }}</p></div><StSwitch v-model="preferences.subscribe" :label="t('studio.updates.subscribe')" :disabled="disabled || (!status.smtpReady && !preferences.subscribe)" /></div>
        <p v-if="!status.smtpReady" class="note">{{ t('studio.updates.smtpRequired') }} <a href="#set-mail">{{ t('studio.updates.configureSMTP') }}</a></p>
        <label v-if="status.smtpReady || preferences.subscribe"><span class="st-flabel">{{ t('studio.updates.recipient') }}</span><span class="st-field"><input v-model.trim="preferences.email" type="email" :required="preferences.subscribe" :disabled="disabled" autocomplete="email" /></span></label>
        <p class="description">{{ t('studio.updates.schedule', { timezone: status.timezone || 'UTC' }) }}<template v-if="status.nextCheck && (preferences.autoUpdate || preferences.subscribe)"> {{ t('studio.updates.nextCheck') }} {{ new Date(status.nextCheck).toLocaleString() }}</template></p>
        <button class="st-btn" :disabled="disabled || !changed">{{ t('studio.updates.savePreferences') }}</button>
      </form>
      <div class="status" role="status" aria-live="polite"><strong>{{ reconnecting ? t('studio.updates.reconnecting') : t(`studio.updates.phase.${status.phase}`) }}</strong><span v-if="status.checkedAt">{{ t('studio.updates.checked') }} {{ new Date(status.checkedAt).toLocaleString() }}</span></div>
      <template v-if="active"><p v-if="status.target">{{ status.target.name }} · {{ status.target.version }}</p><progress :value="percent" max="100" :aria-label="t('studio.updates.progress')" /><p v-if="status.phase === 'downloading'">{{ percent }}% · {{ (status.downloaded / 1048576).toFixed(1) }} / {{ (status.total / 1048576).toFixed(1) }} MB</p></template>
      <p v-if="status.reason" class="note">{{ explain(status.reason) }}</p>
      <p v-if="dirty || changed" class="note">{{ t('studio.updates.saveFirst') }}</p>
      <p v-if="status.backup" class="backup">{{ t('studio.updates.backup') }} <code>{{ status.backup }}</code></p>
      <p v-if="status.notification" class="note">{{ t('studio.updates.notificationResult', { version: status.notification.version, state: t(`studio.updates.notification.${status.notification.status}`), email: status.notification.email }) }}<span v-if="status.notification.error"> · {{ status.notification.error }}</span></p>
      <div v-if="status.available" class="release highlight"><div><b>{{ status.available.name }}</b><span>{{ status.available.version }} · {{ (status.available.size / 1048576).toFixed(1) }} MB</span></div><button v-if="status.canApply" class="st-btn p" :disabled="disabled || dirty || changed" @click="install(status.available)">{{ t('studio.updates.installVersion', { version: status.available.version }) }}</button><p v-else-if="status.available.reason" class="note">{{ explain(status.available.reason) }}</p></div>
      <p v-if="status.error" class="error" role="alert">{{ explain(status.error) }}</p>
    </template>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <div class="actions"><button class="st-btn" :disabled="disabled || changed" @click="check"><SIcon name="refresh" :size="16" />{{ busy && !active ? t('studio.updates.checking') : t('studio.updates.check') }}</button><button v-if="status?.phase === 'installed' || status?.phase === 'rolled_back'" class="st-btn p" @click="refresh">{{ t('studio.updates.refresh') }}</button></div>
    <section v-if="status" class="history">
      <div class="history-heading"><h3>{{ t('studio.updates.history') }}</h3><a v-if="repoURL" :href="repoURL" target="_blank" rel="noopener noreferrer">{{ status.repository }}</a></div>
      <p v-if="listBusy" class="description" role="status">{{ t('studio.updates.loadingVersions') }}</p>
      <p v-else-if="listError" class="error" role="alert">{{ listError }} <button class="st-btn" @click="releases(page)">{{ t('studio.updates.retry') }}</button></p>
      <div v-else-if="catalogue && !catalogue.items.length" class="empty"><SIcon name="archive" :size="30" /><b>{{ t('studio.updates.emptyTitle') }}</b><p>{{ t('studio.updates.emptyBody') }}</p><a :href="repoURL" target="_blank" rel="noopener noreferrer">{{ t('studio.updates.releaseNotes') }}</a></div>
      <div v-else-if="catalogue" class="version-list">
        <article v-for="release in catalogue.items" :key="release.id" class="release">
          <div class="release-top"><div><b>{{ release.name }}</b><p class="description">{{ release.version || t('studio.updates.unknownVersion') }}<template v-if="release.codename"> · {{ release.codename }}</template><template v-if="release.publishedAt"> · {{ new Date(release.publishedAt).toLocaleDateString() }}</template></p></div><button class="st-btn" :disabled="disabled || dirty || changed || !!status.reason || !release.installable" @click="install(release)">{{ t(`studio.updates.select.${release.relation || 'newer'}`) }}</button></div>
          <p v-if="release.reason" class="note">{{ explain(release.reason) }}</p>
          <a :href="release.url" target="_blank" rel="noopener noreferrer">{{ t('studio.updates.releaseNotes') }}</a>
          <details v-if="release.notes"><summary>{{ t('studio.updates.changes') }}</summary><pre>{{ release.notes }}</pre></details>
        </article>
      </div>
      <div v-if="catalogue && (page > 1 || catalogue.hasNext)" class="pagination"><button class="st-btn" :disabled="disabled || page === 1" @click="releases(page - 1)">{{ t('studio.updates.previous') }}</button><span>{{ t('studio.updates.page', { page }) }}</span><button class="st-btn" :disabled="disabled || !catalogue.hasNext" @click="releases(page + 1)">{{ t('studio.updates.next') }}</button></div>
    </section>
  </div>
</template>

<style scoped lang="scss">
.updates { display: grid; gap: 16px; min-width: 0; }
.description,.note,.backup,.option p { color: var(--st-ink-3); font-size: 13px; line-height: 1.7; margin: 0; overflow-wrap: anywhere; }
.facts { display: flex; flex-wrap: wrap; gap: 20px 36px; margin: 0; }
.facts div { display: grid; gap: 5px; min-width: 0; }
dt { font-size: 12px; color: var(--st-ink-3); } dd { margin: 0; font-family: var(--font-mono); font-size: 17px; overflow-wrap: anywhere; }
.preferences { display: grid; gap: 12px; border-block: 1px solid var(--line); padding: 18px 0; }
.preferences label { display: grid; gap: 6px; min-width: 0; }.preferences input { min-width: 0; width: 100%; }.preferences > button { justify-self: start; }
.option { display: flex; justify-content: space-between; align-items: center; gap: 18px; }.option b { font-size: 14px; }
.status,.actions,.pagination,.history-heading { display: flex; flex-wrap: wrap; align-items: center; gap: 10px 18px; }.status span { color: var(--st-ink-3); font-size: 12px; }
.history,.version-list { display: grid; gap: 12px; min-width: 0; }.history-heading { justify-content: space-between; }.history-heading h3 { font-size: 16px; margin: 0; }
.release { border: 1px solid var(--line-2); border-radius: var(--r-md); padding: 16px; display: grid; gap: 10px; min-width: 0; }.release b { overflow-wrap: anywhere; }.release span { color: var(--st-ink-3); font-size: 12px; }.release.highlight > div { display: grid; gap: 5px; }.release > button { justify-self: start; }
.release-top { display: flex; align-items: center; justify-content: space-between; gap: 12px; }.release-top > div { min-width: 0; }.release-top > button { flex-shrink: 0; }
a { color: var(--primary); font-size: 13px; overflow-wrap: anywhere; } pre { white-space: pre-wrap; overflow-wrap: anywhere; font: inherit; font-size: 12px; max-height: 260px; overflow: auto; } summary { cursor: pointer; font-size: 13px; }
progress { width: 100%; height: 8px; accent-color: var(--primary); }.error { color: var(--danger,#c63b46); overflow-wrap: anywhere; font-size: 13px; margin: 0; }.backup code { overflow-wrap: anywhere; }
.empty { display: grid; justify-items: center; gap: 12px; text-align: center; padding: 28px 18px; background: var(--surface-2); border-radius: var(--r-md); }.empty p { max-width: 460px; margin: 0; color: var(--st-ink-3); font-size: 13px; }.pagination { justify-content: flex-end; font-size: 13px; }
@media(max-width:600px) { .release-top { align-items: flex-start; flex-direction: column; }.facts { gap: 14px; } }
</style>
