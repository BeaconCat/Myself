<script setup lang="ts">
import type { Post } from '../../api';
import { placeholderCover } from '../../utils/placeholder';

/** alternate：交错翻转（首页）；compact：窄行（文章页，统一文左图右） */
withDefaults(
  defineProps<{ posts: Post[]; alternate?: boolean; compact?: boolean }>(),
  { alternate: true, compact: false },
);

const fallbackPalette = [
  ['#ff0032', '#7a0020'],
  ['#ffb300', '#7a5200'],
  ['#0078ff', '#00295c'],
  ['#00c853', '#00512a'],
];

function coverOf(post: Post, index: number): string {
  if (post.covers.length) return post.covers[0];
  const [from, to] = fallbackPalette[index % fallbackPalette.length];
  return placeholderCover(from, to, post.tags[0] ?? 'Post');
}
</script>

<template>
  <div class="zig-list">
    <router-link
      v-for="(post, i) in posts"
      :key="post.slug"
      :to="`/articles/${post.slug}`"
      class="zig"
      :class="{ flip: alternate && i % 2 === 1, compact }"
      :style="{ '--i': i % 20 }"
    >
      <div class="zig-text">
        <span class="z-date">{{ post.createdAt.slice(0, 10) }}</span>
        <!-- data-title 供 ::after 渐变层复写同一文本，颜色淡入不重排 -->
        <h3 :data-title="post.title">{{ post.title }}</h3>
        <p>{{ post.excerpt }}</p>
        <div class="z-tags">
          <span v-for="tag in post.tags" :key="tag">{{ tag }}</span>
        </div>
      </div>
      <div class="zig-media">
        <img :src="coverOf(post, i)" :alt="post.title" loading="lazy" draggable="false" />
      </div>
    </router-link>
  </div>
</template>

<style scoped lang="scss">
.zig {
  display: grid;
  grid-template-columns: 1.15fr 1fr;
  gap: 40px;
  align-items: center;
  padding: 34px 28px;
  position: relative;
  transition: background var(--dur-fast);
  /* 逐条浮入（追加加载的新条目挂载时同样生效） */
  animation: zig-in 0.55s var(--ease-out) both;
  animation-delay: calc(var(--i, 0) * 0.06s);

  /* 偶数条翻转：图左文右 */
  &.flip .zig-text { order: 2; }
  &.flip .zig-media { order: 1; }

  /* 悬停直角边框：四角括号式描边 */
  &::before,
  &::after {
    content: '';
    position: absolute;
    width: 26px;
    height: 26px;
    opacity: 0;
    transition: opacity var(--dur-fast) ease, transform var(--dur) var(--ease-out);
    pointer-events: none;
  }

  &::before {
    top: 10px;
    left: 10px;
    border-top: 2px solid var(--primary);
    border-left: 2px solid var(--primary);
    transform: translate(8px, 8px);
  }

  &::after {
    bottom: 10px;
    right: 10px;
    border-bottom: 2px solid var(--primary);
    border-right: 2px solid var(--primary);
    transform: translate(-8px, -8px);
  }

  &:hover {
    background: rgba(var(--primary-rgb), 0.04);

    &::before, &::after { opacity: 1; transform: none; }
    .zig-media img { transform: scale(1.04); }

    /* 基色同步转透明，消除渐变层叠加出的白边 */
    h3 { color: transparent; }
    h3::after { opacity: 1; }
  }
}

.zig-text {
  min-width: 0;

  .z-date {
    font-size: 13px;
    color: var(--text-2);
    font-variant-numeric: tabular-nums;
  }

  /* 标题：渐变层用 ::after 同文本覆盖淡入，仅颜色过渡、文本不动不闪 */
  h3 {
    position: relative;
    font-size: clamp(24px, 2.8vw, 32px);
    margin: 10px 0 12px;
    color: var(--text);
    transition: color var(--dur) ease;

    &::after {
      content: attr(data-title);
      position: absolute;
      inset: 0;
      background: var(--grad-title);
      background-clip: text;
      -webkit-background-clip: text;
      color: transparent;
      opacity: 0;
      transition: opacity var(--dur) ease;
      pointer-events: none;
    }
  }

  p {
    font-size: 14px;
    line-height: 1.8;
    color: var(--text-2);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
}

.z-tags {
  display: flex;
  gap: 8px;
  margin-top: 14px;

  span {
    font-size: 11px;
    font-weight: 600;
    padding: 2px 10px;
    background: rgba(var(--primary-rgb), 0.1);
    color: var(--primary);
  }
}

/* 窄行模式：文章页统一文左图右 */
.zig.compact {
  gap: 28px;
  padding: 20px 24px;

  .zig-text h3 { font-size: clamp(21px, 2.3vw, 26px); margin: 8px 0 8px; }
  .zig-text p { -webkit-line-clamp: 1; }
  .zig-media { aspect-ratio: 21 / 9; }
}

.zig-media {
  overflow: hidden;
  aspect-ratio: 16 / 9;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    transition: transform var(--dur-slow) var(--ease-out);
  }
}

@keyframes zig-in {
  from { opacity: 0; transform: translateY(26px); }
  to { opacity: 1; transform: none; }
}

@media (max-width: 768px) {
  .zig {
    grid-template-columns: 1fr;
    gap: 18px;
    padding: 22px 14px;

    /* 移动端统一图上文下 */
    &.flip .zig-text, .zig-text { order: 2; }
    &.flip .zig-media, .zig-media { order: 1; }
  }
}
</style>
