<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { ArrowLeft, Check, Eye, EyeOff, Mail, MailCheck, PenLine } from 'lucide';
import { siGithub } from 'simple-icons';
import { accountApi, type SessionUser } from '../api';
import { useAuthStore } from '../stores/auth';
import { useConfigStore } from '../stores/config';
import { useIdentity } from '../about/useIdentity';
import { useSiteLogo } from '../utils/siteLogo';
import Icon from '../components/ui/Icon.vue';

/**
 * 前台账号：登录 / 注册 / 邀请注册（/account/join?code=）/ 忘记密码 / 重置密码（?token=）/ 邮箱验证（?token=）。
 * 同一张卡片按路由切换表单，卡片高度随内容平滑过渡。登录后回到 next（站内路径），否则回首页。
 */
type Mode = 'login' | 'register' | 'join' | 'forgot' | 'reset' | 'verify';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const config = useConfigStore();
const { fullName } = useIdentity();
const { logo } = useSiteLogo();

const mode = computed<Mode>(() => String(route.name ?? 'account-login').replace('account-', '') as Mode);
const users = computed(() => config.cfg.users);
const usersOn = computed(() => !!users.value?.enabled);
const signupOpen = computed(() => !!users.value?.readers.enabled && users.value.readers.signup === 'open');
const githubOn = computed(() => !!users.value?.login.github);
const mailReset = computed(() => !!users.value?.login.mailReset);

/** 登录后的落点：只接受站内相对路径 */
const next = computed(() => {
  const n = typeof route.query.next === 'string' ? route.query.next : '';
  return n.startsWith('/') && !n.startsWith('//') ? n : '/';
});
const q = (k: string): string => (typeof route.query[k] === 'string' ? (route.query[k] as string) : '');

const form = reactive({ login: '', email: '', name: '', password: '', confirm: '', website: '' });
const showPw = ref(false);
const busy = ref(false);
const error = ref('');
/** 完成态：注册待验证 / 找回邮件已发 / 验证中 / 验证成功 */
const done = ref<'' | 'pending' | 'sent' | 'verifying' | 'verified' | 'joined'>('');
const invite = ref<{ role: 'author' | 'reader'; email: string } | null>(null);
/** 验证链接是一次换绑邮箱（而不是注册验证） */
const emailChanged = ref(false);
const resetEmail = ref('');
const cardEl = ref<HTMLElement | null>(null);

const ERR: Record<string, string> = {
  bad_credentials: 'account.err.badCredentials',
  too_many_attempts: 'account.err.tooMany',
  pending_verification: 'account.err.pending',
  account_disabled: 'account.err.disabled',
  role_closed: 'account.err.roleClosed',
  invalid_email: 'account.err.email',
  invalid_name: 'account.err.name',
  weak_password: 'account.err.weak',
  email_taken: 'account.err.emailTaken',
  signup_closed: 'account.err.signupClosed',
  invalid_invite: 'account.err.invite',
  invalid_token: 'account.err.token',
  mail_unavailable: 'account.err.noMail',
  github_off: 'account.err.githubOff',
  github_denied: 'account.err.githubDenied',
  github_failed: 'account.err.githubFailed',
  github_taken: 'account.err.githubTaken',
  state_expired: 'account.err.state',
  state_mismatch: 'account.err.state',
  login_required: 'account.err.loginRequired',
};
const explain = (code: string): string => t(ERR[code] ?? 'account.err.generic');

function shake(): void {
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
  cardEl.value?.animate(
    [{ transform: 'none' }, { transform: 'translateX(-6px)' }, { transform: 'translateX(6px)' }, { transform: 'translateX(-4px)' }, { transform: 'none' }],
    { duration: 380, easing: 'cubic-bezier(.2,.8,.3,1)' },
  );
}

function fail(e: unknown): void {
  error.value = explain((e as Error).message);
  shake();
}

function finish(user: SessionUser, mustChange?: boolean): void {
  auth.markLoggedIn(user);
  if (mustChange) void router.replace({ path: '/setup', query: { change: '1' } });
  else void router.replace(next.value);
}

async function submit(): Promise<void> {
  if (busy.value) return;
  error.value = '';
  const m = mode.value;
  if ((m === 'register' || m === 'join' || m === 'reset') && form.password.length < 8) return fail(new Error('weak_password'));
  if (m === 'reset' && form.password !== form.confirm) {
    error.value = t('account.err.mismatch');
    shake();
    return;
  }
  busy.value = true;
  try {
    if (m === 'login') {
      const res = await accountApi.login(form.login.trim(), form.password);
      finish(res.user, res.mustChange);
    } else if (m === 'register' || m === 'join') {
      const res = await accountApi.register({
        email: form.email.trim(),
        name: form.name.trim(),
        password: form.password,
        invite: m === 'join' ? q('code') : undefined,
        website: form.website,
      });
      if (res.pending || !res.user) done.value = 'pending';
      else if (res.user.role === 'author') {
        // 受邀成为协作作者：留在本页说明身份，并给出进入写作后台的入口
        auth.markLoggedIn(res.user);
        done.value = 'joined';
      } else finish(res.user);
    } else if (m === 'forgot') {
      await accountApi.forgot(form.email.trim());
      done.value = 'sent';
    } else if (m === 'reset') {
      const res = await accountApi.reset(q('token'), form.password);
      finish(res.user);
    }
  } catch (e) {
    fail(e);
  } finally {
    busy.value = false;
  }
}

/** 进入各模式时的准备：邀请信息、重置邮箱、邮箱验证、OAuth 回跳错误 */
async function prepare(): Promise<void> {
  error.value = '';
  done.value = '';
  form.password = form.confirm = '';
  if (q('error')) error.value = explain(q('error'));
  if (mode.value === 'join') {
    invite.value = null;
    try {
      invite.value = await accountApi.invite(q('code'));
      if (invite.value.email) form.email = invite.value.email;
    } catch (e) {
      fail(e);
    }
  } else if (mode.value === 'reset') {
    try {
      resetEmail.value = (await accountApi.resetInfo(q('token'))).email;
    } catch (e) {
      fail(e);
    }
  } else if (mode.value === 'verify') {
    done.value = 'verifying';
    try {
      const res = await accountApi.verify(q('token'));
      auth.markLoggedIn(res.user);
      emailChanged.value = !!res.emailChanged;
      done.value = 'verified';
    } catch (e) {
      done.value = '';
      fail(e);
    }
  }
}
onMounted(prepare);
watch(() => route.fullPath, prepare);

const githubHref = computed(() => accountApi.githubUrl({ mode: 'login', next: next.value, invite: mode.value === 'join' ? q('code') : undefined }));

const title = computed(() => {
  if (done.value === 'pending' || done.value === 'sent') return t('account.checkMail');
  if (done.value === 'verified') return emailChanged.value ? t('account.emailChanged') : t('account.verified');
  if (done.value === 'joined') return t('account.joined');
  if (mode.value === 'join' && invite.value) return t('account.joinTitle', { role: t(`account.role_${invite.value.role}`) });
  return t(`account.title_${mode.value}`);
});
const sub = computed(() => {
  if (done.value === 'pending') return t('account.pendingSub', { email: form.email });
  if (done.value === 'sent') return t('account.sentSub');
  if (done.value === 'verifying') return t('account.verifying');
  if (done.value === 'verified') return emailChanged.value ? t('account.emailChangedSub') : t('account.verifiedSub');
  if (done.value === 'joined') return t('account.joinedSub', { site: config.cfg.site.title });
  if (mode.value === 'reset' && resetEmail.value) return t('account.resetSub', { email: resetEmail.value });
  if (mode.value === 'join' && invite.value) return t('account.joinSub', { site: config.cfg.site.title });
  return t(`account.sub_${mode.value}`, { site: config.cfg.site.title });
});

/** 表单是否可用：注册要开放注册；登录要用户系统开启（站长始终可以从后台入口登录） */
const closed = computed(() => {
  if (mode.value === 'register') return !usersOn.value || !signupOpen.value;
  if (mode.value === 'forgot') return !mailReset.value;
  return false;
});
const showForm = computed(() => !done.value && !closed.value && (mode.value !== 'join' || !!invite.value) && mode.value !== 'verify');
const nextQuery = computed(() => (next.value !== '/' ? { next: next.value } : {}));
</script>

<template>
  <main class="acc">
    <div ref="cardEl" class="card rise">
      <router-link to="/" class="brand">
        <img :src="logo" alt="" draggable="false" />
        <span>{{ config.cfg.site.title || fullName }}</span>
      </router-link>

      <Transition name="swap" mode="out-in">
        <div :key="`${mode}-${done}-${!!invite}`" class="pane">
          <span v-if="done === 'pending' || done === 'sent'" class="badge"><Icon :icon="Mail" :size="22" /></span>
          <span v-else-if="done === 'verified'" class="badge ok"><Icon :icon="MailCheck" :size="22" /></span>
          <span v-else-if="done === 'joined'" class="badge ok"><Icon :icon="PenLine" :size="22" /></span>
          <h1>{{ title }}</h1>
          <p class="sub">{{ sub }}</p>

          <p v-if="closed" class="closed">
            {{ mode === 'forgot' ? t('account.noMailHint') : t('account.signupClosed') }}
          </p>

          <form v-if="showForm" class="form" novalidate @submit.prevent="submit">
            <template v-if="mode === 'login'">
              <label class="field">
                <span>{{ t('account.loginId') }}</span>
                <input v-model="form.login" autocomplete="username" spellcheck="false" required />
              </label>
            </template>

            <template v-if="mode === 'register' || mode === 'join'">
              <label class="field">
                <span>{{ t('account.name') }}</span>
                <input v-model="form.name" autocomplete="nickname" maxlength="40" required />
              </label>
              <label class="field">
                <span>{{ t('account.email') }}</span>
                <input v-model="form.email" type="email" autocomplete="email" spellcheck="false" :readonly="!!invite?.email" required />
              </label>
              <!-- 蜜罐：对真人隐藏 -->
              <input v-model="form.website" class="hp" tabindex="-1" autocomplete="off" aria-hidden="true" name="website" />
            </template>

            <template v-if="mode === 'forgot'">
              <label class="field">
                <span>{{ t('account.email') }}</span>
                <input v-model="form.email" type="email" autocomplete="email" spellcheck="false" required />
              </label>
            </template>

            <template v-if="mode !== 'forgot'">
              <label class="field">
                <span class="lb">
                  {{ mode === 'reset' ? t('account.newPassword') : t('account.password') }}
                  <router-link v-if="mode === 'login' && mailReset" :to="{ path: '/account/forgot', query: nextQuery }" class="aux">{{ t('account.forgot') }}</router-link>
                </span>
                <span class="pw">
                  <input
                    v-model="form.password"
                    :aria-label="mode === 'reset' ? t('account.newPassword') : t('account.password')"
                    :type="showPw ? 'text' : 'password'"
                    :autocomplete="mode === 'login' ? 'current-password' : 'new-password'"
                    :placeholder="mode === 'login' ? '' : t('account.pwHint')"
                    required
                  />
                  <button type="button" class="eye" :aria-label="showPw ? t('account.hidePw') : t('account.showPw')" @click="showPw = !showPw">
                    <Icon :icon="showPw ? EyeOff : Eye" :size="17" />
                  </button>
                </span>
              </label>
              <label v-if="mode === 'reset'" class="field">
                <span>{{ t('account.confirmPassword') }}</span>
                <input v-model="form.confirm" :type="showPw ? 'text' : 'password'" autocomplete="new-password" required />
              </label>
            </template>

            <Transition name="err"><p v-if="error" class="err" role="alert">{{ error }}</p></Transition>

            <button type="submit" class="primary" :disabled="busy">
              {{ busy ? t('account.busy') : t(`account.submit_${mode}`) }}
            </button>

            <template v-if="githubOn && (mode === 'login' || mode === 'register' || mode === 'join')">
              <div class="or"><span>{{ t('account.or') }}</span></div>
              <a class="gh" :href="githubHref">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" aria-hidden="true"><path :d="siGithub.path" /></svg>
                {{ t('account.withGithub') }}
              </a>
            </template>
          </form>

          <Transition name="err"><p v-if="error && !showForm" class="err" role="alert">{{ error }}</p></Transition>

          <div class="links">
            <template v-if="mode === 'login'">
              <span v-if="signupOpen">{{ t('account.noAccount') }}<router-link :to="{ path: '/account/register', query: nextQuery }">{{ t('account.goRegister') }}</router-link></span>
            </template>
            <template v-else-if="done === 'joined'">
              <router-link class="primary as-link" :to="{ name: 'admin-posts' }"><Icon :icon="PenLine" :size="16" />{{ t('account.goStudio') }}</router-link>
              <router-link to="/">{{ t('account.goHome') }}</router-link>
            </template>
            <template v-else-if="done === 'verified'">
              <router-link class="primary as-link" :to="next"><Icon :icon="Check" :size="16" />{{ t('account.continue') }}</router-link>
            </template>
            <template v-else>
              <router-link :to="{ path: '/account/login', query: nextQuery }"><Icon :icon="ArrowLeft" :size="15" />{{ t('account.backLogin') }}</router-link>
            </template>
          </div>
        </div>
      </Transition>
    </div>
  </main>
</template>

<style scoped lang="scss">
.acc {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 104px 20px 72px;
}

.card {
  width: min(420px, 100%);
  padding: 30px 30px 26px;
  border-radius: var(--r-xl);
  background: var(--elev);
  box-shadow: inset 0 0 0 0.5px var(--line-2), 0 30px 60px -40px rgb(0 0 0 / 0.45);
}

/* 简洁风格：不要卡片面，表单直接落在页面上 */
:root[data-style='clean'] .card {
  padding: 8px 4px;
  background: transparent;
  box-shadow: none;
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 22px;
  color: var(--text-2);
  font-size: 14px;

  img { width: 30px; height: 30px; border-radius: var(--r-md); object-fit: cover; }
  &:hover { color: var(--text); }
}

.badge {
  display: grid;
  place-items: center;
  width: 46px;
  height: 46px;
  margin-bottom: 14px;
  border-radius: var(--r-md);
  background: color-mix(in oklab, var(--ink) 12%, transparent);
  color: var(--ink);

  &.ok { background: color-mix(in oklab, #22c55e 14%, transparent); color: #16a34a; }
}

h1 { font: 700 26px/1.3 var(--font-serif); letter-spacing: 0.01em; }
.sub { margin-top: 8px; font-size: 14.5px; line-height: 1.7; color: var(--text-2); }
.closed { margin-top: 18px; padding: 12px 14px; border-radius: var(--r-md); background: var(--fill); font-size: 14px; line-height: 1.7; color: var(--text-2); }

.form { display: flex; flex-direction: column; gap: 14px; margin-top: 22px; }

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;

  > span:first-child, .lb { font-size: 13px; color: var(--text-2); }
  .lb { display: flex; justify-content: space-between; }
  .aux { color: var(--text-3); font-size: 12.5px; &:hover { color: var(--ink); } }

  input, .pw {
    height: 44px;
    border-radius: var(--r-md);
    background: var(--fill);
    box-shadow: inset 0 0 0 1px var(--line);
    transition: box-shadow var(--dur-fast), background var(--dur-fast);
  }

  input {
    width: 100%;
    padding: 0 14px;
    border: 0;
    outline: 0;
    color: var(--text);
    font: 15px var(--font-sans);

    &:focus { background: var(--bg); box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--ink) 50%, transparent), 0 0 0 3px color-mix(in oklab, var(--ink) 14%, transparent); }
    &[readonly] { color: var(--text-2); }
  }

  .pw {
    display: flex;
    align-items: center;

    input { height: 100%; background: none; box-shadow: none; }
    &:focus-within { background: var(--bg); box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--ink) 50%, transparent), 0 0 0 3px color-mix(in oklab, var(--ink) 14%, transparent); }
  }

  .eye { display: grid; place-items: center; width: 42px; height: 100%; border: 0; background: none; color: var(--text-3); cursor: pointer; &:hover { color: var(--text); } }
}

.hp { position: absolute; left: -9999px; width: 1px; height: 1px; opacity: 0; }

.err { font-size: 13.5px; line-height: 1.6; color: var(--accent-red); }
.pane > .err { margin-top: 16px; }

.primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 44px;
  border: 0;
  border-radius: var(--r-pill);
  background: var(--solid);
  color: var(--on-solid);
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: transform var(--dur-fast) var(--ease-spring), opacity var(--dur-fast);

  &:active:not(:disabled) { transform: scale(0.98); }
  &:disabled { opacity: 0.6; cursor: progress; }
  &.as-link { width: 100%; }
}

.or {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12.5px;
  color: var(--text-3);

  &::before, &::after { content: ''; flex: 1; height: 0.5px; background: var(--line-2); }
}

.gh {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  height: 44px;
  border-radius: var(--r-pill);
  background: var(--fill);
  box-shadow: inset 0 0 0 1px var(--line);
  color: var(--text);
  font-size: 14.5px;
  font-weight: 500;
  transition: background var(--dur-fast);

  &:hover { background: color-mix(in oklab, var(--text) 8%, transparent); }
}

.links {
  display: flex;
  justify-content: center;
  margin-top: 20px;
  font-size: 13.5px;
  color: var(--text-3);

  a { display: inline-flex; align-items: center; gap: 4px; color: var(--ink); font-weight: 500; }
  .primary.as-link { color: var(--on-solid); }
}

.swap-enter-active, .swap-leave-active { transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-out); }
.swap-enter-from { opacity: 0; transform: translateY(8px); }
.swap-leave-to { opacity: 0; transform: translateY(-6px); }
.err-enter-active, .err-leave-active { transition: opacity var(--dur-fast), transform var(--dur-fast); }
.err-enter-from, .err-leave-to { opacity: 0; transform: translateY(-4px); }

@media (max-width: 767px) {
  /* 移动端：卡片在可视区（去掉底栏占位）内垂直居中 */
  .acc { min-height: 100dvh; padding: 24px 16px calc(104px + env(safe-area-inset-bottom)); place-items: center; }
  .card { padding: 24px 20px 22px; }
}
</style>
