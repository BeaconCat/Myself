<script setup lang="ts">
/**
 * 桌面外壳图标：Lucide 标准线性图标（离线）；GitHub 用 simple-icons 品牌实心图标。
 */
import {
  BookOpen, CornerDownLeft, FileText, Hash, History, LayoutGrid, MessageSquareText, Moon, PenLine, Rows3, Rss,
  Search, Settings, Sun, User, X,
} from 'lucide';
import { siGithub } from 'simple-icons';
import { iconSet } from '../../utils/lucide';

const P: Record<string, string> = {
  ...iconSet({
    search: Search,
    history: History,
    enter: CornerDownLeft,
    book: BookOpen,
    bubble: MessageSquareText,
    user: User,
    hash: Hash,
    rss: Rss,
    close: X,
    sun: Sun,
    cards: LayoutGrid,
    clean: Rows3,
    moon: Moon,
    pen: PenLine,
    doc: FileText,
    gear: Settings,
  }),
  github: `<path d="${siGithub.path}" fill="currentColor" stroke="none" transform="translate(1.2 1.2) scale(.9)"/>`,
};

defineProps<{ name: string }>();
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <svg class="ui-ic" viewBox="0 0 24 24" aria-hidden="true" v-html="P[name] ?? ''" />
</template>

<style lang="scss">
.ui-ic {
  width: 20px;
  height: 20px;
  flex: none;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.6;
  stroke-linecap: round;
  stroke-linejoin: round;

  &.s { width: 16px; height: 16px; }
  &.xs { width: 14px; height: 14px; }
}
</style>
