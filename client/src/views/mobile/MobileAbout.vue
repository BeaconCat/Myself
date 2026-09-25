<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useConfigStore } from '../../stores/config';
import AboutModules from '../../about/AboutModules.vue';
import LargeTitlePage from '../../components/mobile/LargeTitlePage.vue';
import MIcon from '../../components/mobile/MIcon.vue';
import { copyText, toast } from '../../components/mobile/shell';

/**
 * 移动端关于：身份区（头像 / 名字 / 格言大字，后半句主色）+ 模块流（复用 AboutModules，仅在容器上做窄屏适配）
 * + 源代码 / RSS 入口。配置未就绪时显示同几何骨架。
 */
const { t } = useI18n();
const config = useConfigStore();
const about = computed(() => config.cfg.about);
const avatar = computed(() => about.value.avatar || '/favicon-256.png');
const handle = computed(() => config.cfg.github.username || 'myself');

/** 格言按第一个中文逗号拆成两行，后半句用主色 */
const motto = computed(() => {
  const m = about.value.motto ?? '';
  const i = m.search(/[，,]/);
  return i > 0 ? [m.slice(0, i + 1), m.slice(i + 1)] : [m, ''];
});

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
          <div class="ab-id">
            <span class="m-sk ab-av" />
            <div class="ab-who">
              <span class="m-sk m-sk-line" style="width: 60%; height: 26px" />
              <span class="m-sk m-sk-line" style="width: 80%; margin-top: 10px" />
            </div>
          </div>
          <span class="m-sk m-sk-line" style="width: 70%; height: 24px; margin-top: 20px" />
          <span class="m-sk m-sk-line" style="width: 50%; height: 24px; margin-top: 10px" />
          <span class="m-sk m-sk-line" style="width: 92%; margin-top: 18px" />
          <span class="m-sk m-sk-line" style="width: 80%; margin-top: 8px" />
          <div class="m-sk sk-mod" />
        </section>

        <div v-else key="ok">
          <section class="ab-hero">
            <div class="ab-id m-in">
              <div class="ab-av"><img :src="avatar" alt="" draggable="false" /></div>
              <div class="ab-who">
                <h1 class="ab-name">{{ about.name }}</h1>
                <div class="ab-handle">@{{ handle }}<template v-if="about.tagline"> · {{ about.tagline }}</template></div>
              </div>
            </div>
            <p v-if="motto[0]" class="ab-quote m-in" style="--i: 3">
              {{ motto[0] }}<br v-if="motto[1]" /><span>{{ motto[1] }}</span>
            </p>
            <p v-if="about.bio" class="ab-bio m-in" style="--i: 4">{{ about.bio }}</p>
          </section>

          <div class="mods m-in" style="--i: 5">
            <AboutModules :modules="about.modules ?? []" />
          </div>

          <div class="m-list ab-list m-in" style="--i: 6">
            <a class="m-li" href="https://github.com/BeaconCat/Myself" target="_blank" rel="noopener">
              <span class="lic" style="--c: #24292f"><MIcon name="github" /></span>
              <span>{{ t('mobile.about.source') }}</span><small>BeaconCat/Myself</small><MIcon name="chev" class="chev" />
            </a>
            <button class="m-li" @click="copyRss">
              <span class="lic" style="--c: #ff7a1a"><MIcon name="rss" /></span>
              <span>{{ t('mobile.rss') }}</span><small>/feed</small><MIcon name="chev" class="chev" />
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

  &.sk .m-sk-line { display: block; }
}

/* 身份行（高密度）：头像 + 名字 / 账号同一行，不再纵向各占一块 */
.ab-id {
  display: flex;
  align-items: center;
  gap: 16px;
}

.ab-who { flex: 1; min-width: 0; }

.ab-av {
  position: relative;
  flex: none;
  width: 72px;
  height: 72px;
  border-radius: var(--r-lg);
  overflow: hidden;
  box-shadow: var(--shadow-card);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
}

.ab-name {
  font-family: var(--font-serif);
  font-size: 28px;
  line-height: 1.2;
  font-weight: 700;
  letter-spacing: 0.01em;
}

.ab-handle {
  margin-top: 4px;
  font-size: 13px;
  color: var(--text-3);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ab-quote {
  position: relative;
  margin-top: 20px;
  font-family: var(--font-serif);
  font-size: 24px;
  line-height: 1.5;
  font-weight: 700;
  letter-spacing: 0.02em;

  span { color: var(--ink); }
}

.ab-bio {
  position: relative;
  margin-top: 10px;
  font-size: 15px;
  line-height: 1.75;
  color: var(--text-2);
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
  /* 身份由上方移动端原生身份区承担（移动原型设计），模块流里的桌面身份模块不重复显示 */
  :deep(.ak-m.m-profile) { display: none; }
}

.ab-list {
  margin: 16px 16px 0;

  a.m-li {
    color: inherit;
    text-decoration: none;
  }
}

.sk-mod {
  margin-top: 20px;
  height: 180px;
  border-radius: var(--r-xl);
}
</style>
