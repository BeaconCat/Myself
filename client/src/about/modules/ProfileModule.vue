<script setup lang="ts">
import { computed } from 'vue';
import { useConfigStore } from '../../stores/config';
import type { ProfileData } from '../types';
import type { ModProps } from './props';
import KitIcon from '../parts/KitIcon.vue';
import { useI18n } from 'vue-i18n';
import { dayPartKey, pad2, useClock, zoned } from '../useClock';

/**
 * 身份区（profile）：编辑式大字宋体名字 + 一次性扫光、一句话、自述、
 * 实时状态行（正在做 / 城市 / 本地时间逐秒走）与社交入口；
 * 右侧为圆角形象图，按 portrait.fade 向左（或向下）渐隐融入背景，未配置时用站点 logo。
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
    src: p.src || '/favicon-256.png',
    logo: !p.src,
    fade: p.fade || 'left',
    style: { '--pr': r != null && Number.isFinite(r) ? `${r}px` : undefined, '--pf': p.focus || '50% 40%' },
  };
});

const plain = computed(() => props.variant === 'plain');
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

      <nav v-if="d.links?.length" class="pf-social">
        <a
          v-for="l in d.links"
          :key="l.name + l.url"
          :href="l.url"
          class="ak-btn"
          :class="{ pri: l.primary }"
          :target="external(l.url)"
          rel="noopener"
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
.pf {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 32px;
  align-items: center;
  min-height: 600px;
  padding: 96px 0 40px;

  &.plain { grid-template-columns: minmax(0, 1fr); }
}

.pf-text { position: relative; z-index: 1; min-width: 0; container-type: inline-size; }

.pf-kicker {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 26px;
  font: 500 12px/1 var(--ak-mono);
  letter-spacing: 0.14em;
  color: var(--ak-text-3);
  text-transform: uppercase;

  i { width: 44px; height: 1px; background: var(--line-2); }
}

.pf-hi {
  display: block;
  margin-bottom: 14px;
  font: 500 22px/1.2 var(--font-serif);
  letter-spacing: 0.04em;
  color: var(--text-2);
}

.pf-name {
  position: relative;
  white-space: nowrap;
  font: 700 clamp(40px, var(--fit, 17cqi), 148px) / 0.95 var(--font-serif);
  letter-spacing: -0.035em;
  color: var(--text);

  .dotp { display: inline-block; color: var(--primary); }

  /* 一次性扫光：中性明度带（不带主色），扫过后即为纯文字色 */
  .lit {
    background: linear-gradient(100deg, var(--text) 0 40%, var(--text-3) 50%, var(--text) 60% 100%) 100% 0 / 250% 100% no-repeat;
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
    animation: pf-sheen 2.6s 0.5s var(--ease-out) both;
  }
}

@keyframes pf-sheen { from { background-position: 100% 0; } to { background-position: 0 0; } }

.pf-lede {
  margin: 30px 0 16px;
  font: 500 clamp(18px, 3.1cqi, 26px) / 1.55 var(--font-serif);
  letter-spacing: 0.02em;

  em { font-style: normal; color: var(--ak-ink); }
}

.pf-bio { max-width: 560px; font-size: 15.5px; line-height: 1.9; color: var(--text-2); }

.pf-status {
  display: flex;
  flex-wrap: wrap;
  margin: 34px 0 30px;
  border-top: 1px solid var(--ak-line);
  border-bottom: 1px solid var(--ak-line);

  > div { padding: 14px 28px 14px 0; margin-right: 28px; border-right: 1px solid var(--ak-line); }
  > div:last-child { border-right: 0; margin-right: 0; }

  dt { margin-bottom: 9px; font: 500 11px/1 var(--ak-mono); letter-spacing: 0.1em; color: var(--ak-text-3); }
  dd { display: flex; align-items: center; gap: 8px; font-size: 14px; white-space: nowrap; }

  .clock { font: 500 15px var(--ak-mono); font-variant-numeric: tabular-nums; }
  .clock s { text-decoration: none; opacity: 0.45; }
  .dayp { font-size: 12px; color: var(--ak-text-3); }
}

.pf-social {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;

  a {
    height: 40px;
    padding: 0 18px 0 15px;

    .ak-mono { font-size: 11.5px; font-weight: 400; color: var(--ak-text-3); }

    &.pri .ak-mono { color: color-mix(in oklab, var(--on-solid) 68%, transparent); }
  }
}

/* ---------- 形象图：圆角 + 渐隐融入背景 ---------- */
.pf-portrait {
  position: relative;
  width: clamp(300px, 36cqi, 440px);
  aspect-ratio: 440 / 540;
  margin: 0;
  animation: pf-portrait-in 1.2s 0.15s var(--ease-out) both;
}

.pf-frame {
  position: absolute;
  inset: 0;
  display: block;
  overflow: hidden;
  border-radius: var(--pr, var(--r-xl));
  box-shadow: 0 40px 80px -40px rgb(0 0 0 / 0.55);

  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: var(--pf, 50% 40%);
    transition: transform 1.2s var(--ease-out);
  }
}

.pf-portrait:hover .pf-frame img { transform: scale(1.03); }

.fade-left .pf-frame {
  -webkit-mask-image: linear-gradient(to right, transparent, rgb(0 0 0 / 0.08) 8%, rgb(0 0 0 / 0.3) 17%, rgb(0 0 0 / 0.62) 27%, rgb(0 0 0 / 0.88) 35%, #000 42%);
  mask-image: linear-gradient(to right, transparent, rgb(0 0 0 / 0.08) 8%, rgb(0 0 0 / 0.3) 17%, rgb(0 0 0 / 0.62) 27%, rgb(0 0 0 / 0.88) 35%, #000 42%);
}

.fade-bottom .pf-frame {
  -webkit-mask-image: linear-gradient(to top, transparent, rgb(0 0 0 / 0.08) 8%, rgb(0 0 0 / 0.3) 17%, rgb(0 0 0 / 0.62) 27%, rgb(0 0 0 / 0.88) 35%, #000 42%);
  mask-image: linear-gradient(to top, transparent, rgb(0 0 0 / 0.08) 8%, rgb(0 0 0 / 0.3) 17%, rgb(0 0 0 / 0.62) 27%, rgb(0 0 0 / 0.88) 35%, #000 42%);
}

/* 未配置形象图：品牌 logo 置于同色深底上。品牌光影例外：底部一团主色门缝光属于品牌构图 */
.pf-portrait.logo .pf-frame {
  background:
    radial-gradient(60% 50% at 58% 62%, rgba(var(--primary-rgb), 0.28), transparent 70%),
    radial-gradient(120% 90% at 60% 40%, #0b1528, #050b17 70%);

  img { position: absolute; inset: 0 0 0 18%; margin: auto; width: 62%; height: auto; aspect-ratio: 1; object-fit: contain; }
}

.pf-portrait.logo:hover .pf-frame img { transform: scale(1.04); }

@keyframes pf-portrait-in {
  from { opacity: 0; transform: translateX(28px) scale(0.985); filter: blur(12px); }
  to { opacity: 1; transform: none; filter: none; }
}

@container (max-width: 820px) {
  .pf { grid-template-columns: 1fr; min-height: 0; padding: 24px 0 8px; gap: 24px; }

  .pf-portrait {
    order: -1;
    width: 120px;
    height: 120px;

    .pf-frame { border-radius: min(var(--pr, var(--r-xl)), 30px); -webkit-mask-image: none; mask-image: none; box-shadow: 0 16px 36px -18px rgb(0 0 0 / 0.6); }
  }

  .pf-portrait.logo .pf-frame img { inset: 0; width: 86%; }

  .pf-hi { font-size: 16px; }
}

@container (max-width: 560px) {
  .pf-status { flex-direction: column; }
  .pf-status > div { border-right: 0; margin: 0; padding: 12px 0; border-bottom: 1px solid var(--ak-line); }
  .pf-status > div:last-child { border-bottom: 0; }
}
</style>
