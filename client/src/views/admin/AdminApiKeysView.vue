<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type ApiKeyInfo } from '../../api';

const { t } = useI18n();

const keys = ref<ApiKeyInfo[]>([]);
const name = ref('');
const freshKey = ref('');
const busy = ref(false);

async function load(): Promise<void> {
  keys.value = await adminApi.apiKeys();
}

async function create(): Promise<void> {
  if (busy.value || !name.value.trim()) return;
  busy.value = true;
  try {
    const created = await adminApi.createApiKey(name.value.trim());
    freshKey.value = created.key;
    name.value = '';
    await load();
  } finally {
    busy.value = false;
  }
}

async function revoke(key: ApiKeyInfo): Promise<void> {
  if (!window.confirm(t('admin.confirmRevokeKey', { name: key.name }))) return;
  await adminApi.deleteApiKey(key.id);
  await load();
}

async function copyKey(): Promise<void> {
  await navigator.clipboard.writeText(freshKey.value);
}

onMounted(load);
</script>

<template>
  <div>
    <h1 class="page-h">{{ t('admin.menuApi') }}</h1>
    <p class="hint">{{ t('admin.apiHint') }}</p>

    <!-- 创建 -->
    <div class="create-bar">
      <input v-model="name" type="text" :placeholder="t('admin.keyNamePlaceholder')" @keyup.enter="create" />
      <button class="btn primary" :disabled="busy || !name.trim()" @click="create">
        {{ t('admin.createKey') }}
      </button>
    </div>

    <!-- 新 Key 一次性展示 -->
    <div v-if="freshKey" class="fresh">
      <p>{{ t('admin.keyOnce') }}</p>
      <code>{{ freshKey }}</code>
      <button class="op" @click="copyKey">{{ t('admin.copy') }}</button>
    </div>

    <!-- 列表 -->
    <table class="table">
      <thead>
        <tr>
          <th>{{ t('admin.keyName') }}</th>
          <th>{{ t('admin.keyPrefix') }}</th>
          <th>{{ t('admin.keyLastUsed') }}</th>
          <th>{{ t('admin.colDate') }}</th>
          <th />
        </tr>
      </thead>
      <tbody>
        <tr v-for="key in keys" :key="key.id">
          <td>{{ key.name }}</td>
          <td><code>{{ key.prefix }}…</code></td>
          <td>{{ key.lastUsedAt ?? t('admin.neverUsed') }}</td>
          <td>{{ key.createdAt.slice(0, 10) }}</td>
          <td class="ops">
            <button class="op danger" @click="revoke(key)">{{ t('admin.revoke') }}</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- 使用说明 -->
    <div class="usage">
      <h2>{{ t('admin.apiUsage') }}</h2>
      <pre>POST /api/v1/ext/posts
X-Api-Key: myk_xxx
Content-Type: application/json

{"slug":"my-post","title":"标题","excerpt":"摘要","contentMd":"# Markdown 正文","tags":["AI"],"status":"draft"}

POST /api/v1/ext/notes
X-Api-Key: myk_xxx

{"contentMd":"一条随想","mood":"AI","images":[]}</pre>
    </div>
  </div>
</template>

<style scoped lang="scss">
.page-h { font-size: 26px; margin-bottom: 8px; }
.hint { font-size: 13px; color: var(--text-2); margin-bottom: 22px; }

.create-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;

  input {
    flex: 1;
    max-width: 340px;
    padding: 10px 14px;
    border-radius: 10px;
    border: 1px solid var(--border);
    background: var(--surface);
    color: var(--text);
    font-size: 14px;
    font-family: inherit;
    outline: none;

    &:focus { border-color: var(--primary); box-shadow: 0 0 0 3px rgba(var(--primary-rgb), 0.12); }
  }
}

.btn.primary {
  padding: 10px 24px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 700;
  color: #fff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
  background: linear-gradient(180deg, var(--primary), var(--primary-deep));
  box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.4);

  &:hover:not(:disabled) { filter: brightness(1.08); }
  &:disabled { opacity: 0.55; }
}

.fresh {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  padding: 14px 18px;
  margin-bottom: 18px;
  border: 1px solid rgba(var(--primary-rgb), 0.4);
  background: rgba(var(--primary-rgb), 0.06);
  border-radius: var(--radius);

  p { font-size: 13px; font-weight: 600; color: var(--primary); }

  code {
    font-family: Consolas, monospace;
    font-size: 13px;
    padding: 4px 10px;
    background: var(--surface);
    border-radius: 6px;
    word-break: break-all;
  }
}

.table {
  width: 100%;
  border-collapse: collapse;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  margin-bottom: 26px;

  th, td {
    padding: 12px 16px;
    text-align: left;
    font-size: 14px;
    border-bottom: 1px solid var(--border);
  }

  th { background: var(--surface-2); font-size: 12px; color: var(--text-2); }
  code { font-family: Consolas, monospace; font-size: 13px; }
}

.ops { text-align: right; }

.op {
  border: none;
  background: none;
  font-size: 13px;
  font-weight: 600;
  color: var(--primary);

  &.danger { color: var(--accent-red); }
  &:hover { opacity: 0.75; }
}

.usage {
  h2 { font-size: 17px; margin-bottom: 10px; }

  pre {
    padding: 16px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    font-family: Consolas, monospace;
    font-size: 13px;
    line-height: 1.7;
    overflow-x: auto;
  }
}
</style>
