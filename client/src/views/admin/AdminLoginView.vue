<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';
import { staffHome, useAuthStore } from '../../stores/auth';
import { useConfigStore } from '../../stores/config';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import LightCover from './studio/LightCover.vue';
import '@fontsource/noto-serif-sc/400.css';
import '@fontsource/noto-serif-sc/600.css';

/**
 * 登录 · Studio：左侧默认封面（窗格光影），右侧纸面表单。
 * 验证通过：表单原位切到「验证通过」态（输入框右侧对勾、按钮换文案带对勾），
 * 同时立即跳转后台，由路由遮罩盖住整张卡片一起离场，不再单独播卡片退场。
 */
const { t } = useI18n();
const router = useRouter();
const auth = useAuthStore();
const config = useConfigStore();

const username = ref('');
const password = ref('');
const show = ref(false);
const error = ref('');
const busy = ref(false);
/** 表单卡片：报错抖动用 WAAPI 单独播放，不替换 CSS 入场动画（否则移除抖动类时入场会重播） */
const cardEl = ref<HTMLElement | null>(null);
/** 验证通过：表单锁定为只读并显示对勾，等待路由遮罩接走 */
const verified = ref(false);

async function submit(): Promise<void> {
  if (busy.value || verified.value) return;
  busy.value = true;
  error.value = '';
  try {
    const res = await adminApi.login(username.value, password.value);
    if (res.user.role === 'reader') {
      // 读者没有后台权限：不保留这次登录
      await auth.logout();
      throw new Error('not_staff');
    }
    auth.markLoggedIn(res.user);
    verified.value = true;
    (document.activeElement as HTMLElement | null)?.blur();
    void router.push(res.mustChange ? { path: '/setup', query: { change: '1' } } : staffHome(res.user.role));
  } catch (e) {
    const code = (e as Error).message;
    error.value = code === 'too_many_attempts' ? t('studio.login.locked') : code === 'not_staff' ? t('studio.login.notStaff') : t('studio.login.failed');
    shake();
  } finally {
    busy.value = false;
  }
}

function shake(): void {
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
  cardEl.value?.animate(
    [
      { transform: 'translateX(0)' },
      { transform: 'translateX(-6px)', offset: 0.2 },
      { transform: 'translateX(6px)', offset: 0.4 },
      { transform: 'translateX(-6px)', offset: 0.6 },
      { transform: 'translateX(6px)', offset: 0.8 },
      { transform: 'translateX(0)' },
    ],
    { duration: 420, easing: 'cubic-bezier(.2,.8,.3,1)', composite: 'add' },
  );
}

onMounted(() => {
  document.documentElement.dataset.studio = '';
});
onBeforeUnmount(() => {
  delete document.documentElement.dataset.studio;
});
</script>

<template>
  <main class="studio login" :class="{ verified }">
    <div ref="cardEl" class="card">
      <div class="art">
        <LightCover class="door" kind="door" />
        <div class="quote">
          <p>{{ config.cfg.about?.motto || t('studio.login.motto') }}</p>
          <small>{{ config.cfg.site.title }} · {{ t('studio.brandSub') }}</small>
        </div>
      </div>
      <form class="form" @submit.prevent="submit">
        <img class="logo" :src="config.cfg.site.logo || '/favicon-64.png'" alt="" draggable="false" />
        <h1>{{ t('studio.login.title') }}</h1>
        <p class="sub">{{ t('studio.login.sub') }}</p>

        <label>
          <span class="st-flabel">{{ t('studio.login.user') }}</span>
          <span class="st-field">
            <SIcon name="user" :size="16" />
            <input v-model="username" autocomplete="username" required autofocus :readonly="verified" />
            <SIcon v-if="verified" class="ok" data-live name="check" :size="16" />
          </span>
        </label>
        <label>
          <span class="st-flabel">{{ t('studio.login.pass') }}</span>
          <span class="st-field">
            <SIcon name="lock" :size="16" />
            <input v-model="password" :type="show ? 'text' : 'password'" autocomplete="current-password" required :readonly="verified" />
            <SIcon v-if="verified" class="ok" data-live name="check" :size="16" />
            <button v-else type="button" class="eye" :title="show ? t('studio.login.hide') : t('studio.login.show')" @click="show = !show">
              <SIcon :name="show ? 'eyeOff' : 'eye'" :size="16" />
            </button>
          </span>
        </label>
        <p class="err" :class="{ on: !!error }" role="alert">{{ error || '&nbsp;' }}</p>
        <button type="submit" class="st-btn p lg go" :class="{ passed: verified }" :disabled="busy || verified">
          <template v-if="verified">{{ t('studio.login.verified') }}<SIcon name="check" :size="16" data-live /></template>
          <template v-else>{{ busy ? t('studio.login.busy') : t('studio.login.enter') }}<SIcon name="arrowR" :size="16" /></template>
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
  border-radius: var(--r-xl);
  background: var(--paper);
  box-shadow: var(--sh-paper);
  overflow: hidden;
}


.art {
  position: relative;
  min-height: 520px;

  .door { position: absolute; inset: 0; }

  /* 底部压暗，保证格言文字在任意封面上可读 */
  &::after {
    content: '';
    position: absolute;
    inset: 45% 0 0;
    background: linear-gradient(to top, rgb(10 14 22 / 0.7), transparent);
    pointer-events: none;
  }

  .quote { z-index: 1; }

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

  .logo { width: 40px; height: 40px; object-fit: cover; border-radius: var(--r-md); box-shadow: 0 6px 14px -6px rgba(10, 20, 40, 0.5); }

  h1 { font: 600 26px/1.3 var(--font-serif); margin: 6px 0 0; }
  .sub { margin: -8px 0 8px; font-size: 13.5px; color: var(--st-ink-3); }

  label { display: block; }

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

  .go {
    width: 100%;

    .st-ic { transition: transform var(--dur) var(--ease-spring); }
    &:hover .st-ic { transform: translateX(3px); }
  }

  .back { align-self: center; font-size: 12.5px; margin-top: 4px; }

  /* 验证通过：对勾用信号色 --ink 弹入；按钮保持 --solid 实底，禁用态不减淡 */
  .ok { color: var(--ink); animation: ok-pop var(--dur) var(--ease-spring) both; }

  .go.passed {
    opacity: 1;
    cursor: default;

    .st-ic { color: inherit; animation: ok-pop var(--dur) var(--ease-spring) both; }
  }
}

@keyframes ok-pop { from { opacity: 0; transform: scale(0.4); } }

@media (max-width: 860px) {
  .card { grid-template-columns: 1fr; width: min(420px, 100%); }
  .art { min-height: 180px; }
}
</style>
