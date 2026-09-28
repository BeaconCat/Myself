<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { api, type Tag } from '../../api';
import { useConfigStore } from '../../stores/config';
import { useThemeStore, type Mode, type UiStyle } from '../../stores/theme';
import { circularReveal } from '../../utils/circularReveal';
import MIcon from './MIcon.vue';
import { closeDrawer, copyText, shell, toast } from './shell';

/**
 * 侧边抽屉：作者卡 / 外观（深浅 + 色盘）/ 标签云 / 更多（RSS、GitHub、关于站点）/ 管理后台。
 * 位置与跟手由外壳驱动（--m-dp），本组件只负责内容。
 */
const { t } = useI18n();
const router = useRouter();
const config = useConfigStore();
const theme = useThemeStore();

const about = computed(() => config.cfg.about);
const avatar = computed(() => about.value.avatar || '/favicon-256.png');
const handle = computed(() => config.cfg.github.username || 'myself');
const days = computed(() => {
  const start = new Date(`${about.value.foundedAt || '2026-01-01'}T00:00:00`).getTime();
  return Math.max(1, Math.floor((Date.now() - start) / 864e5));
});

const tags = ref<Tag[]>([]);
onMounted(async () => {
  try {
    tags.value = await api.tags();
  } catch { /* 标签云为空即可 */ }
});

function origin(e: MouseEvent): { x: number; y: number } {
  return { x: e.clientX, y: e.clientY };
}

function setMode(mode: Mode, e: MouseEvent): void {
  if (theme.mode === mode) return;
  circularReveal(origin(e), () => theme.setMode(mode), mode === 'light' ? 'expand' : 'contract');
}

function setStyle(style: UiStyle, e: MouseEvent): void {
  if (theme.style === style) return;
  circularReveal(origin(e), () => theme.setStyle(style), style === 'clean' ? 'expand' : 'contract');
}

function setPalette(id: string, e: MouseEvent): void {
  if (theme.paletteId === id) return;
  circularReveal(origin(e), () => theme.setPalette(id), 'expand');
}

function go(path: string | { path: string; query?: Record<string, string> }): void {
  closeDrawer();
  window.setTimeout(() => void router.push(path), 180);
}

async function copyRss(): Promise<void> {
  const ok = await copyText(`${window.location.origin}/feed`);
  toast(ok ? t('mobile.rssCopied') : t('mobile.copyFailed'), ok ? '/feed' : '');
}

function openGithub(): void {
  window.open(`https://github.com/${handle.value}`, '_blank', 'noopener');
}
</script>

<template>
  <aside class="drawer" :aria-hidden="shell.dp < 0.5">
    <div class="author">
      <div class="who">
        <span class="av"><img class="m-avatar" :src="avatar" alt="" draggable="false" /></span>
        <div>
          <b>{{ about.name }}</b>
          <small>@{{ handle }} · {{ t('mobile.online') }}</small>
        </div>
      </div>
      <p v-if="about.motto" class="motto">{{ about.motto }}</p>
    </div>

    <div class="h">{{ t('mobile.appearance') }}</div>
    <div class="seg2" :style="{ '--i': theme.mode === 'dark' ? 1 : 0 }">
      <span class="th" aria-hidden="true" />
      <button class="m-tap" :class="{ on: theme.mode === 'light' }" @click="setMode('light', $event)">
        <MIcon name="sun" class="s" />{{ t('mobile.light') }}
      </button>
      <button class="m-tap" :class="{ on: theme.mode === 'dark' }" @click="setMode('dark', $event)">
        <MIcon name="moon" class="s" />{{ t('mobile.dark') }}
      </button>
    </div>
    <div
      v-if="theme.allowUserStyle"
      class="seg2 style"
      role="group"
      :aria-label="t('theme.style')"
      :style="{ '--i': theme.style === 'cards' ? 1 : 0 }"
    >
      <span class="th" aria-hidden="true" />
      <button class="m-tap" :class="{ on: theme.style === 'clean' }" :aria-pressed="theme.style === 'clean'" @click="setStyle('clean', $event)">
        <MIcon name="clean" class="s" />{{ t('theme.styleClean') }}
      </button>
      <button class="m-tap" :class="{ on: theme.style === 'cards' }" :aria-pressed="theme.style === 'cards'" @click="setStyle('cards', $event)">
        <MIcon name="cards" class="s" />{{ t('theme.styleCards') }}
      </button>
    </div>
    <div v-if="theme.allowUserPalette" class="pals">
      <button
        v-for="p in theme.allPalettes"
        :key="p.id"
        class="pal m-tap"
        :class="{ on: p.id === theme.paletteId }"
        :aria-label="p.nameKey"
        @click="setPalette(p.id, $event)"
      >
        <i :style="{ '--c': p.light.primary }" />
        <span>{{ p.nameKey.split('·')[0].trim() }}</span>
      </button>
    </div>

    <template v-if="tags.length">
      <div class="h">{{ t('mobile.tags') }}</div>
      <div class="tags">
        <button
          v-for="tag in tags"
          :key="tag.name"
          class="m-tap"
          :class="{ big: tag.count > 2 }"
          @click="go({ path: '/articles', query: { tag: tag.name } })"
        >
          <span class="hs">#</span>{{ tag.name }}<em>{{ tag.count }}</em>
        </button>
      </div>
    </template>

    <div class="h">{{ t('mobile.more') }}</div>
    <div class="m-list">
      <button class="m-li" @click="copyRss">
        <span class="lic" style="--c: #ff7a1a"><MIcon name="rss" /></span>
        <span>{{ t('mobile.rss') }}</span><small>{{ t('mobile.copy') }}</small>
      </button>
      <button class="m-li" @click="openGithub">
        <span class="lic" style="--c: #24292f"><MIcon name="github" /></span>
        <span>{{ t('mobile.github') }}</span><MIcon name="chev" class="chev" />
      </button>
      <button class="m-li" @click="go('/about')">
        <span class="lic" style="--c: #5b6b86"><MIcon name="info" /></span>
        <span>{{ t('mobile.aboutSite') }}</span><small>{{ config.cfg.site.title }}</small>
      </button>
    </div>
    <div class="m-list admin">
      <button class="m-li" @click="go('/admin')">
        <span class="lic" style="--c: var(--solid); color: var(--on-solid)"><MIcon name="lock" /></span>
        <span>{{ t('mobile.admin') }}</span><MIcon name="chev" class="chev" />
      </button>
    </div>
    <div class="foot">
      {{ config.cfg.site.title }} · {{ t('mobile.engine') }}<br />{{ t('mobile.runDays', { n: days }) }}
    </div>
  </aside>
</template>

<style scoped lang="scss">
.drawer {
  position: absolute;
  z-index: 1;
  inset: 0 auto 0 0;
  width: min(318px, 84vw);
  padding: calc(var(--m-safe-t) + 28px) 22px calc(var(--m-safe-b) + 32px) 24px;
  overflow-y: auto;
  scrollbar-width: none;
  overscroll-behavior: contain;
  opacity: calc(0.2 + var(--m-dp) * 0.8);
  transform: translateX(calc((1 - var(--m-dp)) * -56px)) scale(calc(0.94 + var(--m-dp) * 0.06));
  transform-origin: left center;
  transition: --m-dp 0.56s var(--m-ease-drawer);

  &.dragging { transition: none; }
  &::-webkit-scrollbar { display: none; }
}

.author {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding-bottom: 8px;

  .who {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .av {
    width: 56px;
    height: 56px;
    flex: none;

    .m-avatar { border-radius: var(--r-lg); }
  }

  b {
    display: block;
    font-family: var(--font-serif);
    font-size: 22px;
    font-weight: 700;
    letter-spacing: 0.01em;
  }

  small {
    font-size: 13px;
    color: var(--text-3);
  }
}

.motto {
  font-family: var(--font-serif);
  font-size: 15px;
  line-height: 1.7;
  color: var(--text-2);

  &::before { content: '「'; color: var(--ink); }
  &::after { content: '」'; color: var(--ink); }
}

.h {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-3);
  letter-spacing: 0.08em;
  margin: 22px 0 12px;
}

.seg2 {
  position: relative;
  display: grid;
  grid-template-columns: 1fr 1fr;
  padding: 3px;
  border-radius: var(--r-md);
  background: var(--fill-2);
  box-shadow: inset 0 0 0 0.5px var(--line);

  button {
    position: relative;
    z-index: 1;
    height: 38px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 7px;
    font-size: 14px;
    color: var(--text-2);
    border-radius: calc(var(--r-md) - 3px);
    transition: color var(--dur) var(--ease-out), transform var(--dur-fast) var(--ease-spring);

    .m-ic { transition: color var(--dur); }

    &.on {
      color: var(--lift-fg);
      font-weight: 500;

      .m-ic { color: var(--ink); }
    }
  }

  .th {
    position: absolute;
    top: 3px;
    bottom: 3px;
    left: 3px;
    width: calc(50% - 3px);
    border-radius: calc(var(--r-md) - 3px);
    background: var(--lift);
    box-shadow: var(--lift-shadow);
    transform: translateX(calc(var(--i, 0) * 100%));
    transition: transform var(--dur) var(--ease-spring), background-color var(--dur), box-shadow var(--dur);
  }
}

.seg2.style { margin-top: 10px; }

.pals {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px 8px;
  margin-top: 14px;
}

.pal {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-3);

  i {
    position: relative;
    width: 46px;
    height: 46px;
    border-radius: 50%;
    background: var(--c);
    box-shadow: inset 0 0 0 0.5px rgb(0 0 0 / 0.12);

    /* 选中：外侧 1.5px 细环（--text），与色块留 3px 间隙 */
    &::after {
      content: '';
      position: absolute;
      inset: -4.5px;
      border-radius: 50%;
      box-shadow: 0 0 0 1.5px var(--text);
      opacity: 0;
      transform: scale(0.8);
      transition: all var(--dur) var(--ease-spring);
    }
  }

  &.on {
    color: var(--text);
    font-weight: 500;

    i::after {
      opacity: 1;
      transform: none;
    }
  }
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;

  button {
    display: inline-flex;
    align-items: baseline;
    gap: 5px;
    padding: 6px 12px;
    border-radius: 999px;
    box-shadow: inset 0 0 0 1px var(--line);
    color: var(--text-2);
    font-size: 13px;

    .hs { color: var(--text-3); margin-right: -3px; }

    em {
      font-style: normal;
      font-size: 11px;
      color: var(--text-3);
      font-family: var(--m-font-mono);
    }

    &.big {
      font-size: 15px;
      color: var(--text);
    }
  }
}

.admin { margin-top: 12px; }

.foot {
  margin-top: 22px;
  font-size: 12px;
  color: var(--text-3);
  line-height: 1.7;
}
</style>
