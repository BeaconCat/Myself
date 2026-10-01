<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import type { Post } from '../../api';
import CoverArt from '../common/CoverArt.vue';
import { attachDrag, clamp, rubber } from './gesture';

/**
 * 首页封面轮播：全宽卡片横滑（跟手 + 两端橡皮筋 + 按速度/位移换页），
 * 相邻卡轻微 rotateY 形成弧面；卡内插画反向视差；进度点即计时器，走满自动翻页。
 */
const props = defineProps<{ posts: Post[]; paused?: boolean; readMin?: (p: Post) => string }>();
const emit = defineEmits<{ open: [slug: string] }>();

const track = ref<HTMLElement | null>(null);
const i = ref(0);
const dx = ref(0);
const dragging = ref(false);
const n = computed(() => props.posts.length);
/** 当前卡文字飞入的随机方向（每次换页刷新） */
const drift = ref({ x: 0, y: 14 });

function go(k: number): void {
  if (!n.value) return;
  i.value = (k + n.value) % n.value;
  dx.value = 0;
  drift.value = { x: Math.round(Math.random() * 40 - 20), y: Math.round(Math.random() * 24 + 6) };
}

watch(n, () => { if (i.value >= n.value) i.value = 0; });

function offset(k: number): number {
  const w = track.value?.clientWidth ?? 350;
  return k - i.value + dx.value / w;
}

let off: (() => void) | null = null;
watch(track, (el) => {
  off?.();
  off = null;
  if (!el) return;
  off = attachDrag(el, {
    axis: 'x',
    stop: true,
    down: (e) => (e.clientX < 24 ? null : {}),
    begin: () => { dragging.value = true; },
    move: (_c, x) => {
      const edge = (i.value === 0 && x > 0) || (i.value === n.value - 1 && x < 0);
      dx.value = edge ? Math.sign(x) * rubber(Math.abs(x), 120) : x;
    },
    end: (_c, vx, _vy, x) => {
      dragging.value = false;
      const w = el.clientWidth;
      let k = i.value;
      if (vx < -0.3 || x < -w / 3) k += 1;
      else if (vx > 0.3 || x > w / 3) k -= 1;
      go(clamp(k, 0, n.value - 1));
    },
  });
});
onBeforeUnmount(() => off?.());

function date(p: Post): string {
  return p.createdAt.slice(0, 10).replace(/-/g, '.');
}
</script>

<template>
  <section class="hc" :class="{ drag: dragging, paused: paused || dragging }">
    <div ref="track" class="track">
      <article
        v-for="(p, k) in posts"
        :key="p.id"
        class="card"
        :class="{ act: k === i }"
        :style="{ '--o': offset(k), '--s': 1 - Math.min(1, Math.abs(offset(k))) * 0.04, '--tx': `${drift.x}px`, '--ty': `${drift.y}px`, zIndex: n - Math.abs(k - i) }"
      >
        <button class="inner m-tap" :tabindex="k === i ? 0 : -1" @click="emit('open', p.slug)">
          <div class="art"><CoverArt :src="p.covers[0]" :seed="p.slug" /></div>
          <div class="shade" />
          <span class="chip">{{ p.tags[0] ?? 'Myself' }}</span>
          <span class="idx">{{ String(k + 1).padStart(2, '0') }} / {{ String(n).padStart(2, '0') }}</span>
          <div class="txt">
            <div class="meta">{{ date(p) }}<template v-if="readMin"> · {{ readMin(p) }}</template></div>
            <h3>{{ p.title }}</h3>
            <p class="post-excerpt">{{ p.excerpt }}</p>
          </div>
        </button>
      </article>
    </div>
    <div class="dots">
      <button
        v-for="(p, k) in posts"
        :key="p.id"
        class="hd"
        :class="{ on: k === i }"
        :aria-label="p.title"
        @click="go(k)"
      >
        <i @animationend="k === i && go(i + 1)" />
      </button>
    </div>
  </section>
</template>

<style scoped lang="scss">
.hc {
  position: relative;
  margin-top: 4px;
  /* 高密度：封面收一档（约 1:1，原 1:1.3），左右 16px 与全站移动边距一致 */
  --w: calc(100vw - 32px);
  --h: min(var(--w), 52vh);
}

.track {
  position: relative;
  height: var(--h);
  perspective: 1100px;
  touch-action: pan-y;
}

.card {
  position: absolute;
  left: 16px;
  top: 0;
  width: var(--w);
  height: var(--h);
  transform:
    translateX(calc(var(--o) * (var(--w) + 12px)))
    rotateY(calc(var(--o) * -10deg))
    scale(var(--s, 1));
  transform-origin: center 60%;
  transition: transform 0.7s var(--ease-out);

  .drag & { transition: none; }
}

.inner {
  position: absolute;
  inset: 0;
  display: block;
  width: 100%;
  border-radius: var(--r-xl);
  overflow: hidden;
  text-align: left;
  color: #fff !important;
  box-shadow: 0 30px 50px -26px rgba(0, 10, 40, 0.9), 0 0 0 0.5px rgba(255, 255, 255, 0.1);
}

html.m-shell[data-mode='light'] .inner {
  box-shadow: 0 30px 50px -24px rgba(20, 40, 90, 0.45);
}

.art {
  position: absolute;
  inset: 0 -48px;
  transform: translateX(calc(var(--o) * -56px));
  transition: transform 0.7s var(--ease-out);

  .drag & { transition: none; }
}

/* 封面压暗层：属于品牌封面光影（中性黑，非主色），保证白字可读 */
.shade {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(2, 6, 14, 0.25) 0%, transparent 22%, transparent 42%, rgba(2, 6, 14, 0.62) 64%, rgba(2, 6, 14, 0.94) 100%);
}

.chip {
  position: absolute;
  left: 18px;
  top: 18px;
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 12.5px;
  font-weight: 500;
  color: #fff;
  background: rgba(255, 255, 255, 0.14);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  box-shadow: inset 0 0 0 0.5px rgba(255, 255, 255, 0.22);
}

.idx {
  position: absolute;
  right: 20px;
  top: 22px;
  font-family: var(--m-font-mono);
  font-size: 11.5px;
  color: rgba(255, 255, 255, 0.6);
  letter-spacing: 0.06em;
}

.txt {
  position: absolute;
  left: 22px;
  right: 22px;
  bottom: 24px;

  > * {
    opacity: 0;
    filter: blur(10px);
    transform: translate(var(--tx, 0), var(--ty, 14px));
    transition: opacity 0.35s var(--ease-out), filter 0.35s var(--ease-out), transform 0.45s var(--ease-out);
  }

  .act & > * {
    opacity: 1;
    filter: none;
    transform: none;
    transition: opacity 0.6s var(--ease-out), filter 0.7s var(--ease-out), transform 0.8s var(--ease-out);
  }

  .act & > :nth-child(1) { transition-delay: 0.12s; }
  .act & > :nth-child(2) { transition-delay: 0.2s; }
  .act & > :nth-child(3) { transition-delay: 0.3s; }
}

.meta {
  font-size: 12.5px;
  color: rgba(255, 255, 255, 0.66);
  margin-bottom: 8px;
  letter-spacing: 0.02em;
}

h3 {
  font-family: var(--font-serif);
  font-size: 26px;
  line-height: 1.3;
  font-weight: 700;
  letter-spacing: 0.01em;
  text-wrap: balance;
}

p {
  margin-top: 10px;
  font-size: 14px;
  line-height: 1.65;
  color: rgba(255, 255, 255, 0.72);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.dots {
  display: flex;
  justify-content: center;
  gap: 6px;
  margin-top: 14px;
}

.hd {
  position: relative;
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: var(--fill-3) !important;
  overflow: hidden;
  transition: width 0.45s var(--ease-spring);

  &.on { width: 30px; }

  i {
    position: absolute;
    inset: 0 auto 0 0;
    width: 0;
    background: var(--ink);
    border-radius: 999px;
  }

  &.on i { animation: fill 6s linear forwards; }

  .paused &.on i { animation-play-state: paused; }
}

@keyframes fill {
  to { width: 100%; }
}

@media (prefers-reduced-motion: reduce) {
  .hd.on i { animation: none; width: 100%; }
}
</style>
