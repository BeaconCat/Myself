<script setup lang="ts">
import type { IconNode } from 'lucide';
import Icon from '../../../components/ui/Icon.vue';

/**
 * 空状态图标：主色轻染的圆角方块里放一枚 Lucide 标准图标，外圈一道极淡的同色环；
 * 入场时轻轻浮起。各页按内容传入图标（素材 ImagePlus、文章 FilePen …）。
 */
defineProps<{ icon: IconNode }>();
</script>

<template>
  <span class="empty-art" aria-hidden="true">
    <Icon :icon="icon" :size="30" :stroke="1.6" />
  </span>
</template>

<style scoped lang="scss">
.empty-art {
  display: grid;
  place-items: center;
  width: 72px;
  height: 72px;
  border-radius: var(--r-xl);
  color: var(--ink);
  background: color-mix(in oklab, var(--primary) 10%, var(--paper));
  box-shadow:
    inset 0 0 0 1px color-mix(in oklab, var(--primary) 18%, transparent),
    0 0 0 8px color-mix(in oklab, var(--primary) 5%, transparent);
  animation: empty-art-in var(--dur-slow) var(--ease-spring) both;
}

@keyframes empty-art-in {
  from { opacity: 0; transform: translateY(8px) scale(0.92); }
}

@media (prefers-reduced-motion: reduce) {
  .empty-art { animation: none; }
}
</style>
