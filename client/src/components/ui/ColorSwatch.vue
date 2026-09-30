<script setup lang="ts">
import { computed } from 'vue';

/**
 * 颜色按钮（替代裸露的原生 <input type="color">）：圆角色块 + 可选十六进制文字，
 * 点击仍唤起系统取色器（透明覆盖在色块上），外观与后台输入框同一套井底样式。
 * showHex = 在色块旁显示 #RRGGBB（等宽）；label 用作无障碍名称与悬停提示。
 */
const props = withDefaults(defineProps<{ modelValue?: string; label?: string; showHex?: boolean; fallback?: string }>(), {
  modelValue: '',
  label: '',
  showHex: false,
  fallback: '#888888',
});
const emit = defineEmits<{ 'update:modelValue': [value: string] }>();

/** 原生取色器只接受 #rrggbb：三位简写展开，非法值用 fallback */
const hex = computed(() => {
  const v = (props.modelValue || '').trim();
  if (/^#[0-9a-f]{6}$/i.test(v)) return v.toLowerCase();
  if (/^#[0-9a-f]{3}$/i.test(v)) return `#${[...v.slice(1)].map((c) => c + c).join('')}`.toLowerCase();
  return props.fallback;
});
</script>

<template>
  <span class="ui-sw" :class="{ wide: showHex }" :title="label || undefined">
    <i class="ui-sw-chip" :style="{ background: modelValue || fallback }" />
    <span v-if="showHex" class="ui-sw-hex">{{ hex.toUpperCase() }}</span>
    <input
      type="color"
      :value="hex"
      :aria-label="label || undefined"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
  </span>
</template>

<style lang="scss">
.ui-sw {
  position: relative;
  display: inline-flex;
  flex-direction: row;
  flex: none;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 36px;
  height: 36px;
  border-radius: var(--r-sm);
  background: var(--well, var(--bg));
  box-shadow: 0 0 0 1px var(--line) inset;
  cursor: pointer;
  transition: box-shadow var(--dur-fast);

  &.wide { width: auto; padding: 0 12px 0 8px; justify-content: flex-start; }
  &:hover { box-shadow: 0 0 0 1px var(--line-3, var(--line-2)) inset; }
  &:focus-within { box-shadow: 0 0 0 1px color-mix(in oklab, var(--ink) 70%, transparent) inset, 0 0 0 3px color-mix(in oklab, var(--ink) 18%, transparent); }

  input {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    padding: 0;
    border: 0;
    opacity: 0;
    cursor: pointer;
  }
}

.ui-sw-chip {
  flex: none;
  width: 20px;
  height: 20px;
  border-radius: var(--r-xs);
  box-shadow: 0 0 0 1px color-mix(in oklab, var(--text) 14%, transparent) inset;
  transition: background-color var(--dur) var(--ease-out), transform var(--dur-fast) var(--ease-spring);

  .ui-sw:active & { transform: scale(0.9); }
}

.ui-sw-hex { font: 12.5px var(--font-mono, monospace); letter-spacing: 0.02em; color: var(--text-2); }
</style>
