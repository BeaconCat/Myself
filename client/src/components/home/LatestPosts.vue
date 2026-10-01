<script setup lang="ts">
import Icon from '../../components/ui/Icon.vue';
import { ArrowRight } from 'lucide';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Post } from '../../api';
import CoverArt from '../common/CoverArt.vue';
import SectionHead from './SectionHead.vue';
import { monthDay } from './format';

/**
 * 首页「最新文章」：7 / 5 并排等高——首篇封面卡（2:1 封面 + 信息块）
 * + 右侧缩略列表卡（行均分高度，底部「全部 N 篇文章」入口压底）。
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
        <div class="sk cover" />
        <div class="info">
          <span class="sk sk-line" style="width: 150px" />
          <span class="sk sk-line" style="width: 72%; height: 24px; margin-top: 12px" />
          <span class="sk sk-line" style="width: 90%; margin-top: 12px" />
        </div>
      </div>
      <div class="list">
        <div v-for="n in 3" :key="n" class="arow">
          <div class="rt">
            <span class="sk sk-line" style="width: 120px" />
            <span class="sk sk-line" style="width: 70%; height: 18px; margin-top: 10px" />
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
        <div class="info">
          <div class="meta">
            <span v-if="feat.tags[0]" class="tag">{{ feat.tags[0] }}</span>
            <i v-if="feat.tags[0]" class="dotsep" />
            <span>{{ md(feat.createdAt) }}</span>
          </div>
          <h3><span>{{ feat.title }}</span></h3>
          <p class="post-excerpt">{{ feat.excerpt }}</p>
        </div>
      </router-link>

      <div class="list">
        <router-link v-for="p in rest" :key="p.id" :to="`/articles/${p.slug}`" class="arow">
          <div class="rt">
            <div class="meta">
              <span v-if="p.tags[0]" class="tag">{{ p.tags[0] }}</span>
              <i v-if="p.tags[0]" class="dotsep" />
              <span>{{ md(p.createdAt) }}</span>
            </div>
            <span class="t"><span>{{ p.title }}</span></span>
            <span class="post-excerpt ex">{{ p.excerpt }}</span>
          </div>
          <div class="thumb">
            <CoverArt :src="p.covers[0]" :seed="p.slug" thumb />
          </div>
        </router-link>
        <router-link to="/articles" class="more">
          <span>{{ t('dense.home.allPosts', { n: total }) }}</span>
          <Icon :icon="ArrowRight" :stroke="2" />
        </router-link>
      </div>
    </div>
  </section>
</template>

<style scoped lang="scss">
.latest {
  display: grid;
  grid-template-columns: minmax(0, 7fr) minmax(0, 5fr);
  gap: var(--card-gap);
  align-items: stretch;
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

/* ===== 卡面：中性，无发光 ===== */
.feat,
.list {
  border-radius: var(--card-r);
  background: var(--card-bg);
  box-shadow: var(--card-shadow);
}

/* ===== 首篇封面卡 ===== */
.feat {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  isolation: isolate;
  color: var(--text);
  transition: transform var(--dur) var(--ease-out), box-shadow var(--dur) var(--ease-out);

  .info {
    display: flex;
    flex: 1;
    flex-direction: column;
    padding: 20px var(--card-pad) var(--card-pad);
  }

  h3 {
    margin-top: 8px;
    font-family: var(--font-serif);
    font-size: 24px;
    font-weight: 700;
    line-height: 1.4;
    text-wrap: balance;

    span {
      background: linear-gradient(currentColor, currentColor) 0 100% / 0 1.5px no-repeat;
      transition: background-size var(--dur-slow) var(--ease-out);
    }
  }

  p {
    margin-top: 8px;
    font-size: 15px;
    line-height: 1.75;
    color: var(--text-2);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  &:hover { transform: translateY(var(--card-rise)); box-shadow: var(--card-shadow-hover); }
  &:hover h3 span { background-size: 100% 1.5px; }
  &:hover .cover :deep(.cv) { transform: scale(1.035); }
  &:focus-visible { outline: none; box-shadow: var(--card-shadow), var(--focus); }
}

.cover {
  position: relative;
  flex: none;
  aspect-ratio: 2 / 1;
  border-radius: var(--card-cover-r);
  overflow: hidden;
  isolation: isolate;
  background: #040914;

  :deep(.cv) { transition: transform var(--dur-slow) var(--ease-out); }

  /* 发丝分界：封面与信息块 */
  &::after {
    content: '';
    position: absolute;
    inset: 0;
    z-index: 3;
    box-shadow: inset 0 -0.5px 0 rgb(255 255 255 / 0.08);
    pointer-events: none;
  }
}

:root[data-mode='light'] .cover::after { box-shadow: inset 0 -0.5px 0 rgb(16 24 40 / 0.1); }

/* ===== 缩略列表卡 ===== */
.list {
  display: flex;
  flex-direction: column;
  padding: 8px var(--card-pad) 20px;
}

.arow {
  position: relative;
  display: flex;
  flex: 1;
  align-items: center;
  gap: 18px;
  margin: 0 -12px;
  padding: 14px 12px;
  border-radius: var(--r-md);
  color: var(--text);
  transition: background-color var(--dur-fast);

  & + &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 12px;
    right: 12px;
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
    margin-top: 4px;
    font-family: var(--font-serif);
    font-size: 19px;
    font-weight: 700;
    line-height: 1.4;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;

    span {
      background: linear-gradient(currentColor, currentColor) 0 100% / 0 1px no-repeat;
      transition: background-size var(--dur) var(--ease-out);
    }
  }

  .ex {
    margin-top: 4px;
    font-size: 15px;
    line-height: 1.6;
    color: var(--text-2);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.thumb {
  position: relative;
  flex: none;
  width: 112px;
  height: 80px;
  border-radius: var(--r-md);
  overflow: hidden;
  isolation: isolate;
  background: #040914;

  :deep(.cv) { transition: transform var(--dur-slow) var(--ease-out); }
}

/* 底部入口：次按钮，压到卡片底部 */
.more {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 40px;
  margin-top: 10px;
  border-radius: var(--r-pill);
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
  background: var(--fill);
  transition: background var(--dur-fast), transform var(--dur-fast) var(--ease-out);

  svg { width: 18px; height: 18px; transition: transform var(--dur-fast) var(--ease-out); }
  &:hover { background: var(--fill-2); }
  &:hover svg { transform: translateX(2px); }
  &:active { transform: scale(0.97); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

/* ===== 简洁风格：封面自带圆角与发丝描边，信息块与列表直接落在页面底色上 ===== */
:root[data-style='clean'] {
  .feat {
    overflow: visible;
    border-radius: 0;

    .info { padding-bottom: 0; }
    &:hover .cover { transform: translateY(-3px); }
  }

  .cover {
    transition: transform var(--dur) var(--ease-out);
    &::after { border-radius: inherit; box-shadow: inset 0 0 0 0.5px rgb(255 255 255 / 0.08); }
  }

  .list { padding-top: 0; padding-bottom: 0; }
}

:root[data-mode='light'][data-style='clean'] .cover::after { box-shadow: inset 0 0 0 0.5px rgb(16 24 40 / 0.1); }

/* ===== 骨架几何 ===== */
.feat .sk.cover { border-radius: var(--card-cover-r); }

.feat .sk-line,
.arow .sk-line {
  display: block;
  height: 12px;
}

.latest .sk.thumb { background: var(--fill-2); }

@media (max-width: 1100px) {
  .latest { grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); }

  .thumb {
    width: 88px;
    height: 68px;
  }
}

@media (max-width: 768px) {
  .latest { grid-template-columns: 1fr; }
  .feat h3 { font-size: 22px; }
}
</style>
