<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import HeroCarousel from '../HeroCarousel.vue';
import type { HeroItem } from './types';
import {
  CARD_CHOREOS,
  DEFAULT_CARD_CHOREO,
  DEFAULT_TEXT_CHOREO,
  TEXT_CHOREOS,
  isCardChoreoId,
  isTextChoreoId,
  type ChoreoMeta,
} from './choreo';
import { prefersReducedMotion } from './useHeroChoreo';

/**
 * 后台「外观」页的 Hero 动效混搭器：
 * 缩放渲染一个真实 HeroCarousel 舞台（与前台同一套组件）+ 两列选择器（文字 / 卡组）。
 * 深浅与色盘跟随全站主题。
 */
const props = defineProps<{
  text?: string;
  card?: string;
}>();

const emit = defineEmits<{
  'update:text': [id: string];
  'update:card': [id: string];
}>();

const { t } = useI18n();

/** 示例文章（与 hero-motion 原型同内容，封面为 CSS 光影构成） */
const SAMPLES: HeroItem[] = [
  {
    title: '欢迎使用 Myself',
    tag: '指南',
    excerpt: '一扇为写作者打开的门：文章、随想与关于页由同一套 Markdown 贯穿，部署只需要一个二进制文件。',
    covers: ['css:door', 'css:slit', 'css:beams'],
  },
  {
    title: '主题系统指南',
    tag: '主题',
    excerpt: '四季色盘 × 深浅模式。只给一个主色，整套明暗变量在运行时自动派生，后台最多保存十组预设。',
    covers: ['css:season', 'css:arcs'],
  },
  {
    title: '用 Markdown 写作',
    tag: '写作',
    excerpt: '标题、代码高亮、目录锚点与配图，所见即所存。TipTap 负责编辑，Markdown 负责落盘，迁移永远自由。',
    covers: ['css:page'],
  },
  {
    title: '把发文托管给 AI：API 中心',
    tag: 'API',
    excerpt: '创建 APIKey，让外部 AI 通过 X-Api-Key 认证的 REST 接口替你发布文章与随想。',
    covers: ['css:signal', 'css:key', 'css:grid'],
  },
];

const textId = computed(() => (isTextChoreoId(props.text) ? props.text : DEFAULT_TEXT_CHOREO));
const cardId = computed(() => (isCardChoreoId(props.card) ? props.card : DEFAULT_CARD_CHOREO));

const playing = ref(true);
const SPEEDS = [0.25, 0.5, 1] as const;
const speed = ref<number>(1);
const reduced = ref(false);

const carousel = ref<InstanceType<typeof HeroCarousel> | null>(null);

/* ===== 缩放舞台：按设计宽度渲染真实 Hero，再整体缩放进容器 ===== */
const DESIGN_W = 1240;
const DESIGN_H = 600;
const frameEl = ref<HTMLElement | null>(null);
const k = ref(0.6);
let ro: ResizeObserver | null = null;

onMounted(() => {
  reduced.value = prefersReducedMotion();
  ro = new ResizeObserver(() => {
    const w = frameEl.value?.clientWidth ?? 0;
    if (w) k.value = Math.min(1, w / DESIGN_W);
  });
  if (frameEl.value) ro.observe(frameEl.value);
});

onBeforeUnmount(() => ro?.disconnect());

/** 选中即预览：恢复播放并立刻切到下一篇 */
async function preview(): Promise<void> {
  playing.value = true;
  await nextTick();
  void carousel.value?.next();
}

function pickText(id: string): void {
  if (id !== textId.value) emit('update:text', id);
}

function pickCard(id: string): void {
  if (id !== cardId.value) emit('update:card', id);
}

watch([textId, cardId], () => void preview());

function fmt(ms: number): string {
  return `${(ms / 1000).toFixed(2)}s`;
}

const groups = computed<{ key: 'text' | 'card'; title: string; sub: string; list: ChoreoMeta[]; value: string; pick: (id: string) => void }[]>(() => [
  { key: 'text', title: t('heroLab.textTitle'), sub: t('heroLab.textSub'), list: TEXT_CHOREOS, value: textId.value, pick: pickText },
  { key: 'card', title: t('heroLab.cardTitle'), sub: t('heroLab.cardSub'), list: CARD_CHOREOS, value: cardId.value, pick: pickCard },
]);

/** 时长条满格参照 */
const MAX_MS = 1600;
</script>

<template>
  <div class="mixer">
    <!-- 舞台 -->
    <div ref="frameEl" class="frame" :style="{ height: `${DESIGN_H * k}px` }">
      <div class="canvas" :style="{ width: `${DESIGN_W}px`, height: `${DESIGN_H}px`, transform: `scale(${k})` }">
        <HeroCarousel
          ref="carousel"
          :items="SAMPLES"
          :photo-ms="2600"
          :text-anim="textId"
          :card-anim="cardId"
          :speed="speed"
          :playing="playing"
        />
      </div>
    </div>

    <!-- 播放控制 -->
    <div class="bar">
      <span class="bar-title">{{ t('heroLab.preview') }}</span>
      <div class="bar-ctl">
        <button
          class="ib play"
          :aria-label="playing ? t('heroLab.pause') : t('heroLab.play')"
          :title="playing ? t('heroLab.pause') : t('heroLab.play')"
          @click="playing = !playing"
        >
          <svg v-if="playing" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M9 6v12M15 6v12" /></svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M8 5.5v13l11-6.5z" /></svg>
        </button>
        <button class="ib" :aria-label="t('heroLab.next')" :title="t('heroLab.next')" @click="carousel?.next()">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M8 6l6 6-6 6" /><path d="M17 6v12" /></svg>
        </button>
        <span class="sep" />
        <div class="seg" role="radiogroup" :aria-label="t('heroLab.speed')">
          <span class="seg-ind" :style="{ transform: `translateX(${SPEEDS.indexOf(speed as 1) * 100}%)` }" />
          <button
            v-for="s in SPEEDS"
            :key="s"
            role="radio"
            :aria-checked="speed === s"
            :class="{ on: speed === s }"
            @click="speed = s"
          >{{ s }}x</button>
        </div>
      </div>
    </div>
    <p class="hint">{{ reduced ? t('heroLab.reduced') : t('heroLab.hint') }}</p>

    <!-- 两列选择器 -->
    <div class="cols">
      <section v-for="g in groups" :key="g.key" class="col">
        <header class="col-head">
          <h3>{{ g.title }}</h3>
          <span>{{ g.sub }}</span>
        </header>
        <div class="opts" role="radiogroup" :aria-label="g.title">
          <button
            v-for="m in g.list"
            :key="m.id"
            class="opt"
            role="radio"
            :aria-checked="g.value === m.id"
            :class="{ on: g.value === m.id }"
            @click="g.pick(m.id)"
          >
            <span class="opt-top">
              <b class="opt-name">{{ m.name }}</b>
              <span class="opt-tag">{{ m.tag }}</span>
            </span>
            <span class="opt-desc">{{ m.desc }}</span>
            <span class="opt-dur">
              <span class="opt-track"><i :style="{ width: `${Math.min(100, (m.duration / MAX_MS) * 100)}%` }" /></span>
              <span class="opt-ms">{{ fmt(m.duration) }}</span>
            </span>
          </button>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped lang="scss">
.mixer {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

/* ===== 舞台 ===== */
.frame {
  position: relative;
  width: 100%;
  overflow: hidden;
  border-radius: 20px;
  border: 1px solid rgba(var(--primary-rgb), 0.12);
  background:
    radial-gradient(560px 300px at 82% 24%, rgba(var(--primary-rgb), 0.08), transparent 65%),
    radial-gradient(480px 260px at 12% 80%, rgba(var(--primary-rgb), 0.05), transparent 65%),
    linear-gradient(180deg, rgba(var(--primary-rgb), 0.045), rgba(var(--primary-rgb), 0.015)),
    var(--bg);
  transition: height var(--dur) var(--ease-out);
}

.canvas {
  position: absolute;
  left: 0;
  top: 0;
  padding: 36px 56px;
  transform-origin: 0 0;

  :deep(.hero) {
    min-height: 0;
    height: 100%;
  }
}

/* ===== 播放控制 ===== */
.bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.bar-title {
  font-family: var(--font-serif);
  font-weight: 700;
  font-size: 15px;
  letter-spacing: 0.02em;
}

.bar-ctl {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px;
  border-radius: 999px;
  background: var(--glass);
  border: 1px solid var(--border);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
}

.ib {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  border: 0;
  background: none;
  color: var(--text-2);
  display: grid;
  place-items: center;
  transition: background var(--dur-fast), color var(--dur-fast), transform var(--dur-fast) var(--ease-out);

  svg { width: 18px; height: 18px; }

  &:hover { background: rgba(var(--primary-rgb), 0.1); color: var(--text); }
  &:active { transform: scale(0.92); }

  &.play {
    background: var(--text);
    color: var(--bg);

    &:hover { transform: scale(1.06); }
  }
}

.sep {
  width: 1px;
  height: 20px;
  background: var(--border);
  margin: 0 4px;
}

.seg {
  position: relative;
  display: grid;
  grid-template-columns: repeat(3, 52px);
  padding: 2px;
  border-radius: 999px;
  background: rgba(var(--primary-rgb), 0.06);

  button {
    position: relative;
    z-index: 1;
    border: 0;
    background: none;
    padding: 6px 0;
    border-radius: 999px;
    font-size: 12.5px;
    font-variant-numeric: tabular-nums;
    color: var(--text-2);
    transition: color var(--dur-fast);

    &.on { color: #fff; }
    &:not(.on):hover { color: var(--text); }
  }
}

.seg-ind {
  position: absolute;
  top: 2px;
  bottom: 2px;
  left: 2px;
  width: 52px;
  border-radius: 999px;
  background: linear-gradient(180deg, var(--primary), var(--primary-deep));
  box-shadow: 0 4px 12px rgba(var(--primary-rgb), 0.4);
  transition: transform var(--dur) var(--ease-spring);
}

.hint {
  margin-top: -6px;
  font-size: 12.5px;
  color: var(--text-2);
}

/* ===== 选择器 ===== */
.cols {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.col-head {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 10px;

  h3 { font-size: 16px; letter-spacing: 0.02em; }
  span { font-size: 12.5px; color: var(--text-2); }
}

.opts {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.opt {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
  text-align: left;
  padding: 12px 14px;
  border-radius: 14px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  transition:
    border-color var(--dur-fast) ease,
    box-shadow var(--dur) var(--ease-out),
    transform var(--dur-fast) var(--ease-out),
    background var(--dur-fast) ease;

  &:hover {
    transform: translateY(-1px);
    border-color: rgba(var(--primary-rgb), 0.35);
  }

  &.on {
    border-color: var(--primary);
    background: color-mix(in oklab, var(--primary) 6%, var(--surface));
    box-shadow: 0 0 0 3px rgba(var(--primary-rgb), 0.14), 0 10px 26px -14px rgba(var(--primary-rgb), 0.55);
  }

  &:focus-visible {
    outline: 2px solid var(--primary);
    outline-offset: 2px;
  }
}

.opt-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.opt-name {
  font-family: var(--font-serif);
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.opt-tag {
  flex: none;
  font-size: 11.5px;
  font-weight: 700;
  letter-spacing: 0.06em;
  padding: 2px 9px;
  border-radius: 999px;
  color: var(--text-2);
  border: 1px solid var(--border);

  .opt.on & {
    color: var(--primary);
    border-color: rgba(var(--primary-rgb), 0.35);
    background: rgba(var(--primary-rgb), 0.08);
  }
}

.opt-desc {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-2);
}

.opt-dur {
  display: flex;
  align-items: center;
  gap: 10px;
}

.opt-track {
  position: relative;
  flex: 1;
  height: 3px;
  border-radius: 3px;
  background: var(--border);
  overflow: hidden;

  i {
    position: absolute;
    inset: 0 auto 0 0;
    border-radius: inherit;
    background: var(--text-2);
    opacity: 0.5;
    transition: background var(--dur-fast), opacity var(--dur-fast);

    .opt.on & {
      background: linear-gradient(90deg, var(--primary), var(--primary-deep));
      opacity: 1;
    }
  }
}

.opt-ms {
  flex: none;
  font-family: ui-monospace, 'Cascadia Mono', Consolas, monospace;
  font-size: 11.5px;
  font-variant-numeric: tabular-nums;
  color: var(--text-2);
}

@media (max-width: 900px) {
  .cols { grid-template-columns: 1fr; }
}
</style>
