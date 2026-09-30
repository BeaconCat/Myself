<script setup lang="ts">
/**
 * 开关（替代原生 checkbox / 单选 radio）：role="switch" 的按钮，默认插槽为右侧文字。
 * 「多项中唯一」（如精选、常住地）由宿主在 update:modelValue 里处理互斥。
 */
defineProps<{ modelValue?: boolean; disabled?: boolean }>();
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>();
</script>

<template>
  <button
    type="button"
    class="ui-sw-t"
    role="switch"
    :class="{ on: modelValue }"
    :aria-checked="!!modelValue"
    :disabled="disabled"
    @click="emit('update:modelValue', !modelValue)"
  >
    <span class="ui-sw-track"><i /></span>
    <span v-if="$slots.default" class="ui-sw-label"><slot /></span>
  </button>
</template>

<style lang="scss">
.ui-sw-t {
  display: inline-flex;
  flex: none;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 0;
  border: 0;
  background: none;
  color: var(--text);
  font: inherit;
  font-size: 13.5px;
  white-space: nowrap;
  cursor: pointer;

  &:disabled { opacity: 0.45; cursor: not-allowed; }
  &:focus-visible { outline: none; }
  &:focus-visible .ui-sw-track { box-shadow: var(--focus); }
}

.ui-sw-track {
  position: relative;
  flex: none;
  width: 32px;
  height: 18px;
  border-radius: var(--r-pill);
  background: var(--fill-3, var(--line-2));
  transition: background-color var(--dur) var(--ease-out);

  i {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 1px 2px rgb(0 0 0 / 0.25);
    transition: transform var(--dur) var(--ease-spring);
  }

  .on > & { background: var(--solid); }
  .on > & i { transform: translateX(14px); }
  .ui-sw-t:active:not(:disabled) & i { width: 17px; }
  .ui-sw-t.on:active:not(:disabled) & i { transform: translateX(11px); }
}

.ui-sw-label { color: var(--text-2); transition: color var(--dur-fast); }
.on .ui-sw-label { color: var(--text); }

@media (prefers-reduced-motion: reduce) {
  .ui-sw-track, .ui-sw-track i { transition: none; }
}
</style>
