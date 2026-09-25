<script lang="ts">
import { reactive } from 'vue';

/** 内容页轻提示（复制链接等）：模块级单例，顶部居中毛玻璃胶囊，2 秒自动收起 */
const state = reactive({ text: '', seq: 0 });
let timer = 0;

export function showToast(text: string): void {
  state.text = text;
  state.seq += 1;
  window.clearTimeout(timer);
  timer = window.setTimeout(() => { state.text = ''; }, 2000);
}
export default {};
</script>

<script setup lang="ts">
import ContentIcon from './ContentIcon.vue';
</script>

<template>
  <Teleport to="body">
    <Transition name="ctoast">
      <div v-if="state.text" :key="state.seq" class="ctoast" role="status">
        <span class="ok"><ContentIcon name="check" size="xs" /></span>{{ state.text }}
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped lang="scss">
.ctoast {
  position: fixed;
  z-index: 9000;
  left: 50%;
  top: 84px;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  height: 44px;
  padding: 0 18px 0 12px;
  border-radius: var(--r-pill);
  background: color-mix(in oklab, var(--surface-2) 86%, transparent);
  backdrop-filter: blur(20px) saturate(170%);
  -webkit-backdrop-filter: blur(20px) saturate(170%);
  box-shadow: var(--shadow-pop);
  font-size: 14px;
  color: var(--text);
  transform: translateX(-50%);
  pointer-events: none;
}

:root[data-mode='light'] .ctoast { background: rgb(255 255 255 / 0.9); }

.ok {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: var(--fill-2);
  color: var(--ink);
}

.ctoast-enter-active { transition: transform 0.45s var(--ease-spring), opacity 0.25s var(--ease-out); }
.ctoast-leave-active { transition: transform 0.3s var(--ease-out), opacity 0.2s var(--ease-out); }

.ctoast-enter-from,
.ctoast-leave-to {
  opacity: 0;
  transform: translate(-50%, -16px) scale(0.9);
}
</style>
