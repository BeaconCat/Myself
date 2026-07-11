<script setup lang="ts">
import { nextTick, ref, watch } from 'vue';
import { useDialogStore } from '../../stores/dialog';

/** 全局模态宿主：挂在 App 根部，消费 dialog store */
const dialog = useDialogStore();

const inputValue = ref('');
const inputEl = ref<HTMLInputElement | null>(null);
/** 退场动画期间保留渲染 */
const leaving = ref(false);
const snapshot = ref(dialog.active);

watch(
  () => dialog.active,
  async (active) => {
    if (active) {
      snapshot.value = active;
      leaving.value = false;
      inputValue.value = active.inputValue ?? '';
      if (active.input) {
        await nextTick();
        inputEl.value?.focus();
      }
    } else if (snapshot.value) {
      leaving.value = true;
      window.setTimeout(() => {
        leaving.value = false;
        snapshot.value = null;
      }, 260);
    }
  },
);

function done(ok: boolean): void {
  const active = dialog.active;
  if (!active) return;
  if (active.input) dialog.settle(ok ? inputValue.value : null);
  else dialog.settle(ok);
}

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Enter') done(true);
  if (e.key === 'Escape') done(false);
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="snapshot"
      class="modal-mask"
      :class="{ leaving }"
      @click.self="done(false)"
      @keydown="onKey"
    >
      <div class="modal" role="dialog" aria-modal="true">
        <h3 v-if="snapshot.title" class="m-title">{{ snapshot.title }}</h3>
        <p v-if="snapshot.message" class="m-msg">{{ snapshot.message }}</p>

        <input
          v-if="snapshot.input"
          ref="inputEl"
          v-model="inputValue"
          class="m-input"
          type="text"
          :placeholder="snapshot.placeholder"
        />

        <div class="m-actions">
          <button
            v-if="!snapshot.alertOnly"
            class="m-btn ghost"
            @click="done(false)"
          >{{ snapshot.cancelText ?? '取消' }}</button>
          <button
            class="m-btn primary"
            :class="{ danger: snapshot.danger }"
            @click="done(true)"
          >{{ snapshot.confirmText ?? '确定' }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.modal-mask {
  position: fixed;
  inset: 0;
  z-index: 9800;
  display: grid;
  place-items: center;
  padding: 24px;
  background: rgba(8, 8, 12, 0.45);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  animation: mask-in 0.25s ease both;

  &.leaving {
    animation: mask-out 0.26s ease both;

    .modal { animation: modal-out 0.26s var(--ease-out) both; }
  }
}

@keyframes mask-in { from { opacity: 0; } }
@keyframes mask-out { to { opacity: 0; } }

.modal {
  width: min(420px, 100%);
  padding: 26px 26px 22px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: 0 30px 80px -20px rgba(0, 0, 0, 0.5);
  /* 灵动入场：缩放 + 上浮 + 微回弹 */
  animation: modal-in 0.38s var(--ease-spring) both;
}

@keyframes modal-in {
  from { opacity: 0; transform: translateY(26px) scale(0.92); }
  60% { opacity: 1; }
  to { opacity: 1; transform: none; }
}

@keyframes modal-out {
  to { opacity: 0; transform: translateY(14px) scale(0.95); }
}

.m-title {
  font-size: 17px;
  margin-bottom: 10px;
}

.m-msg {
  font-size: 14px;
  line-height: 1.8;
  color: var(--text-2);
  word-break: break-word;
}

.m-input {
  width: 100%;
  margin-top: 14px;
  padding: 10px 14px;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text);
  font-size: 14px;
  font-family: inherit;
  outline: none;
  transition: border-color var(--dur-fast), box-shadow var(--dur-fast);

  &:focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(var(--primary-rgb), 0.12);
  }
}

.m-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.m-btn {
  padding: 9px 22px;
  border-radius: 10px;
  border: 1px solid transparent;
  font-size: 14px;
  font-weight: 700;
  transition: all var(--dur-fast) var(--ease-out);

  &.ghost {
    background: var(--surface);
    border-color: var(--border);
    color: var(--text);

    &:hover { border-color: var(--primary); color: var(--primary); }
  }

  &.primary {
    color: #fff;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.4);

    &:hover { filter: brightness(1.08); transform: scale(1.04); }

    &.danger {
      background: linear-gradient(180deg, #ff2450, #d40029);
      box-shadow: 0 4px 14px rgba(255, 0, 50, 0.4);
    }
  }
}
</style>
