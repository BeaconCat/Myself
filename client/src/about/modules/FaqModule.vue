<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { FaqData } from '../types';
import type { ModProps } from './props';
import ModHead from '../parts/ModHead.vue';

/** 问答（faq）：手风琴 —— grid-template-rows 0fr→1fr 高度动画，加号旋转为减号；single 时同时只展开一条 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as FaqData);
const { t } = useI18n();
const open = ref<Set<number>>(new Set([0]));

function toggle(i: number): void {
  const next = new Set(d.value.single ? [] : open.value);
  if (open.value.has(i)) next.delete(i);
  else next.add(i);
  open.value = next;
}

/** 回答支持 `行内代码` */
const parts = (a: string) => a.split('`').map((text, i) => ({ text, code: i % 2 === 1 }));
</script>

<template>
  <ModHead :title="title">{{ t('aboutKit.faq.count', { n: d.items.length }) }}</ModHead>
  <ul class="fq">
    <li v-for="(it, i) in d.items" :key="i" :class="{ open: open.has(i) }">
      <button :aria-expanded="open.has(i)" @click="toggle(i)">{{ it.q }}<span class="pm" /></button>
      <!-- 收起的答案对读屏与 Tab 都不可达（视觉上用高度动画收起，仍在 DOM 里） -->
      <div class="ans" :aria-hidden="!open.has(i)" :inert="!open.has(i)">
        <div>
          <p><template v-for="(s, k) in parts(it.a)" :key="k"><code v-if="s.code">{{ s.text }}</code><template v-else>{{ s.text }}</template></template></p>
        </div>
      </div>
    </li>
  </ul>
</template>

<style scoped lang="scss">
.fq { list-style: none; }

.fq li { border-top: 1px solid var(--ak-line); }
.fq li:first-child { border-top: 0; }

.fq button {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  width: 100%;
  padding: 15px 0;
  text-align: left;
  font: 700 17px/1.5 var(--font-serif);
  transition: color var(--dur-fast);

  &:hover { color: var(--ak-ink); }
}

.fq li:first-child button { padding-top: 0; }

.pm {
  position: relative;
  flex: none;
  width: 30px;
  height: 30px;
  border-radius: var(--r-pill);
  box-shadow: inset 0 0 0 1px var(--ak-line-2);
  color: var(--text-2);
  transition: background-color var(--dur) var(--ease-out), box-shadow var(--dur), color var(--dur), transform var(--dur) var(--ease-out);

  &::before, &::after {
    content: '';
    position: absolute;
    left: 50%;
    top: 50%;
    width: 10px;
    height: 1.5px;
    margin: -0.75px 0 0 -5px;
    border-radius: var(--r-pill);
    background: currentColor;
    transition: transform var(--dur) var(--ease-spring);
  }

  &::after { transform: rotate(90deg); }
}

/* 展开 = 抬升 + 轻染 */
.fq li.open .pm { background: var(--lift); box-shadow: var(--lift-shadow); color: var(--lift-fg); transform: rotate(180deg); }
.fq li.open .pm::after { transform: rotate(0); }

.ans {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows var(--dur-slow) var(--ease-out);

  > div { overflow: hidden; }

  p {
    padding: 0 46px 16px 0;
    font-size: 15px;
    line-height: 1.85;
    color: var(--text-2);
    opacity: 0;
    transform: translateY(-6px);
    transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-out);
  }
}

.fq li.open .ans { grid-template-rows: 1fr; }
.fq li.open .ans p { opacity: 1; transform: none; transition-delay: 0.1s; }

code { padding: 1px 6px; border-radius: var(--r-xs); font: 500 12.5px var(--ak-mono); background: var(--ak-sunken); box-shadow: inset 0 0 0 1px var(--ak-line); }
</style>
