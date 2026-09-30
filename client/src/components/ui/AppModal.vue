<script setup lang="ts">
import { nextTick, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDialogStore } from '../../stores/dialog';

/** 全局模态宿主：挂在 App 根部，消费 dialog store */
const dialog = useDialogStore();
const { t } = useI18n();

const inputValue = ref('');
const inputEl = ref<HTMLInputElement | null>(null);
const okEl = ref<HTMLButtonElement | null>(null);
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
      await nextTick();
      // 聚焦到输入框或确认键，保证 Enter / Esc 立即可用
      if (active.input) inputEl.value?.focus();
      else okEl.value?.focus({ preventScroll: true });
    } else if (snapshot.value) {
      leaving.value = true;
      window.setTimeout(() => {
        leaving.value = false;
        snapshot.value = null;
      }, 220);
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
  // 焦点在按钮上时交给按钮自身的点击（避免在「取消」上按 Enter 却被当成确认）
  if (e.key === 'Enter' && !(e.target instanceof HTMLButtonElement)) done(true);
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
      <div class="modal" role="dialog" aria-modal="true" :aria-labelledby="snapshot.title ? 'app-modal-title' : undefined">
        <h3 v-if="snapshot.title" id="app-modal-title" class="m-title">{{ snapshot.title }}</h3>
        <p v-if="snapshot.message" class="m-msg">{{ snapshot.message }}</p>

        <input
          v-if="snapshot.input"
          ref="inputEl"
          v-model="inputValue"
          class="m-input"
          type="text"
          :placeholder="snapshot.placeholder"
          :aria-label="snapshot.label || snapshot.placeholder || snapshot.title"
        />

        <div class="m-actions">
          <button
            v-if="!snapshot.alertOnly"
            class="m-btn ghost"
            @click="done(false)"
          >{{ snapshot.cancelText ?? t('dialog.cancel') }}</button>
          <button
            ref="okEl"
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
  background: var(--scrim);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  animation: mask-in 0.25s ease both;

  &.leaving {
    animation: mask-out 0.22s ease both;

    .modal { animation: modal-out 0.22s var(--ease-out) both; }
  }
}

@keyframes mask-in { from { opacity: 0; } }
@keyframes mask-out { to { opacity: 0; } }

.modal {
  width: min(420px, 100%);
  padding: 24px 24px 20px;
  background: var(--elev);
  border-radius: var(--r-xl);
  box-shadow: var(--shadow-pop);
  /* 灵动入场：缩放 + 上浮 + 微回弹 */
  animation: modal-in 0.42s var(--ease-spring) both;
}

@keyframes modal-in {
  from { opacity: 0; transform: translateY(16px) scale(0.96); }
  50% { opacity: 1; }
  to { opacity: 1; transform: none; }
}

@keyframes modal-out {
  to { opacity: 0; transform: translateY(8px) scale(0.97); }
}

.m-title {
  font-size: 17px;
  line-height: 1.4;
  margin-bottom: 8px;
}

.m-msg {
  font-size: 14px;
  line-height: 1.8;
  color: var(--text-2);
  word-break: break-word;
}

.m-input {
  width: 100%;
  height: 40px;
  margin-top: 14px;
  padding: 0 14px;
  border: 0;
  border-radius: var(--r-sm);
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line-2);
  color: var(--text);
  font-size: 14px;
  font-family: inherit;
  caret-color: var(--ink);
  outline: none;
  transition: box-shadow var(--dur-fast), background-color var(--dur-fast);

  &::placeholder { color: var(--text-3); }

  &:focus {
    background: transparent;
    box-shadow: inset 0 0 0 1px var(--ink), 0 0 0 3px var(--tint);
  }
}

.m-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 22px;
}

/* 按钮：次要 = 中性填充；主 = 实底 --solid；危险 = 品牌红实底。无渐变 / 外发光 / 文字投影 */
.m-btn {
  height: 38px;
  padding: 0 18px;
  border: 0;
  border-radius: var(--r-pill);
  font-size: 14px;
  font-weight: 500;
  line-height: 1;
  transition: background-color var(--dur-fast) var(--ease-out), color var(--dur-fast), box-shadow var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  &:active { transform: scale(0.97); transition-duration: 0.08s; }
  &:focus-visible { outline: none; box-shadow: var(--focus); }

  &.ghost {
    background: var(--fill-2);
    color: var(--text);
    box-shadow: inset 0 0 0 0.5px var(--line);

    &:hover { background: var(--fill-3); }
    &:focus-visible { box-shadow: var(--focus); }
  }

  &.primary {
    background: var(--solid);
    color: var(--on-solid);
    box-shadow: var(--btn-shadow);

    &:hover { background: var(--solid-hover); }
    &:focus-visible { box-shadow: var(--btn-shadow), var(--focus); }

    &.danger {
      background: #e0002c;
      color: #fff;

      &:hover { background: color-mix(in oklab, #e0002c 88%, black); }
    }
  }
}
</style>
