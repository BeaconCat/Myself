<script setup lang="ts" generic="T">
import { useI18n } from 'vue-i18n';

/**
 * 编辑器通用列表：每项一张小卡，右上角 上移 / 下移 / 删除；底部「添加」。
 * 用法：<EdList :items="d.items" :make="() => ({ name: '' })" v-slot="{ item, index }">…</EdList>
 */
const props = defineProps<{ items: T[]; make: () => T; addLabel?: string; max?: number; compact?: boolean }>();
const { t } = useI18n();

function move(i: number, delta: number): void {
  const j = i + delta;
  if (j < 0 || j >= props.items.length) return;
  const [it] = props.items.splice(i, 1);
  props.items.splice(j, 0, it);
}
</script>

<template>
  <div class="el" :class="{ compact }">
    <TransitionGroup name="el">
      <div v-for="(item, index) in items" :key="index" class="el-item">
        <div class="el-body"><slot :item="item" :index="index" /></div>
        <div class="el-ops">
          <button type="button" :disabled="index === 0" :title="t('aboutKit.ed.up')" @click="move(index, -1)">
            <svg viewBox="0 0 24 24"><path d="m6 15 6-6 6 6" /></svg>
          </button>
          <button type="button" :disabled="index === items.length - 1" :title="t('aboutKit.ed.down')" @click="move(index, 1)">
            <svg viewBox="0 0 24 24"><path d="m6 9 6 6 6-6" /></svg>
          </button>
          <button type="button" class="del" :title="t('aboutKit.ed.remove')" @click="items.splice(index, 1)">
            <svg viewBox="0 0 24 24"><path d="M6 6l12 12M18 6 6 18" /></svg>
          </button>
        </div>
      </div>
    </TransitionGroup>
    <button
      v-if="!max || items.length < max"
      type="button"
      class="a-btn ghost el-add"
      @click="items.push(make())"
    >
      <svg viewBox="0 0 24 24"><path d="M12 5v14M5 12h14" /></svg>{{ addLabel || t('aboutKit.ed.add') }}
    </button>
  </div>
</template>

<style scoped lang="scss">
.el { display: flex; flex-direction: column; gap: 8px; }

.el-item {
  position: relative;
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 12px;
  border-radius: 12px;
  background: color-mix(in oklab, var(--text) 3%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--text) 8%, transparent);
}

.compact .el-item { padding: 8px 10px; align-items: center; }

.el-body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 8px; }

.el-ops {
  display: flex;
  gap: 2px;
  flex-shrink: 0;

  button {
    display: grid;
    place-items: center;
    width: 26px;
    height: 26px;
    padding: 0;
    border: 0;
    border-radius: 7px;
    background: none;
    color: var(--text-2);
    transition: background var(--dur-fast), color var(--dur-fast);

    &:hover:not(:disabled) { background: color-mix(in oklab, var(--text) 8%, transparent); color: var(--text); }
    &:disabled { opacity: 0.3; cursor: default; }
    &.del:hover { color: var(--accent-red); }
  }

  svg { width: 15px; height: 15px; fill: none; stroke: currentColor; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }
}

.el-add {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  align-self: flex-start;
  padding: 7px 14px;
  font-size: 12.5px;

  svg { width: 14px; height: 14px; fill: none; stroke: currentColor; stroke-width: 2.2; stroke-linecap: round; }
}

.el-enter-active, .el-leave-active { transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-out); }
.el-enter-from, .el-leave-to { opacity: 0; transform: translateY(-6px); }
.el-move { transition: transform var(--dur) var(--ease-out); }
</style>
