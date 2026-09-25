<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { QuotesData } from '../types';
import type { ModProps } from './props';
import KitIcon from '../parts/KitIcon.vue';
import ModHead from '../parts/ModHead.vue';

/**
 * 语录（quotes）：rotator = 宋体大引文模糊切入、分段进度条（悬停暂停，进度条动画结束即切下一条）、可手动翻页；
 * list = 列表。
 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as QuotesData);
const { t } = useI18n();

const idx = ref(0);
const out = ref(false);
const cur = computed(() => d.value.items[idx.value % Math.max(1, d.value.items.length)]);
const reduce = () => window.matchMedia('(prefers-reduced-motion: reduce)').matches;

function go(i: number): void {
  const n = d.value.items.length;
  if (!n) return;
  const next = ((i % n) + n) % n;
  out.value = true;
  window.setTimeout(() => { idx.value = next; out.value = false; }, reduce() ? 0 : 380);
}

/** 进度条走满 → 下一条（reduced-motion 下不自动轮播） */
function onFilled(i: number): void {
  if (i === idx.value && !reduce()) go(idx.value + 1);
}
</script>

<template>
  <template v-if="variant === 'list'">
    <ModHead :title="title" />
    <ul class="ql">
      <li v-for="(q, i) in d.items" :key="i"><q>{{ q.text }}</q><cite v-if="q.from">— {{ q.from }}</cite></li>
    </ul>
  </template>
  <template v-else>
    <ModHead :title="title">{{ String(idx + 1).padStart(2, '0') }} / {{ String(d.items.length).padStart(2, '0') }}</ModHead>
    <div v-if="cur" class="qr" :class="{ out }" :style="{ '--iv': `${d.interval || 6}s` }">
      <div class="mark" aria-hidden="true">“</div>
      <blockquote>{{ cur.text }}</blockquote>
      <cite v-if="cur.from">— {{ cur.from }}</cite>
      <div class="qr-ctl">
        <div class="pips">
          <button
            v-for="(_, i) in d.items"
            :key="`${i}-${idx}`"
            :class="{ on: i === idx, done: i < idx }"
            :aria-label="t('aboutKit.quotes.nth', { n: i + 1 })"
            @click="go(i)"
            @animationend="onFilled(i)"
          />
        </div>
        <button class="nb" :aria-label="t('aboutKit.quotes.prev')" @click="go(idx - 1)"><KitIcon name="left" :size="18" /></button>
        <button class="nb" :aria-label="t('aboutKit.quotes.next')" @click="go(idx + 1)"><KitIcon name="right" :size="18" /></button>
      </div>
    </div>
  </template>
</template>

<style scoped lang="scss">
.qr { position: relative; display: flex; flex-direction: column; flex: 1; min-height: 200px; }

.mark { height: 36px; font: 700 80px/0.6 var(--font-serif); color: var(--text-3); }

blockquote {
  font: 700 23px/1.6 var(--font-serif);
  letter-spacing: 0.02em;
  transition: opacity 0.5s var(--ease-out), filter 0.5s var(--ease-out), transform 0.5s var(--ease-out);
}

cite { display: block; margin-top: 12px; font: normal 14px var(--font-sans); color: var(--ak-text-3); transition: opacity 0.5s 0.08s; }

.qr.out blockquote { opacity: 0; filter: blur(8px); transform: translateY(-6px); }
.qr.out cite { opacity: 0; }

.qr-ctl {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: auto;
  padding-top: 22px;

  .pips { display: flex; flex: 1; gap: 6px; }

  .pips button {
    position: relative;
    flex: 1;
    max-width: 36px;
    height: 4px;
    padding: 0;
    overflow: hidden;
    border-radius: var(--r-pill);
    background: var(--fill-3);

    &.done { background: var(--text-3); }
  }

  .pips button.on::after {
    content: '';
    position: absolute;
    inset: 0;
    background: var(--ink);
    transform-origin: left;
    animation: qr-fill var(--iv) linear forwards;
  }

  .nb {
    display: grid;
    place-items: center;
    width: 38px;
    height: 38px;
    border-radius: var(--r-pill);
    color: var(--text-2);
    background: var(--fill);
    transition: color var(--dur-fast), background-color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

    &:hover { color: var(--text); background: var(--fill-2); }
    &:active { transform: scale(0.94); }
  }
}

.qr:hover .pips button.on::after { animation-play-state: paused; }

@keyframes qr-fill { from { transform: scaleX(0); } to { transform: scaleX(1); } }

.ql {
  list-style: none;

  li { padding: 16px 0; border-top: 1px solid var(--ak-line); }
  li:first-child { border-top: 0; padding-top: 0; }
  q { font: 700 17px/1.7 var(--font-serif); quotes: '「' '」'; }
  cite { margin-top: 6px; font-size: 13px; }
}

@container (max-width: 360px) { blockquote { font-size: 19px; } }
</style>
