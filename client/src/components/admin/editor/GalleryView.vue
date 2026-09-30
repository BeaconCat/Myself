<script setup lang="ts">
import { computed } from 'vue';
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3';
import { useI18n } from 'vue-i18n';
import { LayoutGrid, Trash2 } from 'lucide';
import Icon from '../../ui/Icon.vue';
import { renderGalleryHtml, type GalleryData } from '../../../utils/embeds';
import type { EmbedHooks } from './embeds';

/** 拼图节点视图：与前台同一段 HTML（renderGalleryHtml）；点击或「编辑拼图」打开拼图编辑器 */
const props = defineProps(nodeViewProps);
const { t } = useI18n();

const data = computed(() => props.node.attrs.data as GalleryData);
const html = computed(() => renderGalleryHtml(data.value));
const hooks = computed(() => props.extension.options as EmbedHooks);

function edit(): void {
  hooks.value.editGallery?.(data.value, (next) => props.updateAttributes({ data: next }));
}
</script>

<template>
  <NodeViewWrapper class="emb gal" :class="{ sel: selected }" data-drag-handle>
    <!-- eslint-disable-next-line vue/no-v-html -->
    <div v-if="data.images.length" class="emb-body" contenteditable="false" @dblclick="edit" v-html="html" />
    <button v-else type="button" class="emb-empty" contenteditable="false" @click="edit">
      <Icon :icon="LayoutGrid" :size="22" />{{ t('studio.embed.galleryEmpty') }}
    </button>
    <div class="emb-bar" contenteditable="false">
      <span class="emb-info">{{ t('studio.embed.galleryInfo', { n: data.images.length, layout: t(`studio.gallery.layout.${data.layout}`) }) }}</span>
      <button type="button" class="txt" @click="edit"><Icon :icon="LayoutGrid" :size="15" />{{ t('studio.embed.editGallery') }}</button>
      <button type="button" class="danger" :title="t('studio.delete')" @click="deleteNode()"><Icon :icon="Trash2" :size="15" /></button>
    </div>
  </NodeViewWrapper>
</template>
