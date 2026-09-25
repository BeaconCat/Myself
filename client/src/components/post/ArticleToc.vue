<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import type { TocItem } from '../../utils/markdown';
import ContentIcon from './ContentIcon.vue';

/**
 * 详情页右侧目录：1px 轨道 + 2px 主色信号线（随当前节弹性滑动）；
 * 底部阅读进度环 + 百分比 + 剩余时间；回到顶部。
 */
const props = defineProps<{ items: TocItem[]; active: string; progress: number; left: string }>();
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

/* 进度环周长：r = 9 */
const C = 2 * Math.PI * 9;
</script>

<template>
  <aside class="toc">
    <template v-if="items.length">
      <h6>{{ t('article.toc') }}</h6>
      <ol ref="list">
        <i class="bar" :style="{ top: `${bar.top}px`, height: `${bar.height}px`, opacity: bar.height ? 1 : 0 }" />
        <li v-for="it in items" :key="it.id" :class="{ on: it.id === active, l3: it.level === 3 }">
          <a :href="`#${it.id}`" @click.prevent="emit('jump', it.id)">{{ it.text }}</a>
        </li>
      </ol>
    </template>
    <div class="rd" :class="{ solo: !items.length }">
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <circle class="bgc" cx="12" cy="12" r="9" />
        <circle class="fgc" cx="12" cy="12" r="9" :style="{ strokeDasharray: C, strokeDashoffset: C * (1 - progress) }" />
      </svg>
      <div><b>{{ Math.round(progress * 100) }}%</b><span>{{ left }}</span></div>
    </div>
    <button class="totop" @click="emit('top')"><ContentIcon name="up" size="s" />{{ t('content.article.toTop') }}</button>
  </aside>
</template>

<style scoped lang="scss">
.toc {
  position: sticky;
  top: calc(var(--nav-h, 64px) + 40px);
  align-self: start;
  margin-top: 40px;
  font-size: 13.5px;
}

h6 {
  margin-bottom: 12px;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.1em;
  color: var(--text-3);
}

ol {
  position: relative;
  max-height: calc(100vh - 340px);
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
  color: var(--text-3);
  transition: color var(--dur-fast);

  &:hover { color: var(--text); }
  &:focus-visible { outline: none; border-radius: var(--r-xs); box-shadow: var(--focus); }
}

li.l3 a { padding-left: 30px; font-size: 13px; }
li.on a { color: var(--text); font-weight: 500; }

.rd {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 22px;
  padding-top: 18px;
  box-shadow: inset 0 0.5px 0 var(--line);
  font-size: 12.5px;
  color: var(--text-3);

  &.solo { margin-top: 0; padding-top: 0; box-shadow: none; }

  svg {
    width: 26px;
    height: 26px;
    transform: rotate(-90deg);
  }

  circle { fill: none; stroke-width: 2.4; }
  .bgc { stroke: var(--fill-3); }
  .fgc { stroke: var(--ink); stroke-linecap: round; transition: stroke-dashoffset 0.15s linear; }

  b {
    display: block;
    font-family: var(--font-mono);
    font-size: 12.5px;
    font-weight: 500;
    color: var(--text);
  }
}

.totop {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  height: 32px;
  margin: 14px 0 0 -10px;
  padding: 0 12px 0 10px;
  border: 0;
  border-radius: var(--r-pill);
  background: none;
  font-size: 13px;
  color: var(--text-2);
  transition: background-color var(--dur-fast), color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  &:hover { background: var(--fill-2); color: var(--text); }
  &:active { transform: scale(0.97); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}
</style>
