<script setup lang="ts">
import { settle, stableJson } from './studio/state';
import { migrateModules } from '../../about/migrate';
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';
import { useConfigStore, type AboutModule, type SiteConfig } from '../../stores/config';
import { useDialogStore } from '../../stores/dialog';
import { createModule, metaOf, spanOf, titleOf } from '../../about/registry';
import { MODULE_EDITORS } from '../../components/admin/modules';
import ModuleFrame from '../../components/admin/modules/ModuleFrame.vue';
import ModulePicker from '../../components/admin/ModulePicker.vue';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import PopMenu from './studio/PopMenu.vue';
import { toast } from './studio/toast';

/**
 * 关于：模块积木。按住把手拖动排序（FLIP 动画），眼睛开关控制前台显隐，展开即编辑。
 * 编辑器与公共头部（ModuleFrame）由「关于」线提供；保存写入 about.modules。
 */
const { t } = useI18n();
const config = useConfigStore();
const dialog = useDialogStore();

type AboutCfg = SiteConfig['about'];
const about = reactive<AboutCfg>(JSON.parse(JSON.stringify(config.cfg.about)));
const snapshot = ref('');
const loaded = ref(false);
const busy = ref(false);
const pickerOpen = ref(false);
const expanded = ref('');
const identityOpen = ref(false);

const dirty = computed(() => loaded.value && stableJson(about) !== snapshot.value);

async function load(): Promise<void> {
  try {
    const remote = (await adminApi.settings()) as unknown as SiteConfig;
    Object.assign(about, JSON.parse(JSON.stringify(remote.about ?? {})));
    if (!Array.isArray(about.modules)) about.modules = [];
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
  // 与前台同一套迁移：旧结构模块归一，并补出前台会合成的身份区 / 收尾格言，
  // 让后台列表与前台所见一致；快照在迁移之后取，未改动时不显示「未保存」
  about.modules = migrateModules(about.modules, about);
  await settle();
  snapshot.value = stableJson(about);
  loaded.value = true;
}

async function save(): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  try {
    await adminApi.saveSettings({ about: JSON.parse(JSON.stringify(about)) });
    snapshot.value = stableJson(about);
    await config.load();
    toast(t('studio.about.saved'));
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  } finally {
    busy.value = false;
  }
}

/* ===== 头像 ===== */
const avatarInput = ref<HTMLInputElement | null>(null);
const avatarBusy = ref(false);
async function onAvatar(e: Event): Promise<void> {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = '';
  if (!file || avatarBusy.value) return;
  avatarBusy.value = true;
  try {
    const [up] = await adminApi.uploadMedia([file]);
    if (up) about.avatar = up.url;
  } finally {
    avatarBusy.value = false;
  }
}

/* ===== 模块 ===== */
function addModules(types: string[]): void {
  const added = types.map((type) => createModule(type));
  about.modules.push(...added);
  pickerOpen.value = false;
  if (added.length === 1) expanded.value = added[0].id;
  toast(t('studio.about.added', { n: added.length }), { icon: 'plus' });
}

async function removeModule(mod: AboutModule): Promise<void> {
  const ok = await dialog.confirm({
    title: t('studio.about.removeTitle', { name: titleOf(mod) }),
    message: t('studio.about.removeBody'),
    confirmText: t('studio.delete'),
    danger: true,
  });
  if (!ok) return;
  about.modules = about.modules.filter((m) => m.id !== mod.id);
}

function toggleHidden(mod: AboutModule): void {
  mod.hidden = !mod.hidden;
}

function toggleExpand(mod: AboutModule): void {
  expanded.value = expanded.value === mod.id ? '' : mod.id;
}

function spanClass(mod: AboutModule): string {
  return expanded.value === mod.id ? 'span-3' : `span-${spanOf(mod)}`;
}

function menu(mod: AboutModule) {
  return [
    { icon: expanded.value === mod.id ? 'chevronD' : 'pen', label: expanded.value === mod.id ? t('studio.about.collapse') : t('studio.edit'), run: () => toggleExpand(mod) },
    { icon: mod.hidden ? 'eye' : 'eyeOff', label: mod.hidden ? t('studio.about.show') : t('studio.about.hide'), run: () => toggleHidden(mod) },
    { icon: 'trash', label: t('studio.delete'), danger: true, divider: true, run: () => void removeModule(mod) },
  ];
}

/* ===== 拖拽排序：指针拖动 + FLIP ===== */
const grid = ref<HTMLElement | null>(null);
const draggingId = ref('');
let drag: { el: HTMLElement; gx: number; gy: number } | null = null;

function cards(): HTMLElement[] {
  return grid.value ? [...grid.value.querySelectorAll<HTMLElement>('.mod[data-id]')] : [];
}

function place(e: PointerEvent): void {
  if (!drag || !grid.value) return;
  const gr = grid.value.getBoundingClientRect();
  drag.el.style.transform = `translate(${e.clientX - drag.gx - gr.left - drag.el.offsetLeft}px, ${e.clientY - drag.gy - gr.top - drag.el.offsetTop}px) rotate(-0.6deg) scale(1.01)`;
}

function onGripDown(e: PointerEvent, mod: AboutModule): void {
  const el = (e.currentTarget as HTMLElement).closest<HTMLElement>('.mod');
  if (!el) return;
  e.preventDefault();
  expanded.value = '';
  const r = el.getBoundingClientRect();
  drag = { el, gx: e.clientX - r.left, gy: e.clientY - r.top };
  draggingId.value = mod.id;
  el.style.transition = 'none';
  window.addEventListener('pointermove', onMove);
  window.addEventListener('pointerup', onUp, { once: true });
  place(e);
}

let moving = false;
async function onMove(e: PointerEvent): Promise<void> {
  if (!drag) return;
  place(e);
  if (moving) return;
  const target = cards().find((c) => {
    if (c === drag!.el) return false;
    const r = c.getBoundingClientRect();
    return e.clientX > r.left && e.clientX < r.right && e.clientY > r.top && e.clientY < r.bottom;
  });
  if (!target) return;
  const from = about.modules.findIndex((m) => m.id === draggingId.value);
  const to = about.modules.findIndex((m) => m.id === target.dataset.id);
  if (from < 0 || to < 0) return;
  moving = true;
  const first = new Map(cards().map((c) => [c.dataset.id!, c.getBoundingClientRect()]));
  const [m] = about.modules.splice(from, 1);
  about.modules.splice(to, 0, m);
  await nextTick();
  for (const c of cards()) {
    if (c === drag?.el) continue;
    const f = first.get(c.dataset.id!);
    if (!f) continue;
    const l = c.getBoundingClientRect();
    const dx = f.left - l.left;
    const dy = f.top - l.top;
    if (dx || dy) {
      c.animate([{ transform: `translate(${dx}px, ${dy}px)` }, { transform: 'none' }], { duration: 450, easing: 'cubic-bezier(.2,.8,.3,1.2)' });
    }
  }
  place(e);
  moving = false;
}

function onUp(): void {
  window.removeEventListener('pointermove', onMove);
  if (!drag) return;
  const el = drag.el;
  drag = null;
  el.style.transition = 'transform .5s cubic-bezier(.2,.8,.3,1.2), box-shadow .35s';
  el.style.transform = '';
  window.setTimeout(() => {
    draggingId.value = '';
    el.style.transition = '';
  }, 500);
}

onBeforeRouteLeave(async () => {
  if (!dirty.value) return true;
  return dialog.confirm({
    title: t('studio.about.leaveTitle'),
    message: t('studio.about.leaveBody'),
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
});
onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey);
  window.removeEventListener('pointermove', onMove);
});

const visibleCount = computed(() => about.modules.filter((m) => !m.hidden).length);
</script>

<template>
  <section class="studio view">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.about.title') }}</h1>
        <p>{{ t('studio.about.desc') }}</p>
      </div>
      <div class="act">
        <a class="st-btn g" href="/about" target="_blank" rel="noopener"><SIcon name="eye" :size="18" />{{ t('studio.about.preview') }}</a>
        <button type="button" class="st-btn" :class="dirty ? 'p' : 'g'" :disabled="busy || !dirty" @click="save">
          <SIcon name="check" :size="18" />{{ dirty ? t('studio.save') : t('studio.saved') }}
        </button>
      </div>
    </div>

    <!-- 站点身份：问候、建站天数与头像来源 -->
    <div class="identity st-rise" :class="{ open: identityOpen }">
      <button type="button" class="av" :disabled="avatarBusy" :title="t('studio.about.avatar')" @click="avatarInput?.click()">
        <img :src="about.avatar || '/favicon-256.png'" alt="" />
        <span><SIcon name="upload" :size="18" /></span>
      </button>
      <input ref="avatarInput" type="file" accept="image/*" hidden @change="onAvatar" />
      <div class="id-main">
        <b>{{ about.name || 'Myself' }}</b>
        <small>{{ about.tagline }} · {{ t('studio.about.since', { d: about.foundedAt }) }}</small>
      </div>
      <span class="count">{{ t('studio.about.count', { n: about.modules.length, v: visibleCount }) }}</span>
      <button type="button" class="st-btn g sm" @click="identityOpen = !identityOpen">
        {{ identityOpen ? t('studio.about.collapse') : t('studio.about.editIdentity') }}
        <SIcon name="chevronD" :size="16" class="chev" />
      </button>
      <div class="id-form">
        <div>
          <div class="grid2">
            <label><span class="st-flabel">{{ t('studio.about.name') }}</span><span class="st-field"><input v-model="about.name" /></span></label>
            <label><span class="st-flabel">{{ t('studio.about.tagline') }}</span><span class="st-field"><input v-model="about.tagline" /></span></label>
            <label><span class="st-flabel">{{ t('studio.about.founded') }}</span><span class="st-field"><input v-model="about.foundedAt" type="date" /></span></label>
            <label><span class="st-flabel">{{ t('studio.about.motto') }}</span><span class="st-field"><input v-model="about.motto" /></span></label>
          </div>
          <label><span class="st-flabel">{{ t('studio.about.bio') }}</span><span class="st-field ta"><textarea v-model="about.bio" rows="2" /></span></label>
        </div>
      </div>
    </div>

    <div ref="grid" class="mods">
      <div
        v-for="mod in about.modules"
        :key="mod.id"
        class="mod"
        :class="[spanClass(mod), { off: mod.hidden, open: expanded === mod.id, dragging: draggingId === mod.id }]"
        :data-id="mod.id"
      >
        <div class="mh">
          <span class="grip" :title="t('studio.about.drag')" @pointerdown="onGripDown($event, mod)"><SIcon name="grip" :size="18" /></span>
          <span class="mi"><svg class="st-ic" width="16" height="16" viewBox="0 0 24 24"><path :d="metaOf(mod.type)?.icon ?? ''" /></svg></span>
          <b>{{ titleOf(mod) }}</b>
          <span v-if="mod.hidden" class="hid">{{ t('studio.about.hiddenTag') }}</span>
          <span class="sp" />
          <button type="button" class="st-ibtn sm" :class="{ on: !mod.hidden }" :title="mod.hidden ? t('studio.about.show') : t('studio.about.hide')" @click="toggleHidden(mod)">
            <SIcon :name="mod.hidden ? 'eyeOff' : 'eye'" :size="18" />
          </button>
          <button type="button" class="st-ibtn sm" :class="{ on: expanded === mod.id }" :title="t('studio.edit')" @click="toggleExpand(mod)">
            <SIcon name="pen" :size="18" />
          </button>
          <PopMenu :items="menu(mod)" />
        </div>
        <div v-if="expanded !== mod.id" class="mb" @click="toggleExpand(mod)">
          <p class="sum">{{ metaOf(mod.type)?.summary(mod.data) || metaOf(mod.type)?.desc }}</p>
          <p class="desc">{{ metaOf(mod.type)?.desc }}</p>
        </div>
        <div v-else class="ed">
          <ModuleFrame :mod="mod">
            <component :is="MODULE_EDITORS[mod.type]" v-if="MODULE_EDITORS[mod.type]" :mod="mod" />
          </ModuleFrame>
          <div class="ed-ft">
            <button type="button" class="st-btn q sm" @click="expanded = ''">{{ t('studio.about.collapse') }}</button>
          </div>
        </div>
      </div>

      <button type="button" class="mod addm span-1" @click="pickerOpen = true">
        <SIcon name="plus" />{{ t('studio.about.add') }}
      </button>
    </div>

    <ModulePicker v-if="pickerOpen" :existing="about.modules.map((m) => m.type)" @close="pickerOpen = false" @add="addModules" />
  </section>
</template>

<style scoped lang="scss">
.view {
  max-width: 1280px;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

/* ---------- 身份条 ---------- */
.identity {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto auto;
  align-items: center;
  gap: 0 16px;
  padding: 16px 18px;
  border-radius: var(--r-lg);
  background: var(--well);
  margin-bottom: 28px;

  .av {
    position: relative;
    width: 48px;
    height: 48px;
    border-radius: 50%;
    overflow: hidden;
    box-shadow: 0 0 0 2px var(--paper), 0 0 0 3px var(--line-2);

    img { width: 100%; height: 100%; object-fit: cover; display: block; }

    span {
      position: absolute;
      inset: 0;
      display: grid;
      place-items: center;
      color: #fff;
      background: rgba(0, 0, 0, 0.45);
      opacity: 0;
      transition: opacity var(--dur-fast);
    }

    &:hover span { opacity: 1; }
  }

  .id-main {
    min-width: 0;

    b { display: block; font: 600 17px/1.3 var(--font-serif); }
    small { font-size: 12.5px; color: var(--st-ink-3); }
  }

  .count { font-size: 12.5px; color: var(--st-ink-3); }
  .chev { transition: transform var(--dur) var(--ease-spring); }
  &.open .chev { transform: rotate(180deg); }

  .id-form {
    grid-column: 1 / -1;
    display: grid;
    grid-template-rows: 0fr;
    transition: grid-template-rows var(--dur) var(--ease-out);

    > div { overflow: hidden; min-height: 0; }
  }

  &.open .id-form { grid-template-rows: 1fr; }

  .grid2 {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 14px;
    padding-top: 18px;
    margin-bottom: 14px;
  }

  label { display: block; }
}

/* ---------- 积木 ---------- */
.mods {
  position: relative;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.span-1 { grid-column: span 1; }
.span-2 { grid-column: span 2; }
.span-3 { grid-column: span 3; }

.mod {
  position: relative;
  border-radius: var(--r-lg);
  background: var(--paper);
  box-shadow: 0 0 0 1px var(--line-2);
  display: flex;
  flex-direction: column;
  min-width: 0;
  transition: box-shadow var(--dur), opacity var(--dur);

  &:hover { box-shadow: 0 0 0 1px var(--line-3), var(--sh-card-hover); }

  &.dragging {
    z-index: 5;
    box-shadow: 0 0 0 1.5px var(--line-3), var(--shadow-pop);
    user-select: none;

    .grip { cursor: grabbing; }
  }

  &.open { box-shadow: 0 0 0 1px color-mix(in oklab, var(--ink) 55%, transparent), 0 0 0 3px color-mix(in oklab, var(--ink) 16%, transparent); }

  &.off .mb { opacity: 0.4; filter: grayscale(1); }
}

:root[data-mode='dark'] .mod { background: var(--well); }

.mh {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 10px 0 8px;
  font-size: 12.5px;
  color: var(--st-ink-3);

  .grip {
    width: 26px;
    height: 26px;
    display: grid;
    place-items: center;
    border-radius: var(--r-xs);
    cursor: grab;
    color: var(--st-ink-4);
    touch-action: none;

    &:hover { background: var(--hover); color: var(--st-ink-2); }
  }

  .mi {
    width: 26px;
    height: 26px;
    border-radius: var(--r-xs);
    display: grid;
    place-items: center;
    background: var(--tint);
    color: var(--ink);
  }

  b { font-weight: 500; font-size: 13.5px; color: var(--st-ink); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

  .hid {
    font-size: 11px;
    padding: 1px 7px;
    border-radius: var(--r-xs);
    background: var(--well-2);
    color: var(--st-ink-3);
    white-space: nowrap;
  }

  .sp { flex: 1; }
  .st-ibtn.sm { width: 28px; height: 28px; border-radius: var(--r-xs); }
  :deep(.st-ibtn) { width: 28px; height: 28px; border-radius: var(--r-xs); }
}

.mb {
  padding: 12px 18px 18px 44px;
  cursor: pointer;
  transition: opacity var(--dur), filter var(--dur);

  .sum {
    margin: 0 0 4px;
    font: 500 15px/1.6 var(--font-serif);
    color: var(--st-ink);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .desc {
    margin: 0;
    font-size: 12.5px;
    color: var(--st-ink-3);
    display: -webkit-box;
    -webkit-line-clamp: 1;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
}

.ed {
  padding: 14px 18px 16px;
  animation: ed-in var(--dur-slow) var(--ease-out);

  .ed-ft { display: flex; justify-content: flex-end; margin-top: 10px; }
}

@keyframes ed-in { from { opacity: 0; transform: translateY(-6px); } }

.addm {
  align-items: center;
  justify-content: center;
  flex-direction: row;
  gap: 8px;
  min-height: 104px;
  box-shadow: 0 0 0 1.5px var(--line-2) inset;
  background: transparent;
  color: var(--st-ink-3);
  font-size: 14px;
  cursor: pointer;

  &:hover { color: var(--ink); box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--ink) 55%, transparent) inset; }
}

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
  .mods { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .span-3 { grid-column: span 2; }
}
</style>
