<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import './i18n';
import { Archive, FileText, Film, Image as ImageIcon, Music } from 'lucide';
import type { MediaItem } from '../../../api';
import Icon from '../../../components/ui/Icon.vue';
import { extOf, mediaKind } from '../../../utils/mediaKind';

/**
 * 素材缩略块：优先显示图片/内嵌封面；无封面的视频主动读取首帧，其余使用类型占位。
 * 素材库页与素材库模态框共用。
 */
const props = defineProps<{ item: MediaItem; stamp?: string | number }>();
const { t } = useI18n();
const kind = computed(() => mediaKind(props.item));
const ext = computed(() => (props.item.ext || extOf(props.item.name)).toUpperCase());
const ICONS = { image: ImageIcon, video: Film, audio: Music, archive: Archive, file: FileText } as const;
const thumbFailed = ref(false), frameReady = ref(false);
const cover = computed(() => props.item.thumb && !thumbFailed.value && ['image', 'audio', 'video'].includes(kind.value));
const source = computed(() => props.stamp ? `${props.item.thumb}?v=${props.stamp}` : props.item.thumb);
watch(() => [props.item.url, props.item.thumb, props.stamp], () => { thumbFailed.value = false; frameReady.value = false; });
function seekFrame(event: Event): void {
  const video = event.target as HTMLVideoElement;
  if (Number.isFinite(video.duration) && video.duration > 0) video.currentTime = Math.min(.1, video.duration / 2);
}
function showFrame(event: Event): void { frameReady.value = (event.target as HTMLVideoElement).videoWidth > 0; }
</script>

<template>
  <span class="mt" :data-kind="kind">
    <img v-if="cover" :src="source" alt="" loading="lazy" draggable="false" @error="thumbFailed = true" />
    <template v-else-if="kind === 'video'">
      <span v-if="!frameReady" class="ph"><Icon :icon="Film" :size="26" /><em>{{ ext }}</em></span>
      <video :src="item.url" preload="metadata" muted playsinline :class="{ ready: frameReady }" @loadedmetadata="seekFrame" @loadeddata="showFrame" @seeked="showFrame" />
    </template>
    <span v-else class="ph"><Icon :icon="ICONS[kind]" :size="26" /><em>{{ ext }}</em></span>
    <span v-if="kind === 'video' || kind === 'audio' || kind === 'archive'" class="badge" :title="t(`studio.library.kind.${kind}`)"><Icon :icon="ICONS[kind]" :size="13" />{{ ext }}</span>
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
  video { position: relative; opacity: 0; } video.ready { opacity: 1; }
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
