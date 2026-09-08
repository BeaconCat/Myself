<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type CompressJob, type CompressResult, type QualityItem } from '../../api';
import { useDialogStore } from '../../stores/dialog';

const { t } = useI18n();

const items = ref<QualityItem[]>([]);
const selected = ref<Set<string>>(new Set());
const quality = ref(80);
const busy = ref(false);
const results = ref<CompressResult[]>([]);
/** 后台任务进度（轮询） */
const job = ref<CompressJob | null>(null);
let pollTimer = 0;
const dialog = useDialogStore();

const progressPercent = computed(() =>
  job.value && job.value.total ? Math.round((job.value.done / job.value.total) * 100) : 0,
);

const totalSelectedSize = computed(() =>
  items.value.filter((i) => selected.value.has(i.name)).reduce((sum, i) => sum + i.size, 0),
);

async function scan(): Promise<void> {
  items.value = await adminApi.qualityScan();
  selected.value = new Set();
  results.value = [];
}

function toggle(name: string): void {
  const next = new Set(selected.value);
  if (next.has(name)) next.delete(name);
  else next.add(name);
  selected.value = next;
}

function selectAll(): void {
  selected.value = selected.value.size === items.value.length
    ? new Set()
    : new Set(items.value.map((i) => i.name));
}

/** 发起后台任务后每 600ms 轮询，直到 running=false */
async function compress(): Promise<void> {
  if (busy.value || !selected.value.size) return;
  busy.value = true;
  results.value = [];
  try {
    const { id } = await adminApi.qualityCompress([...selected.value], quality.value);
    job.value = await adminApi.qualityJob(id);
    await new Promise<void>((resolve) => {
      const tick = async (): Promise<void> => {
        try {
          job.value = await adminApi.qualityJob(id);
        } catch {
          resolve();
          return;
        }
        if (job.value.running) pollTimer = window.setTimeout(() => void tick(), 600);
        else resolve();
      };
      pollTimer = window.setTimeout(() => void tick(), 600);
    });
    results.value = job.value?.results ?? [];
    items.value = await adminApi.qualityScan();
    selected.value = new Set();
  } catch (err) {
    if ((err as Error).message === 'job_running') void dialog.alert({ message: t('admin.compressRunning') });
    else throw err;
  } finally {
    busy.value = false;
    job.value = null;
  }
}

onBeforeUnmount(() => window.clearTimeout(pollTimer));

function formatSize(bytes: number): string {
  if (bytes > 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
  return `${Math.round(bytes / 1024)} KB`;
}

function savedPercent(r: CompressResult): string {
  if (!r.before || !r.after) return '-';
  return `${Math.round((1 - r.after / r.before) * 100)}%`;
}

onMounted(scan);
</script>

<template>
  <div>
    <header class="head">
      <div>
        <h1 class="page-h">{{ t('admin.menuQuality') }}</h1>
        <p class="hint">{{ t('admin.qualityHint') }}</p>
      </div>
    </header>

    <!-- 压缩控制条 -->
    <div class="control">
      <button class="btn ghost" @click="selectAll">{{ t('admin.selectAll') }}</button>
      <label class="q-label">
        <span>{{ t('admin.compressQuality', { q: quality }) }}</span>
        <input v-model.number="quality" type="range" min="30" max="95" step="5" />
      </label>
      <span class="sel-info">{{ t('admin.selectedInfo', { n: selected.size, size: formatSize(totalSelectedSize) }) }}</span>
      <button class="btn primary" :disabled="busy || !selected.size" @click="compress">
        {{ busy ? t('admin.compressing') : t('admin.compress') }}
      </button>
    </div>

    <!-- 后台任务进度 -->
    <div v-if="job" class="progress">
      <div class="bar"><span :style="{ width: `${progressPercent}%` }" /></div>
      <span class="progress-text">
        {{ t('admin.compressProgress', { done: job.done, total: job.total }) }}
        <em v-if="job.current">{{ job.current }}</em>
      </span>
    </div>

    <!-- 结果 -->
    <div v-if="results.length" class="results">
      <p v-for="r in results" :key="r.name">
        <template v-if="r.error">{{ r.name }} — {{ t('admin.compressFailed') }}</template>
        <template v-else>
          {{ r.name }}<template v-if="r.newName && r.newName !== r.name"> → {{ r.newName }}</template>
          ：{{ formatSize(r.before ?? 0) }} → {{ formatSize(r.after ?? 0) }}（-{{ savedPercent(r) }}）
        </template>
      </p>
    </div>

    <!-- 图片列表 -->
    <table class="table">
      <thead>
        <tr>
          <th />
          <th>{{ t('admin.qualityImage') }}</th>
          <th>{{ t('admin.qualityFormat') }}</th>
          <th>{{ t('admin.qualityDims') }}</th>
          <th>{{ t('admin.size') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="item in items"
          :key="item.name"
          :class="{ on: selected.has(item.name) }"
          @click="toggle(item.name)"
        >
          <td><input type="checkbox" :checked="selected.has(item.name)" @click.stop="toggle(item.name)" /></td>
          <td class="img-cell">
            <img :src="item.url" loading="lazy" alt="" />
            <span>{{ item.name }}</span>
          </td>
          <td>
            <span class="fmt">{{ item.format.toUpperCase() }}</span>
            <span v-if="item.hasAlpha" class="alpha">alpha</span>
          </td>
          <td>{{ item.width }}×{{ item.height }}</td>
          <td>{{ formatSize(item.size) }}</td>
        </tr>
      </tbody>
    </table>

    <p v-if="!items.length" class="empty">{{ t('admin.qualityEmpty') }}</p>
  </div>
</template>

<style scoped lang="scss">
.head { margin-bottom: 18px; }
.page-h { font-size: 26px; margin-bottom: 6px; }
.hint { font-size: 13px; color: var(--text-2); }

.control {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  padding: 14px 18px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  margin-bottom: 16px;
}

/* 进度条：主色填充，宽度过渡走 motion token */
.progress {
  display: flex;
  align-items: center;
  gap: 14px;
  margin: -6px 0 16px;
  font-size: 13px;
  color: var(--text-2);

  .bar {
    flex: 1;
    height: 6px;
    border-radius: 999px;
    background: var(--surface-2);
    overflow: hidden;

    span {
      display: block;
      height: 100%;
      border-radius: inherit;
      background: var(--primary);
      transition: width var(--dur) var(--ease-out);
    }
  }

  em {
    font-style: normal;
    margin-left: 8px;
    font-family: Consolas, monospace;
    opacity: 0.8;
  }
}

.q-label {
  display: flex;
  flex-direction: column;
  gap: 4px;

  span { font-size: 12px; font-weight: 600; color: var(--text-2); }
  input { accent-color: var(--primary); width: 180px; }
}

.sel-info {
  font-size: 13px;
  color: var(--text-2);
  margin-left: auto;
}

.btn {
  padding: 9px 22px;
  border: 1px solid transparent;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 700;
  transition: all var(--dur-fast) var(--ease-out);

  &.primary {
    color: #fff;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.4);

    &:hover:not(:disabled) { filter: brightness(1.08); }
    &:disabled { opacity: 0.55; }
  }

  &.ghost {
    background: var(--surface);
    border-color: var(--border);
    color: var(--text);

    &:hover { border-color: var(--primary); color: var(--primary); }
  }
}

.results {
  padding: 12px 18px;
  margin-bottom: 16px;
  background: rgba(var(--primary-rgb), 0.06);
  border: 1px solid rgba(var(--primary-rgb), 0.3);
  border-radius: var(--radius);
  font-size: 13px;
  line-height: 1.9;
}

.table {
  width: 100%;
  border-collapse: collapse;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;

  th, td {
    padding: 10px 14px;
    text-align: left;
    font-size: 14px;
    border-bottom: 1px solid var(--border);
  }

  th { background: var(--surface-2); font-size: 12px; color: var(--text-2); }

  tbody tr {
    cursor: pointer;
    transition: background var(--dur-fast);

    &:hover { background: rgba(var(--primary-rgb), 0.04); }
    &.on { background: rgba(var(--primary-rgb), 0.08); }
  }

  input[type='checkbox'] { accent-color: var(--primary); }
}

.img-cell {
  display: flex;
  align-items: center;
  gap: 10px;

  img {
    width: 44px;
    height: 33px;
    object-fit: cover;
    border-radius: 6px;
    background: var(--surface-2);
  }

  span {
    font-family: Consolas, monospace;
    font-size: 12px;
  }
}

.fmt {
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--surface-2);
}

.alpha {
  font-size: 10px;
  font-weight: 700;
  margin-left: 6px;
  padding: 2px 6px;
  border-radius: 999px;
  background: rgba(var(--primary-rgb), 0.12);
  color: var(--primary);
}

.empty {
  color: var(--text-2);
  text-align: center;
  padding: 48px 0;
}
</style>
