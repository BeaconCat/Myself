<script setup lang="ts">
import { computed, nextTick, ref, useId, watch } from 'vue';
import { ChevronDown, Search } from 'lucide';
import Icon from './Icon.vue';
import UiListbox from './UiListbox.vue';
import { filterOptions, type UiOption } from './listbox';

/**
 * 下拉选择（替代原生 <select>）：按钮触发 + 浮层列表，可选顶部搜索框（searchable）。
 * 键盘：↑ ↓ 移动、Enter 选中、Esc 关闭、Home / End；非搜索模式下按字母跳到首字匹配项。
 * 插槽 icon（{ option }）：选项前的图标 / 色块 / 缩略图，同时用于按钮里已选项。
 *
 *   <Select v-model="item.icon" :options="opts" searchable>
 *     <template #icon="{ option }"><KitIcon :name="option.value" /></template>
 *   </Select>
 */
const props = withDefaults(
  defineProps<{
    modelValue?: string;
    options: UiOption[];
    placeholder?: string;
    searchable?: boolean;
    searchPlaceholder?: string;
    empty?: string;
    disabled?: boolean;
    /** 浮层最小宽度（px），默认与按钮同宽 */
    minWidth?: number;
    ariaLabel?: string;
  }>(),
  { modelValue: undefined, placeholder: '', searchable: false, searchPlaceholder: '', empty: '', disabled: false, minWidth: 0, ariaLabel: undefined },
);
const emit = defineEmits<{ 'update:modelValue': [value: string] }>();
defineOptions({ inheritAttrs: false });

const uid = useId();
const btn = ref<HTMLButtonElement | null>(null);
const search = ref<HTMLInputElement | null>(null);
const open = ref(false);
const query = ref('');
const active = ref(-1);

const current = computed(() => props.options.find((o) => o.value === props.modelValue));
const items = computed(() => filterOptions(props.options, props.searchable ? query.value : ''));

function show(): void {
  if (props.disabled || open.value) return;
  query.value = '';
  open.value = true;
  active.value = Math.max(0, items.value.findIndex((m) => m.option.value === props.modelValue));
  if (props.searchable) void nextTick(() => search.value?.focus());
}

function hide(refocus = true): void {
  if (!open.value) return;
  open.value = false;
  if (refocus) btn.value?.focus();
}

function pick(o: UiOption): void {
  if (o.disabled) return;
  emit('update:modelValue', o.value);
  hide();
}

function move(delta: number): void {
  const n = items.value.length;
  if (!n) return;
  let i = active.value;
  for (let k = 0; k < n; k++) {
    i = (i + delta + n) % n;
    if (!items.value[i].option.disabled) break;
  }
  active.value = i;
}

function onKey(e: KeyboardEvent): void {
  if (!open.value) {
    if (['ArrowDown', 'ArrowUp', 'Enter', ' '].includes(e.key)) {
      e.preventDefault();
      show();
    }
    return;
  }
  switch (e.key) {
    case 'ArrowDown': e.preventDefault(); move(1); break;
    case 'ArrowUp': e.preventDefault(); move(-1); break;
    case 'Home': if (!props.searchable) { e.preventDefault(); active.value = 0; } break;
    case 'End': if (!props.searchable) { e.preventDefault(); active.value = items.value.length - 1; } break;
    case 'Enter': {
      e.preventDefault();
      const m = items.value[active.value];
      if (m) pick(m.option);
      break;
    }
    case 'Escape': e.preventDefault(); e.stopPropagation(); hide(); break;
    case 'Tab': hide(false); break;
    default:
      // 非搜索模式：按字母跳到首字匹配项
      if (!props.searchable && e.key.length === 1 && !e.ctrlKey && !e.metaKey) {
        const k = e.key.toLowerCase();
        const i = items.value.findIndex((m) => m.option.label.toLowerCase().startsWith(k));
        if (i >= 0) active.value = i;
      }
  }
}

watch(query, () => { active.value = items.value.length ? 0 : -1; });
</script>

<template>
  <button
    v-bind="$attrs"
    ref="btn"
    type="button"
    class="ui-sel"
    :class="{ open, empty: !current }"
    :disabled="disabled"
    role="combobox"
    aria-haspopup="listbox"
    :aria-expanded="open"
    :aria-controls="open ? uid : undefined"
    :aria-activedescendant="open && active >= 0 && !searchable ? `${uid}-${active}` : undefined"
    :aria-label="ariaLabel"
    @click="open ? hide() : show()"
    @keydown="onKey"
  >
    <span v-if="current && $slots.icon" class="ui-sel-ic"><slot name="icon" :option="current" /></span>
    <span class="ui-sel-val">{{ current?.label ?? placeholder }}</span>
    <Icon class="ui-sel-chev" :icon="ChevronDown" :size="15" :stroke="2" />
  </button>
  <UiListbox
    :open="open"
    :anchor="btn"
    :items="items"
    :active="active"
    :selected="modelValue"
    :empty="empty"
    :min-width="minWidth"
    :list-id="uid"
    @hover="active = $event"
    @pick="pick"
    @close="hide(false)"
  >
    <template v-if="searchable" #header>
      <label class="ui-sel-search">
        <Icon :icon="Search" :size="15" />
        <input
          ref="search"
          v-model="query"
          type="text"
          :placeholder="searchPlaceholder"
          :aria-controls="uid"
          :aria-activedescendant="active >= 0 ? `${uid}-${active}` : undefined"
          @keydown="onKey"
        />
      </label>
    </template>
    <template v-if="$slots.icon" #icon="{ option }"><slot name="icon" :option="option" /></template>
  </UiListbox>
</template>

<style lang="scss">
/* 与后台 .a-input 同一套外观：井底 + 内描边，聚焦时信号色环（Teleport 的浮层样式在 UiListbox） */
.ui-sel {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  min-width: 0;
  height: 36px;
  padding: 0 10px 0 12px;
  border: 0;
  border-radius: var(--r-sm);
  background: var(--well, var(--bg));
  box-shadow: 0 0 0 1px var(--line) inset;
  color: var(--text);
  font: inherit;
  font-size: 14px;
  text-align: left;
  cursor: pointer;
  transition: box-shadow var(--dur-fast), background-color var(--dur-fast);

  &:hover:not(:disabled) { box-shadow: 0 0 0 1px var(--line-3, var(--line-2)) inset; }
  &:focus-visible, &.open {
    outline: none;
    background: var(--paper, var(--bg));
    box-shadow: 0 0 0 1px color-mix(in oklab, var(--ink) 70%, transparent) inset, 0 0 0 3px color-mix(in oklab, var(--ink) 18%, transparent);
  }
  &:disabled { opacity: 0.5; cursor: not-allowed; }
  &.empty .ui-sel-val { color: var(--text-2); }
}

.ui-sel-ic { display: grid; place-items: center; flex: none; color: var(--text-2); }
.ui-sel-val { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.ui-sel-chev {
  flex: none;
  color: var(--text-2);
  transition: transform var(--dur) var(--ease-spring);

  .open > & { transform: rotate(180deg); }
}

.ui-sel-search {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: none;
  margin: 0 0 4px;
  padding: 0 10px;
  height: 36px;
  border-bottom: 1px solid var(--line);
  color: var(--text-2);

  input {
    flex: 1;
    min-width: 0;
    height: 100%;
    border: 0;
    outline: none;
    background: none;
    color: var(--text);
    font: inherit;
    font-size: 13.5px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .ui-sel-chev { transition: none; }
}
</style>
