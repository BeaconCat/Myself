<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { BookshelfData } from '../types';
import type { ModProps } from './props';
import ModHead from '../parts/ModHead.vue';

/** 书架（bookshelf）：书脊视图（竖排书名、高矮宽窄不一），在读挂红色书签；点击抽出书并在右侧显示书卡 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as BookshelfData);
const { t } = useI18n();

const picked = ref<number | null>(null);
const sel = computed(() => {
  if (picked.value != null && picked.value < d.value.items.length) return picked.value;
  const reading = d.value.items.findIndex((b) => b.status === '在读');
  return reading >= 0 ? reading : 0;
});
const book = computed(() => d.value.items[sel.value]);
const doneCount = computed(() => d.value.items.filter((b) => b.status === '读完').length);
const readingCount = computed(() => d.value.items.filter((b) => b.status === '在读').length);
</script>

<template>
  <ModHead :title="title">{{ t('aboutKit.bookshelf.meta', { done: doneCount, reading: readingCount }) }}</ModHead>
  <div class="bs">
    <div class="shelf">
      <button
        v-for="(b, i) in d.items"
        :key="i"
        class="spine"
        :class="{ on: i === sel, reading: b.status === '在读', lean: b.lean }"
        :style="{ '--c': b.color, '--tc': b.textColor || '#f4efe4', '--h': `${b.height ?? 88}%`, '--w': `${b.width ?? 34}px`, '--k': i }"
        :aria-label="b.title"
        :aria-pressed="i === sel"
        @click="picked = i"
      >
        <span class="t">{{ b.title }}</span>
        <span class="a">{{ b.author }}</span>
      </button>
    </div>
    <div v-if="book" class="bs-card">
      <div :key="sel" class="bs-anim">
        <small>{{ book.status }}</small>
        <b>{{ book.title }}</b>
        <span>{{ book.author }}</span>
        <p v-if="book.note">「{{ book.note }}」</p>
        <div v-if="book.progress" class="pg"><i :style="{ '--w': `${book.progress}%` }" />{{ book.progress }}%</div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.bs {
  display: grid;
  grid-template-columns: minmax(0, 1fr) clamp(200px, 36cqi, 270px);
  gap: 20px;
  align-items: end;
}

.shelf {
  position: relative;
  display: flex;
  align-items: flex-end;
  gap: 5px;
  height: 230px;
  padding: 0 10px;

  &::after {
    content: '';
    position: absolute;
    left: -6px;
    right: -6px;
    bottom: -12px;
    height: 12px;
    border-radius: calc(var(--r-xs) * 0.6);
    background: linear-gradient(180deg, var(--ak-line-2), var(--ak-sunken));
    box-shadow: 0 14px 24px -12px rgb(0 0 0 / 0.5);
  }
}

.spine {
  position: relative;
  display: flex;
  flex: none;
  flex-direction: column;
  align-items: center;
  width: var(--w);
  height: var(--h);
  padding: 12px 0 10px !important;
  border-radius: calc(var(--r-xs) * 0.6) calc(var(--r-xs) * 0.6) calc(var(--r-xs) * 0.4) calc(var(--r-xs) * 0.4);
  background: var(--c) !important;
  color: var(--tc) !important;
  box-shadow: inset 2px 0 0 rgb(255 255 255 / 0.12), inset -3px 0 6px rgb(0 0 0 / 0.25);
  transform: translateY(0);
  transition: transform 0.45s var(--ease-spring), box-shadow 0.3s;

  &::before {
    content: '';
    position: absolute;
    left: 0;
    right: 0;
    top: 9%;
    height: 5px;
    border-top: 1px solid currentColor;
    border-bottom: 1px solid currentColor;
    opacity: 0.35;
  }

  .t { max-height: 70%; margin-top: 14px; overflow: hidden; white-space: nowrap; writing-mode: vertical-rl; font: 700 13px/1 var(--font-serif); letter-spacing: 0.12em; }
  .a { margin-top: auto; writing-mode: vertical-rl; font: 400 9.5px/1 var(--font-sans); letter-spacing: 0.1em; opacity: 0.6; }

  &:hover { transform: translateY(-10px); }
  /* 选中 = 抬升（书脊上浮 + 中性投影 + 细描边），不挂主色光环 */
  &.on { transform: translateY(-18px); box-shadow: inset 2px 0 0 rgb(255 255 255 / 0.12), 0 18px 28px -10px rgb(0 0 0 / 0.55), 0 0 0 1px var(--line-2); }

  &.lean { margin-left: 10px; transform: rotate(-6deg); transform-origin: bottom left; }
  &.lean:hover { transform: rotate(-6deg) translateY(-8px); }
  &.lean.on { transform: rotate(-3deg) translateY(-14px); }

  &.reading::after {
    content: '';
    position: absolute;
    top: -9px;
    right: 6px;
    width: 7px;
    height: 16px;
    background: var(--ak-red);
    clip-path: polygon(0 0, 100% 0, 100% 100%, 50% 75%, 0 100%);
  }
}

/* 进场：书脊依次立起 */
.shelf .spine { animation: none; }
.in .shelf .spine { animation: bs-rise 0.7s var(--ease-spring) both; animation-delay: calc(var(--k) * 55ms + 150ms); }

@keyframes bs-rise { from { opacity: 0; translate: 0 24px; } to { opacity: 1; translate: 0 0; } }

.bs-card {
  display: flex;
  flex-direction: column;
  min-height: 210px;
  padding: 18px;
  border-radius: var(--r-md);
  background: var(--ak-sunken);

  small { font: 500 12.5px var(--font-sans); color: var(--ak-ink); }
  b { display: block; margin: 6px 0 2px; font: 700 22px/1.35 var(--font-serif); }
  span { font-size: 14px; color: var(--text-2); }
  p { margin-top: 12px; padding-top: 12px; border-top: 1px solid var(--ak-line); font: 400 15px/1.75 var(--font-serif); color: var(--text-2); }
}

.bs-anim { display: flex; flex-direction: column; flex: 1; animation: ak-fade-up 0.45s var(--ease-out); }

.pg {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: auto;
  padding-top: 12px;
  font: 500 12.5px var(--ak-mono);
  color: var(--ak-text-3);

  i { position: relative; flex: 1; height: 5px; overflow: hidden; border-radius: var(--r-pill); background: var(--ak-line); }
  i::after { content: ''; position: absolute; inset: 0; width: var(--w); background: var(--ink); }
}

@container (max-width: 500px) {
  .bs { grid-template-columns: 1fr; }
  .shelf { height: 200px; padding-top: 24px; }
  .bs-card { min-height: 0; }
}

@container (max-width: 480px) { .shelf .spine:nth-child(n + 7) { display: none; } }
@container (max-width: 340px) { .shelf .spine:nth-child(n + 6) { display: none; } }
</style>
