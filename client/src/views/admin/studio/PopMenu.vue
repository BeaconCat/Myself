<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import SIcon from './SIcon.vue';
import type { MenuItem } from './types';
import './i18n';

/** 「更多」弹出菜单：锚定按钮右下，点外部关闭；label 为触发按钮的可访问名称（默认「更多操作」） */
defineProps<{ items: MenuItem[]; title?: string; label?: string }>();
const { t } = useI18n();

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
  <button
    ref="btn"
    type="button"
    class="st-ibtn"
    :class="{ on: open }"
    :title="title"
    :aria-label="label || title || t('studio.a11y.more')"
    aria-haspopup="menu"
    :aria-expanded="open"
    @click.stop="toggle"
  >
    <SIcon name="more" />
  </button>
  <Teleport to="body">
    <div v-if="open" ref="menu" class="studio st-menu" role="menu" :style="{ top: `${pos.top}px`, left: `${pos.left}px` }" @click.stop>
      <template v-for="(item, i) in items" :key="i">
        <hr v-if="item.divider" />
        <button type="button" role="menuitem" :class="{ d: item.danger }" @click="pick(item)">
          <SIcon :name="item.icon" :size="16" />{{ item.label }}
        </button>
      </template>
    </div>
  </Teleport>
</template>
