<script setup lang="ts">
import { computed, ref, useId } from 'vue';
import { ChevronDown } from 'lucide';
import Icon from './Icon.vue';
import UiListbox from './UiListbox.vue';
import { filterOptions, type UiOption } from './listbox';

/**
 * 组合框：可自由输入的文本框 + 模糊搜索下拉（替代原生 datalist）。
 * - 聚焦即展开全部候选；开始输入后按模糊匹配筛选并高亮命中字。
 * - ↑ ↓ 移动、Enter 选中、Esc 收起；点外部收起；不在列表里的文字照样保留（allowFree 默认开）。
 * - 选中候选时 v-model 写入 option.value，并额外抛出 pick(option)，便于联动其他字段（如城市 → 经纬度）。
 *
 *   <Combobox v-model="item.name" :options="cityOptions" @pick="fillCoords(item)" />
 */
const props = withDefaults(
  defineProps<{
    modelValue?: string;
    options: UiOption[];
    placeholder?: string;
    empty?: string;
    /** 最多显示的候选数 */
    limit?: number;
    /** 浮层最小宽度（px），默认与输入框同宽 */
    minWidth?: number;
    disabled?: boolean;
    ariaLabel?: string;
  }>(),
  { modelValue: '', placeholder: '', empty: '', limit: 80, minWidth: 0, disabled: false, ariaLabel: undefined },
);
const emit = defineEmits<{ 'update:modelValue': [value: string]; pick: [option: UiOption] }>();
defineOptions({ inheritAttrs: false });

const uid = useId();
const box = ref<HTMLElement | null>(null);
const input = ref<HTMLInputElement | null>(null);
const open = ref(false);
/** 用户是否在本次展开后输入过：未输入时展示全部候选，而不是只剩当前值 */
const typed = ref(false);
const active = ref(-1);

const items = computed(() => filterOptions(props.options, typed.value ? props.modelValue ?? '' : '', props.limit));

function show(): void {
  if (props.disabled || open.value) return;
  typed.value = false;
  open.value = true;
  active.value = items.value.findIndex((m) => m.option.value === props.modelValue);
}

function hide(): void {
  open.value = false;
  typed.value = false;
}

function onInput(e: Event): void {
  emit('update:modelValue', (e.target as HTMLInputElement).value);
  if (!open.value) show();
  typed.value = true;
  active.value = items.value.length ? 0 : -1;
}

function pick(o: UiOption): void {
  if (o.disabled) return;
  emit('update:modelValue', o.value);
  emit('pick', o);
  hide();
  // 让宿主的 @change 也能感知（与手动输入后失焦一致）
  input.value?.dispatchEvent(new Event('change', { bubbles: true }));
}

function move(delta: number): void {
  const n = items.value.length;
  if (!n) return;
  active.value = active.value < 0 ? (delta > 0 ? 0 : n - 1) : (active.value + delta + n) % n;
}

function onKey(e: KeyboardEvent): void {
  switch (e.key) {
    case 'ArrowDown':
    case 'ArrowUp':
      e.preventDefault();
      if (!open.value) show();
      else move(e.key === 'ArrowDown' ? 1 : -1);
      break;
    case 'Enter':
      if (open.value && active.value >= 0 && items.value[active.value]) {
        e.preventDefault();
        pick(items.value[active.value].option);
      } else hide();
      break;
    case 'Escape':
      if (open.value) {
        e.preventDefault();
        e.stopPropagation();
        hide();
      }
      break;
    case 'Tab':
      hide();
      break;
  }
}

function toggle(): void {
  if (open.value) hide();
  else {
    input.value?.focus();
    show();
  }
}
</script>

<template>
  <div v-bind="$attrs" ref="box" class="ui-cb" :class="{ open, disabled }">
    <input
      ref="input"
      class="ui-cb-in"
      type="text"
      role="combobox"
      autocomplete="off"
      aria-autocomplete="list"
      :aria-expanded="open"
      :aria-controls="open ? uid : undefined"
      :aria-activedescendant="open && active >= 0 ? `${uid}-${active}` : undefined"
      :aria-label="ariaLabel"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      @focus="show"
      @click="show"
      @input="onInput"
      @keydown="onKey"
    />
    <button type="button" class="ui-cb-btn" tabindex="-1" :disabled="disabled" :aria-label="ariaLabel" @click="toggle">
      <Icon :icon="ChevronDown" :size="15" :stroke="2" />
    </button>
  </div>
  <UiListbox
    :open="open"
    :anchor="box"
    :items="items"
    :active="active"
    :selected="modelValue"
    :empty="empty"
    :min-width="minWidth"
    :list-id="uid"
    @hover="active = $event"
    @pick="pick"
    @close="hide"
  >
    <template v-if="$slots.icon" #icon="{ option }"><slot name="icon" :option="option" /></template>
  </UiListbox>
</template>

<style lang="scss">
.ui-cb {
  position: relative;
  display: flex;
  align-items: center;
  min-width: 0;
  height: 36px;
  border-radius: var(--r-sm);
  background: var(--well, var(--bg));
  box-shadow: 0 0 0 1px var(--line) inset;
  transition: box-shadow var(--dur-fast), background-color var(--dur-fast);

  &:hover:not(.disabled) { box-shadow: 0 0 0 1px var(--line-3, var(--line-2)) inset; }
  &:focus-within, &.open {
    background: var(--paper, var(--bg));
    box-shadow: 0 0 0 1px color-mix(in oklab, var(--ink) 70%, transparent) inset, 0 0 0 3px color-mix(in oklab, var(--ink) 18%, transparent);
  }
  &.disabled { opacity: 0.5; }
}

.ui-cb-in {
  flex: 1;
  min-width: 0;
  height: 100%;
  padding: 0 0 0 12px;
  border: 0;
  outline: none;
  background: none;
  color: var(--text);
  font: inherit;
  font-size: 14px;

  &::placeholder { color: var(--text-2); opacity: 0.8; }
}

.ui-cb-btn {
  display: grid;
  place-items: center;
  flex: none;
  width: 30px;
  height: 100%;
  padding: 0;
  border: 0;
  background: none;
  color: var(--text-2);
  cursor: pointer;

  svg { transition: transform var(--dur) var(--ease-spring); }
  .open > & svg { transform: rotate(180deg); }
  &:hover { color: var(--text); }
}

@media (prefers-reduced-motion: reduce) {
  .ui-cb-btn svg { transition: none; }
}
</style>
