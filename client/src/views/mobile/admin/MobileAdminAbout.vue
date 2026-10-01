<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { onBeforeRouteLeave, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { ArrowDown, ArrowUp, Check, ChevronLeft, ChevronRight, Eye, EyeOff, Plus, Trash2, User } from 'lucide';
import { adminApi } from '../../../api';
import { useConfigStore, type AboutModule, type SiteConfig } from '../../../stores/config';
import { useDialogStore } from '../../../stores/dialog';
import { migrateModules } from '../../../about/migrate';
import { injectIdentity, normalizeIdentity, plainText } from '../../../about/identity';
import { createModule, metaOf, MODULE_REGISTRY, spanOf, titleOf, variantOf } from '../../../about/registry';
import { MODULE_COMPONENTS } from '../../../about/components';
import { MODULE_EDITORS } from '../../../components/admin/modules';
import Icon from '../../../components/ui/Icon.vue';
import { BUILTIN_LOGO } from '../../../utils/siteLogo';
import { stableJson } from '../../admin/studio/state';
import { toast } from '../../../components/mobile-admin/state';
import '../../../about/kit.scss';

// Independent mobile page: compact module rows and one full-width editor at a time.
const { t } = useI18n();
const config = useConfigStore();
const dialog = useDialogStore();
const router = useRouter();
const about = reactive<SiteConfig['about']>(JSON.parse(JSON.stringify(config.cfg.about)));
const snapshot = ref('');
const loaded = ref(false), busy = ref(false), failed = ref(false);
const selected = ref(''), adding = ref(false), preview = ref(false), search = ref('');
const dirty = computed(() => loaded.value && stableJson(about) !== snapshot.value);
const current = computed(() => about.modules.find(m => m.id === selected.value));
const metadata = computed(() => current.value ? metaOf(current.value.type) : undefined);
const previewMod = computed(() => current.value ? injectIdentity(current.value, about) : null);
const visible = computed(() => about.modules.filter(m => !m.hidden).length);
const candidates = computed(() => MODULE_REGISTRY.filter(m => `${m.name} ${m.desc} ${m.group}`.includes(search.value.trim())));
const heading = computed(() => adding.value ? t('mobileAdmin.about.add') : current.value ? titleOf(current.value) : t('mobileAdmin.about.title'));

async function load(): Promise<void> {
  loaded.value = false; failed.value = false;
  try {
    const remote = await adminApi.settings() as unknown as SiteConfig;
    Object.assign(about, JSON.parse(JSON.stringify(remote.about ?? {})));
    normalizeIdentity(about);
    about.modules = migrateModules(about.modules ?? [], about);
    snapshot.value = stableJson(about);
    loaded.value = true;
  } catch { failed.value = true; }
}
async function save(): Promise<void> {
  if (busy.value || !loaded.value) return;
  busy.value = true;
  const payload = JSON.parse(JSON.stringify(about));
  try {
    await adminApi.saveSettings({ about: payload });
    snapshot.value = stableJson(payload);
    await config.load();
    toast(t('mobileAdmin.about.saved'));
  } catch { toast(t('mobileAdmin.about.saveFailed'), '', 'error'); }
  finally { busy.value = false; }
}
function edit(mod: AboutModule): void { selected.value = mod.id; preview.value = false; }
function back(): void {
  if (adding.value) { adding.value = false; return; }
  if (current.value) { selected.value = ''; preview.value = false; return; }
  void router.push({ name: 'admin-settings' });
}
function add(type: string): void {
  const mod = createModule(type);
  about.modules.push(mod);
  adding.value = false;
  edit(mod);
}
function move(index: number, delta: number): void {
  const target = index + delta;
  if (target < 0 || target >= about.modules.length) return;
  const [mod] = about.modules.splice(index, 1);
  about.modules.splice(target, 0, mod);
}
async function remove(mod: AboutModule): Promise<void> {
  if (!await dialog.confirm({ title: t('mobileAdmin.about.removeTitle', { name: titleOf(mod) }), message: t('mobileAdmin.about.removeBody'), confirmText: t('mobileAdmin.about.remove'), danger: true })) return;
  about.modules = about.modules.filter(m => m.id !== mod.id);
  selected.value = '';
}
function summary(mod: AboutModule): string { return plainText(metaOf(mod.type)?.summary(mod.data) ?? ''); }
onBeforeRouteLeave(() => !dirty.value || dialog.confirm({ title: t('mobileAdmin.about.leaveTitle'), message: t('mobileAdmin.about.leaveBody'), confirmText: t('mobileAdmin.about.leave'), danger: true }));
function beforeUnload(event: BeforeUnloadEvent): void { if (dirty.value) { event.preventDefault(); event.returnValue = ''; } }
onMounted(() => { void load(); window.addEventListener('beforeunload', beforeUnload); });
onBeforeUnmount(() => window.removeEventListener('beforeunload', beforeUnload));
</script>

<template>
  <section class="ma-about">
    <header class="nav">
      <button class="icon-button" :aria-label="t('mobileAdmin.common.back')" @click="back"><Icon :icon="ChevronLeft" :size="24" /></button>
      <h1>{{ heading }}</h1>
      <button class="save" :disabled="!dirty || busy || !loaded" @click="save"><Icon :icon="Check" :size="17" />{{ t(busy ? 'mobileAdmin.about.saving' : dirty ? 'mobileAdmin.about.save' : 'mobileAdmin.about.saved') }}</button>
    </header>
    <div :key="adding ? 'picker' : selected || 'list'" class="scroll">
      <div v-if="!loaded" class="state" role="status"><p>{{ t(failed ? 'mobileAdmin.about.loadFailed' : 'mobileAdmin.about.loading') }}</p><button v-if="failed" class="action" @click="load">{{ t('mobileAdmin.about.retry') }}</button></div>
      <template v-else-if="adding">
        <input v-model="search" class="search" type="search" :placeholder="t('mobileAdmin.about.search')" :aria-label="t('mobileAdmin.about.search')" />
        <div class="picker">
          <button v-for="meta in candidates" :key="meta.type" class="pick" @click="add(meta.type)"><span class="module-icon"><Icon :icon="meta.icon" :size="22" /></span><span><b>{{ meta.name }}</b><small>{{ meta.desc }}</small></span><Icon :icon="Plus" :size="20" /></button>
          <p v-if="!candidates.length" class="state">{{ t('mobileAdmin.about.noMatch') }}</p>
        </div>
      </template>
      <template v-else-if="current">
        <div class="edit-actions"><button class="action" :aria-pressed="preview" @click="preview = !preview"><Icon :icon="Eye" :size="18" />{{ t(preview ? 'mobileAdmin.about.hidePreview' : 'mobileAdmin.about.preview') }}</button><button class="action" :aria-pressed="!current.hidden" @click="current.hidden = !current.hidden"><Icon :icon="current.hidden ? EyeOff : Eye" :size="18" />{{ t(current.hidden ? 'mobileAdmin.about.hidden' : 'mobileAdmin.about.visible') }}</button></div>
        <div v-if="preview && previewMod" class="mobile-preview ak" inert>
          <section class="ak-m rv in" :class="[`m-${current.type}`, { card: metadata?.chrome(variantOf(current)) }]" :data-span="spanOf(current)" :data-type="current.type"><div class="ak-body"><component :is="MODULE_COMPONENTS[current.type]" v-if="MODULE_COMPONENTS[current.type]" :mod="previewMod" :variant="variantOf(current)" :span="spanOf(current)" :title="titleOf(current)" /></div></section>
        </div>
        <div class="form-card">
          <h2>{{ t('mobileAdmin.about.appearance') }}</h2>
          <label v-if="metadata?.chrome(variantOf(current))"><span>{{ t('mobileAdmin.about.moduleTitle') }}</span><input v-model="current.title" :placeholder="metadata.name" /></label>
          <label v-if="metadata && metadata.variants.length > 1"><span>{{ t('mobileAdmin.about.variant') }}</span><select :value="variantOf(current)" @change="current.variant = ($event.target as HTMLSelectElement).value"><option v-for="variant in metadata.variants" :key="variant.id" :value="variant.id">{{ variant.label }}</option></select></label>
          <label v-if="metadata && metadata.spans.length > 1"><span>{{ t('mobileAdmin.about.desktopWidth') }}</span><select v-model.number="current.span"><option v-for="span in metadata.spans" :key="span" :value="span">{{ t(`mobileAdmin.about.width${span}`) }}</option></select></label>
        </div>
        <div class="form-card mobile-fields"><h2>{{ t('mobileAdmin.about.content') }}</h2><component :is="MODULE_EDITORS[current.type]" v-if="MODULE_EDITORS[current.type]" :key="current.id" :mod="current" /><p v-else>{{ t('mobileAdmin.about.unsupported') }}</p></div>
        <button class="action delete" @click="remove(current)"><Icon :icon="Trash2" :size="18" />{{ t('mobileAdmin.about.remove') }}</button>
      </template>
      <template v-else>
        <div class="identity">
          <img :src="about.avatar || config.cfg.site.logo || BUILTIN_LOGO" alt="" />
          <div class="identity-name"><b>{{ about.name || config.cfg.site.title }}</b><span>{{ t('mobileAdmin.about.count', { total: about.modules.length, visible }) }}</span></div>
          <p v-if="about.tagline" class="tagline">{{ plainText(about.tagline) }}</p>
          <div class="identity-actions"><router-link class="action" :to="{ name: 'admin-identity' }"><Icon :icon="User" :size="17" />{{ t('mobileAdmin.about.identity') }}</router-link><a class="action" href="/about" target="_blank" rel="noopener"><Icon :icon="Eye" :size="17" />{{ t('mobileAdmin.about.preview') }}</a></div>
        </div>
        <p class="hint">{{ t('mobileAdmin.about.listHint') }}</p>
        <TransitionGroup name="module" tag="div" class="modules">
          <article v-for="(mod, index) in about.modules" :key="mod.id" class="module-row" :data-module="mod.type" :class="{ hidden: mod.hidden }">
            <button class="open-module" @click="edit(mod)"><span class="module-icon"><Icon v-if="metaOf(mod.type)" :icon="metaOf(mod.type)!.icon" :size="21" /></span><span class="module-copy"><b>{{ titleOf(mod) }}</b><small>{{ summary(mod) || metaOf(mod.type)?.desc }}</small></span><Icon :icon="ChevronRight" :size="19" /></button>
            <div class="row-actions"><button class="visibility" :aria-pressed="!mod.hidden" :aria-label="`${t(mod.hidden ? 'mobileAdmin.about.hidden' : 'mobileAdmin.about.visible')} · ${titleOf(mod)}`" @click="mod.hidden = !mod.hidden"><Icon :icon="mod.hidden ? EyeOff : Eye" :size="17" />{{ t(mod.hidden ? 'mobileAdmin.about.hidden' : 'mobileAdmin.about.visible') }}</button><span class="order">{{ index + 1 }}</span><button class="icon-button" :disabled="index === 0" :aria-label="`${t('mobileAdmin.about.moveUp')} · ${titleOf(mod)}`" @click="move(index, -1)"><Icon :icon="ArrowUp" :size="18" /></button><button class="icon-button" :disabled="index === about.modules.length - 1" :aria-label="`${t('mobileAdmin.about.moveDown')} · ${titleOf(mod)}`" @click="move(index, 1)"><Icon :icon="ArrowDown" :size="18" /></button></div>
          </article>
        </TransitionGroup>
        <p v-if="!about.modules.length" class="state">{{ t('mobileAdmin.about.empty') }}</p>
      </template>
    </div>
    <footer v-if="loaded && !current && !adding"><button class="add" @click="adding = true; search = ''"><Icon :icon="Plus" :size="21" />{{ t('mobileAdmin.about.add') }}</button></footer>
  </section>
</template>

<style scoped lang="scss">
.ma-about { position: absolute; inset: 0; display: flex; flex-direction: column; background: var(--bg); color: var(--text); min-width: 0; }
.nav { flex: none; display: grid; grid-template-columns: 44px minmax(0,1fr) auto; gap: 8px; align-items: center; padding: calc(var(--safe-t,0px) + 6px) 14px 8px; border-bottom: 1px solid var(--line); background: var(--surface); }
.nav h1 { font-size: 17px; text-align: center; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; margin: 0; }
button, a { -webkit-tap-highlight-color: transparent; } button { cursor: pointer; font: inherit; } button:disabled { opacity: .4; cursor: default; }
.icon-button { display: grid; place-items: center; width: 44px; height: 44px; flex: none; padding: 0; background: none; border: 0; color: var(--ink); border-radius: var(--r-sm); }
.save { display: flex; align-items: center; justify-content: center; gap: 4px; min-height: 44px; padding: 0 8px; color: var(--primary); background: none; border: 0; font-size: 14px; white-space: nowrap; }
.scroll { flex: 1; min-height: 0; overflow-y: auto; overscroll-behavior-y: contain; padding: 18px 16px calc(var(--safe-b,0px) + 24px); display: flex; flex-direction: column; gap: 16px; }
.scroll > * { flex-shrink: 0; min-width: 0; } .hint { margin: 0; color: var(--text-3); font-size: 12px; line-height: 1.6; }
.identity { display: grid; grid-template-columns: 52px minmax(0,1fr); gap: 12px; padding: 18px; border-radius: var(--r-lg); background: var(--fill); }
.identity img { width: 52px; height: 52px; object-fit: cover; border-radius: var(--r-md); }.identity-name { display: flex; flex-direction: column; justify-content: center; gap: 5px; min-width: 0; }.identity-name b { font-size: 21px; overflow-wrap: anywhere; }.identity-name span { font-size: 12px; color: var(--text-3); }
.tagline { grid-column: 1 / -1; margin: 0; font-size: 14px; line-height: 1.8; overflow-wrap: anywhere; }.identity-actions { grid-column: 1/-1; display: flex; flex-wrap: wrap; gap: 8px; }
.action { display: inline-flex; align-items: center; justify-content: center; gap: 7px; min-height: 44px; padding: 8px 13px; border: 1px solid var(--line); border-radius: var(--r-md); background: var(--surface); color: var(--text); font-size: 14px; text-decoration: none; }
.modules,.picker { display: grid; gap: 12px; }.module-row { min-width: 0; border: 1px solid var(--line); border-radius: var(--r-lg); background: var(--surface); overflow: hidden; }.module-row.hidden { background: var(--fill); }
.open-module,.pick { width: 100%; display: grid; grid-template-columns: 38px minmax(0,1fr) 20px; align-items: center; gap: 10px; padding: 14px; border: 0; background: transparent; text-align: left; color: var(--text); }
.module-icon { display: grid; place-items: center; width: 38px; height: 38px; background: color-mix(in oklab,var(--primary) 10%,transparent); color: var(--primary); border-radius: var(--r-sm); }.module-copy,.pick > span:nth-child(2) { min-width: 0; display: grid; gap: 5px; }.module-copy b,.pick b { font-size: 15px; overflow-wrap: anywhere; }.module-copy small { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 12px; color: var(--text-3); }
.row-actions { display: flex; align-items: center; gap: 4px; padding: 0 8px 4px 14px; }.visibility { display: flex; align-items: center; gap: 6px; min-height: 44px; border: 0; background: transparent; color: var(--text-2); font-size: 12px; }.order { margin-left: auto; padding: 0 8px; font: 12px var(--font-mono); color: var(--text-3); }
footer { padding: 10px 16px calc(var(--safe-b,0px) + 10px); border-top: 1px solid var(--line); background: var(--surface); }.add { width: 100%; display: flex; align-items: center; justify-content: center; gap: 8px; min-height: 48px; border: 0; border-radius: var(--r-md); background: var(--solid); color: var(--on-solid); font-weight: 600; }
.pick { border: 1px solid var(--line); background: var(--surface); border-radius: var(--r-lg); }.pick small { font-size: 12px; line-height: 1.6; color: var(--text-3); overflow-wrap: anywhere; }
.edit-actions { display: flex; gap: 8px; }.edit-actions > * { flex: 1; }.form-card { display: grid; gap: 16px; padding: 16px; border: 1px solid var(--line); border-radius: var(--r-lg); background: var(--surface); }.form-card h2 { font-size: 13px; color: var(--text-3); margin: 0; }.form-card > label { display: grid; gap: 8px; min-width: 0; font-size: 14px; }
input,select { width: 100%; min-width: 0; min-height: 44px; border: 1px solid var(--line); border-radius: var(--r-sm); padding: 10px 12px; background: var(--bg); color: var(--text); font: inherit; font-size: 16px; }.delete { color: var(--accent-red); align-self: stretch; }.state { padding: 30px 12px; text-align: center; font-size: 14px; color: var(--text-3); }
.mobile-preview { --ak-pad: 16px; min-width: 0; }.mobile-preview :deep(.ak-m) { min-width: 0; }.mobile-preview :deep(.m-profile .pf) { padding-top: 0; }
/* Shared field widgets receive a mobile form layout, independent of Studio's desktop page. */
.mobile-fields :deep(.ed) { min-width: 0; }.mobile-fields :deep(.el-item) { display: flex; flex-direction: column; padding: 12px; }.mobile-fields :deep(.el-body) { width: 100%; }.mobile-fields :deep(.el-ops) { width: 100%; justify-content: flex-end; }.mobile-fields :deep(.el-ops button) { width: 44px; height: 44px; }
.mobile-fields :deep(.a-input) { min-height: 44px; height: auto; font-size: 16px; }.mobile-fields :deep(.line) { flex-wrap: wrap; }.mobile-fields :deep(.line > .flex-in) { flex-basis: 140px; }.mobile-fields :deep(.grid2),.mobile-fields :deep(.grid3),.mobile-fields :deep(.grid4) { grid-template-columns: minmax(0,1fr); }.mobile-fields :deep(.span2) { grid-column: auto; }.mobile-fields :deep(.seg button) { min-height: 44px; }.mobile-fields :deep(.a-btn) { min-height: 44px; }.mobile-fields :deep(p) { overflow-wrap: anywhere; }
.module-move { transition: transform var(--dur) var(--ease-out); }.module-enter-active,.module-leave-active { transition: opacity var(--dur-fast), transform var(--dur-fast); }.module-enter-from,.module-leave-to { opacity: 0; transform: translateY(8px); }
@media(prefers-reduced-motion: reduce) { .module-move,.module-enter-active,.module-leave-active { transition: none; } }
</style>
