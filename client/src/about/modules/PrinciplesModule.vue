<script setup lang="ts">
import { computed } from 'vue';
import type { PrinciplesData } from '../types';
import type { ModProps } from './props';

/** 信条（principles）：宋体渐隐大号编号 + 一句信条 + 解释；开放排版，通栏四列 / big 变体大字竖排 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as PrinciplesData);
</script>

<template>
  <ol class="pr" :class="{ big: variant === 'big' }" :style="{ '--n': Math.min(4, d.items.length || 1) }">
    <li v-for="(it, i) in d.items" :key="i" :style="{ '--k': i }">
      <span class="n">{{ String(i + 1).padStart(2, '0') }}</span>
      <div>
        <h4>{{ it.title }}</h4>
        <p v-if="it.text">{{ it.text }}</p>
      </div>
    </li>
  </ol>
</template>

<style scoped lang="scss">
.pr {
  display: grid;
  grid-template-columns: repeat(var(--n), minmax(0, 1fr));
  border-top: 1px solid var(--ak-line-2);

  li {
    position: relative;
    list-style: none;
    padding: 26px 28px 6px 0;
    margin-right: 28px;
    border-right: 1px solid var(--ak-line);
    opacity: 0;
    transform: translateY(14px);
    transition: opacity 0.8s var(--ease-out), transform 0.8s var(--ease-out);
    transition-delay: calc(var(--k) * 110ms + 150ms);
  }

  li:last-child { border-right: 0; margin-right: 0; }

  .n {
    display: block;
    margin-bottom: 14px;
    font: 700 82px/1 var(--font-serif);
    letter-spacing: -0.05em;
    color: color-mix(in oklab, var(--text) 18%, transparent);
  }

  h4 { margin-bottom: 10px; font: 700 21px/1.45 var(--font-serif); letter-spacing: 0.01em; }
  p { font-size: 13.5px; line-height: 1.8; color: var(--text-2); }
}

.in .pr li { opacity: 1; transform: none; }

@container (max-width: 900px) {
  .pr { grid-template-columns: 1fr 1fr; }
  .pr li:nth-child(2n) { border-right: 0; margin-right: 0; }
  .pr li:nth-child(n + 3) { border-top: 1px solid var(--ak-line); }
}

@container (max-width: 460px) {
  .pr { grid-template-columns: 1fr; }
  .pr li { display: grid; grid-template-columns: 64px 1fr; padding: 20px 0 16px; margin-right: 0; border-right: 0; border-top: 1px solid var(--ak-line); }
  .pr li:first-child { border-top: 0; }
  .pr .n { font-size: 44px; }
}

.pr.big {
  display: block;
  border-top: 0;

  li {
    display: grid;
    grid-template-columns: 150px 1fr;
    align-items: baseline;
    padding: 26px 0;
    margin: 0;
    border-right: 0;
    border-bottom: 1px solid var(--ak-line);
  }

  .n { margin: 0; font-size: 64px; }
  h4 { font-size: 30px; }
}

@container (max-width: 460px) {
  .pr.big li { grid-template-columns: 64px 1fr; }
  .pr.big .n { font-size: 40px; }
  .pr.big h4 { font-size: 21px; }
}
</style>
