<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Check, LayoutDashboard, LogOut, Mail } from 'lucide';
import { siGithub } from 'simple-icons';
import { accountApi } from '../api';
import { useAuthStore } from '../stores/auth';
import { useConfigStore } from '../stores/config';
import { useDialogStore } from '../stores/dialog';
import Icon from '../components/ui/Icon.vue';

/**
 * 我的账号（/account）：资料（昵称、头像地址）、登录方式（密码 / GitHub 绑定）、改密、退出；
 * 站长与作者多一个进后台入口。未登录时送去登录页并带回跳。
 */
const { t } = useI18n();
const router = useRouter();
const auth = useAuthStore();
const config = useConfigStore();
const dialog = useDialogStore();

const user = computed(() => auth.user);
const githubOn = computed(() => !!config.cfg.users?.login.github);

const profile = reactive({ name: '', avatar: '' });
const pw = reactive({ old: '', next: '', confirm: '' });
const savingProfile = ref(false);
const savingPw = ref(false);
const notice = reactive({ profile: null as null | { ok: boolean; text: string }, pw: null as null | { ok: boolean; text: string } });

const ROLE = computed(() => t(`account.role_${user.value?.role ?? 'reader'}`));
const initial = computed(() => (user.value?.name || user.value?.login || '?').trim()[0]?.toUpperCase() ?? '?');

onMounted(async () => {
  await auth.sync();
  if (!auth.loggedIn || !auth.user) {
    void router.replace({ path: '/account/login', query: { next: '/account' } });
    return;
  }
  profile.name = auth.user.name;
  profile.avatar = auth.user.avatar;
});

async function saveProfile(): Promise<void> {
  if (savingProfile.value) return;
  savingProfile.value = true;
  notice.profile = null;
  try {
    await accountApi.updateMe(profile.name.trim(), profile.avatar.trim());
    await auth.sync();
    notice.profile = { ok: true, text: t('account.me.saved') };
  } catch (e) {
    const code = (e as Error).message;
    notice.profile = { ok: false, text: code === 'invalid_avatar' ? t('account.me.badAvatar') : code === 'invalid_name' ? t('account.err.name') : t('account.err.generic') };
  } finally {
    savingProfile.value = false;
  }
}

async function changePw(): Promise<void> {
  if (savingPw.value) return;
  notice.pw = null;
  if (pw.next.length < 8) notice.pw = { ok: false, text: t('account.err.weak') };
  else if (pw.next !== pw.confirm) notice.pw = { ok: false, text: t('account.err.mismatch') };
  if (notice.pw) return;
  savingPw.value = true;
  try {
    await accountApi.changePassword(pw.old, pw.next);
    pw.old = pw.next = pw.confirm = '';
    await auth.sync();
    notice.pw = { ok: true, text: t('account.me.pwChanged') };
  } catch (e) {
    const code = (e as Error).message;
    notice.pw = { ok: false, text: code === 'bad_credentials' ? t('account.me.pwWrong') : code === 'too_many_attempts' ? t('account.err.tooMany') : t('account.err.generic') };
  } finally {
    savingPw.value = false;
  }
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
      <section class="head rise">
        <span class="av">
          <img v-if="user.avatar" :src="user.avatar" alt="" referrerpolicy="no-referrer" />
          <template v-else>{{ initial }}</template>
        </span>
        <div class="who">
          <h1>{{ user.name || user.login }}</h1>
          <p><span class="role" :class="user.role">{{ ROLE }}</span><span class="mono">{{ user.email || `@${user.login}` }}</span></p>
        </div>
        <router-link v-if="auth.staff" to="/admin" class="btn ghost"><Icon :icon="LayoutDashboard" :size="16" />{{ t('account.me.admin') }}</router-link>
      </section>

      <section class="card rise" style="--i: 1">
        <h2>{{ t('account.me.profile') }}</h2>
        <label class="field">
          <span>{{ t('account.name') }}</span>
          <input v-model="profile.name" maxlength="40" autocomplete="nickname" />
        </label>
        <label class="field">
          <span>{{ t('account.me.avatar') }}<em>{{ t('account.me.avatarHint') }}</em></span>
          <input v-model="profile.avatar" placeholder="https://…" spellcheck="false" />
        </label>
        <div class="ft">
          <Transition name="nt"><span v-if="notice.profile" class="notice" :class="{ ok: notice.profile.ok }">{{ notice.profile.text }}</span></Transition>
          <button type="button" class="btn" :disabled="savingProfile" @click="saveProfile">{{ t('account.me.save') }}</button>
        </div>
      </section>

      <section class="card rise" style="--i: 2">
        <h2>{{ t('account.me.signin') }}</h2>
        <div class="row">
          <span class="ic"><Icon :icon="Mail" :size="18" /></span>
          <div class="tx">
            <b>{{ t('account.me.emailPw') }}</b>
            <small>{{ user.email || t('account.me.noEmail') }}<template v-if="user.email"> · {{ user.emailVerified ? t('account.me.verified') : t('account.me.unverified') }}</template></small>
          </div>
          <span v-if="user.hasPassword" class="state"><Icon :icon="Check" :size="15" />{{ t('account.me.set') }}</span>
        </div>
        <div v-if="githubOn || user.github" class="row">
          <span class="ic"><svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" aria-hidden="true"><path :d="siGithub.path" /></svg></span>
          <div class="tx">
            <b>GitHub</b>
            <small>{{ user.github ? t('account.me.githubOn') : t('account.me.githubOff') }}</small>
          </div>
          <span v-if="user.github" class="state"><Icon :icon="Check" :size="15" />{{ t('account.me.linked') }}</span>
          <a v-else class="btn ghost sm" :href="linkHref">{{ t('account.me.link') }}</a>
        </div>
      </section>

      <section class="card rise" style="--i: 3">
        <h2>{{ user.hasPassword ? t('account.me.changePw') : t('account.me.setPw') }}</h2>
        <p v-if="!user.hasPassword" class="desc">{{ t('account.me.setPwDesc') }}</p>
        <label v-if="user.hasPassword" class="field">
          <span>{{ t('account.me.oldPw') }}</span>
          <input v-model="pw.old" type="password" autocomplete="current-password" />
        </label>
        <div class="two">
          <label class="field">
            <span>{{ t('account.newPassword') }}</span>
            <input v-model="pw.next" type="password" autocomplete="new-password" :placeholder="t('account.pwHint')" />
          </label>
          <label class="field">
            <span>{{ t('account.confirmPassword') }}</span>
            <input v-model="pw.confirm" type="password" autocomplete="new-password" />
          </label>
        </div>
        <div class="ft">
          <Transition name="nt"><span v-if="notice.pw" class="notice" :class="{ ok: notice.pw.ok }">{{ notice.pw.text }}</span></Transition>
          <button type="button" class="btn" :disabled="savingPw" @click="changePw">{{ t('account.me.savePw') }}</button>
        </div>
      </section>

      <button type="button" class="logout rise" style="--i: 4" @click="logout"><Icon :icon="LogOut" :size="16" />{{ t('account.me.logout') }}</button>
    </template>
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

.head {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 8px;

  .av {
    width: 64px;
    height: 64px;
    flex: none;
    display: grid;
    place-items: center;
    overflow: hidden;
    border-radius: 50%;
    background: color-mix(in oklab, var(--ink) 16%, var(--elev));
    color: var(--ink);
    font: 700 24px var(--font-serif);

    img { width: 100%; height: 100%; object-fit: cover; }
  }

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

  .mono { font-family: var(--font-mono); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
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

.card {
  padding: 20px 22px;
  border-radius: var(--r-xl);
  background: var(--elev);
  box-shadow: inset 0 0 0 0.5px var(--line-2);

  h2 { margin-bottom: 14px; font: 600 17px var(--font-serif); }
  .desc { margin: -6px 0 14px; font-size: 13.5px; color: var(--text-3); line-height: 1.6; }
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 12px;

  > span { font-size: 13px; color: var(--text-2); }
  em { margin-left: 8px; font-style: normal; color: var(--text-3); font-size: 12px; }

  input {
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
}

.two { display: grid; grid-template-columns: 1fr 1fr; gap: 0 12px; }

.ft {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 4px;
}

.notice { flex: 1; font-size: 13px; color: var(--accent-red); }
.notice.ok { color: var(--ink); }

.row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;

  & + & { box-shadow: 0 -0.5px 0 var(--line-2); }

  .ic { width: 36px; height: 36px; flex: none; display: grid; place-items: center; border-radius: 50%; background: var(--fill); color: var(--text-2); }
  .tx { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
  b { font-size: 14.5px; font-weight: 500; }
  small { font-size: 12.5px; color: var(--text-3); }
  .state { display: inline-flex; align-items: center; gap: 4px; font-size: 13px; color: var(--ink); }
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 38px;
  padding: 0 18px;
  border: 0;
  border-radius: var(--r-pill);
  background: var(--solid);
  color: var(--on-solid);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: transform var(--dur-fast) var(--ease-spring), opacity var(--dur-fast);

  &:active:not(:disabled) { transform: scale(0.97); }
  &:disabled { opacity: 0.6; cursor: progress; }

  &.ghost { background: var(--fill); color: var(--text); box-shadow: inset 0 0 0 1px var(--line); }
  &.sm { height: 32px; padding: 0 14px; font-size: 13px; }
}

.logout {
  align-self: center;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
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

@media (max-width: 767px) {
  .me { padding: 24px 16px 120px; }
  .two { grid-template-columns: 1fr; }
}
</style>
