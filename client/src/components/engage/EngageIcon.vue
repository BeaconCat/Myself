<script setup lang="ts">
/** 互动图标：Lucide 标准线性图标（离线）；filled 时实心（已点的喜欢等） */
import { Check, Copy, Heart, MessageCircle, Radio, Share, Smile, SmilePlus, Sparkles } from 'lucide';
import { iconSet } from '../../utils/lucide';

defineProps<{ name: string; filled?: boolean; size?: number }>();

const P: Record<string, string> = iconSet({
  like: Heart,
  spark: Sparkles,
  smile: Smile,
  resonate: Radio,
  comment: MessageCircle,
  react: SmilePlus,
  share: Share,
  copy: Copy,
  check: Check,
});
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <svg
    class="eg-i"
    :class="{ filled }"
    viewBox="0 0 24 24"
    :width="size ?? 18"
    :height="size ?? 18"
    fill="none"
    stroke="currentColor"
    stroke-width="1.6"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
    v-html="P[name] ?? ''"
  />
</template>

<style scoped>
.eg-i { display: block; flex: none; transition: transform var(--dur-fast) var(--ease-spring); }
.eg-i.filled :deep(path) { fill: currentColor; }
</style>
