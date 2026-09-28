<script lang="ts">
import { coverForKind } from '../../../utils/defaultCovers';

/** 无封面文章在首页轮播里的默认封面种类（历史 `css:<kind>` 取值；现映射到默认封面图） */
export const COVER_KINDS = ['door', 'slit', 'beams', 'season', 'arcs', 'page', 'signal', 'key', 'grid'] as const;
export type CoverKind = (typeof COVER_KINDS)[number];

/**
 * 封面源 → 默认封面种类：
 * `css:<kind>` 显式指定；无图或服务端 SVG 渐变占位（data:image/svg+xml）按内容哈希稳定映射，
 * 真实图片返回 null。
 */
export function coverKindOf(src: string): CoverKind | null {
  if (src.startsWith('css:')) {
    const k = src.slice(4) as CoverKind;
    return COVER_KINDS.includes(k) ? k : 'door';
  }
  if (!src || src.startsWith('data:image/svg+xml') || src.startsWith('/api/v1/img/')) {
    let h = 0;
    for (let i = 0; i < src.length; i++) h = (h * 31 + src.charCodeAt(i)) | 0;
    return COVER_KINDS[Math.abs(h) % COVER_KINDS.length];
  }
  return null;
}
</script>

<script setup lang="ts">
/** 首页轮播卡面：默认封面图（utils/defaultCovers.ts），铺满卡片 */
const props = defineProps<{ kind: CoverKind }>();
</script>

<template>
  <img class="art" :src="coverForKind(props.kind)" alt="" draggable="false" aria-hidden="true" />
</template>

<style scoped lang="scss">
.art {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
</style>
