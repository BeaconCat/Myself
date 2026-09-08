<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import type { HeroItem, HeroPhase } from './types';

const props = defineProps<{
  item: HeroItem | undefined;
  /** 条目序号：作 key 重建文字节点，触发入场动画 */
  itemIndex: number;
  phase: HeroPhase;
}>();

const { t } = useI18n();
const router = useRouter();

/**
 * 标题分词：只在空格/标点/`|` 标记处允许换行，词段内绝不从中间断开。
 * `|` 为编辑期换行标记（仅作可断点，不渲染）：一行放得下就完整一行。
 */
interface TitleWord {
  chars: { ch: string; delay: number }[];
  gap: boolean;
}

const titleWords = computed<TitleWord[]>(() => {
  const raw = (props.item?.title ?? '').split('|').join('​');
  const tokens = raw.match(/[^\s​，。：；！？、,.:;!?]+[，。：；！？、,.:;!?]?|\s+|​/g) ?? [];
  const words: TitleWord[] = [];
  let index = 0;
  for (const token of tokens) {
    if (token === '​') {
      words.push({ chars: [], gap: true });
      continue;
    }
    if (/^\s+$/.test(token)) {
      words.push({ chars: [{ ch: ' ', delay: index * 0.035 }], gap: true });
      index += 1;
      continue;
    }
    words.push({
      chars: [...token].map((ch) => ({ ch, delay: (index += 1) * 0.035 })),
      gap: false,
    });
  }
  return words;
});
</script>

<template>
  <!-- 左：大标题 + 简介（phase 挂根节点，退场动画按 .out 触发） -->
  <div class="hero-text" :class="phase">
    <span :key="`tag-${itemIndex}`" class="hero-tag">{{ item?.tag }}</span>
    <h1 class="hero-title" aria-live="polite">
      <span
        v-for="(word, wi) in titleWords"
        :key="`${itemIndex}-${wi}`"
        class="word"
        :class="{ gap: word.gap }"
      ><span
        v-for="(c, ci) in word.chars"
        :key="ci"
        class="char"
        :style="{ '--d': c.delay + 's' }"
      >{{ c.ch }}</span></span>
    </h1>
    <p :key="`ex-${itemIndex}`" class="hero-excerpt">{{ item?.excerpt }}</p>
    <button
      :key="`btn-${itemIndex}`"
      class="hero-btn"
      @click="item?.slug && router.push(`/articles/${item.slug}`)"
    >{{ t('hero.readMore') }}</button>
  </div>
</template>

<style scoped lang="scss">
.hero-text {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 18px;
}

.hero-tag {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  padding: 4px 12px;
  border-radius: 999px;
  color: var(--primary);
  background: rgba(var(--primary-rgb), 0.1);
  border: 1px solid rgba(var(--primary-rgb), 0.25);
}

.hero-title {
  font-size: clamp(30px, 4.4vw, 54px);
  line-height: 1.2;
  min-height: 2.4em;
}

/* 词段整体不拆行；gap（空格/换行标记）处才允许换行 */
.word {
  display: inline-block;
  white-space: nowrap;

  &.gap { display: inline; white-space: normal; }
}

.char {
  display: inline-block;
  white-space: pre;
}

/* 文字入场绑元素挂载（key 随条目重建），不受状态机提前切 idle 影响，动画完整播完 */
.char {
  animation: char-in var(--dur-slow) var(--ease-out) both;
  animation-delay: var(--d);
}

@keyframes char-in {
  from { opacity: 0; filter: blur(12px); transform: translateX(36px); }
  to { opacity: 1; filter: blur(0); transform: none; }
}

.hero-text.out .hero-title,
.hero-text.out .hero-excerpt,
.hero-text.out .hero-tag,
.hero-text.out .hero-btn {
  animation: text-out 0.45s var(--ease-out) both;
}

/* 55% 时间即完全透明：与卡片同步瞬隐，支撑出/入场重叠切换 */
@keyframes text-out {
  55% { opacity: 0; filter: blur(10px); }
  to { opacity: 0; filter: blur(10px); transform: translateX(-48px); }
}

.hero-excerpt {
  font-size: 16px;
  line-height: 1.8;
  color: var(--text-2);
  max-width: 46ch;
}

.hero-excerpt,
.hero-tag {
  animation: fade-up var(--dur-slow) var(--ease-out) both;
  animation-delay: 0.25s;
}

.hero-btn {
  animation: blur-up var(--dur-slow) var(--ease-out) both;
  animation-delay: 0.4s;
}

@keyframes fade-up {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: none; }
}

@keyframes blur-up {
  from { opacity: 0; filter: blur(10px); transform: translateY(16px); }
  to { opacity: 1; filter: blur(0); transform: none; }
}

.hero-btn {
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.04em;
  padding: 12px 32px;
  border-radius: 12px;
  border: none;
  color: #fff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.4);
  background: linear-gradient(180deg, var(--primary), var(--primary-deep));
  box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.45), inset 0 1px 0 rgba(255, 255, 255, 0.25);
  transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast), filter var(--dur-fast);

  &:hover {
    filter: brightness(1.08);
    transform: scale(1.05);
    box-shadow: 0 8px 24px rgba(var(--primary-rgb), 0.55), inset 0 1px 0 rgba(255, 255, 255, 0.25);
  }
}

@media (max-width: 900px) {
  .hero-text { align-items: center; text-align: center; }

  .hero-title {
    font-size: clamp(24px, 6.6vw, 34px);
    /* 移动端标题不占位两行，贴紧下文 */
    min-height: 0;
  }

  .hero-excerpt { font-size: 14px; }
}
</style>
