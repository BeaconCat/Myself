<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, thumbOf, type CompressJob, type CompressResult, type QualityItem } from '../../../api';
import { useDialogStore } from '../../../stores/dialog';
import './i18n';
import SIcon from './SIcon.vue';
import DoorArt from './DoorArt.vue';
import { toast } from './toast';
import { formatSize } from './format';

/** 图片优化：扫描 → 勾选 → 后台压缩任务（轮询进度）→ 结果汇总 */
const emit = defineEmits<{ done: []; count: [n: number] }>();
const { t } = useI18n();
const dialog = useDialogStore();

const items = ref<QualityItem[]>([]);
const loaded = ref(false);
const onlyCompressible = ref(true);
const selected = ref<Set<string>>(new Set());
const quality = ref(80);
const job = ref<CompressJob | null>(null);
const results = ref<CompressResult[]>([]);
let pollTimer = 0;

const shown = computed(() => (onlyCompressible.value ? items.value.filter((i) => i.compressible) : items.value));
const compressibleCount = computed(() => items.value.filter((i) => i.compressible).length);
const selectedSize = computed(() => items.value.filter((i) => selected.value.has(i.name)).reduce((s, i) => s + i.size, 0));
const allOn = computed(() => shown.value.length > 0 && shown.value.every((i) => selected.value.has(i.name)));
const percent = computed(() => (job.value && job.value.total ? Math.round((job.value.done / job.value.total) * 100) : 0));
const summary = computed(() => {
  const ok = results.value.filter((r) => !r.error);
  const before = ok.reduce((s, r) => s + (r.before ?? 0), 0);
  const after = ok.reduce((s, r) => s + (r.after ?? 0), 0);
  return { n: ok.length, failed: results.value.length - ok.length, before, after, pct: before ? Math.round((1 - after / before) * 100) : 0 };
});

async function scan(): Promise<void> {
  try {
    items.value = await adminApi.qualityScan();
    selected.value = new Set(items.value.filter((i) => i.compressible).map((i) => i.name));
    emit('count', compressibleCount.value);
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
  loaded.value = true;
}

function toggle(name: string): void {
  const next = new Set(selected.value);
  if (next.has(name)) next.delete(name);
  else next.add(name);
  selected.value = next;
}

function toggleAll(): void {
  selected.value = allOn.value ? new Set() : new Set(shown.value.map((i) => i.name));
}

async function run(): Promise<void> {
  if (job.value || !selected.value.size) return;
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
    toast(t('studio.media.qDone', { size: formatSize(Math.max(0, summary.value.before - summary.value.after)) }));
    await scan();
    emit('done');
  } catch (err) {
    if ((err as Error).message === 'job_running') void dialog.alert({ title: t('studio.media.jobRunning') });
    else toast(t('studio.saveFailed'), { icon: 'x' });
  } finally {
    job.value = null;
  }
}

onMounted(scan);
onBeforeUnmount(() => window.clearTimeout(pollTimer));
</script>

<template>
  <div class="q">
    <div class="ctl">
      <div class="q-range">
        <div class="st-flabel"><span>{{ t('studio.media.qQuality') }}</span><span class="mono">{{ quality }}</span></div>
        <input v-model.number="quality" type="range" min="30" max="95" step="5" :style="{ '--p': `${((quality - 30) / 65) * 100}%` }" />
        <small>{{ t('studio.media.qHint') }}</small>
      </div>
      <div class="sel">
        <b class="mono">{{ selected.size }}</b>
        <small>{{ t('studio.media.qSelected', { size: formatSize(selectedSize) }) }}</small>
      </div>
      <button type="button" class="st-btn p lg" :disabled="!!job || !selected.size" @click="run">
        <SIcon name="compress" :size="16" />{{ job ? t('studio.media.qRunning') : t('studio.media.qRun') }}
      </button>
    </div>

    <div v-if="job" class="progress">
      <div class="bar"><i :style="{ width: `${percent}%` }" /></div>
      <span class="mono">{{ job.done }} / {{ job.total }}</span>
      <em class="mono">{{ job.current }}</em>
    </div>

    <div v-if="results.length && !job" class="result">
      <span class="ri"><SIcon name="check" /></span>
      <div>
        <b>{{ t('studio.media.qResult', { n: summary.n, pct: summary.pct }) }}</b>
        <small class="mono">{{ formatSize(summary.before) }} → {{ formatSize(summary.after) }}</small>
        <small v-if="summary.failed">{{ t('studio.media.qFailed', { n: summary.failed }) }}</small>
      </div>
    </div>

    <div class="list-h">
      <label class="st-ckrow" @click.prevent="toggleAll">
        <span class="st-ck" :class="{ on: allOn }"><svg viewBox="0 0 24 24"><path d="M5 12.5l4.5 4.5L19 7" /></svg></span>
        {{ t('studio.media.qAll') }}
      </label>
      <span class="sp" />
      <label class="st-ckrow small" @click.prevent="onlyCompressible = !onlyCompressible">
        <span class="st-ck" :class="{ on: onlyCompressible }"><svg viewBox="0 0 24 24"><path d="M5 12.5l4.5 4.5L19 7" /></svg></span>
        {{ t('studio.media.qOnly', { n: compressibleCount }) }}
      </label>
    </div>

    <div v-if="loaded && !shown.length" class="st-empty">
      <DoorArt />
      <h4>{{ t('studio.media.qEmpty') }}</h4>
      <p>{{ t('studio.media.qEmptySub') }}</p>
    </div>

    <div class="rows">
      <div
        v-for="(it, i) in shown"
        :key="it.name"
        class="row st-rise"
        :class="{ on: selected.has(it.name) }"
        :style="{ '--i': Math.min(i, 10) }"
        @click="toggle(it.name)"
      >
        <span class="st-ck" :class="{ on: selected.has(it.name) }"><svg viewBox="0 0 24 24"><path d="M5 12.5l4.5 4.5L19 7" /></svg></span>
        <img :src="thumbOf(it.url)" alt="" loading="lazy" />
        <div class="nm">
          <b class="mono">{{ it.name }}</b>
          <small>{{ it.width }} × {{ it.height }}</small>
        </div>
        <span class="fmt">{{ it.format.toUpperCase() }}<i v-if="it.hasAlpha">alpha</i></span>
        <span class="mono size">{{ formatSize(it.size) }}</span>
        <span class="flag" :class="{ yes: it.compressible }">{{ it.compressible ? t('studio.media.compressible') : t('studio.media.optimal') }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.ctl {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  gap: 32px;
  align-items: center;
  padding: 20px 24px;
  border-radius: var(--r-lg);
  background: var(--well);
  margin-bottom: 20px;
}

.q-range {
  max-width: 420px;

  small { display: block; font-size: 13px; color: var(--st-ink-3); margin-top: 8px; }

  input[type='range'] {
    appearance: none;
    width: 100%;
    height: 6px;
    border-radius: calc(var(--r-xs) / 2);
    background: linear-gradient(90deg, var(--ink) var(--p), var(--well-2) var(--p));
    outline: none;

    &::-webkit-slider-thumb {
      appearance: none;
      width: 20px;
      height: 20px;
      border-radius: 50%;
      background: #fff;
      box-shadow: 0 1px 4px rgba(0, 0, 0, 0.25), 0 0 0 1px var(--line-2);
      cursor: pointer;
      transition: transform var(--dur-fast) var(--ease-spring);
    }

    &:active::-webkit-slider-thumb { transform: scale(1.15); }
  }
}

.sel {
  text-align: right;

  b { display: block; font-size: 30px; line-height: 1.1; font-weight: 600; font-variant-numeric: tabular-nums; }
  small { font-size: 12.5px; color: var(--st-ink-3); }
}

.progress {
  display: flex;
  align-items: center;
  gap: 14px;
  margin: 0 0 22px;
  font-size: 12.5px;
  color: var(--st-ink-3);

  .bar {
    flex: 1;
    height: 6px;
    border-radius: calc(var(--r-xs) / 2);
    background: var(--well-2);
    overflow: hidden;

    i { display: block; height: 100%; border-radius: inherit; background: var(--ink); transition: width var(--dur) var(--ease-out); }
  }

  em { font-style: normal; max-width: 240px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
}

.result {
  display: flex;
  gap: 14px;
  align-items: center;
  padding: 16px 18px;
  border-radius: var(--r-md);
  margin-bottom: 22px;
  background: color-mix(in oklab, var(--green) 8%, var(--paper));
  box-shadow: 0 0 0 1px color-mix(in oklab, var(--green) 22%, transparent) inset;
  animation: rise-in var(--dur-slow) var(--ease-spring);

  .ri {
    width: 38px;
    height: 38px;
    border-radius: var(--r-sm);
    display: grid;
    place-items: center;
    background: color-mix(in oklab, var(--green) 16%, var(--paper));
    color: color-mix(in oklab, var(--green) 70%, var(--st-ink));
  }

  b { display: block; font-size: 14.5px; font-weight: 500; }
  small { font-size: 12.5px; color: var(--st-ink-3); margin-right: 12px; }
}

@keyframes rise-in { from { opacity: 0; transform: translateY(10px); } }

.list-h {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 12px 10px;
  border-bottom: 1px solid var(--line);
  margin-bottom: 6px;

  .sp { flex: 1; }
  .small { font-size: 13px; }
}

.row {
  display: grid;
  grid-template-columns: 18px 64px minmax(0, 1fr) 90px 90px 76px;
  gap: 16px;
  align-items: center;
  padding: 10px 12px;
  border-radius: var(--r-md);
  cursor: pointer;
  transition: background var(--dur-fast);

  &:hover { background: var(--well); }
  &.on { background: var(--tint); }

  img { width: 64px; height: 44px; object-fit: cover; border-radius: var(--r-xs); background: var(--well-2); display: block; }

  .nm {
    min-width: 0;

    b { display: block; font-size: 12.5px; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    small { font-size: 12px; color: var(--st-ink-3); }
  }

  .fmt {
    font-size: 11.5px;
    font-weight: 500;
    color: var(--st-ink-2);

    i {
      font-style: normal;
      margin-left: 6px;
      padding: 1px 6px;
      border-radius: var(--r-xs);
      background: var(--tint);
      color: var(--ink);
      font-size: 10.5px;
    }
  }

  .size { font-size: 12px; color: var(--st-ink-2); text-align: right; }

  .flag {
    font-size: 11.5px;
    text-align: center;
    padding: 2px 0;
    border-radius: var(--r-xs);
    color: var(--st-ink-3);
    background: var(--well-2);

    &.yes {
      background: color-mix(in oklab, var(--yellow) 18%, var(--paper));
      color: color-mix(in oklab, var(--yellow) 50%, var(--st-ink));
    }
  }
}
</style>
