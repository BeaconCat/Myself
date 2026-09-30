<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi } from '../../../api';
import type { MailConfig } from '../../../stores/config';
import '../studio/i18n';
import SIcon from '../studio/SIcon.vue';
import StSeg from '../studio/StSeg.vue';
import StSwitch from '../studio/StSwitch.vue';

/**
 * 邮件（SMTP）：注册验证、找回密码、邀请与重置链接都经由这里发信。
 * 测试发信使用已保存的配置，所以有未保存改动时先提示保存。
 */
const props = defineProps<{ mail: MailConfig; dirty: boolean }>();
const { t } = useI18n();

const PRESET_PORT: Record<MailConfig['security'], number> = { starttls: 587, tls: 465, none: 25 };
const security = computed({
  get: () => props.mail.security,
  set: (v: MailConfig['security']) => {
    // 端口仍是上一个模式的默认值时跟着切换
    if (props.mail.port === PRESET_PORT[props.mail.security]) props.mail.port = PRESET_PORT[v];
    props.mail.security = v;
  },
});

const showPw = ref(false);
const testTo = ref('');
const testing = ref(false);
const result = ref<{ ok: boolean; text: string } | null>(null);

async function test(): Promise<void> {
  if (testing.value || !testTo.value.trim()) return;
  if (props.dirty) {
    result.value = { ok: false, text: t('studio.settings.mailSaveFirst') };
    return;
  }
  testing.value = true;
  result.value = null;
  try {
    await adminApi.mailTest(testTo.value.trim());
    result.value = { ok: true, text: t('studio.settings.mailTestOk', { to: testTo.value.trim() }) };
  } catch (e) {
    const code = (e as Error).message;
    result.value = { ok: false, text: code === 'invalid_email' ? t('studio.settings.mailBadTo') : t('studio.settings.mailTestFail', { reason: code }) };
  } finally {
    testing.value = false;
  }
}
</script>

<template>
  <div class="st-opt">
    <div>{{ t('studio.settings.mailOn') }}<small>{{ t('studio.settings.mailOnSub') }}</small></div>
    <StSwitch v-model="mail.enabled" :label="t('studio.settings.mailOn')" />
  </div>

  <div class="fold" :class="{ open: mail.enabled }" :inert="!mail.enabled">
    <div class="clip">
      <div class="grid">
        <label class="c8">
          <span class="st-flabel">{{ t('studio.settings.mailHost') }}</span>
          <span class="st-field"><input v-model.trim="mail.host" placeholder="smtp.example.com" spellcheck="false" /></span>
        </label>
        <label class="c4">
          <span class="st-flabel">{{ t('studio.settings.mailPort') }}</span>
          <span class="st-field"><input v-model.number="mail.port" type="number" min="1" max="65535" /></span>
        </label>
        <div class="c12">
          <span class="st-flabel">{{ t('studio.settings.mailSecurity') }}</span>
          <StSeg
            v-model="security"
            :options="[
              { value: 'starttls', label: 'STARTTLS' },
              { value: 'tls', label: 'SSL / TLS' },
              { value: 'none', label: t('studio.settings.mailPlain') },
            ]"
          />
        </div>
        <label class="c6">
          <span class="st-flabel">{{ t('studio.settings.mailUser') }}</span>
          <span class="st-field"><input v-model.trim="mail.username" autocomplete="off" spellcheck="false" /></span>
        </label>
        <label class="c6">
          <span class="st-flabel">{{ t('studio.settings.mailPass') }}<em>{{ t('studio.settings.mailPassHint') }}</em></span>
          <span class="st-field">
            <input v-model="mail.password" :type="showPw ? 'text' : 'password'" autocomplete="new-password" />
            <button type="button" class="eye" :title="showPw ? t('studio.login.hide') : t('studio.login.show')" @click="showPw = !showPw">
              <SIcon :name="showPw ? 'eyeOff' : 'eye'" :size="16" />
            </button>
          </span>
        </label>
        <label class="c12">
          <span class="st-flabel">{{ t('studio.settings.mailFrom') }}<em>{{ t('studio.settings.mailFromHint') }}</em></span>
          <span class="st-field"><input v-model.trim="mail.from" placeholder="Myself <no-reply@example.com>" spellcheck="false" /></span>
        </label>
      </div>

      <div class="test">
        <span class="st-flabel">{{ t('studio.settings.mailTest') }}</span>
        <div class="row">
          <label class="st-field"><SIcon name="mail" :size="16" /><input v-model="testTo" :aria-label="t('studio.settings.mailTest')" type="email" placeholder="you@example.com" @keydown.enter="test" /></label>
          <button type="button" class="st-btn g" :disabled="testing || !testTo.trim()" @click="test">
            <SIcon name="send" :size="16" />{{ testing ? t('studio.settings.mailTesting') : t('studio.settings.mailTestSend') }}
          </button>
        </div>
        <Transition name="msg">
          <p v-if="result" class="msg" :class="{ ok: result.ok }"><SIcon :name="result.ok ? 'check' : 'info'" :size="16" />{{ result.text }}</p>
        </Transition>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.fold {
  display: grid;
  grid-template-rows: 0fr;
  opacity: 0;
  transition: grid-template-rows var(--dur-slow) var(--ease-out), opacity var(--dur) var(--ease-out);

  &.open { grid-template-rows: 1fr; opacity: 1; }
}

.clip { min-height: 0; overflow: hidden; }

.grid {
  display: grid;
  grid-template-columns: repeat(12, minmax(0, 1fr));
  gap: 14px 16px;
  padding-top: 16px;

  .c4 { grid-column: span 4; }
  .c6 { grid-column: span 6; }
  .c8 { grid-column: span 8; }
  .c12 { grid-column: span 12; }

  label, > div { display: flex; flex-direction: column; min-width: 0; }
}

.st-flabel em { margin-left: 6px; font-style: normal; font-weight: 400; color: var(--st-ink-3); }

.eye {
  display: grid;
  place-items: center;
  color: var(--st-ink-3);

  &:hover { color: var(--st-ink); }
}

.test {
  margin-top: 18px;
  padding: 14px 16px 16px;
  border-radius: var(--r-md);
  background: var(--well);

  .row { display: flex; gap: 10px; }
  .st-field { flex: 1; gap: 8px; color: var(--st-ink-3); }
}

.msg {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin: 10px 0 0;
  font-size: 13.5px;
  line-height: 1.6;
  color: color-mix(in oklab, var(--red) 70%, var(--st-ink));
  word-break: break-word;

  &.ok { color: color-mix(in oklab, var(--green) 70%, var(--st-ink)); }
  :deep(svg) { flex: none; margin-top: 3px; }
}

.msg-enter-active, .msg-leave-active { transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-out); }
.msg-enter-from, .msg-leave-to { opacity: 0; transform: translateY(-4px); }

@media (prefers-reduced-motion: reduce) {
  .fold { transition: none; }
}
</style>
