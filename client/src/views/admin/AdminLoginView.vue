<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';
import { useAuthStore } from '../../stores/auth';

const { t } = useI18n();
const router = useRouter();
const auth = useAuthStore();

const username = ref('');
const password = ref('');
const error = ref('');
const busy = ref(false);

async function submit(): Promise<void> {
  if (busy.value) return;
  busy.value = true;
  error.value = '';
  try {
    auth.setToken(await adminApi.login(username.value, password.value));
    void router.push('/admin/posts');
  } catch {
    error.value = t('admin.loginFailed');
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <main class="page">
    <form v-reveal class="login-card" @submit.prevent="submit">
      <img class="logo" src="/favicon-256.png" alt="" draggable="false" />
      <h1>{{ t('admin.loginTitle') }}</h1>

      <label>
        <span>{{ t('admin.username') }}</span>
        <input v-model="username" type="text" autocomplete="username" required />
      </label>
      <label>
        <span>{{ t('admin.password') }}</span>
        <input v-model="password" type="password" autocomplete="current-password" required />
      </label>

      <p v-if="error" class="error">{{ error }}</p>

      <button class="submit" type="submit" :disabled="busy">
        {{ busy ? t('admin.loggingIn') : t('admin.login') }}
      </button>
    </form>
  </main>
</template>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
}

.login-card {
  width: min(380px, 100%);
  padding: 40px 34px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow);
  display: flex;
  flex-direction: column;
  gap: 18px;

  .logo {
    width: 56px;
    height: 56px;
    margin: 0 auto;
  }

  h1 {
    text-align: center;
    font-size: 22px;
    margin-bottom: 6px;
  }
}

label {
  display: flex;
  flex-direction: column;
  gap: 7px;

  span {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-2);
  }
}

input {
  padding: 11px 14px;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text);
  font-size: 14px;
  font-family: inherit;
  outline: none;
  transition: border-color var(--dur-fast), box-shadow var(--dur-fast);

  &:focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px rgba(var(--primary-rgb), 0.15);
  }
}

.error {
  font-size: 13px;
  color: var(--accent-red);
}

.submit {
  padding: 12px;
  border: none;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 700;
  color: #fff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
  background: linear-gradient(180deg, var(--primary), var(--primary-deep));
  box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.45);
  transition: transform var(--dur-fast) var(--ease-out), filter var(--dur-fast);

  &:hover:not(:disabled) { filter: brightness(1.08); transform: scale(1.02); }
  &:disabled { opacity: 0.6; }
}
</style>
