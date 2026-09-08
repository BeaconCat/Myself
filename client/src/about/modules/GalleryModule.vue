<script setup lang="ts">
import { ref } from 'vue';
import type { AboutModule } from '../../stores/config';
import ImageViewer from '../../components/media/ImageViewer.vue';

/** 照片墙：宫格 + 全屏查看器 */
defineProps<{ mod: AboutModule }>();

const viewerImages = ref<string[]>([]);
const viewerIndex = ref(0);
const viewerOpen = ref(false);

function openGallery(images: string[], index: number): void {
  viewerImages.value = images;
  viewerIndex.value = index;
  viewerOpen.value = true;
}
</script>

<template>
  <h2 class="block-title">照片墙</h2>
  <div class="gallery">
    <button
      v-for="(src, i) in mod.data.images"
      :key="src"
      class="g-cell"
      @click="openGallery(mod.data.images, Number(i))"
    >
      <img :src="src" loading="lazy" alt="" />
    </button>
  </div>
  <p v-if="!mod.data.images?.length" class="empty">在后台上传照片后展示。</p>

  <ImageViewer
    v-if="viewerOpen"
    :images="viewerImages"
    :start-index="viewerIndex"
    @close="viewerOpen = false"
  />
</template>

<style scoped lang="scss">
@use './shared';

.gallery {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 6px;
}

.g-cell {
  border: none;
  padding: 0;
  background: var(--surface-2);
  aspect-ratio: 1;
  overflow: hidden;
  cursor: zoom-in;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    transition: transform var(--dur) var(--ease-out);
  }

  &:hover img { transform: scale(1.06); }
}
</style>
