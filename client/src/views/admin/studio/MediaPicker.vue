<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type MediaItem } from '../../../api';
import { mediaKind } from '../../../utils/mediaKind';
import './i18n';
import SIcon from './SIcon.vue';
import StModal from './StModal.vue';
import { toast } from './toast';

/**
 * 从素材库选一张图（单选）：点缩略图即选中并关闭；也可以在这里直接上传（走查重，重复的直接复用已有那张）。
 * 身份页的站点 logo / 头像 / 形象图 / 名片头图共用。
 */
const props = defineProps<{ open: boolean; title: string; current?: string }>();
const emit = defineEmits<{ pick: [url: string]; close: [] }>();
const { t } = useI18n();

const items = ref<MediaItem[] | null>(null);
const uploading = ref(false);
const fileInput = ref<HTMLInputElement | null>(null);

watch(
  () => props.open,
  async (open) => {
    if (!open) return;
    try {
      items.value = (await adminApi.media()).filter((item) => mediaKind(item) === 'image');
    } catch {
      items.value = [];
      toast(t('studio.loadFailed'), { icon: 'x' });
    }
  },
  { immediate: true },
);

function pick(url: string): void {
  emit('pick', url);
  emit('close');
}

async function onFile(e: Event): Promise<void> {
  const input = e.target as HTMLInputElement;
  const f = input.files?.[0];
  input.value = '';
  if (!f || !f.type.startsWith('image/') || uploading.value) return;
  uploading.value = true;
  try {
    const [up] = await adminApi.uploadMedia([f]);
    if (up && mediaKind(up) === 'image') {
      if (up.duplicate) toast(t('studio.picker.reused'), { icon: 'copy' });
      pick(up.url);
    }
  } catch {
    toast(t('studio.identity.uploadFailed'), { icon: 'x' });
  } finally {
    uploading.value = false;
  }
}
</script>

<template>
  <StModal :open="open" wide panel-class="mpick" @close="emit('close')">
    <h3>{{ title }}</h3>
    <p>{{ t('studio.picker.desc') }}</p>
    <div class="grid">
      <button type="button" class="up" :disabled="uploading" @click="fileInput?.click()">
        <SIcon name="upload" :size="20" />
        <span>{{ uploading ? t('studio.media.uploading') : t('studio.picker.upload') }}</span>
      </button>
      <button
        v-for="m in items ?? []"
        :key="m.name"
        type="button"
        :class="{ on: current === m.url }"
        :title="m.name"
        @click="pick(m.url)"
      >
        <img :src="m.thumb" alt="" loading="lazy" />
        <span v-if="current === m.url" class="tick"><SIcon name="check" :size="14" /></span>
      </button>
    </div>
    <p v-if="items && !items.length" class="empty">{{ t('studio.picker.empty') }}</p>
    <input ref="fileInput" type="file" accept="image/*" hidden @change="onFile" />
    <div class="ft">
      <button type="button" class="st-btn g" @click="emit('close')">{{ t('studio.cancel') }}</button>
    </div>
  </StModal>
</template>

<style scoped lang="scss">
:global(.st-modal.mpick) { width: min(760px, calc(100vw - 32px)); }

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(112px, 1fr));
  gap: 10px;
  max-height: 52vh;
  overflow: auto;
  padding: 4px;

  button {
    position: relative;
    aspect-ratio: 1;
    border-radius: var(--r-md);
    overflow: hidden;
    background: var(--well-2);
    box-shadow: 0 0 0 1px var(--line);
    transition: transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast);

    img { width: 100%; height: 100%; object-fit: cover; display: block; }
    &:hover { transform: translateY(-2px); }
    &.on { box-shadow: 0 0 0 2px var(--paper), 0 0 0 3.5px var(--st-ink); }
  }

  .up {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    color: var(--st-ink-3);
    font-size: 12.5px;
    background: var(--well);
    box-shadow: inset 0 0 0 1.5px var(--line-2);

    &:hover { color: var(--st-ink); }
    &:disabled { opacity: 0.6; cursor: progress; }
  }

  .tick {
    position: absolute;
    top: 6px;
    right: 6px;
    display: grid;
    place-items: center;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: var(--st-ink);
    color: var(--paper);
  }
}

.empty { margin: 12px 0 0; font-size: 13.5px; color: var(--st-ink-3); }
</style>
