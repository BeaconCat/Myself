<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import type { HeroItem } from './types';
import type { LineEls, TextEls } from './choreo/types';

/**
 * 左侧文字层：一篇文章一个实例（双缓冲：切换期间新旧两层同时在场）。
 * 标题拆字 + 拆行、摘要拆行；行按挂载后的实际排版分组，宽度变化时重拆。
 */
const props = defineProps<{
  item: HeroItem;
  /** 离场层：不接事件、对读屏隐藏，也不再重拆 */
  leaving?: boolean;
  /** 编舞进行中：推迟重拆（否则动画目标节点被替换） */
  frozen?: boolean;
}>();

const { t } = useI18n();
const router = useRouter();

type TokKind = 'w' | 'sp' | 'br';
interface Tok {
  kind: TokKind;
  text: string;
  chars: string[];
}

/**
 * 标题分词：只在空格 / 标点 / `|` 标记处允许换行，词段内绝不从中间断开。
 * `|` 为编辑期换行标记（仅作可断点，不渲染）。
 */
function tokenizeTitle(raw: string): Tok[] {
  const out: Tok[] = [];
  const parts = raw.match(/[^\s|，。：；！？、,.:;!?]+[，。：；！？、,.:;!?]*|[，。：；！？、,.:;!?]+|\s+|\|/g) ?? [];
  for (const p of parts) {
    if (p === '|') out.push({ kind: 'br', text: '​', chars: [] });
    else if (/^\s+$/.test(p)) out.push({ kind: 'sp', text: ' ', chars: [] });
    else out.push({ kind: 'w', text: p, chars: [...p] });
  }
  return out;
}

/** 摘要分词：中日文逐字可断，西文单词成组，标点随前字 */
function tokenizeBody(raw: string): Tok[] {
  const parts = raw.match(/(?:[A-Za-z0-9\-_.]+|[^\s])[，。：；！？、,.:;!?」）]*|\s+/g) ?? [];
  return parts.map((p) => (/^\s+$/.test(p)
    ? { kind: 'sp' as const, text: ' ', chars: [] }
    : { kind: 'w' as const, text: p, chars: [...p] }));
}

const titleToks = computed(() => tokenizeTitle(props.item.title ?? ''));
const bodyToks = computed(() => tokenizeBody(props.item.excerpt ?? ''));

const allIdx = (n: number): number[] => Array.from({ length: n }, (_, i) => i);
const titleLines = ref<number[][]>([allIdx(titleToks.value.length)]);
const bodyLines = ref<number[][]>([allIdx(bodyToks.value.length)]);
/** 测量态：行容器按 inline 自然流排版，读取每个词段的 offsetTop */
const measuring = ref(true);

const rootEl = ref<HTMLElement | null>(null);
const tagEl = ref<HTMLElement | null>(null);
const titleEl = ref<HTMLElement | null>(null);
const bodyEl = ref<HTMLElement | null>(null);
const btnEl = ref<HTMLElement | null>(null);

function group(container: HTMLElement, toks: Tok[]): number[][] {
  const nodes = Array.from(container.querySelectorAll<HTMLElement>('[data-t]'));
  const groups: number[][] = [];
  let cur: number[] | null = null;
  let top: number | null = null;
  for (const node of nodes) {
    const i = Number(node.dataset.t);
    if (toks[i].kind !== 'w') {
      cur?.push(i);
      continue;
    }
    const y = node.offsetTop;
    if (top === null || Math.abs(y - top) > 4) {
      cur = [];
      groups.push(cur);
      top = y;
    }
    cur!.push(i);
  }
  // 行尾空白 / 断点不进行（nowrap 行宽 = 可见字宽）
  for (const g of groups) {
    while (g.length && toks[g[g.length - 1]].kind !== 'w') g.pop();
  }
  return groups.length ? groups : [[]];
}

let splitting: Promise<void> | null = null;
let lastWidth = 0;

async function split(): Promise<void> {
  measuring.value = true;
  titleLines.value = [allIdx(titleToks.value.length)];
  bodyLines.value = [allIdx(bodyToks.value.length)];
  await nextTick();
  if (!titleEl.value || !bodyEl.value) return;
  titleLines.value = group(titleEl.value, titleToks.value);
  bodyLines.value = group(bodyEl.value, bodyToks.value);
  lastWidth = rootEl.value?.offsetWidth ?? 0;
  measuring.value = false;
  await nextTick();
  snap();
}

/*
 * 像素对齐：编舞期间行 / 字会被提升为合成层，合成层按整像素栅格化；
 * 若静止态文字落在小数像素上，动画结束交还 DOM 的那一帧会出现半像素「跳动」。
 * 这里把整层文字平移到设备像素网格上（行高也按设备像素取整，见样式），
 * 让合成层与静止态栅格化结果一致。移动端（切换时叠放居中）与离场层不处理。
 */
const narrow = window.matchMedia('(max-width: 900px)');

function snap(): void {
  const el = rootEl.value;
  if (!el || props.leaving) return;
  const dpr = window.devicePixelRatio || 1;
  el.style.setProperty('--px', `${1 / dpr}px`);
  el.style.translate = '';
  if (narrow.matches) return;
  const r = el.getBoundingClientRect();
  const fx = (r.left * dpr - Math.round(r.left * dpr)) / dpr;
  const fy = (r.top * dpr - Math.round(r.top * dpr)) / dpr;
  if (Math.abs(fx) > 0.01 || Math.abs(fy) > 0.01) el.style.translate = `${-fx.toFixed(3)}px ${-fy.toFixed(3)}px`;
}

function resplit(): Promise<void> {
  splitting = (splitting ?? Promise.resolve()).then(split);
  return splitting;
}

/** 首次拆行完成（父级在其后采集元素、启动入场） */
const readyPromise = resplit();

let ro: ResizeObserver | null = null;
let pageRo: ResizeObserver | null = null;
let roTimer = 0;
let ready = false;
/** 冻结期间积压的重拆请求 */
let dirty = false;

function requestSplit(): void {
  if (props.leaving) return;
  if (props.frozen || !ready) {
    dirty = true;
    return;
  }
  void resplit();
}

watch(() => props.frozen, (on) => {
  if (!on && dirty) {
    dirty = false;
    requestSplit();
  }
});

onMounted(() => {
  ro = new ResizeObserver(() => {
    const w = rootEl.value?.offsetWidth ?? 0;
    if (!ready || Math.abs(w - lastWidth) < 1) return;
    window.clearTimeout(roTimer);
    roTimer = window.setTimeout(requestSplit, 120);
  });
  if (rootEl.value) ro.observe(rootEl.value);
  // 视口 / 滚动条变化会让整页水平居中偏移半像素：宽度不变也要重新对齐
  pageRo = new ResizeObserver(() => {
    if (ready && !props.frozen) snap();
  });
  pageRo.observe(document.documentElement);
  void readyPromise.then(() => {
    ready = true;
    if (dirty) {
      dirty = false;
      requestSplit();
    }
  });
  // 字体晚到会改变字宽：就绪后重拆一次
  const fonts = document.fonts;
  if (fonts && fonts.status !== 'loaded') void fonts.ready.then(requestSplit);
});

onBeforeUnmount(() => {
  ro?.disconnect();
  pageRo?.disconnect();
  window.clearTimeout(roTimer);
});

function lineEls(container: HTMLElement): LineEls[] {
  return Array.from(container.querySelectorAll<HTMLElement>(':scope > .ln')).map((ln) => {
    const inner = ln.firstElementChild as HTMLElement;
    return { ln, inner, chars: Array.from(inner.querySelectorAll<HTMLElement>('.ch')) };
  });
}

/** 编舞采集：拆行后的真实 DOM */
function els(): TextEls {
  const lines = lineEls(titleEl.value!);
  return {
    root: rootEl.value!,
    tag: tagEl.value!,
    title: titleEl.value!,
    excerpt: bodyEl.value!,
    btn: btnEl.value!,
    lines,
    exLines: lineEls(bodyEl.value!),
    chars: lines.flatMap((l) => l.chars),
  };
}

function open(): void {
  if (props.item.slug) void router.push(`/articles/${props.item.slug}`);
}

/** 等到最近一次拆行（含拆行期间追加的重拆）全部落定 */
async function settled(): Promise<void> {
  await readyPromise;
  let p: Promise<void> | null;
  do {
    p = splitting;
    await p;
  } while (p !== splitting);
}

defineExpose({ els, ready: settled, resplit });
</script>

<template>
  <div
    ref="rootEl"
    class="hero-text"
    :class="{ leaving }"
    :aria-hidden="leaving ? 'true' : undefined"
  >
    <span ref="tagEl" class="hero-eyebrow">
      <span class="hero-tag">{{ item.tag }}</span>
      <template v-if="item.date"><i class="dot" aria-hidden="true" /><span>{{ item.date }}</span></template>
    </span>
    <h1 ref="titleEl" class="hero-title" :class="{ measure: measuring }" :aria-label="item.title.split('|').join('')">
      <span v-for="(ln, li) in titleLines" :key="li" class="ln" aria-hidden="true"><span class="ln-in"><template v-for="ti in ln" :key="ti"><span
        v-if="titleToks[ti].kind === 'w'"
        class="w"
        :data-t="ti"
      ><span v-for="(c, ci) in titleToks[ti].chars" :key="ci" class="ch">{{ c }}</span></span><span
        v-else
        :class="titleToks[ti].kind"
        :data-t="ti"
      >{{ titleToks[ti].text }}</span></template></span></span>
    </h1>
    <p ref="bodyEl" class="hero-excerpt" :class="{ measure: measuring }">
      <span v-for="(ln, li) in bodyLines" :key="li" class="ln"><span class="ln-in"><span
        v-for="ti in ln"
        :key="ti"
        :class="bodyToks[ti].kind"
        :data-t="ti"
      >{{ bodyToks[ti].text }}</span></span></span>
    </p>
    <div ref="btnEl" class="hero-cta">
      <button class="hero-btn primary" :tabindex="leaving ? -1 : undefined" @click="open">
        {{ t('hero.readMore') }}
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M5 12h14M13 6l6 6-6 6" /></svg>
      </button>
      <router-link to="/articles" class="hero-btn ghost" :tabindex="leaving ? -1 : undefined">
        {{ t('home.allPosts') }}
      </router-link>
    </div>
  </div>
</template>

<style scoped lang="scss">
.hero-text {
  position: relative;
  z-index: 5;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 18px;
  min-width: 0;

  &.leaving { pointer-events: none; }
}

/* 眉题：「# 标签 · 日期」纯文字，# 用三级灰 */
.hero-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--text-3);
}

.hero-tag {
  color: var(--text-2);

  &::before {
    content: '#';
    margin-right: 3px;
    color: var(--text-3);
    font-family: var(--font-mono);
    font-size: 0.92em;
  }
}

.dot {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: currentColor;
  opacity: 0.7;
}

.hero-title {
  font-family: var(--font-serif);
  font-weight: 900;
  font-size: clamp(30px, 3.9vw, 52px);
  /* 行高按设备像素取整（--px = 1 / devicePixelRatio，HeroText 写入）：每行都落在像素网格上 */
  line-height: round(1.2em, var(--px, 1px));
  letter-spacing: 0.005em;
  /* 测量态按平衡换行分组：避免「API / 中心」这种末行孤字 */
  text-wrap: balance;
  min-height: calc(2 * round(1.2em, var(--px, 1px)));
  width: 100%;
}

.hero-excerpt {
  font-size: 16px;
  line-height: round(1.8em, var(--px, 1px));
  color: var(--text-2);
  max-width: 30em;
  width: 100%;
}

/* 行：块级容器（光扫 / 基线的定位参照），行内 nowrap 保证字距动画不重排 */
.ln {
  display: block;
  position: relative;
}

.ln-in {
  display: inline-block;
  white-space: nowrap;
  transform-origin: 0 60%;
  /*
   * 遮罩安全区：行遮罩（mask）作用在 border-box 上，字形上伸 / 下伸（y、g、标点）会伸出行高。
   * 上下各留 0.26em、左右 0.12em，再用等量负 margin 抵消：布局、基线与字距完全不变。
   */
  --pad-y: round(0.26em, var(--px, 1px));
  --pad-x: round(0.12em, var(--px, 1px));

  padding: var(--pad-y) var(--pad-x);
  margin: calc(-1 * var(--pad-y)) calc(-1 * var(--pad-x));
}

.w,
.ch {
  display: inline-block;
  white-space: pre;
}

.sp { white-space: pre; }

/* 测量态：行容器退化为 inline，词段按自然流换行 */
.measure {
  .ln,
  .ln-in {
    display: inline;
    white-space: normal;
    padding: 0;
    margin: 0;
  }

  .sp { white-space: normal; }
}

/* 摘要按 CJK 逐字可断：测量态词段 inline 参与自然断行 */
.hero-excerpt .w { display: inline; }

/* CTA：主按钮实底（--solid / --on-solid，按对比度派生），次按钮 ghost；无渐变、无发光、无文字投影 */
.hero-cta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 14px;
}

.hero-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  height: 46px;
  padding: 0 24px;
  border: 0;
  border-radius: var(--r-pill);
  font-size: 15px;
  font-weight: 500;
  line-height: 1;
  white-space: nowrap;
  transition:
    background-color var(--dur-fast) var(--ease-out),
    color var(--dur-fast),
    box-shadow var(--dur-fast),
    transform var(--dur-fast) var(--ease-spring);

  svg {
    width: 16px;
    height: 16px;
    transition: transform var(--dur) var(--ease-spring);
  }

  &:active { transform: scale(0.97); transition-duration: 0.08s; }
  &:focus-visible { outline: none; box-shadow: var(--focus); }

  &.primary {
    background: var(--solid);
    color: var(--on-solid);
    box-shadow: var(--btn-shadow);

    &:hover { background: var(--solid-hover); }
    &:hover svg { transform: translateX(2px); }
    &:focus-visible { box-shadow: var(--btn-shadow), var(--focus); }
  }

  &.ghost {
    background: transparent;
    color: var(--text-2);

    &:hover { background: var(--fill-2); color: var(--text); }
  }
}

@media (max-width: 900px) {
  .hero-text { align-items: center; text-align: center; }

  .hero-title {
    font-size: clamp(24px, 6.6vw, 34px);
    min-height: 0;
  }

  .ln-in { transform-origin: 50% 60%; }

  .hero-excerpt { font-size: 14px; }

  .hero-cta { justify-content: center; }
}
</style>
