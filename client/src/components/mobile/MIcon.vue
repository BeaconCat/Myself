<script setup lang="ts">
/**
 * 移动端图标：Lucide 标准线性图标（离线）；GitHub 用 simple-icons 品牌实心图标。
 * 选中态由父级加粗描边（见 TabBarFloating）。
 */
import {
  ArrowUp, BookOpen, Check, ChevronLeft, ChevronRight, Clock, History, House, Image, Info, LayoutGrid, Link, List,
  Lock, MessageSquareText, Moon, Rows3, Rss, Search, Share, Sun, Tag, Type, User, X,
} from 'lucide';
import { siGithub } from 'simple-icons';
import { iconSet } from '../../utils/lucide';

const P: Record<string, string> = {
  ...iconSet({
    home: House,
    book: BookOpen,
    bubble: MessageSquareText,
    user: User,
    search: Search,
    image: Image,
    back: ChevronLeft,
    chev: ChevronRight,
    close: X,
    share: Share,
    list: List,
    type: Type,
    rss: Rss,
    info: Info,
    lock: Lock,
    sun: Sun,
    moon: Moon,
    clock: Clock,
    tag: Tag,
    check: Check,
    link: Link,
    cards: LayoutGrid,
    clean: Rows3,
    arrowUp: ArrowUp,
    history: History,
  }),
  github: `<path d="${siGithub.path}" fill="currentColor" stroke="none" transform="translate(1.2 1.2) scale(.9)"/>`,
};

defineProps<{ name: string }>();
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <svg class="m-ic" viewBox="0 0 24 24" aria-hidden="true" v-html="P[name] ?? ''" />
</template>

<style lang="scss">
.m-ic {
  width: 24px;
  height: 24px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.5;
  stroke-linecap: round;
  stroke-linejoin: round;
  flex: none;
  transition: stroke-width var(--dur) var(--ease-out);

  &.s { width: 18px; height: 18px; }
  &.xs { width: 15px; height: 15px; }
}
</style>
