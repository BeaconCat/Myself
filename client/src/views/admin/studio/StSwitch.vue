<script setup lang="ts">
/** 开关：按下时滑块拉伸，带回弹 */
const model = defineModel<boolean>({ default: false });
defineProps<{ disabled?: boolean; label?: string }>();
</script>

<template>
  <button
    type="button"
    role="switch"
    class="sw"
    :class="{ on: model }"
    :aria-checked="model"
    :aria-label="label"
    :disabled="disabled"
    @click="model = !model"
  />
</template>

<style scoped lang="scss">
.sw {
  position: relative;
  width: 38px;
  height: 22px;
  border-radius: 11px;
  background: var(--line-3);
  flex: none;
  cursor: pointer;
  transition: background var(--dur) var(--ease-out);

  &::after {
    content: '';
    position: absolute;
    top: 2px;
    left: 2px;
    width: 18px;
    height: 18px;
    border-radius: 9px;
    background: #fff;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2), 0 0 0 0.5px rgba(0, 0, 0, 0.04);
    transition: all var(--dur) var(--ease-spring);
  }

  &:active:not(:disabled)::after { width: 23px; }

  &.on {
    background: var(--primary);

    &::after { left: 18px; }
    &:active:not(:disabled)::after { left: 13px; }
  }

  &:disabled { opacity: 0.45; cursor: not-allowed; }
}
</style>
