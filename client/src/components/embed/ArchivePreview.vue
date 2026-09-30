<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { ChevronRight, Download, FileText, Folder, FolderOpen, X } from 'lucide';
import { api, type ArchiveListing } from '../../api';
import Icon from '../ui/Icon.vue';

/**
 * 压缩包在线预览：正文里压缩包卡片的「预览内容」按钮（data-archive，见 utils/embeds.ts）经全局点击委托打开，
 * 只读中央目录列出文件树（文件夹可折叠），不解压、不下载内容。前台 / 编辑器通用，挂在 App 根部。
 */
const { t } = useI18n();
const open = ref(false);
const src = ref('');
const title = ref('');
const data = ref<ArchiveListing | null>(null);
const error = ref('');
const collapsed = ref<Set<string>>(new Set());

function onClick(e: MouseEvent): void {
  const btn = (e.target as HTMLElement | null)?.closest<HTMLElement>('[data-archive]');
  if (!btn) return;
  e.preventDefault();
  e.stopPropagation();
  src.value = btn.dataset.archive ?? '';
  title.value = btn.dataset.title ?? '';
  open.value = true;
}

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape' && open.value) open.value = false;
}

onMounted(() => {
  document.addEventListener('click', onClick, true);
  document.addEventListener('keydown', onKey);
});
onBeforeUnmount(() => {
  document.removeEventListener('click', onClick, true);
  document.removeEventListener('keydown', onKey);
});

watch(open, async (v) => {
  if (!v) return;
  data.value = null;
  error.value = '';
  collapsed.value = new Set();
  const name = src.value.split(/[?#]/)[0].split('/').pop() ?? '';
  try {
    data.value = await api.archive(decodeURIComponent(name));
  } catch (err) {
    error.value = String((err as Error).message).includes('422') ? t('content.embed.archiveBad') : t('content.embed.archiveFailed');
  }
});

interface Row { path: string; name: string; depth: number; dir: boolean; size: number }

/** 条目 → 树形行：补全缺失的中间目录，按目录在前、名称排序；折叠的目录隐藏其子项 */
const rows = computed<Row[]>(() => {
  const list = data.value?.entries ?? [];
  const map = new Map<string, Row>();
  for (const e of list) {
    const parts = e.name.replace(/\/$/, '').split('/').filter(Boolean);
    parts.forEach((p, i) => {
      const path = parts.slice(0, i + 1).join('/');
      const isLeaf = i === parts.length - 1;
      if (!map.has(path)) map.set(path, { path, name: p, depth: i, dir: !isLeaf || e.dir, size: isLeaf ? e.size : 0 });
    });
  }
  const all = [...map.values()].sort((a, b) => {
    const pa = a.path.split('/');
    const pb = b.path.split('/');
    for (let i = 0; i < Math.min(pa.length, pb.length); i++) {
      if (pa[i] === pb[i]) continue;
      const da = i < pa.length - 1 || a.dir;
      const db = i < pb.length - 1 || b.dir;
      if (da !== db) return da ? -1 : 1;
      return pa[i].localeCompare(pb[i]);
    }
    return pa.length - pb.length;
  });
  return all.filter((r) => ![...collapsed.value].some((c) => r.path.startsWith(`${c}/`)));
});

const stats = computed(() => {
  const list = data.value?.entries ?? [];
  const files = list.filter((e) => !e.dir).length;
  const dirs = new Set(list.flatMap((e) => {
    const parts = e.name.replace(/\/$/, '').split('/');
    return parts.slice(0, e.dir ? parts.length : -1).map((_, i) => parts.slice(0, i + 1).join('/'));
  })).size;
  return { files, dirs };
});

function toggle(r: Row): void {
  if (!r.dir) return;
  const next = new Set(collapsed.value);
  if (next.has(r.path)) next.delete(r.path);
  else next.add(r.path);
  collapsed.value = next;
}

function size(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / 1024 / 1024).toFixed(1)} MB`;
}
</script>

<template>
  <Teleport to="body">
    <Transition name="arc">
      <div v-if="open" class="arc-scrim" @mousedown.self="open = false">
        <div class="arc" role="dialog" aria-modal="true">
          <header>
            <div class="tt">
              <b>{{ title || data?.title || t('content.embed.archiveTitle') }}</b>
              <small v-if="data">{{ t('content.embed.archiveCount', { n: stats.files, dirs: stats.dirs }) }} · {{ size(data.size) }}</small>
            </div>
            <a class="dl" :href="src" download><Icon :icon="Download" :size="16" />{{ t('content.embed.download') }}</a>
            <button type="button" class="x" :aria-label="t('content.embed.open')" @click="open = false"><Icon :icon="X" :size="18" /></button>
          </header>
          <div class="body">
            <p v-if="error" class="msg">{{ error }}</p>
            <div v-else-if="!data" class="sk"><i v-for="n in 6" :key="n" :style="{ width: `${40 + ((n * 37) % 50)}%` }" /></div>
            <p v-else-if="!rows.length" class="msg">{{ t('content.embed.empty') }}</p>
            <ul v-else class="tree">
              <li v-for="r in rows" :key="r.path" :class="{ dir: r.dir }" :style="{ '--d': r.depth }" @click="toggle(r)">
                <Icon v-if="r.dir" class="chev" :class="{ shut: collapsed.has(r.path) }" :icon="ChevronRight" :size="14" />
                <Icon :icon="r.dir ? (collapsed.has(r.path) ? Folder : FolderOpen) : FileText" :size="16" />
                <span class="nm">{{ r.name }}</span>
                <span v-if="!r.dir" class="sz">{{ size(r.size) }}</span>
              </li>
            </ul>
            <p v-if="data?.truncated" class="msg">{{ t('content.embed.archiveTruncated', { n: data.entries.length }) }}</p>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped lang="scss">
.arc-scrim {
  position: fixed;
  inset: 0;
  z-index: 1200;
  display: grid;
  place-items: center;
  padding: 24px;
  background: color-mix(in oklab, var(--bg) 40%, rgba(0, 0, 0, 0.45));
  backdrop-filter: blur(6px);
}

.arc {
  width: min(640px, 100%);
  max-height: min(78vh, 760px);
  display: flex;
  flex-direction: column;
  border-radius: var(--r-lg);
  background: var(--bg);
  box-shadow: 0 0 0 1px var(--line, rgba(128, 128, 128, 0.2)), 0 24px 64px rgba(0, 0, 0, 0.25);
  overflow: hidden;
}

header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 18px 18px 14px 22px;
  border-bottom: 1px solid var(--line, rgba(128, 128, 128, 0.18));

  .tt { flex: 1; min-width: 0; }
  b { display: block; font: 700 17px/1.4 var(--font-serif); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  small { font: 12.5px var(--font-mono); color: var(--text-2); }

  .dl {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 34px;
    padding: 0 13px;
    border-radius: var(--r-sm);
    font-size: 13.5px;
    color: var(--text);
    box-shadow: 0 0 0 1px var(--line, rgba(128, 128, 128, 0.25));
    transition: color var(--dur-fast);

    &:hover { color: var(--primary); }
  }

  .x {
    width: 34px;
    height: 34px;
    display: grid;
    place-items: center;
    border: 0;
    border-radius: var(--r-sm);
    cursor: pointer;
    color: var(--text-2);
    background: none;

    &:hover { color: var(--text); background: var(--fill-2, rgba(128, 128, 128, 0.12)); }
  }
}

.body { flex: 1; overflow: auto; padding: 8px 10px 14px; }

.tree {
  list-style: none;
  margin: 0;
  padding: 0;

  li {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 34px;
    padding: 0 12px 0 calc(12px + var(--d) * 18px);
    border-radius: var(--r-sm);
    font-size: 14px;
    color: var(--text);

    &:hover { background: var(--fill-2, rgba(128, 128, 128, 0.1)); }
    &.dir { cursor: pointer; font-weight: 600; }
    &:not(.dir) { padding-left: calc(34px + var(--d) * 18px); }
    :deep(svg) { flex: none; color: var(--text-2); }
    &.dir :deep(svg) { color: var(--primary); }
  }

  .chev { transform: rotate(90deg); transition: transform var(--dur-fast) var(--ease-out); }
  .chev.shut { transform: none; }
  .nm { flex: 1; min-width: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .sz { font: 12px var(--font-mono); color: var(--text-2); }
}

.msg { margin: 16px 12px; font-size: 13.5px; color: var(--text-2); text-align: center; }

.sk {
  display: grid;
  gap: 12px;
  padding: 14px 12px;

  i { height: 14px; border-radius: var(--r-xs); background: var(--fill-2, rgba(128, 128, 128, 0.12)); animation: arc-pulse 1.2s ease-in-out infinite alternate; }
}

@keyframes arc-pulse { to { opacity: 0.45; } }

.arc-enter-active, .arc-leave-active { transition: opacity var(--dur) var(--ease-out); }
.arc-enter-active .arc, .arc-leave-active .arc { transition: transform var(--dur) var(--ease-spring); }
.arc-enter-from, .arc-leave-to { opacity: 0; }
.arc-enter-from .arc, .arc-leave-to .arc { transform: translateY(12px) scale(0.97); }

@media (prefers-reduced-motion: reduce) {
  .arc-enter-from .arc, .arc-leave-to .arc { transform: none; }
}
</style>
