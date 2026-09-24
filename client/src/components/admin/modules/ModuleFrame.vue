<script setup lang="ts">
/**
 * ModuleFrame —— 关于页模块编辑器的公共头部（宽度 span / 变体 variant / 标题覆盖 / 隐藏开关）。
 *
 * 用法（后台「关于管理」页在每个模块编辑器外包一层）：
 *
 *   import ModuleFrame from '@/components/admin/modules/ModuleFrame.vue';
 *   import { MODULE_EDITORS } from '@/components/admin/modules';
 *
 *   <ModuleFrame :mod="mod">
 *     <component :is="MODULE_EDITORS[mod.type]" :mod="mod" />
 *   </ModuleFrame>
 *
 * - 直接修改 mod.span / mod.variant / mod.title / mod.hidden（与 mod.data 一样原地编辑，保存时随 about.modules 提交）。
 * - setup 时先调用 migrateInPlace(mod)，旧 schema（如 devices → uses）在编辑器渲染前即被迁移。
 * - 可选 prop `compact`：只显示 span + 隐藏开关（用于折叠态行内）。
 * - 可选 slot `aside`：头部右侧追加操作按钮（如删除 / 拖拽手柄）。
 */
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { AboutModule } from '../../../stores/config';
import { metaOf, spanOf, variantOf } from '../../../about/registry';
import { migrateInPlace } from '../../../about/migrate';

const props = defineProps<{ mod: AboutModule; compact?: boolean }>();
const { t } = useI18n();

migrateInPlace(props.mod);

const meta = computed(() => metaOf(props.mod.type));
const span = computed(() => spanOf(props.mod));
const variant = computed(() => variantOf(props.mod));

/* eslint-disable vue/no-mutating-props -- 模块对象由关于管理页持有，编辑器约定原地修改 */
function setSpan(s: 1 | 2 | 3): void { props.mod.span = s; }
function setVariant(v: string): void { props.mod.variant = v; }
function setTitle(e: Event): void {
  const v = (e.target as HTMLInputElement).value.trim();
  if (v) props.mod.title = v;
  else delete props.mod.title;
}
function toggleHidden(): void { props.mod.hidden = !props.mod.hidden; }
/* eslint-enable vue/no-mutating-props */

const SPAN_COLS: Record<number, number> = { 1: 4, 2: 8, 3: 12 };
</script>

<template>
  <div class="mf" :class="{ hidden: mod.hidden }">
    <div class="mf-head">
      <div class="mf-field">
        <span class="mf-label">{{ t('aboutKit.ed.span') }}</span>
        <div class="mf-seg">
          <button
            v-for="s in [1, 2, 3] as const"
            :key="s"
            type="button"
            :class="{ on: span === s }"
            :disabled="!meta?.spans.includes(s)"
            :title="t('aboutKit.ed.spanCols', { n: SPAN_COLS[s] })"
            @click="setSpan(s)"
          >
            <i class="mf-span" :style="{ '--f': SPAN_COLS[s] / 12 }" />{{ s }}
          </button>
        </div>
      </div>

      <div v-if="!compact && meta && meta.variants.length > 1" class="mf-field">
        <span class="mf-label">{{ t('aboutKit.ed.variant') }}</span>
        <div class="mf-seg">
          <button
            v-for="v in meta.variants"
            :key="v.id"
            type="button"
            :class="{ on: variant === v.id }"
            @click="setVariant(v.id)"
          >{{ v.label }}</button>
        </div>
      </div>

      <label v-if="!compact && meta?.chrome(variant)" class="mf-field mf-title">
        <span class="mf-label">{{ t('aboutKit.ed.title') }}</span>
        <input class="a-input" type="text" :value="mod.title ?? ''" :placeholder="meta?.name" @change="setTitle" />
      </label>

      <button type="button" class="mf-hide" :class="{ on: mod.hidden }" :aria-pressed="!!mod.hidden" @click="toggleHidden">
        <svg v-if="mod.hidden" viewBox="0 0 24 24"><path d="M3 3l18 18M10.6 10.6a2 2 0 0 0 2.8 2.8M9.4 5.2A9.8 9.8 0 0 1 12 5c5 0 8.5 4.5 9.5 7a13 13 0 0 1-3 4.2M6.2 6.6C4.3 8 3 10 2.5 12c1 2.5 4.5 7 9.5 7 1.6 0 3-.4 4.3-1" /></svg>
        <svg v-else viewBox="0 0 24 24"><path d="M2.5 12C3.5 9.5 7 5 12 5s8.5 4.5 9.5 7c-1 2.5-4.5 7-9.5 7s-8.5-4.5-9.5-7Z" /><circle cx="12" cy="12" r="3" /></svg>
        {{ mod.hidden ? t('aboutKit.ed.hidden') : t('aboutKit.ed.visible') }}
      </button>

      <slot name="aside" />
    </div>

    <div v-if="$slots.default" class="mf-body"><slot /></div>
  </div>
</template>

<style scoped lang="scss">
.mf { display: flex; flex-direction: column; gap: 14px; }

.mf-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 12px 16px;
  padding-bottom: 14px;
  border-bottom: 1px dashed color-mix(in oklab, var(--text) 12%, transparent);
}

.mf-field { display: flex; flex-direction: column; gap: 6px; }
.mf-title { flex: 1; min-width: 160px; }
.mf-title .a-input { padding-top: 6px; padding-bottom: 6px; height: 32px; }

.mf-label { font-size: 11.5px; font-weight: 600; letter-spacing: 0.04em; color: var(--text-2); }

.mf-seg {
  display: inline-flex;
  gap: 2px;
  padding: 3px;
  border-radius: 10px;
  background: color-mix(in oklab, var(--text) 5%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--text) 8%, transparent);

  button {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 26px;
    padding: 0 10px;
    border: 0;
    border-radius: 7px;
    background: none;
    font-size: 12.5px;
    color: var(--text-2);
    transition: background var(--dur-fast), color var(--dur-fast), box-shadow var(--dur-fast);

    &:hover:not(:disabled) { color: var(--text); }
    &:disabled { opacity: 0.3; cursor: not-allowed; }

    &.on {
      background: var(--surface);
      color: var(--text);
      box-shadow: 0 1px 2px rgb(0 0 0 / 0.12), inset 0 0 0 1px color-mix(in oklab, var(--text) 12%, transparent);
    }
  }
}

/* span 图示：12 栏里占多少 */
.mf-span {
  position: relative;
  width: 24px;
  height: 8px;
  border-radius: 2px;
  background: color-mix(in oklab, var(--text) 12%, transparent);

  &::after {
    content: '';
    position: absolute;
    inset: 0 auto 0 0;
    width: calc(var(--f) * 100%);
    border-radius: 2px;
    background: currentColor;
    opacity: 0.8;
  }
}

.on .mf-span::after { background: var(--primary); opacity: 1; }

.mf-hide {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  margin-left: auto;
  padding: 0 12px;
  border: 0;
  border-radius: 9px;
  background: color-mix(in oklab, var(--text) 5%, transparent);
  font-size: 12.5px;
  color: var(--text-2);
  transition: background var(--dur-fast), color var(--dur-fast);

  svg { width: 16px; height: 16px; fill: none; stroke: currentColor; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }
  &:hover { color: var(--text); }
  &.on { color: var(--accent-red); background: color-mix(in oklab, var(--accent-red) 10%, transparent); }
}

.mf.hidden .mf-body { opacity: 0.55; }
.mf-body { transition: opacity var(--dur); }
</style>
