<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Folder, FolderPlus, Inbox } from 'lucide';
import type { MediaFolder } from '../../../api';
import Icon from '../../../components/ui/Icon.vue';
import StModal from './StModal.vue';
import './i18n';

/** 「移动到…」：选一个目标文件夹（含未归类），也可就地新建 */
const props = defineProps<{ open: boolean; folders: MediaFolder[]; count: number; current?: string }>();
const emit = defineEmits<{ pick: [folder: string]; create: []; close: [] }>();
const { t } = useI18n();

const rows = computed(() => [...props.folders].map((f) => f.path).sort((a, b) => a.localeCompare(b, 'zh')));
</script>

<template>
  <StModal :open="open" panel-class="fpick" @close="emit('close')">
    <h3>{{ t('studio.folder.moveTitle', { n: count }) }}</h3>
    <div class="list">
      <button type="button" :class="{ cur: current === '' }" @click="emit('pick', '')">
        <Icon :icon="Inbox" :size="16" />{{ t('studio.folder.unfiled') }}
      </button>
      <button
        v-for="p in rows"
        :key="p"
        type="button"
        :class="{ cur: current === p }"
        :style="{ '--d': p.split('/').length - 1 }"
        @click="emit('pick', p)"
      >
        <Icon :icon="Folder" :size="16" />{{ p.slice(p.lastIndexOf('/') + 1) }}
      </button>
    </div>
    <div class="ft">
      <button type="button" class="st-btn g sm" @click="emit('create')"><Icon :icon="FolderPlus" :size="15" />{{ t('studio.folder.new') }}</button>
      <span class="sp" />
      <button type="button" class="st-btn g" @click="emit('close')">{{ t('studio.cancel') }}</button>
    </div>
  </StModal>
</template>

<style scoped lang="scss">
:global(.st-modal.fpick) { width: min(420px, calc(100vw - 32px)); }

.list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  max-height: 50vh;
  overflow: auto;
  margin: 10px -6px 0;

  button {
    display: flex;
    align-items: center;
    gap: 8px;
    min-height: 36px;
    padding: 0 10px 0 calc(10px + var(--d, 0) * 16px);
    border: 0;
    border-radius: var(--r-sm);
    font: 500 14px var(--font-sans);
    text-align: left;
    color: var(--st-ink);
    background: none;
    cursor: pointer;

    &:hover { background: var(--hover); }
    &.cur { color: var(--st-ink-4); pointer-events: none; }
  }
}

.ft { display: flex; align-items: center; gap: 8px; margin-top: 16px; }
.sp { flex: 1; }
</style>
