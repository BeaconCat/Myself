<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useConfigStore } from '../../stores/config';
import AboutModules from '../../about/AboutModules.vue';
import KitIcon from '../../about/parts/KitIcon.vue';
import { brandColor, cardLinks } from '../../about/brands';
import { dayPartKey, pad2, useClock, zoned } from '../../about/useClock';
import { safeHref } from '../../utils/safeUrl';
import LargeTitlePage from '../../components/mobile/LargeTitlePage.vue';
import MIcon from '../../components/mobile/MIcon.vue';
import { copyText, toast } from '../../components/mobile/shell';

/**
 * 移动端关于：
 * - 身份名片：可选头图（身份页开关，默认关闭）；头像居左、名字（别名）/ 账号在其右，右侧为名片按钮（身份页勾选，
 *   与桌面身份区同一组）；下接签名（*高亮*）/ 自述 / 状态条（正在做 · 城市 · 本地时间）
 * - 模块流（复用 AboutModules，仅在容器上做窄屏适配；身份与格言不在流里重复）
 * - 格言收尾（与桌面一致放在「想聊聊」之后），最后是「本站基于 Myself」与 RSS 入口
 * 配置未就绪时显示同几何骨架。
 */
const { t } = useI18n();
const config = useConfigStore();
const about = computed(() => config.cfg.about);
const avatar = computed(() => about.value.avatar || '/favicon-256.png');
const handle = computed(() => config.cfg.github.username || 'myself');
const banner = computed(() => (about.value.banner?.show ? about.value.banner.src || '/covers/05.webp' : ''));
const bannerFocus = computed(() => about.value.banner?.focus || '50% 50%');
const links = computed(() => cardLinks(about.value.links));
const external = (url: string) => (/^https?:/.test(url) ? '_blank' : undefined);

/** *星号* 包裹的片段高亮 */
const lede = computed(() =>
  (about.value.tagline ?? '').split('*').map((text, i) => ({ text, em: i % 2 === 1 })).filter((x) => x.text),
);

const now = useClock();
const tz = computed(() => Number(about.value.status?.tz ?? 8));
const clock = computed(() => zoned(now.value, tz.value));

/** 格言模块单独放到页尾（「想聊聊」之后）；其余模块照常成流 */
const flow = computed(() => (about.value.modules ?? []).filter((m) => m.type !== 'motto'));
const closing = computed(() => (about.value.modules ?? []).filter((m) => m.type === 'motto' && !m.hidden));

async function share(): Promise<void> {
  const ok = await copyText(window.location.href);
  toast(ok ? t('mobile.linkCopied') : t('mobile.copyFailed'));
}

async function copyRss(): Promise<void> {
  const ok = await copyText(`${window.location.origin}/feed`);
  toast(ok ? t('mobile.rssCopied') : t('mobile.copyFailed'), ok ? '/feed' : '');
}

async function refresh(): Promise<void> {
  await config.load();
}
</script>

<template>
  <LargeTitlePage :title="t('mobile.about.title')" :large="false" :refresh="refresh">
    <template #right>
      <button class="m-icbtn m-tap" :aria-label="t('mobile.share')" @click="share"><MIcon name="share" /></button>
    </template>

    <template #hero>
      <Transition name="m-swap" mode="out-in">
        <section v-if="!config.loaded" key="sk" class="ab-hero sk">
          <div class="m-sk ab-banner" />
          <span class="m-sk ab-av" />
          <span class="m-sk m-sk-line" style="width: 46%; height: 26px; margin-top: 14px" />
          <span class="m-sk m-sk-line" style="width: 30%; margin-top: 10px" />
          <span class="m-sk m-sk-line" style="width: 92%; margin-top: 18px" />
          <span class="m-sk m-sk-line" style="width: 80%; margin-top: 8px" />
          <div class="m-sk sk-mod" />
        </section>

        <div v-else key="ok">
          <section class="ab-hero">
            <div class="ab-card m-in" :class="{ 'has-banner': !!banner }">
              <div v-if="banner" class="ab-banner">
                <img :src="banner" :style="{ objectPosition: bannerFocus }" alt="" draggable="false" />
              </div>
              <div class="ab-head">
                <div class="ab-av">
                  <img :src="avatar" alt="" draggable="false" />
                  <i v-if="about.status?.doing" class="ab-live" />
                </div>
                <div class="ab-id">
                  <h1 class="ab-name">{{ about.name }}</h1>
                  <div class="ab-handle">
                    <span v-if="about.alias?.trim()" class="alias">{{ about.alias.trim() }}</span>@{{ handle }}
                  </div>
                </div>
                <nav v-if="links.length" class="ab-links">
                  <a
                    v-for="l in links"
                    :key="l.name + l.url"
                    :href="safeHref(l.url)"
                    :target="external(l.url)"
                    rel="noopener noreferrer"
                    :aria-label="l.name"
                    :title="l.name"
                    :style="{ '--c': brandColor(l.icon) }"
                    class="m-tap"
                  ><KitIcon :name="l.icon" :size="18" /></a>
                </nav>
              </div>

              <div class="ab-body">
                <p v-if="lede.length" class="ab-lede">
                  <template v-for="(x, i) in lede" :key="i"><em v-if="x.em">{{ x.text }}</em><template v-else>{{ x.text }}</template></template>
                </p>
                <p v-if="about.bio" class="ab-bio">{{ about.bio }}</p>
              </div>

              <dl class="ab-status">
                <div v-if="about.status?.doing">
                  <dt>{{ t('aboutKit.doing') }}</dt>
                  <dd><i class="live-dot" />{{ about.status.doing }}</dd>
                </div>
                <div v-if="about.status?.city">
                  <dt>{{ t('aboutKit.city') }}</dt>
                  <dd>{{ about.status.city }}</dd>
                </div>
                <div>
                  <dt>{{ t('aboutKit.localTime') }}</dt>
                  <dd><span class="mono">{{ pad2(clock.h) }}:{{ pad2(clock.m) }}</span><small>{{ t(dayPartKey(clock.h)) }}</small></dd>
                </div>
              </dl>
            </div>
          </section>

          <div class="mods m-in" style="--i: 3">
            <AboutModules :modules="flow" :about="about" />
          </div>

          <div v-if="closing.length" class="mods closing">
            <AboutModules :modules="closing" :about="about" />
          </div>

          <div class="m-list ab-list m-in" style="--i: 4">
            <a class="m-li" href="https://github.com/BeaconCat/Myself" target="_blank" rel="noopener">
              <span class="lic" style="--c: #24292f"><MIcon name="github" /></span>
              <span class="li-t"><b>{{ t('mobile.about.source') }}</b><small>{{ t('mobile.about.sourceSub') }}</small></span>
              <MIcon name="chev" class="chev" />
            </a>
            <button class="m-li" @click="copyRss">
              <span class="lic" style="--c: #ff7a1a"><MIcon name="rss" /></span>
              <span class="li-t"><b>{{ t('mobile.rss') }}</b><small>{{ t('mobile.about.rssSub') }}</small></span>
              <MIcon name="chev" class="chev" />
            </button>
          </div>
        </div>
      </Transition>
    </template>
  </LargeTitlePage>
</template>

<style scoped lang="scss">
.ab-hero {
  position: relative;
  padding: 8px 16px 4px;

  &.sk {
    .m-sk-line { display: block; }
    .ab-av { display: block; margin: -40px 0 0 16px; }
  }
}

/* 身份名片：可选头图；头像居左、名字 / 账号在右、名片按钮靠右 */
.ab-card {
  position: relative;
  border-radius: var(--r-xl);
  background: var(--elev);
  box-shadow: inset 0 0 0 0.5px var(--line-2), 0 18px 40px -30px rgb(0 0 0 / 0.45);
  overflow: hidden;
}

.ab-banner {
  position: relative;
  height: 120px;
  overflow: hidden;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    animation: ab-banner-in 1.1s var(--ease-out) both;
  }

  /* 底缘轻压暗，头像压边时更稳 */
  &::after {
    content: '';
    position: absolute;
    inset: 45% 0 0;
    background: linear-gradient(transparent, rgb(0 0 0 / 0.14));
  }
}

.sk .ab-banner { border-radius: var(--r-xl); }

@keyframes ab-banner-in {
  from { transform: scale(1.08); filter: blur(6px); }
}

/* 头部行：无头图时三者垂直居中；有头图时头像上移压边，名字与按钮落在头图下沿之下 */
.ab-head {
  position: relative;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px 16px 0;
}

.has-banner .ab-head {
  align-items: flex-start;
  margin-top: -36px;
  padding-top: 0;

  .ab-id { padding-top: 44px; }
  .ab-links { padding-top: 46px; }
}

.ab-av {
  position: relative;
  flex: none;
  width: 72px;
  height: 72px;
  border-radius: var(--r-xl);
  background: var(--elev);
  box-shadow: 0 0 0 1px var(--line-2), 0 10px 24px -14px rgb(0 0 0 / 0.5);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    border-radius: inherit;
  }
}

.has-banner .ab-av { box-shadow: 0 0 0 4px var(--elev), 0 12px 26px -14px rgb(0 0 0 / 0.55); }

.ab-live {
  position: absolute;
  right: -4px;
  bottom: -4px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #22c55e;
  box-shadow: 0 0 0 3px var(--elev);
}

.ab-id { flex: 1; min-width: 0; }

.ab-name {
  font-family: var(--font-serif);
  font-size: 26px;
  line-height: 1.2;
  font-weight: 700;
  letter-spacing: 0.01em;
  overflow-wrap: anywhere;
}

.ab-handle {
  margin-top: 4px;
  font: 500 12.5px var(--font-mono);
  color: var(--text-3);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;

  .alias {
    margin-right: 8px;
    font-family: var(--font-sans);
    font-size: 13px;
    color: var(--text-2);
  }
}

/* 名片按钮：品牌色轻染底 + 品牌色图标（黑色系品牌与通用图标用正文色） */
.ab-links {
  display: flex;
  flex: none;
  gap: 6px;

  a {
    --c: var(--text-2);

    display: grid;
    place-items: center;
    width: 34px;
    height: 34px;
    border-radius: 50%;
    color: var(--c);
    background: color-mix(in oklab, var(--c) 11%, transparent);
    box-shadow: inset 0 0 0 0.5px color-mix(in oklab, var(--c) 22%, transparent);
    transition: transform var(--dur-fast) var(--ease-spring), background var(--dur-fast);

    &:active { transform: scale(0.9); background: color-mix(in oklab, var(--c) 18%, transparent); }
  }

  /* KitIcon 的线性描边样式挂在 .ak 作用域下，名片在模块流之外，这里补上（品牌实心图标由 KitIcon 自带样式覆盖） */
  :deep(svg.ak-i) {
    fill: none;
    stroke: currentColor;
    stroke-width: 1.6;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.ab-body { padding: 0 16px; }

.ab-lede {
  margin-top: 14px;
  font-family: var(--font-serif);
  font-size: 19px;
  line-height: 1.55;
  font-weight: 600;

  em { font-style: normal; color: var(--ink); }
}

.ab-bio {
  margin-top: 8px;
  font-size: 15px;
  line-height: 1.75;
  color: var(--text-2);
}

/* 状态条：等分格 + 细分隔线 */
.ab-status {
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: minmax(0, 1fr);
  margin: 16px 16px 0;
  padding: 12px 0 14px;
  border-top: 0.5px solid var(--line-2);

  > div { min-width: 0; padding: 0 12px; }
  > div:first-child { padding-left: 0; }
  > div + div { box-shadow: -0.5px 0 0 var(--line-2); }

  dt { font-size: 11.5px; color: var(--text-3); letter-spacing: 0.04em; }

  dd {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 3px;
    font-size: 14px;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;

    small { font-size: 12px; font-weight: 400; color: var(--text-3); }
  }

  .mono { font-family: var(--font-mono); }

  .live-dot {
    flex: none;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #22c55e;
    box-shadow: 0 0 0 3px color-mix(in oklab, #22c55e 22%, transparent);
  }
}

/* 模块流窄屏适配：只调容器与外壳间距，不改模块本身 */
.mods {
  padding: 16px 16px 0;
  overflow-x: clip;

  :deep(.ak) {
    --ak-pad: 20px;
    --ak-gap: 12px;
    --ak-r-lg: var(--r-xl);
  }

  /* 列向 flex 中 flex-basis:0 在不定高卡片里会把模块体压扁，窄屏回到内容高度 */
  :deep(.ak-m.card > .ak-body) { flex: 1 0 auto; }
  :deep(.ak-m.card:hover) { transform: none; }
  /* 身份由上方名片承担，模块流里不重复 */
  :deep(.ak-m.m-profile) { display: none; }
}

/* 分发器迁移旧配置时可能补回格言模块：流里一律不显示，格言只在页尾出现 */
.mods:not(.closing) :deep(.ak-m.m-motto) { display: none; }

/* 页尾格言：只留格言本身（分发器缺身份区时会自动补一个，这里隐去） */
.closing {
  padding-top: 0;

  :deep(.ak-m:not(.m-motto)) { display: none; }
  /* 落款线在底部上方 36px：底部留足 68px，线与文字间隔约 32px */
  :deep(.mo) { padding: 32px 0 68px; }
}

.ab-list {
  margin: 8px 16px 0;

  .m-li { height: 62px; }

  a.m-li {
    color: inherit;
    text-decoration: none;
  }

  .li-t {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 1px;
    text-align: left;

    b { font-weight: 500; }

    small {
      margin: 0;
      font-size: 12px;
      color: var(--text-3);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
}

.sk-mod {
  margin-top: 20px;
  height: 180px;
  border-radius: var(--r-xl);
}
</style>
