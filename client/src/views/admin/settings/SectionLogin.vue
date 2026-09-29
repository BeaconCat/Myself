<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { SiteConfig } from '../../../stores/config';
import '../studio/i18n';
import SIcon from '../studio/SIcon.vue';
import StSwitch from '../studio/StSwitch.vue';
import { copyText } from '../studio/state';
import { toast } from '../studio/toast';

/**
 * 登录方式：GitHub OAuth。需要在 GitHub 创建 OAuth App，把下面的回调地址填进去，
 * 再把 Client ID / Secret 填回这里；用户系统总开关打开后前台才出现「用 GitHub 登录」。
 */
const props = defineProps<{ cfg: SiteConfig }>();
const { t } = useI18n();

const oauth = computed(() => props.cfg.oauth!.github);
const login = computed(() => props.cfg.users!.login);
const showSecret = ref(false);

/** 与服务端 siteBase 一致：配置了站点地址用站点地址，否则用当前访问的地址 */
const base = computed(() => (props.cfg.site.url?.startsWith('http') ? props.cfg.site.url.replace(/\/+$/, '') : window.location.origin));
const callback = computed(() => `${base.value}/api/v1/auth/github/callback`);
const ready = computed(() => login.value.github && !!oauth.value.clientId && !!oauth.value.clientSecret);

async function copy(): Promise<void> {
  toast((await copyText(callback.value)) ? t('studio.copied') : t('studio.users.copyFailed'), { icon: 'copy' });
}
</script>

<template>
  <div class="st-opt">
    <div>{{ t('studio.settings.ghLogin') }}<small>{{ t('studio.settings.ghLoginSub') }}</small></div>
    <StSwitch v-model="login.github" />
  </div>

  <div class="fold" :class="{ open: login.github }" :inert="!login.github">
    <div class="clip">
      <ol class="steps">
        <li>
          <span>{{ t('studio.settings.ghStep1') }}</span>
          <a class="st-link" href="https://github.com/settings/applications/new" target="_blank" rel="noopener">
            github.com/settings/applications/new<SIcon name="external" :size="14" />
          </a>
        </li>
        <li>
          <span>{{ t('studio.settings.ghStep2') }}</span>
          <div class="cb">
            <code>{{ callback }}</code>
            <button type="button" class="st-btn g sm" @click="copy"><SIcon name="copy" :size="15" />{{ t('studio.copy') }}</button>
          </div>
          <small v-if="!cfg.site.url">{{ t('studio.settings.ghCallbackHint') }}</small>
        </li>
        <li><span>{{ t('studio.settings.ghStep3') }}</span></li>
      </ol>

      <div class="grid">
        <label>
          <span class="st-flabel">Client ID</span>
          <span class="st-field"><input v-model.trim="oauth.clientId" spellcheck="false" autocomplete="off" /></span>
        </label>
        <label>
          <span class="st-flabel">Client Secret</span>
          <span class="st-field">
            <input v-model.trim="oauth.clientSecret" :type="showSecret ? 'text' : 'password'" autocomplete="new-password" spellcheck="false" />
            <button type="button" class="eye" :title="showSecret ? t('studio.login.hide') : t('studio.login.show')" @click="showSecret = !showSecret">
              <SIcon :name="showSecret ? 'eyeOff' : 'eye'" :size="16" />
            </button>
          </span>
        </label>
      </div>

      <p class="state" :class="{ ok: ready && cfg.users?.enabled }">
        <SIcon :name="ready && cfg.users?.enabled ? 'check' : 'info'" :size="16" />
        {{ !ready ? t('studio.settings.ghNotReady') : cfg.users?.enabled ? t('studio.settings.ghReady') : t('studio.settings.ghNeedsUsers') }}
      </p>
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

.steps {
  margin: 16px 0 0;
  padding: 0;
  list-style: none;
  counter-reset: step;
  display: flex;
  flex-direction: column;
  gap: 12px;
  font-size: 14px;
  color: var(--st-ink-2);

  /* 序号画在盒内（折叠容器 overflow:hidden 会裁掉盒外的 marker） */
  li {
    position: relative;
    padding-left: 30px;
    counter-increment: step;

    &::before {
      content: counter(step);
      position: absolute;
      left: 0;
      top: 1px;
      width: 20px;
      height: 20px;
      display: grid;
      place-items: center;
      border-radius: 50%;
      background: var(--well-2);
      font: 600 11.5px var(--font-mono);
      color: var(--st-ink-2);
    }
  }
  li > span { display: block; margin-bottom: 6px; }
  .st-link { display: inline-flex; align-items: center; gap: 4px; font-family: var(--font-mono); font-size: 13px; }
  small { display: block; margin-top: 6px; font-size: 12.5px; color: var(--st-ink-3); }
}

.cb {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 8px 8px 12px;
  border-radius: var(--r-md);
  background: var(--well);
  box-shadow: 0 0 0 1px var(--line-2);

  code { flex: 1; min-width: 0; font: 12.5px var(--font-mono); word-break: break-all; color: var(--st-ink); }
}

.grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px 16px;
  margin-top: 16px;

  label { display: flex; flex-direction: column; min-width: 0; }
}

.eye {
  display: grid;
  place-items: center;
  color: var(--st-ink-3);

  &:hover { color: var(--st-ink); }
}

.state {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 14px 0 0;
  font-size: 13.5px;
  color: var(--st-ink-3);

  &.ok { color: color-mix(in oklab, var(--green) 70%, var(--st-ink)); }
}

@media (prefers-reduced-motion: reduce) {
  .fold { transition: none; }
}
</style>
