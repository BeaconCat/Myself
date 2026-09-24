<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref } from 'vue';
import SIcon from './SIcon.vue';
import type { MenuItem } from './types';

/** 「更多」弹出菜单：锚定按钮右下，点外部关闭 */
defineProps<{ items: MenuItem[]; title?: string }>();

const open = ref(false);
const pos = ref({ top: 0, left: 0 });
const btn = ref<HTMLElement | null>(null);
const menu = ref<HTMLElement | null>(null);

function outside(e: MouseEvent): void {
  if (!menu.value?.contains(e.target as Node) && !btn.value?.contains(e.target as Node)) close();
}

function close(): void {
  open.value = false;
  document.removeEventListener('mousedown', outside);
  window.removeEventListener('scroll', close, true);
}

async function toggle(): Promise<void> {
  if (open.value) return close();
  const r = btn.value!.getBoundingClientRect();
  pos.value = { top: r.bottom + 6, left: r.right };
  open.value = true;
  await nextTick();
  const w = menu.value?.offsetWidth ?? 172;
  const h = menu.value?.offsetHeight ?? 0;
  const top = r.bottom + 6 + h > window.innerHeight - 8 ? r.top - 6 - h : r.bottom + 6;
  pos.value = { top, left: Math.max(12, r.right - w) };
  document.addEventListener('mousedown', outside);
  window.addEventListener('scroll', close, true);
}

function pick(item: MenuItem): void {
  close();
  item.run();
}

onBeforeUnmount(close);
defineExpose({ close });
</script>

<template>
  <button ref="btn" type="button" class="st-ibtn" :class="{ on: open }" :title="title" @click.stop="toggle">
    <SIcon name="more" />
  </button>
  <Teleport to="body">
    <div v-if="open" ref="menu" class="studio st-menu" :style="{ top: `${pos.top}px`, left: `${pos.left}px` }" @click.stop>
      <template v-for="(item, i) in items" :key="i">
        <hr v-if="item.divider" />
        <button type="button" :class="{ d: item.danger }" @click="pick(item)">
          <SIcon :name="item.icon" :size="16" />{{ item.label }}
        </button>
      </template>
    </div>
  </Teleport>
</template>
