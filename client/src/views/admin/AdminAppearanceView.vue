<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';
import { FALLBACK_CONFIG, useConfigStore, type SiteConfig, type ThemePreset } from '../../stores/config';
import { useDialogStore } from '../../stores/dialog';
import { useThemeStore } from '../../stores/theme';
import HeroMixer from '../../components/home/hero/HeroMixer.vue';
import type { CardChoreoId, TextChoreoId } from '../../components/home/hero/choreo/types';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import StSwitch from './studio/StSwitch.vue';
import StSeg from './studio/StSeg.vue';
import LightCover from './studio/LightCover.vue';
import { themeTransition } from './studio/state';
import { toast } from './studio/toast';

/**
 * 外观：色盘预设（拖拽排序 / 编辑 / 最多 10 组）× 默认模式 × 访客换肤，右侧实时预览；
 * 下方「首页轮播」：Hero 规则 + 文字 / 卡组动效混搭器（HeroMixer，由 Hero 线提供）。
 */
const { t } = useI18n();
const config = useConfigStore();
const theme = useThemeStore();
const dialog = useDialogStore();

type Pick2 = Pick<SiteConfig, 'theme' | 'hero'>;
const cfg = reactive<Pick2>(JSON.parse(JSON.stringify({ theme: config.cfg.theme, hero: config.cfg.hero })));
const snapshot = ref('');
const loaded = ref(false);
const busy = ref(false);
const editing = ref('');
const dirty = computed(() => loaded.value && JSON.stringify(cfg) !== snapshot.value);

async function load(): Promise<void> {
  try {
    const remote = (await adminApi.settings()) as unknown as SiteConfig;
    cfg.theme = JSON.parse(JSON.stringify(remote.theme ?? FALLBACK_CONFIG.theme));
    cfg.hero = { ...FALLBACK_CONFIG.hero, ...(remote.hero ?? {}) };
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
  await nextTick();
  snapshot.value = JSON.stringify(cfg);
  loaded.value = true;
}

async function save(): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  try {
    cfg.theme.displayCount = Math.min(4, Math.max(1, cfg.theme.displayCount));
    await adminApi.saveSettings(JSON.parse(JSON.stringify(cfg)));
    snapshot.value = JSON.stringify(cfg);
    await config.load();
    theme.init();
    toast(t('studio.appearance.saved'));
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  } finally {
    busy.value = false;
  }
}

/* ===== 色盘 ===== */
const presets = computed(() => cfg.theme.presets);
const editingPreset = computed(() => presets.value.find((p) => p.id === editing.value) ?? null);

function lum(hex: string): number {
  const v = hex.replace('#', '');
  const c = [0, 2, 4].map((i) => {
    const x = parseInt(v.slice(i, i + 2), 16) / 255;
    return x <= 0.03928 ? x / 12.92 : ((x + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * c[0] + 0.7152 * c[1] + 0.0722 * c[2];
}

function pickPreset(p: ThemePreset, e: MouseEvent): void {
  cfg.theme.defaultPaletteId = p.id;
  editing.value = editing.value === p.id ? '' : p.id;
  if (theme.allPalettes.some((x) => x.id === p.id) && theme.paletteId !== p.id) {
    themeTransition(e, () => theme.setPalette(p.id, true));
  }
}

function addPreset(): void {
  if (presets.value.length >= 10) return;
  const id = `custom-${Date.now().toString(36)}`;
  presets.value.push({ id, name: t('studio.appearance.customName'), primary: '#00a3a3', primaryDeep: '#00706f' });
  editing.value = id;
}

async function removePreset(p: ThemePreset): Promise<void> {
  if (presets.value.length <= 1) return;
  const ok = await dialog.confirm({
    title: t('studio.appearance.removeTitle', { name: p.name }),
    message: t('studio.appearance.removeBody'),
    confirmText: t('studio.delete'),
    danger: true,
  });
  if (!ok) return;
  cfg.theme.presets = presets.value.filter((x) => x.id !== p.id);
  if (cfg.theme.defaultPaletteId === p.id) cfg.theme.defaultPaletteId = cfg.theme.presets[0].id;
  editing.value = '';
}

/* 拖拽排序（HTML5 DnD + 目标位提示） */
const dragFrom = ref(-1);
const dragOver = ref(-1);
function onDrop(i: number): void {
  const from = dragFrom.value;
  dragFrom.value = -1;
  dragOver.value = -1;
  if (from < 0 || from === i) return;
  const list = cfg.theme.presets;
  const [m] = list.splice(from, 1);
  list.splice(i, 0, m);
}

/* ===== 模式 ===== */
function setMode(m: 'light' | 'dark', e: MouseEvent): void {
  cfg.theme.defaultMode = m;
  if (theme.mode !== m) themeTransition(e, () => theme.setMode(m));
}

const autoSeason = computed({
  get: () => cfg.theme.autoSwitch === 'season',
  set: (v: boolean) => (cfg.theme.autoSwitch = v ? 'season' : 'off'),
});

const pvName = computed(() => {
  const p = presets.value.find((x) => x.id === theme.paletteId);
  return `${p?.name ?? theme.paletteId} · ${theme.mode === 'dark' ? t('studio.dark') : t('studio.light')}`;
});

/* ===== 首页轮播 ===== */
const intervalSec = computed(() => (cfg.hero.intervalMs / 1000).toFixed(1).replace(/\.0$/, ''));
function stepInterval(d: number): void {
  cfg.hero.intervalMs = Math.min(10000, Math.max(1000, cfg.hero.intervalMs + d * 500));
}
function stepCount(d: number): void {
  cfg.hero.count = Math.min(10, Math.max(1, cfg.hero.count + d));
}
function stepDisplay(d: number): void {
  cfg.theme.displayCount = Math.min(Math.min(4, presets.value.length), Math.max(1, cfg.theme.displayCount + d));
}

onBeforeRouteLeave(async () => {
  if (!dirty.value) return true;
  return dialog.confirm({
    title: t('studio.appearance.leaveTitle'),
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
});
onBeforeUnmount(() => window.removeEventListener('keydown', onKey));
</script>

<template>
  <section class="studio view">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.appearance.title') }}</h1>
        <p>{{ t('studio.appearance.desc') }}</p>
      </div>
      <div class="act">
        <button type="button" class="st-btn" :class="dirty ? 'p' : 'g'" :disabled="busy || !dirty" @click="save">
          <SIcon name="check" :size="16" />{{ dirty ? t('studio.save') : t('studio.saved') }}
        </button>
      </div>
    </div>

    <div class="ap">
      <div>
        <h2>{{ t('studio.appearance.palette') }}</h2>
        <p class="desc">{{ t('studio.appearance.paletteDesc', { n: presets.length }) }}</p>
        <div class="swatches">
          <button
            v-for="(p, i) in presets"
            :key="p.id"
            type="button"
            class="swatch"
            :class="{
              on: cfg.theme.defaultPaletteId === p.id,
              ink: lum(p.primary) > 0.4,
              over: dragOver === i && dragFrom !== i,
              back: i >= cfg.theme.displayCount,
            }"
            :style="{ '--c': p.primary, '--d': p.primaryDeep }"
            draggable="true"
            @click="pickPreset(p, $event)"
            @dragstart="dragFrom = i"
            @dragover.prevent="dragOver = i"
            @dragleave="dragOver = -1"
            @drop="onDrop(i)"
          >
            <span class="shine" />
            <span class="ok"><SIcon name="check" :size="16" /></span>
            <span v-if="i >= cfg.theme.displayCount" class="lock">{{ t('studio.appearance.adminOnly') }}</span>
            <b>{{ p.name.split('·')[0].trim() }}</b>
            <small class="mono">{{ (p.name.split('·')[1] ?? '').trim() }} {{ p.primary.toUpperCase() }}</small>
          </button>
          <button v-if="presets.length < 10" type="button" class="swatch add" @click="addPreset">
            <SIcon name="plus" :size="16" />{{ t('studio.appearance.newPalette') }}
          </button>
        </div>

        <Transition name="pe">
          <div v-if="editingPreset" :key="editingPreset.id" class="p-edit">
            <label class="st-field"><input v-model="editingPreset.name" :placeholder="t('studio.appearance.pName')" /></label>
            <label class="color"><input v-model="editingPreset.primary" type="color" /><span><small>{{ t('studio.appearance.primary') }}</small><b class="mono">{{ editingPreset.primary }}</b></span></label>
            <label class="color"><input v-model="editingPreset.primaryDeep" type="color" /><span><small>{{ t('studio.appearance.deep') }}</small><b class="mono">{{ editingPreset.primaryDeep }}</b></span></label>
            <button type="button" class="st-ibtn ring" :disabled="presets.length <= 1" :title="t('studio.delete')" @click="removePreset(editingPreset)"><SIcon name="trash" /></button>
          </div>
        </Transition>

        <div class="opts">
          <div class="st-opt">
            <div>{{ t('studio.appearance.display') }}<small>{{ t('studio.appearance.displaySub') }}</small></div>
            <div class="st-stepper">
              <button type="button" :disabled="cfg.theme.displayCount <= 1" @click="stepDisplay(-1)"><SIcon name="minus" :size="16" /></button>
              <span>{{ cfg.theme.displayCount }}</span>
              <button type="button" :disabled="cfg.theme.displayCount >= Math.min(4, presets.length)" @click="stepDisplay(1)"><SIcon name="plus" :size="16" /></button>
            </div>
          </div>
          <div class="st-opt">
            <div>{{ t('studio.appearance.season') }}<small>{{ t('studio.appearance.seasonSub') }}</small></div>
            <StSwitch v-model="autoSeason" />
          </div>
          <div class="st-opt">
            <div>{{ t('studio.appearance.allowUser') }}<small>{{ t('studio.appearance.allowUserSub') }}</small></div>
            <StSwitch v-model="cfg.theme.allowUserPalette" />
          </div>
        </div>

        <h2>{{ t('studio.appearance.mode') }}</h2>
        <p class="desc">{{ t('studio.appearance.modeDesc') }}</p>
        <div class="modes">
          <button type="button" class="mode-c" :class="{ on: cfg.theme.defaultMode === 'light' }" @click="setMode('light', $event)">
            <div class="mini light"><i class="a" /><i class="b" /><i class="c" /></div>
            <span>{{ t('studio.light') }}</span>
          </button>
          <button type="button" class="mode-c" :class="{ on: cfg.theme.defaultMode === 'dark' }" @click="setMode('dark', $event)">
            <div class="mini dark"><i class="a" /><i class="b" /><i class="c" /></div>
            <span>{{ t('studio.dark') }}</span>
          </button>
        </div>

        <h2>{{ t('studio.appearance.brand') }}</h2>
        <p class="desc">{{ t('studio.appearance.brandDesc') }}</p>
        <div class="brand-row">
          <div class="bs"><i style="--c: var(--red)" /><i style="--c: var(--yellow)" /><i style="--c: var(--blue)" /></div>
          <span>{{ t('studio.appearance.brandNames') }}</span>
          <span class="mono">#FF0032 · #FFB300 · #0078FF</span>
        </div>
      </div>

      <div class="pv-wrap">
        <div class="pv-cap"><span>{{ t('studio.appearance.livePreview') }}</span><span class="mono">{{ pvName }}</span></div>
        <div class="pv">
          <div class="pv-bar"><i /><i /><i /><span>{{ config.cfg.site.title.toLowerCase() }}.blog</span></div>
          <div class="pv-body">
            <div class="pv-nav">
              <div class="cap"><span class="on">{{ t('studio.appearance.pvHome') }}</span><span>{{ t('studio.nav.posts') }}</span><span>{{ t('studio.nav.notes') }}</span><span>{{ t('studio.nav.about') }}</span></div>
              <div class="cap dots"><i v-for="p in presets.slice(0, cfg.theme.displayCount)" :key="p.id" :class="{ on: theme.paletteId === p.id }" :style="{ background: p.primary }" /></div>
            </div>
            <div class="pv-hero">
              <div>
                <span class="tg">{{ t('studio.appearance.pvTag') }}</span>
                <h4>{{ t('studio.appearance.pvTitle') }}</h4>
                <p>{{ t('studio.appearance.pvExcerpt') }}</p>
                <span class="b">{{ t('studio.write.readMore') }}</span>
              </div>
              <LightCover class="hcv" kind="pages" />
            </div>
            <div class="pv-cards">
              <div><LightCover class="ccv" kind="door" /><b>{{ t('studio.appearance.pvCard1') }}</b><small>{{ t('studio.appearance.pvMeta1') }}</small></div>
              <div><LightCover class="ccv" kind="band" /><b>{{ t('studio.appearance.pvCard2') }}</b><small>{{ t('studio.appearance.pvMeta2') }}</small></div>
              <div><LightCover class="ccv" kind="dawn" /><b>{{ t('studio.appearance.pvCard3') }}</b><small>{{ t('studio.appearance.pvMeta3') }}</small></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 首页轮播 -->
    <div class="hero-sec">
      <div class="st-sec-t">
        <h2>{{ t('studio.appearance.hero') }}</h2>
      </div>
      <p class="desc">{{ t('studio.appearance.heroDesc') }}</p>
      <div class="rules">
        <div class="rule">
          <small>{{ t('studio.appearance.heroCount') }}</small>
          <div class="st-stepper">
            <button type="button" :disabled="cfg.hero.count <= 1" @click="stepCount(-1)"><SIcon name="minus" :size="16" /></button>
            <span>{{ cfg.hero.count }}</span>
            <button type="button" :disabled="cfg.hero.count >= 10" @click="stepCount(1)"><SIcon name="plus" :size="16" /></button>
          </div>
        </div>
        <div class="rule">
          <small>{{ t('studio.appearance.heroInterval') }}</small>
          <div class="st-stepper">
            <button type="button" :disabled="cfg.hero.intervalMs <= 1000" @click="stepInterval(-1)"><SIcon name="minus" :size="16" /></button>
            <span>{{ intervalSec }}s</span>
            <button type="button" :disabled="cfg.hero.intervalMs >= 10000" @click="stepInterval(1)"><SIcon name="plus" :size="16" /></button>
          </div>
        </div>
        <div class="rule">
          <small>{{ t('studio.appearance.heroPinned') }}</small>
          <StSeg
            v-model="cfg.hero.pinnedRule"
            :options="[
              { value: 'pinned-first', label: t('studio.appearance.pinnedFirst') },
              { value: 'ignore', label: t('studio.appearance.pinnedIgnore') },
            ]"
          />
        </div>
      </div>
      <div class="mixer-wrap">
        <HeroMixer
          :text="cfg.hero.textAnim"
          :card="cfg.hero.cardAnim"
          @update:text="(v: string) => (cfg.hero.textAnim = v as TextChoreoId)"
          @update:card="(v: string) => (cfg.hero.cardAnim = v as CardChoreoId)"
        />
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

h2 { font: 600 18px var(--font-serif); margin: 0 0 6px; }
.desc { font-size: 13.5px; color: var(--ink-3); margin: 0 0 18px; }

.ap {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 420px;
  gap: 48px;
  align-items: start;
}

.swatches {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 14px;
}

.swatch {
  position: relative;
  height: 118px;
  border-radius: 18px;
  overflow: hidden;
  text-align: left;
  padding: 14px 16px;
  color: #fff;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  cursor: pointer;
  background:
    radial-gradient(90% 80% at 85% 0%, color-mix(in oklab, var(--c) 70%, #fff), transparent 60%),
    linear-gradient(150deg, var(--c), var(--d));
  box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.06) inset, 0 10px 24px -14px var(--d);
  transition: transform var(--dur) var(--ease-spring), box-shadow var(--dur), opacity var(--dur);

  &:hover { transform: translateY(-3px); }
  &:active { transform: scale(0.97); }
  &.ink { color: #241a00; }
  &.back { opacity: 0.72; }
  &.over { transform: scale(1.04); box-shadow: 0 0 0 2px var(--paper), 0 0 0 4px var(--ink-4); }

  b { font: 700 24px/1 var(--font-serif); }
  small { font-size: 11px; opacity: 0.82; margin-top: 6px; display: block; letter-spacing: 0.02em; }

  .ok {
    position: absolute;
    right: 12px;
    top: 12px;
    width: 26px;
    height: 26px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.95);
    color: #1e1c19;
    display: grid;
    place-items: center;
    transform: scale(0);
    transition: transform var(--dur) var(--ease-bounce);
  }

  .lock {
    position: absolute;
    left: 12px;
    top: 12px;
    font-size: 11px;
    padding: 1px 8px;
    border-radius: 8px;
    background: rgba(0, 0, 0, 0.22);
    color: #fff;
  }

  &.on { box-shadow: 0 0 0 3px var(--paper), 0 0 0 5px var(--c), 0 16px 30px -14px var(--d); }
  &.on .ok { transform: scale(1); }

  .shine {
    position: absolute;
    inset: 0;
    background: linear-gradient(115deg, transparent 35%, rgba(255, 255, 255, 0.25) 50%, transparent 65%);
    transform: translateX(-100%);
    transition: transform 0.8s var(--ease-out);
  }

  &:hover .shine { transform: translateX(100%); }

  &.add {
    background: var(--well);
    color: var(--ink-3);
    box-shadow: 0 0 0 1.5px var(--line-2) inset;
    align-items: center;
    justify-content: center;
    gap: 8px;
    flex-direction: row;
    font-size: 14px;

    &:hover { color: var(--primary-ink); }
  }
}

.p-edit {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto auto;
  gap: 10px;
  align-items: center;
  padding: 12px;
  border-radius: 16px;
  background: var(--well);
  margin-bottom: 10px;

  .st-field { background: var(--paper); }

  .color {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 10px 4px 4px;
    border-radius: 11px;
    background: var(--paper);
    box-shadow: 0 0 0 1px var(--line) inset;
    cursor: pointer;

    input {
      width: 30px;
      height: 30px;
      padding: 0;
      border: 0;
      border-radius: 8px;
      background: none;
      cursor: pointer;
    }

    input::-webkit-color-swatch-wrapper { padding: 0; }
    input::-webkit-color-swatch { border: 0; border-radius: 8px; }

    small { display: block; font-size: 11px; color: var(--ink-3); line-height: 1.2; }
    b { font-size: 12px; font-weight: 500; }
  }
}

.pe-enter-active, .pe-leave-active { transition: all var(--dur) var(--ease-out); }
.pe-enter-from, .pe-leave-to { opacity: 0; transform: translateY(-6px); }

.opts { margin: 8px 0 40px; }

.modes {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 40px;
}

.mode-c {
  border-radius: 16px;
  padding: 10px;
  text-align: left;
  box-shadow: 0 0 0 1px var(--line-2) inset;
  transition: all var(--dur-fast);

  &:active { transform: scale(0.97); }
  &.on { box-shadow: 0 0 0 2px var(--primary) inset; background: var(--primary-soft); }

  span { font-size: 13.5px; font-weight: 500; padding: 0 4px; }

  .mini {
    height: 74px;
    border-radius: 10px;
    margin-bottom: 10px;
    position: relative;
    overflow: hidden;

    i { position: absolute; border-radius: 4px; }
    .a { left: 10px; top: 10px; right: 40%; height: 8px; }
    .b { left: 10px; top: 26px; width: 44%; height: 34px; }
    .c { right: 10px; top: 26px; width: 34%; height: 34px; background: var(--primary); }

    &.light { background: color-mix(in oklab, var(--primary) 5%, #f5f6f8); .a { background: #e3e5ea; } .b { background: #fff; box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.05); } }
    &.dark { background: color-mix(in oklab, var(--primary) 8%, #0b1220); .a { background: #1a2438; } .b { background: #101a2e; } }
  }
}

.brand-row {
  display: flex;
  gap: 18px;
  align-items: center;
  padding: 16px 18px;
  border-radius: 14px;
  background: var(--well);
  font-size: 13px;
  color: var(--ink-2);

  .bs { display: flex; gap: 6px; }
  .bs i { width: 22px; height: 22px; border-radius: 7px; background: var(--c); box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.08) inset; }
  .mono { color: var(--ink-3); font-size: 11.5px; }
}

/* ---------- 预览 ---------- */
.pv-wrap { position: sticky; top: 24px; }

.pv-cap {
  display: flex;
  justify-content: space-between;
  font-size: 12.5px;
  color: var(--ink-3);
  margin-bottom: 10px;
}

.pv {
  --pv-bg: color-mix(in oklab, var(--primary) 4%, #f5f6f8);
  --pv-card: #fff;
  --pv-ink: #141821;
  --pv-ink2: #6b7280;
  --pv-line: rgba(0, 0, 0, 0.06);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: var(--sh-pop);
  background: var(--pv-bg);
  transition: background var(--dur-slow);
}

:root[data-mode='dark'] .pv {
  --pv-bg: color-mix(in oklab, var(--primary) 6%, #0b1220);
  --pv-card: color-mix(in oklab, var(--primary) 6%, #101a2e);
  --pv-ink: #e8edf6;
  --pv-ink2: #7c89a3;
  --pv-line: rgba(255, 255, 255, 0.06);
}

.pv-bar {
  height: 30px;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 12px;
  background: var(--well-2);

  i { width: 9px; height: 9px; border-radius: 50%; background: var(--line-3); }
  span { margin: 0 auto; font: 500 10.5px var(--font-mono); color: var(--ink-3); background: var(--paper); padding: 2px 30px; border-radius: 6px; }
}

.pv-body { padding: 14px 18px 20px; color: var(--pv-ink); }

.pv-nav {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-bottom: 16px;

  .cap {
    display: flex;
    gap: 2px;
    padding: 3px;
    border-radius: 99px;
    background: color-mix(in srgb, var(--pv-card) 80%, transparent);
    box-shadow: 0 0 0 1px var(--pv-line);

    span { font-size: 9px; padding: 3px 9px; border-radius: 99px; color: var(--pv-ink2); }
    span.on { background: var(--primary); color: var(--on-primary); }
  }

  .dots i {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin: 3px 2px;
    display: block;

    &.on { box-shadow: 0 0 0 1.5px var(--pv-card), 0 0 0 3px currentColor; color: var(--ink-4); }
  }
}

.pv-hero {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  align-items: center;
  padding: 16px;
  border-radius: 12px;
  background: var(--pv-card);
  box-shadow: 0 0 0 1px var(--pv-line);
  margin-bottom: 12px;

  .tg { display: inline-block; font-size: 8px; padding: 2px 7px; border-radius: 9px; color: var(--primary-ink); background: var(--primary-soft); margin-bottom: 6px; }
  h4 { font: 700 15px/1.35 var(--font-serif); margin: 0 0 6px; }
  p { font-size: 8.5px; color: var(--pv-ink2); margin: 0 0 10px; line-height: 1.6; }
  .b { display: inline-block; font-size: 8.5px; padding: 4px 10px; border-radius: 6px; background: var(--primary); color: var(--on-primary); box-shadow: 0 4px 10px -4px var(--primary); }
  .hcv { aspect-ratio: 4 / 3; border-radius: 8px; transform: perspective(500px) rotateY(-12deg); box-shadow: 0 10px 20px -8px rgba(0, 0, 0, 0.4); }
}

.pv-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;

  > div { border-radius: 8px; background: var(--pv-card); box-shadow: 0 0 0 1px var(--pv-line); padding: 5px; }
  .ccv { aspect-ratio: 16 / 10; border-radius: 5px; margin-bottom: 5px; }
  b { display: block; font: 600 8.5px/1.4 var(--font-serif); padding: 0 2px; }
  small { display: block; font-size: 7px; color: var(--pv-ink2); padding: 2px; }
}

/* ---------- 首页轮播 ---------- */
.hero-sec {
  margin-top: 64px;
  padding-top: 40px;
  border-top: 1px solid var(--line);
}

.rules {
  display: flex;
  gap: 36px;
  flex-wrap: wrap;
  align-items: flex-end;
  padding: 18px 22px;
  border-radius: 18px;
  background: var(--well);
  margin-bottom: 24px;

  .rule small { display: block; font-size: 12.5px; color: var(--ink-3); margin-bottom: 8px; }
  .st-stepper { background: var(--paper); }
}

.mixer-wrap {
  border-radius: 20px;
  padding: 18px;
  box-shadow: 0 0 0 1px var(--line-2);
}

@media (max-width: 1180px) {
  .view { padding: 40px 36px 80px; }
  .ap { grid-template-columns: 1fr; }
  .pv-wrap { position: static; }
}
</style>
