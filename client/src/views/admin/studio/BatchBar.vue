<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import './i18n';
import SIcon from './SIcon.vue';

/**
 * 批量操作条：列表里勾选了条目后贴底浮起。左侧「已选 N 项」+ 全选 / 全不选，右侧放各页自己的操作（默认插槽），末尾取消选择。
 * 进出场由父级 v-if 控制，这里自带 Transition。
 */
defineProps<{
  show: boolean;
  count: number;
  /** 当前可选的总数：全部已选时按钮变为「全不选」 */
  total: number;
  /** 计数后的补充说明（如总大小） */
  note?: string;
}>();
const emit = defineEmits<{ all: []; clear: [] }>();
const { t } = useI18n();
</script>

<template>
  <Transition name="batch">
    <div v-if="show" class="batch" role="toolbar">
      <span class="n">{{ t('studio.batch.picked', { n: count }) }}<small v-if="note" class="mono">{{ note }}</small></span>
      <button type="button" class="st-btn g sm" @click="emit('all')">{{ count >= total ? t('studio.batch.none') : t('studio.batch.all') }}</button>
      <span class="sp" />
      <slot />
      <button type="button" class="st-ibtn sm" :title="t('studio.batch.clear')" @click="emit('clear')"><SIcon name="x" :size="16" /></button>
    </div>
  </Transition>
</template>

<style scoped lang="scss">
.batch {
  position: sticky;
  bottom: 20px;
  z-index: 20;
  display: flex;
  align-items: center;
  gap: 10px;
  width: fit-content;
  min-width: min(560px, 100%);
  margin: 16px auto 0;
  padding: 8px 8px 8px 18px;
  border-radius: var(--r-lg);
  background: var(--paper);
  box-shadow: 0 0 0 1px var(--line-2), var(--shadow-pop);

  .n { font-weight: 600; font-size: 14px; white-space: nowrap; }
  .n small { margin-left: 8px; font-size: 12px; font-weight: 400; color: var(--st-ink-3); }
  .sp { flex: 1; }
}

.batch-enter-active, .batch-leave-active { transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-spring); }
.batch-enter-from, .batch-leave-to { opacity: 0; transform: translateY(16px) scale(0.97); }
</style>
