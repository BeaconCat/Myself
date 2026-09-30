<script setup lang="ts">
import { computed } from 'vue';
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3';
import type { MediaAlign } from '../../../utils/mediaSize';
import ResizeBox from './ResizeBox.vue';
import AlignBar from './AlignBar.vue';

/**
 * 正文图片节点视图：选中后四周出拖柄自由改尺寸（Shift 等比），上方浮出对齐工具条。
 * 行内（缺省）/ 左 / 中 / 右；对齐后图片独占一行。
 */
const props = defineProps(nodeViewProps);
const a = computed(() => props.node.attrs as { src: string; alt: string | null; width: number | null; height: number | null; align: MediaAlign });

/** 改属性后重新选中本节点：保持拖柄与工具条，便于连续调整 */
function update(attrs: Record<string, unknown>): void {
  props.updateAttributes(attrs);
  const pos = props.getPos();
  if (typeof pos === 'number') props.editor.commands.setNodeSelection(pos);
}

function onSize(s: { w: number | null; h: number | null }): void {
  update({ width: s.w, height: s.h });
}
</script>

<template>
  <NodeViewWrapper as="span" class="rimg" :class="[a.align && `al-${a.align}`, { sel: selected }]">
    <ResizeBox :w="a.width" :h="a.height" :selected="selected" :centered="a.align === 'center'" @change="onSize">
      <img :src="a.src" :alt="a.alt ?? ''" draggable="false" data-drag-handle />
    </ResizeBox>
    <AlignBar
      v-if="selected"
      :align="a.align"
      inline
      :sized="!!a.width"
      @align="(v) => update({ align: v })"
      @reset="onSize({ w: null, h: null })"
    />
  </NodeViewWrapper>
</template>
