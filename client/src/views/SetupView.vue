<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../api';
import { useAuthStore } from '../stores/auth';
import { useConfigStore } from '../stores/config';
import './admin/studio/i18n';
import SIcon from './admin/studio/SIcon.vue';
import '@fontsource/noto-serif-sc/400.css';
import '@fontsource/noto-serif-sc/600.css';

/**
 * 首次启动初始化（/setup）：初始化码 → 管理员账号 → 站点与身份 → 示例内容 → 完成。
 * 初始化码只打印在服务端启动日志里，证明操作者能接触到服务器。
 * 改密模式（/setup?change=1）：旧库仍在用历史默认口令时，登录后必须先在这里改密。
 * 左侧封面在程序生成的默认封面间缓慢交叠轮换；步骤之间横向滑动切换。
 */
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const config = useConfigStore();

const changeMode = computed(() => route.query.change === '1');

/* ---------- 左侧封面轮换 ---------- */
const COVERS = ['05', '12', '01', '09', '02', '10'];
const coverIdx = ref(0);
let coverTimer = 0;

/* ---------- 步骤 ---------- */
type Step = 'code' | 'database' | 'admin' | 'site' | 'content' | 'done';
const canConfigureDatabase = ref(false);
const currentDatabase = ref('sqlite');
const databaseChoice = ref<'current' | 'sqlite' | 'mysql'>('current');
const mysql = reactive({ host: '127.0.0.1', port: 3306, name: 'myself', user: 'myself', password: '', tls: 'false' as 'false' | 'true' });
const STEPS = computed<Step[]>(() => ['code', ...(canConfigureDatabase.value ? ['database' as const] : []), 'admin', 'site', 'content']);
const step = ref<Step>('code');
const dir = ref<1 | -1>(1);
const stepNo = computed(() => STEPS.value.indexOf(step.value));

const form = reactive({
  code: '',
  username: 'admin',
  password: '',
  confirm: '',
  title: config.cfg.site.title || 'Myself',
  url: '',
  name: '',
  tagline: '',
  motto: config.cfg.about.motto || '',
  demo: true,
});
const show = ref(false);
const error = ref('');
const busy = ref(false);

const strength = computed(() => {
  const v = form.password;
  let s = 0;
  if (v.length >= 8) s++;
  if (v.length >= 12) s++;
  if (/[A-Z]/.test(v) && /[a-z]/.test(v)) s++;
  if (/\d/.test(v) && /[^A-Za-z0-9]/.test(v)) s++;
  return v ? Math.max(1, s) : 0;
});

function validate(s: Step): string {
  if (s === 'code' && !/^[0-9a-fA-F]{16}$/.test(form.code.replace(/[\s-]/g, ''))) return t('setup.err.code');
  if (s === 'admin') {
    if (!/^[A-Za-z0-9_.-]{3,32}$/.test(form.username.trim())) return t('setup.err.username');
    if (form.password.length < 8) return t('setup.err.short');
    if (form.password !== form.confirm) return t('setup.err.mismatch');
  }
  if (s === 'site') {
    if (!form.title.trim()) return t('setup.err.title');
    if (!form.name.trim()) return t('setup.err.name');
    if (form.url.trim() && !/^https?:\/\/[^\s/]+/.test(form.url.trim())) return t('setup.err.url');
  }
  return '';
}

function go(to: Step): void {
  dir.value = STEPS.value.indexOf(to) >= stepNo.value ? 1 : -1;
  error.value = '';
  step.value = to;
}

async function next(): Promise<void> {
  const msg = validate(step.value);
  if (msg) {
    error.value = msg;
    shake();
    return;
  }
  if (step.value === 'content') {
    void finish();
    return;
  }
  if (step.value === 'database') {
    busy.value = true;
    try {
      const driver = databaseChoice.value;
      const expectedDriver = driver === 'current' ? currentDatabase.value : driver;
      await adminApi.setupDatabase(form.code.trim(), { driver, ...(driver === 'mysql' ? { mysql: { ...mysql } } : {}) });
      let ready = false;
      for (let i = 0; i < 60; i++) {
        await new Promise(resolve => window.setTimeout(resolve, 500));
        try {
          const status = await adminApi.setupStatus();
          if (status.databaseDriver === expectedDriver) { ready = true; break; }
        } catch { /* The server closes and reopens its listener during reconfiguration. */ }
      }
      if (!ready) throw new Error('database_restart_timeout');
      currentDatabase.value = expectedDriver;
      databaseChoice.value = 'current';
      mysql.password = '';
    } catch (e) {
      error.value = explain(e);
      shake();
      return;
    } finally { busy.value = false; }
  }
  // 初始化码当场向服务端核对，错了不必等到最后一步才知道
  if (step.value === 'code') {
    if (busy.value) return;
    busy.value = true;
    try {
      await adminApi.verifySetupCode(form.code.trim());
    } catch (e) {
      error.value = explain(e);
      shake();
      return;
    } finally {
      busy.value = false;
    }
  }
  go(STEPS.value[stepNo.value + 1]);
}

function back(): void {
  if (stepNo.value > 0) go(STEPS.value[stepNo.value - 1]);
}

const ERRORS: Record<string, string> = {
  bad_setup_code: 'setup.err.badCode',
  too_many_attempts: 'setup.err.locked',
  already_setup: 'setup.err.already',
  invalid_username: 'setup.err.username',
  weak_password: 'setup.err.short',
  invalid_site_url: 'setup.err.url',
  bad_credentials: 'setup.err.oldWrong',
  change_code_required: 'setup.err.changeCode',
  database_configuration_failed: 'setup.database.failed',
  database_configuration_unavailable: 'setup.database.unavailable',
  database_restart_timeout: 'setup.database.timeout',
};

function explain(e: unknown): string {
  const key = ERRORS[(e as Error).message];
  return key ? t(key) : t('setup.err.generic');
}

async function finish(): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  error.value = '';
  try {
    const res = await adminApi.setup({
      code: form.code.trim(),
      username: form.username.trim(),
      password: form.password,
      site: { title: form.title.trim(), url: form.url.trim() || undefined },
      identity: { name: form.name.trim(), tagline: form.tagline.trim(), motto: form.motto.trim() },
      demo: form.demo,
    });
    auth.markLoggedIn(res.user);
    await config.load();
    dir.value = 1;
    step.value = 'done';
  } catch (e) {
    error.value = explain(e);
    // 初始化码错了回到第一步重填
    if ((e as Error).message === 'bad_setup_code') go('code');
    shake();
  } finally {
    busy.value = false;
  }
}

/* ---------- 改密模式 ---------- */
const pw = reactive({ old: '', next: '', confirm: '', code: '' });

async function changePassword(): Promise<void> {
  if (busy.value) return;
  error.value = '';
  if (pw.next.length < 8) error.value = t('setup.err.short');
  else if (pw.next !== pw.confirm) error.value = t('setup.err.mismatch');
  else if (pw.next === pw.old) error.value = t('setup.err.same');
  if (error.value) {
    shake();
    return;
  }
  busy.value = true;
  try {
    await adminApi.changePassword(pw.old, pw.next, pw.code.trim());
    auth.markLoggedIn();
    void router.replace({ name: 'admin-today' });
  } catch (e) {
    error.value = explain(e);
    shake();
  } finally {
    busy.value = false;
  }
}

/* ---------- 报错抖动 ---------- */
const cardEl = ref<HTMLElement | null>(null);
function shake(): void {
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
  cardEl.value?.animate(
    [
      { transform: 'none' },
      { transform: 'translateX(-6px)', offset: 0.2 },
      { transform: 'translateX(6px)', offset: 0.4 },
      { transform: 'translateX(-4px)', offset: 0.6 },
      { transform: 'translateX(3px)', offset: 0.8 },
      { transform: 'none' },
    ],
    { duration: 420, easing: 'cubic-bezier(.2,.8,.3,1)', composite: 'add' },
  );
}

onMounted(() => {
  void adminApi.setupStatus().then(status => {
    canConfigureDatabase.value = status.canConfigureDatabase;
    currentDatabase.value = status.databaseDriver;
  }).catch(() => { /* Existing configuration loading already reports backend outages. */ });
  document.documentElement.dataset.studio = '';
  coverTimer = window.setInterval(() => { coverIdx.value = (coverIdx.value + 1) % COVERS.length; }, 5200);
});
onBeforeUnmount(() => {
  delete document.documentElement.dataset.studio;
  window.clearInterval(coverTimer);
});
</script>

<template>
  <main class="studio setup">
    <div ref="cardEl" class="card">
      <!-- 左：默认封面缓慢交叠轮换 + 当前步骤的一句话 -->
      <div class="art">
        <TransitionGroup name="cv" tag="div" class="covers">
          <img v-for="(c, i) in COVERS" v-show="i === coverIdx" :key="c" :src="`/covers/${c}.webp`" alt="" draggable="false" />
        </TransitionGroup>
        <div class="art-text">
          <img class="logo" src="/favicon-64.png" alt="" draggable="false" />
          <p>{{ changeMode ? t('setup.change.aside') : t(`setup.aside.${step === 'done' ? 'done' : step}`) }}</p>
          <small>Myself · {{ changeMode ? t('setup.change.kicker') : t('setup.brand') }}</small>
        </div>
      </div>

      <!-- 右：步骤表单 -->
      <div class="pane">
        <template v-if="changeMode">
          <form class="form" @submit.prevent="changePassword">
            <span class="kicker">{{ t('setup.change.kicker') }}</span>
            <h1>{{ t('setup.change.title') }}</h1>
            <p class="sub">{{ t('setup.change.sub') }}</p>
            <label><span class="st-flabel">{{ t('setup.change.code') }}<em>{{ t('setup.change.codeHint') }}</em></span><span class="st-field mono-in"><SIcon name="terminal" :size="16" /><input v-model="pw.code" maxlength="24" autocomplete="one-time-code" spellcheck="false" required placeholder="A1B2-C3D4-E5F6-A7B8" /></span></label>
            <label><span class="st-flabel">{{ t('setup.change.old') }}</span><span class="st-field"><SIcon name="lock" :size="16" /><input v-model="pw.old" type="password" autocomplete="current-password" required autofocus /></span></label>
            <label><span class="st-flabel">{{ t('setup.password') }}</span><span class="st-field"><SIcon name="key" :size="16" /><input v-model="pw.next" type="password" autocomplete="new-password" required /></span></label>
            <label><span class="st-flabel">{{ t('setup.confirm') }}</span><span class="st-field"><SIcon name="check" :size="16" /><input v-model="pw.confirm" type="password" autocomplete="new-password" required /></span></label>
            <p class="err" :class="{ on: !!error }" :role="error ? 'alert' : undefined" :aria-hidden="!error">{{ error || '&nbsp;' }}</p>
            <button type="submit" class="st-btn p lg go" :disabled="busy">{{ busy ? t('setup.working') : t('setup.change.submit') }}<SIcon name="arrowR" :size="16" /></button>
          </form>
        </template>

        <template v-else>
          <!-- 进度：四段，当前段实底 -->
          <div v-if="step !== 'done'" class="progress" :aria-label="t('setup.progress', { n: stepNo + 1, total: STEPS.length })">
            <i v-for="(s, i) in STEPS" :key="s" :class="{ on: i <= stepNo }" />
            <span>{{ stepNo + 1 }} / {{ STEPS.length }}</span>
          </div>

          <Transition :name="dir > 0 ? 'fwd' : 'bwd'" mode="out-in">
            <form v-if="step === 'code'" key="code" class="form" @submit.prevent="next">
              <span class="kicker">{{ t('setup.code.kicker') }}</span>
              <h1>{{ t('setup.code.title') }}</h1>
              <p class="sub">{{ t('setup.code.sub') }}</p>
              <pre class="log"><span>[myself-server]</span> {{ t('setup.code.logLine') }} <b>A1B2C3D4</b></pre>
              <label>
                <span class="st-flabel">{{ t('setup.code.label') }}</span>
                <span class="st-field mono-in"><SIcon name="terminal" :size="16" /><input v-model="form.code" maxlength="24" autocomplete="one-time-code" spellcheck="false" required autofocus placeholder="A1B2-C3D4-E5F6-A7B8" /></span>
              </label>
              <p class="err" :class="{ on: !!error }" :role="error ? 'alert' : undefined" :aria-hidden="!error">{{ error || '&nbsp;' }}</p>
              <button type="submit" class="st-btn p lg go" :disabled="busy">{{ busy ? t('setup.checking') : t('setup.next') }}<SIcon name="arrowR" :size="16" /></button>
            </form>

            <form v-else-if="step === 'database'" key="database" class="form" @submit.prevent="next">
              <span class="kicker">{{ t('setup.database.kicker') }}</span>
              <h1>{{ t('setup.database.title') }}</h1>
              <p class="sub">{{ t('setup.database.sub') }}</p>
              <label><span class="st-flabel">{{ t('setup.database.driver') }}</span><span class="st-field"><select v-model="databaseChoice" :disabled="busy"><option value="current">{{ t('setup.database.current', { driver: currentDatabase === 'mysql' ? 'MySQL' : 'SQLite' }) }}</option><option value="sqlite">SQLite</option><option value="mysql">MySQL</option></select></span></label>
              <template v-if="databaseChoice === 'mysql'">
                <div class="two">
                  <label><span class="st-flabel">{{ t('setup.database.host') }}</span><span class="st-field"><input v-model.trim="mysql.host" required autocomplete="off" /></span></label>
                  <label><span class="st-flabel">{{ t('setup.database.port') }}</span><span class="st-field"><input v-model.number="mysql.port" type="number" min="1" max="65535" required /></span></label>
                </div>
                <div class="two">
                  <label><span class="st-flabel">{{ t('setup.database.name') }}</span><span class="st-field"><input v-model.trim="mysql.name" required autocomplete="off" /></span></label>
                  <label><span class="st-flabel">{{ t('setup.database.user') }}</span><span class="st-field"><input v-model.trim="mysql.user" required autocomplete="off" /></span></label>
                </div>
                <label><span class="st-flabel">{{ t('setup.database.password') }}</span><span class="st-field"><input v-model="mysql.password" type="password" autocomplete="new-password" /></span></label>
                <label><span class="st-flabel">{{ t('setup.database.tls') }}</span><span class="st-field"><select v-model="mysql.tls"><option value="false">{{ t('setup.database.tlsOff') }}</option><option value="true">{{ t('setup.database.tlsOn') }}</option></select></span></label>
              </template>
              <p class="err" :class="{ on: !!error }" :role="error ? 'alert' : undefined" :aria-hidden="!error">{{ error || '&nbsp;' }}</p>
              <div class="nav"><button type="button" class="st-btn g lg" :disabled="busy" @click="back">{{ t('setup.back') }}</button><button type="submit" class="st-btn p lg go" :disabled="busy">{{ busy ? t('setup.database.connecting') : t('setup.next') }}<SIcon name="arrowR" :size="16" /></button></div>
            </form>

            <form v-else-if="step === 'admin'" key="admin" class="form" @submit.prevent="next">
              <span class="kicker">{{ t('setup.admin.kicker') }}</span>
              <h1>{{ t('setup.admin.title') }}</h1>
              <p class="sub">{{ t('setup.admin.sub') }}</p>
              <label><span class="st-flabel">{{ t('setup.username') }}</span><span class="st-field"><SIcon name="user" :size="16" /><input v-model="form.username" autocomplete="username" required autofocus /></span></label>
              <label>
                <span class="st-flabel">{{ t('setup.password') }}<em>{{ t('setup.passwordHint') }}</em></span>
                <span class="st-field">
                  <SIcon name="lock" :size="16" />
                  <input v-model="form.password" :type="show ? 'text' : 'password'" autocomplete="new-password" required />
                  <button type="button" class="eye" :title="show ? t('studio.login.hide') : t('studio.login.show')" @click="show = !show"><SIcon :name="show ? 'eyeOff' : 'eye'" :size="16" /></button>
                </span>
                <span class="meter" :data-s="strength"><i /><i /><i /><i /></span>
              </label>
              <label><span class="st-flabel">{{ t('setup.confirm') }}</span><span class="st-field"><SIcon name="check" :size="16" /><input v-model="form.confirm" :type="show ? 'text' : 'password'" autocomplete="new-password" required /></span></label>
              <p class="err" :class="{ on: !!error }" :role="error ? 'alert' : undefined" :aria-hidden="!error">{{ error || '&nbsp;' }}</p>
              <div class="nav">
                <button type="button" class="st-btn g lg" @click="back"><SIcon name="arrowL" :size="16" />{{ t('setup.back') }}</button>
                <button type="submit" class="st-btn p lg go">{{ t('setup.next') }}<SIcon name="arrowR" :size="16" /></button>
              </div>
            </form>

            <form v-else-if="step === 'site'" key="site" class="form" @submit.prevent="next">
              <span class="kicker">{{ t('setup.site.kicker') }}</span>
              <h1>{{ t('setup.site.title') }}</h1>
              <p class="sub">{{ t('setup.site.sub') }}</p>
              <div class="two">
                <label><span class="st-flabel">{{ t('setup.siteTitle') }}</span><span class="st-field"><input v-model="form.title" required autofocus /></span></label>
                <label><span class="st-flabel">{{ t('setup.name') }}</span><span class="st-field"><input v-model="form.name" :placeholder="t('setup.namePh')" required /></span></label>
              </div>
              <label><span class="st-flabel">{{ t('setup.tagline') }}<em>{{ t('setup.optional') }}</em></span><span class="st-field"><input v-model="form.tagline" :placeholder="t('setup.taglinePh')" /></span></label>
              <label><span class="st-flabel">{{ t('setup.motto') }}<em>{{ t('setup.optional') }}</em></span><span class="st-field"><input v-model="form.motto" /></span></label>
              <label>
                <span class="st-flabel">{{ t('setup.siteUrl') }}<em>{{ t('setup.siteUrlHint') }}</em></span>
                <span class="st-field mono-in"><SIcon name="globe" :size="16" /><input v-model="form.url" inputmode="url" placeholder="https://example.com" /></span>
              </label>
              <p class="err" :class="{ on: !!error }" :role="error ? 'alert' : undefined" :aria-hidden="!error">{{ error || '&nbsp;' }}</p>
              <div class="nav">
                <button type="button" class="st-btn g lg" @click="back"><SIcon name="arrowL" :size="16" />{{ t('setup.back') }}</button>
                <button type="submit" class="st-btn p lg go">{{ t('setup.next') }}<SIcon name="arrowR" :size="16" /></button>
              </div>
            </form>

            <form v-else-if="step === 'content'" key="content" class="form" @submit.prevent="next">
              <span class="kicker">{{ t('setup.content.kicker') }}</span>
              <h1>{{ t('setup.content.title') }}</h1>
              <p class="sub">{{ t('setup.content.sub') }}</p>
              <div class="choices" role="radiogroup">
                <button type="button" class="choice" role="radio" :aria-checked="form.demo" :class="{ on: form.demo }" @click="form.demo = true">
                  <span class="thumbs"><img src="/covers/03-s.webp" alt="" /><img src="/covers/01-s.webp" alt="" /><img src="/covers/10-s.webp" alt="" /></span>
                  <b>{{ t('setup.content.demo') }}</b>
                  <small>{{ t('setup.content.demoSub') }}</small>
                </button>
                <button type="button" class="choice" role="radio" :aria-checked="!form.demo" :class="{ on: !form.demo }" @click="form.demo = false">
                  <span class="thumbs blank"><i /><i /><i /></span>
                  <b>{{ t('setup.content.blank') }}</b>
                  <small>{{ t('setup.content.blankSub') }}</small>
                </button>
              </div>
              <p class="err" :class="{ on: !!error }" :role="error ? 'alert' : undefined" :aria-hidden="!error">{{ error || '&nbsp;' }}</p>
              <div class="nav">
                <button type="button" class="st-btn g lg" :disabled="busy" @click="back"><SIcon name="arrowL" :size="16" />{{ t('setup.back') }}</button>
                <button type="submit" class="st-btn p lg go" :disabled="busy">{{ busy ? t('setup.working') : t('setup.finish') }}<SIcon name="check" :size="16" /></button>
              </div>
            </form>

            <div v-else key="done" class="form done">
              <span class="badge"><SIcon name="check" :size="26" /></span>
              <h1>{{ t('setup.done.title') }}</h1>
              <p class="sub">{{ t('setup.done.sub', { name: form.name.trim() }) }}</p>
              <ul class="next-steps">
                <li><SIcon name="user" :size="16" />{{ t('setup.done.s1') }}</li>
                <li><SIcon name="palette" :size="16" />{{ t('setup.done.s2') }}</li>
                <li><SIcon name="pen" :size="16" />{{ t('setup.done.s3') }}</li>
              </ul>
              <div class="nav">
                <a class="st-btn g lg" href="/">{{ t('setup.done.site') }}</a>
                <router-link class="st-btn p lg go" :to="{ name: 'admin-identity' }">{{ t('setup.done.admin') }}<SIcon name="arrowR" :size="16" /></router-link>
              </div>
            </div>
          </Transition>
        </template>
      </div>
    </div>
  </main>
</template>

<style scoped lang="scss">
.setup {
  position: fixed;
  inset: 0;
  display: grid;
  place-items: center;
  padding: 24px;
  overflow: auto;
  background:
    radial-gradient(900px 600px at 80% -10%, color-mix(in oklab, var(--primary) 12%, transparent), transparent 60%),
    radial-gradient(700px 500px at 10% 110%, color-mix(in oklab, var(--primary) 7%, transparent), transparent 60%),
    var(--desk);
}

.card {
  display: grid;
  grid-template-columns: 400px 480px;
  min-height: 620px;
  border-radius: var(--r-xl);
  background: var(--paper);
  box-shadow: var(--sh-paper);
  overflow: hidden;
  animation: card-in var(--dur-slow) var(--ease-out) both;
}

@keyframes card-in { from { opacity: 0; transform: translateY(14px) scale(0.99); } }

/* ---------- 左：封面 ---------- */
.art {
  position: relative;
  overflow: hidden;
  background: #1b2436;

  .covers img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  /* 底部压暗，保证文字可读 */
  &::after {
    content: '';
    position: absolute;
    inset: 40% 0 0;
    background: linear-gradient(to top, rgb(10 14 22 / 0.72), transparent);
    pointer-events: none;
  }
}

.cv-enter-active { transition: opacity 1.6s ease, transform 6s ease-out; }
.cv-leave-active { transition: opacity 1.6s ease; }
.cv-enter-from { opacity: 0; transform: scale(1.04); }
.cv-leave-to { opacity: 0; }

.art-text {
  position: absolute;
  left: 30px;
  right: 30px;
  bottom: 28px;
  z-index: 1;
  color: #fff;

  .logo { width: 36px; height: 36px; margin-bottom: 16px; object-fit: cover; border-radius: var(--r-md); box-shadow: 0 8px 18px -8px rgb(0 0 0 / 0.6); }
  p { margin: 0 0 8px; font: 600 21px/1.5 var(--font-serif); letter-spacing: 0.03em; }
  small { font-size: 12px; letter-spacing: 0.08em; color: rgb(255 255 255 / 0.62); }
}

/* ---------- 右：步骤 ---------- */
.pane {
  position: relative;
  display: flex;
  flex-direction: column;
  padding: 36px 44px 34px;
  overflow: hidden;
}

.progress {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 26px;

  i {
    flex: 1;
    height: 3px;
    border-radius: var(--r-pill);
    background: var(--well-2);
    transition: background var(--dur-slow) var(--ease-out);

    &.on { background: var(--solid); }
  }

  span { margin-left: 10px; font: 500 12px var(--font-mono); color: var(--st-ink-3); }
}

.form {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 16px;

  label { display: block; }

  .kicker { font: 500 12px var(--font-mono); letter-spacing: 0.14em; color: var(--ink); text-transform: uppercase; }
  h1 { margin: -6px 0 0; font: 600 27px/1.3 var(--font-serif); }
  .sub { margin: -8px 0 4px; font-size: 13.5px; line-height: 1.7; color: var(--st-ink-3); }

  .st-flabel em { font-style: normal; font-weight: 400; color: var(--st-ink-4); }

  .two { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 12px; }
  .two > label { min-width: 0; }

  .eye {
    width: 28px;
    height: 28px;
    display: grid;
    place-items: center;
    border-radius: var(--r-xs);
    color: var(--st-ink-3);

    &:hover { background: var(--hover); color: var(--st-ink); }
  }

  .err {
    margin: -6px 0 -4px;
    min-height: 20px;
    font-size: 13px;
    color: var(--red);
    opacity: 0;
    transform: translateY(-4px);
    transition: all var(--dur) var(--ease-out);

    &.on { opacity: 1; transform: none; }
  }

  .nav {
    display: flex;
    gap: 10px;
    margin-top: auto;

    .go { flex: 1; }
  }

  > .go { margin-top: auto; }

  .go .st-ic { transition: transform var(--dur) var(--ease-spring); }
  .go:hover .st-ic { transform: translateX(3px); }
}

/* 启动日志示意：告诉用户初始化码长什么样、在哪儿 */
.log {
  margin: 0;
  padding: 12px 14px;
  border-radius: var(--r-md);
  background: var(--well);
  box-shadow: inset 0 0 0 1px var(--line);
  font: 12.5px/1.6 var(--font-mono);
  color: var(--st-ink-2);
  white-space: pre-wrap;
  word-break: break-all;

  span { color: var(--st-ink-4); }
  b { color: var(--ink); font-weight: 600; letter-spacing: 0.08em; }
}

.meter {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 4px;
  margin-top: 8px;

  i { height: 3px; border-radius: var(--r-pill); background: var(--well-2); transition: background var(--dur); }

  &[data-s='1'] i:nth-child(-n + 1) { background: var(--red); }
  &[data-s='2'] i:nth-child(-n + 2) { background: var(--yellow); }
  &[data-s='3'] i:nth-child(-n + 3) { background: color-mix(in oklab, var(--green) 70%, var(--yellow)); }
  &[data-s='4'] i { background: var(--green); }
}

/* 示例内容 / 空白开始：两张可选卡，选中抬升 + 轻染 */
.choices {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.choice {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
  padding: 12px 12px 14px;
  border-radius: var(--r-lg);
  text-align: left;
  background: var(--paper);
  box-shadow: inset 0 0 0 1px var(--line-2);
  transition: box-shadow var(--dur) var(--ease-out), background var(--dur) var(--ease-out), transform var(--dur-fast) var(--ease-spring);

  &:hover { background: var(--well); }
  &:active { transform: scale(0.98); }

  &.on {
    background: color-mix(in oklab, var(--ink) 6%, var(--paper));
    box-shadow: inset 0 0 0 1.5px color-mix(in oklab, var(--ink) 60%, transparent), 0 10px 24px -18px var(--st-shade);
  }

  b { margin-top: 4px; font-size: 14.5px; font-weight: 600; color: var(--st-ink); }
  small { font-size: 12.5px; line-height: 1.6; color: var(--st-ink-3); }
}

.thumbs {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 10;
  border-radius: var(--r-md);
  overflow: hidden;
  background: var(--well);

  img {
    position: absolute;
    width: 62%;
    aspect-ratio: 16 / 10;
    object-fit: cover;
    border-radius: calc(var(--r-md) - 2px);
    box-shadow: 0 8px 18px -10px rgb(0 0 0 / 0.5);

    &:nth-child(1) { left: 6%; top: 10%; rotate: -5deg; }
    &:nth-child(2) { left: 22%; top: 22%; rotate: 2deg; }
    &:nth-child(3) { left: 34%; top: 36%; rotate: 6deg; }
  }

  &.blank {
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 8px;
    padding: 0 16%;

    i { height: 6px; border-radius: var(--r-pill); background: var(--well-2); }
    i:nth-child(1) { width: 70%; }
    i:nth-child(3) { width: 45%; }
  }
}

/* 完成 */
.done {
  justify-content: center;

  .badge {
    width: 56px;
    height: 56px;
    display: grid;
    place-items: center;
    border-radius: 50%;
    color: var(--on-solid);
    background: var(--solid);
    animation: pop 0.5s var(--ease-spring) both;
  }

  h1 { margin-top: 6px; }
}

@keyframes pop { from { opacity: 0; transform: scale(0.5); } }

.next-steps {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 4px 0 12px;
  padding: 16px;
  list-style: none;
  border-radius: var(--r-md);
  background: var(--well);

  li { display: flex; align-items: center; gap: 10px; font-size: 13.5px; color: var(--st-ink-2); }
  .st-ic { color: var(--ink); }
}

/* 步骤切换：前进从右滑入，后退从左滑入 */
.fwd-enter-active, .fwd-leave-active, .bwd-enter-active, .bwd-leave-active { transition: opacity 0.28s var(--ease-out), transform 0.28s var(--ease-out); }
.fwd-enter-from, .bwd-leave-to { opacity: 0; transform: translateX(28px); }
.fwd-leave-to, .bwd-enter-from { opacity: 0; transform: translateX(-28px); }

@media (max-width: 960px) {
  .card { grid-template-columns: 1fr; width: min(500px, 100%); min-height: 0; }
  .art { height: 180px; }
  .art-text { bottom: 18px; p { font-size: 17px; } .logo { display: none; } }
  .pane { padding: 28px 24px 26px; }
}

@media (max-width: 480px) {
  .setup { padding: 0; place-items: stretch; }
  .card { border-radius: 0; width: 100%; }
  .form .two, .choices { grid-template-columns: 1fr; }
}

@media (prefers-reduced-motion: reduce) {
  .card, .done .badge { animation: none; }
  .fwd-enter-active, .fwd-leave-active, .bwd-enter-active, .bwd-leave-active, .cv-enter-active, .cv-leave-active { transition: none; }
}
</style>
