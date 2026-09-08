<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

/** 时间选择器：快捷区间（今天 / 7 天 / 30 天）+ 自定义起止日期 */
const from = defineModel<string>('from', { required: true });
const to = defineModel<string>('to', { required: true });

const emit = defineEmits<{
  /** 快捷区间 */
  quick: [days: number];
  /** 清除筛选 */
  clear: [];
  /** 自定义日期变更 */
  change: [];
}>();

const { t } = useI18n();
const dateOpen = ref(false);

const hasDateFilter = computed(() => !!(from.value || to.value));
</script>

<template>
  <div class="date-wrap">
    <button
      class="date-btn"
      :class="{ on: hasDateFilter || dateOpen }"
      :title="t('thoughts.dateFilter')"
      @click="dateOpen = !dateOpen"
    >
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="3" y="5" width="18" height="16" rx="2" />
        <path d="M8 3v4M16 3v4M3 10h18" />
      </svg>
    </button>

    <transition name="pop">
      <div v-if="dateOpen" class="date-pop">
        <div class="dp-quick">
          <button @click="emit('quick', 1)">{{ t('thoughts.today') }}</button>
          <button @click="emit('quick', 7)">{{ t('thoughts.last7') }}</button>
          <button @click="emit('quick', 30)">{{ t('thoughts.last30') }}</button>
          <button class="clear" @click="emit('clear')">{{ t('thoughts.clearDate') }}</button>
        </div>
        <div class="dp-range">
          <label>
            <span>{{ t('thoughts.dateFrom') }}</span>
            <input v-model="from" type="date" @change="emit('change')" />
          </label>
          <label>
            <span>{{ t('thoughts.dateTo') }}</span>
            <input v-model="to" type="date" @change="emit('change')" />
          </label>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped lang="scss">
.date-wrap { position: relative; }

.date-btn {
  width: 40px;
  height: 40px;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text-2);
  display: grid;
  place-items: center;
  transition: all var(--dur-fast) var(--ease-out);

  svg { width: 17px; height: 17px; }

  &:hover { border-color: var(--primary); color: var(--primary); transform: scale(1.06); }

  &.on {
    border-color: rgba(var(--primary-rgb), 0.5);
    color: var(--primary);
    background: rgba(var(--primary-rgb), 0.08);
  }
}

.date-pop {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  z-index: 30;
  width: 260px;
  padding: 14px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--shadow), 0 16px 40px -12px rgba(var(--primary-rgb), 0.25);
}

.pop-enter-active, .pop-leave-active { transition: opacity var(--dur-fast), transform var(--dur-fast) var(--ease-out); }
.pop-enter-from, .pop-leave-to { opacity: 0; transform: translateY(-8px) scale(0.96); }

.dp-quick {
  display: flex;
  gap: 6px;
  margin-bottom: 12px;

  button {
    flex: 1;
    padding: 6px 0;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: none;
    color: var(--text-2);
    font-size: 12px;
    font-weight: 600;
    transition: all var(--dur-fast);

    &:hover { border-color: var(--primary); color: var(--primary); }
    &.clear:hover { border-color: var(--accent-red); color: var(--accent-red); }
  }
}

.dp-range {
  display: flex;
  flex-direction: column;
  gap: 10px;

  label {
    display: flex;
    align-items: center;
    gap: 10px;

    span { font-size: 12px; color: var(--text-2); width: 28px; flex-shrink: 0; }

    input {
      flex: 1;
      padding: 7px 10px;
      border-radius: 8px;
      border: 1px solid var(--border);
      background: var(--bg);
      color: var(--text);
      font-size: 13px;
      font-family: inherit;
      outline: none;

      &:focus { border-color: var(--primary); }
    }
  }
}
</style>
