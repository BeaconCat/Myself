<script setup lang="ts">
import { ref, watch } from 'vue';
import { shell } from './shell';
import MIcon from './MIcon.vue';

/** 灵动岛轻提示：黑色胶囊从顶部展开，图标与文字延迟浮现，2.2s 后收回。 */
const wide = ref(false);
const text = ref('');
const sub = ref('');
let timer = 0;

watch(
  () => shell.toast.seq,
  () => {
    text.value = shell.toast.text;
    sub.value = shell.toast.sub;
    wide.value = true;
    window.clearTimeout(timer);
    timer = window.setTimeout(() => { wide.value = false; }, 2200);
  },
);
</script>

<template>
  <div class="island" :class="{ wide }" role="status" aria-live="polite">
    <span class="isl-ic"><MIcon name="check" class="xs" /></span>
    <span class="isl-t">{{ text }}<small v-if="sub">{{ sub }}</small></span>
  </div>
</template>

<style scoped lang="scss">
.island {
  position: fixed;
  z-index: 200;
  top: calc(var(--m-safe-t) + 10px);
  left: 50%;
  width: 124px;
  height: 36px;
  margin-left: -62px;
  border-radius: 20px;
  background: #000;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 16px;
  color: #fff;
  overflow: hidden;
  pointer-events: none;
  opacity: 0;
  transform: translateY(-8px) scale(0.9);
  box-shadow: 0 12px 30px -10px rgba(0, 0, 0, 0.5);
  transition:
    width 0.5s var(--ease-spring),
    margin-left 0.5s var(--ease-spring),
    height 0.5s var(--ease-spring),
    border-radius 0.5s var(--ease-spring),
    opacity 0.3s var(--ease-out) 0.2s,
    transform 0.4s var(--ease-spring) 0.1s;

  &.wide {
    width: min(300px, calc(100vw - 48px));
    margin-left: calc(min(300px, calc(100vw - 48px)) / -2);
    height: 44px;
    border-radius: 24px;
    opacity: 1;
    transform: none;
    transition-delay: 0s;
  }
}

.isl-ic {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: var(--primary);
  color: var(--m-on-primary, #fff);
  flex: none;
  opacity: 0;
  transform: scale(0.4);
  transition: all 0.35s var(--ease-spring);

  .m-ic { stroke-width: 2.4; }
}

.isl-t {
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  opacity: 0;
  transform: translateY(4px);
  transition: all 0.3s var(--ease-out);

  small {
    color: rgba(255, 255, 255, 0.55);
    margin-left: 6px;
    font-size: 12px;
  }
}

.wide .isl-ic,
.wide .isl-t {
  opacity: 1;
  transform: none;
  transition-delay: 0.12s;
}
</style>
