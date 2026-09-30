<script setup lang="ts">
import { safeHref } from '../../utils/safeUrl';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useConfigStore } from '../../stores/config';
import type { ContactData, SocialLink } from '../types';
import type { ModProps } from './props';
import KitIcon from '../parts/KitIcon.vue';
import { toast } from '../toast';
import { copyText } from '../../utils/clipboard';

/**
 * 联系（contact）：宋体大标题、其他渠道（取 profile 社交入口，不含邮件）、邮箱一键复制（实底主按钮）、
 * 写信按钮、回复时效；右上一道主色信号线。
 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as ContactData);
const { t } = useI18n();
const copied = ref(false);
const config = useConfigStore();

/** 其他渠道：复用身份区的社交入口（邮件已由下方复制框承担，排除 mailto） */
const channels = computed<SocialLink[]>(() => {
  const profile = config.cfg.about.modules?.find((m) => m.type === 'profile');
  const links = (profile?.data?.links ?? []) as SocialLink[];
  return links.filter((l) => l?.url && !/^mailto:/.test(l.url)).slice(0, props.variant === 'wide' ? 3 : 4);
});
const external = (url: string) => (/^https?:/.test(url) ? '_blank' : undefined);

async function copy(): Promise<void> {
  if (!d.value.email) return;
  await copyText(d.value.email);
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
    <ul v-if="channels.length" class="ct-ch">
      <li v-for="l in channels" :key="l.name + l.url">
        <a :href="safeHref(l.url)" :target="external(l.url)" rel="noopener noreferrer">
          <span class="ic"><KitIcon :name="l.icon" /></span>
          <span class="tx"><b>{{ l.name }}</b><small>{{ l.handle || l.url.replace(/^https?:\/\//, '') }}</small></span>
          <KitIcon class="arr" name="arrow" :size="16" />
        </a>
      </li>
    </ul>
    <div class="ct-foot">
      <div v-if="d.email" class="ct-mail">
        <span>{{ d.email }}</span>
        <button class="ak-btn pri cp" :class="{ ok: copied }" @click="copy">
          <KitIcon :name="copied ? 'ok' : 'copy'" :size="16" />{{ copied ? t('aboutKit.contact.done') : t('aboutKit.contact.copy') }}
        </button>
      </div>
      <div class="ct-row">
        <a v-if="d.url" class="ak-btn" :href="safeHref(d.url)"><KitIcon name="mail" />{{ d.buttonText }}</a>
        <small v-if="d.sla"><span class="ak-dot live" />{{ d.sla }}</small>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.ct { position: relative; display: flex; flex-direction: column; flex: 1; min-height: 240px; }

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
.ct h3 { font: 700 26px/1.25 var(--font-serif); letter-spacing: 0.02em; }
.ct p { margin-top: 8px; font-size: 15px; line-height: 1.7; color: var(--text-2); }

/* 其他渠道：圆形图标 + 名称 + 账号，行高紧凑；悬停箭头右上飞出 */
.ct-ch {
  list-style: none;
  margin-top: 18px;

  a {
    display: grid;
    grid-template-columns: 38px minmax(0, 1fr) auto;
    gap: 12px;
    align-items: center;
    padding: 8px 10px;
    margin: 0 -10px;
    border-radius: var(--r-md);
    transition: background-color var(--dur-fast);

    &:hover { background: var(--ak-sunken); }
    &:hover .arr { transform: translate(2px, -2px); color: var(--text); }
  }

  .ic { display: grid; place-items: center; width: 38px; height: 38px; border-radius: var(--r-pill); background: var(--fill); color: var(--text-2); }
  .tx { min-width: 0; }
  b { display: block; font-size: 15px; font-weight: 500; line-height: 1.35; }
  small { display: block; overflow: hidden; font: 400 12.5px var(--ak-mono); color: var(--ak-text-3); white-space: nowrap; text-overflow: ellipsis; }
  .arr { color: var(--ak-text-3); transition: transform var(--dur) var(--ease-spring), color var(--dur); }
}

.ct-foot { position: relative; margin-top: auto; padding-top: 22px; }

.ct-mail {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 5px 5px 16px;
  border-radius: var(--r-pill);
  background: var(--ak-sunken);
  box-shadow: inset 0 0 0 1px var(--ak-line);

  span { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font: 500 14px var(--ak-mono); }
}

.cp {
  gap: 6px;
  height: 34px;
  padding: 0 14px;
  font-size: 13.5px;
}

.ct-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px 12px;
  margin-top: 12px;

  small { display: flex; align-items: center; gap: 7px; font-size: 13px; color: var(--ak-text-3); }
}

.ct.wide {
  flex-direction: row;
  align-items: center;
  gap: 40px;
  min-height: 0;

  .ct-head { flex: 1; min-width: 0; }
  .ct-slit { display: none; }
  .ct-ch { flex: none; margin: 0; }
  .ct-ch a { padding: 6px 10px; }
  .ct-foot { margin-top: 0; padding-top: 0; min-width: 340px; }
}

/* 中等宽度：标题独占一行，渠道与邮箱并排在下 */
@container (max-width: 1000px) {
  .ct.wide { flex-wrap: wrap; gap: 16px 28px; }
  .ct.wide .ct-head { flex: 1 0 100%; }
  .ct.wide .ct-ch { display: flex; gap: 4px 12px; }
  .ct.wide .ct-foot { flex: 1; }
}

@container (max-width: 640px) {
  .ct.wide { flex-direction: column; flex-wrap: nowrap; align-items: stretch; gap: 20px; }
  .ct.wide .ct-ch { flex-wrap: wrap; }
  /* 列向时上一档的 flex 基准（标题 100%、底部 0）会按高度计算：标题撑满、底部压扁溢出，这里统一回到内容高度 */
  .ct.wide .ct-head, .ct.wide .ct-foot { flex: none; }
  .ct.wide .ct-foot { min-width: 0; }
}
</style>
