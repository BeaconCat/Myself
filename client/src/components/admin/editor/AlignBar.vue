<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { AlignCenter, AlignLeft, AlignRight, RotateCcw, WrapText } from 'lucide';
import Icon from '../../ui/Icon.vue';
import type { MediaAlign } from '../../../utils/mediaSize';

/** 媒体对齐 + 恢复原始尺寸的小工具条（图片 / 视频 / 音频节点共用）；inline 为行内（仅图片） */
defineProps<{ align: MediaAlign; inline?: boolean; sized: boolean }>();
const emit = defineEmits<{ align: [a: MediaAlign]; reset: [] }>();
const { t } = useI18n();
</script>

<template>
  <span class="abar" contenteditable="false" @mousedown.prevent>
    <button v-if="inline" type="button" :class="{ on: !align }" :title="t('studio.embed.alignInline')" @click="emit('align', '')"><Icon :icon="WrapText" :size="15" /></button>
    <button type="button" :class="{ on: align === 'left' || (!inline && !align) }" :title="t('studio.embed.alignLeft')" @click="emit('align', 'left')"><Icon :icon="AlignLeft" :size="15" /></button>
    <button type="button" :class="{ on: align === 'center' }" :title="t('studio.embed.alignCenter')" @click="emit('align', 'center')"><Icon :icon="AlignCenter" :size="15" /></button>
    <button type="button" :class="{ on: align === 'right' }" :title="t('studio.embed.alignRight')" @click="emit('align', 'right')"><Icon :icon="AlignRight" :size="15" /></button>
    <span class="sep" />
    <button type="button" :disabled="!sized" :title="t('studio.embed.resetSize')" @click="emit('reset')"><Icon :icon="RotateCcw" :size="15" /></button>
  </span>
</template>
