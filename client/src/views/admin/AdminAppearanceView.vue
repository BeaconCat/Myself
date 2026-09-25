<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';
import { FALLBACK_CONFIG, useConfigStore, type SiteConfig, type ThemePreset } from '../../stores/config';
import { useDialogStore } from '../../stores/dialog';
import { applyRadius, useThemeStore } from '../../stores/theme';
import { derivePalette } from '../../themes/derive';
import HeroMixer from '../../components/home/hero/HeroMixer.vue';
import type { CardChoreoId, RotateChoreoId, TextChoreoId } from '../../components/home/hero/choreo/types';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import StSwitch from './studio/StSwitch.vue';
import StSeg from './studio/StSeg.vue';
import LightCover from './studio/LightCover.vue';
import { themeTransition, settle, stableJson } from './studio/state';
import { toast } from './studio/toast';

/**
 * 外观：色盘预设（拖拽排序 / 编辑 / 最多 10 组）× 默认模式 × 访客换肤 × 全局圆角，右侧实时预览；
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
const dirty = computed(() => loaded.value && stableJson(cfg) !== snapshot.value);

async function load(): Promise<void> {
  try {
    const remote = (await adminApi.settings()) as unknown as SiteConfig;
    cfg.theme = JSON.parse(JSON.stringify(remote.theme ?? FALLBACK_CONFIG.theme));
    cfg.hero = { ...FALLBACK_CONFIG.hero, ...(remote.hero ?? {}) };
    cfg.theme.radius = clampRadius(cfg.theme.radius ?? FALLBACK_CONFIG.theme.radius ?? 10);
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
  await settle();
  snapshot.value = stableJson(cfg);
  loaded.value = true;
}

async function save(): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  try {
    cfg.theme.displayCount = Math.min(4, Math.max(1, cfg.theme.displayCount));
    await adminApi.saveSettings(JSON.parse(JSON.stringify(cfg)));
    snapshot.value = stableJson(cfg);
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

/** 色卡文字色：与全局主按钮同一套对比度派生（秋黄等亮色 → 深色字） */
function swatchInk(p: ThemePreset): string {
  try {
    return derivePalette(p).light.onSolid;
  } catch {
    return '#fff';
  }
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

/* ===== 圆角：滑块 0–24 + 预设，实时写入 --r-base；未保存离开时恢复 ===== */
const RADIUS_PRESETS = [
  { v: 4, key: 'radiusSharp' },
  { v: 10, key: 'radiusStandard' },
  { v: 16, key: 'radiusRound' },
] as const;

function clampRadius(v: number): number {
  const n = Math.round(Number(v));
  return Number.isFinite(n) ? Math.min(24, Math.max(0, n)) : 10;
}

const radius = computed({
  get: () => cfg.theme.radius ?? 10,
  set: (v: number) => (cfg.theme.radius = clampRadius(v)),
});
const radiusPct = computed(() => `${(radius.value / 24) * 100}%`);

watch(radius, (v) => applyRadius(v));

/** 恢复为已保存（站点配置）的圆角 */
function restoreRadius(): void {
  applyRadius(config.cfg.theme.radius ?? 10);
}

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
  const ok = await dialog.confirm({
    title: t('studio.appearance.leaveTitle'),
    message: t('studio.appearance.leaveBody'),
    confirmText: t('studio.write.leave'),
    danger: true,
  });
  if (ok) restoreRadius();
  return ok;
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
  restoreRadius();
});
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
          <SIcon name="check" :size="18" />{{ dirty ? t('studio.save') : t('studio.saved') }}
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
              over: dragOver === i && dragFrom !== i,
              back: i >= cfg.theme.displayCount,
            }"
            :style="{ '--c': p.primary, '--d': p.primaryDeep, '--fg': swatchInk(p) }"
            draggable="true"
            @click="pickPreset(p, $event)"
            @dragstart="dragFrom = i"
            @dragover.prevent="dragOver = i"
            @dragleave="dragOver = -1"
            @drop="onDrop(i)"
          >
            <span class="ok"><SIcon name="check" :size="18" /></span>
            <span v-if="i >= cfg.theme.displayCount" class="lock">{{ t('studio.appearance.adminOnly') }}</span>
            <b>{{ p.name.split('·')[0].trim() }}</b>
            <small><span>{{ (p.name.split('·')[1] ?? '').trim() }}</span><span class="mono">{{ p.primary.toUpperCase() }}</span></small>
          </button>
          <button v-if="presets.length < 10" type="button" class="swatch add" @click="addPreset">
            <SIcon name="plus" :size="18" />{{ t('studio.appearance.newPalette') }}
          </button>
        </div>

        <Transition name="pe">
          <div v-if="editingPreset" class="p-edit">
            <label class="st-field"><input v-model="editingPreset.name" :placeholder="t('studio.appearance.pName')" /></label>
            <label class="color"><i class="sw" :style="{ background: editingPreset.primary }" /><input v-model="editingPreset.primary" type="color" /><span><small>{{ t('studio.appearance.primary') }}</small><b class="mono">{{ editingPreset.primary }}</b></span></label>
            <label class="color"><i class="sw" :style="{ background: editingPreset.primaryDeep }" /><input v-model="editingPreset.primaryDeep" type="color" /><span><small>{{ t('studio.appearance.deep') }}</small><b class="mono">{{ editingPreset.primaryDeep }}</b></span></label>
            <button type="button" class="st-ibtn ring" :disabled="presets.length <= 1" :title="t('studio.delete')" @click="removePreset(editingPreset)"><SIcon name="trash" /></button>
          </div>
        </Transition>

        <div class="opts">
          <div class="st-opt">
            <div>{{ t('studio.appearance.display') }}<small>{{ t('studio.appearance.displaySub') }}</small></div>
            <div class="st-stepper">
              <button type="button" :disabled="cfg.theme.displayCount <= 1" @click="stepDisplay(-1)"><SIcon name="minus" :size="18" /></button>
              <span>{{ cfg.theme.displayCount }}</span>
              <button type="button" :disabled="cfg.theme.displayCount >= Math.min(4, presets.length)" @click="stepDisplay(1)"><SIcon name="plus" :size="18" /></button>
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

        <div class="pair">
        <div>
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

        </div>
        <div>
        <h2>{{ t('studio.appearance.radius') }}</h2>
        <p class="desc">{{ t('studio.appearance.radiusDesc') }}</p>
        <div class="radius">
          <div class="r-top">
            <span>{{ t('studio.appearance.radiusBase') }}</span>
            <b class="mono">{{ radius }}px</b>
          </div>
          <input
            v-model.number="radius"
            class="r-range"
            type="range"
            min="0"
            max="24"
            step="1"
            :style="{ '--p': radiusPct }"
            :aria-label="t('studio.appearance.radiusBase')"
          />
          <div class="r-presets">
            <button
              v-for="rp in RADIUS_PRESETS"
              :key="rp.v"
              type="button"
              class="st-chip"
              :class="{ on: radius === rp.v }"
              @click="radius = rp.v"
            >
              {{ t(`studio.appearance.${rp.key}`) }}<span class="n">{{ rp.v }}</span>
            </button>
          </div>
        </div>
        </div>
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

        <div class="pv-cap demo-cap"><span>{{ t('studio.appearance.radiusDemo') }}</span><span class="mono">{{ radius }}px</span></div>
        <div class="demo">
          <div class="d-row">
            <button type="button" class="st-btn p sm" tabindex="-1">{{ t('studio.appearance.radiusBtn') }}</button>
            <button type="button" class="st-btn g sm" tabindex="-1">{{ t('studio.appearance.radiusBtn2') }}</button>
            <span class="st-chip on">{{ t('studio.appearance.radiusChipOn') }}</span>
            <span class="st-chip">{{ t('studio.appearance.radiusChip') }}</span>
          </div>
          <label class="st-field"><SIcon name="search" :size="18" /><input :placeholder="t('studio.appearance.radiusInput')" /></label>
          <div class="d-card">
            <LightCover class="d-cv" kind="door" />
            <div><b>{{ t('studio.appearance.radiusCard') }}</b><small>{{ t('studio.appearance.radiusCardMeta') }}</small></div>
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
            <button type="button" :disabled="cfg.hero.count <= 1" @click="stepCount(-1)"><SIcon name="minus" :size="18" /></button>
            <span>{{ cfg.hero.count }}</span>
            <button type="button" :disabled="cfg.hero.count >= 10" @click="stepCount(1)"><SIcon name="plus" :size="18" /></button>
          </div>
        </div>
        <div class="rule">
          <small>{{ t('studio.appearance.heroInterval') }}</small>
          <div class="st-stepper">
            <button type="button" :disabled="cfg.hero.intervalMs <= 1000" @click="stepInterval(-1)"><SIcon name="minus" :size="18" /></button>
            <span>{{ intervalSec }}s</span>
            <button type="button" :disabled="cfg.hero.intervalMs >= 10000" @click="stepInterval(1)"><SIcon name="plus" :size="18" /></button>
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
          :rotate="cfg.hero.rotateAnim"
          @update:rotate="(v: string) => (cfg.hero.rotateAnim = v as RotateChoreoId)"
        />
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

h2 { font: 700 22px/1.3 var(--font-serif); margin: 0 0 4px; }
.desc { font-size: 14px; color: var(--st-ink-3); margin: 0 0 14px; }

.ap {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 440px;
  gap: 32px;
  align-items: start;
}

.swatches {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 10px;
  margin-bottom: 14px;
}

.swatch {
  position: relative;
  height: 104px;
  border-radius: var(--r-lg);
  overflow: hidden;
  text-align: left;
  padding: 12px 14px;
  color: var(--fg, #fff);
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  cursor: pointer;
  /* 色卡 = 色样本身：纯色 + 顶部内高光，中性阴影（不做主色渐变 / 同色投影） */
  background: var(--c);
  box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.22), inset 0 0 0 0.5px rgb(0 0 0 / 0.12), var(--shadow-card);
  transition: transform var(--dur) var(--ease-spring), box-shadow var(--dur), opacity var(--dur);

  &:hover { transform: translateY(-3px); }
  &:active { transform: scale(0.97); }
  &.back { opacity: 0.72; }
  &.over { transform: scale(1.04); box-shadow: 0 0 0 2px var(--paper), 0 0 0 4px var(--st-ink-4); }

  b { font: 700 24px/1 var(--font-serif); }
  small {
    display: flex;
    flex-direction: column;
    margin-top: 6px;
    font-size: 12px;
    line-height: 1.35;
    opacity: 0.85;
    letter-spacing: 0.02em;

    span { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  }

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
    border-radius: var(--r-xs);
    background: rgba(0, 0, 0, 0.22);
    color: #fff;
  }

  /* 选中：中性双环（纸面间隙 + 正文色 1.5px），与前台色盘选中同构 */
  &.on { box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.22), 0 0 0 3px var(--paper), 0 0 0 4.5px var(--st-ink), var(--shadow-card); }
  &.on .ok { transform: scale(1); }

  &.add {
    background: var(--well);
    color: var(--st-ink-3);
    box-shadow: 0 0 0 1.5px var(--line-2) inset;
    align-items: center;
    justify-content: center;
    gap: 6px;
    flex-direction: column;
    font-size: 13.5px;

    &:hover { color: var(--ink); }
  }
}

.p-edit {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto auto;
  gap: 10px;
  align-items: center;
  padding: 12px;
  border-radius: var(--r-md);
  background: var(--well);
  margin-bottom: 10px;

  .st-field { background: var(--paper); }

  /* 色块容器撑满行高，四边等距；色块为正方形，颜色切换走过渡（原生取色器透明覆盖在色块上） */
  .color {
    --pad: 6px;

    position: relative;
    align-self: stretch;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: var(--pad) 14px var(--pad) var(--pad);
    border-radius: var(--r-sm);
    background: var(--paper);
    box-shadow: 0 0 0 1px var(--line) inset;
    cursor: pointer;

    .sw {
      height: 100%;
      aspect-ratio: 1;
      flex: none;
      border-radius: max(2px, calc(var(--r-sm) - var(--pad)));
      box-shadow: 0 0 0 1px rgb(0 0 0 / 0.08) inset;
      transition: background-color var(--dur) var(--ease-out);
    }

    input {
      position: absolute;
      inset: var(--pad) auto var(--pad) var(--pad);
      aspect-ratio: 1;
      height: calc(100% - var(--pad) * 2);
      padding: 0;
      border: 0;
      opacity: 0;
      cursor: pointer;
    }

    span { transition: opacity var(--dur-fast); }

    small { display: block; font-size: 12px; color: var(--st-ink-3); line-height: 1.2; }
    b { font-size: 13px; font-weight: 500; }
  }
}

.pe-enter-active, .pe-leave-active { transition: all var(--dur) var(--ease-out); }
.pe-enter-from, .pe-leave-to { opacity: 0; transform: translateY(-6px); }

.opts { margin: 4px 0 28px; }

/* 默认模式 ｜ 圆角 并排等高 */
.pair {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-items: stretch;
  gap: 20px;

  > div { display: flex; flex-direction: column; min-width: 0; }
  .desc { min-height: 2.9em; }
}

.modes {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  flex: 1;
}

.mode-c {
  border-radius: var(--r-md);
  padding: 10px;
  text-align: left;
  box-shadow: 0 0 0 1px var(--line-2) inset;
  transition: all var(--dur-fast);

  &:hover:not(.on) { background: var(--hover); }
  &:active { transform: scale(0.97); }
  &.on { background: var(--lift); box-shadow: var(--lift-shadow); }

  span { font-size: 14px; font-weight: 500; padding: 0 4px; }

  .mini {
    height: 68px;
    border-radius: var(--r-sm);
    margin-bottom: 10px;
    position: relative;
    overflow: hidden;

    i { position: absolute; border-radius: var(--r-xs); }
    .a { left: 10px; top: 10px; right: 40%; height: 8px; }
    .b { left: 10px; top: 26px; width: 44%; height: 34px; }
    .c { right: 10px; top: 26px; width: 34%; height: 34px; background: var(--solid); }

    &.light { background: color-mix(in oklab, var(--primary) 5%, #f5f6f8); .a { background: #e3e5ea; } .b { background: #fff; box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.05); } }
    &.dark { background: color-mix(in oklab, var(--primary) 8%, #0b1220); .a { background: #1a2438; } .b { background: #101a2e; } }
  }
}

/* ---------- 圆角 ---------- */
.radius {
  flex: 1;
  padding: 14px 16px;
  border-radius: var(--r-md);
  background: var(--well);

  .r-top {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    font-size: 13.5px;
    color: var(--st-ink-2);

    b { font: 600 22px/1 var(--font-mono); color: var(--st-ink); }
  }

  .r-presets { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 10px; }
}

.r-range {
  --p: 42%;
  width: 100%;
  height: 22px;
  margin: 8px 0 0;
  background: none;
  appearance: none;
  cursor: pointer;

  &::-webkit-slider-runnable-track {
    height: 4px;
    border-radius: var(--r-pill);
    background: linear-gradient(90deg, var(--st-ink-2) var(--p), var(--line-2) var(--p));
  }

  &::-moz-range-track { height: 4px; border-radius: var(--r-pill); background: var(--line-2); }
  &::-moz-range-progress { height: 4px; border-radius: var(--r-pill); background: var(--st-ink-2); }

  &::-webkit-slider-thumb {
    appearance: none;
    width: 18px;
    height: 18px;
    margin-top: -7px;
    border-radius: 50%;
    background: var(--elev);
    box-shadow: 0 0 0 0.5px rgb(16 24 40 / 0.16), 0 1px 3px rgb(16 24 40 / 0.24);
    transition: transform var(--dur-fast) var(--ease-spring);
  }

  &::-moz-range-thumb {
    width: 18px;
    height: 18px;
    border: 0;
    border-radius: 50%;
    background: var(--elev);
    box-shadow: 0 0 0 0.5px rgb(16 24 40 / 0.16), 0 1px 3px rgb(16 24 40 / 0.24);
  }

  &:active::-webkit-slider-thumb { transform: scale(1.15); }
  &:focus-visible { outline: 0; }
  &:focus-visible::-webkit-slider-thumb { box-shadow: var(--focus); }
}


/* ---------- 预览 ---------- */
.pv-wrap { position: sticky; top: 24px; }

.pv-cap {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: var(--st-ink-3);
  margin-bottom: 10px;
}

.pv {
  --pv-bg: color-mix(in oklab, var(--primary) 4%, #f5f6f8);
  --pv-card: #fff;
  --pv-ink: #141821;
  --pv-ink2: #6b7280;
  --pv-line: rgba(0, 0, 0, 0.06);
  border-radius: var(--r-md);
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
  span { margin: 0 auto; font: 500 10.5px var(--font-mono); color: var(--st-ink-3); background: var(--paper); padding: 2px 30px; border-radius: var(--r-xs); }
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
    border-radius: var(--r-pill);
    background: color-mix(in srgb, var(--pv-card) 80%, transparent);
    box-shadow: 0 0 0 1px var(--pv-line);

    span { font-size: 9px; padding: 3px 9px; border-radius: var(--r-pill); color: var(--pv-ink2); }
    span.on { background: var(--lift); box-shadow: var(--lift-shadow); color: var(--pv-ink); font-weight: 500; }
  }

  .dots i {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin: 3px 2px;
    display: block;

    &.on { box-shadow: 0 0 0 1.5px var(--pv-card), 0 0 0 2.5px var(--pv-ink); }
  }
}

.pv-hero {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  align-items: center;
  padding: 16px;
  border-radius: var(--r-md);
  background: var(--pv-card);
  box-shadow: 0 0 0 1px var(--pv-line);
  margin-bottom: 12px;

  /* 标签：「# 名称」纯文字，# 为三级灰 */
  .tg { display: inline-block; font-size: 8.5px; color: var(--pv-ink2); margin-bottom: 6px; }
  .tg::before { content: '#'; margin-right: 2px; font-family: var(--font-mono); opacity: 0.6; }
  h4 { font: 700 15px/1.35 var(--font-serif); margin: 0 0 6px; }
  p { font-size: 8.5px; color: var(--pv-ink2); margin: 0 0 10px; line-height: 1.6; }
  .b { display: inline-block; font-size: 8.5px; padding: 4px 10px; border-radius: var(--r-pill); background: var(--solid); color: var(--on-solid); box-shadow: var(--btn-shadow); }
  .hcv { aspect-ratio: 4 / 3; border-radius: var(--r-xs); transform: perspective(500px) rotateY(-12deg); box-shadow: 0 10px 20px -8px rgba(0, 0, 0, 0.4); }
}

.pv-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;

  > div { border-radius: var(--r-sm); background: var(--pv-card); box-shadow: 0 0 0 1px var(--pv-line); padding: 5px; }
  .ccv { aspect-ratio: 16 / 10; border-radius: var(--r-xs); margin-bottom: 5px; }
  b { display: block; font: 600 8.5px/1.4 var(--font-serif); padding: 0 2px; }
  small { display: block; font-size: 7px; color: var(--pv-ink2); padding: 2px; }
}

/* ---------- 控件示意（随圆角实时变化） ---------- */
.demo-cap { margin-top: 18px; }

.demo {
  display: grid;
  gap: 12px;
  padding: 16px;
  border-radius: var(--r-lg);
  background: var(--well);
  box-shadow: 0 0 0 1px var(--line) inset;

  .d-row { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
  .st-chip { cursor: default; }
  .st-field { background: var(--paper); }

  .d-card {
    display: grid;
    grid-template-columns: 96px minmax(0, 1fr);
    gap: 12px;
    align-items: center;
    padding: 8px;
    border-radius: var(--r-lg);
    background: var(--paper);
    box-shadow: var(--shadow-card);

    b { display: block; font: 600 14px/1.4 var(--font-serif); }
    small { font-size: 12px; color: var(--st-ink-3); }
  }

  .d-cv { aspect-ratio: 16 / 10; border-radius: var(--r-md); }
}

/* ---------- 首页轮播 ---------- */
.hero-sec {
  margin-top: 40px;
  padding-top: 28px;
  border-top: 1px solid var(--line);

  > .st-sec-t { margin-bottom: 4px; }
}

.rules {
  display: flex;
  gap: 32px;
  flex-wrap: wrap;
  align-items: flex-end;
  padding: 16px 20px;
  border-radius: var(--r-lg);
  background: var(--well);
  margin-bottom: 16px;

  .rule small { display: block; font-size: 13px; color: var(--st-ink-3); margin-bottom: 8px; }
  .st-stepper { background: var(--paper); }
}

.mixer-wrap {
  border-radius: var(--r-lg);
  padding: 16px;
  box-shadow: 0 0 0 1px var(--line-2);
}

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
  .ap { grid-template-columns: 1fr; }
  .pv-wrap { position: static; }
}
</style>
