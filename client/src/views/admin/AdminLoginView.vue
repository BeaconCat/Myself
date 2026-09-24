<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';
import { useAuthStore } from '../../stores/auth';
import { useConfigStore } from '../../stores/config';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import LightCover from './studio/LightCover.vue';
import '@fontsource/noto-serif-sc/400.css';
import '@fontsource/noto-serif-sc/600.css';

/** 登录 · Studio：左侧门缝光，右侧纸面表单；成功后推门进入「今天」 */
const { t } = useI18n();
const router = useRouter();
const auth = useAuthStore();
const config = useConfigStore();

const username = ref('');
const password = ref('');
const show = ref(false);
const error = ref('');
const busy = ref(false);
const shaking = ref(false);
const entering = ref(false);

async function submit(): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  error.value = '';
  try {
    auth.setToken(await adminApi.login(username.value, password.value));
    entering.value = true;
    window.setTimeout(() => void router.push({ name: 'admin-today' }), 520);
  } catch {
    error.value = t('studio.login.failed');
    shaking.value = false;
    requestAnimationFrame(() => (shaking.value = true));
  } finally {
    busy.value = false;
  }
}

onMounted(() => {
  document.documentElement.dataset.studio = '';
});
onBeforeUnmount(() => {
  delete document.documentElement.dataset.studio;
});
</script>

<template>
  <main class="studio login" :class="{ entering }">
    <div class="card" :class="{ shake: shaking }" @animationend="shaking = false">
      <div class="art">
        <LightCover class="door" kind="door" />
        <div class="quote">
          <p>{{ config.cfg.about?.motto || t('studio.login.motto') }}</p>
          <small>{{ config.cfg.site.title }} · {{ t('studio.brandSub') }}</small>
        </div>
      </div>
      <form class="form" @submit.prevent="submit">
        <img class="logo" src="/favicon-64.png" alt="" draggable="false" />
        <h1>{{ t('studio.login.title') }}</h1>
        <p class="sub">{{ t('studio.login.sub') }}</p>

        <label>
          <span class="st-flabel">{{ t('studio.login.user') }}</span>
          <span class="st-field"><SIcon name="user" :size="16" /><input v-model="username" autocomplete="username" required autofocus /></span>
        </label>
        <label>
          <span class="st-flabel">{{ t('studio.login.pass') }}</span>
          <span class="st-field">
            <SIcon name="lock" :size="16" />
            <input v-model="password" :type="show ? 'text' : 'password'" autocomplete="current-password" required />
            <button type="button" class="eye" :title="show ? t('studio.login.hide') : t('studio.login.show')" @click="show = !show">
              <SIcon :name="show ? 'eyeOff' : 'eye'" :size="16" />
            </button>
          </span>
        </label>
        <p class="err" :class="{ on: !!error }">{{ error || '&nbsp;' }}</p>
        <button type="submit" class="st-btn p lg go" :disabled="busy">
          {{ busy ? t('studio.login.busy') : t('studio.login.enter') }}<SIcon name="arrowR" :size="16" />
        </button>
        <a class="back st-link" href="/"><SIcon name="arrowL" :size="14" />{{ t('studio.login.back') }}</a>
      </form>
    </div>
  </main>
</template>

<style scoped lang="scss">
.login {
  position: fixed;
  inset: 0;
  display: grid;
  place-items: center;
  padding: 24px;
  background:
    radial-gradient(900px 600px at 80% -10%, color-mix(in oklab, var(--primary) 14%, transparent), transparent 60%),
    radial-gradient(700px 500px at 10% 110%, color-mix(in oklab, var(--primary) 8%, transparent), transparent 60%),
    var(--desk);
  overflow: auto;
}

.card {
  display: grid;
  grid-template-columns: 380px 400px;
  border-radius: 26px;
  background: var(--paper);
  box-shadow: var(--sh-paper);
  overflow: hidden;
  animation: card-in var(--dur-slow) var(--ease-spring) both;

  &.shake { animation: shake 0.42s var(--ease-out); }

  .entering & {
    transition: transform 0.5s var(--ease-out), opacity 0.5s var(--ease-out);
    transform: scale(1.03);
    opacity: 0;
  }
}

@keyframes card-in { from { opacity: 0; transform: translateY(18px) scale(0.97); } }
@keyframes shake { 20%, 60% { transform: translateX(-6px); } 40%, 80% { transform: translateX(6px); } }

.art {
  position: relative;
  min-height: 520px;

  .door { position: absolute; inset: 0; }

  .quote {
    position: absolute;
    left: 28px;
    right: 28px;
    bottom: 26px;
    color: #fff;

    p { margin: 0 0 6px; font: 600 20px/1.5 var(--font-serif); letter-spacing: 0.04em; }
    small { font-size: 12px; color: rgba(255, 255, 255, 0.6); letter-spacing: 0.08em; }
  }
}

.form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 44px 40px 32px;

  .logo { width: 40px; height: 40px; border-radius: 11px; box-shadow: 0 6px 14px -6px rgba(10, 20, 40, 0.5); }

  h1 { font: 600 26px/1.3 var(--font-serif); margin: 6px 0 0; }
  .sub { margin: -8px 0 8px; font-size: 13.5px; color: var(--ink-3); }

  label { display: block; }

  .eye {
    width: 28px;
    height: 28px;
    display: grid;
    place-items: center;
    border-radius: 8px;
    color: var(--ink-3);

    &:hover { background: var(--hover); color: var(--ink); }
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

  .go {
    width: 100%;

    .st-ic { transition: transform var(--dur) var(--ease-spring); }
    &:hover .st-ic { transform: translateX(3px); }
  }

  .back { align-self: center; font-size: 12.5px; margin-top: 4px; }
}

@media (max-width: 860px) {
  .card { grid-template-columns: 1fr; width: min(420px, 100%); }
  .art { min-height: 180px; }
}
</style>
