<script setup lang="ts">
/**
 * 内容区图标：Lucide 标准线性图标（离线）；size：s = 16，xs = 14，默认 18。
 */
defineProps<{ name: string; size?: 's' | 'xs' }>();
</script>

<script lang="ts">
import {
  ArrowLeft, ArrowRight, ArrowUp, Calendar, Check, ChevronLeft, ChevronRight, Copy, Image, Link, List,
  MessageSquareText, Search, Share, Type, X,
} from 'lucide';
import { iconSet } from '../../utils/lucide';

const P: Record<string, string> = iconSet({
  search: Search,
  bubble: MessageSquareText,
  image: Image,
  cal: Calendar,
  arrowL: ArrowLeft,
  arrowR: ArrowRight,
  chevL: ChevronLeft,
  chevR: ChevronRight,
  share: Share,
  link: Link,
  type: Type,
  list: List,
  up: ArrowUp,
  x: X,
  check: Check,
  copy: Copy,
});
export default {};
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <svg class="ci" :class="size" viewBox="0 0 24 24" aria-hidden="true" v-html="P[name]" />
</template>

<style scoped>
.ci {
  width: 18px;
  height: 18px;
  flex: none;
  display: block;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.5;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.ci.s { width: 16px; height: 16px; }
.ci.xs { width: 14px; height: 14px; }
</style>
