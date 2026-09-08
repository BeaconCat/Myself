<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { SiteConfig, ThemePreset } from '../../../stores/config';

/** 主题：默认色盘 / 模式 / 自动切换 + 预设列表（拖拽排序、增删） */
const props = defineProps<{ cfg: SiteConfig }>();
const { t } = useI18n();

/* ===== 主题预设：拖拽排序 / 增删 ===== */
let dragIndex = -1;

function onDragStart(i: number): void {
  dragIndex = i;
}

function onDrop(i: number): void {
  if (dragIndex < 0 || dragIndex === i) return;
  const list = props.cfg.theme.presets;
  const [moved] = list.splice(dragIndex, 1);
  list.splice(i, 0, moved);
  dragIndex = -1;
}

function addPreset(): void {
  if (props.cfg.theme.presets.length >= 10) return;
  props.cfg.theme.presets.push({
    id: `custom-${Date.now().toString(36)}`,
    name: '自定义',
    primary: '#7c4dff',
    primaryDeep: '#5e35b1',
  });
}

function removePreset(preset: ThemePreset): void {
  const theme = props.cfg.theme;
  if (theme.presets.length <= 1) return;
  theme.presets = theme.presets.filter((p) => p.id !== preset.id);
  if (theme.defaultPaletteId === preset.id) {
    theme.defaultPaletteId = theme.presets[0].id;
  }
}
</script>

<template>
  <section id="sec-theme" class="card">
    <h2>{{ t('admin.secTheme') }}</h2>
    <div class="row3">
      <label>
        <span>{{ t('admin.defaultPalette') }}</span>
        <select v-model="cfg.theme.defaultPaletteId">
          <option v-for="p in cfg.theme.presets" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
      </label>
      <label>
        <span>{{ t('admin.defaultMode') }}</span>
        <select v-model="cfg.theme.defaultMode">
          <option value="light">{{ t('theme.light') }}</option>
          <option value="dark">{{ t('theme.dark') }}</option>
        </select>
      </label>
      <label>
        <span>{{ t('admin.autoSwitch') }}</span>
        <select v-model="cfg.theme.autoSwitch">
          <option value="off">{{ t('admin.autoOff') }}</option>
          <option value="season">{{ t('admin.autoSeason') }}</option>
        </select>
      </label>
    </div>
    <div class="row3">
      <label class="switch">
        <input v-model="cfg.theme.allowUserPalette" type="checkbox" />
        <i class="track" aria-hidden="true" />
        <span>{{ t('admin.allowUserPalette') }}</span>
      </label>
      <label>
        <span>{{ t('admin.displayCount') }}</span>
        <input v-model.number="cfg.theme.displayCount" type="number" min="1" max="4" />
      </label>
    </div>

    <!-- 预设列表：拖拽排序，前 N 个展示 -->
    <p class="sub-h">{{ t('admin.presets', { n: cfg.theme.presets.length }) }}</p>
    <ul class="presets">
      <li
        v-for="(preset, i) in cfg.theme.presets"
        :key="preset.id"
        :class="{ shown: i < cfg.theme.displayCount }"
        draggable="true"
        @dragstart="onDragStart(i)"
        @dragover.prevent
        @drop="onDrop(i)"
      >
        <span class="grip" aria-hidden="true">⋮⋮</span>
        <input v-model="preset.name" class="p-name" type="text" />
        <label class="color"><span>{{ t('admin.primary') }}</span><input v-model="preset.primary" type="color" /></label>
        <label class="color"><span>{{ t('admin.primaryDeep') }}</span><input v-model="preset.primaryDeep" type="color" /></label>
        <span class="p-flag">{{ i < cfg.theme.displayCount ? t('admin.shown') : t('admin.hidden') }}</span>
        <button class="op danger" @click="removePreset(preset)">{{ t('admin.delete') }}</button>
      </li>
    </ul>
    <button class="btn ghost" :disabled="cfg.theme.presets.length >= 10" @click="addPreset">
      {{ t('admin.addPreset') }}
    </button>
  </section>
</template>

<style scoped lang="scss">
@use './settings-shared';

/* 预设行 */
.presets {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;

  li {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--bg);
    cursor: grab;
    opacity: 0.6;

    &.shown { opacity: 1; border-color: rgba(var(--primary-rgb), 0.35); }
    &:active { cursor: grabbing; }
  }
}

.grip { color: var(--text-2); letter-spacing: -2px; user-select: none; }

.p-name { flex: 1; max-width: 180px; }

.color {
  flex-direction: row !important;
  align-items: center;
  gap: 6px;

  input[type='color'] {
    width: 34px;
    height: 26px;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: none;
    padding: 0;
    cursor: pointer;
  }
}

.p-flag {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 10px;
  border-radius: 999px;
  background: var(--surface-2);
  color: var(--text-2);
  margin-left: auto;
}

li.shown .p-flag {
  background: rgba(var(--primary-rgb), 0.12);
  color: var(--primary);
}
</style>
