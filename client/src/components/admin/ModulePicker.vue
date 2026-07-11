<script setup lang="ts">
import { computed, ref } from 'vue';
import { MODULE_REGISTRY } from '../../about/registry';

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
              <div v-if="m.type === 'stats'" class="pv-stats">
                <span v-for="s in [['365', '天'], ['42', '文'], ['128', '想'], ['9', '签']]" :key="s[1]">
                  <strong>{{ s[0] }}</strong><i>{{ s[1] }}</i>
                </span>
              </div>

              <blockquote v-else-if="m.type === 'motto'" class="pv-motto">记录本身，就是意义。</blockquote>

              <div v-else-if="m.type === 'skills' || m.type === 'favorites'" class="pv-chips">
                <b>{{ m.type === 'skills' ? '创作' : '电影' }}</b>
                <span v-for="c in (m.type === 'skills' ? ['文章', '摄影'] : ['科幻', '悬疑'])" :key="c">{{ c }}</span>
              </div>

              <div v-else-if="m.type === 'skillbars'" class="pv-bars">
                <div v-for="b in [['写作', 80], ['摄影', 55]]" :key="b[0]">
                  <em>{{ b[0] }}</em>
                  <i><u :style="{ width: b[1] + '%' }" /></i>
                </div>
              </div>

              <div v-else-if="m.type === 'languages'" class="pv-lang">
                <i style="width: 45%; background: #0078ff" />
                <i style="width: 35%; background: #00c853" />
                <i style="width: 20%; background: #ffb300" />
              </div>

              <div v-else-if="m.type === 'milestones'" class="pv-ms">
                <div><b>2026</b><i /><span>点亮站点</span></div>
                <div><b>2025</b><i /><span>开始记录</span></div>
              </div>

              <div v-else-if="m.type === 'gallery'" class="pv-gallery">
                <i v-for="n in 4" :key="n" />
              </div>

              <figure v-else-if="m.type === 'quotes'" class="pv-quote">
                <span>把喜欢的话收藏在这里。</span>
                <em>—— 出处</em>
              </figure>

              <div v-else-if="m.type === 'devices' || m.type === 'stack'" class="pv-tiles">
                <span v-for="d in (m.type === 'devices' ? ['主机', '相机'] : ['Vue', 'Vite'])" :key="d">{{ d }}</span>
              </div>

              <div v-else-if="m.type === 'faq'" class="pv-faq">
                <div><span>一个常见的问题？</span><b>+</b></div>
                <div><span>另一个问题？</span><b>+</b></div>
              </div>

              <ul v-else-if="m.type === 'now'" class="pv-now">
                <li><i />在写一篇新文章</li>
                <li><i />在学一门新语言</li>
              </ul>

              <div v-else-if="m.type === 'github'" class="pv-stats">
                <span v-for="s in [['14', '仓'], ['4', '星'], ['7', '粉'], ['99', '提']]" :key="s[1]">
                  <strong>{{ s[0] }}</strong><i>{{ s[1] }}</i>
                </span>
              </div>

              <div v-else-if="m.type === 'socials'" class="pv-pills">
                <span>GitHub</span><span>Email</span><span>RSS</span>
              </div>

              <div v-else-if="m.type === 'contact'" class="pv-contact">
                <div><b>想聊聊？</b><span>合作或打个招呼</span></div>
                <i>发邮件</i>
              </div>
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

/* 真实迷你预览 */
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

.pv-stats {
  display: flex;
  gap: 8px;

  span {
    flex: 1;
    text-align: center;
    padding: 6px 2px;
    border-radius: 6px;
    background: var(--surface-2);

    strong {
      display: block;
      font-family: var(--font-serif);
      font-size: 14px;
      background: var(--grad-title);
      background-clip: text;
      -webkit-background-clip: text;
      color: transparent;
    }

    i { font-style: normal; font-size: 9.5px; color: var(--text-2); }
  }
}

.pv-motto {
  margin: 0 auto;
  width: fit-content;
  padding: 6px 14px;
  font-family: var(--font-serif);
  font-size: 12px;
  color: var(--text-2);
  border-left: 2px solid rgba(var(--primary-rgb), 0.6);
  border-right: 2px solid rgba(var(--primary-rgb), 0.6);
  background: rgba(var(--primary-rgb), 0.05);
  border-radius: 5px;
}

.pv-chips {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;

  b { font-size: 11px; color: var(--primary); margin-right: 2px; }

  span {
    font-size: 10.5px;
    font-weight: 600;
    padding: 2px 9px;
    background: var(--surface-2);
  }
}

.pv-bars {
  display: flex;
  flex-direction: column;
  gap: 7px;

  > div { display: flex; align-items: center; gap: 8px; }
  em { font-style: normal; font-size: 10.5px; width: 30px; flex-shrink: 0; }

  i {
    flex: 1;
    height: 6px;
    border-radius: 999px;
    background: var(--surface-2);
    overflow: hidden;
    display: block;

    u {
      display: block;
      height: 100%;
      border-radius: inherit;
      background: linear-gradient(90deg, var(--primary), var(--primary-deep));
    }
  }
}

.pv-lang {
  display: flex;
  height: 12px;
  border-radius: 999px;
  overflow: hidden;

  i { display: block; }
}

.pv-ms {
  display: flex;
  flex-direction: column;
  gap: 6px;

  > div { display: flex; align-items: center; gap: 8px; font-size: 10.5px; }
  b { font-family: var(--font-serif); color: var(--primary); font-size: 12px; }

  i {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--primary);
  }

  span { color: var(--text-2); }
}

.pv-gallery {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 4px;

  i {
    aspect-ratio: 1;
    border-radius: 4px;
    background: linear-gradient(135deg, rgba(var(--primary-rgb), 0.35), var(--surface-2));
  }
}

.pv-quote {
  margin: 0;

  span {
    display: block;
    font-family: var(--font-serif);
    font-size: 11.5px;
  }

  em {
    display: block;
    text-align: right;
    font-style: normal;
    font-size: 10px;
    color: var(--text-2);
    margin-top: 4px;
  }
}

.pv-tiles {
  display: flex;
  gap: 6px;

  span {
    flex: 1;
    text-align: center;
    font-family: var(--font-serif);
    font-size: 11px;
    padding: 8px 4px;
    background: var(--surface-2);
  }
}

.pv-faq {
  display: flex;
  flex-direction: column;
  gap: 5px;

  > div {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 5px 9px;
    background: var(--surface-2);
    border-radius: 6px;
    font-size: 10.5px;
  }

  b { color: var(--primary); }
}

.pv-now {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;

  li {
    display: flex;
    align-items: center;
    gap: 7px;
    font-size: 10.5px;
  }

  i {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--primary);
    box-shadow: 0 0 5px rgba(var(--primary-rgb), 0.6);
  }
}

.pv-pills {
  display: flex;
  gap: 6px;

  span {
    font-size: 10.5px;
    font-weight: 600;
    padding: 4px 12px;
    border: 1px solid var(--border);
    background: var(--surface-2);
  }
}

.pv-contact {
  display: flex;
  align-items: center;
  gap: 10px;

  > div { flex: 1; }
  b { display: block; font-size: 12px; }
  span { font-size: 10px; color: var(--text-2); }

  i {
    font-style: normal;
    font-size: 10.5px;
    font-weight: 700;
    color: #fff;
    padding: 5px 12px;
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    border-radius: 6px;
  }
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
