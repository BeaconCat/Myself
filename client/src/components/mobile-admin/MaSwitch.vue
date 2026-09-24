<script setup lang="ts">
/** iOS 风格开关：滑块弹性，按下时横向拉伸。 */
defineProps<{ modelValue: boolean; label?: string }>();
const emit = defineEmits<{ 'update:modelValue': [v: boolean] }>();
</script>

<template>
  <button
    class="ma-switch"
    :class="{ on: modelValue }"
    role="switch"
    :aria-checked="modelValue"
    :aria-label="label"
    @click="emit('update:modelValue', !modelValue)"
  >
    <i />
  </button>
</template>

<style scoped lang="scss">
.ma-switch {
  position: relative;
  width: 51px;
  height: 31px;
  border-radius: 999px;
  background: var(--fill-2);
  box-shadow: inset 0 0 0 0.5px var(--line);
  flex: none;
  transition: background var(--dur) var(--ease-out);

  i {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 27px;
    height: 27px;
    border-radius: 999px;
    background: #fff;
    box-shadow: 0 3px 8px rgba(0, 0, 0, 0.18), 0 1px 1px rgba(0, 0, 0, 0.1);
    transition: transform 0.4s var(--ease-spring), width 0.2s var(--ease-out);
  }

  &:active i { width: 33px; }

  &.on {
    background: var(--primary);

    i { transform: translateX(20px); }
    &:active i { transform: translateX(14px); }
  }
}
</style>
