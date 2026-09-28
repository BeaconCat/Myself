<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useThemeStore, type UiStyle } from '../../stores/theme';
import { useAuthStore } from '../../stores/auth';
import { adminApi } from '../../api';
import { circularReveal } from '../../utils/circularReveal';
import UiIcon from '../ui/UiIcon.vue';

const { t } = useI18n();
const theme = useThemeStore();
const auth = useAuthStore();

/** 关闭访客换肤只影响访客：管理员永远保留色盘条 */
const showPalettes = computed(() => theme.allowUserPalette || auth.loggedIn);
/** 界面风格切换同理：关闭访客切换时管理员仍可见 */
const showStyle = computed(() => theme.allowUserStyle || auth.loggedIn);
const solo = computed(() => !showPalettes.value && !showStyle.value);

function origin(e: MouseEvent) {
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
  return { x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 };
}

/** 管理员切换 = 顺手写入全局默认主题 */
function persistAsDefault(patch: Record<string, string>): void {
  if (!auth.loggedIn) return;
  adminApi.saveSettings({ theme: patch }).catch(() => undefined);
}

function onPalette(e: MouseEvent, id: string): void {
  if (id === theme.paletteId) return;
  circularReveal(origin(e), () => theme.setPalette(id, auth.loggedIn), 'expand');
  persistAsDefault({ defaultPaletteId: id });
}

/** 界面风格：卡片 ⇄ 简洁，圆形揭幕与深浅切换同一套过渡 */
function onStyle(e: MouseEvent, style: UiStyle): void {
  if (style === theme.style) return;
  circularReveal(origin(e), () => theme.setStyle(style, auth.loggedIn), style === 'clean' ? 'expand' : 'contract');
  persistAsDefault({ defaultStyle: style });
}

function onMode(e: MouseEvent): void {
  const toLight = theme.mode === 'dark';
  circularReveal(origin(e), () => theme.toggleMode(), toLight ? 'expand' : 'contract');
  persistAsDefault({ defaultMode: toLight ? 'light' : 'dark' });
}
</script>

<template>
  <!-- 仅深浅切换时收成正圆按钮 -->
  <div class="switcher" :class="{ solo }">
    <template v-if="showPalettes">
      <button
        v-for="p in theme.allPalettes"
        :key="p.id"
        class="dot"
        :class="{ active: theme.paletteId === p.id }"
        :style="{ '--c': p[theme.mode].primary }"
        :title="p.nameKey"
        :aria-label="p.nameKey"
        :aria-pressed="theme.paletteId === p.id"
        @click="onPalette($event, p.id)"
      />
      <span class="divider" />
    </template>
    <div
      v-if="showStyle"
      class="style-seg"
      role="group"
      :aria-label="t('theme.style')"
      :style="{ '--i': theme.style === 'cards' ? 1 : 0 }"
    >
      <span class="thumb" aria-hidden="true" />
      <button
        v-for="s in (['clean', 'cards'] as const)"
        :key="s"
        type="button"
        :class="{ on: theme.style === s }"
        :title="s === 'cards' ? t('theme.toCards') : t('theme.toClean')"
        :aria-label="s === 'cards' ? t('theme.styleCards') : t('theme.styleClean')"
        :aria-pressed="theme.style === s"
        @click="onStyle($event, s)"
      >
        <UiIcon :name="s" class="style-icon" />
      </button>
    </div>
    <button
      class="mode-btn"
      :title="theme.mode === 'light' ? t('theme.dark') : t('theme.light')"
      :aria-label="theme.mode === 'light' ? t('theme.dark') : t('theme.light')"
      @click="onMode($event)"
    >
      <UiIcon :name="theme.mode === 'light' ? 'sun' : 'moon'" class="mode-icon" />
    </button>
  </div>
</template>

<style scoped lang="scss">
/* 外观胶囊：与导航胶囊同一中性玻璃面 */
.switcher {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 44px;
  padding: 0 4px 0 10px;
  /* 与导航胶囊同一圆角规则（--r-base 默认时为胶囊） */
  border-radius: var(--nav-r, var(--r-pill));
  background: color-mix(in oklab, var(--bg) 72%, transparent);
  backdrop-filter: blur(20px) saturate(170%);
  -webkit-backdrop-filter: blur(20px) saturate(170%);
  box-shadow: inset 0 0 0 0.5px var(--line-2), 0 8px 24px -16px rgb(0 0 0 / 0.5);

  :root[data-mode='light'] & {
    background: color-mix(in oklab, var(--bg) 70%, rgb(255 255 255 / 0.4));
    box-shadow: inset 0 0 0 0.5px var(--line-2), 0 1px 2px rgb(16 24 40 / 0.04), 0 10px 28px -18px rgb(16 24 40 / 0.35);
  }

  /* 正圆模式：只剩深浅切换 */
  &.solo {
    width: 44px;
    padding: 0;
    justify-content: center;
  }
}

/* 色盘圆点：实色 + 中性内描边 / 顶光；选中 = 外圈细环（文字色），无同色投影 */
.dot {
  position: relative;
  width: 26px;
  height: 26px;
  border: 0;
  border-radius: 50%;
  background: none;
  transition: transform var(--dur-fast) var(--ease-spring);

  &::before {
    content: '';
    position: absolute;
    inset: 6px;
    border-radius: 50%;
    background: var(--c);
    box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.25), inset 0 0 0 0.5px rgb(0 0 0 / 0.2);
    transition: transform var(--dur) var(--ease-spring);
  }

  &::after {
    content: '';
    position: absolute;
    inset: 2px;
    border-radius: 50%;
    box-shadow: 0 0 0 1.5px var(--text);
    opacity: 0;
    transform: scale(0.8);
    transition: opacity var(--dur) var(--ease-out), transform var(--dur) var(--ease-spring);
  }

  &:hover::before { transform: scale(1.15); }
  &:active { transform: scale(0.9); }

  &.active::after {
    opacity: 0.85;
    transform: none;
  }

  &:focus-visible {
    outline: none;
    box-shadow: var(--focus);
  }
}

.divider {
  width: 1px;
  height: 18px;
  margin: 0 4px 0 6px;
  background: var(--line-2);
}

/* 界面风格：两格迷你 segmented，选中 = 抬升 + 轻染，滑块在两格之间 morph */
.style-seg {
  position: relative;
  display: grid;
  grid-template-columns: 30px 30px;
  height: 32px;
  padding: 2px;
  margin-right: 2px;
  border-radius: min(var(--r-pill), calc(var(--nav-r-in, 16px) - 2px));
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line);

  button {
    position: relative;
    z-index: 1;
    display: grid;
    place-items: center;
    border: 0;
    border-radius: inherit;
    background: none;
    color: var(--text-3);
    transition: color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

    &:hover { color: var(--text); }
    &:active { transform: scale(0.9); }
    &.on { color: var(--ink); }

    &:focus-visible {
      outline: none;
      box-shadow: var(--focus);
    }
  }

  .thumb {
    position: absolute;
    top: 2px;
    bottom: 2px;
    left: 2px;
    width: 30px;
    border-radius: inherit;
    background: var(--lift);
    box-shadow: var(--lift-shadow);
    transform: translateX(calc(var(--i, 0) * 100%));
    transition: transform var(--dur) var(--ease-spring);
  }
}

.style-icon {
  width: 16px;
  height: 16px;
}

/* 深浅切换：中性图标按钮 */
.mode-btn {
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  border: 0;
  border-radius: min(50%, var(--nav-r-in, 50%));
  background: none;
  color: var(--text-2);
  transition: color var(--dur-fast), background-color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  &:hover {
    color: var(--text);
    background: var(--fill-2);
  }

  &:active { transform: scale(0.92); }

  &:focus-visible {
    outline: none;
    box-shadow: var(--focus);
  }
}

.mode-icon {
  width: 18px;
  height: 18px;
  transition: transform var(--dur) var(--ease-spring);

  .mode-btn:hover & { transform: rotate(18deg); }
}
</style>
