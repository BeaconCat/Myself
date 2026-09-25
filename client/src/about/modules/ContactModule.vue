<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { ContactData } from '../types';
import type { ModProps } from './props';
import KitIcon from '../parts/KitIcon.vue';
import { toast } from '../toast';

/** 联系（contact）：宋体大标题、邮箱一键复制（实底主按钮）、写信按钮、回复时效；右上一道主色信号线 */
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
    <span class="ct-slit" aria-hidden="true" />
    <div class="ct-head">
      <h3>{{ d.title }}</h3>
      <p>{{ d.text }}</p>
    </div>
    <div class="ct-foot">
      <div v-if="d.email" class="ct-mail">
        <span>{{ d.email }}</span>
        <button class="ak-btn pri cp" :class="{ ok: copied }" @click="copy">
          <KitIcon :name="copied ? 'ok' : 'copy'" :size="14" />{{ copied ? t('aboutKit.contact.done') : t('aboutKit.contact.copy') }}
        </button>
      </div>
      <div class="ct-row">
        <a v-if="d.url" class="ak-btn" :href="d.url"><KitIcon name="mail" />{{ d.buttonText }}</a>
        <small v-if="d.sla"><span class="ak-dot live" />{{ d.sla }}</small>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.ct { position: relative; display: flex; flex-direction: column; flex: 1; min-height: 260px; }

.ct-slit {
  position: absolute;
  right: 14px;
  top: -4px;
  width: 2px;
  height: 64px;
  border-radius: var(--r-pill);
  background: linear-gradient(var(--ink), transparent);
}

.ct-head { position: relative; padding-right: 24px; }
.ct h3 { font: 700 30px/1.25 var(--font-serif); letter-spacing: 0.02em; }
.ct p { margin-top: 10px; font-size: 14px; color: var(--text-2); }

.ct-foot { position: relative; margin-top: auto; padding-top: 28px; }

.ct-mail {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 5px 5px 16px;
  border-radius: var(--r-pill);
  background: var(--ak-sunken);
  box-shadow: inset 0 0 0 1px var(--ak-line);

  span { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font: 500 13px var(--ak-mono); }
}

.cp {
  gap: 6px;
  height: 32px;
  padding: 0 14px;
  font-size: 12.5px;
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
