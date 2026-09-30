<script setup lang="ts">
import { computed } from 'vue';
import { Archive, FileText, Film, Music } from 'lucide';
import type { MediaItem } from '../../../api';
import Icon from '../../../components/ui/Icon.vue';
import { extOf, mediaKind } from '../../../utils/mediaKind';

/**
 * 素材缩略块：图片用缩略图；视频取首帧（preload=metadata，不自动播放）；音频 / 压缩包 / 文件显示类型图标 + 扩展名。
 * 素材库页与素材库模态框共用。
 */
const props = defineProps<{ item: MediaItem; stamp?: string | number }>();
const kind = computed(() => mediaKind(props.item));
const ext = computed(() => (props.item.ext || extOf(props.item.name)).toUpperCase());
const ICONS = { video: Film, audio: Music, archive: Archive, file: FileText } as const;
</script>

<template>
  <span class="mt" :data-kind="kind">
    <img v-if="kind === 'image'" :src="stamp ? `${item.thumb}?v=${stamp}` : item.thumb" alt="" loading="lazy" draggable="false" />
    <video v-else-if="kind === 'video'" :src="`${item.url}#t=0.1`" preload="metadata" muted playsinline />
    <span v-else class="ph"><Icon :icon="ICONS[kind]" :size="26" /><em>{{ ext }}</em></span>
    <span v-if="kind === 'video'" class="badge"><Icon :icon="Film" :size="13" />{{ ext }}</span>
  </span>
</template>

<style scoped lang="scss">
.mt {
  position: relative;
  display: block;
  width: 100%;
  height: 100%;
  background: var(--well-2);

  img, video { display: block; width: 100%; height: 100%; object-fit: cover; }
}

.ph {
  position: absolute;
  inset: 0;
  display: grid;
  place-content: center;
  justify-items: center;
  gap: 6px;
  color: var(--st-ink-3);

  em { font: 600 11.5px/1 var(--font-mono); font-style: normal; letter-spacing: 0.06em; color: var(--st-ink-2); }
}

.mt[data-kind='audio'] .ph { color: var(--ink); background: color-mix(in oklab, var(--ink) 8%, var(--well-2)); }
.mt[data-kind='archive'] .ph { background: color-mix(in oklab, var(--st-ink-4) 12%, var(--well-2)); }

.badge {
  position: absolute;
  left: 8px;
  bottom: 8px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 7px;
  border-radius: var(--r-xs);
  font: 600 11px/18px var(--font-mono);
  color: #fff;
  background: rgba(0, 0, 0, 0.5);
}
</style>
