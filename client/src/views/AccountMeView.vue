<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { AtSign, Camera, Check, Clock, KeyRound, LayoutDashboard, LogOut, Mail } from 'lucide';
import { siGithub } from 'simple-icons';
import { accountApi, type SessionUser } from '../api';
import { useAuthStore } from '../stores/auth';
import { useConfigStore } from '../stores/config';
import { useDialogStore } from '../stores/dialog';
import Icon from '../components/ui/Icon.vue';
import AvatarCropper from '../components/ui/AvatarCropper.vue';

/**
 * 我的账号（/account）：
 * - 头像：点击上传 → 裁切 → 站长直接生效，其他人进入待审（可撤回）；站长默认用「身份」头像，也可单独设置 / 恢复
 * - 资料：昵称
 * - 账号与登录：登录名（站长随时可改，其他角色每 30 天一次）、邮箱（可改，改后重新验证；登录名与邮箱都能登录）、密码、GitHub 绑定
 * 区块随界面风格：cards 为卡片，clean 为透明底 + 发丝线分隔。
 */
const { t } = useI18n();
const router = useRouter();
const auth = useAuthStore();
const config = useConfigStore();
const dialog = useDialogStore();

const user = computed(() => auth.user);
const isOwner = computed(() => user.value?.role === 'admin');
const githubOn = computed(() => !!config.cfg.users?.login.github);
const mailReady = computed(() => !!config.cfg.users?.login.mailReset);
const ROLE = computed(() => t(`account.role_${user.value?.role ?? 'reader'}`));
const initial = computed(() => (user.value?.name || user.value?.login || '?').trim()[0]?.toUpperCase() ?? '?');

onMounted(async () => {
  await auth.sync();
  if (!auth.loggedIn || !auth.user) {
    void router.replace({ path: '/account/login', query: { next: '/account' } });
    return;
  }
  name.value = auth.user.name;
});

/** 服务端返回最新资料后同步到会话 */
function apply(u: SessionUser): void {
  auth.setUser(u);
}

/* ---------- 通知条：每个区块各自一条，几秒后淡出 ---------- */
type Note = { ok: boolean; text: string } | null;
const notes = reactive<Record<'avatar' | 'profile' | 'account', Note>>({ avatar: null, profile: null, account: null });
const timers: Record<string, number> = {};
function notify(where: keyof typeof notes, ok: boolean, text: string): void {
  notes[where] = { ok, text };
  window.clearTimeout(timers[where]);
  // 成功提示几秒后自动收起；错误一直留着，直到下一次操作或重新提交
  if (ok) timers[where] = window.setTimeout(() => (notes[where] = null), 4200);
}

const ERR: Record<string, string> = {
  bad_credentials: 'account.me.pwWrong',
  too_many_attempts: 'account.err.tooMany',
  invalid_name: 'account.err.name',
  invalid_email: 'account.err.email',
  email_taken: 'account.err.emailTaken',
  invalid_login: 'account.me.loginRule',
  login_taken: 'account.me.loginTaken',
  login_cooldown: 'account.me.loginCooldown',
  weak_password: 'account.err.weak',
  invalid_image: 'account.me.badImage',
  file_too_large: 'account.me.tooLarge',
  mail_unavailable: 'account.me.emailNoMail',
  site_url_required: 'account.me.emailNoSiteUrl',
  image_too_large: 'account.me.tooLarge',
  mail_failed: 'account.me.mailFailed',
};
const explain = (e: unknown): string => t(ERR[(e as Error).message] ?? 'account.err.generic');

/* ---------- 头像 ---------- */
const fileInput = ref<HTMLInputElement | null>(null);
const cropping = ref<File | null>(null);
const uploading = ref(false);

function pick(): void {
  fileInput.value?.click();
}
function onFile(e: Event): void {
  const input = e.target as HTMLInputElement;
  const f = input.files?.[0];
  input.value = '';
  if (!f) return;
  if (!f.type.startsWith('image/')) return notify('avatar', false, t('account.me.badImage'));
  if (f.size > 20 * 1024 * 1024) return notify('avatar', false, t('account.me.tooLarge'));
  cropping.value = f;
}
async function upload(blob: Blob): Promise<void> {
  uploading.value = true;
  try {
    const res = await accountApi.uploadAvatar(blob);
    apply(res.user);
    cropping.value = null;
    notify('avatar', true, res.user.avatarPending ? t('account.me.avatarSubmitted') : t('account.me.avatarSaved'));
  } catch (e) {
    notify('avatar', false, explain(e));
  } finally {
    uploading.value = false;
  }
}
async function withdraw(): Promise<void> {
  try {
    apply((await accountApi.deleteAvatar(true)).user);
  } catch (e) {
    notify('avatar', false, explain(e));
  }
}
async function removeAvatar(): Promise<void> {
  const ok = await dialog.confirm({
    title: isOwner.value ? t('account.me.useIdentityTitle') : t('account.me.removeAvatarTitle'),
    message: isOwner.value ? t('account.me.useIdentityBody') : t('account.me.removeAvatarBody'),
    confirmText: isOwner.value ? t('account.me.useIdentity') : t('account.me.removeAvatar'),
  });
  if (!ok) return;
  try {
    apply((await accountApi.deleteAvatar()).user);
  } catch (e) {
    notify('avatar', false, explain(e));
  }
}

/* ---------- 资料 ---------- */
const name = ref('');
const savingName = ref(false);
const nameDirty = computed(() => !!user.value && name.value.trim() !== user.value.name);
async function saveName(): Promise<void> {
  if (savingName.value || !nameDirty.value) return;
  savingName.value = true;
  try {
    apply((await accountApi.updateMe(name.value.trim())).user);
    notify('profile', true, t('account.me.saved'));
  } catch (e) {
    notify('profile', false, explain(e));
  } finally {
    savingName.value = false;
  }
}

/* ---------- 账号与登录：同一时间只展开一个编辑框 ---------- */
type Editing = 'login' | 'email' | 'password' | null;
const editing = ref<Editing>(null);
const busy = ref(false);
const form = reactive({ login: '', email: '', emailPw: '', old: '', next: '', confirm: '' });
/** 账号区块：展开编辑框后聚焦其中第一个输入框 */
const accountSec = ref<HTMLElement | null>(null);

async function edit(which: Editing): Promise<void> {
  editing.value = editing.value === which ? null : which;
  notes.account = null;
  Object.assign(form, { login: user.value?.login ?? '', email: '', emailPw: '', old: '', next: '', confirm: '' });
  await nextTick();
  accountSec.value?.querySelector<HTMLInputElement>('.fold.open input')?.focus();
}

const loginLocked = computed(() => !isOwner.value && !!user.value?.loginNextChange);
const nextChangeDate = computed(() => {
  const v = user.value?.loginNextChange;
  if (!v) return '';
  const d = new Date(`${v.replace(' ', 'T')}Z`);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
});

async function run(fn: () => Promise<string>): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  try {
    const msg = await fn();
    editing.value = null;
    // 成功后清空表单（尤其是密码框），收起的编辑框里不留旧值
    Object.assign(form, { email: '', emailPw: '', old: '', next: '', confirm: '' });
    notify('account', true, msg);
  } catch (e) {
    notify('account', false, explain(e));
  } finally {
    busy.value = false;
  }
}

function saveLogin(): Promise<void> {
  return run(async () => {
    apply((await accountApi.changeLogin(form.login.trim())).user);
    if (isOwner.value) await config.load();
    return t('account.me.loginSaved');
  });
}

function saveEmail(): Promise<void> {
  return run(async () => {
    const res = await accountApi.changeEmail(form.email.trim(), form.emailPw);
    apply(res.user);
    return t('account.me.emailSent', { email: res.user.emailPending ?? form.email.trim() });
  });
}

async function cancelEmail(): Promise<void> {
  try {
    apply((await accountApi.cancelEmailChange()).user);
  } catch (e) {
    notify('account', false, explain(e));
  }
}

function savePassword(): Promise<void> {
  if (form.next.length < 8) return Promise.resolve(notify('account', false, t('account.err.weak')));
  if (form.next !== form.confirm) return Promise.resolve(notify('account', false, t('account.err.mismatch')));
  return run(async () => {
    await accountApi.changePassword(form.old, form.next);
    await auth.sync();
    return t('account.me.pwChanged');
  });
}

async function logout(): Promise<void> {
  const ok = await dialog.confirm({ title: t('account.me.logoutTitle'), message: t('account.me.logoutBody'), confirmText: t('account.me.logout') });
  if (!ok) return;
  await auth.logout();
  void router.replace('/');
}

const linkHref = computed(() => accountApi.githubUrl({ mode: 'link', next: '/account' }));
</script>

<template>
  <main class="me">
    <template v-if="user">
      <!-- 头像 + 名字 -->
      <section class="head rise">
        <button type="button" class="av" :title="t('account.me.changeAvatar')" :aria-label="t('account.me.changeAvatar')" @click="pick">
          <img v-if="auth.avatar" :src="auth.avatar" alt="" referrerpolicy="no-referrer" />
          <span v-else class="ini" aria-hidden="true">{{ initial }}</span>
          <span class="cam"><Icon :icon="Camera" :size="20" /></span>
        </button>
        <input ref="fileInput" type="file" accept="image/*" hidden @change="onFile" />
        <div class="who">
          <h1>{{ user.name || user.login }}</h1>
          <p><span class="role" :class="user.role">{{ ROLE }}</span><span class="mono">{{ user.login.includes('@') ? user.login : `@${user.login}` }}</span></p>
          <p class="av-line">
            <template v-if="isOwner && user.avatarDefault">{{ t('account.me.usingIdentity') }}</template>
            <button v-else-if="user.avatar && !user.avatarDefault" type="button" class="link" @click="removeAvatar">
              {{ isOwner ? t('account.me.useIdentity') : t('account.me.removeAvatar') }}
            </button>
            <template v-else>{{ t('account.me.avatarHint') }}</template>
          </p>
        </div>
        <router-link v-if="auth.staff" to="/admin" class="btn ghost"><Icon :icon="LayoutDashboard" :size="16" />{{ t('account.me.admin') }}</router-link>
      </section>

      <Transition name="fold">
        <div v-if="user.avatarPending" class="pending rise">
          <img :src="user.avatarPending" alt="" />
          <div class="tx">
            <b><Icon :icon="Clock" :size="14" />{{ t('account.me.avatarPending') }}</b>
            <small>{{ t('account.me.avatarPendingSub') }}</small>
          </div>
          <button type="button" class="link" @click="withdraw">{{ t('account.me.withdraw') }}</button>
        </div>
      </Transition>
      <Transition name="nt"><p v-if="notes.avatar" class="notice top" :class="{ ok: notes.avatar.ok }">{{ notes.avatar.text }}</p></Transition>

      <!-- 资料 -->
      <section class="sec rise" style="--i: 1">
        <h2>{{ t('account.me.profile') }}</h2>
        <label class="field">
          <span>{{ t('account.name') }}</span>
          <span class="inline">
            <input v-model="name" maxlength="40" autocomplete="nickname" @keydown.enter="saveName" />
            <button type="button" class="btn" :disabled="savingName || !nameDirty" @click="saveName">{{ t('account.me.save') }}</button>
          </span>
        </label>
        <Transition name="nt"><p v-if="notes.profile" class="notice" :class="{ ok: notes.profile.ok }">{{ notes.profile.text }}</p></Transition>
      </section>

      <!-- 账号与登录 -->
      <section ref="accountSec" class="sec rise" style="--i: 2">
        <h2>{{ t('account.me.signin') }}</h2>
        <p class="desc">{{ t('account.me.signinDesc') }}</p>

        <!-- 登录名 -->
        <div class="row" :class="{ open: editing === 'login' }">
          <span class="ic"><Icon :icon="AtSign" :size="17" /></span>
          <div class="tx">
            <b>{{ t('account.me.login') }}</b>
            <span class="val mono">{{ user.login }}</span>
            <small>{{ isOwner ? t('account.me.loginOwnerHint') : loginLocked ? t('account.me.loginNext', { date: nextChangeDate }) : t('account.me.loginHint') }}</small>
          </div>
          <button type="button" class="btn ghost sm" :disabled="loginLocked" @click="edit('login')">
            {{ editing === 'login' ? t('account.me.cancel') : t('account.me.edit') }}
          </button>
        </div>
        <div class="fold" :class="{ open: editing === 'login' }" :inert="editing !== 'login'" :aria-hidden="editing !== 'login'">
          <div class="clip">
            <form class="editor" @submit.prevent="saveLogin">
              <label class="field">
                <span>{{ t('account.me.newLogin') }}<em>{{ t('account.me.loginRule') }}</em></span>
                <input v-model="form.login" maxlength="24" spellcheck="false" autocomplete="username" />
              </label>
              <p class="warn">{{ t(isOwner ? 'account.me.loginOwnerWarn' : 'account.me.loginWarn') }}</p>
              <div class="ft"><Transition name="nt"><p v-if="notes.account && !notes.account.ok" class="notice in" role="alert">{{ notes.account.text }}</p></Transition><button type="submit" class="btn" :disabled="busy">{{ t('account.me.save') }}</button></div>
            </form>
          </div>
        </div>

        <!-- 邮箱 -->
        <div class="row" :class="{ open: editing === 'email' }">
          <span class="ic"><Icon :icon="Mail" :size="17" /></span>
          <div class="tx">
            <b>{{ t('account.email') }}</b>
            <span class="val">
              {{ user.email || t('account.me.noEmail') }}
              <em v-if="user.email" class="chip" :class="{ ok: user.emailVerified }">{{ user.emailVerified ? t('account.me.verified') : t('account.me.unverified') }}</em>
            </span>
            <small v-if="user.emailPending" class="pend">
              {{ t('account.me.emailPending', { email: user.emailPending }) }}
              <button type="button" class="link" @click="cancelEmail">{{ t('account.me.cancelChange') }}</button>
            </small>
            <small v-else>{{ t('account.me.emailHint') }}</small>
          </div>
          <button type="button" class="btn ghost sm" @click="edit('email')">
            {{ editing === 'email' ? t('account.me.cancel') : user.email ? t('account.me.edit') : t('account.me.add') }}
          </button>
        </div>
        <div class="fold" :class="{ open: editing === 'email' }" :inert="editing !== 'email'" :aria-hidden="editing !== 'email'">
          <div class="clip">
            <form class="editor" @submit.prevent="saveEmail">
              <div class="two">
                <label class="field">
                  <span>{{ t('account.me.newEmail') }}</span>
                  <input v-model="form.email" type="email" autocomplete="email" spellcheck="false" />
                </label>
                <label v-if="user.hasPassword" class="field">
                  <span>{{ t('account.me.oldPw') }}</span>
                  <input v-model="form.emailPw" type="password" autocomplete="current-password" />
                </label>
              </div>
              <p v-if="mailReady" class="warn">{{ user.email ? t('account.me.emailWarnChange') : t('account.me.emailWarnAdd') }}</p>
              <p v-else class="warn bad">
                {{ t('account.me.emailNoMail') }}
                <router-link v-if="isOwner" :to="{ name: 'admin-settings' }">{{ t('account.me.goMail') }}</router-link>
              </p>
              <div class="ft"><Transition name="nt"><p v-if="notes.account && !notes.account.ok" class="notice in" role="alert">{{ notes.account.text }}</p></Transition><button type="submit" class="btn" :disabled="busy || !mailReady || !form.email.trim()">{{ t('account.me.sendConfirm') }}</button></div>
            </form>
          </div>
        </div>

        <!-- 密码 -->
        <div class="row" :class="{ open: editing === 'password' }">
          <span class="ic"><Icon :icon="KeyRound" :size="17" /></span>
          <div class="tx">
            <b>{{ t('account.password') }}</b>
            <span class="val">{{ user.hasPassword ? t('account.me.pwSet') : t('account.me.pwNotSet') }}</span>
            <small>{{ user.hasPassword ? t('account.me.pwHint') : t('account.me.setPwDesc') }}</small>
          </div>
          <button type="button" class="btn ghost sm" @click="edit('password')">
            {{ editing === 'password' ? t('account.me.cancel') : user.hasPassword ? t('account.me.edit') : t('account.me.add') }}
          </button>
        </div>
        <div class="fold" :class="{ open: editing === 'password' }" :inert="editing !== 'password'" :aria-hidden="editing !== 'password'">
          <div class="clip">
            <form class="editor" @submit.prevent="savePassword">
              <label v-if="user.hasPassword" class="field">
                <span>{{ t('account.me.oldPw') }}</span>
                <input v-model="form.old" type="password" autocomplete="current-password" />
              </label>
              <div class="two">
                <label class="field">
                  <span>{{ t('account.newPassword') }}</span>
                  <input v-model="form.next" type="password" autocomplete="new-password" :placeholder="t('account.pwHint')" />
                </label>
                <label class="field">
                  <span>{{ t('account.confirmPassword') }}</span>
                  <input v-model="form.confirm" type="password" autocomplete="new-password" />
                </label>
              </div>
              <div class="ft"><Transition name="nt"><p v-if="notes.account && !notes.account.ok" class="notice in" role="alert">{{ notes.account.text }}</p></Transition><button type="submit" class="btn" :disabled="busy">{{ t('account.me.savePw') }}</button></div>
            </form>
          </div>
        </div>

        <!-- GitHub -->
        <div v-if="githubOn || user.github" class="row">
          <span class="ic"><svg viewBox="0 0 24 24" width="17" height="17" fill="currentColor" aria-hidden="true"><path :d="siGithub.path" /></svg></span>
          <div class="tx">
            <b>GitHub</b>
            <span class="val">{{ user.github ? t('account.me.linked') : t('account.me.notLinked') }}</span>
            <small>{{ user.github ? t('account.me.githubOn') : t('account.me.githubOff') }}</small>
          </div>
          <span v-if="user.github" class="state"><Icon :icon="Check" :size="15" /></span>
          <a v-else class="btn ghost sm" :href="linkHref">{{ t('account.me.link') }}</a>
        </div>

        <Transition name="nt"><p v-if="notes.account && (notes.account.ok || !editing)" class="notice" :class="{ ok: notes.account.ok }" role="status">{{ notes.account.text }}</p></Transition>
      </section>

      <button type="button" class="logout rise" style="--i: 3" @click="logout"><Icon :icon="LogOut" :size="16" />{{ t('account.me.logout') }}</button>
    </template>

    <AvatarCropper :file="cropping" :busy="uploading" @done="upload" @cancel="cropping = null" />
  </main>
</template>

<style scoped lang="scss">
.me {
  max-width: 680px;
  margin: 0 auto;
  padding: 104px 24px 96px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* ---------- 头像 + 名字 ---------- */
.head {
  display: flex;
  align-items: center;
  gap: 18px;
  margin-bottom: 4px;

  .who { flex: 1; min-width: 0; }
  h1 { font: 700 26px/1.25 var(--font-serif); }

  p {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-top: 6px;
    font-size: 13px;
    color: var(--text-3);
  }

  .av-line { margin-top: 4px; font-size: 12.5px; }
  .mono { font-family: var(--font-mono); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
}

.av {
  position: relative;
  width: 72px;
  height: 72px;
  flex: none;
  padding: 0;
  border: 0;
  overflow: hidden;
  border-radius: 50%;
  background: color-mix(in oklab, var(--ink) 16%, var(--elev));
  cursor: pointer;

  img { width: 100%; height: 100%; object-fit: cover; display: block; }
  .ini { display: grid; place-items: center; width: 100%; height: 100%; color: var(--ink); font: 700 26px var(--font-serif); }

  .cam {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    color: #fff;
    background: rgb(0 0 0 / 0.42);
    opacity: 0;
    transition: opacity var(--dur-fast);
  }

  &:hover .cam, &:focus-visible .cam { opacity: 1; }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

.role {
  flex: none;
  padding: 2px 9px;
  border-radius: var(--r-pill);
  background: var(--fill);
  color: var(--text-2);
  font-size: 12px;

  &.admin, &.author { background: color-mix(in oklab, var(--ink) 14%, transparent); color: var(--ink); }
}

.link {
  padding: 0;
  border: 0;
  background: none;
  color: var(--ink);
  font-size: inherit;
  cursor: pointer;

  &:hover { text-decoration: underline; }
}

.pending {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border-radius: var(--r-lg);
  background: color-mix(in oklab, var(--accent-yellow) 10%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--accent-yellow) 30%, transparent);

  img { width: 40px; height: 40px; border-radius: 50%; object-fit: cover; }
  .tx { flex: 1; display: flex; flex-direction: column; gap: 2px; }
  b { display: inline-flex; align-items: center; gap: 6px; font-size: 14px; font-weight: 600; }
  small { font-size: 12.5px; color: var(--text-3); }
}

/* ---------- 区块：跟随界面风格 ---------- */
.sec {
  padding: var(--card-pad-sm) calc(var(--card-pad-sm) + 2px);
  border-radius: var(--card-r);
  background: var(--card-bg);
  box-shadow: var(--card-shadow), inset 0 0 0 0.5px var(--line-2);

  h2 { margin-bottom: 14px; font: 600 17px var(--font-serif); }
  .desc { margin: -8px 0 10px; font-size: 13px; color: var(--text-3); line-height: 1.6; }
}

/* 简洁：透明底、无投影，区块之间一道发丝线 */
:root[data-style='clean'] .sec {
  padding: 26px 0 4px;
  border-radius: 0;
  box-shadow: inset 0 1px 0 var(--line);
}

:root[data-style='clean'] .me { gap: 8px; }

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;

  > span:first-child { font-size: 13px; color: var(--text-2); }
  em { margin-left: 8px; font-style: normal; color: var(--text-3); font-size: 12px; }

  input {
    width: 100%;
    min-width: 0;
    height: 42px;
    padding: 0 14px;
    border: 0;
    outline: 0;
    border-radius: var(--r-md);
    background: var(--fill);
    box-shadow: inset 0 0 0 1px var(--line);
    color: var(--text);
    font: 15px var(--font-sans);
    transition: box-shadow var(--dur-fast), background var(--dur-fast);

    &:focus { background: var(--bg); box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--ink) 50%, transparent), 0 0 0 3px color-mix(in oklab, var(--ink) 14%, transparent); }
  }

  .inline { display: flex; gap: 10px; }
  .inline .btn { flex: none; height: 42px; }
}

.row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 0;

  & + .fold + .row, & + .row { box-shadow: 0 -0.5px 0 var(--line-2); }

  .ic { width: 38px; height: 38px; flex: none; display: grid; place-items: center; border-radius: 50%; background: var(--fill); color: var(--text-2); }
  .tx { flex: 1; min-width: 0; display: grid; grid-template-columns: 76px minmax(0, 1fr); gap: 2px 12px; align-items: center; }
  /* 标签跨两行，与右侧「值 + 说明」整体垂直居中 */
  b { grid-row: 1 / span 2; align-self: center; font-size: 14px; font-weight: 500; color: var(--text-2); }
  .val { font-size: 15px; color: var(--text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: inline-flex; align-items: center; gap: 8px; }
  .val.mono { font-family: var(--font-mono); font-size: 14px; }
  small { grid-column: 2; font-size: 12.5px; color: var(--text-3); line-height: 1.5; }
  small.pend { color: color-mix(in oklab, var(--accent-yellow) 70%, var(--text)); }
  small .link { margin-left: 6px; font-size: 12.5px; }
  .state { color: var(--ink); display: grid; place-items: center; width: 32px; }
}

.chip {
  padding: 1px 8px;
  border-radius: var(--r-pill);
  font-style: normal;
  font-size: 11.5px;
  background: color-mix(in oklab, var(--accent-yellow) 16%, transparent);
  color: color-mix(in oklab, var(--accent-yellow) 70%, var(--text));

  &.ok { background: color-mix(in oklab, var(--ink) 12%, transparent); color: var(--ink); }
}

/* 行内编辑：grid 行高展开 / 收起 */
.fold {
  display: grid;
  grid-template-rows: 0fr;
  opacity: 0;
  transition: grid-template-rows var(--dur) var(--ease-out), opacity var(--dur) var(--ease-out);

  &.open { grid-template-rows: 1fr; opacity: 1; }
}

.clip { min-height: 0; overflow: hidden; }

.editor {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 0 0 14px 52px;
  padding: 14px 16px;
  border-radius: var(--r-lg);
  background: var(--fill);

  .field input { background: var(--bg); }
  .warn { font-size: 12.5px; line-height: 1.6; color: var(--text-3); }
  .warn.bad { color: color-mix(in oklab, var(--accent-red) 75%, var(--text)); }
  .warn a { margin-left: 6px; color: var(--ink); }
  .ft { display: flex; justify-content: flex-end; align-items: center; gap: 12px; }
}

.two { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }

.notice { margin-top: 8px; font-size: 13px; color: var(--accent-red); }
.notice.ok { color: var(--ink); }
.notice.top { margin: -6px 0 0; }
.notice.in { flex: 1; margin: 0; align-self: center; }

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 38px;
  padding: 0 18px;
  border: 0;
  border-radius: var(--r-pill);
  background: var(--solid);
  color: var(--on-solid);
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  cursor: pointer;
  transition: transform var(--dur-fast) var(--ease-spring), opacity var(--dur-fast), background var(--dur-fast);

  &:active:not(:disabled) { transform: scale(0.97); }
  &:disabled { opacity: 0.5; cursor: not-allowed; }

  &.ghost { background: var(--fill); color: var(--text); box-shadow: inset 0 0 0 1px var(--line); }
  &.ghost:hover:not(:disabled) { background: color-mix(in oklab, var(--text) 8%, transparent); }
  &.sm { height: 32px; padding: 0 14px; font-size: 13px; }
}

.logout {
  align-self: center;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  padding: 8px 14px;
  border: 0;
  border-radius: var(--r-pill);
  background: none;
  color: var(--text-3);
  font-size: 14px;
  cursor: pointer;
  transition: color var(--dur-fast), background var(--dur-fast);

  &:hover { color: var(--accent-red); background: color-mix(in oklab, var(--accent-red) 8%, transparent); }
}

.nt-enter-active, .nt-leave-active { transition: opacity var(--dur-fast); }
.nt-enter-from, .nt-leave-to { opacity: 0; }
.fold-enter-active, .fold-leave-active { transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-out); }
.fold-enter-from, .fold-leave-to { opacity: 0; transform: translateY(-6px); }

@media (max-width: 767px) {
  .me { padding: 24px 16px 120px; }
  .two { grid-template-columns: 1fr; }
  .head { flex-wrap: wrap; }
  .row .tx { grid-template-columns: 1fr; }
  .row b { grid-row: auto; }
  .row small { grid-column: 1; }
  .editor { margin-left: 0; }
}

@media (prefers-reduced-motion: reduce) {
  .fold { transition: none; }
}
</style>
