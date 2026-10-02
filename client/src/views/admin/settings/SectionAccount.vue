<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../../api';
import { useAuthStore } from '../../../stores/auth';
import '../studio/i18n';
import SIcon from '../studio/SIcon.vue';
import { toast } from '../studio/toast';

/** 账号：修改管理员密码（独立提交，不随站点设置保存） */
const { t } = useI18n();
const auth = useAuthStore();

const oldPw = ref('');
const newPw = ref('');
const confirmPw = ref('');
const busy = ref(false);
const error = ref('');

const strength = computed(() => {
  const v = newPw.value;
  let s = 0;
  if (v.length >= 8) s += 1;
  if (v.length >= 12) s += 1;
  if (/[A-Z]/.test(v) && /[a-z]/.test(v)) s += 1;
  if (/\d/.test(v) && /[^\w]/.test(v)) s += 1;
  return v ? Math.max(1, s) : 0;
});

async function submit(): Promise<void> {
  error.value = '';
  if (newPw.value.length < 8) {
    error.value = t('studio.settings.pwShort');
    return;
  }
  if (newPw.value !== confirmPw.value) {
    error.value = t('studio.settings.pwMismatch');
    return;
  }
  busy.value = true;
  try {
    await adminApi.changePassword(oldPw.value, newPw.value);
    // 改密后其它会话全部失效；当前会话已由服务端换发新 Cookie
    auth.markLoggedIn();
    oldPw.value = newPw.value = confirmPw.value = '';
    toast(t('studio.settings.pwChanged'), { icon: 'lock' });
  } catch (e) {
    error.value = (e as Error).message === 'too_many_attempts' ? t('studio.login.locked') : t('studio.settings.pwWrong');
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <form class="acc" @submit.prevent="submit">
    <input type="text" autocomplete="username" value="admin" hidden />
    <label><span class="st-flabel">{{ t('studio.settings.pwOld') }}</span><span class="st-field"><SIcon name="lock" :size="18" /><input v-model="oldPw" type="password" autocomplete="current-password" /></span></label>
    <label>
      <span class="st-flabel">{{ t('studio.settings.pwNew') }}</span>
      <span class="st-field"><SIcon name="key" :size="18" /><input v-model="newPw" type="password" autocomplete="new-password" /></span>
      <span class="meter" :data-s="strength"><i /><i /><i /><i /></span>
    </label>
    <label><span class="st-flabel">{{ t('studio.settings.pwConfirm') }}</span><span class="st-field"><SIcon name="check" :size="18" /><input v-model="confirmPw" type="password" autocomplete="new-password" /></span></label>
    <div class="ft">
      <span v-if="error" class="err">{{ error }}</span>
      <span v-else class="hint">{{ t('studio.settings.pwHint') }}</span>
      <button type="submit" class="st-btn g" :disabled="busy || !oldPw || !newPw">{{ t('studio.settings.pwSubmit') }}</button>
    </div>
  </form>
</template>

<style scoped lang="scss">
.acc {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
  padding-top: 6px;

  label { display: block; }
}

.meter {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 4px;
  margin-top: 8px;

  i { height: 3px; border-radius: calc(var(--r-xs) / 2); background: var(--well-2); transition: background var(--dur); }

  &[data-s='1'] i:nth-child(-n + 1) { background: var(--red); }
  &[data-s='2'] i:nth-child(-n + 2) { background: var(--yellow); }
  &[data-s='3'] i:nth-child(-n + 3) { background: color-mix(in oklab, var(--green) 70%, var(--yellow)); }
  &[data-s='4'] i { background: var(--green); }
}

.ft {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;

  .hint { font-size: 12.5px; color: var(--st-ink-3); }
  .err { font-size: 13px; color: var(--red); }
  > button { margin-inline-start: auto; }
}

@media (max-width: 767px) {
  .acc { grid-template-columns: minmax(0, 1fr); }
  .ft { flex-wrap: wrap; }
}
</style>
