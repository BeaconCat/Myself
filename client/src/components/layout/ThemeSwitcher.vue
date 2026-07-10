<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useThemeStore } from '../../stores/theme';

const { t } = useI18n();
const theme = useThemeStore();
</script>

<template>
  <div class="switcher">
    <!-- 色盘：四季圆点，选中放大描边 -->
    <button
      v-for="p in theme.allPalettes"
      :key="p.id"
      class="dot"
      :class="{ active: theme.paletteId === p.id }"
      :style="{ background: p[theme.mode].primary }"
      :title="t(p.nameKey)"
      @click="theme.setPalette(p.id)"
    />
    <span class="divider" />
    <!-- 深浅切换：日/月形状纯 CSS 图标 -->
    <button
      class="mode-btn"
      :title="theme.mode === 'light' ? t('theme.dark') : t('theme.light')"
      @click="theme.toggleMode()"
    >
      <span class="mode-icon" :class="theme.mode" />
    </button>
  </div>
</template>

<style scoped lang="scss">
.switcher {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 999px;
  background: var(--glass);
  backdrop-filter: blur(14px) saturate(1.5);
  -webkit-backdrop-filter: blur(14px) saturate(1.5);
  border: 1px solid rgba(var(--primary-rgb), 0.18);
  box-shadow: var(--shadow), inset 0 1px 0 rgba(255, 255, 255, 0.12);
}

.dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid transparent;
  transition: transform var(--dur-fast) var(--ease-spring), border-color var(--dur-fast), box-shadow var(--dur-fast);

  &:hover { transform: scale(1.2); }

  &.active {
    border-color: var(--text);
    transform: scale(1.25);
    box-shadow: 0 0 10px rgba(var(--primary-rgb), 0.6);
  }
}

.divider {
  width: 1px;
  height: 16px;
  background: var(--border);
}

.mode-btn {
  border: none;
  background: none;
  width: 24px;
  height: 24px;
  display: grid;
  place-items: center;
}

/* 日/月纯 CSS 图标：light 显示太阳（圆+光晕），dark 显示月牙 */
.mode-icon {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  transition: all var(--dur) var(--ease-spring);

  &.light {
    background: var(--text);
    box-shadow: 0 0 0 2.5px var(--bg), 0 0 0 4px var(--text-2);
  }

  &.dark {
    background: transparent;
    box-shadow: inset -4px -3px 0 0 var(--text);
    transform: rotate(-20deg);
  }
}
</style>
