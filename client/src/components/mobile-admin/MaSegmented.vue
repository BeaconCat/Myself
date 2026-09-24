<script setup lang="ts">
/** 分段控件：滑块弹性平移，选中文字加粗；可带计数角标。 */
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
  margin: 0 20px;
  padding: 3px;
  border-radius: 999px;
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
    border-radius: 999px;
    transition: color var(--dur);
    white-space: nowrap;

    &.on {
      color: var(--text);
      font-weight: 600;
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
    border-radius: 999px;
    background: var(--elev-2);
    box-shadow: 0 3px 10px -3px rgba(0, 0, 0, 0.45), inset 0 0 0 0.5px var(--line-2);
    transform: translateX(calc(var(--i) * 100%));
    transition: transform 0.45s var(--ease-spring);
  }
}

:root[data-mode='light'] .ma-seg .th {
  background: #fff;
  box-shadow: 0 3px 10px -4px rgba(20, 40, 80, 0.25);
}
</style>
