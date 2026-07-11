<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';
import { useConfigStore, type SiteConfig, type ThemePreset } from '../../stores/config';
import { useThemeStore } from '../../stores/theme';

const { t } = useI18n();
const configStore = useConfigStore();

const cfg = reactive<SiteConfig>(JSON.parse(JSON.stringify(configStore.cfg)));
const message = ref('');
const busy = ref(false);

const TIMEZONES = [
  'Asia/Shanghai', 'Asia/Tokyo', 'Asia/Singapore', 'Asia/Hong_Kong',
  'UTC', 'Europe/London', 'Europe/Berlin', 'America/New_York', 'America/Los_Angeles',
];

/** 锚点子导航 */
const SECTIONS = [
  { id: 'sec-platform', key: 'admin.secPlatform' },
  { id: 'sec-loading', key: 'admin.secLoading' },
  { id: 'sec-theme', key: 'admin.secTheme' },
  { id: 'sec-hero', key: 'admin.secHero' },
  { id: 'sec-timezone', key: 'admin.secTimezone' },
  { id: 'sec-github', key: 'admin.secGithub' },
  { id: 'sec-about', key: 'admin.secAbout' },
  { id: 'sec-users', key: 'admin.secUsers' },
];

/* GitHub 同步面板 */
interface SyncLogEntry { at: string; ok: boolean; message: string }
const ghLog = ref<SyncLogEntry[]>([]);
const ghPreview = ref<{ stats?: Record<string, number>; fetchedAt?: string } | null>(null);
const ghSyncing = ref(false);
const ghMsg = ref('');

async function loadGhLog(): Promise<void> {
  try {
    const data = await adminApi.githubLog();
    ghLog.value = data.log;
    ghPreview.value = data.preview as typeof ghPreview.value;
  } catch { /* 忽略 */ }
}

async function syncNow(): Promise<void> {
  if (ghSyncing.value) return;
  ghSyncing.value = true;
  ghMsg.value = '';
  try {
    await adminApi.githubSync();
    ghMsg.value = t('admin.ghSyncOk');
  } catch {
    ghMsg.value = t('admin.ghSyncFail');
  } finally {
    ghSyncing.value = false;
    await loadGhLog();
  }
}

async function load(): Promise<void> {
  const remote = await adminApi.settings() as unknown as SiteConfig;
  Object.assign(cfg, JSON.parse(JSON.stringify(remote)));
}

async function save(): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  message.value = '';
  try {
    cfg.theme.displayCount = Math.min(4, Math.max(1, cfg.theme.displayCount));
    await adminApi.saveSettings(cfg as unknown as Record<string, unknown>);
    await configStore.load();
    useThemeStore().init();
    message.value = t('admin.saved');
  } catch {
    message.value = t('admin.saveFailed');
  } finally {
    busy.value = false;
  }
}

/* ===== 主题预设：拖拽排序 / 增删 ===== */
let dragIndex = -1;

function onDragStart(i: number): void {
  dragIndex = i;
}

function onDrop(i: number): void {
  if (dragIndex < 0 || dragIndex === i) return;
  const list = cfg.theme.presets;
  const [moved] = list.splice(dragIndex, 1);
  list.splice(i, 0, moved);
  dragIndex = -1;
}

function addPreset(): void {
  if (cfg.theme.presets.length >= 10) return;
  cfg.theme.presets.push({
    id: `custom-${Date.now().toString(36)}`,
    name: '自定义',
    primary: '#7c4dff',
    primaryDeep: '#5e35b1',
  });
}

function removePreset(preset: ThemePreset): void {
  if (cfg.theme.presets.length <= 1) return;
  cfg.theme.presets = cfg.theme.presets.filter((p) => p.id !== preset.id);
  if (cfg.theme.defaultPaletteId === preset.id) {
    cfg.theme.defaultPaletteId = cfg.theme.presets[0].id;
  }
}

/* ===== 改密 ===== */
const oldPassword = ref('');
const newPassword = ref('');
const confirmPw = ref('');
const pwMessage = ref('');
const pwOk = ref(false);

async function changePassword(): Promise<void> {
  pwMessage.value = '';
  if (newPassword.value.length < 8) {
    pwMessage.value = t('admin.pwTooShort');
    return;
  }
  if (newPassword.value !== confirmPw.value) {
    pwMessage.value = t('admin.pwMismatch');
    return;
  }
  try {
    await adminApi.changePassword(oldPassword.value, newPassword.value);
    pwOk.value = true;
    pwMessage.value = t('admin.pwChanged');
    oldPassword.value = newPassword.value = confirmPw.value = '';
  } catch {
    pwOk.value = false;
    pwMessage.value = t('admin.pwWrong');
  }
}

onMounted(async () => {
  await load();
  await loadGhLog();
});
</script>

<template>
  <div class="settings">
    <header class="head">
      <h1 class="page-h">{{ t('admin.menuSettings') }}</h1>
      <div class="actions">
        <span v-if="message" class="msg">{{ message }}</span>
        <button class="btn primary" :disabled="busy" @click="save">{{ t('admin.saveAll') }}</button>
      </div>
    </header>

    <!-- 锚点子导航 -->
    <nav class="subnav">
      <a v-for="s in SECTIONS" :key="s.id" :href="`#${s.id}`">{{ t(s.key) }}</a>
    </nav>

    <!-- 平台 -->
    <section id="sec-platform" class="card">
      <h2>{{ t('admin.secPlatform') }}</h2>
      <div class="row2">
        <label><span>{{ t('admin.siteTitle') }}</span><input v-model="cfg.site.title" type="text" /></label>
        <label><span>{{ t('admin.siteSubtitle') }}</span><input v-model="cfg.site.subtitle" type="text" /></label>
      </div>
    </section>

    <!-- Loading 文案 -->
    <section id="sec-loading" class="card">
      <h2>{{ t('admin.secLoading') }}</h2>
      <div class="row2">
        <label><span>{{ t('admin.bootText') }}</span><input v-model="cfg.loading.bootText" type="text" /></label>
        <label><span>{{ t('admin.routeText') }}</span><input v-model="cfg.loading.routeText" type="text" /></label>
      </div>
    </section>

    <!-- 主题 -->
    <section id="sec-theme" class="card">
      <h2>{{ t('admin.secTheme') }}</h2>
      <div class="row3">
        <label>
          <span>{{ t('admin.defaultPalette') }}</span>
          <select v-model="cfg.theme.defaultPaletteId">
            <option v-for="p in cfg.theme.presets" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
        </label>
        <label>
          <span>{{ t('admin.defaultMode') }}</span>
          <select v-model="cfg.theme.defaultMode">
            <option value="light">{{ t('theme.light') }}</option>
            <option value="dark">{{ t('theme.dark') }}</option>
          </select>
        </label>
        <label>
          <span>{{ t('admin.autoSwitch') }}</span>
          <select v-model="cfg.theme.autoSwitch">
            <option value="off">{{ t('admin.autoOff') }}</option>
            <option value="season">{{ t('admin.autoSeason') }}</option>
          </select>
        </label>
      </div>
      <div class="row3">
        <label class="switch">
          <input v-model="cfg.theme.allowUserPalette" type="checkbox" />
          <i class="track" aria-hidden="true" />
          <span>{{ t('admin.allowUserPalette') }}</span>
        </label>
        <label>
          <span>{{ t('admin.displayCount') }}</span>
          <input v-model.number="cfg.theme.displayCount" type="number" min="1" max="4" />
        </label>
      </div>

      <!-- 预设列表：拖拽排序，前 N 个展示 -->
      <p class="sub-h">{{ t('admin.presets', { n: cfg.theme.presets.length }) }}</p>
      <ul class="presets">
        <li
          v-for="(preset, i) in cfg.theme.presets"
          :key="preset.id"
          :class="{ shown: i < cfg.theme.displayCount }"
          draggable="true"
          @dragstart="onDragStart(i)"
          @dragover.prevent
          @drop="onDrop(i)"
        >
          <span class="grip" aria-hidden="true">⋮⋮</span>
          <input v-model="preset.name" class="p-name" type="text" />
          <label class="color"><span>{{ t('admin.primary') }}</span><input v-model="preset.primary" type="color" /></label>
          <label class="color"><span>{{ t('admin.primaryDeep') }}</span><input v-model="preset.primaryDeep" type="color" /></label>
          <span class="p-flag">{{ i < cfg.theme.displayCount ? t('admin.shown') : t('admin.hidden') }}</span>
          <button class="op danger" @click="removePreset(preset)">{{ t('admin.delete') }}</button>
        </li>
      </ul>
      <button class="btn ghost" :disabled="cfg.theme.presets.length >= 10" @click="addPreset">
        {{ t('admin.addPreset') }}
      </button>
    </section>

    <!-- 首页轮播规则 -->
    <section id="sec-hero" class="card">
      <h2>{{ t('admin.secHero') }}</h2>
      <div class="row3">
        <label>
          <span>{{ t('admin.heroInterval') }}</span>
          <input v-model.number="cfg.hero.intervalMs" type="number" min="1000" step="500" />
        </label>
        <label>
          <span>{{ t('admin.heroCount') }}</span>
          <input v-model.number="cfg.hero.count" type="number" min="1" max="10" />
        </label>
        <label>
          <span>{{ t('admin.heroPinnedRule') }}</span>
          <select v-model="cfg.hero.pinnedRule">
            <option value="pinned-first">{{ t('admin.pinnedFirst') }}</option>
            <option value="ignore">{{ t('admin.pinnedIgnore') }}</option>
          </select>
        </label>
      </div>
      <label>
        <span>{{ t('admin.thoughtsSubtitle') }}</span>
        <input v-model="cfg.thoughts.subtitle" type="text" />
      </label>
    </section>

    <!-- 时区 -->
    <section id="sec-timezone" class="card">
      <h2>{{ t('admin.secTimezone') }}</h2>
      <label class="narrow">
        <span>{{ t('admin.timezone') }}</span>
        <select v-model="cfg.timezone">
          <option v-for="tz in TIMEZONES" :key="tz" :value="tz">{{ tz }}</option>
        </select>
      </label>
    </section>

    <!-- GitHub -->
    <section id="sec-github" class="card">
      <h2>{{ t('admin.secGithub') }}</h2>
      <div class="row3">
        <label><span>{{ t('admin.ghUser') }}</span><input v-model="cfg.github.username" type="text" /></label>
        <label>
          <span>{{ t('admin.ghMode') }}</span>
          <select v-model="cfg.github.mode">
            <option value="manual">{{ t('admin.ghManual') }}</option>
            <option value="api">{{ t('admin.ghApi') }}</option>
          </select>
        </label>
        <label><span>{{ t('admin.ghRefresh') }}</span><input v-model.number="cfg.github.refreshMinutes" type="number" min="1" /></label>
      </div>
      <label v-if="cfg.github.mode === 'api'">
        <span>{{ t('admin.ghToken') }}</span>
        <input v-model="cfg.github.token" type="password" autocomplete="off" placeholder="ghp_…（可留空走匿名公开接口）" />
      </label>
      <div v-if="cfg.github.mode === 'api'" class="row2" style="margin-top: 14px">
        <label>
          <span>{{ t('admin.ghProxy') }}</span>
          <input v-model="cfg.github.proxy" type="text" placeholder="http://127.0.0.1:7890（留空读 HTTPS_PROXY）" />
        </label>
        <label class="switch" style="align-self: end">
          <input v-model="cfg.github.insecureTls" type="checkbox" />
          <i class="track" aria-hidden="true" />
          <span>{{ t('admin.ghInsecure') }}</span>
        </label>
      </div>
      <div v-if="cfg.github.mode === 'manual'" class="row3" style="margin-top: 14px">
        <label><span>{{ t('admin.ghRepos') }}</span><input v-model.number="cfg.github.stats.repos" type="number" /></label>
        <label><span>Stars</span><input v-model.number="cfg.github.stats.stars" type="number" /></label>
        <label><span>{{ t('admin.ghFollowers') }}</span><input v-model.number="cfg.github.stats.followers" type="number" /></label>
      </div>
      <div v-if="cfg.github.mode === 'manual'" class="row3">
        <label><span>{{ t('admin.ghCommits') }}</span><input v-model.number="cfg.github.stats.commits" type="number" /></label>
      </div>

      <!-- 同步面板：立即同步 / 数据预览 / 同步日志 -->
      <div v-if="cfg.github.mode === 'api'" class="gh-panel">
        <div class="gh-actions">
          <button class="btn primary" :disabled="ghSyncing" @click="syncNow">
            {{ ghSyncing ? t('admin.ghSyncing') : t('admin.ghSyncNow') }}
          </button>
          <span v-if="ghMsg" class="msg">{{ ghMsg }}</span>
        </div>

        <div v-if="ghPreview?.stats" class="gh-preview">
          <div class="gp"><strong>{{ ghPreview.stats.repos }}</strong><span>{{ t('admin.ghRepos') }}</span></div>
          <div class="gp"><strong>{{ ghPreview.stats.stars }}</strong><span>Stars</span></div>
          <div class="gp"><strong>{{ ghPreview.stats.followers }}</strong><span>{{ t('admin.ghFollowers') }}</span></div>
          <div class="gp"><strong>{{ ghPreview.stats.commits }}</strong><span>{{ t('admin.ghCommits') }}</span></div>
        </div>

        <ul v-if="ghLog.length" class="gh-log">
          <li v-for="(entry, i) in ghLog" :key="i" :class="{ err: !entry.ok }">
            <time>{{ entry.at.slice(5, 19).replace('T', ' ') }}</time>
            <span>{{ entry.message }}</span>
          </li>
        </ul>
        <p v-else class="hint">{{ t('admin.ghNoLog') }}</p>
      </div>
    </section>

    <!-- 关于：迁至独立管理页 -->
    <section id="sec-about" class="card about-link">
      <h2>{{ t('admin.secAbout') }}</h2>
      <p class="hint">{{ t('admin.aboutMoved') }}</p>
      <router-link to="/admin/about" class="btn primary">{{ t('admin.goAbout') }}</router-link>
    </section>

    <!-- 用户管理 -->
    <section id="sec-users" class="card">
      <h2>{{ t('admin.secUsers') }}</h2>
      <p class="hint">{{ t('admin.usersHint') }}</p>
      <div class="pw-grid">
        <label><span>{{ t('admin.oldPassword') }}</span><input v-model="oldPassword" type="password" autocomplete="current-password" /></label>
        <label><span>{{ t('admin.newPassword') }}</span><input v-model="newPassword" type="password" autocomplete="new-password" /></label>
        <label><span>{{ t('admin.confirmPassword') }}</span><input v-model="confirmPw" type="password" autocomplete="new-password" /></label>
      </div>
      <div class="pw-actions">
        <button class="btn ghost" @click="changePassword">{{ t('admin.changePassword') }}</button>
        <span v-if="pwMessage" class="msg" :class="{ err: !pwOk }">{{ pwMessage }}</span>
      </div>
    </section>
  </div>
</template>

<style scoped lang="scss">
.settings {
  max-width: 860px;
}

.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 22px;
}

.page-h { font-size: 26px; }

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.msg {
  font-size: 13px;
  color: var(--primary);

  &.err { color: var(--accent-red); }
}

/* 锚点子导航 */
.subnav {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 10px 12px;
  margin-bottom: 16px;
  border-radius: var(--radius);
  background: var(--glass);
  backdrop-filter: blur(14px) saturate(1.4);
  -webkit-backdrop-filter: blur(14px) saturate(1.4);
  border: 1px solid var(--border);

  a {
    font-size: 12.5px;
    font-weight: 600;
    padding: 6px 14px;
    border-radius: 999px;
    color: var(--text-2);
    transition: all var(--dur-fast);

    &:hover {
      background: rgba(var(--primary-rgb), 0.1);
      color: var(--primary);
    }
  }
}

/* 开关 */
.switch {
  flex-direction: row !important;
  align-items: center;
  gap: 10px;
  cursor: pointer;

  input { display: none; }

  .track {
    width: 40px;
    height: 22px;
    border-radius: 999px;
    background: var(--surface-2);
    border: 1px solid var(--border);
    position: relative;
    transition: background var(--dur-fast), border-color var(--dur-fast);

    &::after {
      content: '';
      position: absolute;
      top: 2px;
      left: 2px;
      width: 16px;
      height: 16px;
      border-radius: 50%;
      background: var(--text-2);
      transition: transform var(--dur-fast) var(--ease-spring), background var(--dur-fast);
    }
  }

  input:checked + .track {
    background: rgba(var(--primary-rgb), 0.25);
    border-color: rgba(var(--primary-rgb), 0.5);

    &::after {
      transform: translateX(18px);
      background: var(--primary);
    }
  }

  span { font-size: 13px; color: var(--text); }
}

/* GitHub 同步面板 */
.gh-panel {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px dashed var(--border);
}

.gh-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}

.gh-preview {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
  margin-bottom: 14px;
}

.gp {
  text-align: center;
  padding: 12px 6px;
  border-radius: 10px;
  background: var(--bg);
  border: 1px solid var(--border);

  strong {
    display: block;
    font-family: var(--font-serif);
    font-size: 20px;
    background: var(--grad-title);
    background-clip: text;
    -webkit-background-clip: text;
    color: transparent;
  }

  span { font-size: 11px; color: var(--text-2); }
}

.gh-log {
  list-style: none;
  max-height: 180px;
  overflow-y: auto;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--bg);

  li {
    display: flex;
    gap: 12px;
    padding: 8px 12px;
    font-size: 12.5px;
    border-bottom: 1px solid var(--border);

    time { color: var(--text-2); flex-shrink: 0; font-variant-numeric: tabular-nums; }
    span { color: var(--text); }

    &.err span { color: var(--accent-red); }
    &:last-child { border-bottom: none; }
  }
}

.about-link .btn { margin-top: 4px; display: inline-block; }

.card {
  padding: 22px 24px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  margin-bottom: 18px;

  h2 {
    font-size: 17px;
    margin-bottom: 16px;
    padding-left: 12px;
    border-left: 4px solid var(--primary);
  }
}

.hint { font-size: 13px; color: var(--text-2); margin-bottom: 14px; }
.sub-h { font-size: 13px; font-weight: 600; color: var(--text-2); margin: 18px 0 10px; }

.row2, .row3, .pw-grid {
  display: grid;
  gap: 14px;
  margin-bottom: 14px;
}

.row2 { grid-template-columns: repeat(2, 1fr); }
.row3, .pw-grid { grid-template-columns: repeat(3, 1fr); }

label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;

  span { font-size: 12px; font-weight: 600; color: var(--text-2); }

  &.narrow { max-width: 280px; }

  &.check {
    flex-direction: row;
    align-items: center;
    gap: 8px;

    span { font-size: 13px; color: var(--text); }
    input { accent-color: var(--primary); width: 16px; height: 16px; }
  }
}

input:not([type='checkbox']):not([type='color']), select, textarea {
  padding: 10px 13px;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text);
  font-size: 14px;
  font-family: inherit;
  outline: none;
  resize: vertical;
  transition: border-color var(--dur-fast), box-shadow var(--dur-fast);

  &:focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(var(--primary-rgb), 0.12);
  }
}

/* 预设行 */
.presets {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;

  li {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--bg);
    cursor: grab;
    opacity: 0.6;

    &.shown { opacity: 1; border-color: rgba(var(--primary-rgb), 0.35); }
    &:active { cursor: grabbing; }
  }
}

.grip { color: var(--text-2); letter-spacing: -2px; user-select: none; }

.p-name { flex: 1; max-width: 180px; }

.color {
  flex-direction: row !important;
  align-items: center;
  gap: 6px;

  input[type='color'] {
    width: 34px;
    height: 26px;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: none;
    padding: 0;
    cursor: pointer;
  }
}

.p-flag {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 10px;
  border-radius: 999px;
  background: var(--surface-2);
  color: var(--text-2);
  margin-left: auto;
}

li.shown .p-flag {
  background: rgba(var(--primary-rgb), 0.12);
  color: var(--primary);
}

.op {
  border: none;
  background: none;
  font-size: 13px;
  font-weight: 600;

  &.danger { color: var(--accent-red); }
  &:hover { opacity: 0.75; }
}

.btn {
  padding: 10px 24px;
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

    &:hover:not(:disabled) { border-color: var(--primary); color: var(--primary); }
    &:disabled { opacity: 0.5; }
  }
}

.pw-actions {
  display: flex;
  align-items: center;
  gap: 14px;
}

@media (max-width: 800px) {
  .row2, .row3, .pw-grid { grid-template-columns: 1fr; }
}
</style>
