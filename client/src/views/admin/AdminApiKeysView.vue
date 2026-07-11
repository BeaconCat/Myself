<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type ApiKeyInfo } from '../../api';
import { readToken } from '../../stores/auth';
import { useDialogStore } from '../../stores/dialog';
import { API_CATALOG, buildAgentPrompt, type Endpoint } from './apiCatalog';

const { t } = useI18n();

/* ===== Key 管理 ===== */
const keys = ref<ApiKeyInfo[]>([]);
const name = ref('');
const busy = ref(false);
/** 本会话内创建的 Key 明文（服务端只存哈希，仅此处可用于调试/提示词） */
const sessionKeys = ref<{ name: string; key: string }[]>([]);

async function load(): Promise<void> {
  keys.value = await adminApi.apiKeys();
}

async function create(): Promise<void> {
  if (busy.value || !name.value.trim()) return;
  busy.value = true;
  try {
    const created = await adminApi.createApiKey(name.value.trim());
    sessionKeys.value.unshift({ name: created.name, key: created.key });
    selectedKey.value = created.key;
    name.value = '';
    await load();
  } finally {
    busy.value = false;
  }
}

async function revoke(key: ApiKeyInfo): Promise<void> {
  const ok = await useDialogStore().confirm({
    title: t('admin.revoke'),
    message: t('admin.confirmRevokeKey', { name: key.name }),
    danger: true,
  });
  if (!ok) return;
  await adminApi.deleteApiKey(key.id);
  await load();
}

/* ===== 调试台 ===== */
const openId = ref('');
const reqPath = ref('');
const reqBody = ref('');
const respText = ref('');
const respStatus = ref<number | null>(null);
const testing = ref(false);
const selectedKey = ref('');
const manualKey = ref('');

const effectiveKey = computed(() => manualKey.value.trim() || selectedKey.value);

function endpointId(e: Endpoint): string {
  return `${e.method} ${e.path}`;
}

function toggle(e: Endpoint): void {
  const id = endpointId(e);
  if (openId.value === id) {
    openId.value = '';
    return;
  }
  openId.value = id;
  let path = e.path;
  for (const [param, value] of Object.entries(e.sample ?? {})) {
    path = path.replace(param, value);
  }
  reqPath.value = path;
  reqBody.value = e.sampleBody ?? '';
  respText.value = '';
  respStatus.value = null;
}

async function send(e: Endpoint): Promise<void> {
  if (testing.value) return;
  testing.value = true;
  respText.value = '';
  respStatus.value = null;
  try {
    const headers: Record<string, string> = {};
    if (e.auth === 'jwt') headers.Authorization = `Bearer ${readToken()}`;
    if (e.auth === 'apikey') headers['X-Api-Key'] = effectiveKey.value;
    const hasBody = ['POST', 'PUT'].includes(e.method) && reqBody.value.trim();
    if (hasBody) headers['Content-Type'] = 'application/json';
    const res = await fetch(reqPath.value, {
      method: e.method,
      headers,
      body: hasBody ? reqBody.value : undefined,
    });
    respStatus.value = res.status;
    const text = await res.text();
    try {
      respText.value = JSON.stringify(JSON.parse(text), null, 2);
    } catch {
      respText.value = text.slice(0, 4000);
    }
  } catch (err) {
    respText.value = String(err);
  } finally {
    testing.value = false;
  }
}

/* ===== Agent 提示词 ===== */
const prompt = computed(() =>
  buildAgentPrompt(window.location.origin, effectiveKey.value),
);

async function copyText(text: string): Promise<void> {
  await navigator.clipboard.writeText(text);
}

async function copyPrompt(): Promise<void> {
  await copyText(prompt.value);
}

function downloadPrompt(): void {
  const blob = new Blob([prompt.value], { type: 'text/plain;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'Myself Prompt.txt';
  a.click();
  URL.revokeObjectURL(url);
}

onMounted(load);
</script>

<template>
  <div>
    <header class="a-head">
      <div>
        <h1>{{ t('admin.menuApi') }}</h1>
        <p>{{ t('admin.apiHint') }}</p>
      </div>
    </header>

    <!-- Key 管理 -->
    <section class="a-card block">
      <h2 class="sec-h">{{ t('admin.keySection') }}</h2>
      <div class="create-bar">
        <input v-model="name" class="a-input" type="text" :placeholder="t('admin.keyNamePlaceholder')" @keyup.enter="create" />
        <button class="a-btn primary" :disabled="busy || !name.trim()" @click="create">{{ t('admin.createKey') }}</button>
      </div>

      <div v-for="sk in sessionKeys" :key="sk.key" class="fresh">
        <span class="fresh-name">{{ sk.name }}</span>
        <code>{{ sk.key }}</code>
        <button class="op" @click="copyText(sk.key)">{{ t('admin.copy') }}</button>
      </div>

      <table class="k-table">
        <thead>
          <tr><th>{{ t('admin.keyName') }}</th><th>{{ t('admin.keyPrefix') }}</th><th>{{ t('admin.keyLastUsed') }}</th><th /></tr>
        </thead>
        <tbody>
          <tr v-for="key in keys" :key="key.id">
            <td>{{ key.name }}</td>
            <td><code>{{ key.prefix }}…</code></td>
            <td>{{ key.lastUsedAt ?? t('admin.neverUsed') }}</td>
            <td class="ops"><button class="op danger" @click="revoke(key)">{{ t('admin.revoke') }}</button></td>
          </tr>
        </tbody>
      </table>
    </section>

    <!-- 调试身份 -->
    <section class="a-card block">
      <h2 class="sec-h">{{ t('admin.testerAuth') }}</h2>
      <div class="auth-bar">
        <label>
          <span>{{ t('admin.pickSessionKey') }}</span>
          <select v-model="selectedKey" class="a-input">
            <option value="">{{ t('admin.noSessionKey') }}</option>
            <option v-for="sk in sessionKeys" :key="sk.key" :value="sk.key">{{ sk.name }}</option>
          </select>
        </label>
        <label>
          <span>{{ t('admin.manualKey') }}</span>
          <input v-model="manualKey" class="a-input" type="text" placeholder="myk_…" />
        </label>
      </div>
      <p class="tip">{{ t('admin.testerTip') }}</p>
    </section>

    <!-- 接口目录 + 调试台 -->
    <section v-for="group in API_CATALOG" :key="group.title" class="block">
      <h2 class="grp-h">{{ group.title }}</h2>
      <div class="ep-list">
        <div v-for="e in group.endpoints" :key="endpointId(e)" class="ep a-card" :class="{ open: openId === endpointId(e) }">
          <button class="ep-row" @click="toggle(e)">
            <span class="method" :class="e.method.toLowerCase()">{{ e.method }}</span>
            <code class="path">{{ e.path }}</code>
            <span class="desc">{{ e.desc }}</span>
            <span class="auth-tag" :class="e.auth">{{ e.auth === 'none' ? t('admin.authNone') : e.auth === 'jwt' ? 'JWT' : 'APIKey' }}</span>
          </button>

          <div v-if="openId === endpointId(e)" class="tester">
            <label class="t-row">
              <span>URL</span>
              <input v-model="reqPath" class="a-input" type="text" />
            </label>
            <label v-if="['POST', 'PUT'].includes(e.method)" class="t-row">
              <span>Body</span>
              <textarea v-model="reqBody" class="a-input mono" rows="6" />
            </label>
            <div class="t-actions">
              <button class="a-btn primary" :disabled="testing" @click="send(e)">
                {{ testing ? t('admin.sending') : t('admin.sendRequest') }}
              </button>
              <span v-if="respStatus !== null" class="status" :class="{ ok: respStatus < 400 }">HTTP {{ respStatus }}</span>
            </div>
            <pre v-if="respText" class="resp">{{ respText }}</pre>
          </div>
        </div>
      </div>
    </section>

    <!-- Agent 提示词 -->
    <section class="a-card block">
      <h2 class="sec-h">{{ t('admin.promptTitle') }}</h2>
      <p class="tip">{{ t('admin.promptHint') }}</p>
      <pre class="prompt">{{ prompt }}</pre>
      <div class="t-actions">
        <button class="a-btn ghost" @click="copyPrompt">{{ t('admin.copy') }}</button>
        <button class="a-btn primary" @click="downloadPrompt">{{ t('admin.downloadPrompt') }}</button>
      </div>
    </section>
  </div>
</template>

<style scoped lang="scss">
.block { margin-bottom: 22px; }

.sec-h {
  font-size: 16px;
  margin-bottom: 14px;
  padding-left: 12px;
  border-left: 4px solid var(--primary);
}

.grp-h {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-2);
  margin: 26px 0 10px;
}

.create-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;

  input { flex: 1; max-width: 340px; }
}

.fresh {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  padding: 10px 14px;
  margin-bottom: 10px;
  border: 1px solid rgba(var(--primary-rgb), 0.4);
  background: rgba(var(--primary-rgb), 0.06);
  border-radius: 10px;

  .fresh-name { font-size: 13px; font-weight: 700; color: var(--primary); }

  code {
    font-family: Consolas, monospace;
    font-size: 12.5px;
    word-break: break-all;
  }
}

.k-table {
  width: 100%;
  border-collapse: collapse;

  th, td {
    padding: 10px 12px;
    text-align: left;
    font-size: 13.5px;
    border-bottom: 1px solid var(--border);
  }

  th { font-size: 12px; color: var(--text-2); }
  code { font-family: Consolas, monospace; font-size: 12.5px; }
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

.auth-bar {
  display: flex;
  gap: 14px;
  flex-wrap: wrap;

  label {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 240px;

    span { font-size: 12px; font-weight: 600; color: var(--text-2); }
  }
}

.tip { font-size: 12.5px; color: var(--text-2); margin-top: 10px; }

/* 接口条目 */
.ep-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ep {
  padding: 0;
  overflow: hidden;
  transition: border-color var(--dur-fast);

  &.open { border-color: rgba(var(--primary-rgb), 0.45); }
}

.ep-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px 16px;
  border: none;
  background: none;
  text-align: left;
  cursor: pointer;
  transition: background var(--dur-fast);

  &:hover { background: rgba(var(--primary-rgb), 0.04); }
}

.method {
  font-size: 11px;
  font-weight: 800;
  padding: 3px 10px;
  border-radius: 6px;
  width: 58px;
  text-align: center;
  flex-shrink: 0;

  &.get { background: rgba(0, 120, 255, 0.14); color: var(--accent-blue); }
  &.post { background: rgba(0, 200, 83, 0.14); color: #00a344; }
  &.put { background: rgba(255, 179, 0, 0.16); color: #c78800; }
  &.delete { background: rgba(255, 0, 50, 0.12); color: var(--accent-red); }
}

.path {
  font-family: Consolas, monospace;
  font-size: 12.5px;
  color: var(--text);
  flex-shrink: 0;
}

.desc {
  font-size: 12.5px;
  color: var(--text-2);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.auth-tag {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 999px;
  flex-shrink: 0;

  &.none { background: var(--surface-2); color: var(--text-2); }
  &.jwt { background: rgba(0, 120, 255, 0.12); color: var(--accent-blue); }
  &.apikey { background: rgba(var(--primary-rgb), 0.12); color: var(--primary); }
}

/* 调试台 */
.tester {
  padding: 14px 16px 16px;
  border-top: 1px dashed var(--border);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.t-row {
  display: flex;
  flex-direction: column;
  gap: 6px;

  span { font-size: 11px; font-weight: 700; color: var(--text-2); }
}

.mono { font-family: Consolas, monospace; font-size: 13px; }

.t-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 10px;
}

.status {
  font-size: 13px;
  font-weight: 700;
  color: var(--accent-red);

  &.ok { color: #00a344; }
}

.resp, .prompt {
  padding: 14px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 10px;
  font-family: Consolas, monospace;
  font-size: 12.5px;
  line-height: 1.7;
  max-height: 320px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
}

.prompt { max-height: 420px; margin-top: 10px; }
</style>
