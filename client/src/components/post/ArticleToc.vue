<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import type { TocItem } from '../../utils/markdown';
import ContentIcon from './ContentIcon.vue';

/**
 * 详情页右栏阅读卡（高密度）：统计条（字数 / 分钟 / 章节 / 已读%）+ 撑满宽度的阅读进度条与剩余时间
 * → 目录（1px 轨道 + 2px 主色信号线，随当前节弹性滑动）→ 回到顶部。
 */
const props = withDefaults(
  defineProps<{ items: TocItem[]; active: string; progress: number; left: string; words?: number; minutes?: number }>(),
  { words: 0, minutes: 0 },
);
const emit = defineEmits<{ jump: [id: string]; top: [] }>();
const { t } = useI18n();

const list = ref<HTMLElement | null>(null);
const bar = ref({ top: 0, height: 0 });

function placeBar(): void {
  const on = list.value?.querySelector<HTMLElement>('li.on');
  if (!on) {
    bar.value = { top: 0, height: 0 };
    return;
  }
  bar.value = { top: on.offsetTop + 6, height: on.offsetHeight - 12 };
}

watch(() => [props.active, props.items], () => void nextTick(placeBar));
onMounted(() => {
  placeBar();
  window.addEventListener('resize', placeBar);
});
onBeforeUnmount(() => window.removeEventListener('resize', placeBar));
</script>

<template>
  <aside class="toc">
    <div class="stats">
      <div class="stat"><strong>{{ words }}</strong><small>{{ t('dense.article.words') }}</small></div>
      <div class="stat"><strong>{{ minutes }}</strong><small>{{ t('dense.article.minutes') }}</small></div>
      <div class="stat"><strong>{{ items.length }}</strong><small>{{ t('dense.article.sections') }}</small></div>
      <div class="stat"><strong>{{ Math.round(progress * 100) }}<i>%</i></strong><small>{{ t('dense.article.read') }}</small></div>
    </div>

    <div class="rd">
      <span class="track"><i :style="{ transform: `scaleX(${progress})` }" /></span>
      <span class="left">{{ left }}</span>
    </div>

    <template v-if="items.length">
      <h6>{{ t('article.toc') }}</h6>
      <ol ref="list">
        <i class="bar" :style="{ top: `${bar.top}px`, height: `${bar.height}px`, opacity: bar.height ? 1 : 0 }" />
        <li v-for="it in items" :key="it.id" :class="{ on: it.id === active, l3: it.level === 3 }">
          <a :href="`#${it.id}`" @click.prevent="emit('jump', it.id)">{{ it.text }}</a>
        </li>
      </ol>
    </template>

    <button class="totop" @click="emit('top')"><ContentIcon name="up" />{{ t('content.article.toTop') }}</button>
  </aside>
</template>

<style scoped lang="scss">
.toc {
  position: sticky;
  top: calc(var(--nav-h, 64px) + 24px);
  align-self: start;
  display: flex;
  flex-direction: column;
  margin-top: 28px;
  padding: 24px;
  border-radius: var(--r-lg);
  background: var(--elev);
  box-shadow: var(--shadow-card);
  font-size: 14px;
}

/* 统计条：2 x 2 等分格 */
.stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  border-radius: var(--r-md);
  background: var(--fill);
}

.stat {
  min-width: 0;
  padding: 16px 16px 14px;

  &:nth-child(2n) { box-shadow: -1px 0 0 var(--line); }
  &:nth-child(3) { box-shadow: 0 -1px 0 var(--line); }
  &:nth-child(4) { box-shadow: -1px 0 0 var(--line), 0 -1px 0 var(--line); }

  strong {
    display: block;
    font-family: var(--font-mono);
    font-size: 30px;
    line-height: 1.1;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--text);

    i { font-style: normal; font-size: 0.6em; color: var(--text-3); margin-left: 1px; }
  }

  small {
    display: block;
    margin-top: 6px;
    font-size: 12.5px;
    color: var(--text-3);
  }
}

/* 阅读进度：撑满卡宽的细条 + 剩余时间 */
.rd {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 18px;
  font-size: 13px;
  color: var(--text-3);

  .track {
    flex: 1;
    height: 6px;
    overflow: hidden;
    border-radius: var(--r-pill);
    background: var(--fill-2);
  }

  .track i {
    display: block;
    height: 100%;
    border-radius: inherit;
    background: var(--ink);
    transform-origin: left;
    transition: transform 0.15s linear;
  }

  .left { flex: none; white-space: nowrap; }
}

h6 {
  margin: 20px 0 8px;
  font-family: var(--font-serif);
  font-size: 20px;
  font-weight: 700;
  color: var(--text);
}

ol {
  position: relative;
  max-height: calc(100vh - 460px);
  min-height: 60px;
  overflow-y: auto;
  scrollbar-width: none;
  list-style: none;

  &::-webkit-scrollbar { display: none; }

  /* 1px 轨道 */
  &::before {
    content: '';
    position: absolute;
    left: 0;
    top: 4px;
    bottom: 4px;
    width: 1px;
    background: var(--line-2);
  }
}

/* 2px 主色信号线 */
.bar {
  position: absolute;
  left: -0.5px;
  width: 2px;
  border-radius: var(--r-pill);
  background: var(--ink);
  transition: top 0.4s var(--ease-spring), height 0.4s var(--ease-spring), opacity var(--dur);
}

li a {
  position: relative;
  display: block;
  padding: 6px 0 6px 16px;
  line-height: 1.5;
  color: var(--text-2);
  transition: color var(--dur-fast);

  &:hover { color: var(--text); }
  &:focus-visible { outline: none; border-radius: var(--r-xs); box-shadow: var(--focus); }
}

li.l3 a { padding-left: 30px; font-size: 13.5px; color: var(--text-3); }
li.on a { color: var(--text); font-weight: 500; }

.totop {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  height: 38px;
  margin-top: 18px;
  padding: 0 14px;
  border: 0;
  border-radius: var(--r-pill);
  background: var(--fill);
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
  transition: background-color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  &:hover { background: var(--fill-2); }
  &:active { transform: scale(0.97); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

@media (max-width: 1179px) {
  .toc { padding: 20px; }
  .stat { padding: 14px 14px 12px; }
  .stat strong { font-size: 26px; }
}
</style>
