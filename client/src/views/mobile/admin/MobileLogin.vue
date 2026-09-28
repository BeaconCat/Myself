<script setup lang="ts">
/**
 * 移动端登录：logo + 衬线大标题 + iOS 分组输入框 + 实底主按钮；错误时卡片抖动。
 * 验证通过：表单原位切到「验证通过」态（输入行右侧对勾、按钮换文案带对勾），立即跳转后台，
 * 由路由遮罩盖住整页一起离场；欢迎提示在揭幕开始时弹出。
 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../../api';
import { useAuthStore } from '../../../stores/auth';
import { useConfigStore } from '../../../stores/config';
import { REVEAL_EVENT } from '../../../stores/loading';
import MaIcon from '../../../components/mobile-admin/MaIcon.vue';
import MaIsland from '../../../components/mobile-admin/MaIsland.vue';
import MaRing from '../../../components/mobile-admin/MaRing.vue';
import { toast } from '../../../components/mobile-admin/state';

const { t } = useI18n();
const router = useRouter();
const auth = useAuthStore();
const config = useConfigStore();

const username = ref('');
const password = ref('');
const reveal = ref(false);
const busy = ref(false);
/** 报错抖动用 WAAPI 单独播放，不替换 CSS 入场动画（否则移除抖动类时入场会重播） */
const cardEl = ref<HTMLElement | null>(null);
function shake(): void {
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
  cardEl.value?.animate(
    [
      { transform: 'none' },
      { transform: 'translateX(-10px)', offset: 0.2 },
      { transform: 'translateX(8px)', offset: 0.4 },
      { transform: 'translateX(-5px)', offset: 0.6 },
      { transform: 'translateX(3px)', offset: 0.8 },
      { transform: 'none' },
    ],
    { duration: 500, easing: 'cubic-bezier(.2,.8,.3,1)', composite: 'add' },
  );
}
const success = ref(false);
const error = ref('');

const canSubmit = computed(() => !!username.value && !!password.value && !busy.value && !success.value);

async function submit(): Promise<void> {
  if (!canSubmit.value) return;
  busy.value = true;
  error.value = '';
  try {
    const res = await adminApi.login(username.value, password.value);
    auth.setToken(res.token);
    success.value = true;
    (document.activeElement as HTMLElement | null)?.blur();
    if (res.mustChange) {
      void router.replace({ path: '/setup', query: { change: '1' } });
      return;
    }
    const name = config.cfg.about.name || '';
    window.addEventListener(REVEAL_EVENT, () => toast(t('mobileAdmin.login.welcome'), name), { once: true });
    void router.replace('/admin');
  } catch (e) {
    error.value = (e as Error).message === 'too_many_attempts' ? t('mobileAdmin.login.locked') : t('mobileAdmin.login.failed');
    shake();
    navigator.vibrate?.([12, 40, 12]);
  } finally {
    busy.value = false;
  }
}

onMounted(() => document.documentElement.classList.add('ma-lock'));
onBeforeUnmount(() => document.documentElement.classList.remove('ma-lock'));
</script>

<template>
  <div class="ma-root ml" :class="{ success }">
    <main class="ml-main">
      <div class="brand">
        <img class="logo" src="/favicon-256.png" alt="" draggable="false" />
      </div>
      <h1>{{ t('mobileAdmin.login.title') }}</h1>
      <p class="sub">{{ config.cfg.site.title }} · {{ t('mobileAdmin.login.sub') }}</p>

      <form ref="cardEl" class="card" @submit.prevent="submit">
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
              :readonly="success"
            />
            <MaIcon v-if="success" class="ok" data-live name="check" :size="18" />
          </label>
          <label class="row">
            <MaIcon name="lock" :size="19" />
            <input
              v-model="password"
              :type="reveal ? 'text' : 'password'"
              autocomplete="current-password"
              enterkeyhint="go"
              :placeholder="t('mobileAdmin.login.password')"
              :readonly="success"
            />
            <MaIcon v-if="success" class="ok" data-live name="check" :size="18" />
            <button v-else type="button" class="eye tap" :aria-label="t('mobileAdmin.login.reveal')" @click="reveal = !reveal">
              <MaIcon name="eye" :size="18" :class="{ off: !reveal }" />
            </button>
          </label>
        </div>
        <p class="err" :class="{ on: !!error }" role="alert">{{ error }}</p>
        <button class="go tap" :class="{ passed: success }" type="submit" :disabled="!canSubmit">
          <template v-if="success">
            <span>{{ t('mobileAdmin.login.verified') }}</span>
            <MaIcon class="go-ok" name="check" data-live :size="18" />
          </template>
          <template v-else>
            <MaRing v-if="busy" indeterminate :size="18" :stroke="2.2" class="go-ring" />
            <span>{{ busy ? t('mobileAdmin.login.loggingIn') : t('mobileAdmin.login.submit') }}</span>
          </template>
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

  .logo {
    position: relative;
    width: 100%;
    height: 100%;
    border-radius: var(--r-xl);
    box-shadow: var(--shadow-card);
    transition: transform 0.6s var(--ease-sheet);
  }
}

h1 {
  margin-top: 26px;
  text-align: center;
  font-family: var(--font-serif);
  font-size: 32px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.sub {
  margin-top: 6px;
  text-align: center;
  font-size: 14px;
  color: var(--text-3);
}


.card {
  margin-top: 34px;

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
    caret-color: var(--ink);

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
  border-radius: var(--r-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 17px;
  font-weight: 600;
  color: var(--on-solid);
  background: var(--solid);
  box-shadow: var(--btn-shadow);
  transition: opacity var(--dur), background-color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  &:active:not(:disabled) { transform: scale(0.97); background: var(--solid-hover); }
  &:disabled { opacity: 0.45; box-shadow: none; }

  .go-ring {
    --ring-bg: color-mix(in oklab, var(--on-solid) 30%, transparent);
    --ring-fg: var(--on-solid);
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

/* 验证通过：原位对勾（信号色 --ink 弹入），按钮保持 --solid 实底不减淡；整页随路由遮罩离场 */
.ok {
  color: var(--ink);
  animation: ok-pop var(--dur) var(--ease-spring) both;
}

.go.passed {
  opacity: 1;
  box-shadow: var(--btn-shadow);

  .go-ok { animation: ok-pop var(--dur) var(--ease-spring) both; }
}

@keyframes ok-pop { from { opacity: 0; transform: scale(0.4); } }
</style>
