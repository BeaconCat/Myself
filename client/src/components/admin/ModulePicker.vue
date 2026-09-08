<script setup lang="ts">
import { computed, ref } from 'vue';
import { MODULE_REGISTRY } from '../../about/registry';
import ModulePreview from './modules/ModulePreview.vue';

/** 添加模块弹窗：搜索 + 多选 + 预览卡 */
const emit = defineEmits<{ close: []; add: [types: string[]] }>();

const keyword = ref('');
const selected = ref<Set<string>>(new Set());

const filtered = computed(() => {
  const kw = keyword.value.trim().toLowerCase();
  if (!kw) return MODULE_REGISTRY;
  return MODULE_REGISTRY.filter(
    (m) => m.name.toLowerCase().includes(kw) || m.desc.toLowerCase().includes(kw),
  );
});

function toggle(type: string): void {
  const next = new Set(selected.value);
  if (next.has(type)) next.delete(type);
  else next.add(type);
  selected.value = next;
}

function confirm(): void {
  if (!selected.value.size) return;
  emit('add', [...selected.value]);
}
</script>

<template>
  <Teleport to="body">
    <div class="picker-mask" @click.self="emit('close')">
      <div class="picker">
        <header class="p-head">
          <h3>添加模块</h3>
          <div class="p-search">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round">
              <circle cx="11" cy="11" r="7" />
              <path d="M20 20l-3.8-3.8" />
            </svg>
            <input v-model="keyword" type="search" placeholder="搜索模块…" />
          </div>
          <button class="p-close" aria-label="关闭" @click="emit('close')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round">
              <path d="M6 6l12 12M18 6L6 18" />
            </svg>
          </button>
        </header>

        <div class="p-grid">
          <button
            v-for="m in filtered"
            :key="m.type"
            class="p-card"
            :class="{ on: selected.has(m.type) }"
            @click="toggle(m.type)"
          >
            <div class="p-top">
              <span class="p-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                  <path :d="m.icon" />
                </svg>
              </span>
              <div class="p-meta">
                <strong>{{ m.name }}</strong>
                <span>{{ m.desc }}</span>
              </div>
              <span class="p-check" aria-hidden="true">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M5 13l4 4 10-10" />
                </svg>
              </span>
            </div>
            <!-- 模块真实迷你预览 -->
            <div class="p-preview">
              <ModulePreview :type="m.type" />
            </div>
          </button>
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
  background: rgba(8, 8, 12, 0.5);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  animation: mask-in 0.25s ease both;
}

@keyframes mask-in { from { opacity: 0; } }

.picker {
  width: min(760px, 100%);
  max-height: 84vh;
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: 0 30px 80px -20px rgba(0, 0, 0, 0.5);
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
  border-bottom: 1px solid var(--border);

  h3 { font-size: 17px; flex-shrink: 0; }
}

.p-search {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  padding: 8px 14px;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--bg);

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

  &:focus-within { border-color: var(--primary); }
}

.p-close {
  width: 34px;
  height: 34px;
  border: none;
  border-radius: 50%;
  background: var(--surface-2);
  color: var(--text-2);
  display: grid;
  place-items: center;
  flex-shrink: 0;
  transition: all var(--dur-fast);

  svg { width: 15px; height: 15px; }
  &:hover { color: var(--accent-red); transform: rotate(90deg); }
}

.p-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  padding: 18px 20px;
  overflow-y: auto;
}

@media (max-width: 640px) {
  .p-grid { grid-template-columns: 1fr; }
}

.p-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px;
  border: 1.5px solid var(--border);
  border-radius: 12px;
  background: var(--bg);
  text-align: left;
  cursor: pointer;
  transition: all var(--dur-fast) var(--ease-out);

  &:hover { border-color: rgba(var(--primary-rgb), 0.5); transform: scale(1.02); }

  &.on {
    border-color: var(--primary);
    background: rgba(var(--primary-rgb), 0.06);
    box-shadow: 0 0 0 3px rgba(var(--primary-rgb), 0.12);

    .p-check { opacity: 1; transform: scale(1); }
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
  border-radius: 10px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
  background: rgba(var(--primary-rgb), 0.1);
  color: var(--primary);

  svg { width: 19px; height: 19px; }
}

.p-meta {
  flex: 1;
  min-width: 0;

  strong { display: block; font-size: 14px; }

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
  background: var(--primary);
  color: #fff;
  display: grid;
  place-items: center;
  flex-shrink: 0;
  opacity: 0;
  transform: scale(0.5);
  transition: all var(--dur-fast) var(--ease-spring);

  svg { width: 11px; height: 11px; }
}

/* 真实迷你预览（各模块样式见 modules/ModulePreview.vue） */
.p-preview {
  padding: 12px;
  border-radius: 8px;
  background: var(--surface);
  border: 1px solid var(--border);
  min-height: 64px;
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
  border-top: 1px solid var(--border);
}

.p-count {
  flex: 1;
  font-size: 12.5px;
  color: var(--text-2);
}
</style>
