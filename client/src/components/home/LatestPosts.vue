<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Post } from '../../api';
import CoverArt from '../common/CoverArt.vue';
import SectionHead from './SectionHead.vue';
import { monthDay } from './format';

/**
 * 首页「最新文章」：首篇大卡（16:9 封面 + 宋体大标题）+ 右侧缩略列表（与移动端 feat + m-arow 同构）。
 * 加载中渲染同形骨架，数据到达后 rise-stagger 入场。
 */
const props = defineProps<{ posts: Post[]; total: number; loading: boolean }>();

const { t } = useI18n();

const feat = computed(() => props.posts[0] ?? null);
const rest = computed(() => props.posts.slice(1, 4));

function md(s: string): string {
  return t('home.md', monthDay(s));
}

const sub = computed(() => {
  if (!feat.value) return '';
  return t('home.latestSub', { n: props.total, ...monthDay(feat.value.createdAt) });
});
</script>

<template>
  <section class="latest-sec">
    <SectionHead v-reveal :title="t('home.latest')" :sub="sub" :loading="loading" to="/articles" :link-text="t('home.allPosts')" />

    <!-- 同形骨架 -->
    <div v-if="loading" class="latest" aria-hidden="true">
      <div class="feat">
        <div class="sk cover-sk" />
        <span class="sk sk-line" style="width: 150px; margin-top: 22px" />
        <span class="sk sk-line" style="width: 72%; height: 26px; margin-top: 14px" />
        <span class="sk sk-line" style="width: 90%; margin-top: 14px" />
        <span class="sk sk-line" style="width: 58%; margin-top: 8px" />
      </div>
      <div class="list">
        <div v-for="n in 3" :key="n" class="arow">
          <div class="rt">
            <span class="sk sk-line" style="width: 120px" />
            <span class="sk sk-line" style="width: 70%; height: 18px; margin-top: 12px" />
            <span class="sk sk-line" style="width: 92%; margin-top: 10px" />
          </div>
          <div class="sk thumb" />
        </div>
      </div>
    </div>

    <div v-else-if="feat" v-reveal class="latest rise-stagger">
      <router-link :to="`/articles/${feat.slug}`" class="feat">
        <div class="cover">
          <CoverArt :src="feat.covers[0]" :seed="feat.slug" />
        </div>
        <div class="meta">
          <span v-if="feat.tags[0]" class="tag">{{ feat.tags[0] }}</span>
          <i v-if="feat.tags[0]" class="dotsep" />
          <span>{{ md(feat.createdAt) }}</span>
        </div>
        <h3><span>{{ feat.title }}</span></h3>
        <p>{{ feat.excerpt }}</p>
      </router-link>

      <div class="list rise-stagger">
        <router-link v-for="p in rest" :key="p.id" :to="`/articles/${p.slug}`" class="arow">
          <div class="rt">
            <div class="meta">
              <span v-if="p.tags[0]" class="tag">{{ p.tags[0] }}</span>
              <i v-if="p.tags[0]" class="dotsep" />
              <span>{{ md(p.createdAt) }}</span>
            </div>
            <span class="t"><span>{{ p.title }}</span></span>
            <span class="ex">{{ p.excerpt }}</span>
          </div>
          <div class="thumb">
            <CoverArt :src="p.covers[0]" :seed="p.slug" thumb />
          </div>
        </router-link>
      </div>
    </div>
  </section>
</template>

<style scoped lang="scss">
.latest {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(0, 1fr);
  gap: 56px;
  align-items: start;
}

/* ===== 元信息：「# 标签 · 日期」纯文字 ===== */
.meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0 10px;
  font-size: 13px;
  color: var(--text-3);
}

.tag {
  color: var(--text-2);

  &::before {
    content: '#';
    margin-right: 2px;
    color: var(--text-3);
    font-family: var(--font-mono);
    font-size: 0.92em;
  }
}

.dotsep {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: currentColor;
  opacity: 0.7;
}

/* ===== 首篇大卡 ===== */
.feat {
  display: block;
  border-radius: var(--r-xl);
  color: var(--text);

  .meta { margin-top: 20px; }

  h3 {
    margin-top: 8px;
    font-family: var(--font-serif);
    font-size: 30px;
    font-weight: 700;
    line-height: 1.35;
    text-wrap: balance;

    span {
      background: linear-gradient(currentColor, currentColor) 0 100% / 0 1.5px no-repeat;
      transition: background-size var(--dur-slow) var(--ease-out);
    }
  }

  p {
    margin-top: 10px;
    max-width: 60ch;
    font-size: 15px;
    line-height: 1.8;
    color: var(--text-2);
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  &:hover h3 span { background-size: 100% 1.5px; }
  &:hover .cover { transform: translateY(-4px); box-shadow: var(--shadow-card-hover); }
  &:hover .cover :deep(.cv) { transform: scale(1.035); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

.cover {
  position: relative;
  aspect-ratio: 16 / 9;
  border-radius: var(--r-xl);
  overflow: hidden;
  isolation: isolate;
  background: #040914;
  box-shadow: var(--shadow-card);
  transition: transform var(--dur) var(--ease-out), box-shadow var(--dur) var(--ease-out);

  :deep(.cv) { transition: transform var(--dur-slow) var(--ease-out); }

  /* 发丝内描边：封面与底色分界 */
  &::after {
    content: '';
    position: absolute;
    inset: 0;
    z-index: 3;
    border-radius: inherit;
    box-shadow: inset 0 0 0 0.5px rgb(255 255 255 / 0.08);
    pointer-events: none;
  }
}

:root[data-mode='light'] .cover::after { box-shadow: inset 0 0 0 0.5px rgb(16 24 40 / 0.1); }

/* ===== 缩略列表 ===== */
.list { padding-top: 2px; }

.arow {
  position: relative;
  display: flex;
  align-items: center;
  gap: 24px;
  margin: 0 -18px;
  padding: 20px 18px;
  border-radius: var(--r-lg);
  color: var(--text);
  transition: background-color var(--dur-fast);

  & + &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 18px;
    right: 18px;
    height: 0.5px;
    background: var(--line-2);
    transition: opacity var(--dur-fast);
  }

  &:hover { background: var(--fill); }
  &:hover::before,
  &:hover + &::before { opacity: 0; }
  &:hover .thumb :deep(.cv) { transform: scale(1.05); }
  &:hover .t span { background-size: 100% 1px; }
  &:focus-visible { outline: none; box-shadow: var(--focus); }

  .rt {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .t {
    display: block;
    margin-top: 6px;
    font-family: var(--font-serif);
    font-size: 19px;
    font-weight: 700;
    line-height: 1.45;

    span {
      background: linear-gradient(currentColor, currentColor) 0 100% / 0 1px no-repeat;
      transition: background-size var(--dur) var(--ease-out);
    }
  }

  .ex {
    margin-top: 6px;
    font-size: 14px;
    line-height: 1.7;
    color: var(--text-2);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
}

.thumb {
  position: relative;
  flex: none;
  width: 104px;
  height: 104px;
  border-radius: var(--r-md);
  overflow: hidden;
  isolation: isolate;
  background: #040914;

  :deep(.cv) { transition: transform var(--dur-slow) var(--ease-out); }
}

/* ===== 骨架几何 ===== */
.cover-sk {
  display: block;
  aspect-ratio: 16 / 9;
  border-radius: var(--r-xl);
}

.feat .sk-line,
.arow .sk-line {
  display: block;
  height: 12px;
}

.latest .sk.thumb { background: var(--fill-2); }

@media (max-width: 1100px) {
  .latest {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
    gap: 32px;
  }

  .thumb {
    width: 84px;
    height: 84px;
  }

  .arow .ex { -webkit-line-clamp: 1; }
}

@media (max-width: 768px) {
  .latest {
    grid-template-columns: 1fr;
    gap: 20px;
  }

  .feat h3 { font-size: 24px; }

  .arow {
    gap: 16px;
    padding: 16px 18px;
  }
}
</style>
