<script setup lang="ts">
import SIcon from './SIcon.vue';
import { dismiss, toasts, type Toast } from './toast';

function act(t: Toast): void {
  t.fn?.();
  dismiss(t.id);
}
</script>

<template>
  <div class="toasts" aria-live="polite">
    <div v-for="t in toasts" :key="t.id" class="toast" :class="{ bye: t.leaving, act: !!t.action }">
      <SIcon :name="t.icon" :size="16" />
      <span>{{ t.msg }}</span>
      <button v-if="t.action" type="button" @click="act(t)">{{ t.action }}</button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.toasts {
  position: fixed;
  top: 18px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 100;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  pointer-events: none;
}

.toast {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 42px;
  padding: 0 18px 0 14px;
  border-radius: var(--r-pill);
  background: var(--toast-bg);
  color: var(--toast-ink);
  font-size: 13.5px;
  box-shadow: 0 14px 34px -10px rgba(0, 0, 0, 0.4);
  pointer-events: auto;
  animation: toast-in var(--dur-slow) var(--ease-spring) both;

  .st-ic { color: var(--primary); }

  &.act { padding-right: 6px; }

  button {
    height: 30px;
    padding: 0 12px;
    border-radius: var(--r-pill);
    font-size: 13px;
    font-weight: 500;
    color: var(--toast-ink);
    background: color-mix(in srgb, var(--toast-ink) 12%, transparent);

    &:hover { background: color-mix(in srgb, var(--toast-ink) 20%, transparent); }
  }

  &.bye { animation: toast-out 0.3s ease-in forwards; }
}

@keyframes toast-in { from { opacity: 0; transform: translateY(-24px) scale(0.8); } }
@keyframes toast-out { to { opacity: 0; transform: translateY(-14px) scale(0.9); } }
</style>
