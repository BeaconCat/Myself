<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../../api';

/** 用户管理：修改管理员密码（自持状态，不依赖 cfg） */
const { t } = useI18n();

const oldPassword = ref('');
const newPassword = ref('');
const confirmPw = ref('');
const pwMessage = ref('');
const pwOk = ref(false);

async function changePassword(): Promise<void> {
  pwMessage.value = '';
  if (newPassword.value.length < 8) {
    pwMessage.value = t('admin.pwTooShort');
    return;
  }
  if (newPassword.value !== confirmPw.value) {
    pwMessage.value = t('admin.pwMismatch');
    return;
  }
  try {
    await adminApi.changePassword(oldPassword.value, newPassword.value);
    pwOk.value = true;
    pwMessage.value = t('admin.pwChanged');
    oldPassword.value = newPassword.value = confirmPw.value = '';
  } catch {
    pwOk.value = false;
    pwMessage.value = t('admin.pwWrong');
  }
}
</script>

<template>
  <section id="sec-users" class="card">
    <h2>{{ t('admin.secUsers') }}</h2>
    <p class="hint">{{ t('admin.usersHint') }}</p>
    <div class="pw-grid">
      <label><span>{{ t('admin.oldPassword') }}</span><input v-model="oldPassword" type="password" autocomplete="current-password" /></label>
      <label><span>{{ t('admin.newPassword') }}</span><input v-model="newPassword" type="password" autocomplete="new-password" /></label>
      <label><span>{{ t('admin.confirmPassword') }}</span><input v-model="confirmPw" type="password" autocomplete="new-password" /></label>
    </div>
    <div class="pw-actions">
      <button class="btn ghost" @click="changePassword">{{ t('admin.changePassword') }}</button>
      <span v-if="pwMessage" class="msg" :class="{ err: !pwOk }">{{ pwMessage }}</span>
    </div>
  </section>
</template>

<style scoped lang="scss">
@use './settings-shared';

.pw-actions {
  display: flex;
  align-items: center;
  gap: 14px;
}
</style>
