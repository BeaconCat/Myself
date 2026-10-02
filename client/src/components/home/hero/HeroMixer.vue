<script setup lang="ts">
import Icon from '../../../components/ui/Icon.vue';
import { GalleryHorizontalEnd, Pause, Play, SkipForward } from 'lucide';
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import HeroCarousel from '../HeroCarousel.vue';
import type { HeroItem } from './types';
import {
  CARD_CHOREOS,
  DEFAULT_CARD_CHOREO,
  DEFAULT_ROTATE_CHOREO,
  DEFAULT_TEXT_CHOREO,
  ROTATE_CHOREOS,
  TEXT_CHOREOS,
  isCardChoreoId,
  isRotateChoreoId,
  isTextChoreoId,
  type ChoreoMeta,
} from './choreo';
import { prefersReducedMotion } from './useHeroChoreo';

/**
 * 后台「外观」页的 Hero 动效混搭器：
 * 缩放渲染一个真实 HeroCarousel 舞台（与前台同一套组件）+ 三列选择器（文字 / 卡组 / 组内切换）。
 * 深浅与色盘跟随全站主题。支持 v-model:text / v-model:card / v-model:rotate。
 */
const props = defineProps<{
  text?: string;
  card?: string;
  rotate?: string;
}>();

const emit = defineEmits<{
  'update:text': [id: string];
  'update:card': [id: string];
  'update:rotate': [id: string];
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
const rotateId = computed(() => (isRotateChoreoId(props.rotate) ? props.rotate : DEFAULT_ROTATE_CHOREO));

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

function pickRotate(id: string): void {
  if (id !== rotateId.value) emit('update:rotate', id);
}

/** 只播组内切换：不换文章，只让当前卡组轮转一位（当前文章单封面时先切到多封面的一篇） */
function previewRotate(): void {
  carousel.value?.rotate();
}

watch([textId, cardId], () => void preview());
watch(rotateId, () => previewRotate());

function fmt(ms: number): string {
  return `${(ms / 1000).toFixed(2)}s`;
}

const groups = computed<{ key: 'text' | 'card' | 'rotate'; title: string; sub: string; list: ChoreoMeta[]; value: string; pick: (id: string) => void }[]>(() => [
  { key: 'text', title: t('heroLab.textTitle'), sub: t('heroLab.textSub'), list: TEXT_CHOREOS, value: textId.value, pick: pickText },
  { key: 'card', title: t('heroLab.cardTitle'), sub: t('heroLab.cardSub'), list: CARD_CHOREOS, value: cardId.value, pick: pickCard },
  { key: 'rotate', title: t('heroLab.rotateTitle'), sub: t('heroLab.rotateSub'), list: ROTATE_CHOREOS, value: rotateId.value, pick: pickRotate },
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
          :rotate-anim="rotateId"
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
          <Icon :icon="playing ? Pause : Play" />
        </button>
        <button class="ib" :aria-label="t('heroLab.next')" :title="t('heroLab.next')" @click="carousel?.next()">
          <Icon :icon="SkipForward" />
        </button>
        <button class="ib txt" :title="t('heroLab.rotateOnly')" @click="previewRotate">
          <Icon :icon="GalleryHorizontalEnd" />
          <span>{{ t('heroLab.rotateOnly') }}</span>
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
  border-radius: var(--r-lg);
  background: var(--bg);
  box-shadow: inset 0 0 0 0.5px var(--line-2);
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
  border-radius: var(--r-pill);
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line);
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

  &:hover { background: var(--fill-2); color: var(--text); }
  &:active { transform: scale(0.94); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }

  &.play {
    background: var(--text);
    color: var(--bg);

    &:hover { transform: scale(1.06); }
  }

  /* 带文字的胶囊钮：只播组内切换 */
  &.txt {
    width: auto;
    display: inline-flex;
    gap: 6px;
    padding: 0 12px 0 10px;
    border-radius: var(--r-pill);
    font-size: 12.5px;

    svg { width: 16px; height: 16px; }
  }
}

.sep {
  width: 1px;
  height: 20px;
  background: var(--line-2);
  margin: 0 4px;
}

.seg {
  position: relative;
  display: grid;
  grid-template-columns: repeat(3, 52px);
  padding: 2px;
  border-radius: var(--r-pill);
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line);

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

    &.on { color: var(--lift-fg); font-weight: 500; }
    &:not(.on):hover { color: var(--text); }
  }
}

.seg-ind {
  position: absolute;
  top: 2px;
  bottom: 2px;
  left: 2px;
  width: 52px;
  border-radius: var(--r-pill);
  background: var(--lift);
  box-shadow: var(--lift-shadow);
  transition: transform var(--dur) var(--ease-spring), background-color var(--dur);
}

.hint {
  margin-top: -6px;
  font-size: 12.5px;
  color: var(--text-2);
}

/* ===== 选择器 ===== */
.cols {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.col-head {
  display: flex;
  flex-wrap: wrap;
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
  border: 0;
  border-radius: var(--r-md);
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line);
  color: var(--text);
  transition:
    box-shadow var(--dur) var(--ease-out),
    transform var(--dur-fast) var(--ease-spring),
    background-color var(--dur-fast) ease;

  &:hover { background: var(--fill-2); }
  &:active { transform: scale(0.98); }

  /* 选中 = 抬升 + 轻染（与导航 / chip 同一套 token） */
  &.on {
    background: var(--lift);
    box-shadow: var(--lift-shadow);
  }

  &:focus-visible {
    outline: none;
    box-shadow: var(--focus);
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
  font-size: 12px;
  color: var(--text-2);

  &::before {
    content: '#';
    margin-right: 2px;
    color: var(--text-3);
    font-family: var(--font-mono);
    font-size: 0.92em;
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
  border-radius: var(--r-pill);
  background: var(--fill-3);
  overflow: hidden;

  i {
    position: absolute;
    inset: 0 auto 0 0;
    border-radius: inherit;
    background: var(--text-2);
    opacity: 0.5;
    transition: background var(--dur-fast), opacity var(--dur-fast);

    .opt.on & {
      background: var(--ink);
      opacity: 1;
    }
  }
}

.opt-ms {
  flex: none;
  font-family: var(--font-mono);
  font-size: 11.5px;
  font-variant-numeric: tabular-nums;
  color: var(--text-2);
}

@media (max-width: 1100px) {
  .cols { grid-template-columns: 1fr 1fr; }
}

@media (max-width: 900px) {
  .cols { grid-template-columns: 1fr; }
}

@media (max-width: 767px) {
  .bar-ctl { flex-wrap: wrap; max-width: 100%; border-radius: var(--r-md); gap: 6px; }
  .bar-ctl .seg { margin-left: auto; }
  .bar-ctl .sep { display: none; }
  .opt-top { flex-wrap: wrap; gap: 4px 8px; }
}
</style>
