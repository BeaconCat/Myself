<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3';
import { useI18n } from 'vue-i18n';
import { Archive, Eye, EyeOff, Replace, Trash2 } from 'lucide';
import Icon from '../../ui/Icon.vue';
import { renderMediaHtml, type MediaData, type MediaKind } from '../../../utils/embeds';
import { extOf } from '../../../utils/mediaKind';
import type { MediaAlign } from '../../../utils/mediaSize';
import ResizeBox from './ResizeBox.vue';
import AlignBar from './AlignBar.vue';
import type { EmbedHooks } from './embeds';

/**
 * 媒体节点视图：与前台同一段 HTML（renderMediaHtml）所见即所得；
 * 悬停 / 选中时右上角浮出工具条：改标题、换文件、压缩包预览开关、删除。
 */
const props = defineProps(nodeViewProps);
const { t } = useI18n();

const data = computed(() => props.node.attrs as MediaData & { width: number | null; height: number | null; align: MediaAlign });
/** 视频 / 音频可拖动改尺寸（音频只改宽）；文件卡片整行，不改尺寸 */
const sizable = computed(() => data.value.kind === 'video' || data.value.kind === 'audio');
const html = computed(() => renderMediaHtml({ kind: data.value.kind, src: data.value.src, title: data.value.title }, {
  download: t('content.embed.download'), preview: t('content.embed.preview'), open: t('content.embed.open'),
}));
const zip = computed(() => extOf(data.value.src) === 'zip');
const hooks = computed(() => props.extension.options as EmbedHooks);

const title = ref(data.value.title);
watch(() => data.value.title, (v) => { title.value = v; });

function saveTitle(): void {
  if (title.value.trim() !== data.value.title) props.updateAttributes({ title: title.value.trim() });
}

function replace(): void {
  hooks.value.replaceMedia?.(data.value.kind, (next) => props.updateAttributes(next));
}

/** 改属性后重新选中本节点：保持拖柄与工具条 */
function update(attrs: Record<string, unknown>): void {
  props.updateAttributes(attrs);
  const pos = props.getPos();
  if (typeof pos === 'number') props.editor.commands.setNodeSelection(pos);
}

function onSize(s: { w: number | null; h: number | null }): void {
  update({ width: s.w, height: data.value.kind === 'audio' ? null : s.h });
}

function togglePreview(): void {
  const kind: MediaKind = data.value.kind === 'archive' ? 'file' : 'archive';
  props.updateAttributes({ kind });
}
</script>

<template>
  <NodeViewWrapper class="emb" :class="[{ sel: selected }, sizable && data.align && `al-${data.align}`]" data-drag-handle>
    <ResizeBox
      v-if="sizable"
      class="emb-rz"
      :w="data.width"
      :h="data.height"
      :selected="selected"
      :axis="data.kind === 'audio' ? 'x' : 'both'"
      :centered="data.align === 'center'"
      @change="onSize"
    >
      <!-- eslint-disable-next-line vue/no-v-html -->
      <div class="emb-body" contenteditable="false" v-html="html" />
    </ResizeBox>
    <!-- eslint-disable-next-line vue/no-v-html -->
    <div v-else class="emb-body" contenteditable="false" v-html="html" />
    <AlignBar
      v-if="sizable && selected"
      class="emb-align"
      :align="data.align"
      :sized="!!data.width"
      @align="(v) => update({ align: v })"
      @reset="onSize({ w: null, h: null })"
    />
    <div class="emb-bar" contenteditable="false">
      <input
        v-model="title"
        class="emb-title"
        :placeholder="t('studio.embed.titlePh')"
        @keydown.enter.prevent="($event.target as HTMLInputElement).blur()"
        @blur="saveTitle"
      />
      <button v-if="hooks.replaceMedia" type="button" :title="t('studio.embed.replace')" @click="replace"><Icon :icon="Replace" :size="15" /></button>
      <button
        v-if="zip"
        type="button"
        :class="{ on: data.kind === 'archive' }"
        :title="data.kind === 'archive' ? t('studio.embed.previewOff') : t('studio.embed.previewOn')"
        @click="togglePreview"
      >
        <Icon :icon="data.kind === 'archive' ? Eye : EyeOff" :size="15" /><Icon :icon="Archive" :size="13" />
      </button>
      <button type="button" class="danger" :title="t('studio.delete')" @click="deleteNode()"><Icon :icon="Trash2" :size="15" /></button>
    </div>
  </NodeViewWrapper>
</template>
