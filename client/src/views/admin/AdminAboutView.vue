<script setup lang="ts">
import { settle, stableJson } from './studio/state';
import { migrateModules } from '../../about/migrate';
import { injectIdentity, normalizeIdentity, plainText } from '../../about/identity';
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { onBeforeRouteLeave } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';
import { useConfigStore, type AboutModule, type SiteConfig } from '../../stores/config';
import { useDialogStore } from '../../stores/dialog';
import { createModule, metaOf, spanOf, titleOf } from '../../about/registry';
import type { Span } from '../../about/types';
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
  normalizeIdentity(about);
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

/* ===== 模块 ===== */
/** 卡片摘要：身份区 / 格言的内容来自站点身份，先注入再取摘要 */
function summaryOf(mod: AboutModule): string {
  const meta = metaOf(mod.type);
  return plainText(meta?.summary(injectIdentity(mod, about).data)) || meta?.desc || '';
}

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
    ...spanItems(mod),
    { icon: 'trash', label: t('studio.delete'), danger: true, divider: true, run: () => void removeModule(mod) },
  ];
}

/* ===== 网格排版：拖动排序（幽灵卡 + 占位）与右缘拖柄改宽 ===== */
const grid = ref<HTMLElement | null>(null);
const draggingId = ref('');
const resizingId = ref('');
const ghost = ref<HTMLElement | null>(null);
const ghostBox = reactive({ w: 0, h: 0, x: 0, y: 0 });

const draggingMod = computed(() => about.modules.find((m) => m.id === draggingId.value) ?? null);

function cards(): HTMLElement[] {
  return grid.value ? [...grid.value.querySelectorAll<HTMLElement>(':scope > [data-id]')] : [];
}

/** 布局矩形：offset 系坐标不受 FLIP 动画的 transform 影响，命中检测因此不会来回抖 */
function layoutRect(el: HTMLElement): { l: number; t: number; r: number; b: number } {
  return { l: el.offsetLeft, t: el.offsetTop, r: el.offsetLeft + el.offsetWidth, b: el.offsetTop + el.offsetHeight };
}

/** FLIP：记录旧位置 → 变更 → 从旧位置补间到新位置 */
async function flip(change: () => void): Promise<void> {
  const first = new Map(cards().map((c) => [c.dataset.id!, c.getBoundingClientRect()]));
  change();
  await nextTick();
  for (const c of cards()) {
    const f = first.get(c.dataset.id!);
    if (!f) continue;
    const l = c.getBoundingClientRect();
    const dx = f.left - l.left;
    const dy = f.top - l.top;
    if (Math.abs(dx) < 0.5 && Math.abs(dy) < 0.5) continue;
    c.getAnimations().forEach((an) => an.cancel());
    c.animate([{ transform: `translate(${dx}px, ${dy}px)` }, { transform: 'none' }], { duration: 380, easing: 'cubic-bezier(.2,.8,.3,1)' });
  }
}

/** 最近的滚动容器（后台主面板自己滚动，不一定是 window） */
function scroller(el: HTMLElement | null): HTMLElement {
  for (let n = el?.parentElement; n; n = n.parentElement) {
    const oy = getComputedStyle(n).overflowY;
    if ((oy === 'auto' || oy === 'scroll') && n.scrollHeight > n.clientHeight) return n;
  }
  return document.scrollingElement as HTMLElement;
}

interface DragState { gx: number; gy: number; x: number; y: number; scroll: HTMLElement; raf: number; last: number }
let drag: DragState | null = null;

function onGripDown(e: PointerEvent, mod: AboutModule): void {
  const el = (e.currentTarget as HTMLElement).closest<HTMLElement>('.mod');
  if (!el || e.button !== 0) return;
  e.preventDefault();
  expanded.value = '';
  const r = el.getBoundingClientRect();
  ghostBox.w = r.width;
  ghostBox.h = r.height;
  ghostBox.x = r.left;
  ghostBox.y = r.top;
  drag = { gx: e.clientX - r.left, gy: e.clientY - r.top, x: e.clientX, y: e.clientY, scroll: scroller(grid.value), raf: 0, last: 0 };
  // 拖动期间网格改为按顺序排列（不回填），插入位置才可预期；切换本身走 FLIP
  void flip(() => { draggingId.value = mod.id; });
  window.addEventListener('pointermove', onMove);
  window.addEventListener('pointerup', onUp);
  window.addEventListener('pointercancel', onUp);
  drag.raf = requestAnimationFrame(tick);
}

function onMove(e: PointerEvent): void {
  if (!drag) return;
  drag.x = e.clientX;
  drag.y = e.clientY;
}

/** 每帧：幽灵卡跟手、贴边自动滚动、按指针落在哪张卡的左 / 右半区决定插到它前 / 后 */
function tick(): void {
  if (!drag) return;
  const { x, y } = drag;
  if (ghost.value) ghost.value.style.transform = `translate(${x - drag.gx}px, ${y - drag.gy}px) rotate(-0.8deg)`;

  const sc = drag.scroll;
  const sr = sc === document.scrollingElement ? { top: 0, bottom: window.innerHeight } : sc.getBoundingClientRect();
  const edge = 80;
  if (y < sr.top + edge) sc.scrollTop -= Math.ceil((sr.top + edge - y) / 5);
  else if (y > sr.bottom - edge) sc.scrollTop += Math.ceil((y - (sr.bottom - edge)) / 5);

  const now = performance.now();
  if (grid.value && now - drag.last > 80) {
    const gr = grid.value.getBoundingClientRect();
    const px = x - gr.left;
    const py = y - gr.top;
    const from = about.modules.findIndex((m) => m.id === draggingId.value);
    for (const c of cards()) {
      const id = c.dataset.id!;
      if (id === draggingId.value) continue;
      const rc = layoutRect(c);
      if (px < rc.l || px > rc.r || py < rc.t || py > rc.b) continue;
      let to = about.modules.findIndex((m) => m.id === id);
      const after = px > (rc.l + rc.r) / 2;
      if (after && to < from) to += 1;
      if (!after && to > from) to -= 1;
      if (to !== from && from >= 0 && to >= 0) {
        drag.last = now;
        void flip(() => {
          const [m] = about.modules.splice(from, 1);
          about.modules.splice(to, 0, m);
        });
      }
      break;
    }
  }
  drag.raf = requestAnimationFrame(tick);
}

async function onUp(): Promise<void> {
  window.removeEventListener('pointermove', onMove);
  window.removeEventListener('pointerup', onUp);
  window.removeEventListener('pointercancel', onUp);
  if (!drag) return;
  cancelAnimationFrame(drag.raf);
  drag = null;
  // 幽灵卡落回占位框，再换回真卡
  const g = ghost.value;
  const slot = grid.value?.querySelector<HTMLElement>('.slot');
  if (g && slot) {
    const to = slot.getBoundingClientRect();
    const anim = g.animate(
      [{ transform: g.style.transform }, { transform: `translate(${to.left}px, ${to.top}px)` }],
      { duration: 300, easing: 'cubic-bezier(.2,.8,.3,1)', fill: 'forwards' },
    );
    await anim.finished.catch(() => undefined);
  }
  // 松手后恢复与前台一致的 dense 回填
  await flip(() => { draggingId.value = ''; });
}

/* ---------- 改宽：右缘拖柄按 1/3 · 2/3 · 整行吸附（只落在模块支持的宽度上） ---------- */
function setSpan(mod: AboutModule, span: Span): void {
  if (spanOf(mod) === span) return;
  void flip(() => { mod.span = span; });
}

function spanItems(mod: AboutModule) {
  const spans = metaOf(mod.type)?.spans ?? [];
  if (spans.length < 2) return [];
  return spans.map((sp, i) => ({
    icon: spanOf(mod) === sp ? 'check' : 'grid',
    label: t(`studio.about.span${sp}`),
    divider: i === 0,
    run: () => setSpan(mod, sp),
  }));
}

function resizable(mod: AboutModule): boolean {
  return (metaOf(mod.type)?.spans.length ?? 0) > 1;
}

function onResizeDown(e: PointerEvent, mod: AboutModule): void {
  const el = (e.currentTarget as HTMLElement).closest<HTMLElement>('.mod');
  const meta = metaOf(mod.type);
  if (!el || !grid.value || !meta || e.button !== 0) return;
  e.preventDefault();
  e.stopPropagation();
  resizingId.value = mod.id;
  const gs = getComputedStyle(grid.value);
  const cols = gs.gridTemplateColumns.split(' ').length;
  const gap = parseFloat(gs.columnGap) || 0;
  const colW = (grid.value.clientWidth - gap * (cols - 1)) / cols;
  const move = (ev: PointerEvent): void => {
    const left = el.getBoundingClientRect().left;
    const want = Math.max(1, Math.min(cols, Math.round((ev.clientX - left + gap) / (colW + gap))));
    const span = meta.spans.reduce((a, b) => (Math.abs(b - want) < Math.abs(a - want) ? b : a));
    setSpan(mod, span);
  };
  const up = (): void => {
    window.removeEventListener('pointermove', move);
    window.removeEventListener('pointerup', up);
    window.removeEventListener('pointercancel', up);
    resizingId.value = '';
  };
  window.addEventListener('pointermove', move);
  window.addEventListener('pointerup', up);
  window.addEventListener('pointercancel', up);
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
  if (drag) cancelAnimationFrame(drag.raf);
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

    <!-- 站点身份：只读摘要，内容在「身份」页统一编辑 -->
    <div class="identity st-rise">
      <img class="av" :src="about.avatar || '/favicon-256.png'" alt="" />
      <div class="id-main">
        <b>{{ about.name || 'Myself' }}</b>
        <small>{{ plainText(about.tagline) }} · {{ t('studio.about.since', { d: about.foundedAt }) }}</small>
      </div>
      <span class="count">{{ t('studio.about.count', { n: about.modules.length, v: visibleCount }) }}</span>
      <router-link class="st-btn g sm" :to="{ name: 'admin-identity' }" :title="t('studio.about.identityHint')">
        <SIcon name="user" :size="16" />{{ t('studio.about.editIdentity') }}
      </router-link>
    </div>

    <div ref="grid" class="mods" :class="{ arranging: draggingId || resizingId, ordering: draggingId }">
      <template v-for="mod in about.modules" :key="mod.id">
      <!-- 拖动中：原位显示同宽占位框，真卡以幽灵卡形式跟手 -->
      <div v-if="draggingId === mod.id" class="slot" :class="spanClass(mod)" :data-id="mod.id" :style="{ minHeight: `${ghostBox.h}px` }" />
      <div
        v-else
        class="mod"
        :class="[spanClass(mod), { off: mod.hidden, open: expanded === mod.id, resizing: resizingId === mod.id }]"
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
          <p class="sum">{{ summaryOf(mod) }}</p>
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
        <span
          v-if="resizable(mod) && expanded !== mod.id"
          class="rsz"
          :title="t('studio.about.resize')"
          @pointerdown="onResizeDown($event, mod)"
        ><i /></span>
      </div>
      </template>

      <button type="button" class="mod addm" @click="pickerOpen = true">
        <SIcon name="plus" />{{ t('studio.about.add') }}
      </button>
    </div>

    <!-- 幽灵卡：fixed 定位跟手，只画头部与摘要 -->
    <div v-if="draggingMod" ref="ghost" class="mod ghost" :style="{ width: `${ghostBox.w}px`, height: `${ghostBox.h}px`, transform: `translate(${ghostBox.x}px, ${ghostBox.y}px)` }">
      <div class="mh">
        <span class="grip"><SIcon name="grip" :size="18" /></span>
        <span class="mi"><svg class="st-ic" width="16" height="16" viewBox="0 0 24 24"><path :d="metaOf(draggingMod.type)?.icon ?? ''" /></svg></span>
        <b>{{ titleOf(draggingMod) }}</b>
      </div>
      <div class="mb">
        <p class="sum">{{ summaryOf(draggingMod) }}</p>
        <p class="desc">{{ metaOf(draggingMod.type)?.desc }}</p>
      </div>
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

/* ---------- 身份条（只读摘要） ---------- */
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
    width: 48px;
    height: 48px;
    border-radius: 50%;
    object-fit: cover;
    box-shadow: 0 0 0 2px var(--paper), 0 0 0 3px var(--line-2);
  }

  .id-main {
    min-width: 0;

    b { display: block; font: 600 17px/1.3 var(--font-serif); }
    small { font-size: 12.5px; color: var(--st-ink-3); }
  }

  .count { font-size: 12.5px; color: var(--st-ink-3); }
}

/* ---------- 积木 ---------- */
.mods {
  position: relative;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

/* 与前台一致：按顺序排列并 dense 回填空格；拖动排序期间按纯顺序排列，插入位置可预期 */
.mods { grid-auto-flow: row dense; }
.mods.arranging { user-select: none; cursor: grabbing; }
.mods.ordering { grid-auto-flow: row; }

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

  &.resizing { box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--ink) 55%, transparent); }

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

/* 占位框：与被拖卡片同宽同高的虚线槽，主色轻染 */
.slot {
  border-radius: var(--r-lg);
  background: color-mix(in oklab, var(--ink) 6%, transparent);
  box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--ink) 45%, transparent) inset;
  background-image: repeating-linear-gradient(-45deg, transparent 0 10px, color-mix(in oklab, var(--ink) 5%, transparent) 10px 20px);
}

/* 幽灵卡：浮在最上层跟手，抬升阴影；不参与布局 */
.ghost {
  position: fixed;
  left: 0;
  top: 0;
  z-index: 60;
  overflow: hidden;
  pointer-events: none;
  box-shadow: 0 0 0 1.5px var(--line-3), var(--shadow-pop);
  will-change: transform;
  animation: ghost-lift 0.22s var(--ease-out);
}

@keyframes ghost-lift { from { box-shadow: 0 0 0 1px var(--line-2); } }

/* 改宽拖柄：右缘一条竖向握把，悬停卡片时浮现 */
.rsz {
  position: absolute;
  top: 50%;
  right: -7px;
  width: 14px;
  height: 44px;
  translate: 0 -50%;
  display: grid;
  place-items: center;
  cursor: ew-resize;
  touch-action: none;
  opacity: 0;
  transition: opacity var(--dur-fast);

  i {
    width: 5px;
    height: 28px;
    border-radius: var(--r-pill);
    background: var(--paper);
    box-shadow: 0 0 0 1px var(--line-3), 0 2px 6px var(--line-2);
    transition: background var(--dur-fast), height var(--dur-fast) var(--ease-spring);
  }

  &:hover i { height: 36px; background: var(--ink); box-shadow: none; }
}

.mod:hover .rsz,
.mod.resizing .rsz { opacity: 1; }
.mod.resizing .rsz i { height: 36px; background: var(--ink); box-shadow: none; }

/* 添加模块：网格末尾整行细条，不参与 dense 回填 */
.addm {
  grid-column: 1 / -1;
  align-items: center;
  justify-content: center;
  flex-direction: row;
  gap: 8px;
  min-height: 64px;
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
