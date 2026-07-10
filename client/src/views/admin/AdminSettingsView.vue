<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../api';

const { t } = useI18n();

const oldPassword = ref('');
const newPassword = ref('');
const confirm = ref('');
const message = ref('');
const ok = ref(false);
const busy = ref(false);

async function submit(): Promise<void> {
  if (busy.value) return;
  message.value = '';
  if (newPassword.value.length < 8) {
    message.value = t('admin.pwTooShort');
    return;
  }
  if (newPassword.value !== confirm.value) {
    message.value = t('admin.pwMismatch');
    return;
  }
  busy.value = true;
  try {
    await adminApi.changePassword(oldPassword.value, newPassword.value);
    ok.value = true;
    message.value = t('admin.pwChanged');
    oldPassword.value = '';
    newPassword.value = '';
    confirm.value = '';
  } catch {
    ok.value = false;
    message.value = t('admin.pwWrong');
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div>
    <h1 class="page-h">{{ t('admin.menuSettings') }}</h1>

    <form class="card" @submit.prevent="submit">
      <h2>{{ t('admin.changePassword') }}</h2>
      <label>
        <span>{{ t('admin.oldPassword') }}</span>
        <input v-model="oldPassword" type="password" autocomplete="current-password" required />
      </label>
      <label>
        <span>{{ t('admin.newPassword') }}</span>
        <input v-model="newPassword" type="password" autocomplete="new-password" required />
      </label>
      <label>
        <span>{{ t('admin.confirmPassword') }}</span>
        <input v-model="confirm" type="password" autocomplete="new-password" required />
      </label>

      <p v-if="message" class="msg" :class="{ ok }">{{ message }}</p>

      <button class="btn primary" type="submit" :disabled="busy">{{ t('admin.save') }}</button>
    </form>
  </div>
</template>

<style scoped lang="scss">
.page-h {
  font-size: 26px;
  margin-bottom: 22px;
}

.card {
  width: min(420px, 100%);
  padding: 26px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  display: flex;
  flex-direction: column;
  gap: 16px;

  h2 { font-size: 18px; }
}

label {
  display: flex;
  flex-direction: column;
  gap: 6px;

  span {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-2);
  }
}

input {
  padding: 10px 13px;
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
    box-shadow: 0 0 0 3px rgba(var(--primary-rgb), 0.12);
  }
}

.msg {
  font-size: 13px;
  color: var(--accent-red);

  &.ok { color: var(--primary); }
}

.btn.primary {
  padding: 11px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 700;
  color: #fff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
  background: linear-gradient(180deg, var(--primary), var(--primary-deep));
  box-shadow: 0 4px 14px rgba(var(--primary-rgb), 0.4);
  transition: transform var(--dur-fast) var(--ease-out), filter var(--dur-fast);

  &:hover:not(:disabled) { filter: brightness(1.08); transform: scale(1.02); }
  &:disabled { opacity: 0.6; }
}
</style>
