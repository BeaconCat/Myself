<script setup lang="ts">
/** 分段控件：选中块 = 抬升 + 轻染，弹性平移；选中项图标着 --ink；可带计数角标。 */
defineProps<{ modelValue: number; items: { label: string; count?: number | string }[] }>();
const emit = defineEmits<{ 'update:modelValue': [v: number] }>();
</script>

<template>
  <div class="ma-seg" role="tablist" :style="{ '--i': modelValue, '--n': items.length }">
    <span class="th" />
    <button
      v-for="(it, k) in items"
      :key="k"
      role="tab"
      :aria-selected="modelValue === k"
      :class="{ on: modelValue === k }"
      @click="emit('update:modelValue', k)"
    >
      {{ it.label }}<em v-if="it.count !== undefined">{{ it.count }}</em>
    </button>
  </div>
</template>

<style scoped lang="scss">
.ma-seg {
  position: relative;
  display: grid;
  grid-template-columns: repeat(var(--n), 1fr);
  height: 40px;
  margin: 0 16px;
  padding: 3px;
  border-radius: var(--r-md);
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line);

  button {
    position: relative;
    z-index: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    font-size: 14px;
    color: var(--text-2);
    border-radius: calc(var(--r-md) - 3px);
    transition: color var(--dur);
    white-space: nowrap;

    &.on {
      color: var(--lift-fg);
      font-weight: 500;

      :deep(svg) { color: var(--ink); }
    }

    em {
      font-style: normal;
      font-family: var(--font-mono);
      font-size: 11px;
      color: var(--text-3);
    }
  }

  .th {
    position: absolute;
    top: 3px;
    bottom: 3px;
    left: 3px;
    width: calc((100% - 6px) / var(--n));
    border-radius: calc(var(--r-md) - 3px);
    background: var(--lift);
    box-shadow: var(--lift-shadow);
    transform: translateX(calc(var(--i) * 100%));
    transition: transform 0.45s var(--ease-spring), background-color var(--dur), box-shadow var(--dur);
  }
}
</style>
