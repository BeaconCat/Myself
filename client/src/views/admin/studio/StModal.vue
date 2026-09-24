<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue';

/** Studio 纸面模态：遮罩 + 弹簧入场；内容由插槽提供（确认类统一走 stores/dialog） */
const props = defineProps<{ open: boolean; wide?: boolean; panelClass?: string }>();
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
</script>

<template>
  <Teleport to="body">
    <Transition name="st-modal">
      <div v-if="open" class="studio st-scrim" @mousedown.self="emit('close')">
        <div class="st-modal" :class="[panelClass, { wide }]" role="dialog" aria-modal="true">
          <slot />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
