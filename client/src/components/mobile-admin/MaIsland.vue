<script setup lang="ts">
/** 灵动岛式轻提示：顶部黑色胶囊弹性展开，图标弹入、文字上浮，自动收回后淡出。 */
import { island } from './state';
import MaIcon from './MaIcon.vue';
</script>

<template>
  <div class="ma-island" :class="[{ open: island.open }, island.kind]" role="status" aria-live="polite">
    <span :key="island.seq" class="isl-ic">
      <MaIcon :name="island.kind === 'error' ? 'alert' : island.kind === 'info' ? 'spark' : 'check'" :size="14" />
    </span>
    <span :key="`t${island.seq}`" class="isl-t">{{ island.text }}<small v-if="island.sub">{{ island.sub }}</small></span>
  </div>
</template>

<style scoped lang="scss">
.ma-island {
  position: fixed;
  z-index: 90;
  top: calc(var(--safe-t, 10px) + 2px);
  left: 50%;
  width: 118px;
  height: 34px;
  margin-left: -59px;
  border-radius: 20px;
  background: #000;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 14px;
  color: #fff;
  overflow: hidden;
  pointer-events: none;
  opacity: 0;
  transform: translateY(-8px) scale(0.9);
  box-shadow: 0 10px 30px -10px rgba(0, 0, 0, 0.6), inset 0 0 0 0.5px rgba(255, 255, 255, 0.12);
  transition:
    width 0.5s var(--ease-spring),
    margin-left 0.5s var(--ease-spring),
    height 0.5s var(--ease-spring),
    border-radius 0.5s var(--ease-spring),
    opacity 0.3s var(--ease-out) 0.25s,
    transform 0.4s var(--ease-out) 0.2s;

  &.open {
    width: min(320px, calc(100vw - 32px));
    margin-left: calc(min(320px, calc(100vw - 32px)) / -2);
    height: 44px;
    border-radius: 24px;
    opacity: 1;
    transform: none;
    transition:
      width 0.5s var(--ease-spring) 0.06s,
      margin-left 0.5s var(--ease-spring) 0.06s,
      height 0.5s var(--ease-spring) 0.06s,
      border-radius 0.5s var(--ease-spring),
      opacity 0.15s,
      transform 0.35s var(--ease-spring);
  }
}

.isl-ic {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: var(--primary);
  color: var(--on-primary, #fff);
  flex: none;
  opacity: 0;
  transform: scale(0.4);
  transition: all 0.35s var(--ease-spring);

  :deep(.ma-ic) { stroke-width: 2.4; }
}

.error .isl-ic { background: var(--accent-red); color: #fff; }
.info .isl-ic { background: var(--accent-yellow); color: #1a1200; }

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

.open .isl-ic,
.open .isl-t {
  opacity: 1;
  transform: none;
  transition-delay: 0.14s;
}
</style>
