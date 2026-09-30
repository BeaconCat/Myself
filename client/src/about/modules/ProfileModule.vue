<script setup lang="ts">
import { safeHref } from '../../utils/safeUrl';
import { computed } from 'vue';
import { useConfigStore } from '../../stores/config';
import type { ProfileData } from '../types';
import type { ModProps } from './props';
import KitIcon from '../parts/KitIcon.vue';
import { brandColor, cardLinks } from '../brands';
import { useI18n } from 'vue-i18n';
import { dayPartKey, pad2, useClock, zoned } from '../useClock';

/**
 * 身份区（profile）：编辑式大字宋体名字 + 一次性扫光、一句话、自述、
 * 实时状态行（正在做 / 城市 / 本地时间逐秒走）与社交入口；
 * 右侧形象图撑满右栏全尺寸显示，按 portrait.fade 左缘（或底缘）部分渐隐融入背景，未配置时用站点 logo。
 * 形象图圆角：data.portrait.radius（px）优先，未配置时跟随全局 --r-xl。
 * 社交入口：primary = 实底主按钮（--solid），其余为次级按钮。
 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as ProfileData);
const config = useConfigStore();
const { t } = useI18n();
const now = useClock();

const days = computed(() => {
  const from = new Date(`${config.cfg.about.foundedAt || '2026-01-01'}T00:00:00`).getTime();
  return Math.max(1, Math.floor((now.value - from) / 864e5));
});

const kicker = computed(() => d.value.kicker?.trim() || t('aboutKit.kicker', { n: days.value }));

/** *星号* 或 <em> 包裹的片段高亮 */
const lede = computed(() => {
  const src = (d.value.lede ?? '').replace(/<\/?em>/g, '*');
  return src.split('*').map((text, i) => ({ text, em: i % 2 === 1 })).filter((s) => s.text);
});

/** 名字字号：按字形宽度估算（汉字 1em、西文约 .58em），保证一行放下 */
const nameStyle = computed(() => {
  const w = [...(d.value.name ?? '')].reduce((a, ch) => a + (ch.charCodeAt(0) >= 0x2e80 ? 1 : 0.58), 0.45);
  return { '--fit': `${Math.min(17, 92 / Math.max(w, 1)).toFixed(2)}cqi` };
});

const tz = computed(() => Number(d.value.status?.tz ?? 8));
const clock = computed(() => zoned(now.value, tz.value));
const tzLabel = computed(() => `UTC${tz.value >= 0 ? '+' : ''}${tz.value}`);

const portrait = computed(() => {
  const p: Partial<ProfileData['portrait']> = d.value.portrait ?? {};
  const r = typeof p.radius === 'number' ? p.radius : null;
  return {
    src: p.src || config.cfg.site.logo || '/logo-1024.webp',
    // 内置 logo 图自带圆角底（按原样显示、不套圆角）；上传的站点 logo 与普通形象图一样跟随圆角设置
    logo: !p.src && !config.cfg.site.logo,
    fade: p.fade || 'left',
    style: {
      '--pr': r != null && Number.isFinite(r) ? `${r}px` : undefined,
      '--pf': p.focus || '50% 40%',
      '--fw': `${fadeWidth(p.fadeWidth)}%`,
    },
  };
});

/** 渐隐宽度：10–80%，缺省 50% */
function fadeWidth(v: unknown): number {
  const n = Number(v);
  return Number.isFinite(n) && n > 0 ? Math.min(80, Math.max(10, n)) : 50;
}

const plain = computed(() => props.variant === 'plain');
/** 名片按钮：身份页勾选的链接（最多 3 个），与移动端名片一致 */
const links = computed(() => cardLinks(d.value.links));
const external = (url: string) => (/^https?:/.test(url) ? '_blank' : undefined);
</script>

<template>
  <div class="pf" :class="{ plain }">
    <div class="pf-text">
      <p class="pf-kicker"><i />{{ kicker }}</p>
      <h1 class="pf-name" :style="nameStyle">
        <span v-if="d.hello" class="pf-hi">{{ d.hello }}</span>
        <span class="lit">{{ d.name }}</span><span class="dotp">.</span>
      </h1>
      <p v-if="d.alias?.trim()" class="pf-alias">{{ d.alias.trim() }}</p>
      <p v-if="lede.length" class="pf-lede">
        <template v-for="(s, i) in lede" :key="i"><em v-if="s.em">{{ s.text }}</em><template v-else>{{ s.text }}</template></template>
      </p>
      <p v-if="d.bio" class="pf-bio">{{ d.bio }}</p>

      <dl class="pf-status">
        <div v-if="d.status?.doing">
          <dt>{{ t('aboutKit.doing') }}</dt>
          <dd><span class="ak-dot live" />{{ d.status.doing }}</dd>
        </div>
        <div v-if="d.status?.city">
          <dt>{{ t('aboutKit.city') }}</dt>
          <dd>{{ d.status.city }}</dd>
        </div>
        <div>
          <dt>{{ t('aboutKit.localTime') }} · {{ tzLabel }}</dt>
          <dd>
            <span class="clock">{{ pad2(clock.h) }}<s>:</s>{{ pad2(clock.m) }}<s>:</s>{{ pad2(clock.s) }}</span>
            <span class="dayp">{{ t(dayPartKey(clock.h)) }}</span>
          </dd>
        </div>
      </dl>

      <nav v-if="links.length" class="pf-social">
        <a
          v-for="l in links"
          :key="l.name + l.url"
          :href="safeHref(l.url)"
          class="ak-btn"
          :class="{ pri: l.primary }"
          :style="{ '--bc': brandColor(l.icon) }"
          :target="external(l.url)"
          rel="noopener noreferrer"
        >
          <KitIcon :name="l.icon" />{{ l.name }}<span v-if="l.handle" class="ak-mono">{{ l.handle }}</span>
        </a>
      </nav>
    </div>

    <figure
      v-if="!plain"
      class="pf-portrait"
      :class="[`fade-${portrait.fade}`, { logo: portrait.logo }]"
      :style="portrait.style"
    >
      <span class="pf-frame">
        <img :src="portrait.src" :alt="d.name" draggable="false" />
      </span>
    </figure>
  </div>
</template>

<style scoped lang="scss">
/* 身份区（高密度）：去掉整屏留白，名字 / 一句话 / 自述 / 统计条式状态 / 社交紧凑成一个信息块，形象图撑满右栏 */
.pf {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 0.92fr);
  align-items: center;
  padding: 36px 0 4px;

  &.plain { grid-template-columns: minmax(0, 1fr); }
}

.pf-text { position: relative; z-index: 1; min-width: 0; container-type: inline-size; }

.pf-kicker {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 18px;
  font: 500 12.5px/1 var(--ak-mono);
  letter-spacing: 0.12em;
  color: var(--ak-text-3);
  text-transform: uppercase;

  i { width: 36px; height: 1px; background: var(--line-2); }
}

.pf-hi {
  display: block;
  margin-bottom: 10px;
  font: 500 20px/1.2 var(--font-serif);
  letter-spacing: 0.04em;
  color: var(--text-2);
}

.pf-name {
  position: relative;
  white-space: nowrap;
  font: 700 clamp(40px, var(--fit, 15cqi), 120px) / 0.98 var(--font-serif);
  letter-spacing: -0.035em;
  color: var(--text);

  .dotp { display: inline-block; color: var(--primary); }

  /* 一次性扫光：中性明度带（不带主色），扫过后即为纯文字色 */
  .lit {
    background: linear-gradient(100deg, var(--text) 0 40%, var(--text-3) 50%, var(--text) 60% 100%) 100% 0 / 250% 100% no-repeat;
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
    /* background-clip:text 只在盒内着色：给 f 的顶钩、y 的下伸留出余量，负外边距抵消，排版不动 */
    padding: 0 0.14em 0.12em 0;
    margin-right: -0.14em;
    -webkit-box-decoration-break: clone;
    animation: pf-sheen 2.6s 0.5s var(--ease-out) both;
  }
}

@keyframes pf-sheen { from { background-position: 100% 0; } to { background-position: 0 0; } }

.pf-alias {
  margin: 10px 0 0;
  font: 500 15px/1.4 var(--font-serif);
  letter-spacing: 0.06em;
  color: var(--text-3);
}

.pf-lede {
  margin: 34px 0 10px;
  font: 600 clamp(18px, 2.9cqi, 24px) / 1.5 var(--font-serif);
  letter-spacing: 0.02em;

  em { font-style: normal; color: var(--ak-ink); }
}

.pf-bio { max-width: 660px; font-size: 15px; line-height: 1.8; color: var(--text-2); }

/* 状态行 = 统计条写法：下沉面等分格，标签 12.5px 三级灰，值 17px（时间 20px 等宽） */
.pf-status {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, max-content));
  gap: var(--stat-gap);
  margin: 22px 0 20px;
  padding: var(--statbar-pad);
  border-radius: var(--statbar-r);
  background: var(--statbar-bg);
  box-shadow: var(--statbar-shadow);

  > div { min-width: 0; padding: 14px 22px 12px 18px; }
  > div + div { box-shadow: -1px 0 0 var(--ak-line); }

  /* 简洁风格：纯数字行 */
  :root[data-style='clean'] & {
    > div { padding: 0 20px 0 0; }
    > div + div { box-shadow: none; }
  }

  dt { margin-bottom: 6px; font-size: 12.5px; line-height: 1.3; color: var(--ak-text-3); }
  dd { display: flex; align-items: center; gap: 8px; min-height: 28px; font-size: 17px; font-weight: 600; white-space: nowrap; }

  .clock { font: 600 20px var(--ak-mono); font-variant-numeric: tabular-nums; }
  .clock s { text-decoration: none; opacity: 0.45; }
  .dayp { font-size: 13px; font-weight: 400; color: var(--ak-text-3); }
}

.pf-social {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;

  a {
    height: 40px;
    padding: 0 18px 0 15px;

    .ak-mono { font-size: 12px; font-weight: 400; color: var(--ak-text-3); }

    &.pri .ak-mono { color: color-mix(in oklab, var(--on-solid) 68%, transparent); }

    /* 非主按钮的品牌图标用品牌色（黑色系品牌由 brandColor 回落为正文色） */
    &:not(.pri) .ak-brand { color: var(--bc, currentColor); }
  }
}

/* ---------- 形象图：按原图比例完整显示（不裁切），左缘部分渐隐融入背景 ---------- */
.pf-portrait {
  position: relative;
  display: flex;
  justify-content: flex-end;
  min-width: 0;
  margin: 0;
  animation: pf-portrait-in 1.2s 0.15s var(--ease-out) both;
}

.pf-frame {
  position: relative;
  display: block;
  max-width: 100%;
  overflow: hidden;
  border-radius: var(--pr, var(--r-xl));

  img {
    display: block;
    width: auto;
    max-width: 100%;
    height: auto;
    max-height: 540px;
    transition: transform 1.2s var(--ease-out);
  }
}

.pf-portrait:hover .pf-frame img { transform: scale(1.03); }

/* 渐隐：仅左缘，缓动曲线式多段停靠，过渡区约 45%，无硬边；底部保持清晰 */
.fade-left .pf-frame {
  -webkit-mask-image: linear-gradient(to right, transparent 0%, rgb(0 0 0 / 0.04) calc(var(--fw, 50%) * 0.115), rgb(0 0 0 / 0.14) calc(var(--fw, 50%) * 0.25), rgb(0 0 0 / 0.3) calc(var(--fw, 50%) * 0.385), rgb(0 0 0 / 0.5) calc(var(--fw, 50%) * 0.52), rgb(0 0 0 / 0.7) calc(var(--fw, 50%) * 0.654), rgb(0 0 0 / 0.86) calc(var(--fw, 50%) * 0.77), rgb(0 0 0 / 0.96) calc(var(--fw, 50%) * 0.885), #000 var(--fw, 50%));
  mask-image: linear-gradient(to right, transparent 0%, rgb(0 0 0 / 0.04) calc(var(--fw, 50%) * 0.115), rgb(0 0 0 / 0.14) calc(var(--fw, 50%) * 0.25), rgb(0 0 0 / 0.3) calc(var(--fw, 50%) * 0.385), rgb(0 0 0 / 0.5) calc(var(--fw, 50%) * 0.52), rgb(0 0 0 / 0.7) calc(var(--fw, 50%) * 0.654), rgb(0 0 0 / 0.86) calc(var(--fw, 50%) * 0.77), rgb(0 0 0 / 0.96) calc(var(--fw, 50%) * 0.885), #000 var(--fw, 50%));
}

.fade-bottom .pf-frame {
  -webkit-mask-image: linear-gradient(to top, transparent 0%, rgb(0 0 0 / 0.04) calc(var(--fw, 50%) * 0.115), rgb(0 0 0 / 0.14) calc(var(--fw, 50%) * 0.25), rgb(0 0 0 / 0.3) calc(var(--fw, 50%) * 0.385), rgb(0 0 0 / 0.5) calc(var(--fw, 50%) * 0.52), rgb(0 0 0 / 0.7) calc(var(--fw, 50%) * 0.654), rgb(0 0 0 / 0.86) calc(var(--fw, 50%) * 0.77), rgb(0 0 0 / 0.96) calc(var(--fw, 50%) * 0.885), #000 var(--fw, 50%));
  mask-image: linear-gradient(to top, transparent 0%, rgb(0 0 0 / 0.04) calc(var(--fw, 50%) * 0.115), rgb(0 0 0 / 0.14) calc(var(--fw, 50%) * 0.25), rgb(0 0 0 / 0.3) calc(var(--fw, 50%) * 0.385), rgb(0 0 0 / 0.5) calc(var(--fw, 50%) * 0.52), rgb(0 0 0 / 0.7) calc(var(--fw, 50%) * 0.654), rgb(0 0 0 / 0.86) calc(var(--fw, 50%) * 0.77), rgb(0 0 0 / 0.96) calc(var(--fw, 50%) * 0.885), #000 var(--fw, 50%));
}

/* 未配置形象图也未上传站点 logo：内置 logo 原图自带圆角底，直接全尺寸显示，不再套底框 */
.pf-portrait.logo .pf-frame {
  width: min(100%, 520px);
  border-radius: 0;

  img { width: 100%; max-height: none; }
}

.pf-portrait.logo:hover .pf-frame img { transform: scale(1.02); }

@keyframes pf-portrait-in {
  from { opacity: 0; transform: translateX(28px) scale(0.985); filter: blur(12px); }
  to { opacity: 1; transform: none; filter: none; }
}

@container (max-width: 820px) {
  .pf { grid-template-columns: 1fr; padding: 20px 0 4px; gap: 20px; }

  .pf-portrait {
    order: -1;
    justify-content: flex-start;
    width: 120px;
    height: 120px;

    .pf-frame {
      width: 100%;
      border-radius: min(var(--pr, var(--r-xl)), 30px);
      -webkit-mask-image: none;
      mask-image: none;
      box-shadow: 0 16px 36px -18px rgb(0 0 0 / 0.6);

      img { width: 100%; height: 100%; max-height: none; object-fit: cover; object-position: var(--pf, 50% 40%); }
    }
  }

  .pf-portrait.logo .pf-frame { border-radius: 26px; }

  .pf-hi { font-size: 16px; }
}

@container (max-width: 560px) {
  .pf-status { grid-template-columns: 1fr; }
  .pf-status > div + div { box-shadow: 0 -1px 0 var(--ak-line); }
}
</style>
