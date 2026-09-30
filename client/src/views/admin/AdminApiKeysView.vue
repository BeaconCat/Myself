<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type ApiKeyInfo, type ApiKeyScope, type ApiLogEntry } from '../../api';
import { useDialogStore } from '../../stores/dialog';
import { API_CATALOG, buildAgentPrompt, type Endpoint } from './apiCatalog';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import StSeg from './studio/StSeg.vue';
import StModal from './studio/StModal.vue';
import PopMenu from './studio/PopMenu.vue';
import EmptyArt from './studio/EmptyArt.vue';
import StPager from './studio/StPager.vue';
import { KeyRound, ScrollText } from 'lucide';
import { copyText, saveBlob } from './studio/state';
import { toast } from './studio/toast';
import { dateText, parseTime, relTime } from './studio/format';
import type { MenuItem } from './studio/types';

/** API 中心：Key 管理（明文只显示一次）、快速上手、接口目录与调试台、Agent 提示词、调用日志 */
const { t } = useI18n();
const dialog = useDialogStore();

const keys = ref<ApiKeyInfo[]>([]);
const loaded = ref(false);
/** 本会话创建的明文 Key（服务端只存哈希）：用于调试台与提示词 */
const sessionKeys = ref<{ id: number; name: string; key: string }[]>([]);

async function load(): Promise<void> {
  try {
    keys.value = await adminApi.apiKeys();
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
  loaded.value = true;
}

/* ===== 新建 ===== */
const createOpen = ref(false);
const newName = ref('');
/** 新建 Key 的权限：默认仅投稿（最小权限） */
const newScope = ref<ApiKeyScope>('contrib');
const created = ref<{ name: string; key: string } | null>(null);
const creating = ref(false);

function openCreate(): void {
  newName.value = '';
  newScope.value = 'contrib';
  created.value = null;
  createOpen.value = true;
}

async function create(): Promise<void> {
  if (creating.value || !newName.value.trim()) return;
  creating.value = true;
  try {
    const res = await adminApi.createApiKey(newName.value.trim(), newScope.value);
    created.value = { name: res.name, key: res.key };
    sessionKeys.value.unshift({ id: res.id, name: res.name, key: res.key });
    testerKey.value = res.key;
    await load();
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  } finally {
    creating.value = false;
  }
}

async function copy(text: string, msg = t('studio.copied')): Promise<void> {
  if (await copyText(text)) toast(msg, { icon: 'copy' });
}

async function revoke(k: ApiKeyInfo): Promise<void> {
  const ok = await dialog.confirm({
    title: t('studio.api.revokeTitle', { name: k.name }),
    message: t('studio.api.revokeBody'),
    confirmText: t('studio.api.revoke'),
    danger: true,
  });
  if (!ok) return;
  try {
    await adminApi.deleteApiKey(k.id);
    sessionKeys.value = sessionKeys.value.filter((s) => s.id !== k.id);
    keys.value = keys.value.filter((x) => x.id !== k.id);
    toast(t('studio.api.revoked'), { icon: 'key' });
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

function keyMenu(k: ApiKeyInfo): MenuItem[] {
  return [
    { icon: 'copy', label: t('studio.api.copyPrefix'), run: () => void copy(k.prefix) },
    { icon: 'trash', label: t('studio.api.revoke'), danger: true, divider: true, run: () => void revoke(k) },
  ];
}

/* ===== 快速上手 ===== */
const lang = ref<'curl' | 'js'>('curl');
const origin = window.location.origin;
const sampleKey = computed(() => sessionKeys.value[0]?.key ?? '');

/**
 * 页面上只显示打码后的 Key（前 8 位 + 圆点）：完整 Key 只在「只会显示这一次」的弹窗里出现，
 * 关掉后示例代码与提示词都不再明文展示；点「复制」时仍复制带完整 Key 的版本，方便直接粘贴使用。
 */
function maskKey(key: string): string {
  return key ? `${key.slice(0, 8)}••••••••` : '';
}

function codeFor(key: string): string {
  const k = key || '<YOUR_API_KEY>';
  return lang.value === 'curl'
    ? [
      `# ${t('studio.api.codeCurl')}`,
      `curl -X POST ${origin}/api/v1/ext/posts \\`,
      `  -H "X-Api-Key: ${k}" \\`,
      '  -H "Content-Type: application/json" \\',
      `  -d '{"slug":"weekly-notes","title":"${t('studio.api.codeTitle')}","contentMd":"## ${t('studio.api.codeBody')}","tags":["API"],"status":"draft"}'`,
    ].join('\n')
    : [
      `// ${t('studio.api.codeJs')}`,
      `await fetch("${origin}/api/v1/ext/notes", {`,
      '  method: "POST",',
      `  headers: { "X-Api-Key": process.env.MYSELF_KEY, "Content-Type": "application/json" },`,
      `  body: JSON.stringify({ contentMd: "${t('studio.api.codeNote')}", mood: "AI", images: [] }),`,
      '});',
    ].join('\n');
}
const code = computed(() => codeFor(maskKey(sampleKey.value)));

/* ===== 调试台 ===== */
const openId = ref('');
const reqPath = ref('');
const reqBody = ref('');
const resp = ref<{ status: number; text: string; ms: number } | null>(null);
const testing = ref(false);
const testerKey = ref('');
const manualKey = ref('');
const effectiveKey = computed(() => manualKey.value.trim() || testerKey.value);
const catGroup = ref(1);

const epId = (e: Endpoint) => `${e.method} ${e.path}`;

function toggle(e: Endpoint): void {
  const id = epId(e);
  if (openId.value === id) {
    openId.value = '';
    return;
  }
  openId.value = id;
  let path = e.path;
  for (const [k, v] of Object.entries(e.sample ?? {})) path = path.replace(k, v);
  reqPath.value = path;
  reqBody.value = e.sampleBody ?? '';
  resp.value = null;
}

async function send(e: Endpoint): Promise<void> {
  if (testing.value) return;
  testing.value = true;
  resp.value = null;
  const t0 = performance.now();
  try {
    const headers: Record<string, string> = {};
    // 管理接口凭当前登录的会话 Cookie（同源自动携带）+ CSRF 头
    if (e.auth === 'jwt') headers['X-Requested-With'] = 'myself';
    if (e.auth === 'apikey') headers['X-Api-Key'] = effectiveKey.value;
    const hasBody = ['POST', 'PUT'].includes(e.method) && reqBody.value.trim();
    if (hasBody) headers['Content-Type'] = 'application/json';
    const res = await fetch(reqPath.value, { method: e.method, headers, body: hasBody ? reqBody.value : undefined });
    const raw = await res.text();
    let text = raw.slice(0, 6000);
    try {
      text = JSON.stringify(JSON.parse(raw), null, 2);
    } catch { /* 非 JSON 原样显示 */ }
    resp.value = { status: res.status, text, ms: Math.round(performance.now() - t0) };
  } catch (err) {
    resp.value = { status: 0, text: String(err), ms: Math.round(performance.now() - t0) };
  } finally {
    testing.value = false;
  }
}

/* ===== 提示词 ===== */
/** 显示用（打码）与复制 / 下载用（完整 Key）两份提示词 */
const prompt = computed(() => buildAgentPrompt(origin, maskKey(effectiveKey.value)));
const promptFull = computed(() => buildAgentPrompt(origin, effectiveKey.value));
function downloadPrompt(): void {
  saveBlob(new Blob([promptFull.value], { type: 'text/plain;charset=utf-8' }), 'Myself Prompt.txt');
}

/* ===== 调用日志（服务端分页；旧后端没有该接口时显示不可用） ===== */
type LogStatus = 'all' | 'ok' | 'error';
const logs = ref<ApiLogEntry[]>([]);
const logPage = ref(1);
const logSize = ref(20);
const logTotal = ref(0);
const logCounts = ref({ all: 0, ok: 0, error: 0 });
/** 0 = 全部 Key */
const logKey = ref(0);
const logStatus = ref<LogStatus>('all');
const logState = ref<'loading' | 'ready' | 'unavailable'>('loading');
const logBusy = ref(false);
const logSec = ref<HTMLElement | null>(null);
let logSeq = 0;

async function loadLogs(): Promise<void> {
  const seq = ++logSeq;
  logBusy.value = true;
  try {
    const res = await adminApi.apiLogs({
      page: logPage.value,
      pageSize: logSize.value,
      key: logKey.value || undefined,
      status: logStatus.value === 'all' ? undefined : logStatus.value,
    });
    if (seq !== logSeq) return;
    logs.value = res.items ?? [];
    logTotal.value = res.total ?? 0;
    logCounts.value = res.counts ?? { all: 0, ok: 0, error: 0 };
    logState.value = 'ready';
  } catch {
    if (seq !== logSeq) return;
    if (logState.value === 'ready') toast(t('studio.loadFailed'), { icon: 'x' });
    else logState.value = 'unavailable';
  } finally {
    if (seq === logSeq) logBusy.value = false;
  }
}

// 换筛选回到第 1 页；翻页 / 换每页条数直接重取
watch([logKey, logStatus, logPage, logSize], (n, o) => {
  if ((n[0] !== o[0] || n[1] !== o[1]) && logPage.value !== 1) {
    logPage.value = 1;
    return;
  }
  void loadLogs();
});

const LOG_STATUS = computed(() =>
  (['all', 'ok', 'error'] as const).map((v) => ({
    value: v,
    label: t(`studio.api.${{ all: 'lAll', ok: 'lOk', error: 'lErr' }[v]}`),
    count: logState.value === 'ready' ? logCounts.value[v] || undefined : undefined,
  })),
);

/** 从 Key 卡片跳到它的调用记录 */
function showKeyLogs(k: ApiKeyInfo): void {
  logKey.value = k.id;
  logStatus.value = 'all';
  logSec.value?.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

const pad2 = (n: number) => String(n).padStart(2, '0');
/** 日志时间：本地 MM-DD HH:mm:ss（非今年带年份） */
function logTime(s: string): string {
  const d = parseTime(s);
  if (!d) return s;
  const md = `${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}`;
  const day = d.getFullYear() === new Date().getFullYear() ? md : `${d.getFullYear()}-${md}`;
  return `${day} ${pad2(d.getHours())}:${pad2(d.getMinutes())}:${pad2(d.getSeconds())}`;
}

/** 状态码配色：2xx/3xx 绿；认证失败与 5xx 红；其余 4xx（参数、未找到、冲突）黄 */
function statusTone(code: number): string {
  if (code < 400) return '';
  return code >= 500 || code === 401 || code === 403 ? 'bad' : 'warn';
}

/** 仍存在的 Key：日志里其余的 Key 标为已吊销 */
const liveKeys = computed(() => new Set(keys.value.map((k) => k.id)));

onMounted(() => {
  void load();
  void loadLogs();
});
</script>

<template>
  <section class="studio view">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.api.title') }}</h1>
        <p>{{ t('studio.api.descPre') }}<span class="mono hk">X-Api-Key</span>{{ t('studio.api.descPost') }}</p>
      </div>
      <div class="act">
        <button type="button" class="st-btn p" @click="openCreate"><SIcon name="plus" :size="18" />{{ t('studio.api.new') }}</button>
      </div>
    </div>

    <div v-if="loaded && !keys.length" class="st-empty keys-empty">
      <EmptyArt :icon="KeyRound" />
      <h4>{{ t('studio.api.empty') }}</h4>
      <p>{{ t('studio.api.emptySub') }}</p>
    </div>

    <div class="keys">
      <div v-for="(k, i) in keys" :key="k.id" class="key st-rise" :style="{ '--i': i }">
        <span class="ki"><SIcon name="key" :size="20" /></span>
        <div class="kn">
          <h4>{{ k.name }}</h4>
          <div class="kv2">
            <span class="mono">{{ k.prefix }}…</span>
            <button type="button" :title="t('studio.api.copyPrefix')" @click="copy(k.prefix)"><SIcon name="copy" :size="16" /></button>
            <span v-if="sessionKeys.some((s) => s.id === k.id)" class="fresh">{{ t('studio.api.thisSession') }}</span>
          </div>
          <div class="scopes" :class="k.scope">
            <b>{{ k.scope === 'full' ? t('studio.api.scopeFull') : t('studio.api.scopeContrib') }}</b>
            <template v-if="k.scope === 'full'"><span>{{ t('studio.api.scopePosts') }}</span><span>{{ t('studio.api.scopeNotes') }}</span></template>
            <span v-else>{{ t('studio.api.scopeDrafts') }}</span>
          </div>
        </div>
        <div class="lu">
          {{ k.lastUsedAt ? relTime(k.lastUsedAt) : t('studio.api.never') }}
          <small>{{ t('studio.api.lastUsed') }}</small>
          <button v-if="k.calls" type="button" class="calls" @click="showKeyLogs(k)">
            {{ t('studio.api.calls', { n: k.calls }) }}<span v-if="k.errors" class="bad"> · {{ t('studio.api.callsErr', { n: k.errors }) }}</span>
          </button>
        </div>
        <div class="lu">
          <span class="mono">{{ dateText(k.createdAt) }}</span>
          <small>{{ t('studio.api.createdAt') }}</small>
        </div>
        <PopMenu :items="keyMenu(k)" :label="t('studio.a11y.moreOf', { name: k.name })" />
      </div>
    </div>

    <div class="api-grid">
      <div>
        <div class="st-sec-t">
          <h2>{{ t('studio.api.quick') }}</h2>
          <StSeg v-model="lang" :options="[{ value: 'curl', label: 'cURL' }, { value: 'js', label: 'JavaScript' }]" />
        </div>
        <pre class="code"><code>{{ code }}</code><button type="button" class="st-ibtn cp" :title="t('studio.copy')" @click="copy(codeFor(sampleKey))"><SIcon name="copy" :size="18" /></button></pre>
      </div>
      <div>
        <div class="st-sec-t"><h2>{{ t('studio.api.endpoints') }}</h2></div>
        <div class="endp">
          <div v-for="e in API_CATALOG[1].endpoints" :key="epId(e)"><em :class="e.method.toLowerCase()">{{ e.method }}</em><span>{{ e.path.split('?')[0] }}</span></div>
        </div>
        <p class="note">{{ t('studio.api.draftNote') }}</p>
      </div>
    </div>

    <!-- 调试台 -->
    <div class="tester">
      <div class="st-sec-t">
        <h2>{{ t('studio.api.tester') }}</h2>
        <StSeg v-model="catGroup" :options="API_CATALOG.map((_g, i) => ({ value: i, label: t(`studio.api.group${i}`) }))" />
      </div>
      <div class="auth">
        <label class="st-field sel">
          <SIcon name="key" :size="18" />
          <select v-model="testerKey" :aria-label="t('studio.a11y.testerKey')">
            <option value="">{{ t('studio.api.noSessionKey') }}</option>
            <option v-for="s in sessionKeys" :key="s.id" :value="s.key">{{ s.name }}</option>
          </select>
        </label>
        <label class="st-field mono-in"><input v-model="manualKey" :placeholder="t('studio.api.manualKey')" :aria-label="t('studio.api.manualKey')" spellcheck="false" /></label>
        <small>{{ t('studio.api.testerTip') }}</small>
      </div>
      <div class="eps">
        <div v-for="e in API_CATALOG[catGroup].endpoints" :key="epId(e)" class="ep" :class="{ open: openId === epId(e) }">
          <button type="button" class="ep-row" @click="toggle(e)">
            <em :class="e.method.toLowerCase()">{{ e.method }}</em>
            <code>{{ e.path }}</code>
            <span class="d">{{ e.desc }}</span>
            <span class="au" :class="e.auth">{{ e.auth === 'none' ? t('studio.api.authNone') : e.auth === 'jwt' ? 'JWT' : 'APIKey' }}</span>
            <SIcon name="chevronD" :size="16" class="chev" />
          </button>
          <div v-if="openId === epId(e)" class="ep-body">
            <label class="st-field mono-in"><span class="suffix">URL</span><input v-model="reqPath" spellcheck="false" /></label>
            <label v-if="['POST', 'PUT'].includes(e.method)" class="st-field ta mono-in"><textarea v-model="reqBody" rows="6" :aria-label="t('studio.a11y.reqBody')" spellcheck="false" /></label>
            <div class="send">
              <button type="button" class="st-btn p sm" :disabled="testing" @click="send(e)"><SIcon name="play" :size="16" />{{ testing ? t('studio.api.sending') : t('studio.api.send') }}</button>
              <span v-if="resp" class="status" :class="{ ok: resp.status > 0 && resp.status < 400 }"><i class="st-dot" />HTTP {{ resp.status }} · {{ resp.ms }} ms</span>
            </div>
            <pre v-if="resp" class="code resp"><code>{{ resp.text }}</code></pre>
          </div>
        </div>
      </div>
    </div>

    <!-- Agent 提示词 -->
    <div class="prompt-sec">
      <div class="st-sec-t">
        <h2>{{ t('studio.api.prompt') }}</h2>
        <span class="acts">
          <button type="button" class="st-btn g sm" @click="copy(promptFull)"><SIcon name="copy" :size="16" />{{ t('studio.copy') }}</button>
          <button type="button" class="st-btn g sm" @click="downloadPrompt"><SIcon name="download" :size="16" />{{ t('studio.api.download') }}</button>
        </span>
      </div>
      <p class="note">{{ t('studio.api.promptHint') }}</p>
      <pre class="prompt"><code>{{ prompt }}</code></pre>
    </div>

    <!-- 调用日志 -->
    <div ref="logSec" class="log-sec">
      <div class="st-sec-t">
        <h2>{{ t('studio.api.logs') }}</h2>
        <span v-if="logState !== 'unavailable'" class="log-tools">
          <label class="st-field sel key-pick">
            <SIcon name="key" :size="16" />
            <select v-model.number="logKey" :aria-label="t('studio.api.lAllKeys')">
              <option :value="0">{{ t('studio.api.lAllKeys') }}</option>
              <option v-for="k in keys" :key="k.id" :value="k.id">{{ k.name }}</option>
            </select>
            <SIcon name="chevronD" :size="14" class="chev" />
          </label>
          <StSeg v-model="logStatus" :options="LOG_STATUS" />
          <button type="button" class="st-ibtn" :class="{ spin: logBusy }" :title="t('studio.api.lRefresh')" :aria-label="t('studio.api.lRefresh')" @click="loadLogs">
            <SIcon name="refresh" :size="17" />
          </button>
        </span>
      </div>
      <p class="note log-sub">{{ t('studio.api.logsSub') }}</p>

      <div v-if="logState === 'unavailable'" class="st-empty">
        <EmptyArt :icon="ScrollText" />
        <h4>{{ t('studio.api.lUnavailable') }}</h4>
        <p>{{ t('studio.api.lUnavailableSub') }}</p>
      </div>
      <div v-else-if="logState === 'ready' && !logs.length" class="st-empty">
        <EmptyArt :icon="ScrollText" />
        <h4>{{ logKey || logStatus !== 'all' ? t('studio.api.lNoMatch') : t('studio.api.lEmpty') }}</h4>
        <p>{{ logKey || logStatus !== 'all' ? t('studio.api.lNoMatchSub') : t('studio.api.lEmptySub') }}</p>
      </div>
      <template v-else-if="logState === 'ready'">
        <table class="st-table logs" :class="{ busy: logBusy }">
          <thead>
            <tr>
              <th>{{ t('studio.api.lTime') }}</th>
              <th>Key</th>
              <th>{{ t('studio.api.lReq') }}</th>
              <th>{{ t('studio.api.lStatus') }}</th>
              <th class="r">{{ t('studio.api.lMs') }}</th>
              <th>{{ t('studio.api.lFrom') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(l, i) in logs" :key="l.id" :style="{ '--i': i }">
              <td class="mono dim nw">{{ logTime(l.at) }}</td>
              <td class="kc">
                <template v-if="l.keyId !== null">
                  <b>{{ l.keyName }}</b>
                  <span v-if="!liveKeys.has(l.keyId)" class="tag">{{ t('studio.api.lRevoked') }}</span>
                </template>
                <span v-else class="tag warn">{{ l.keyPrefix ? t('studio.api.lInvalid') : t('studio.api.lMissing') }}</span>
                <small v-if="l.keyPrefix" class="mono">{{ l.keyPrefix }}…</small>
              </td>
              <td>
                <span class="req">
                  <em :class="l.method.toLowerCase()">{{ l.method }}</em>
                  <span class="mono p" :title="l.path">{{ l.path }}</span>
                </span>
              </td>
              <td><span class="sc" :class="statusTone(l.status)"><i class="st-dot" />{{ l.status }}</span></td>
              <td class="mono dim r nw">{{ l.ms }} ms</td>
              <td class="mono dim ip" :title="l.ua">{{ l.ip }}</td>
            </tr>
          </tbody>
        </table>
        <StPager v-model:page="logPage" v-model:page-size="logSize" :total="logTotal" :page-sizes="[20, 50, 100]" :busy="logBusy" />
      </template>
    </div>

    <StModal :open="createOpen" @close="createOpen = false">
      <template v-if="!created">
        <div class="mi"><SIcon name="key" :size="22" /></div>
        <h3>{{ t('studio.api.newTitle') }}</h3>
        <p>{{ t('studio.api.newDesc') }}</p>
        <div class="st-flabel">{{ t('studio.api.name') }}</div>
        <label class="st-field"><input v-model="newName" :placeholder="t('studio.api.namePh')" :aria-label="t('studio.a11y.keyName')" @keydown.enter="create" /></label>
        <div class="st-flabel scope-label">{{ t('studio.api.scopeLabel') }}</div>
        <div class="scope-pick" role="radiogroup">
          <button
            v-for="sc in (['contrib', 'full'] as const)"
            :key="sc"
            type="button"
            role="radio"
            :aria-checked="newScope === sc"
            :class="{ on: newScope === sc }"
            @click="newScope = sc"
          >
            <SIcon :name="sc === 'full' ? 'key' : 'pen'" :size="18" />
            <b>{{ sc === 'full' ? t('studio.api.scopeFull') : t('studio.api.scopeContrib') }}</b>
            <small>{{ sc === 'full' ? t('studio.api.scopeFullSub') : t('studio.api.scopeContribSub') }}</small>
          </button>
        </div>
        <div class="ft">
          <button type="button" class="st-btn g" @click="createOpen = false">{{ t('studio.cancel') }}</button>
          <button type="button" class="st-btn p" :disabled="creating || !newName.trim()" @click="create">{{ t('studio.api.create') }}</button>
        </div>
      </template>
      <template v-else>
        <div class="mi ok"><SIcon name="check" :size="22" /></div>
        <h3>{{ t('studio.api.createdTitle', { name: created.name }) }}</h3>
        <p>{{ t('studio.api.createdDesc') }}</p>
        <div class="newkey"><span class="mono">{{ created.key }}</span><button type="button" class="st-ibtn" :title="t('studio.copy')" @click="copy(created.key)"><SIcon name="copy" :size="18" /></button></div>
        <div class="warnline"><SIcon name="lock" :size="18" />{{ t('studio.api.onlyOnce') }}</div>
        <div class="ft">
          <button type="button" class="st-btn p" @click="createOpen = false">{{ t('studio.api.savedIt') }}</button>
        </div>
      </template>
    </StModal>
  </section>
</template>

<style scoped lang="scss">
.view {
  max-width: 1280px;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

.hk { font-size: 13px; padding: 1px 6px; border-radius: var(--r-xs); background: var(--well); margin: 0 3px; }

.keys { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; margin-bottom: 32px; }
/* 无 Key：压成一条横向提示卡，不占整屏 */
.keys-empty {
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr);
  grid-template-rows: auto auto;
  column-gap: 20px;
  align-items: center;
  text-align: left;
  padding: 18px 22px;
  margin-bottom: 28px;
  border-radius: var(--r-lg);
  box-shadow: 0 0 0 1px var(--line-2);

  :deep(.empty-art) { grid-row: 1 / 3; }
  h4 { margin: 0 0 2px; align-self: end; }
  p { margin: 0; align-self: start; }
}

.key {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr) auto auto auto;
  align-items: center;
  gap: 16px;
  padding: 16px 18px;
  border-radius: var(--r-md);
  box-shadow: 0 0 0 1px var(--line-2);
  transition: box-shadow var(--dur);

  &:hover { box-shadow: 0 0 0 1px var(--line-3), var(--sh-card-hover); }

  .ki { width: 44px; height: 44px; border-radius: var(--r-md); display: grid; place-items: center; background: var(--tint); color: var(--ink); }
  .kn { min-width: 0; }
  h4 { margin: 0 0 4px; font: 700 17px/1.35 var(--font-serif); }

  .kv2 {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: var(--st-ink-3);

    button { width: 28px; height: 28px; display: grid; place-items: center; border-radius: var(--r-xs); color: var(--st-ink-3); }
    button:hover { background: var(--hover); color: var(--st-ink); }

    .fresh { font-size: 11px; padding: 1px 7px; border-radius: var(--r-xs); background: color-mix(in oklab, var(--green) 14%, var(--paper)); color: color-mix(in oklab, var(--green) 70%, var(--st-ink)); }
  }

  .scopes { display: flex; flex-wrap: wrap; align-items: center; gap: 5px; margin-top: 8px; }
  .scopes span { font-size: 12.5px; padding: 3px 10px; border-radius: var(--r-pill); box-shadow: 0 0 0 1px var(--line-2) inset; color: var(--st-ink-2); }
  .scopes b { font-size: 12px; font-weight: 600; padding: 3px 10px; border-radius: var(--r-pill); background: var(--well-2); color: var(--st-ink-2); }
  .scopes.full b { background: color-mix(in oklab, var(--yellow) 22%, transparent); color: color-mix(in oklab, var(--yellow) 55%, var(--st-ink)); }

  .lu { font-size: 14px; white-space: nowrap; }
  .lu small { display: block; font-size: 12.5px; color: var(--st-ink-3); margin-top: 3px; }
}

.code {
  position: relative;
  margin: 0;
  border-radius: var(--r-md);
  background: var(--code-bg);
  color: #d9d6cf;
  padding: 20px 22px;
  font: 13px/1.9 var(--font-mono);
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.04) inset;

  code { font: inherit; background: none; padding: 0; color: inherit; }

  .cp { position: absolute; right: 12px; top: 12px; color: #aaa; background: rgba(255, 255, 255, 0.06); }
  .cp:hover { color: #fff; background: rgba(255, 255, 255, 0.12); }
}

.api-grid {
  display: grid;
  grid-template-columns: minmax(0, 7fr) minmax(0, 5fr);
  align-items: stretch;
  gap: 20px;
  margin-bottom: 36px;

  > div { display: flex; flex-direction: column; min-width: 0; }
  .code { flex: 1; }
}

em {
  font-style: normal;
  font-weight: 600;
  font-size: 11.5px;
  width: 54px;
  flex: none;
  text-align: center;
  padding: 2px 0;
  border-radius: var(--r-xs);
  font-family: var(--font-mono);

  &.post { background: color-mix(in oklab, var(--green) 16%, var(--paper)); color: color-mix(in oklab, var(--green) 70%, var(--st-ink)); }
  &.get { background: color-mix(in oklab, var(--blue) 14%, var(--paper)); color: color-mix(in oklab, var(--blue) 70%, var(--st-ink)); }
  &.put { background: color-mix(in oklab, var(--yellow) 18%, var(--paper)); color: color-mix(in oklab, var(--yellow) 50%, var(--st-ink)); }
  &.delete { background: color-mix(in oklab, var(--red) 12%, var(--paper)); color: color-mix(in oklab, var(--red) 70%, var(--st-ink)); }
}

.endp {
  display: flex;
  flex-direction: column;

  div { display: flex; align-items: center; gap: 10px; padding: 10px 0; border-bottom: 1px solid var(--line); font: 13.5px var(--font-mono); color: var(--st-ink-2); }
}

.note { font-size: 13px; color: var(--st-ink-3); line-height: 1.7; margin: 12px 0 0; }

.tester { margin-bottom: 36px; }

.auth {
  display: grid;
  grid-template-columns: 220px 280px 1fr;
  gap: 10px;
  align-items: center;
  margin-bottom: 16px;

  small { font-size: 13px; color: var(--st-ink-3); line-height: 1.5; }
}

.eps { display: flex; flex-direction: column; gap: 6px; }

.ep {
  border-radius: var(--r-md);
  transition: background var(--dur-fast), box-shadow var(--dur-fast);

  &.open { background: var(--well); }

  .ep-row {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    padding: 10px 12px;
    border-radius: var(--r-md);
    text-align: left;

    &:hover { background: var(--well); }

    code { font: 13.5px var(--font-mono); color: var(--st-ink); white-space: nowrap; }
    .d { flex: 1; min-width: 0; font-size: 14px; color: var(--st-ink-3); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

    .au {
      font-size: 12px;
      padding: 2px 9px;
      border-radius: var(--r-xs);
      background: var(--well-2);
      color: var(--st-ink-3);

      &.apikey { background: var(--tint); color: var(--ink); }
    }

    .chev { color: var(--st-ink-4); transition: transform var(--dur) var(--ease-spring); }
  }

  &.open .chev { transform: rotate(180deg); }

  .ep-body {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 4px 12px 14px;
    animation: body-in var(--dur) var(--ease-out);

    .st-field { background: var(--paper); }
  }

  .send { display: flex; align-items: center; gap: 12px; }

  .status {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font: 12px var(--font-mono);
    color: var(--red);

    .st-dot { --c: var(--red); }
    &.ok { color: color-mix(in oklab, var(--green) 70%, var(--st-ink)); .st-dot { --c: var(--green); } }
  }

  .resp { max-height: 320px; font-size: 12px; line-height: 1.7; }
}

@keyframes body-in { from { opacity: 0; transform: translateY(-4px); } }

.prompt-sec { margin-bottom: 36px; }
.prompt-sec .acts { display: flex; gap: 8px; }

.prompt {
  margin: 14px 0 0;
  max-height: 280px;
  overflow: auto;
  padding: 18px 20px;
  border-radius: var(--r-md);
  background: var(--well);
  font: 13px/1.8 var(--font-mono);
  color: var(--st-ink-2);
  white-space: pre-wrap;

  code { font: inherit; background: none; padding: 0; }
}

/* ----- 调用日志 ----- */
.log-sec { scroll-margin-top: 24px; }
.log-sub { margin: -4px 0 14px; }

.log-tools {
  display: flex;
  align-items: center;
  gap: 8px;

  .key-pick {
    position: relative;
    width: 200px;
    min-height: 36px;

    select { flex: 1; min-width: 0; padding-right: 18px; background: none; border: 0; }
    .chev { position: absolute; right: 10px; pointer-events: none; color: var(--st-ink-4); }
  }

  .spin :deep(svg) { animation: log-spin 0.8s linear infinite; }
}

@keyframes log-spin { to { transform: rotate(360deg); } }

.logs {
  table-layout: fixed;
  transition: opacity var(--dur-fast);

  &.busy { opacity: 0.6; }

  th:nth-child(1) { width: 146px; }
  th:nth-child(2) { width: 22%; }
  th:nth-child(4) { width: 80px; }
  th:nth-child(5) { width: 84px; }
  th:nth-child(6) { width: 128px; }
  .r { text-align: right; }
  .nw { white-space: nowrap; }
  td.dim { font-size: 12.5px; color: var(--st-ink-3); }
  td.ip { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

  tbody tr { animation: log-row-in var(--dur) var(--ease-out) both; animation-delay: calc(var(--i, 0) * 14ms); }

  .kc {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;

    b { font-weight: 500; color: var(--st-ink); margin-right: 6px; }
    small { display: block; font-size: 11.5px; color: var(--st-ink-4); margin-top: 1px; }
  }

  .tag {
    font-size: 11.5px;
    padding: 1px 7px;
    border-radius: var(--r-xs);
    background: var(--well-2);
    color: var(--st-ink-3);

    &.warn { background: color-mix(in oklab, var(--red) 10%, var(--paper)); color: color-mix(in oklab, var(--red) 70%, var(--st-ink)); }
  }

  .req {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;

    .p { flex: 1; min-width: 0; font-size: 12.5px; color: var(--st-ink-2); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  }

  .sc {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font: 500 12.5px var(--font-mono);
    color: color-mix(in oklab, var(--green) 70%, var(--st-ink));

    .st-dot { --c: var(--green); }
    &.bad { color: var(--red); .st-dot { --c: var(--red); } }
    &.warn { color: color-mix(in oklab, var(--yellow) 55%, var(--st-ink)); .st-dot { --c: var(--yellow); } }
  }
}

@keyframes log-row-in { from { opacity: 0; transform: translateY(4px); } }

.key .lu .calls {
  display: block;
  margin-top: 3px;
  font-size: 12.5px;
  color: var(--ink);
  text-align: left;

  &:hover { text-decoration: underline; text-underline-offset: 3px; }
  .bad { color: var(--red); }
}

@media (prefers-reduced-motion: reduce) {
  .logs tbody tr, .log-tools .spin :deep(svg) { animation: none; }
}

.newkey {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 14px;
  border-radius: var(--r-sm);
  background: var(--code-bg);
  color: #e9e6df;
  font-size: 13px;
  word-break: break-all;

  .mono { flex: 1; }
  .st-ibtn { color: #bbb; }
  .st-ibtn:hover { color: #fff; background: rgba(255, 255, 255, 0.1); }
}

.warnline {
  display: flex;
  gap: 8px;
  align-items: center;
  font-size: 12.5px;
  color: color-mix(in oklab, var(--yellow) 55%, var(--st-ink));
  margin-top: 10px;
}

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
  .api-grid { grid-template-columns: 1fr; }
  .keys { grid-template-columns: 1fr; }
  .auth { grid-template-columns: 1fr 1fr; }
  .auth small { grid-column: 1 / -1; }
}

/* 新建 Key：权限二选一，选中抬升 + 轻染 */
.scope-label { margin-top: 14px; }

.scope-pick {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;

  button {
    display: grid;
    grid-template-columns: auto 1fr;
    align-items: center;
    gap: 4px 8px;
    padding: 12px 14px;
    border-radius: var(--r-md);
    text-align: left;
    color: var(--st-ink-2);
    box-shadow: inset 0 0 0 1px var(--line-2);
    transition: background var(--dur) var(--ease-out), box-shadow var(--dur) var(--ease-out);

    &:hover { background: var(--well); }
    b { font-size: 14px; font-weight: 600; color: var(--st-ink); }
    small { grid-column: 1 / -1; font-size: 12.5px; line-height: 1.55; color: var(--st-ink-3); }

    &.on {
      color: var(--ink);
      background: color-mix(in oklab, var(--ink) 6%, var(--paper));
      box-shadow: inset 0 0 0 1.5px color-mix(in oklab, var(--ink) 60%, transparent);
    }
  }
}
</style>
