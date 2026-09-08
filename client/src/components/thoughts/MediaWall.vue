<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { thumbOf, type Note } from '../../api';

/** 媒体墙（X 风）：铺平所有随想配图为三列宫格 */
const props = defineProps<{ notes: Note[]; loading: boolean }>();

const emit = defineEmits<{
  /** 打开查看器：图组 + 起始索引 */
  open: [images: string[], index: number];
}>();

const { t } = useI18n();

interface MediaCell {
  src: string;
  note: Note;
  index: number;
}

const mediaCells = computed<MediaCell[]>(() =>
  props.notes.flatMap((note) =>
    note.images.map((src, index) => ({ src, note, index })),
  ),
);
</script>

<template>
  <div class="media-wall" :class="{ loading }">
    <button
      v-for="(cell, i) in mediaCells"
      :key="`${cell.note.id}-${cell.index}`"
      class="media-cell"
      :style="{ '--i': i % 30 }"
      @click="emit('open', cell.note.images, cell.index)"
    >
      <img :src="thumbOf(cell.src)" loading="lazy" alt="" />
    </button>
    <p v-if="!loading && !mediaCells.length" class="empty">{{ t('thoughts.mediaEmpty') }}</p>
  </div>
</template>

<style scoped lang="scss">
.media-wall {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 4px;
  border-radius: var(--radius);
  overflow: hidden;
  transition: opacity var(--dur-fast);

  &.loading { opacity: 0.55; }
}

.media-cell {
  border: none;
  padding: 0;
  background: var(--surface-2);
  aspect-ratio: 1;
  overflow: hidden;
  cursor: zoom-in;
  animation: tweet-in 0.45s var(--ease-out) both;
  animation-delay: calc(var(--i, 0) * 0.03s);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    transition: transform var(--dur) var(--ease-out), filter var(--dur-fast);
  }

  &:hover img { transform: scale(1.06); filter: brightness(1.06); }
}

@keyframes tweet-in {
  from { opacity: 0; transform: translateY(22px); }
  to { opacity: 1; transform: none; }
}

.empty {
  grid-column: 1 / -1;
  color: var(--text-2);
  text-align: center;
  padding: 48px 0;
}
</style>
