<script setup lang="ts">
import { nextTick, onBeforeUnmount, onUpdated, ref, useId, watch } from 'vue';

/**
 * Studio 纸面模态：遮罩 + 弹簧入场；内容由插槽提供（确认类统一走 stores/dialog）。
 * 对话框名称：显式 label 优先，否则指向插槽里的第一个标题（内容切换后重新指向）。
 */
const props = defineProps<{ open: boolean; wide?: boolean; panelClass?: string; label?: string }>();
const emit = defineEmits<{ close: [] }>();

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape') emit('close');
}

watch(
  () => props.open,
  (v) => {
    if (v) document.addEventListener('keydown', onKey);
    else document.removeEventListener('keydown', onKey);
  },
  { immediate: true },
);
onBeforeUnmount(() => document.removeEventListener('keydown', onKey));

const panel = ref<HTMLElement | null>(null);
const labelledby = ref<string | undefined>();
const uid = useId();
function nameFromHeading(): void {
  if (props.label) return;
  const h = panel.value?.querySelector<HTMLElement>('h1, h2, h3, h4');
  if (h && !h.id) h.id = `${uid}-title`;
  labelledby.value = h?.id || undefined;
}
watch(() => props.open, (v) => { if (v) void nextTick(nameFromHeading); });
onUpdated(nameFromHeading);
</script>

<template>
  <Teleport to="body">
    <Transition name="st-modal">
      <div v-if="open" class="studio st-scrim" @mousedown.self="emit('close')">
        <div
          ref="panel"
          class="st-modal"
          :class="[panelClass, { wide }]"
          role="dialog"
          aria-modal="true"
          :aria-label="label || undefined"
          :aria-labelledby="label ? undefined : labelledby"
        >
          <slot />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
