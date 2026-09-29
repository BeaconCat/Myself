<script setup lang="ts">
import Icon from '../../components/ui/Icon.vue';
import { Check, Search, X } from 'lucide';
import { computed, ref } from 'vue';
import { MODULE_GROUPS, MODULE_REGISTRY } from '../../about/registry';
import ModulePreview from './modules/ModulePreview.vue';

/**
 * 添加模块弹窗：搜索 + 按分组（身份 / 数据 / 经历 / 喜好 / 工具 / 互动）排列 + 多选 + 迷你预览。
 * 可选 existing：当前已有模块类型（profile 只允许一个，已存在时置灰）。
 */
const props = defineProps<{ existing?: string[] }>();
const emit = defineEmits<{ close: []; add: [types: string[]] }>();

const SINGLE = new Set(['profile']);
const keyword = ref('');
const selected = ref<Set<string>>(new Set());

const filtered = computed(() => {
  const kw = keyword.value.trim().toLowerCase();
  if (!kw) return MODULE_REGISTRY;
  return MODULE_REGISTRY.filter(
    (m) => m.name.toLowerCase().includes(kw) || m.desc.toLowerCase().includes(kw) || m.type.includes(kw),
  );
});

const groups = computed(() =>
  MODULE_GROUPS.map((g) => ({ name: g, items: filtered.value.filter((m) => m.group === g) })).filter((g) => g.items.length),
);

const blocked = (type: string) => SINGLE.has(type) && !!props.existing?.includes(type);

function toggle(type: string): void {
  if (blocked(type)) return;
  const next = new Set(selected.value);
  if (next.has(type)) next.delete(type);
  else next.add(type);
  selected.value = next;
}

function confirm(): void {
  if (!selected.value.size) return;
  emit('add', [...selected.value]);
}

const SPAN_COLS: Record<number, number> = { 1: 4, 2: 8, 3: 12 };
</script>

<template>
  <Teleport to="body">
    <div class="picker-mask" @click.self="emit('close')">
      <div class="picker">
        <header class="p-head">
          <h3>添加模块</h3>
          <div class="p-search">
            <Icon :icon="Search" :stroke="2.2" />
            <input v-model="keyword" type="search" placeholder="搜索模块…" />
          </div>
          <button class="p-close" aria-label="关闭" @click="emit('close')">
            <Icon :icon="X" :stroke="2.4" />
          </button>
        </header>

        <div class="p-body">
          <section v-for="g in groups" :key="g.name" class="p-group">
            <h4>{{ g.name }}<small>{{ g.items.length }}</small></h4>
            <div class="p-grid">
              <button
                v-for="m in g.items"
                :key="m.type"
                class="p-card"
                :class="{ on: selected.has(m.type), off: blocked(m.type) }"
                :disabled="blocked(m.type)"
                @click="toggle(m.type)"
              >
                <div class="p-top">
                  <span class="p-icon">
                    <Icon :icon="m.icon" :stroke="1.8" />
                  </span>
                  <div class="p-meta">
                    <strong>{{ m.name }}<em>{{ m.type }}</em></strong>
                    <span>{{ m.desc }}</span>
                  </div>
                  <span class="p-check" aria-hidden="true">
                    <Icon :icon="Check" :stroke="3" />
                  </span>
                </div>
                <!-- 模块迷你预览 -->
                <div class="p-preview">
                  <ModulePreview :type="m.type" />
                </div>
                <div class="p-spans">
                  <span v-for="s in m.spans" :key="s" :class="{ def: s === m.defaultSpan }"><i :style="{ width: `${(SPAN_COLS[s] / 12) * 100}%` }" /></span>
                  <small>{{ m.variants.map((v) => v.label).join(' / ') }}</small>
                </div>
              </button>
            </div>
          </section>
        </div>

        <footer class="p-foot">
          <span class="p-count">{{ selected.size ? `已选 ${selected.size} 个模块` : '点击卡片可多选' }}</span>
          <button class="a-btn ghost" @click="emit('close')">取消</button>
          <button class="a-btn primary" :disabled="!selected.size" @click="confirm">
            添加所选模块
          </button>
        </footer>
      </div>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.picker-mask {
  position: fixed;
  inset: 0;
  z-index: 9700;
  display: grid;
  place-items: center;
  padding: 24px;
  background: var(--scrim);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  animation: mask-in 0.25s ease both;
}

@keyframes mask-in { from { opacity: 0; } }

.picker {
  width: min(860px, 100%);
  max-height: 84vh;
  display: flex;
  flex-direction: column;
  background: var(--elev);
  border-radius: var(--r-xl);
  box-shadow: var(--shadow-pop);
  overflow: hidden;
  animation: picker-in 0.38s var(--ease-spring) both;
}

@keyframes picker-in {
  from { opacity: 0; transform: translateY(26px) scale(0.94); }
  to { opacity: 1; transform: none; }
}

.p-head {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px 20px;
  box-shadow: inset 0 -1px 0 var(--line);

  h3 { font-size: 17px; flex-shrink: 0; }
}

.p-search {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  height: 38px;
  padding: 0 14px;
  border-radius: var(--r-pill);
  background: var(--fill);
  box-shadow: inset 0 0 0 1px var(--line);
  transition: background-color var(--dur-fast), box-shadow var(--dur-fast);

  svg { width: 15px; height: 15px; color: var(--text-2); flex-shrink: 0; }

  input {
    flex: 1;
    min-width: 0;
    border: none;
    outline: none;
    background: none;
    color: var(--text);
    font-size: 13.5px;
    font-family: inherit;
  }

  &:hover { box-shadow: inset 0 0 0 1px var(--line-2); }
  &:focus-within { background: var(--elev); box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--ink) 70%, transparent), 0 0 0 3px color-mix(in oklab, var(--ink) 18%, transparent); }
}

.p-close {
  width: 34px;
  height: 34px;
  border: none;
  border-radius: 50%;
  background: var(--fill);
  color: var(--text-2);
  display: grid;
  place-items: center;
  flex-shrink: 0;
  transition: background-color var(--dur-fast), color var(--dur-fast), transform var(--dur) var(--ease-spring);

  svg { width: 15px; height: 15px; }
  &:hover { color: var(--text); background: var(--fill-2); transform: rotate(90deg); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

.p-body {
  padding: 6px 20px 18px;
  overflow-y: auto;
}

.p-group {
  h4 {
    position: sticky;
    top: 0;
    z-index: 1;
    display: flex;
    align-items: baseline;
    gap: 8px;
    padding: 12px 0 8px;
    font-size: 12px;
    letter-spacing: 0.08em;
    color: var(--text-2);
    background: var(--elev);

    small { font-weight: 400; color: var(--text-3); }
  }
}

.p-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

@media (max-width: 640px) {
  .p-grid { grid-template-columns: 1fr; }
}

/* 模块卡：默认中性面 + 细描边；选中 = 抬升 + 轻染（--lift），右上勾为实底 --solid */
.p-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px;
  border: 0;
  border-radius: var(--r-lg);
  background: var(--fill);
  box-shadow: inset 0 0 0 1px var(--line);
  color: var(--text);
  text-align: left;
  cursor: pointer;
  transition: background-color var(--dur) var(--ease-out), box-shadow var(--dur) var(--ease-out), transform var(--dur-fast) var(--ease-spring);

  &:hover { background: var(--fill-2); box-shadow: inset 0 0 0 1px var(--line-2); }
  &:active { transform: scale(0.985); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }

  &.on {
    background: var(--lift);
    box-shadow: var(--lift-shadow);

    .p-check { opacity: 1; transform: scale(1); }
    .p-icon { color: var(--ink); }
  }
}

.p-top {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.p-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--r-sm);
  display: grid;
  place-items: center;
  flex-shrink: 0;
  background: var(--fill-2);
  color: var(--text-2);
  transition: color var(--dur);

  svg { width: 19px; height: 19px; }
}

.p-meta {
  flex: 1;
  min-width: 0;

  strong { display: flex; align-items: baseline; gap: 8px; font-size: 14px; }
  em { font-style: normal; font-weight: 400; font-size: 11px; font-family: var(--font-mono); color: var(--text-3); }

  span {
    display: block;
    margin-top: 3px;
    font-size: 11.5px;
    line-height: 1.5;
    color: var(--text-2);
  }
}

.p-check {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--solid);
  color: var(--on-solid);
  box-shadow: var(--btn-shadow);
  display: grid;
  place-items: center;
  flex-shrink: 0;
  opacity: 0;
  transform: scale(0.5);
  transition: all var(--dur-fast) var(--ease-spring);

  svg { width: 11px; height: 11px; }
}

.p-card.off { opacity: 0.45; cursor: not-allowed; &:hover { transform: none; background: var(--fill); box-shadow: inset 0 0 0 1px var(--line); } }

.p-spans {
  display: flex;
  align-items: center;
  gap: 4px;

  span { position: relative; width: 22px; height: 6px; border-radius: calc(var(--r-xs) * 0.45); background: var(--fill-3); overflow: hidden; }
  span i { position: absolute; inset: 0 auto 0 0; background: var(--text-3); }
  span.def i { background: var(--ink); }
  small { margin-left: auto; font-size: 11px; color: var(--text-2); }
}

/* 真实迷你预览（各模块样式见 modules/ModulePreview.vue） */
.p-preview {
  padding: 12px;
  border-radius: var(--r-md);
  background: var(--elev);
  box-shadow: 0 0 0 0.5px var(--line), 0 1px 2px rgb(16 24 40 / 0.06);
  min-height: 78px;
  display: flex;
  align-items: center;
  pointer-events: none;

  > * { width: 100%; }
}

.p-foot {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 20px;
  box-shadow: inset 0 1px 0 var(--line);
}

.p-count {
  flex: 1;
  font-size: 12.5px;
  color: var(--text-2);
}
</style>
