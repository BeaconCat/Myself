<script setup lang="ts">
import { ref, useId } from 'vue';
import { ChevronDown } from 'lucide';
import Icon from '../ui/Icon.vue';
import { useI18n } from 'vue-i18n';
import RichEditor from './RichEditor.vue';
const props = defineProps<{ modelValue: string; placeholder?: string }>();
const emit = defineEmits<{ 'update:modelValue': [value: string] }>();
const { t } = useI18n();
const source = ref(false);
const showToolbar = ref(false);
const toolbarId = `markdown-tools-${useId()}`;
const rich = ref<InstanceType<typeof RichEditor> | null>(null);
const textarea = ref<HTMLTextAreaElement | null>(null);
defineExpose({ focus: () => source.value ? textarea.value?.focus() : rich.value?.focus() });
</script>
<template>
  <div class="markdown-editor studio">
    <div class="mode" role="group" :aria-label="t('markdownEditor.mode')">
      <button type="button" :aria-pressed="!source" @click="source = false">{{ t('markdownEditor.rich') }}</button>
      <button type="button" :aria-pressed="source" @click="source = true">{{ t('markdownEditor.source') }}</button>
      <button v-if="!source" type="button" class="toolbar-toggle" :aria-expanded="showToolbar" :aria-controls="toolbarId"
        :aria-label="t(showToolbar ? 'markdownEditor.collapseToolbar' : 'markdownEditor.expandToolbar')"
        @click="showToolbar = !showToolbar">
        {{ t(showToolbar ? 'markdownEditor.collapseToolbar' : 'markdownEditor.expandToolbar') }}
        <Icon :icon="ChevronDown" :size="14" />
      </button>
    </div>
    <textarea v-if="source" ref="textarea" class="source" :value="props.modelValue" :placeholder="placeholder"
      :aria-label="t('markdownEditor.source')" @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)" />
    <RichEditor v-else ref="rich" :toolbar-visible="showToolbar" :toolbar-id="toolbarId" :model-value="modelValue" :placeholder="placeholder" @update:model-value="emit('update:modelValue', $event)" />
  </div>
</template>
<style scoped lang="scss">
.markdown-editor { min-width: 0; width: 100%; }
.mode { display: flex; gap: 4px; margin-bottom: 8px; }
.mode button { padding: 6px 10px; border: 0; border-radius: var(--r-xs); background: none; color: var(--text-2); font-size: 12px; }
.mode button[aria-pressed='true'] { background: var(--well); color: var(--text); }
.mode .toolbar-toggle { margin-left: auto; display: inline-flex; align-items: center; gap: 4px; white-space: nowrap; }
.toolbar-toggle :deep(svg) { transition: transform .28s var(--ease-out); }
.toolbar-toggle[aria-expanded="true"] :deep(svg) { transform: rotate(180deg); }
@media (prefers-reduced-motion: reduce) { .toolbar-toggle :deep(svg) { transition: none; } }
.source { width: 100%; min-height: 240px; resize: vertical; padding: 14px; border: 1px solid var(--line-2); border-radius: var(--r-sm); background: var(--well); color: var(--text); font: 14px/1.9 var(--font-mono); }
</style>
