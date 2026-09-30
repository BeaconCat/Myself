<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ChevronRight, Folder, FolderOpen, FolderPlus, Inbox, Layers } from 'lucide';
import type { MediaFolder } from '../../../api';
import Icon from '../../../components/ui/Icon.vue';
import PopMenu from './PopMenu.vue';
import { ALL, DRAG_TYPE } from './useFolders';
import type { MenuItem } from './types';
import './i18n';

/**
 * 素材文件夹树：全部 / 未归类 / 各级文件夹（可折叠）。editable 时带新建、重命名、删除菜单，
 * 并接收拖进来的素材（dataTransfer: DRAG_TYPE）以移动到该文件夹。
 */
const props = defineProps<{
  folders: MediaFolder[];
  total: number;
  unfiled: number;
  editable?: boolean;
  compact?: boolean;
}>();
const current = defineModel<string>({ required: true });
const emit = defineEmits<{ create: [parent: string]; rename: [path: string]; remove: [path: string]; drop: [names: string[], folder: string] }>();
const { t } = useI18n();

const collapsed = ref<Set<string>>(new Set());
const dropOn = ref<string | null>(null);

interface Row { path: string; name: string; depth: number; count: number; hasKids: boolean }

/** 路径列表 → 按层级排序的行；折叠的文件夹隐藏其子项 */
const rows = computed<Row[]>(() => {
  const paths = [...props.folders].map((f) => f.path).sort((a, b) => a.localeCompare(b, 'zh'));
  const count = new Map(props.folders.map((f) => [f.path, f.count]));
  return paths
    .filter((p) => ![...collapsed.value].some((c) => p.startsWith(`${c}/`)))
    .map((p) => ({
      path: p,
      name: p.slice(p.lastIndexOf('/') + 1),
      depth: p.split('/').length - 1,
      count: count.get(p) ?? 0,
      hasKids: paths.some((q) => q.startsWith(`${p}/`)),
    }));
});

function toggle(p: string): void {
  const next = new Set(collapsed.value);
  if (next.has(p)) next.delete(p);
  else next.add(p);
  collapsed.value = next;
}

function menu(p: string): MenuItem[] {
  return [
    { icon: 'plus', label: t('studio.folder.newSub'), run: () => emit('create', p) },
    { icon: 'pen', label: t('studio.folder.rename'), run: () => emit('rename', p) },
    { icon: 'trash', label: t('studio.delete'), danger: true, divider: true, run: () => emit('remove', p) },
  ];
}

/* ---------- 拖放：素材拖到文件夹上即移动 ---------- */
function onOver(e: DragEvent, target: string): void {
  if (!props.editable || !e.dataTransfer?.types.includes(DRAG_TYPE)) return;
  e.preventDefault();
  e.dataTransfer.dropEffect = 'move';
  dropOn.value = target;
}
function onDrop(e: DragEvent, target: string): void {
  dropOn.value = null;
  const raw = e.dataTransfer?.getData(DRAG_TYPE);
  if (!raw) return;
  e.preventDefault();
  try {
    const names = JSON.parse(raw) as string[];
    if (Array.isArray(names) && names.length) emit('drop', names, target);
  } catch { /* 忽略 */ }
}
</script>

<template>
  <nav class="ftree" :class="{ compact }" :aria-label="t('studio.folder.title')">
    <div class="fh">
      <span>{{ t('studio.folder.title') }}</span>
      <button v-if="editable" type="button" class="st-ibtn sm" :title="t('studio.folder.new')" @click="emit('create', '')">
        <Icon :icon="FolderPlus" :size="16" />
      </button>
    </div>
    <button type="button" class="fr" :class="{ on: current === ALL }" @click="current = ALL">
      <Icon :icon="Layers" :size="16" /><span class="nm">{{ t('studio.folder.all') }}</span><em>{{ total }}</em>
    </button>
    <button
      type="button"
      class="fr"
      :class="{ on: current === '', drop: dropOn === '' }"
      @click="current = ''"
      @dragover="onOver($event, '')"
      @dragleave="dropOn = null"
      @drop="onDrop($event, '')"
    >
      <Icon :icon="Inbox" :size="16" /><span class="nm">{{ t('studio.folder.unfiled') }}</span><em>{{ unfiled }}</em>
    </button>
    <TransitionGroup name="fr" tag="div" class="fl">
      <div
        v-for="r in rows"
        :key="r.path"
        class="fr"
        :class="{ on: current === r.path, drop: dropOn === r.path }"
        :style="{ '--d': r.depth }"
        role="button"
        tabindex="0"
        @click="current = r.path"
        @keydown.enter="current = r.path"
        @dragover="onOver($event, r.path)"
        @dragleave="dropOn = null"
        @drop="onDrop($event, r.path)"
      >
        <span class="tw" :class="{ shut: collapsed.has(r.path), none: !r.hasKids }" @click.stop="r.hasKids && toggle(r.path)">
          <Icon :icon="ChevronRight" :size="13" />
        </span>
        <Icon :icon="current === r.path ? FolderOpen : Folder" :size="16" />
        <span class="nm" :title="r.path">{{ r.name }}</span>
        <em v-if="r.count">{{ r.count }}</em>
        <span v-if="editable" class="mm" @click.stop><PopMenu :items="menu(r.path)" :label="t('studio.a11y.moreOf', { name: r.name })" /></span>
      </div>
    </TransitionGroup>
    <p v-if="!rows.length && editable" class="hint">{{ t('studio.folder.hint') }}</p>
  </nav>
</template>

<style scoped lang="scss">
.ftree {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.fh {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 4px 8px 10px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.08em;
  color: var(--st-ink-4);
}

.fl { display: flex; flex-direction: column; gap: 2px; }

.fr {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 34px;
  padding: 0 6px 0 calc(10px + var(--d, 0) * 16px);
  border: 0;
  border-radius: var(--r-sm);
  text-align: left;
  font: 500 13.5px var(--font-sans);
  color: var(--st-ink-2);
  background: none;
  cursor: pointer;
  transition: background var(--dur-fast), color var(--dur-fast), box-shadow var(--dur-fast);

  &:hover { background: var(--hover); }
  &.on { color: var(--st-ink); background: var(--tint); }
  &.on :deep(svg) { color: var(--ink); }
  &.drop { background: var(--tint); box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--ink) 60%, transparent) inset; }

  .nm { flex: 1; min-width: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  em { font: 500 11.5px var(--font-mono); font-style: normal; color: var(--st-ink-4); }

  .tw {
    width: 16px;
    display: grid;
    place-items: center;
    color: var(--st-ink-4);
    transform: rotate(90deg);
    transition: transform var(--dur-fast) var(--ease-out);

    &.shut { transform: none; }
    &.none { visibility: hidden; }
  }

  .mm { opacity: 0; transition: opacity var(--dur-fast); }
  &:hover .mm, .mm:focus-within { opacity: 1; }
  .mm :deep(.st-ibtn) { width: 26px; height: 26px; }
}

.hint { margin: 8px 10px 0; font-size: 12px; line-height: 1.6; color: var(--st-ink-4); }

.compact .fr { min-height: 30px; font-size: 13px; }

.fr-enter-active, .fr-leave-active { transition: opacity var(--dur-fast), transform var(--dur-fast) var(--ease-out); }
.fr-enter-from, .fr-leave-to { opacity: 0; transform: translateY(-4px); }
</style>
