<script setup lang="ts">
/**
 * 移动端登录：门光 logo + 衬线大标题 + iOS 分组输入框；错误时卡片抖动，成功后门光放大转场进入后台。
 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../../api';
import { useAuthStore } from '../../../stores/auth';
import { useConfigStore } from '../../../stores/config';
import { useThemeStore } from '../../../stores/theme';
import MaIcon from '../../../components/mobile-admin/MaIcon.vue';
import MaIsland from '../../../components/mobile-admin/MaIsland.vue';
import MaRing from '../../../components/mobile-admin/MaRing.vue';
import { toast } from '../../../components/mobile-admin/state';
import { onColor } from '../../../components/mobile-admin/format';

const { t } = useI18n();
const router = useRouter();
const auth = useAuthStore();
const config = useConfigStore();
const theme = useThemeStore();

const username = ref('');
const password = ref('');
const reveal = ref(false);
const busy = ref(false);
const shake = ref(false);
const success = ref(false);
const error = ref('');

const onPrimary = computed(() => onColor(theme.palette?.[theme.mode]?.primary ?? '#0078ff'));
const canSubmit = computed(() => !!username.value && !!password.value && !busy.value);

async function submit(): Promise<void> {
  if (!canSubmit.value) return;
  busy.value = true;
  error.value = '';
  try {
    auth.setToken(await adminApi.login(username.value, password.value));
    success.value = true;
    (document.activeElement as HTMLElement | null)?.blur();
    window.setTimeout(() => {
      void router.replace('/admin');
      window.setTimeout(() => toast(t('mobileAdmin.login.welcome'), config.cfg.about.name || ''), 500);
    }, 520);
  } catch {
    error.value = t('mobileAdmin.login.failed');
    shake.value = false;
    requestAnimationFrame(() => (shake.value = true));
    navigator.vibrate?.([12, 40, 12]);
  } finally {
    busy.value = false;
  }
}

onMounted(() => document.documentElement.classList.add('ma-lock'));
onBeforeUnmount(() => document.documentElement.classList.remove('ma-lock'));
</script>

<template>
  <div class="ma-root ml" :class="{ success }" :style="{ '--on-primary': onPrimary }">
    <div class="glow" aria-hidden="true" />
    <main class="ml-main">
      <div class="brand">
        <span class="halo" aria-hidden="true" />
        <img class="logo" src="/favicon-256.png" alt="" draggable="false" />
      </div>
      <h1>{{ t('mobileAdmin.login.title') }}</h1>
      <p class="sub">{{ config.cfg.site.title }} · {{ t('mobileAdmin.login.sub') }}</p>

      <form class="card" :class="{ shake }" @submit.prevent="submit" @animationend="shake = false">
        <div class="ma-list fields">
          <label class="row">
            <MaIcon name="user" :size="19" />
            <input
              v-model="username"
              type="text"
              autocomplete="username"
              autocapitalize="off"
              spellcheck="false"
              enterkeyhint="next"
              :placeholder="t('mobileAdmin.login.username')"
            />
          </label>
          <label class="row">
            <MaIcon name="lock" :size="19" />
            <input
              v-model="password"
              :type="reveal ? 'text' : 'password'"
              autocomplete="current-password"
              enterkeyhint="go"
              :placeholder="t('mobileAdmin.login.password')"
            />
            <button type="button" class="eye tap" :aria-label="t('mobileAdmin.login.reveal')" @click="reveal = !reveal">
              <MaIcon name="eye" :size="18" :class="{ off: !reveal }" />
            </button>
          </label>
        </div>
        <p class="err" :class="{ on: !!error }" role="alert">{{ error }}</p>
        <button class="go tap" type="submit" :disabled="!canSubmit">
          <MaRing v-if="busy" indeterminate :size="18" :stroke="2.2" class="go-ring" />
          <span>{{ busy ? t('mobileAdmin.login.loggingIn') : t('mobileAdmin.login.submit') }}</span>
        </button>
      </form>

      <a class="back" href="/"><MaIcon name="back" :size="16" />{{ t('mobileAdmin.login.backSite') }}</a>
    </main>
    <MaIsland />
  </div>
</template>

<style lang="scss">
@use '../../../components/mobile-admin/ma-tokens.scss';

html.ma-lock,
html.ma-lock body {
  overflow: hidden;
  overscroll-behavior: none;
}
</style>

<style scoped lang="scss">
.ml {
  position: fixed;
  inset: 0;
  z-index: 100;
  overflow-y: auto;
  background: var(--bg);
  display: flex;
  flex-direction: column;
}

.glow {
  position: absolute;
  inset: -20% -30% auto;
  height: 70%;
  background:
    radial-gradient(50% 50% at 50% 60%, color-mix(in oklab, var(--primary) 26%, transparent), transparent 70%),
    radial-gradient(30% 30% at 50% 70%, rgba(255, 255, 255, 0.08), transparent 70%);
  pointer-events: none;
  animation: ml-breathe 6s ease-in-out infinite alternate;
}

@keyframes ml-breathe {
  from { opacity: 0.75; transform: scale(0.96); }
  to { opacity: 1; transform: scale(1.04); }
}

.ml-main {
  position: relative;
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: calc(var(--safe-t) + 40px) 24px calc(var(--safe-b) + 28px);
  max-width: 440px;
  width: 100%;
  margin: 0 auto;
}

.brand {
  position: relative;
  width: 88px;
  height: 88px;
  margin: 0 auto;
  animation: ml-in 0.8s var(--ease-spring) backwards;

  .logo {
    position: relative;
    width: 100%;
    height: 100%;
    border-radius: 26px;
    box-shadow: 0 20px 40px -14px color-mix(in oklab, var(--primary) 60%, black), 0 0 0 0.5px rgba(255, 255, 255, 0.14);
    transition: transform 0.6s var(--ease-sheet);
  }

  .halo {
    position: absolute;
    inset: -30px;
    border-radius: 50%;
    background: radial-gradient(closest-side, color-mix(in oklab, var(--primary) 45%, transparent), transparent);
    filter: blur(8px);
    transition: transform 0.7s var(--ease-sheet), opacity 0.7s;
  }
}

h1 {
  margin-top: 26px;
  text-align: center;
  font-family: var(--font-serif);
  font-size: 32px;
  font-weight: 700;
  letter-spacing: 0.02em;
  animation: ml-in 0.7s var(--ease-out) 0.08s backwards;
}

.sub {
  margin-top: 6px;
  text-align: center;
  font-size: 14px;
  color: var(--text-3);
  animation: ml-in 0.7s var(--ease-out) 0.14s backwards;
}

@keyframes ml-in {
  from { opacity: 0; transform: translateY(16px) scale(0.96); filter: blur(6px); }
}

.card {
  margin-top: 34px;
  animation: ml-in 0.7s var(--ease-out) 0.2s backwards;

  &.shake { animation: ml-shake 0.5s var(--ease-out); }
}

@keyframes ml-shake {
  0%, 100% { transform: none; }
  20% { transform: translateX(-10px); }
  40% { transform: translateX(8px); }
  60% { transform: translateX(-5px); }
  80% { transform: translateX(3px); }
}

.fields {
  background: var(--elev);
  box-shadow: inset 0 0 0 0.5px var(--line), var(--shadow);
}

.row {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  height: 56px;
  padding: 0 14px 0 16px;
  color: var(--text-3);

  & + .row::before {
    content: '';
    position: absolute;
    top: 0;
    left: 47px;
    right: 0;
    height: 0.5px;
    background: var(--line-2);
  }

  &:focus-within { color: var(--ink); }

  input {
    flex: 1;
    min-width: 0;
    height: 100%;
    font-size: 17px;
    color: var(--text);
    caret-color: var(--primary);

    &::placeholder { color: var(--text-3); }
  }
}

.eye {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: var(--text-3);

  .off { opacity: 0.55; }
}

.err {
  min-height: 20px;
  margin: 10px 4px 0;
  font-size: 13px;
  color: var(--accent-red);
  opacity: 0;
  transform: translateY(-4px);
  transition: all var(--dur) var(--ease-out);

  &.on { opacity: 1; transform: none; }
}

.go {
  width: 100%;
  height: 52px;
  margin-top: 8px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 17px;
  font-weight: 600;
  color: var(--on-primary);
  background: linear-gradient(180deg, color-mix(in oklab, var(--primary) 88%, white), var(--primary) 50%, var(--primary-deep));
  box-shadow: 0 14px 28px -12px var(--glow), inset 0 1px 0 rgba(255, 255, 255, 0.3);
  transition: opacity var(--dur), transform var(--dur-fast) var(--ease-spring);

  &:disabled { opacity: 0.45; box-shadow: none; }

  .go-ring {
    --ring-bg: color-mix(in oklab, var(--on-primary) 30%, transparent);
    --ring-fg: var(--on-primary);
  }
}

.back {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin: 28px auto 0;
  font-size: 14px;
  color: var(--text-3);
}

/* 登录成功：门光放大，内容淡出 */
.success {
  .brand .halo { transform: scale(3.2); opacity: 0.9; }
  .brand .logo { transform: scale(1.08); }

  h1,
  .sub,
  .card,
  .back {
    opacity: 0;
    transform: translateY(-10px);
    transition: opacity 0.35s, transform 0.45s var(--ease-sheet);
  }
}
</style>
