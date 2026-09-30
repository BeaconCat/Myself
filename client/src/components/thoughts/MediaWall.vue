<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Note } from '../../api';
import CoverArt from '../common/CoverArt.vue';
import { ymdOf } from '../post/content';

/** 媒体墙：铺平随想配图，按月分组的四列方格，角标为日期；点开进入查看器（该条随想的图组） */
const props = defineProps<{ notes: Note[] }>();
const emit = defineEmits<{ open: [images: string[], index: number, rect: DOMRect] }>();
const { t } = useI18n();

interface Cell { key: string; src: string; note: Note; index: number; label: string }
interface Group { key: string; label: string; cells: Cell[] }

const groups = computed<Group[]>(() => {
  const out: Group[] = [];
  for (const note of props.notes) {
    if (!note.images.length) continue;
    const ym = note.createdAt.slice(0, 7);
    const { y, m, d } = ymdOf(note.createdAt);
    let g = out[out.length - 1];
    if (!g || g.key !== ym) {
      g = { key: ym, label: t('content.thoughts.monthTitle', { y, m }), cells: [] };
      out.push(g);
    }
    note.images.forEach((src, index) => {
      g.cells.push({ key: `${note.id}-${index}`, src, note, index, label: `${m}月${d}日` });
    });
  }
  return out;
});

function openAt(e: MouseEvent, c: Cell): void {
  emit('open', c.note.images, c.index, (e.currentTarget as HTMLElement).getBoundingClientRect());
}
</script>

<template>
  <div class="mwall">
    <section v-for="g in groups" :key="g.key">
      <div class="mg-h"><b>{{ g.label }}</b><span>{{ t('content.thoughts.mediaCount', { n: g.cells.length }) }}</span></div>
      <div class="mgrid">
        <button v-for="c in g.cells" :key="c.key" type="button" class="cell" :aria-label="t('a11y.viewImageOf', { label: c.label })" @click="openAt($event, c)">
          <CoverArt :src="c.src" :seed="c.key" pool="all" thumb />
          <span class="n">{{ c.label }}</span>
        </button>
      </div>
    </section>
    <p v-if="!groups.length" class="nores">{{ t('thoughts.mediaEmpty') }}</p>
  </div>
</template>

<style scoped lang="scss">
.mg-h {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding: 20px 0 10px;
  font-size: 13px;
  letter-spacing: 0.06em;
  color: var(--text-3);

  b { font-weight: 500; color: var(--text-2); }
}

.mgrid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 4px;
  overflow: hidden;
  isolation: isolate;
  border-radius: var(--r-lg);
}

.cell {
  position: relative;
  aspect-ratio: 1;
  overflow: hidden;
  padding: 0;
  border: 0;
  background: #040914;
  cursor: zoom-in;

  &:hover :deep(.cv) { transform: scale(1.05); }
  &:focus-visible { outline: none; box-shadow: inset 0 0 0 2px var(--ink); }

  /* 底部轻压暗，保证角标在亮图上可读 */
  &::after {
    content: '';
    position: absolute;
    inset: 55% 0 0;
    z-index: 3;
    background: linear-gradient(transparent, rgb(0 0 0 / 0.32));
    pointer-events: none;
  }

  /* 日期角标：白字 72% */
  .n {
    position: absolute;
    right: 8px;
    bottom: 6px;
    z-index: 4;
    font-family: var(--font-mono);
    font-size: 12px;
    color: rgb(255 255 255 / 0.72);
  }
}

.nores {
  padding: 72px 0;
  text-align: center;
  color: var(--text-3);
}

@media (max-width: 1100px) {
  .mgrid { grid-template-columns: repeat(4, 1fr); }
}

@media (max-width: 640px) {
  .mgrid { grid-template-columns: repeat(3, 1fr); }
}
</style>
