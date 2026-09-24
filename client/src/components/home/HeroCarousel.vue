<script lang="ts">
/** 对外类型沿用原入口，消费方（HomeView）无需改动导入 */
export type { HeroItem } from './hero/types';
</script>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, toRef, watch } from 'vue';
import type { HeroItem } from './hero/types';
import { useConfigStore } from '../../stores/config';
import { getCardChoreo, getTextChoreo } from './hero/choreo';
import type { CardChoreoId, FxEls, TextChoreoId } from './hero/choreo/types';
import { EASE } from './hero/choreo/engine';
import { useHeroRotation } from './hero/useHeroRotation';
import { useHeroChoreo } from './hero/useHeroChoreo';
import HeroText from './hero/HeroText.vue';
import HeroDeck from './hero/HeroDeck.vue';
import HeroDeckNav from './hero/HeroDeckNav.vue';
import './hero/choreo/choreo.scss';

const props = withDefaults(
  defineProps<{
    items: HeroItem[];
    photoMs?: number;
    /** 覆盖站点配置 hero.textAnim（混搭器用） */
    textAnim?: TextChoreoId | string;
    /** 覆盖站点配置 hero.cardAnim（混搭器用） */
    cardAnim?: CardChoreoId | string;
    /** 全部动效倍速（混搭器 0.25x / 0.5x / 1x） */
    speed?: number;
    /** false = 暂停自动轮播并冻结进行中的切换 */
    playing?: boolean;
  }>(),
  { photoMs: 3000, textAnim: undefined, cardAnim: undefined, speed: 1, playing: true },
);

const config = useConfigStore();
const textChoreo = computed(() => getTextChoreo(props.textAnim ?? config.cfg.hero?.textAnim));
const cardChoreo = computed(() => getCardChoreo(props.cardAnim ?? config.cfg.hero?.cardAnim));
const navMode = computed<'pills' | 'rail'>(() => (cardChoreo.value.chrome === 'rail' ? 'rail' : 'pills'));

/* ===== 双缓冲：切换期间新旧两套内容同时在场 ===== */
interface Slide {
  key: number;
  index: number;
  /** 离场层冻结的组内照片序号 */
  photo: number;
  leaving: boolean;
}

let seq = 0;
const slides = ref<Slide[]>([{ key: seq, index: 0, photo: 0, leaving: false }]);
const busy = ref(false);
/** 首次入场前隐藏（入场首帧由编舞接管） */
const pre = ref(true);

const heroEl = ref<HTMLElement | null>(null);
const textCol = ref<HTMLElement | null>(null);
const fxEl = ref<FxEls | null>(null);
const flareEl = ref<HTMLElement | null>(null);
const floorEl = ref<HTMLElement | null>(null);
const vigEl = ref<HTMLElement | null>(null);
const navRef = ref<InstanceType<typeof HeroDeckNav> | null>(null);

type TextInst = InstanceType<typeof HeroText>;
type DeckInst = InstanceType<typeof HeroDeck>;
const textRefs = new Map<number, TextInst>();
const deckRefs = new Map<number, DeckInst>();

function setTextRef(key: number, inst: unknown): void {
  if (inst) textRefs.set(key, inst as TextInst);
  else textRefs.delete(key);
}

function setDeckRef(key: number, inst: unknown): void {
  if (inst) deckRefs.set(key, inst as DeckInst);
  else deckRefs.delete(key);
}

const {
  itemIndex,
  photoIndex,
  hovering,
  lightboxOn,
  curtainDown,
  covers,
  wrap,
  onFillEnd,
  stepPhoto,
} = useHeroRotation(toRef(props, 'items'), (next) => void swapTo(next));

const speed = toRef(props, 'speed');
/** 编舞时钟：混搭器暂停或幕布未揭开时冻结 */
const clockOn = computed(() => props.playing && !curtainDown.value);
/** 自动步进计时暂停 */
const paused = computed(() => hovering.value || lightboxOn.value || !clockOn.value);

const choreo = useHeroChoreo({ host: heroEl, fx: fxEl, speed, playing: clockOn });

/* ===== 文字列高度：新旧层绝对居中叠放，列高从旧层平滑过渡到新层（移动端下方卡组不跳） ===== */
let colAnim: Animation | null = null;

function lockTextCol(oldEl: HTMLElement | null, newEl: HTMLElement): void {
  const col = textCol.value;
  if (!col) return;
  const hNew = newEl.offsetHeight;
  const hOld = oldEl?.offsetHeight ?? hNew;
  colAnim?.cancel();
  col.classList.add('swapping');
  colAnim = col.animate(
    [{ height: `${hOld}px` }, { height: `${hNew}px` }],
    { duration: 700 / Math.max(0.05, props.speed), easing: EASE.expoOut, fill: 'forwards' },
  );
}

function unlockTextCol(): void {
  colAnim?.cancel();
  colAnim = null;
  textCol.value?.classList.remove('swapping');
}

/** 切换收尾：移除离场层（同一微任务内完成，浏览器不会画出「动画已取消、节点未移除」的中间帧） */
function settle(): void {
  unlockTextCol();
  slides.value = slides.value.filter((s) => !s.leaving);
  busy.value = false;
}

let settingUp = false;

async function swapTo(target: number): Promise<void> {
  if (settingUp || !props.items.length) return;
  const next = wrap(target);
  // 打断进行中的切换：直接跳终态并收尾
  choreo.finish();
  if (next === itemIndex.value) return;
  settingUp = true;
  try {
    const cur = slides.value[slides.value.length - 1];
    cur.leaving = true;
    cur.photo = photoIndex.value;
    const prevIdx = itemIndex.value;
    itemIndex.value = next;
    photoIndex.value = 0;
    busy.value = true;
    const fresh: Slide = { key: ++seq, index: next, photo: 0, leaving: false };
    slides.value.push(fresh);
    await nextTick();
    const oldText = textRefs.get(cur.key);
    const newText = textRefs.get(fresh.key);
    const oldDeck = deckRefs.get(cur.key);
    const newDeck = deckRefs.get(fresh.key);
    if (!newText || !newDeck) {
      settle();
      return;
    }
    await newText.ready();
    const newTextEls = newText.els();
    const oldTextEls = oldText?.els() ?? null;
    lockTextCol(oldTextEls?.root ?? null, newTextEls.root);
    const run = choreo.play({
      text: textChoreo.value,
      card: cardChoreo.value,
      oldText: oldTextEls,
      newText: newTextEls,
      oldDeck: oldDeck?.els() ?? null,
      newDeck: newDeck.els(),
      nav: navRef.value?.nav(prevIdx) ?? null,
    });
    run.onSettle(settle);
  } finally {
    settingUp = false;
  }
}

/** 首次入场：只播 enter（幕布未揭开时编舞时钟冻结在首帧） */
async function intro(): Promise<void> {
  const s = slides.value[0];
  await nextTick();
  const text = textRefs.get(s.key);
  const deck = deckRefs.get(s.key);
  if (!text || !deck) {
    pre.value = false;
    return;
  }
  await text.ready();
  busy.value = true;
  const run = choreo.play({
    text: textChoreo.value,
    card: cardChoreo.value,
    newText: text.els(),
    newDeck: deck.els(),
    nav: navRef.value?.nav(itemIndex.value) ?? null,
  });
  // 首屏没有出场段：跳过入场前的等待（出场让位的那段空拍），揭幕即开演
  run.seek(Math.min(run.firstDelay, 360));
  pre.value = false;
  run.onSettle(() => {
    busy.value = false;
  });
}

/* ===== 键盘：悬停或焦点在 Hero 内时左右切换文章（lightbox 打开时交给 lightbox） ===== */
function onKey(e: KeyboardEvent): void {
  if (lightboxOn.value || (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight')) return;
  const target = e.target as HTMLElement | null;
  if (target?.closest('input, textarea, select, [contenteditable="true"]')) return;
  const hero = heroEl.value;
  if (!hero || !(hovering.value || hero.contains(document.activeElement))) return;
  e.preventDefault();
  void swapTo(itemIndex.value + (e.key === 'ArrowRight' ? 1 : -1));
}

/** lightbox 关闭后按真实指针位置回填悬停态（遮罩期间 mouseleave 不可靠） */
function onLightbox(on: boolean): void {
  lightboxOn.value = on;
  if (on) return;
  const sync = (): void => {
    hovering.value = heroEl.value?.matches(':hover') ?? false;
  };
  sync();
  // 遮罩淡出期间 :hover 仍落在遮罩上：下一次指针移动时再校准一次
  window.addEventListener('pointermove', sync, { once: true });
}

/* 条目数据被整体替换（后台保存等）：回到第一篇 */
watch(() => props.items, (list, old) => {
  if (list === old) return;
  if (itemIndex.value >= list.length) void swapTo(0);
});

onMounted(() => {
  fxEl.value = { flare: flareEl.value!, floor: floorEl.value!, vig: vigEl.value! };
  window.addEventListener('keydown', onKey);
  void intro();
});

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey);
  colAnim?.cancel();
});

defineExpose({
  next: () => swapTo(itemIndex.value + 1),
  prev: () => swapTo(itemIndex.value - 1),
  go: (i: number) => swapTo(i),
});
</script>

<template>
  <section
    ref="heroEl"
    class="hero"
    :class="{ pre, paused }"
    aria-roledescription="carousel"
    @mouseenter="hovering = true"
    @mouseleave="hovering = false"
  >
    <!-- 左：标签 + 大标题 + 摘要 + 按钮（新旧两层叠放） -->
    <div ref="textCol" class="hero-text-col">
      <HeroText
        v-for="s in slides"
        :key="s.key"
        :ref="(inst) => setTextRef(s.key, inst)"
        :item="items[s.index] ?? items[0]"
        :leaving="s.leaving"
        :frozen="busy"
      />
    </div>

    <!-- 右：3D 卡组（新旧两组叠放）+ 常驻导航 -->
    <div class="hero-stage">
      <div class="deck-stack">
        <HeroDeck
          v-for="s in slides"
          :key="s.key"
          :ref="(inst) => setDeckRef(s.key, inst)"
          :covers="(items[s.index] ?? items[0]).covers.slice(0, 3)"
          :photo-index="s.leaving ? s.photo : photoIndex"
          :idx="s.index"
          :title="(items[s.index] ?? items[0]).title"
          :leaving="s.leaving"
          :busy="busy"
          :speed="speed"
          @update:photo-index="photoIndex = $event"
          @update:lightbox="onLightbox"
          @step="stepPhoto"
        />
      </div>

      <HeroDeckNav
        ref="navRef"
        :items="items"
        :item-index="itemIndex"
        :photo-index="photoIndex"
        :photo-count="covers.length"
        :photo-ms="photoMs"
        :speed="speed"
        :paused="paused"
        :busy="busy"
        :mode="navMode"
        @go="swapTo"
        @fill-end="onFillEnd"
      />
    </div>

    <!-- 编舞光效层（静止态全透明） -->
    <div class="hero-fx" aria-hidden="true">
      <div ref="flareEl" class="hc-flare" />
      <div ref="floorEl" class="hc-floor" />
      <div ref="vigEl" class="hc-vig" />
    </div>
  </section>
</template>

<style scoped lang="scss">
/*
 * 层叠：.hero 自成层叠上下文（isolation），内部各层 z 交错：
 * 地面光 4 < 文字 5 < 溢光 8 < 卡片 10–35 < 切换钮 40 < 导航 46 < 共享飞行卡 48/49 < 暗角 59。
 * 因此 hero-text-col / hero-stage / deck-stack 都不能建立层叠上下文。
 */
.hero {
  position: relative;
  isolation: isolate;
  min-height: 62vh;
  display: grid;
  grid-template-columns: 1.05fr 1fr;
  align-items: center;
  gap: 48px;
  /* 展示区禁选中：双击（退出 lightbox 等）不再拉出文字选区 */
  user-select: none;

  &.pre .hero-text-col,
  &.pre .hero-stage { visibility: hidden; }
}

/* 文字列：静止时单层在流；切换时新旧层绝对居中叠放，列高由 JS 过渡 */
.hero-text-col {
  display: grid;
  min-width: 0;

  > * {
    grid-area: 1 / 1;
    align-self: center;
  }

  &.swapping {
    display: block;
    position: relative;

    > * {
      position: absolute;
      left: 0;
      right: 0;
      top: 50%;
      translate: 0 -50%;
    }
  }
}

.hero-stage {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 0;
}

.deck-stack {
  display: grid;
  /* 大屏自适应放大：宽屏时右半区不空 */
  width: min(100%, clamp(430px, 36vw, 580px));

  > * { grid-area: 1 / 1; }
}

/* 暗角覆盖到外层背景壳的留白 */
.hero-fx :deep(.hc-vig) { inset: -36px -56px; }

@media (max-width: 900px) {
  .hero {
    grid-template-columns: 1fr;
    gap: 26px;
    min-height: auto;
  }

  /* 相册收窄居中：两侧各留 ~46px 给后排卡探出与切换钮 */
  .deck-stack { width: min(100% - 92px, 300px); }

  .hero-fx :deep(.hc-vig) { inset: -20px -16px; }
}
</style>
