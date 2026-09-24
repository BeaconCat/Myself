<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { ContactData } from '../types';
import type { ModProps } from './props';
import KitIcon from '../parts/KitIcon.vue';
import { toast } from '../toast';

/** 联系（contact）：宋体大标题、邮箱一键复制、主按钮、回复时效；右上一道光缝 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as ContactData);
const { t } = useI18n();
const copied = ref(false);

async function copy(): Promise<void> {
  if (!d.value.email) return;
  try { await navigator.clipboard?.writeText(d.value.email); } catch { /* 剪贴板不可用时仍给反馈 */ }
  copied.value = true;
  toast(t('aboutKit.contact.copied'));
  window.setTimeout(() => { copied.value = false; }, 1800);
}
</script>

<template>
  <div class="ct" :class="{ wide: variant === 'wide' }">
    <span class="ct-glow" aria-hidden="true" />
    <span class="ct-slit" aria-hidden="true" />
    <div class="ct-head">
      <h3>{{ d.title }}</h3>
      <p>{{ d.text }}</p>
    </div>
    <div class="ct-foot">
      <div v-if="d.email" class="ct-mail">
        <span>{{ d.email }}</span>
        <button class="cp" :class="{ ok: copied }" @click="copy">
          <KitIcon :name="copied ? 'ok' : 'copy'" :size="14" />{{ copied ? t('aboutKit.contact.done') : t('aboutKit.contact.copy') }}
        </button>
      </div>
      <div class="ct-row">
        <a v-if="d.url" class="ak-btn pri" :href="d.url"><KitIcon name="mail" />{{ d.buttonText }}</a>
        <small v-if="d.sla"><span class="ak-dot live" />{{ d.sla }}</small>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.ct { position: relative; display: flex; flex-direction: column; flex: 1; min-height: 260px; }

.ct-glow {
  position: absolute;
  right: -28px;
  top: -28px;
  width: 240px;
  height: 240px;
  background: radial-gradient(closest-side, rgba(var(--primary-rgb), 0.28), transparent);
  pointer-events: none;
}

.ct-slit {
  position: absolute;
  right: 14px;
  top: -4px;
  width: 3px;
  height: 74px;
  border-radius: 2px;
  background: linear-gradient(#fff, rgb(var(--primary-rgb)));
  box-shadow: 0 0 18px 2px rgba(var(--primary-rgb), 0.8);
}

.ct-head { position: relative; }
.ct h3 { font: 700 30px/1.25 var(--font-serif); letter-spacing: 0.02em; }
.ct p { margin-top: 10px; font-size: 14px; color: var(--text-2); }

.ct-foot { position: relative; margin-top: auto; padding-top: 28px; }

.ct-mail {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 6px 6px 14px;
  border-radius: 13px;
  background: var(--ak-sunken);
  border: 1px solid var(--ak-line);

  span { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font: 500 13px var(--ak-mono); }
}

.cp {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 12px;
  border-radius: 9px;
  font-size: 12.5px;
  color: var(--text-2);
  border: 1px solid var(--ak-line-2) !important;
  transition: color var(--dur-fast), border-color var(--dur-fast);

  &:hover { color: var(--text); }
  &.ok { color: var(--ak-green); border-color: rgb(0 200 83 / 0.5) !important; }
}

.ct-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px 12px;
  margin-top: 12px;

  small { display: flex; align-items: center; gap: 7px; font-size: 12px; color: var(--ak-text-3); }
}

.ct.wide {
  flex-direction: row;
  align-items: center;
  gap: 40px;
  min-height: 0;

  .ct-head { flex: 1; }
  .ct-foot { margin-top: 0; padding-top: 0; min-width: 330px; }
}

@container (max-width: 640px) {
  .ct.wide { flex-direction: column; align-items: stretch; gap: 20px; }
  .ct.wide .ct-foot { min-width: 0; }
}
</style>
